package resolver

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/grpc/federation"
	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/types"
)

type Resolver struct {
	files                      []*descriptorpb.FileDescriptorProto
	celRegistry                *CELRegistry
	defToFileMap               map[*descriptorpb.FileDescriptorProto]*File
	protoPackageNameToFileDefs map[string][]*descriptorpb.FileDescriptorProto
	protoPackageNameToPackage  map[string]*Package

	serviceToRuleMap   map[*Service]*federation.ServiceRule
	methodToRuleMap    map[*Method]*federation.MethodRule
	messageToRuleMap   map[*Message]*federation.MessageRule
	enumToRuleMap      map[*Enum]*federation.EnumRule
	enumValueToRuleMap map[*EnumValue]*federation.EnumValueRule
	fieldToRuleMap     map[*Field]*federation.FieldRule
	oneofToRuleMap     map[*Oneof]*federation.OneofRule

	cachedMessageMap map[string]*Message
	cachedEnumMap    map[string]*Enum
	cachedMethodMap  map[string]*Method
	cachedServiceMap map[string]*Service
}

func New(files []*descriptorpb.FileDescriptorProto) *Resolver {
	msgMap := make(map[string]*Message)
	return &Resolver{
		files:                      files,
		celRegistry:                newCELRegistry(msgMap),
		defToFileMap:               make(map[*descriptorpb.FileDescriptorProto]*File),
		protoPackageNameToFileDefs: make(map[string][]*descriptorpb.FileDescriptorProto),
		protoPackageNameToPackage:  make(map[string]*Package),

		serviceToRuleMap:   make(map[*Service]*federation.ServiceRule),
		methodToRuleMap:    make(map[*Method]*federation.MethodRule),
		messageToRuleMap:   make(map[*Message]*federation.MessageRule),
		enumToRuleMap:      make(map[*Enum]*federation.EnumRule),
		enumValueToRuleMap: make(map[*EnumValue]*federation.EnumValueRule),
		fieldToRuleMap:     make(map[*Field]*federation.FieldRule),
		oneofToRuleMap:     make(map[*Oneof]*federation.OneofRule),

		cachedMessageMap: msgMap,
		cachedEnumMap:    make(map[string]*Enum),
		cachedMethodMap:  make(map[string]*Method),
		cachedServiceMap: make(map[string]*Service),
	}
}

// Result of resolver processing.
type Result struct {
	// Services list of services that specify the grpc.federation.service option.
	Services []*Service
	// Warnings all warnings occurred during the resolve process.
	Warnings []*Warning
}

// Warning represents what should be warned that is not an error that occurred during the Resolver.Resolve().
type Warning struct {
	Location *source.Location
	Message  string
}

func (r *Resolver) ResolveWellknownFiles() (Files, error) {
	ctx := newContext()

	fds := stdFileDescriptors()
	r.resolvePackageAndFileReference(ctx, fds)

	files := make([]*File, 0, len(fds))
	for _, fileDef := range fds {
		files = append(files, r.resolveFile(ctx, fileDef))
	}
	return files, ctx.error()
}

func (r *Resolver) Resolve() (*Result, error) {
	if err := r.celRegistry.RegisterFiles(r.files...); err != nil {
		return nil, err
	}
	// In order to return multiple errors with source code location information,
	// we add all errors to the context when they occur.
	// Therefore, functions called from Resolve() do not return errors directly.
	// Instead, it must return all errors captured by context in ctx.error().
	ctx := newContext()

	r.resolvePackageAndFileReference(ctx, r.files)

	files := r.resolveFiles(ctx)

	r.resolveRule(ctx, files)

	if !r.existsServiceRule(files) {
		return &Result{Warnings: ctx.warnings()}, ctx.error()
	}

	r.resolveMessageArgument(ctx, files)
	r.resolveMessageDependencies(ctx, files)

	services := r.servicesWithRule(ctx, files)
	return &Result{Services: services, Warnings: ctx.warnings()}, ctx.error()
}

// resolvePackageAndFileReference create instances of Package and File to be used inside the resolver from all file descriptor and link them together.
// This process must always be done at the beginning of the Resolve().
func (r *Resolver) resolvePackageAndFileReference(ctx *context, files []*descriptorpb.FileDescriptorProto) {
	for _, fileDef := range files {
		protoPackageName := fileDef.GetPackage()
		pkg, exists := r.protoPackageNameToPackage[protoPackageName]
		if !exists {
			pkg = &Package{Name: fileDef.GetPackage()}
		}
		file := &File{Name: fileDef.GetName()}
		gopkg, err := ResolveGoPackage(fileDef)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.GoPackageLocation(fileDef.GetName()),
				),
			)
		} else {
			file.GoPackage = gopkg
		}
		file.Package = pkg
		file.Desc = fileDef
		pkg.Files = append(pkg.Files, file)

		r.defToFileMap[fileDef] = file
		r.protoPackageNameToFileDefs[protoPackageName] = append(
			r.protoPackageNameToFileDefs[protoPackageName],
			fileDef,
		)
		r.protoPackageNameToPackage[protoPackageName] = pkg
	}
}

// resolveFiles resolve all references except custom option.
func (r *Resolver) resolveFiles(ctx *context) []*File {
	files := make([]*File, 0, len(r.files))
	for _, fileDef := range r.files {
		files = append(files, r.resolveFile(ctx, fileDef))
	}
	return files
}

func ResolveGoPackage(def *descriptorpb.FileDescriptorProto) (*GoPackage, error) {
	opts := def.GetOptions()
	if opts == nil {
		return nil, nil
	}
	importPath, gopkgName, err := splitGoPackageName(opts.GetGoPackage())
	if err != nil {
		return nil, err
	}
	return &GoPackage{
		Name:       gopkgName,
		ImportPath: importPath,
	}, nil
}

func (r *Resolver) existsServiceRule(files []*File) bool {
	for _, file := range files {
		for _, service := range file.Services {
			if service.Rule != nil {
				return true
			}
		}
	}
	return false
}

func (r *Resolver) allMessages(files []*File) []*Message {
	msgs := make([]*Message, 0, len(r.cachedMessageMap))
	for _, file := range files {
		msgs = append(msgs, file.Messages...)
	}
	return msgs
}

// servicesWithRule returns all services that have a federation rule.
func (r *Resolver) servicesWithRule(ctx *context, files []*File) []*Service {
	services := make([]*Service, 0, len(r.cachedServiceMap))
	for _, file := range files {
		ctx := ctx.withFile(file)
		for _, service := range file.Services {
			ctx := ctx.withService(service)
			if service.Rule != nil {
				r.validateServiceDependency(ctx, service)
				r.validateMethodResponse(ctx, service)
				services = append(services, service)
			}
		}
	}
	return services
}

func (r *Resolver) validateServiceDependency(ctx *context, service *Service) {
	useSvcMap := map[string]struct{}{}
	useServices := service.UseServices()
	for _, svc := range useServices {
		useSvcMap[svc.FQDN()] = struct{}{}
	}
	depSvcMap := map[string]struct{}{}
	for idx, dep := range service.Rule.Dependencies {
		if dep.Service == nil {
			continue
		}
		depSvcName := dep.Service.FQDN()
		if _, exists := useSvcMap[depSvcName]; !exists {
			ctx.addWarning(&Warning{
				Location: source.ServiceDependencyLocation(service.File.Name, service.Name, idx),
				Message:  fmt.Sprintf(`%q defined in "dependencies" of "grpc.federation.service" but it is not used`, depSvcName),
			})
		}
		depSvcMap[depSvcName] = struct{}{}
	}
}

func (r *Resolver) validateMethodResponse(ctx *context, service *Service) {
	for _, method := range service.Methods {
		response := method.Response
		if response.Rule == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`"%s.%s" message needs to specify "grpc.federation.message" option`, response.PackageName(), response.Name),
					source.MessageLocation(ctx.fileName(), response.Name),
				),
			)
		}
	}
}

func (r *Resolver) resolveFile(ctx *context, def *descriptorpb.FileDescriptorProto) *File {
	file := r.defToFileMap[def]
	ctx = ctx.withFile(file)
	for _, serviceDef := range def.GetService() {
		service := r.resolveService(ctx, file.Package, serviceDef.GetName())
		if service == nil {
			continue
		}
		file.Services = append(file.Services, service)
	}
	for _, msgDef := range def.GetMessageType() {
		msg := r.resolveMessage(ctx, file.Package, msgDef.GetName())
		if msg == nil {
			continue
		}
		file.Messages = append(file.Messages, msg)
	}
	for _, enumDef := range def.GetEnumType() {
		file.Enums = append(file.Enums, r.resolveEnum(ctx, file.Package, enumDef.GetName()))
	}
	return file
}

func (r *Resolver) resolveService(ctx *context, pkg *Package, name string) *Service {
	fqdn := fmt.Sprintf("%s.%s", pkg.Name, name)
	cachedService, exists := r.cachedServiceMap[fqdn]
	if exists {
		return cachedService
	}
	file, serviceDef, err := r.lookupService(pkg, name)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.ServiceLocation(ctx.fileName(), name),
			),
		)
		return nil
	}
	ruleDef, err := getServiceRule(serviceDef)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.ServiceOptionLocation(ctx.fileName(), name),
			),
		)
		return nil
	}
	service := &Service{
		File:    file,
		Name:    name,
		Methods: make([]*Method, 0, len(serviceDef.GetMethod())),
	}
	r.serviceToRuleMap[service] = ruleDef
	ctx = ctx.withService(service)
	for _, methodDef := range serviceDef.GetMethod() {
		method := r.resolveMethod(ctx, service, methodDef)
		if method == nil {
			continue
		}
		service.Methods = append(service.Methods, method)
	}
	r.cachedServiceMap[fqdn] = service
	return service
}

func (r *Resolver) resolveServiceDependency(ctx *context, def *federation.ServiceDependency) *ServiceDependency {
	var service *Service
	serviceWithPkgName := def.GetService()
	if serviceWithPkgName == "" {
		ctx.addError(
			ErrWithLocation(
				`"service" must be specified`,
				source.ServiceDependencyServiceLocation(ctx.fileName(), ctx.serviceName(), ctx.depIndex()),
			),
		)
	} else {
		pkg, err := r.lookupPackage(serviceWithPkgName)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.ServiceDependencyServiceLocation(ctx.fileName(), ctx.serviceName(), ctx.depIndex()),
				),
			)
		}
		if pkg != nil {
			serviceName := r.trimPackage(pkg, serviceWithPkgName)
			service = r.resolveService(ctx, pkg, serviceName)
			if service == nil {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`%q does not exist`, serviceWithPkgName),
						source.ServiceDependencyServiceLocation(ctx.fileName(), ctx.serviceName(), ctx.depIndex()),
					),
				)
			}
		}
	}
	return &ServiceDependency{Name: def.GetName(), Service: service}
}

func (r *Resolver) resolveMethod(ctx *context, service *Service, methodDef *descriptorpb.MethodDescriptorProto) *Method {
	fqdn := fmt.Sprintf("%s.%s/%s", service.PackageName(), service.Name, methodDef.GetName())
	cachedMethod, exists := r.cachedMethodMap[fqdn]
	if exists {
		return cachedMethod
	}
	reqPkg, err := r.lookupPackage(methodDef.GetInputType())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.ServiceMethodLocation(ctx.fileName(), service.Name, methodDef.GetName()),
			),
		)
	}
	resPkg, err := r.lookupPackage(methodDef.GetOutputType())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.ServiceMethodLocation(ctx.fileName(), service.Name, methodDef.GetName()),
			),
		)
	}
	var (
		req *Message
		res *Message
	)
	if reqPkg != nil {
		reqType := r.trimPackage(reqPkg, methodDef.GetInputType())
		req = r.resolveMessage(ctx, reqPkg, reqType)
	}
	if resPkg != nil {
		resType := r.trimPackage(resPkg, methodDef.GetOutputType())
		res = r.resolveMessage(ctx, resPkg, resType)
	}
	ruleDef, err := getMethodRule(methodDef)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.ServiceMethodOptionLocation(ctx.fileName(), service.Name, methodDef.GetName()),
			),
		)
		return nil
	}
	method := &Method{
		Service:  service,
		Name:     methodDef.GetName(),
		Request:  req,
		Response: res,
	}
	r.methodToRuleMap[method] = ruleDef
	r.cachedMethodMap[fqdn] = method
	return method
}

func (r *Resolver) resolveMessage(ctx *context, pkg *Package, name string) *Message {
	fqdn := fmt.Sprintf("%s.%s", pkg.Name, name)
	cachedMessage, exists := r.cachedMessageMap[fqdn]
	if exists {
		return cachedMessage
	}
	file, msgDef, err := r.lookupMessage(pkg, name)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageLocation(ctx.fileName(), name),
			),
		)
		return nil
	}
	msg := &Message{File: file, Name: msgDef.GetName()}
	for _, nestedMsgDef := range msgDef.GetNestedType() {
		nestedMsg := r.resolveMessage(ctx, pkg, fmt.Sprintf("%s.%s", name, nestedMsgDef.GetName()))
		if nestedMsg == nil {
			continue
		}
		nestedMsg.ParentMessage = msg
		msg.NestedMessages = append(msg.NestedMessages, nestedMsg)
	}
	for _, enumDef := range msgDef.GetEnumType() {
		enum := r.resolveEnum(ctx, pkg, fmt.Sprintf("%s.%s", name, enumDef.GetName()))
		if enum == nil {
			continue
		}
		enum.Message = msg
		msg.Enums = append(msg.Enums, enum)
	}
	opt := msgDef.GetOptions()
	msg.IsMapEntry = opt.GetMapEntry()
	rule, err := getMessageRule(msgDef)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageOptionLocation(ctx.fileName(), msgDef.GetName()),
			),
		)
	}
	r.cachedMessageMap[fqdn] = msg
	r.messageToRuleMap[msg] = rule
	ctx = ctx.withMessage(msg)
	var oneofs []*Oneof
	for _, oneofDef := range msgDef.GetOneofDecl() {
		oneof := r.resolveOneof(ctx, oneofDef)
		oneof.Message = msg
		oneofs = append(oneofs, oneof)
	}
	msg.Fields = r.resolveFields(ctx, msgDef.GetField(), oneofs)
	msg.Oneofs = oneofs
	return msg
}

func (r *Resolver) resolveOneof(ctx *context, def *descriptorpb.OneofDescriptorProto) *Oneof {
	rule, err := getOneofRule(def)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.OneofOptionLocation(ctx.fileName(), ctx.messageName(), def.GetName()),
			),
		)
	}
	oneof := &Oneof{
		Name: def.GetName(),
	}
	r.oneofToRuleMap[oneof] = rule
	return oneof
}

func (r *Resolver) resolveEnum(ctx *context, pkg *Package, name string) *Enum {
	fqdn := fmt.Sprintf("%s.%s", pkg.Name, name)
	cachedEnum, exists := r.cachedEnumMap[fqdn]
	if exists {
		return cachedEnum
	}
	file, def, err := r.lookupEnum(pkg, name)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.EnumLocation(ctx.fileName(), ctx.messageName(), name),
			),
		)
		return nil
	}
	values := make([]*EnumValue, 0, len(def.GetValue()))
	enum := &Enum{
		File: file,
		Name: def.GetName(),
	}
	for _, valueDef := range def.GetValue() {
		valueName := valueDef.GetName()
		rule, err := getEnumValueRule(valueDef)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.EnumValueLocation(ctx.fileName(), ctx.messageName(), name, valueName),
				),
			)
		}
		enumValue := &EnumValue{Value: valueName, Enum: enum}
		values = append(values, enumValue)
		r.enumValueToRuleMap[enumValue] = rule
	}
	enum.Values = values
	rule, err := getEnumRule(def)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.EnumLocation(ctx.fileName(), ctx.messageName(), name),
			),
		)
	}
	r.cachedEnumMap[fqdn] = enum
	r.enumToRuleMap[enum] = rule
	return enum
}

// resolveRule resolve the rule defined in grpc.federation custom option.
func (r *Resolver) resolveRule(ctx *context, files []*File) {
	for _, file := range files {
		ctx := ctx.withFile(file)
		r.resolveServiceRules(ctx, file.Services)
		r.resolveMessageRules(ctx, file.Messages)
		r.resolveEnumRules(ctx, file.Enums)
	}
}

func (r *Resolver) resolveServiceRules(ctx *context, svcs []*Service) {
	for _, svc := range svcs {
		ctx := ctx.withService(svc)
		svc.Rule = r.resolveServiceRule(ctx, r.serviceToRuleMap[svc])
		r.resolveMethodRules(ctx, svc.Methods)
	}
}

func (r *Resolver) resolveMethodRules(ctx *context, mtds []*Method) {
	for _, mtd := range mtds {
		mtd.Rule = r.resolveMethodRule(ctx.withMethod(mtd), r.methodToRuleMap[mtd])
	}
}

func (r *Resolver) resolveMessageRules(ctx *context, msgs []*Message) {
	for _, msg := range msgs {
		ctx := ctx.withMessage(msg)
		r.resolveMessageRule(ctx, msg, r.messageToRuleMap[msg])
		r.resolveFieldRules(ctx, msg)
		r.resolveEnumRules(ctx, msg.Enums)
		r.resolveAutoBindFields(ctx, msg)
		if msg.HasRule() {
			if msg.HasCustomResolver() || msg.HasCustomResolverFields() {
				// If use custom resolver, set the `Used` flag true
				// because all dependency message references are passed as arguments for custom resolver.
				msg.UseAllNameReference()
			}
		}
		r.resolveMessageRules(ctx, msg.NestedMessages)
	}
}

func (r *Resolver) resolveFieldRules(ctx *context, msg *Message) {
	for _, field := range msg.Fields {
		field.Rule = r.resolveFieldRule(ctx, msg, field, r.fieldToRuleMap[field])
	}
}

func (r *Resolver) resolveAutoBindFields(ctx *context, msg *Message) {
	if msg.Rule == nil {
		return
	}
	if msg.HasCustomResolver() {
		return
	}
	rule := msg.Rule
	autobindFieldMap := make(map[string][]*AutoBindField)
	if rule.MethodCall != nil && rule.MethodCall.Response != nil {
		for _, responseField := range rule.MethodCall.Response.Fields {
			if !responseField.AutoBind {
				continue
			}
			if responseField.Type == nil {
				continue
			}
			if responseField.Type.Ref == nil {
				continue
			}
			for _, field := range responseField.Type.Ref.Fields {
				autobindFieldMap[field.Name] = append(autobindFieldMap[field.Name], &AutoBindField{
					ResponseField: responseField,
					Field:         field,
				})
			}
		}
	}
	for _, dep := range rule.MessageDependencies {
		if !dep.AutoBind {
			continue
		}
		for _, field := range dep.Message.Fields {
			autobindFieldMap[field.Name] = append(autobindFieldMap[field.Name], &AutoBindField{
				MessageDependency: dep,
				Field:             field,
			})
		}
	}
	for _, field := range msg.Fields {
		if field.HasRule() {
			continue
		}
		autoBindFields, exists := autobindFieldMap[field.Name]
		if !exists {
			continue
		}
		if len(autoBindFields) > 1 {
			var locates []string
			for _, autoBindField := range autoBindFields {
				if autoBindField.ResponseField != nil {
					locates = append(locates, fmt.Sprintf(`%q name at response`, autoBindField.ResponseField.Name))
				} else if autoBindField.MessageDependency != nil {
					locates = append(locates, fmt.Sprintf(`%q name at messages`, autoBindField.MessageDependency.Name))
				}
			}
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are %s`, field.Name, strings.Join(locates, " and ")),
					source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
				),
			)
			continue
		}
		autoBindField := autoBindFields[0]
		if autoBindField.Field.Type == nil || field.Type == nil {
			continue
		}
		if autoBindField.Field.Type.Type != field.Type.Type {
			continue
		}
		switch {
		case autoBindField.ResponseField != nil:
			autoBindField.ResponseField.Used = true
		case autoBindField.MessageDependency != nil:
			autoBindField.MessageDependency.Used = true
		}
		field.Rule = &FieldRule{
			AutoBindField: autoBindField,
		}
	}
}

const namePattern = `^[a-zA-Z][a-zA-Z0-9_]*$`

var nameRe = regexp.MustCompile(namePattern)

func (r *Resolver) validateName(name string) error {
	if name == "" {
		return nil
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf(`%q is invalid name. name should be in the following pattern: %s`, name, namePattern)
	}
	return nil
}

func (r *Resolver) validateMessages(ctx *context, msgs []*Message) {
	for _, msg := range msgs {
		ctx := ctx.withFile(msg.File).withMessage(msg)
		r.validateMessageFields(ctx, msg)
		r.validateMessages(ctx, msg.NestedMessages)
	}
}

func (r *Resolver) validateMessageFields(ctx *context, msg *Message) {
	if msg.Rule == nil {
		return
	}
	for _, field := range msg.Fields {
		if !field.HasRule() {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q field in %q message needs to specify "grpc.federation.field" option`, field.Name, msg.FQDN()),
					source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
				),
			)
			continue
		}
		if field.HasMessageCustomResolver() || field.HasCustomResolver() {
			continue
		}
		if field.Type == nil {
			continue
		}
		fieldType := field.Type
		if fieldType.Type != types.Message && fieldType.Type != types.Enum {
			continue
		}
		rule := field.Rule
		if rule.Value != nil {
			r.validateBindFieldType(ctx, rule.Value.Type(), field)
		}
		if rule.Alias != nil {
			r.validateBindFieldType(ctx, rule.Alias.Type, field)
		}
		if rule.AutoBindField != nil {
			r.validateBindFieldType(ctx, rule.AutoBindField.Field.Type, field)
		}
	}
}

func (r *Resolver) validateBindFieldType(ctx *context, fromType *Type, toField *Field) {
	if fromType == nil || toField == nil {
		return
	}
	toType := toField.Type
	if fromType.Type == types.Message {
		if fromType.Ref == nil || toType.Ref == nil {
			return
		}
		if fromType.Ref.IsMapEntry {
			// If it is a map entry, ignore it.
			return
		}
		fromMessageName := fromType.Ref.FQDN()
		toMessage := toType.Ref
		toMessageName := toMessage.FQDN()
		if fromMessageName == toMessageName {
			// assignment of the same type is okay.
			return
		}
		if toMessage.Rule == nil || toMessage.Rule.Alias == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromMessageName, toMessageName, ctx.messageName(), toField.Name,
					),
					source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), toField.Name),
				),
			)
			return
		}
		toMessageAliasName := toMessage.Rule.Alias.FQDN()
		if toMessageAliasName != fromMessageName {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromMessageName, toMessageName, ctx.messageName(), toField.Name,
					),
					source.MessageAliasLocation(toMessage.File.Name, toMessage.Name),
				),
			)
			return
		}
		return
	}
	if fromType.Type == types.Enum {
		if fromType.Enum == nil || toType.Enum == nil {
			return
		}
		fromEnumName := fromType.Enum.FQDN()
		toEnum := toType.Enum
		toEnumName := toEnum.FQDN()
		if toEnumName == fromEnumName {
			// assignment of the same type is okay.
			return
		}
		var toEnumMessageName string
		if toEnum.Message != nil {
			toEnumMessageName = toEnum.Message.Name
		}
		if toEnum.Rule == nil || toEnum.Rule.Alias == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.enum option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromEnumName, toEnumName, ctx.messageName(), toField.Name,
					),
					source.EnumLocation(ctx.fileName(), toEnumMessageName, toEnum.Name),
				),
			)
			return
		}
		toEnumAliasName := toEnum.Rule.Alias.FQDN()
		if toEnumAliasName != fromEnumName {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.enum option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromEnumName, toEnumName, ctx.messageName(), toField.Name,
					),
					source.EnumAliasLocation(ctx.fileName(), toEnumMessageName, toEnum.Name),
				),
			)
		}
		return
	}
}

func (r *Resolver) resolveEnumRules(ctx *context, enums []*Enum) {
	for _, enum := range enums {
		ctx := ctx.withEnum(enum)
		enum.Rule = r.resolveEnumRule(ctx, r.enumToRuleMap[enum])
		for _, value := range enum.Values {
			value.Rule = r.resolveEnumValueRule(ctx, enum, value, r.enumValueToRuleMap[value])
		}
	}
}

func (r *Resolver) resolveServiceRule(ctx *context, def *federation.ServiceRule) *ServiceRule {
	if def == nil {
		return nil
	}
	deps := make([]*ServiceDependency, 0, len(def.GetDependencies()))
	svcNameMap := map[string]struct{}{}
	for idx, depDef := range def.GetDependencies() {
		dep := r.resolveServiceDependency(ctx.withDepIndex(idx), depDef)
		if dep.Service == nil {
			continue
		}
		if dep.Name != "" {
			if _, exists := svcNameMap[dep.Name]; exists {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`%q name duplicated`, dep.Name),
						source.ServiceDependencyNameLocation(ctx.fileName(), ctx.serviceName(), idx),
					),
				)
			}
			svcNameMap[dep.Name] = struct{}{}
		}
		deps = append(deps, dep)
	}
	return &ServiceRule{Dependencies: deps}
}

func (r *Resolver) resolveMethodRule(ctx *context, def *federation.MethodRule) *MethodRule {
	if def == nil {
		return nil
	}
	rule := &MethodRule{}
	timeout := def.GetTimeout()
	if timeout != "" {
		duration, err := time.ParseDuration(timeout)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.ServiceMethodTimeoutLocation(ctx.fileName(), ctx.serviceName(), ctx.methodName()),
				),
			)
		} else {
			rule.Timeout = &duration
		}
	}
	return rule
}

func (r *Resolver) resolveMessageRule(ctx *context, msg *Message, ruleDef *federation.MessageRule) {
	if ruleDef == nil {
		return
	}
	methodCall := r.resolveMethodCall(ctx, ruleDef.Resolver)
	depMsgs := r.resolveMessages(ctx, msg, ruleDef.Messages)
	for _, depMsg := range depMsgs {
		depMsg.Owner = &MessageDependencyOwner{
			Type:    MessageDependencyOwnerMessage,
			Message: msg,
		}
	}
	msg.Rule = &MessageRule{
		MethodCall:          methodCall,
		MessageDependencies: depMsgs,
		CustomResolver:      ruleDef.GetCustomResolver(),
		Alias:               r.resolveMessageAlias(ctx, ruleDef.GetAlias()),
		DependencyGraph:     &MessageDependencyGraph{},
	}
}

func (r *Resolver) resolveMessageAlias(ctx *context, aliasName string) *Message {
	if aliasName == "" {
		return nil
	}
	if strings.Contains(aliasName, ".") {
		pkg, err := r.lookupPackage(aliasName)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MessageAliasLocation(ctx.fileName(), ctx.messageName()),
				),
			)
			return nil
		}
		name := r.trimPackage(pkg, aliasName)
		return r.resolveMessage(ctx, pkg, name)
	}
	return r.resolveMessage(ctx, ctx.file().Package, aliasName)
}

func (r *Resolver) resolveFieldRule(ctx *context, msg *Message, field *Field, ruleDef *federation.FieldRule) *FieldRule {
	if ruleDef == nil {
		if msg.Rule == nil {
			return nil
		}
		if msg.Rule.CustomResolver {
			return &FieldRule{MessageCustomResolver: true}
		}
		if msg.Rule.Alias != nil {
			alias := r.resolveFieldAlias(ctx, msg, field, "")
			if alias == nil {
				return nil
			}
			return &FieldRule{Alias: alias}
		}
		return nil
	}
	oneof := r.resolveFieldOneofRule(ctx, msg, field, ruleDef.GetOneof())
	var value *Value
	if oneof == nil {
		v, err := r.resolveValue(ctx, fieldRuleToCommonValueDef(ruleDef))
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
				),
			)
			return nil
		}
		value = v
	}
	return &FieldRule{
		Value:          value,
		CustomResolver: ruleDef.GetCustomResolver(),
		Alias:          r.resolveFieldAlias(ctx, msg, field, ruleDef.GetAlias()),
		Oneof:          oneof,
	}
}

func (r *Resolver) resolveFieldOneofRule(ctx *context, msg *Message, field *Field, def *federation.FieldOneof) *FieldOneofRule {
	if def == nil {
		return nil
	}
	if field.Oneof == nil {
		ctx.addError(
			ErrWithLocation(
				`"oneof" feature can only be used for fields within oneof`,
				source.MessageFieldOneofLocation(ctx.fileName(), msg.Name, field.Name),
			),
		)
		return nil
	}
	depMsgs := r.resolveMessages(ctx, msg, def.GetMessages())
	for _, depMsg := range depMsgs {
		depMsg.Owner = &MessageDependencyOwner{
			Type:  MessageDependencyOwnerOneofField,
			Field: field,
		}
	}
	return &FieldOneofRule{
		Expr:                &CELValue{Expr: def.GetExpr()},
		MessageDependencies: depMsgs,
		By:                  &CELValue{Expr: def.GetBy()},
	}
}

func (r *Resolver) resolveFieldRuleByAutoAlias(ctx *context, msg *Message, field *Field) *Field {
	if msg.Rule == nil {
		return nil
	}
	if msg.Rule.Alias == nil {
		return nil
	}
	msgAlias := msg.Rule.Alias
	aliasField := msgAlias.Field(field.Name)
	if aliasField == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`specified "alias" in grpc.federation.message option, but %q field does not exist in %q message`,
					field.Name, msgAlias.FQDN(),
				),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if field.Type.Type != aliasField.Type.Type {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, types.ToString(field.Type.Type),
					msgAlias.FQDN(), types.ToString(aliasField.Type.Type),
				),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
			),
		)
		return nil
	}
	return aliasField
}

func (r *Resolver) resolveFieldAlias(ctx *context, msg *Message, field *Field, fieldAlias string) *Field {
	if fieldAlias == "" {
		return r.resolveFieldRuleByAutoAlias(ctx, msg, field)
	}
	if msg.Rule == nil || msg.Rule.Alias == nil {
		ctx.addError(
			ErrWithLocation(
				`use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option`,
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
			),
		)
		return nil
	}
	msgAlias := msg.Rule.Alias
	aliasField := msgAlias.Field(fieldAlias)
	if aliasField == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q field does not exist in %q message`, fieldAlias, msgAlias.FQDN()),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if field.Type.Type != aliasField.Type.Type {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, types.ToString(field.Type.Type),
					msgAlias.FQDN(), types.ToString(aliasField.Type.Type),
				),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), field.Name),
			),
		)
		return nil
	}
	return aliasField
}

func (r *Resolver) resolveEnumRule(ctx *context, ruleDef *federation.EnumRule) *EnumRule {
	if ruleDef == nil {
		return nil
	}
	return &EnumRule{
		Alias: r.resolveEnumAlias(ctx, ruleDef.GetAlias()),
	}
}

func (r *Resolver) resolveEnumAlias(ctx *context, aliasName string) *Enum {
	if aliasName == "" {
		return nil
	}
	if strings.Contains(aliasName, ".") {
		pkg, err := r.lookupPackage(aliasName)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.EnumAliasLocation(ctx.fileName(), ctx.messageName(), ctx.enumName()),
				),
			)
			return nil
		}
		name := r.trimPackage(pkg, aliasName)
		return r.resolveEnum(ctx, pkg, name)
	}
	return r.resolveEnum(ctx, ctx.file().Package, aliasName)
}

func (r *Resolver) resolveEnumValueRule(ctx *context, enum *Enum, enumValue *EnumValue, ruleDef *federation.EnumValueRule) *EnumValueRule {
	if ruleDef == nil {
		if enum.Rule == nil {
			return nil
		}
		if enum.Rule.Alias != nil {
			return r.resolveEnumValueRuleByAutoAlias(ctx, enum, enumValue.Value)
		}
		return nil
	}
	aliases := make([]*EnumValue, 0, len(ruleDef.GetAlias()))
	for _, aliasName := range ruleDef.GetAlias() {
		alias := r.resolveEnumValueAlias(ctx, enum, enumValue.Value, aliasName)
		if alias == nil {
			break
		}
		aliases = append(aliases, alias)
	}
	return &EnumValueRule{
		Default: ruleDef.GetDefault(),
		Aliases: aliases,
	}
}

func (r *Resolver) resolveEnumValueRuleByAutoAlias(ctx *context, enum *Enum, enumValueName string) *EnumValueRule {
	if enum.Rule == nil {
		return nil
	}
	if enum.Rule.Alias == nil {
		return nil
	}
	enumAlias := enum.Rule.Alias
	enumValue := enumAlias.Value(enumValueName)
	if enumValue == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`specified "alias" in grpc.federation.enum option, but %q value does not exist in %q enum`,
					enumValueName, enumAlias.FQDN(),
				),
				source.EnumValueLocation(ctx.fileName(), ctx.messageName(), ctx.enumName(), enumValueName),
			),
		)
		return nil
	}
	return &EnumValueRule{Aliases: []*EnumValue{enumValue}}
}

func (r *Resolver) resolveEnumValueAlias(ctx *context, enum *Enum, enumValueName, enumValueAlias string) *EnumValue {
	if enumValueAlias == "" {
		return nil
	}
	if enum.Rule == nil || enum.Rule.Alias == nil {
		ctx.addError(
			ErrWithLocation(
				`use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option`,
				source.EnumValueAliasLocation(ctx.fileName(), ctx.messageName(), ctx.enumName(), enumValueName),
			),
		)
		return nil
	}
	enumAlias := enum.Rule.Alias
	value := enumAlias.Value(enumValueAlias)
	if value == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q value does not exist in %q enum`, enumValueAlias, enumAlias.FQDN()),
				source.EnumValueLocation(ctx.fileName(), ctx.messageName(), ctx.enumName(), enumValueName),
			),
		)
		return nil
	}
	return value
}

func (r *Resolver) resolveFields(ctx *context, fieldsDef []*descriptorpb.FieldDescriptorProto, oneofs []*Oneof) []*Field {
	fields := make([]*Field, 0, len(fieldsDef))
	for _, fieldDef := range fieldsDef {
		field := r.resolveField(ctx, fieldDef, oneofs)
		if field == nil {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func (r *Resolver) resolveField(ctx *context, fieldDef *descriptorpb.FieldDescriptorProto, oneofs []*Oneof) *Field {
	typ, err := r.resolveType(ctx, fieldDef.GetTypeName(), fieldDef.GetType(), fieldDef.GetLabel())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), fieldDef.GetName()),
			),
		)
		return nil
	}
	field := &Field{Name: fieldDef.GetName(), Type: typ}
	if fieldDef.OneofIndex != nil {
		oneof := oneofs[fieldDef.GetOneofIndex()]
		oneof.Fields = append(oneof.Fields, field)
		field.Oneof = oneof
		typ.OneofField = &OneofField{Field: field}
	}
	rule, err := getFieldRule(fieldDef)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageFieldLocation(ctx.fileName(), ctx.messageName(), fieldDef.GetName()),
			),
		)
	}
	r.fieldToRuleMap[field] = rule
	return field
}

func (r *Resolver) resolveType(ctx *context, typeName string, typ types.Type, label descriptorpb.FieldDescriptorProto_Label) (*Type, error) {
	var (
		ref  *Message
		enum *Enum
	)
	switch typ {
	case types.Message:
		var pkg *Package
		if !strings.Contains(typeName, ".") {
			file := ctx.file()
			if file == nil {
				return nil, fmt.Errorf(`package name is missing for %q message`, typeName)
			}
			pkg = file.Package
		} else {
			p, err := r.lookupPackage(typeName)
			if err != nil {
				return nil, err
			}
			pkg = p
		}
		name := r.trimPackage(pkg, typeName)
		ref = r.resolveMessage(ctx, pkg, name)
	case types.Enum:
		var pkg *Package
		if !strings.Contains(typeName, ".") {
			file := ctx.file()
			if file == nil {
				return nil, fmt.Errorf(`package name is missing for %q enum`, typeName)
			}
			pkg = file.Package
		} else {
			p, err := r.lookupPackage(typeName)
			if err != nil {
				return nil, err
			}
			pkg = p
		}
		name := r.trimPackage(pkg, typeName)
		enum = r.resolveEnum(ctx, pkg, name)
	}
	return &Type{
		Type:     typ,
		Repeated: label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED,
		Ref:      ref,
		Enum:     enum,
	}, nil
}

func (r *Resolver) resolveMethodCall(ctx *context, resolverDef *federation.Resolver) *MethodCall {
	if resolverDef == nil {
		return nil
	}
	pkgName, serviceName, methodName, err := r.splitMethodFullName(ctx.file().Package, resolverDef.GetMethod())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MethodLocation(ctx.fileName(), ctx.messageName()),
			),
		)
		return nil
	}
	pkg, exists := r.protoPackageNameToPackage[pkgName]
	if !exists {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q package does not exist`, pkgName),
				source.MethodLocation(ctx.fileName(), ctx.messageName()),
			),
		)
		return nil
	}
	service := r.resolveService(ctx, pkg, serviceName)
	if service == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`cannot find %q method because the service to which the method belongs does not exist`, methodName),
				source.MethodLocation(ctx.fileName(), ctx.messageName()),
			),
		)
		return nil
	}

	method := service.Method(methodName)
	if method == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q method does not exist in %s service`, methodName, service.Name),
				source.MethodLocation(ctx.fileName(), ctx.messageName()),
			),
		)
		return nil
	}

	var timeout *time.Duration
	timeoutDef := resolverDef.GetTimeout()
	if timeoutDef != "" {
		duration, err := time.ParseDuration(timeoutDef)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MethodTimeoutLocation(ctx.fileName(), ctx.messageName()),
				),
			)
		} else {
			timeout = &duration
		}
	}

	return &MethodCall{
		Method:   method,
		Request:  r.resolveRequest(ctx, method, resolverDef.GetRequest()),
		Response: r.resolveResponse(ctx, method, resolverDef.GetResponse()),
		Timeout:  timeout,
		Retry:    r.resolveRetry(ctx, resolverDef.GetRetry(), timeout),
	}
}

func (r *Resolver) resolveRetry(ctx *context, def *federation.RetryPolicy, timeout *time.Duration) *RetryPolicy {
	if def == nil {
		return nil
	}
	return &RetryPolicy{
		Constant:    r.resolveRetryConstant(ctx, def.GetConstant()),
		Exponential: r.resolveRetryExponential(ctx, def.GetExponential(), timeout),
	}
}

var (
	DefaultRetryMaxRetryCount = uint64(5)

	DefaultRetryConstantInterval = time.Second

	DefaultRetryExponentialInitialInterval     = 500 * time.Millisecond
	DefaultRetryExponentialRandomizationFactor = float64(0.5)
	DefaultRetryExponentialMultiplier          = float64(1.5)
	DefaultRetryExponentialMaxInterval         = 60 * time.Second
)

func (r *Resolver) resolveRetryConstant(ctx *context, def *federation.RetryPolicyConstant) *RetryPolicyConstant {
	if def == nil {
		return nil
	}

	interval := DefaultRetryConstantInterval
	if def.Interval != nil {
		duration, err := time.ParseDuration(def.GetInterval())
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MethodRetryConstantIntervalLocation(ctx.fileName(), ctx.messageName()),
				),
			)
		} else {
			interval = duration
		}
	}

	maxRetries := DefaultRetryMaxRetryCount
	if def.MaxRetries != nil {
		maxRetries = def.GetMaxRetries()
	}

	return &RetryPolicyConstant{
		Interval:   interval,
		MaxRetries: maxRetries,
	}
}

func (r *Resolver) resolveRetryExponential(ctx *context, def *federation.RetryPolicyExponential, timeout *time.Duration) *RetryPolicyExponential {
	if def == nil {
		return nil
	}

	initialInterval := DefaultRetryExponentialInitialInterval
	if def.InitialInterval != nil {
		interval, err := time.ParseDuration(def.GetInitialInterval())
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MethodRetryExponentialInitialIntervalLocation(ctx.fileName(), ctx.messageName()),
				),
			)
		} else {
			initialInterval = interval
		}
	}

	randomizationFactor := DefaultRetryExponentialRandomizationFactor
	if def.RandomizationFactor != nil {
		randomizationFactor = def.GetRandomizationFactor()
	}

	multiplier := DefaultRetryExponentialMultiplier
	if def.Multiplier != nil {
		multiplier = def.GetMultiplier()
	}

	maxInterval := DefaultRetryExponentialMaxInterval
	if def.MaxInterval != nil {
		interval, err := time.ParseDuration(def.GetMaxInterval())
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MethodRetryExponentialMaxIntervalLocation(ctx.fileName(), ctx.messageName()),
				),
			)
		} else {
			maxInterval = interval
		}
	}

	maxRetries := DefaultRetryMaxRetryCount
	if def.MaxRetries != nil {
		maxRetries = def.GetMaxRetries()
	}

	var maxElapsedTime time.Duration
	if timeout != nil {
		maxElapsedTime = *timeout
	}

	return &RetryPolicyExponential{
		InitialInterval:     initialInterval,
		RandomizationFactor: randomizationFactor,
		Multiplier:          multiplier,
		MaxInterval:         maxInterval,
		MaxRetries:          maxRetries,
		MaxElapsedTime:      maxElapsedTime,
	}
}

func (r *Resolver) resolveRequest(ctx *context, method *Method, requestDef []*federation.MethodRequest) *Request {
	reqType := method.Request
	args := make([]*Argument, 0, len(requestDef))
	for idx, req := range requestDef {
		fieldName := req.GetField()
		var argType *Type
		if !reqType.HasField(fieldName) {
			ctx.addError(ErrWithLocation(
				fmt.Sprintf(`%q field does not exist in "%s.%s" message for method request`, fieldName, reqType.PackageName(), reqType.Name),
				source.RequestFieldLocation(ctx.fileName(), ctx.messageName(), idx),
			))
		} else {
			argType = reqType.Field(fieldName).Type
		}
		value, err := r.resolveValue(ctx, methodRequestToCommonValueDef(req))
		if err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				source.RequestByLocation(ctx.fileName(), ctx.messageName(), idx),
			))
		}
		args = append(args, &Argument{
			Name:  fieldName,
			Type:  argType,
			Value: value,
		})
	}
	return &Request{Args: args, Type: reqType}
}

func (r *Resolver) resolveResponse(ctx *context, method *Method, responseDef []*federation.MethodResponse) *Response {
	resType := method.Response
	fields := make([]*ResponseField, 0, len(responseDef))
	for idx, res := range responseDef {
		name := res.GetName()
		if err := r.validateName(name); err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				source.ResponseFieldLocation(ctx.fileName(), ctx.messageName(), idx),
			))
			continue
		}
		if name == "" {
			name = "_" + r.protoFQDNToNormalizedName(method.FQDN())
		}
		fieldName := res.GetField()
		if fieldName == "" {
			fields = append(fields, &ResponseField{
				Name:     name,
				Type:     &Type{Type: types.Message, Ref: resType},
				AutoBind: res.GetAutobind(),
			})
			continue
		}
		var fieldType *Type
		if !resType.HasField(fieldName) {
			ctx.addError(ErrWithLocation(
				fmt.Sprintf(`%q field does not exist in "%s.%s" message for method response`, fieldName, resType.PackageName(), resType.Name),
				source.ResponseFieldLocation(ctx.fileName(), ctx.messageName(), idx),
			))
		} else {
			fieldType = resType.Field(fieldName).Type
		}
		fields = append(fields, &ResponseField{
			Name:      name,
			FieldName: fieldName,
			Type:      fieldType,
			AutoBind:  res.GetAutobind(),
		})
	}
	return &Response{Fields: fields, Type: resType}
}

func (r *Resolver) protoFQDNToNormalizedName(fqdn string) string {
	names := strings.Split(fqdn, ".")
	formattedNames := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.Replace(name, "-", "_", -1)
		name = strings.Replace(name, "/", "_", -1)
		formattedNames = append(formattedNames, name)
	}
	return strings.Join(formattedNames, "_")
}

func (r *Resolver) resolveMessages(ctx *context, baseMsg *Message, msgs []*federation.Message) []*MessageDependency {
	deps := make([]*MessageDependency, 0, len(msgs))
	for idx, msgDef := range msgs {
		ctx := ctx.withDepIndex(idx)
		msgName := msgDef.GetMessage()
		var msg *Message
		if strings.Contains(msgName, ".") {
			pkg, err := r.lookupPackage(msgName)
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.MessageDependencyMessageLocation(ctx.fileName(), ctx.messageName(), idx),
					),
				)
				continue
			}
			name := r.trimPackage(pkg, msgName)
			msg = r.resolveMessage(ctx, pkg, name)
		} else {
			file := ctx.file()
			msg = r.resolveMessage(ctx, file.Package, msgName)
		}
		if msg == baseMsg {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`recursive definition: %q is own message name`, msg.Name),
					source.MessageDependencyMessageLocation(ctx.fileName(), ctx.messageName(), idx),
				),
			)
			continue
		}
		args := make([]*Argument, 0, len(msgDef.GetArgs()))
		for idx, argDef := range msgDef.GetArgs() {
			args = append(args, r.resolveMessageDependencyArgument(ctx.withArgIndex(idx), argDef))
		}

		name := msgDef.GetName()
		if err := r.validateName(name); err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				source.MessageDependencyMessageLocation(ctx.fileName(), ctx.messageName(), idx),
			))
			continue
		}
		if name == "" {
			name = "_" + r.protoFQDNToNormalizedName(msg.FQDN())
		}
		deps = append(deps, &MessageDependency{
			Name:     name,
			Message:  msg,
			Args:     args,
			AutoBind: msgDef.GetAutobind(),
		})
	}
	return deps
}

func (r *Resolver) resolveMessageDependencyArgument(ctx *context, argDef *federation.Argument) *Argument {
	value, err := r.resolveValue(ctx, argumentToCommonValueDef(argDef))
	if err != nil {
		switch {
		case argDef.GetBy() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MessageDependencyArgumentByLocation(
						ctx.fileName(),
						ctx.messageName(),
						ctx.depIndex(),
						ctx.argIndex(),
					),
				),
			)
		case argDef.GetInline() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.MessageDependencyArgumentInlineLocation(
						ctx.fileName(),
						ctx.messageName(),
						ctx.depIndex(),
						ctx.argIndex(),
					),
				),
			)
		}
	}
	name := argDef.GetName()
	if err := r.validateName(name); err != nil {
		ctx.addError(ErrWithLocation(
			err.Error(),
			source.MessageDependencyArgumentNameLocation(
				ctx.fileName(),
				ctx.messageName(),
				ctx.depIndex(),
				ctx.argIndex(),
			),
		))
	}
	return &Argument{
		Name:  name,
		Value: value,
	}
}

func (r *Resolver) resolveMessageConstValue(ctx *context, val *federation.MessageValue) (*Type, map[string]*Value, error) {
	msgName := val.GetName()
	t, err := r.resolveType(
		ctx,
		msgName,
		types.Message,
		descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL,
	)
	if err != nil {
		return nil, nil, err
	}
	if t.Ref == nil {
		return nil, nil, fmt.Errorf(`%q message does not exist`, msgName)
	}
	fieldMap := map[string]*Value{}
	for _, field := range val.GetFields() {
		fieldName := field.GetField()
		if !t.Ref.HasField(fieldName) {
			return nil, nil, fmt.Errorf(`%q field does not exist in %s message`, fieldName, msgName)
		}
		value, err := r.resolveValue(ctx, messageFieldValueToCommonValueDef(field))
		if err != nil {
			return nil, nil, err
		}
		fieldMap[field.GetField()] = value
	}
	return t, fieldMap, nil
}

func (r *Resolver) resolveEnumConstValue(ctx *context, enumValueName string) (*EnumValue, error) {
	pkg, err := r.lookupPackageFromTypeName(ctx, enumValueName)
	if err != nil {
		return nil, err
	}
	return r.lookupEnumValue(enumValueName, pkg)
}

func (r *Resolver) lookupPackageFromTypeName(ctx *context, name string) (*Package, error) {
	if strings.Contains(name, ".") {
		p, err := r.lookupPackage(name)
		if err == nil {
			return p, nil
		}
	}
	file := ctx.file()
	if file == nil {
		return nil, fmt.Errorf(`cannot find package from %q name`, name)
	}
	return file.Package, nil
}

func (r *Resolver) lookupEnumValue(name string, pkg *Package) (*EnumValue, error) {
	valueName := r.trimPackage(pkg, name)
	isGlobalEnumValue := !strings.Contains(valueName, ".")
	if isGlobalEnumValue {
		for _, file := range pkg.Files {
			for _, enum := range file.Enums {
				for _, value := range enum.Values {
					if value.Value == valueName {
						return value, nil
					}
				}
			}
		}
	} else {
		for _, file := range pkg.Files {
			for _, msg := range file.Messages {
				if value := r.lookupEnumValueFromMessage(name, msg); value != nil {
					return value, nil
				}
			}
		}
	}
	return nil, fmt.Errorf(`cannot find enum value from %q`, name)
}

func (r *Resolver) lookupEnumValueFromMessage(name string, msg *Message) *EnumValue {
	msgName := strings.Join(append(msg.ParentMessageNames(), msg.Name), ".")
	for _, enum := range msg.Enums {
		for _, value := range enum.Values {
			valueName := fmt.Sprintf("%s.%s", msgName, value.Value)
			if valueName == name {
				return value
			}
		}
	}
	for _, msg := range msg.NestedMessages {
		if value := r.lookupEnumValueFromMessage(name, msg); value != nil {
			return value
		}
	}
	return nil
}

// resolveMessageArgument constructs message arguments using a dependency graph and assigns them to each message.
func (r *Resolver) resolveMessageArgument(ctx *context, files []*File) {
	// create a dependency graph for all messages.
	graph := CreateAllMessageDependencyGraph(ctx, r.allMessages(files))
	if graph == nil {
		return
	}

	svcMsgSet := make(map[*Service]map[*Message]struct{})
	for _, root := range graph.Roots {
		reqMsg := r.lookupRequestMessageFromResponseMessage(root.Message)
		if reqMsg == nil {
			continue
		}
		msgArg := newMessageArgument(root.Message)
		msgArg.Fields = append(msgArg.Fields, reqMsg.Fields...)
		r.cachedMessageMap[msgArg.FQDN()] = msgArg
		// Store the messages to serviceMsgMap first to avoid inserting duplicated ones to Service.Messages
		for _, svc := range root.Message.File.Services {
			if _, exists := svcMsgSet[svc]; !exists {
				svcMsgSet[svc] = make(map[*Message]struct{})
			}
			for _, msg := range r.resolveMessageArgumentRecursive(ctx, root) {
				svcMsgSet[svc][msg] = struct{}{}
			}
		}
	}

	for svc, msgSet := range svcMsgSet {
		msgs := make([]*Message, 0, len(msgSet))
		for msg := range msgSet {
			msgs = append(msgs, msg)
		}
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].Name < msgs[j].Name
		})
		args := make([]*Message, 0, len(msgs))
		for _, msg := range msgs {
			args = append(args, msg.Rule.MessageArgument)
		}
		svc.MessageArgs = append(svc.MessageArgs, args...)
		svc.Messages = append(svc.Messages, msgs...)
	}
}

func (r *Resolver) resolveMessageArgumentRecursive(ctx *context, node *AllMessageDependencyGraphNode) []*Message {
	msg := node.Message
	arg := msg.Rule.MessageArgument
	fileDesc := r.messageArgumentFileDescriptor(arg)
	if err := r.celRegistry.RegisterFiles(append(r.files, fileDesc)...); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageLocation(msg.File.Name, msg.Name),
			),
		)
		return nil
	}
	env, err := r.createCELEnv(msg)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.MessageLocation(msg.File.Name, msg.Name),
			),
		)
		return nil
	}
	r.resolveMessageCELValues(ctx, env, msg)

	msgs := []*Message{msg}
	msgToDepsMap := make(map[*Message][]*MessageDependency)
	depToIdx := make(map[*MessageDependency]int)
	for idx, dep := range msg.Rule.MessageDependencies {
		msgToDepsMap[dep.Message] = append(msgToDepsMap[dep.Message], dep)
		depToIdx[dep] = idx
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for idx, dep := range field.Rule.Oneof.MessageDependencies {
			msgToDepsMap[dep.Message] = append(msgToDepsMap[dep.Message], dep)
			depToIdx[dep] = idx
		}
	}
	for _, child := range node.Children {
		depMsg := child.Message
		depMsgArg := newMessageArgument(depMsg)
		r.cachedMessageMap[depMsgArg.FQDN()] = depMsgArg
		deps := msgToDepsMap[depMsg]
		depMsgArg.Fields = append(depMsgArg.Fields, r.resolveMessageArgumentFields(ctx, msg, deps, depToIdx)...)
		m := r.resolveMessageArgumentRecursive(ctx, child)
		msgs = append(msgs, m...)
	}
	return msgs
}

func (r *Resolver) resolveMessageArgumentFields(ctx *context, msg *Message, deps []*MessageDependency, depToIdx map[*MessageDependency]int) []*Field {
	argNameMap := make(map[string]struct{})
	for _, dep := range deps {
		for _, arg := range dep.Args {
			if arg.Name == "" {
				continue
			}
			argNameMap[arg.Name] = struct{}{}
		}
	}

	evaluatedArgNameMap := make(map[string]struct{})
	var fields []*Field
	for _, dep := range deps {
		depIdx := depToIdx[dep]
		r.validateMessageDependencyArgumentName(ctx, argNameMap, msg, dep, depIdx)
		for argIdx, arg := range dep.Args {
			if _, exists := evaluatedArgNameMap[arg.Name]; exists {
				continue
			}
			if arg.Value == nil {
				continue
			}
			fieldType := arg.Value.Type()
			if fieldType == nil {
				continue
			}
			if arg.Value.CEL != nil && arg.Value.CEL.Inline {
				if fieldType.Type != types.Message {
					ctx.addError(
						ErrWithLocation(
							"inline value is not message type",
							source.MessageDependencyArgumentInlineLocation(
								msg.File.Name,
								msg.Name,
								depToIdx[dep],
								argIdx,
							),
						),
					)
					continue
				}
				fields = append(fields, fieldType.Ref.Fields...)
			} else {
				fields = append(fields, &Field{
					Name: arg.Name,
					Type: fieldType,
				})
			}
			evaluatedArgNameMap[arg.Name] = struct{}{}
		}
	}
	return fields
}

func (r *Resolver) validateMessageDependencyArgumentName(ctx *context, argNameMap map[string]struct{}, msg *Message, dep *MessageDependency, depIdx int) {
	curDepArgNameMap := make(map[string]struct{})
	for _, arg := range dep.Args {
		if arg.Name == "" {
			continue
		}
		curDepArgNameMap[arg.Name] = struct{}{}
	}
	for name := range argNameMap {
		if _, exists := curDepArgNameMap[name]; exists {
			continue
		}
		var errLoc *source.Location
		switch dep.Owner.Type {
		case MessageDependencyOwnerMessage:
			errLoc = source.MessageDependencyArgumentLocation(
				msg.File.Name,
				msg.Name,
				depIdx,
				0,
			)
		case MessageDependencyOwnerOneofField:
			errLoc = source.MessageFieldOneofMessageDependencyArgumentLocation(
				msg.File.Name,
				msg.Name,
				dep.Owner.Field.Name,
				depIdx,
				0,
			)
		}
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf("%q argument is defined in other message dependency arguments, but not in this context", name),
				errLoc,
			),
		)
	}
}

func (r *Resolver) resolveMessageCELValues(ctx *context, env *cel.Env, msg *Message) {
	if msg.Rule == nil {
		return
	}
	methodCall := msg.Rule.MethodCall
	if methodCall != nil && methodCall.Request != nil {
		for idx, arg := range methodCall.Request.Args {
			if arg.Value == nil {
				continue
			}
			if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.RequestByLocation(msg.File.Name, msg.Name, idx),
					),
				)
			}
		}
	}
	for depIdx, depMessage := range msg.Rule.MessageDependencies {
		for argIdx, arg := range depMessage.Args {
			if arg.Value == nil {
				continue
			}
			if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
				if arg.Value.CEL != nil && arg.Value.CEL.Inline {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.MessageDependencyArgumentInlineLocation(
								msg.File.Name,
								msg.Name,
								depIdx,
								argIdx,
							),
						),
					)
				} else {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.MessageDependencyArgumentByLocation(
								msg.File.Name,
								msg.Name,
								depIdx,
								argIdx,
							),
						),
					)
				}
			}
		}
	}
	for _, field := range msg.Fields {
		if !field.HasRule() {
			continue
		}
		if field.Rule.Value != nil {
			if err := r.resolveCELValue(ctx, env, field.Rule.Value.CEL); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.MessageFieldByLocation(
							msg.File.Name,
							msg.Name,
							field.Name,
						),
					),
				)
			}
		}
		if field.Rule.Oneof != nil {
			oneof := field.Rule.Oneof
			if err := r.resolveCELValue(ctx, env, oneof.Expr); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.MessageFieldOneofExprLocation(
							msg.File.Name,
							msg.Name,
							field.Name,
						),
					),
				)
			}
			if oneof.Expr.Out != nil {
				if oneof.Expr.Out.Type != types.Bool {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`return value of "expr" must be bool type but got %s type`, oneof.Expr.Out.Type),
							source.MessageFieldOneofExprLocation(
								msg.File.Name,
								msg.Name,
								field.Name,
							),
						),
					)
				}
			}
			var envOpts []cel.EnvOption
			for _, dep := range oneof.MessageDependencies {
				if dep.Name == "" {
					continue
				}
				envOpts = append(envOpts, cel.Variable(dep.Name, ToCELType(NewMessageType(dep.Message, false))))
			}
			newEnv, err := env.Extend(envOpts...)
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`failed to extend cel.Env from variables of messages: %s`, err.Error()),
						source.MessageFieldOneofLocation(
							msg.File.Name,
							msg.Name,
							field.Name,
						),
					),
				)
				continue
			}
			if err := r.resolveCELValue(ctx, newEnv, oneof.By); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.MessageFieldOneofByLocation(
							msg.File.Name,
							msg.Name,
							field.Name,
						),
					),
				)
			}
		}
	}

	r.resolveUsedNameReference(msg)
}

func (r *Resolver) resolveUsedNameReference(msg *Message) {
	if msg.Rule == nil {
		return
	}
	methodCall := msg.Rule.MethodCall
	nameMap := make(map[string]struct{})
	for _, name := range msg.ReferenceNames() {
		nameMap[name] = struct{}{}
	}
	if methodCall != nil && methodCall.Response != nil {
		for _, field := range methodCall.Response.Fields {
			if _, exists := nameMap[field.Name]; exists {
				field.Used = true
			}
		}
	}
	for _, depMessage := range msg.Rule.MessageDependencies {
		if _, exists := nameMap[depMessage.Name]; exists {
			depMessage.Used = true
		}
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		oneof := field.Rule.Oneof
		oneofNameMap := make(map[string]struct{})
		for _, name := range oneof.By.ReferenceNames() {
			oneofNameMap[name] = struct{}{}
		}
		for _, depMessage := range oneof.MessageDependencies {
			if _, exists := oneofNameMap[depMessage.Name]; exists {
				depMessage.Used = true
			}
		}
	}
}

func (r *Resolver) resolveCELValue(ctx *context, env *cel.Env, value *CELValue) error {
	if value == nil {
		return nil
	}
	if strings.Contains(value.Expr, federation.MessageArgumentVariableName) {
		return fmt.Errorf("%q is a reserved keyword and cannot be used as a variable name", federation.MessageArgumentVariableName)
	}
	expr := strings.Replace(value.Expr, "$", federation.MessageArgumentVariableName, -1)
	ast, issues := env.Compile(expr)
	if issues.Err() != nil {
		return issues.Err()
	}
	out, err := r.fromCELType(ctx, ast.OutputType())
	if err != nil {
		return err
	}
	checkedExpr, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		return err
	}
	value.Out = out
	value.CheckedExpr = checkedExpr
	return nil
}

func (r *Resolver) createCELEnv(msg *Message) (*cel.Env, error) {
	arg := msg.Rule.MessageArgument
	envOpts := []cel.EnvOption{
		cel.StdLib(),
		cel.CustomTypeAdapter(r.celRegistry),
		cel.CustomTypeProvider(r.celRegistry),
		cel.Variable(federation.MessageArgumentVariableName, cel.ObjectType(arg.FQDN())),
	}
	if msg.Rule != nil && msg.Rule.MethodCall != nil && msg.Rule.MethodCall.Response != nil {
		resType := msg.Rule.MethodCall.Response.Type
		for _, responseField := range msg.Rule.MethodCall.Response.Fields {
			if responseField.Name == "" {
				continue
			}
			typ := NewMessageType(resType, false)
			if responseField.FieldName != "" {
				field := resType.Field(responseField.FieldName)
				if field == nil {
					// The case of an invalid field in the response field has already been verified in the `resolveResponse()`,
					// so there is no need to add the error to context here.
					continue
				}
				typ = field.Type
			}
			envOpts = append(envOpts, cel.Variable(responseField.Name, ToCELType(typ)))
		}
	}
	if msg.Rule != nil {
		for _, dep := range msg.Rule.MessageDependencies {
			if dep.Name == "" {
				continue
			}
			envOpts = append(envOpts, cel.Variable(dep.Name, ToCELType(NewMessageType(dep.Message, false))))
		}
	}
	env, err := cel.NewCustomEnv(envOpts...)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (r *Resolver) fromCELType(ctx *context, typ *cel.Type) (*Type, error) {
	switch typ.Kind() {
	case celtypes.BoolKind:
		return BoolType, nil
	case celtypes.BytesKind:
		return BytesType, nil
	case celtypes.DoubleKind:
		return DoubleType, nil
	case celtypes.IntKind:
		if enum, found := r.celRegistry.LookupEnum(typ); found {
			return &Type{Type: types.Enum, Enum: enum}, nil
		}
		return Int64Type, nil
	case celtypes.UintKind:
		return Uint64Type, nil
	case celtypes.StringKind:
		return StringType, nil
	case celtypes.AnyKind:
		return AnyType, nil
	case celtypes.DurationKind:
		return DurationType, nil
	case celtypes.TimestampKind:
		return TimestampType, nil
	case celtypes.MapKind:
		mapKey, err := r.fromCELType(ctx, typ.Parameters()[0])
		if err != nil {
			return nil, err
		}
		mapValue, err := r.fromCELType(ctx, typ.Parameters()[1])
		if err != nil {
			return nil, err
		}
		return NewMessageType(&Message{
			IsMapEntry: true,
			Fields: []*Field{
				{Name: "key", Type: mapKey},
				{Name: "value", Type: mapValue},
			},
		}, false), nil
	case celtypes.ListKind:
		typ, err := r.fromCELType(ctx, typ.Parameters()[0])
		if err != nil {
			return nil, err
		}
		typ = typ.Clone()
		typ.Repeated = true
		return typ, nil
	case celtypes.StructKind:
		return r.resolveType(
			ctx,
			typ.TypeName(),
			types.Message,
			descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL,
		)
	}
	return nil, errors.New("unknown type is required")
}

func (r *Resolver) messageArgumentFileDescriptor(arg *Message) *descriptorpb.FileDescriptorProto {
	desc := arg.File.Desc
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String(arg.Name),
	}
	for idx, field := range arg.Fields {
		var typeName string
		if field.Type.Ref != nil {
			typeName = field.Type.Ref.FQDN()
		}
		if field.Type.Enum != nil {
			typeName = field.Type.Enum.FQDN()
		}
		label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
		if field.Type.Repeated {
			label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		}
		typ := field.Type.Type
		msg.Field = append(msg.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String(field.Name),
			Number:   proto.Int32(int32(idx) + 1),
			Type:     &typ,
			TypeName: proto.String(typeName),
			Label:    &label,
		})
	}
	return &descriptorpb.FileDescriptorProto{
		Name:             proto.String(arg.Name),
		Package:          proto.String(federation.PrivatePackageName),
		Dependency:       append(desc.Dependency, arg.File.Name),
		PublicDependency: desc.PublicDependency,
		WeakDependency:   desc.WeakDependency,
		MessageType:      []*descriptorpb.DescriptorProto{msg},
	}
}

// resolveMessageRuleDependencies resolve dependencies for each message.
func (r *Resolver) resolveMessageDependencies(ctx *context, files []*File) {
	msgs := r.allMessages(files)
	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}
		if graph := CreateMessageDependencyGraph(ctx, msg); graph != nil {
			msg.Rule.DependencyGraph = graph
			msg.Rule.Resolvers = graph.MessageResolverGroups(ctx)
		}
		for _, field := range msg.Fields {
			if field.Rule == nil {
				continue
			}
			if field.Rule.Oneof == nil {
				continue
			}
			if graph := CreateMessageDependencyGraphByFieldOneof(ctx, msg, field); graph != nil {
				field.Rule.Oneof.DependencyGraph = graph
				field.Rule.Oneof.Resolvers = graph.MessageResolverGroups(ctx)
			}
		}
	}
	r.validateMessages(ctx, msgs)
}

func (r *Resolver) resolveValue(ctx *context, def *commonValueDef) (*Value, error) {
	const (
		customResolverOpt = "custom_resolver"
		aliasOpt          = "alias"
		byOpt             = "by"
		inlineOpt         = "inline"
		doubleOpt         = "double"
		doublesOpt        = "doubles"
		floatOpt          = "float"
		floatsOpt         = "floats"
		int32Opt          = "int32"
		int32sOpt         = "int32s"
		int64Opt          = "int64"
		int64sOpt         = "int64s"
		uint32Opt         = "uint32"
		uint32sOpt        = "uint32s"
		uint64Opt         = "uint64"
		uint64sOpt        = "uint64s"
		sint32Opt         = "sint32"
		sint32sOpt        = "sint32s"
		sint64Opt         = "sint64"
		sint64sOpt        = "sint64s"
		fixed32Opt        = "fixed32"
		fixed32sOpt       = "fixed32s"
		fixed64Opt        = "fixed64"
		fixed64sOpt       = "fixed64s"
		sfixed32Opt       = "sfixed32"
		sfixed32sOpt      = "sfixed32s"
		sfixed64Opt       = "sfixed64"
		sfixed64sOpt      = "sfixed64s"
		boolOpt           = "bool"
		boolsOpt          = "bools"
		stringOpt         = "string"
		stringsOpt        = "strings"
		byteStringOpt     = "byte_string"
		byteStringsOpt    = "byte_strings"
		messageOpt        = "message"
		messagesOpt       = "messages"
		enumOpt           = "enum"
		enumsOpt          = "enums"
		envOpt            = "env"
		envsOpt           = "envs"
	)
	var (
		value    *Value
		optNames []string
	)
	if def.CustomResolver != nil {
		optNames = append(optNames, customResolverOpt)
	}
	if def.Alias != nil {
		optNames = append(optNames, aliasOpt)
	}
	if def.By != nil {
		value = &Value{CEL: &CELValue{Expr: def.GetBy()}}
		optNames = append(optNames, byOpt)
	}
	if def.Inline != nil {
		value = &Value{CEL: &CELValue{Expr: def.GetInline(), Inline: true}}
		optNames = append(optNames, inlineOpt)
	}
	if def.Double != nil {
		value = NewDoubleValue(def.GetDouble())
		optNames = append(optNames, doubleOpt)
	}
	if def.Doubles != nil {
		value = NewDoublesValue(def.GetDoubles()...)
		optNames = append(optNames, doublesOpt)
	}
	if def.Float != nil {
		value = NewFloatValue(def.GetFloat())
		optNames = append(optNames, floatOpt)
	}
	if def.Floats != nil {
		value = NewFloatsValue(def.GetFloats()...)
		optNames = append(optNames, floatsOpt)
	}
	if def.Int32 != nil {
		value = NewInt32Value(def.GetInt32())
		optNames = append(optNames, int32Opt)
	}
	if def.Int32S != nil {
		value = NewInt32sValue(def.GetInt32S()...)
		optNames = append(optNames, int32sOpt)
	}
	if def.Int64 != nil {
		value = NewInt64Value(def.GetInt64())
		optNames = append(optNames, int64Opt)
	}
	if def.Int64S != nil {
		value = NewInt64sValue(def.GetInt64S()...)
		optNames = append(optNames, int64sOpt)
	}
	if def.Uint32 != nil {
		value = NewUint32Value(def.GetUint32())
		optNames = append(optNames, uint32Opt)
	}
	if def.Uint32S != nil {
		value = NewUint32sValue(def.GetUint32S()...)
		optNames = append(optNames, uint32sOpt)
	}
	if def.Uint64 != nil {
		value = NewUint64Value(def.GetUint64())
		optNames = append(optNames, uint64Opt)
	}
	if def.Uint64S != nil {
		value = NewUint64sValue(def.GetUint64S()...)
		optNames = append(optNames, uint64sOpt)
	}
	if def.Sint32 != nil {
		value = NewSint32Value(def.GetSint32())
		optNames = append(optNames, sint32Opt)
	}
	if def.Sint32S != nil {
		value = NewSint32sValue(def.GetSint32S()...)
		optNames = append(optNames, sint32sOpt)
	}
	if def.Sint64 != nil {
		value = NewSint64Value(def.GetSint64())
		optNames = append(optNames, sint64Opt)
	}
	if def.Sint64S != nil {
		value = NewSint64sValue(def.GetSint64S()...)
		optNames = append(optNames, sint64sOpt)
	}
	if def.Fixed32 != nil {
		value = NewFixed32Value(def.GetFixed32())
		optNames = append(optNames, fixed32Opt)
	}
	if def.Fixed32S != nil {
		value = NewFixed32sValue(def.GetFixed32S()...)
		optNames = append(optNames, fixed32sOpt)
	}
	if def.Fixed64 != nil {
		value = NewFixed64Value(def.GetFixed64())
		optNames = append(optNames, fixed64Opt)
	}
	if def.Fixed64S != nil {
		value = NewFixed64sValue(def.GetFixed64S()...)
		optNames = append(optNames, fixed64sOpt)
	}
	if def.Sfixed32 != nil {
		value = NewSfixed32Value(def.GetSfixed32())
		optNames = append(optNames, sfixed32Opt)
	}
	if def.Sfixed32S != nil {
		value = NewSfixed32sValue(def.GetSfixed32S()...)
		optNames = append(optNames, sfixed32sOpt)
	}
	if def.Sfixed64 != nil {
		value = NewSfixed64Value(def.GetSfixed64())
		optNames = append(optNames, sfixed64Opt)
	}
	if def.Sfixed64S != nil {
		value = NewSfixed64sValue(def.GetSfixed64S()...)
		optNames = append(optNames, sfixed64sOpt)
	}
	if def.Bool != nil {
		value = NewBoolValue(def.GetBool())
		optNames = append(optNames, boolOpt)
	}
	if def.Bools != nil {
		value = NewBoolsValue(def.GetBools()...)
		optNames = append(optNames, boolsOpt)
	}
	if def.String != nil {
		value = NewStringValue(def.GetString())
		optNames = append(optNames, stringOpt)
	}
	if def.Strings != nil {
		value = NewStringsValue(def.GetStrings()...)
		optNames = append(optNames, stringsOpt)
	}
	if def.ByteString != nil {
		value = NewByteStringValue(def.GetByteString())
		optNames = append(optNames, byteStringOpt)
	}
	if def.ByteStrings != nil {
		value = NewByteStringsValue(def.GetByteStrings()...)
		optNames = append(optNames, byteStringsOpt)
	}
	if def.Message != nil {
		typ, val, err := r.resolveMessageConstValue(ctx, def.GetMessage())
		if err != nil {
			return nil, err
		}
		value = NewMessageValue(typ, val)
		optNames = append(optNames, messageOpt)
	}
	if def.Messages != nil {
		var (
			typ  *Type
			vals []map[string]*Value
		)
		for _, msg := range def.GetMessages() {
			t, val, err := r.resolveMessageConstValue(ctx, msg)
			if err != nil {
				return nil, err
			}
			if typ == nil {
				typ = t
			} else if typ.Ref != t.Ref {
				return nil, fmt.Errorf(`"messages" value unsupported multiple message type`)
			}
			vals = append(vals, val)
		}
		typ.Repeated = true
		value = NewMessagesValue(typ, vals...)
		optNames = append(optNames, messagesOpt)
	}
	if def.Enum != nil {
		enumValue, err := r.resolveEnumConstValue(ctx, def.GetEnum())
		if err != nil {
			return nil, err
		}
		value = NewEnumValue(enumValue)
		optNames = append(optNames, enumOpt)
	}
	if def.Enums != nil {
		var (
			enumValues []*EnumValue
			enum       *Enum
		)
		for _, enumName := range def.GetEnums() {
			enumValue, err := r.resolveEnumConstValue(ctx, enumName)
			if err != nil {
				return nil, err
			}
			if enum == nil {
				enum = enumValue.Enum
			} else if enum != enumValue.Enum {
				return nil, fmt.Errorf(`different enum values are used in enums: %q and %q`, enum.FQDN(), enumValue.Enum.FQDN())
			}
			enumValues = append(enumValues, enumValue)
		}
		value = NewEnumsValue(enumValues...)
		optNames = append(optNames, enumsOpt)
	}
	if def.Env != nil {
		value = NewEnvValue(EnvKey(def.GetEnv()))
		optNames = append(optNames, envOpt)
	}
	if def.Envs != nil {
		envKeys := make([]EnvKey, 0, len(def.GetEnvs()))
		for _, key := range def.GetEnvs() {
			envKeys = append(envKeys, EnvKey(key))
		}
		value = NewEnvsValue(envKeys...)
		optNames = append(optNames, envsOpt)
	}
	if len(optNames) == 0 {
		return nil, fmt.Errorf("value must be specified")
	}
	if len(optNames) != 1 {
		return nil, fmt.Errorf("multiple values cannot be specified at the same time: %s", strings.Join(optNames, ","))
	}
	return value, nil
}

func (r *Resolver) lookupRequestMessageFromResponseMessage(resMsg *Message) *Message {
	for _, method := range r.cachedMethodMap {
		if method.Response == resMsg {
			return method.Request
		}
	}
	return nil
}

func (r *Resolver) splitMethodFullName(pkg *Package, name string) (string, string, string, error) {
	serviceWithPkgAndMethod := strings.Split(name, "/")
	if len(serviceWithPkgAndMethod) != 2 {
		return "", "", "", fmt.Errorf(`invalid method format. required format is "<package-name>.<service-name>/<method-name>" but specified %q`, name)
	}
	serviceWithPkgName := serviceWithPkgAndMethod[0]
	methodName := serviceWithPkgAndMethod[1]
	if !strings.Contains(serviceWithPkgName, ".") {
		return pkg.Name, serviceWithPkgName, methodName, nil
	}
	names := strings.Split(serviceWithPkgName, ".")
	if len(names) <= 1 {
		return "", "", "", fmt.Errorf(`invalid method format. required package name but not specified: %q`, serviceWithPkgName)
	}
	pkgName := strings.Join(names[:len(names)-1], ".")
	serviceName := names[len(names)-1]
	return pkgName, serviceName, methodName, nil
}

func (r *Resolver) lookupMessage(pkg *Package, name string) (*File, *descriptorpb.DescriptorProto, error) {
	files, exists := r.protoPackageNameToFileDefs[pkg.Name]
	if !exists {
		return nil, nil, fmt.Errorf(`%q package does not exist`, pkg.Name)
	}
	for _, file := range files {
		for _, msg := range file.GetMessageType() {
			if msg.GetName() == name {
				return r.defToFileMap[file], msg, nil
			}
			parent := msg.GetName()
			for _, msg := range msg.GetNestedType() {
				if found := r.lookupMessageRecursive(name, parent, msg); found != nil {
					return r.defToFileMap[file], found, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf(`"%s.%s" message does not exist`, pkg.Name, name)
}

func (r *Resolver) lookupEnum(pkg *Package, name string) (*File, *descriptorpb.EnumDescriptorProto, error) {
	files, exists := r.protoPackageNameToFileDefs[pkg.Name]
	if !exists {
		return nil, nil, fmt.Errorf(`%q package does not exist`, pkg.Name)
	}
	for _, file := range files {
		for _, enum := range file.GetEnumType() {
			if enum.GetName() == name {
				return r.defToFileMap[file], enum, nil
			}
		}
		for _, msg := range file.GetMessageType() {
			msgName := msg.GetName()
			for _, enum := range msg.GetEnumType() {
				enumName := fmt.Sprintf("%s.%s", msgName, enum.GetName())
				if enumName == name {
					return r.defToFileMap[file], enum, nil
				}
			}
			for _, subMsg := range msg.GetNestedType() {
				if found := r.lookupEnumRecursive(name, msgName, subMsg); found != nil {
					return r.defToFileMap[file], found, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf(`"%s.%s" enum does not exist`, pkg.Name, name)
}

func (r *Resolver) lookupEnumRecursive(name, parent string, msg *descriptorpb.DescriptorProto) *descriptorpb.EnumDescriptorProto {
	prefix := fmt.Sprintf("%s.%s", parent, msg.GetName())
	for _, enum := range msg.GetEnumType() {
		enumName := fmt.Sprintf("%s.%s", prefix, enum.GetName())
		if enumName == name {
			return enum
		}
	}
	for _, subMsg := range msg.GetNestedType() {
		enum := r.lookupEnumRecursive(name, prefix, subMsg)
		if enum != nil {
			return enum
		}
	}
	return nil
}

func (r *Resolver) lookupMessageRecursive(name, parent string, msg *descriptorpb.DescriptorProto) *descriptorpb.DescriptorProto {
	fullMsgName := fmt.Sprintf("%s.%s", parent, msg.GetName())
	if fullMsgName == name {
		return msg
	}
	for _, nestedMsg := range msg.GetNestedType() {
		msg := r.lookupMessageRecursive(name, fullMsgName, nestedMsg)
		if msg != nil {
			return msg
		}
	}
	return nil
}

func (r *Resolver) lookupService(pkg *Package, name string) (*File, *descriptorpb.ServiceDescriptorProto, error) {
	files, exists := r.protoPackageNameToFileDefs[pkg.Name]
	if !exists {
		return nil, nil, fmt.Errorf(`%q package does not exist`, pkg.Name)
	}
	for _, file := range files {
		for _, svc := range file.GetService() {
			if svc.GetName() == name {
				return r.defToFileMap[file], svc, nil
			}
		}
	}
	return nil, nil, fmt.Errorf(`"%s.%s" service does not exist`, pkg.Name, name)
}

func (r *Resolver) lookupPackage(name string) (*Package, error) {
	name = strings.TrimPrefix(name, ".")
	names := strings.Split(name, ".")
	if len(names) <= 1 {
		return nil, fmt.Errorf(`unexpected package name %q`, name)
	}
	for lastIdx := len(names) - 1; lastIdx > 0; lastIdx-- {
		pkgName := strings.Join(names[:lastIdx], ".")
		pkg, exists := r.protoPackageNameToPackage[pkgName]
		if exists {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf(`cannot find package from %q`, name)
}

func (r *Resolver) trimPackage(pkg *Package, name string) string {
	name = strings.TrimPrefix(name, ".")
	if !strings.Contains(name, ".") {
		return name
	}
	return strings.TrimPrefix(name, fmt.Sprintf("%s.", pkg.Name))
}

func splitGoPackageName(goPackage string) (string, string, error) {
	importPathAndPkgName := strings.Split(goPackage, ";")
	if len(importPathAndPkgName) == 1 {
		path := importPathAndPkgName[0]
		paths := strings.Split(path, "/")
		if len(paths) == 0 {
			return path, path, nil
		}
		return path, paths[len(paths)-1], nil
	}
	if len(importPathAndPkgName) != 2 {
		return "", "", fmt.Errorf(`go_package option %q is invalid`, goPackage)
	}
	return importPathAndPkgName[0], importPathAndPkgName[1], nil
}
