package resolver

import (
	gocontext "context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/grpc/federation"
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/types"
)

type Resolver struct {
	files                      []*descriptorpb.FileDescriptorProto
	celRegistry                *CELRegistry
	defToFileMap               map[*descriptorpb.FileDescriptorProto]*File
	protoPackageNameToFileDefs map[string][]*descriptorpb.FileDescriptorProto
	protoPackageNameToPackage  map[string]*Package
	celPluginMap               map[string]*CELPlugin

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
		celPluginMap:               make(map[string]*CELPlugin),

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
	// Files list of files with services with the grpc.federation.service option.
	Files []*File
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
	r.resolveAutoBind(ctx, files)
	r.resolveMessageDependencies(ctx, files)

	r.validateServiceFromFiles(ctx, files)

	return &Result{
		Files:    r.resultFiles(files),
		Warnings: ctx.warnings(),
	}, ctx.error()
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
					source.NewLocationBuilder(fileDef.GetName()).WithGoPackage().Location(),
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

func (r *Resolver) validateServiceFromFiles(ctx *context, files []*File) {
	for _, file := range files {
		ctx := ctx.withFile(file)
		for _, svc := range file.Services {
			ctx := ctx.withService(svc)
			r.validateService(ctx, svc)
		}
	}
}

func (r *Resolver) resultFiles(allFiles []*File) []*File {
	fileMap := make(map[*File]struct{})
	ret := make([]*File, 0, len(allFiles))
	for _, file := range r.hasServiceOrPluginRuleFiles(allFiles) {
		ret = append(ret, file)
		fileMap[file] = struct{}{}

		for _, samePkgFile := range r.samePackageFiles(file) {
			if _, exists := fileMap[samePkgFile]; exists {
				continue
			}
			ret = append(ret, samePkgFile)
			fileMap[samePkgFile] = struct{}{}
		}
	}
	return ret
}

func (r *Resolver) hasServiceOrPluginRuleFiles(files []*File) []*File {
	var ret []*File
	for _, file := range files {
		switch {
		case file.HasServiceWithRule():
			ret = append(ret, file)
		case len(file.CELPlugins) != 0:
			ret = append(ret, file)
		}
	}
	return ret
}

func (r *Resolver) samePackageFiles(src *File) []*File {
	ret := make([]*File, 0, len(src.Package.Files))
	for _, file := range src.Package.Files {
		if file == src {
			continue
		}
		ret = append(ret, file)
	}
	return ret
}

func (r *Resolver) validateService(ctx *context, svc *Service) {
	if svc.Rule == nil {
		return
	}
	r.validateServiceDependency(ctx, svc)
	r.validateMethodResponse(ctx, svc)
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
				Location: source.NewLocationBuilder(service.File.Name).
					WithService(service.Name).
					WithOption().WithDependencies(idx).Location(),
				Message: fmt.Sprintf(`%q defined in "dependencies" of "grpc.federation.service" but it is not used`, depSvcName),
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
					source.NewLocationBuilder(ctx.fileName()).WithMessage(response.Name).Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveFile(ctx *context, def *descriptorpb.FileDescriptorProto) *File {
	file := r.defToFileMap[def]
	ctx = ctx.withFile(file)
	pluginRuleDef, err := getExtensionRule[*federation.PluginRule](def.GetOptions(), federation.E_Plugin)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).Location(),
			),
		)
	}
	if pluginRuleDef != nil {
		if plugin := r.resolveCELPlugin(ctx, def, pluginRuleDef.Export); plugin != nil {
			file.CELPlugins = append(file.CELPlugins, plugin)
		}
	}

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

func (r *Resolver) resolveCELPlugin(ctx *context, fileDef *descriptorpb.FileDescriptorProto, def *federation.Export) *CELPlugin {
	if def == nil {
		return nil
	}
	pkgName := fileDef.GetPackage()
	plugin := &CELPlugin{Name: def.GetName()}
	ctx = ctx.withPlugin(plugin)
	for idx, fn := range def.GetFunctions() {
		pluginFunc := r.resolvePluginGlobalFunction(ctx.withPluginFunctionIndex(idx), pkgName, fn)
		if pluginFunc == nil {
			continue
		}
		plugin.Functions = append(plugin.Functions, pluginFunc)
	}
	for idx, msgType := range def.GetTypes() {
		ctx := ctx.withPluginTypeIndex(idx)
		msg, err := r.resolveMessageByName(ctx, msgType.GetName())
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewLocationBuilder(ctx.fileName()).
						WithExport(plugin.Name).
						WithTypes(idx).WithName().
						Location(),
				),
			)
			continue
		}
		for idx, fn := range msgType.GetMethods() {
			pluginFunc := r.resolvePluginMethod(ctx.withPluginIsMethod(true).withPluginFunctionIndex(idx), msg, fn)
			if pluginFunc == nil {
				continue
			}
			plugin.Functions = append(plugin.Functions, pluginFunc)
		}
	}
	r.celPluginMap[def.GetName()] = plugin
	return plugin
}

func (r *Resolver) resolvePluginMethod(ctx *context, msg *Message, fn *federation.CELFunction) *CELFunction {
	msgType := NewMessageType(msg, false)
	pluginFunc := &CELFunction{
		Name:     fn.GetName(),
		Receiver: msg,
		Args:     []*Type{msgType},
	}
	args, ret := r.resolvePluginFunctionArgumentsAndReturn(ctx, fn.GetArgs(), fn.GetReturn())
	pluginFunc.Args = append(pluginFunc.Args, args...)
	pluginFunc.Return = ret
	pluginFunc.ID = r.toPluginFunctionID(fmt.Sprintf("%s_%s", msg.FQDN(), fn.GetName()), append(pluginFunc.Args, pluginFunc.Return))
	return pluginFunc
}

func (r *Resolver) resolvePluginGlobalFunction(ctx *context, pkgName string, fn *federation.CELFunction) *CELFunction {
	pluginFunc := &CELFunction{
		Name: fmt.Sprintf("%s.%s", pkgName, fn.GetName()),
	}
	args, ret := r.resolvePluginFunctionArgumentsAndReturn(ctx, fn.GetArgs(), fn.GetReturn())
	pluginFunc.Args = append(pluginFunc.Args, args...)
	pluginFunc.Return = ret
	pluginFunc.ID = r.toPluginFunctionID(fmt.Sprintf("%s_%s", pkgName, fn.GetName()), append(pluginFunc.Args, pluginFunc.Return))
	return pluginFunc
}

func (r *Resolver) resolvePluginFunctionArgumentsAndReturn(ctx *context, args []*federation.CELFunctionArgument, ret *federation.CELType) ([]*Type, *Type) {
	argTypes := r.resolvePluginFunctionArguments(ctx, args)
	if ret == nil {
		return argTypes, nil
	}
	retType, err := r.resolvePluginFunctionType(ctx, ret.GetType(), ret.GetRepeated())
	if err != nil {
		if ctx.pluginIsMethod() {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewLocationBuilder(ctx.fileName()).
						WithExport(ctx.pluginName()).
						WithTypes(ctx.pluginTypeIndex()).
						WithMethods(ctx.pluginFunctionIndex()).WithReturnType().Location(),
				),
			)
		} else {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewLocationBuilder(ctx.fileName()).
						WithExport(ctx.pluginName()).
						WithFunctions(ctx.pluginFunctionIndex()).WithReturnType().Location(),
				),
			)
		}
	}
	return argTypes, retType
}

func (r *Resolver) resolvePluginFunctionArguments(ctx *context, args []*federation.CELFunctionArgument) []*Type {
	var ret []*Type
	for argIdx, arg := range args {
		typ, err := r.resolvePluginFunctionType(ctx, arg.GetType(), arg.GetRepeated())
		if err != nil {
			if ctx.pluginIsMethod() {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewLocationBuilder(ctx.fileName()).
							WithExport(ctx.pluginName()).
							WithTypes(ctx.pluginTypeIndex()).
							WithMethods(ctx.pluginFunctionIndex()).
							WithArgs(argIdx).Location(),
					),
				)
			} else {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewLocationBuilder(ctx.fileName()).
							WithExport(ctx.pluginName()).
							WithFunctions(ctx.pluginFunctionIndex()).
							WithArgs(argIdx).Location(),
					),
				)
			}
			continue
		}
		ret = append(ret, typ)
	}
	return ret
}

func (r *Resolver) resolvePluginFunctionType(ctx *context, typeName string, repeated bool) (*Type, error) {
	var label descriptorpb.FieldDescriptorProto_Label
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	} else {
		label = descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
	}
	kind := types.ToKind(typeName)
	if kind == types.Unknown {
		// TODO: do not consider enums at first.
		kind = types.Message
	}
	return r.resolveType(ctx, typeName, kind, label)
}

func (r *Resolver) toPluginFunctionID(prefix string, t []*Type) string {
	var typeNames []string
	for _, tt := range t {
		if tt == nil {
			continue
		}
		typeNames = append(typeNames, tt.FQDN())
	}
	return strings.ReplaceAll(strings.Join(append([]string{prefix}, typeNames...), "_"), ".", "_")
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
				source.NewLocationBuilder(ctx.fileName()).WithService(name).Location(),
			),
		)
		return nil
	}
	ruleDef, err := getExtensionRule[*federation.ServiceRule](serviceDef.GetOptions(), federation.E_Service)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).WithService(name).Location(),
			),
		)
		return nil
	}
	plugins := make([]*CELPlugin, 0, len(r.celPluginMap))
	for _, plugin := range r.celPluginMap {
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	service := &Service{
		File:       file,
		Name:       name,
		Methods:    make([]*Method, 0, len(serviceDef.GetMethod())),
		CELPlugins: plugins,
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
	location := source.NewLocationBuilder(ctx.fileName()).
		WithService(ctx.serviceName()).
		WithOption().WithDependencies(ctx.depIndex()).WithService().Location()
	if serviceWithPkgName == "" {
		ctx.addError(
			ErrWithLocation(
				`"service" must be specified`,
				location,
			),
		)
	} else {
		pkg, err := r.lookupPackage(serviceWithPkgName)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					location,
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
						location,
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
				source.NewLocationBuilder(ctx.fileName()).
					WithService(service.Name).
					WithMethod(methodDef.GetName()).Location(),
			),
		)
	}
	resPkg, err := r.lookupPackage(methodDef.GetOutputType())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).
					WithService(service.Name).
					WithMethod(methodDef.GetName()).Location(),
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
	ruleDef, err := getExtensionRule[*federation.MethodRule](methodDef.GetOptions(), federation.E_Method)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).
					WithService(service.Name).
					WithMethod(methodDef.GetName()).
					WithOption().Location(),
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

func (r *Resolver) resolveMessageByName(ctx *context, name string) (*Message, error) {
	if strings.Contains(name, ".") {
		pkg, err := r.lookupPackage(name)
		if err != nil {
			return nil, err
		}
		msgName := r.trimPackage(pkg, name)
		return r.resolveMessage(ctx, pkg, msgName), nil
	}
	return r.resolveMessage(ctx, ctx.file().Package, name), nil
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
				source.NewLocationBuilder(ctx.fileName()).WithMessage(name).Location(),
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
	rule, err := getExtensionRule[*federation.MessageRule](msgDef.GetOptions(), federation.E_Message)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).WithMessage(msgDef.GetName()).WithOption().Location(),
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
	rule, err := getExtensionRule[*federation.OneofRule](def.GetOptions(), federation.E_Oneof)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).
					WithMessage(ctx.messageName()).
					WithOneof(def.GetName()).Location(),
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
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), name).Location(),
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
		rule, err := getExtensionRule[*federation.EnumValueRule](valueDef.GetOptions(), federation.E_EnumValue)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), name).WithValue(valueName).Location(),
				),
			)
		}
		enumValue := &EnumValue{Value: valueName, Enum: enum}
		values = append(values, enumValue)
		r.enumValueToRuleMap[enumValue] = rule
	}
	enum.Values = values
	rule, err := getExtensionRule[*federation.EnumRule](def.GetOptions(), federation.E_Enum)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), name).Location(),
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
		if msg.HasCustomResolver() || msg.HasCustomResolverFields() {
			// If use custom resolver, set the `Used` flag true
			// because all dependency message references are passed as arguments for custom resolver.
			msg.UseAllNameReference()
		}
		r.resolveMessageRules(ctx, msg.NestedMessages)
	}
}

func (r *Resolver) resolveFieldRules(ctx *context, msg *Message) {
	for _, field := range msg.Fields {
		field.Rule = r.resolveFieldRule(ctx, msg, field, r.fieldToRuleMap[field])
		if msg.Rule == nil && field.Rule != nil {
			msg.Rule = &MessageRule{}
		}
	}
	r.validateFieldsOneofRule(ctx, msg)
}

func (r *Resolver) validateFieldsOneofRule(ctx *context, msg *Message) {
	var usedDefault bool
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		oneof := field.Rule.Oneof
		if oneof == nil {
			continue
		}
		if oneof.Default {
			if usedDefault {
				ctx.addError(
					ErrWithLocation(
						`"default" found multiple times in the "grpc.federation.field.oneof". "default" can only be specified once per oneof`,
						source.NewLocationBuilder(msg.File.Name).
							WithMessage(msg.Name).
							WithField(field.Name).
							WithOption().
							WithOneOf().WithDefault().Location(),
					),
				)
			} else {
				usedDefault = true
			}
		}
		if !oneof.Default && oneof.If == nil {
			ctx.addError(
				ErrWithLocation(
					`"if" or "default" must be specified in "grpc.federation.field.oneof"`,
					source.NewLocationBuilder(msg.File.Name).
						WithMessage(msg.Name).
						WithField(field.Name).
						WithOption().WithOneOf().Location(),
				),
			)
		}
		if oneof.By == nil {
			ctx.addError(
				ErrWithLocation(
					`"by" must be specified in "grpc.federation.field.oneof"`,
					source.NewLocationBuilder(msg.File.Name).
						WithMessage(msg.Name).
						WithField(field.Name).
						WithOption().WithOneOf().Location(),
				),
			)
		}
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
	for _, varDef := range rule.VariableDefinitions {
		if !varDef.AutoBind {
			continue
		}
		typ := varDef.Expr.Type
		if typ == nil {
			continue
		}
		if typ.Kind != types.Message {
			continue
		}
		for _, field := range typ.Message.Fields {
			autobindFieldMap[field.Name] = append(autobindFieldMap[field.Name], &AutoBindField{
				VariableDefinition: varDef,
				Field:              field,
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
				if autoBindField.VariableDefinition != nil {
					locates = append(locates, fmt.Sprintf(`%q name at def`, autoBindField.VariableDefinition.Name))
				}
			}
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q field found multiple times in the message specified by autobind. since it is not possible to determine one, please use "grpc.federation.field" to explicitly bind it. found message names are %s`, field.Name, strings.Join(locates, " and ")),
					source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
				),
			)
			continue
		}
		autoBindField := autoBindFields[0]
		if autoBindField.Field.Type == nil || field.Type == nil {
			continue
		}
		if autoBindField.Field.Type.Kind != field.Type.Kind {
			continue
		}
		if autoBindField.VariableDefinition != nil {
			autoBindField.VariableDefinition.Used = true
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
					source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
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
		if fieldType.Kind != types.Message && fieldType.Kind != types.Enum {
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
	if fromType.Kind == types.Message {
		if fromType.Message == nil || toType.Message == nil {
			return
		}
		if fromType.Message.IsMapEntry {
			// If it is a map entry, ignore it.
			return
		}
		fromMessageName := fromType.Message.FQDN()
		toMessage := toType.Message
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
					source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(toField.Name).Location(),
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
					source.NewLocationBuilder(toMessage.File.Name).
						WithMessage(toMessage.Name).
						WithOption().WithAlias().Location(),
				),
			)
			return
		}
		return
	}
	if fromType.Kind == types.Enum {
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
					source.NewEnumBuilder(ctx.fileName(), toEnumMessageName, toEnum.Name).Location(),
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
					source.NewEnumBuilder(ctx.fileName(), toEnumMessageName, toEnum.Name).WithOption().Location(),
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
						source.NewLocationBuilder(ctx.fileName()).
							WithService(ctx.serviceName()).
							WithOption().WithDependencies(idx).WithName().Location(),
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
					source.NewLocationBuilder(ctx.fileName()).
						WithService(ctx.serviceName()).
						WithMethod(ctx.methodName()).
						WithOption().WithTimeout().Location(),
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
	ctx = ctx.withDefOwner(&VariableDefinitionOwner{
		Type:    VariableDefinitionOwnerMessage,
		Message: msg,
	})
	msg.Rule = &MessageRule{
		VariableDefinitions: r.resolveVariableDefinitions(ctx, ruleDef.GetDef()),
		CustomResolver:      ruleDef.GetCustomResolver(),
		Alias:               r.resolveMessageAlias(ctx, ruleDef.GetAlias()),
		DependencyGraph:     &MessageDependencyGraph{},
	}
}

func (r *Resolver) resolveVariableDefinitions(ctx *context, varDefs []*federation.VariableDefinition) []*VariableDefinition {
	ctx.clearVariableDefinitions()

	var ret []*VariableDefinition
	for idx, varDef := range varDefs {
		ctx := ctx.withDefIndex(idx)
		vd := r.resolveVariableDefinition(ctx, varDef)
		ctx.addVariableDefinition(vd)
		ret = append(ret, vd)
	}
	return ret
}

func (r *Resolver) resolveVariableDefinition(ctx *context, varDef *federation.VariableDefinition) *VariableDefinition {
	var ifValue *CELValue
	if varDef.GetIf() != "" {
		ifValue = &CELValue{Expr: varDef.GetIf()}
	}
	return &VariableDefinition{
		Idx:      ctx.defIndex(),
		Owner:    ctx.defOwner(),
		Name:     r.resolveVariableName(ctx, varDef.GetName()),
		If:       ifValue,
		AutoBind: varDef.GetAutobind(),
		Expr:     r.resolveVariableExpr(ctx, varDef),
	}
}

func (r *Resolver) resolveVariableName(ctx *context, name string) string {
	if name != "" {
		return name
	}
	return fmt.Sprintf("_def%d", ctx.defIndex())
}

func (r *Resolver) resolveVariableExpr(ctx *context, varDef *federation.VariableDefinition) *VariableExpr {
	switch varDef.GetExpr().(type) {
	case *federation.VariableDefinition_By:
		return &VariableExpr{By: &CELValue{Expr: varDef.GetBy()}}
	case *federation.VariableDefinition_Map:
		return &VariableExpr{Map: r.resolveMapExpr(ctx, varDef.GetMap())}
	case *federation.VariableDefinition_Message:
		return &VariableExpr{Message: r.resolveMessageExpr(ctx, varDef.GetMessage())}
	case *federation.VariableDefinition_Call:
		return &VariableExpr{Call: r.resolveCallExpr(ctx, varDef.GetCall())}
	case *federation.VariableDefinition_Validation:
		return &VariableExpr{Validation: r.resolveValidationExpr(ctx, varDef.GetValidation())}
	}
	return nil
}

func (r *Resolver) resolveMapExpr(ctx *context, def *federation.MapExpr) *MapExpr {
	if def == nil {
		return nil
	}
	return &MapExpr{
		Iterator: r.resolveMapIterator(ctx, def.GetIterator()),
		Expr:     r.resolveMapIteratorExpr(ctx, def),
	}
}

func (r *Resolver) resolveMapIterator(ctx *context, def *federation.Iterator) *Iterator {
	name := def.GetName()
	if name == "" {
		ctx.addError(
			ErrWithLocation(
				"map iterator name must be specified",
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithMap().WithIteratorName().Location(),
			),
		)
	}
	var srcDef *VariableDefinition
	if src := def.GetSrc(); src == "" {
		ctx.addError(
			ErrWithLocation(
				"map iterator src must be specified",
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithMap().WithIteratorSource().Location(),
			),
		)
	} else {
		srcDef = ctx.variableDef(src)
		if srcDef == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q variable is not defined`, src),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithMap().WithIteratorSource().Location(),
				),
			)
		}
	}
	return &Iterator{
		Name:   name,
		Source: srcDef,
	}
}

func (r *Resolver) resolveMapIteratorExpr(ctx *context, def *federation.MapExpr) *MapIteratorExpr {
	switch def.GetExpr().(type) {
	case *federation.MapExpr_By:
		return &MapIteratorExpr{By: &CELValue{Expr: def.GetBy()}}
	case *federation.MapExpr_Message:
		return &MapIteratorExpr{Message: r.resolveMessageExpr(ctx, def.GetMessage())}
	}
	return nil
}

func (r *Resolver) resolveMessageExpr(ctx *context, def *federation.MessageExpr) *MessageExpr {
	if def == nil {
		return nil
	}
	msg, err := r.resolveMessageByName(ctx, def.GetName())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithMessage().WithName().Location(),
			),
		)
	}
	if ctx.msg == msg {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`recursive definition: %q is own message name`, msg.Name),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithMessage().WithName().Location(),
			),
		)
	}
	args := make([]*Argument, 0, len(def.GetArgs()))
	for idx, argDef := range def.GetArgs() {
		args = append(args, r.resolveMessageExprArgument(ctx.withArgIndex(idx), argDef))
	}
	return &MessageExpr{
		Message: msg,
		Args:    args,
	}
}

func (r *Resolver) resolveCallExpr(ctx *context, def *federation.CallExpr) *CallExpr {
	if def == nil {
		return nil
	}

	pkgName, serviceName, methodName, err := r.splitMethodFullName(ctx.file().Package, def.GetMethod())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().WithMethod().Location(),
			),
		)
		return nil
	}
	pkg, exists := r.protoPackageNameToPackage[pkgName]
	if !exists {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q package does not exist`, pkgName),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().WithMethod().Location(),
			),
		)
		return nil
	}
	service := r.resolveService(ctx, pkg, serviceName)
	if service == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`cannot find %q method because the service to which the method belongs does not exist`, methodName),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().WithMethod().Location(),
			),
		)
		return nil
	}

	method := service.Method(methodName)
	if method == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q method does not exist in %s service`, methodName, service.Name),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().WithMethod().Location(),
			),
		)
		return nil
	}

	var timeout *time.Duration
	timeoutDef := def.GetTimeout()
	if timeoutDef != "" {
		duration, err := time.ParseDuration(timeoutDef)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithCall().WithTimeout().Location(),
				),
			)
		} else {
			timeout = &duration
		}
	}

	return &CallExpr{
		Method:  method,
		Request: r.resolveRequest(ctx, method, def.GetRequest()),
		Timeout: timeout,
		Retry:   r.resolveRetry(ctx, def.GetRetry(), timeout),
	}
}

func (r *Resolver) resolveValidationExpr(ctx *context, def *federation.ValidationExpr) *ValidationExpr {
	e := def.GetError()
	if e.GetIf() != "" && len(e.GetDetails()) != 0 {
		ctx.addError(
			ErrWithLocation(
				"cannot set both rule and details at the same time",
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithValidation().WithIf().Location(),
			),
		)
		return nil
	}

	vr := &ValidationExpr{
		Error: &ValidationError{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
		},
	}

	if expr := e.GetIf(); expr != "" {
		vr.Error.If = &CELValue{
			Expr: expr,
		}
		return vr
	}

	if details := e.GetDetails(); len(details) != 0 {
		vr.Error.Details = r.resolveMessageRuleValidationDetails(ctx, details)
		return vr
	}

	ctx.addError(
		ErrWithLocation(
			"either rule or details should be specified",
			source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
				WithValidation().WithIf().Location(),
		),
	)
	return nil
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
					source.NewLocationBuilder(ctx.fileName()).
						WithMessage(ctx.messageName()).
						WithOption().WithAlias().Location(),
				),
			)
			return nil
		}
		name := r.trimPackage(pkg, aliasName)
		return r.resolveMessage(ctx, pkg, name)
	}
	return r.resolveMessage(ctx, ctx.file().Package, aliasName)
}

func (r *Resolver) resolveMessageRuleValidationDetails(ctx *context, details []*federation.ValidationErrorDetail) []*ValidationErrorDetail {
	result := make([]*ValidationErrorDetail, 0, len(details))
	for idx, detail := range details {
		ctx := ctx.withErrDetailIndex(idx)
		result = append(result, &ValidationErrorDetail{
			If: &CELValue{
				Expr: detail.GetIf(),
			},
			Messages:             r.resolveValidationDetailMessages(ctx, detail.GetMessage()),
			PreconditionFailures: r.resolvePreconditionFailures(detail.GetPreconditionFailure()),
			BadRequests:          r.resolveBadRequests(detail.GetBadRequest()),
			LocalizedMessages:    r.resolveLocalizedMessages(detail.GetLocalizedMessage()),
		})
	}
	return result
}

func (r *Resolver) resolveValidationDetailMessages(ctx *context, messages []*federation.MessageExpr) VariableDefinitions {
	if len(messages) == 0 {
		return nil
	}
	msgs := make([]*federation.VariableDefinition, 0, len(messages))
	for idx, message := range messages {
		name := fmt.Sprintf("_def%d_err_detail%d_msg%d", ctx.defIndex(), ctx.errDetailIndex(), idx)
		msgs = append(msgs, &federation.VariableDefinition{
			Name: &name,
			Expr: &federation.VariableDefinition_Message{
				Message: message,
			},
		})
	}
	ctx = ctx.withDefOwner(&VariableDefinitionOwner{
		Type: VariableDefinitionOwnerValidationErrorDetailMessage,
		ValidationErrorIndexes: &ValidationErrorIndexes{
			DefIdx:       ctx.defIndex(),
			ErrDetailIdx: ctx.errDetailIndex(),
		},
	})
	defs := make([]*VariableDefinition, 0, len(msgs))
	for _, def := range r.resolveVariableDefinitions(ctx, msgs) {
		def.Used = true
		defs = append(defs, def)
	}
	return defs
}

func (r *Resolver) resolvePreconditionFailures(failures []*errdetails.PreconditionFailure) []*PreconditionFailure {
	if len(failures) == 0 {
		return nil
	}
	result := make([]*PreconditionFailure, 0, len(failures))
	for _, failure := range failures {
		result = append(result, &PreconditionFailure{
			Violations: r.resolvePreconditionFailureViolations(failure.GetViolations()),
		})
	}
	return result
}

func (r *Resolver) resolvePreconditionFailureViolations(violations []*errdetails.PreconditionFailure_Violation) []*PreconditionFailureViolation {
	if len(violations) == 0 {
		return nil
	}
	result := make([]*PreconditionFailureViolation, 0, len(violations))
	for _, violation := range violations {
		result = append(result, &PreconditionFailureViolation{
			Type: &CELValue{
				Expr: violation.GetType(),
			},
			Subject: &CELValue{
				Expr: violation.GetSubject(),
			},
			Description: &CELValue{
				Expr: violation.GetDescription(),
			},
		})
	}
	return result
}

func (r *Resolver) resolveBadRequests(reqs []*errdetails.BadRequest) []*BadRequest {
	if len(reqs) == 0 {
		return nil
	}
	result := make([]*BadRequest, 0, len(reqs))
	for _, req := range reqs {
		result = append(result, &BadRequest{
			FieldViolations: r.resolveBadRequestFieldViolations(req.GetFieldViolations()),
		})
	}
	return result
}

func (r *Resolver) resolveBadRequestFieldViolations(violations []*errdetails.BadRequest_FieldViolation) []*BadRequestFieldViolation {
	if len(violations) == 0 {
		return nil
	}
	result := make([]*BadRequestFieldViolation, 0, len(violations))
	for _, violation := range violations {
		result = append(result, &BadRequestFieldViolation{
			Field: &CELValue{
				Expr: violation.GetField(),
			},
			Description: &CELValue{
				Expr: violation.GetDescription(),
			},
		})
	}
	return result
}

func (r *Resolver) resolveLocalizedMessages(messages []*errdetails.LocalizedMessage) []*LocalizedMessage {
	if len(messages) == 0 {
		return nil
	}
	result := make([]*LocalizedMessage, 0, len(messages))
	for _, req := range messages {
		result = append(result, &LocalizedMessage{
			Locale: req.GetLocale(),
			Message: &CELValue{
				Expr: req.GetMessage(),
			},
		})
	}
	return result
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
					source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
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
				source.NewLocationBuilder(ctx.fileName()).WithMessage(msg.Name).WithField(field.Name).Location(),
			),
		)
		return nil
	}
	var (
		ifValue *CELValue
		by      *CELValue
	)
	if v := def.GetIf(); v != "" {
		ifValue = &CELValue{Expr: v}
	}
	if b := def.GetBy(); b != "" {
		by = &CELValue{Expr: b}
	}
	ctx = ctx.withDefOwner(&VariableDefinitionOwner{
		Type:  VariableDefinitionOwnerOneofField,
		Field: field,
	})
	return &FieldOneofRule{
		If:                  ifValue,
		Default:             def.GetDefault(),
		VariableDefinitions: r.resolveVariableDefinitions(ctx, def.GetDef()),
		By:                  by,
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
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if field.Type.Kind != aliasField.Type.Kind {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, field.Type.Kind.ToString(),
					msgAlias.FQDN(), aliasField.Type.Kind.ToString(),
				),
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
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
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
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
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if field.Type.Kind != aliasField.Type.Kind {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, field.Type.Kind.ToString(),
					msgAlias.FQDN(), aliasField.Type.Kind.ToString(),
				),
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(field.Name).Location(),
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
					source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithOption().Location(),
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
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithValue(enumValueName).Location(),
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
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).
					WithValue(enumValueName).
					WithOption().WithAlias().Location(),
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
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithValue(enumValueName).Location(),
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
	typ, err := r.resolveType(ctx, fieldDef.GetTypeName(), types.Kind(fieldDef.GetType()), fieldDef.GetLabel())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(fieldDef.GetName()).Location(),
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
	rule, err := getExtensionRule[*federation.FieldRule](fieldDef.GetOptions(), federation.E_Field)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(ctx.fileName()).WithMessage(ctx.messageName()).WithField(fieldDef.GetName()).Location(),
			),
		)
	}
	r.fieldToRuleMap[field] = rule
	return field
}

func (r *Resolver) resolveType(ctx *context, typeName string, kind types.Kind, label descriptorpb.FieldDescriptorProto_Label) (*Type, error) {
	var (
		msg  *Message
		enum *Enum
	)
	switch kind {
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
		msg = r.resolveMessage(ctx, pkg, name)
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
		Kind:     kind,
		Repeated: label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED,
		Message:  msg,
		Enum:     enum,
	}, nil
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
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithCall().
						WithRetry().WithConstantInterval().Location(),
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
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithCall().
						WithRetry().WithExponentialInitialInterval().Location(),
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
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithCall().
						WithRetry().WithExponentialMaxInterval().Location(),
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
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().WithRequest(idx).WithField().Location(),
			))
		} else {
			argType = reqType.Field(fieldName).Type
		}
		value, err := r.resolveValue(ctx, methodRequestToCommonValueDef(req))
		if err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
					WithCall().
					WithRequest(idx).WithBy().Location(),
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

func (r *Resolver) resolveMessageExprArgument(ctx *context, argDef *federation.Argument) *Argument {
	value, err := r.resolveValue(ctx, argumentToCommonValueDef(argDef))
	if err != nil {
		switch {
		case argDef.GetBy() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithMessage().
						WithArgs(ctx.argIndex()).WithBy().Location(),
				),
			)
		case argDef.GetInline() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithMessage().
						WithArgs(ctx.argIndex()).
						WithInline().Location(),
				),
			)
		}
	}
	name := argDef.GetName()
	if err := r.validateName(name); err != nil {
		ctx.addError(ErrWithLocation(
			err.Error(),
			source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
				WithMessage().
				WithArgs(ctx.argIndex()).WithName().Location(),
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
	if t.Message == nil {
		return nil, nil, fmt.Errorf(`%q message does not exist`, msgName)
	}
	fieldMap := map[string]*Value{}
	for _, field := range val.GetFields() {
		fieldName := field.GetField()
		if !t.Message.HasField(fieldName) {
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
		// The root message is always the response message of the method.
		resMsg := root.Message

		var msgArg *Message
		if resMsg.Rule.MessageArgument != nil {
			msgArg = resMsg.Rule.MessageArgument
		} else {
			msgArg = newMessageArgument(resMsg)
			resMsg.Rule.MessageArgument = msgArg
		}

		// The message argument of the response message is the request message.
		// Therefore, the request message is retrieved from the response message.
		reqMsg := r.lookupRequestMessageFromResponseMessage(resMsg)
		if reqMsg == nil {
			// A non-response message may also become a root message.
			// In such a case, the message argument field does not exist.
			// However, since it is necessary to resolve the CEL reference, needs to call recursive message argument resolver.
			_ = r.resolveMessageArgumentRecursive(ctx, root)
			continue
		}

		msgArg.Fields = append(msgArg.Fields, reqMsg.Fields...)
		r.cachedMessageMap[msgArg.FQDN()] = msgArg

		// Store the messages to serviceMsgMap first to avoid inserting duplicated ones to Service.Messages
		for _, svc := range r.cachedServiceMap {
			if _, exists := svcMsgSet[svc]; !exists {
				svcMsgSet[svc] = make(map[*Message]struct{})
			}
			// If the method of the service has not a response message, it is excluded.
			if !svc.HasMessageInMethod(resMsg) {
				continue
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
				source.NewLocationBuilder(msg.File.Name).WithMessage(msg.Name).Location(),
			),
		)
		return nil
	}
	env, err := r.createCELEnv(msg)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewLocationBuilder(msg.File.Name).WithMessage(msg.Name).Location(),
			),
		)
		return nil
	}
	r.resolveMessageCELValues(ctx.withFile(msg.File).withMessage(msg), env, msg)

	msgs := []*Message{msg}
	msgToDefsMap := make(map[*Message][]*VariableDefinition)
	for _, varDef := range msg.Rule.VariableDefinitions {
		for _, msgExpr := range varDef.MessageExprs() {
			msgToDefsMap[msgExpr.Message] = append(msgToDefsMap[msgExpr.Message], varDef)
		}
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for _, varDef := range field.Rule.Oneof.VariableDefinitions {
			for _, msgExpr := range varDef.MessageExprs() {
				msgToDefsMap[msgExpr.Message] = append(msgToDefsMap[msgExpr.Message], varDef)
			}
		}
	}
	for _, child := range node.Children {
		depMsg := child.Message
		var depMsgArg *Message
		if depMsg.Rule == nil {
			continue
		}
		if depMsg.Rule.MessageArgument != nil {
			depMsgArg = depMsg.Rule.MessageArgument
		} else {
			depMsgArg = newMessageArgument(depMsg)
			depMsg.Rule.MessageArgument = depMsgArg
		}
		if _, exists := r.cachedMessageMap[depMsgArg.FQDN()]; !exists {
			r.cachedMessageMap[depMsgArg.FQDN()] = depMsgArg
			defs := msgToDefsMap[depMsg]
			depMsgArg.Fields = append(depMsgArg.Fields, r.resolveMessageArgumentFields(ctx, msg, defs)...)
		}
		m := r.resolveMessageArgumentRecursive(ctx, child)
		msgs = append(msgs, m...)
	}
	return msgs
}

func (r *Resolver) resolveMessageArgumentFields(ctx *context, msg *Message, defs []*VariableDefinition) []*Field {
	argNameMap := make(map[string]struct{})
	for _, varDef := range defs {
		for _, msgExpr := range varDef.MessageExprs() {
			for _, arg := range msgExpr.Args {
				if arg.Name == "" {
					continue
				}
				argNameMap[arg.Name] = struct{}{}
			}
		}
	}

	evaluatedArgNameMap := make(map[string]struct{})
	var fields []*Field
	for _, varDef := range defs {
		for _, msgExpr := range varDef.MessageExprs() {
			r.validateMessageDependencyArgumentName(ctx, argNameMap, msg, varDef)
			for argIdx, arg := range msgExpr.Args {
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
				if arg.Value.CEL != nil && arg.Value.Inline {
					if fieldType.Kind != types.Message {
						ctx.addError(
							ErrWithLocation(
								"inline value is not message type",
								source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, varDef.Idx).
									WithMessage().
									WithArgs(argIdx).
									WithInline().Location(),
							),
						)
						continue
					}
					fields = append(fields, fieldType.Message.Fields...)
				} else {
					fields = append(fields, &Field{
						Name: arg.Name,
						Type: fieldType,
					})
				}
				evaluatedArgNameMap[arg.Name] = struct{}{}
			}
		}
	}
	return fields
}

func (r *Resolver) validateMessageDependencyArgumentName(ctx *context, argNameMap map[string]struct{}, msg *Message, def *VariableDefinition) {
	for _, msgExpr := range def.MessageExprs() {
		curDepArgNameMap := make(map[string]struct{})
		for _, arg := range msgExpr.Args {
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
			switch def.Owner.Type {
			case VariableDefinitionOwnerMessage:
				errLoc = source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, def.Idx).
					WithMessage().
					WithArgs(0).Location()
			case VariableDefinitionOwnerOneofField:
				errLoc = source.NewLocationBuilder(msg.File.Name).
					WithMessage(msg.Name).
					WithField(def.Owner.Field.Name).
					WithOption().
					WithOneOf().
					WithVariableDefinitions(def.Idx).
					WithMessage().
					WithArgs(0).Location()
			case VariableDefinitionOwnerValidationErrorDetailMessage:
				errLoc = source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, def.Owner.ValidationErrorIndexes.DefIdx).
					WithValidation().
					WithDetail(def.Owner.ValidationErrorIndexes.ErrDetailIdx).
					WithMessage(def.Idx).
					WithMessage().
					WithArgs(0).Location()
			}
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf("%q argument is defined in other message dependency arguments, but not in this context", name),
					errLoc,
				),
			)
		}
	}
}

func (r *Resolver) resolveMessageCELValues(ctx *context, env *cel.Env, msg *Message) {
	if msg.Rule == nil {
		return
	}
	for idx, varDef := range msg.Rule.VariableDefinitions {
		ctx := ctx.withDefIndex(idx)
		if varDef.If != nil {
			if err := r.resolveCELValue(ctx, env, varDef.If); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, idx).WithIf().Location(),
					),
				)
			}
			if varDef.If.Out != nil {
				if varDef.If.Out.Kind != types.Bool {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`return value of "if" must be bool type but got %s type`, varDef.If.Out.Kind.ToString()),
							source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, idx).WithIf().Location(),
						),
					)
				}
			}
		}
		if varDef.Expr == nil {
			continue
		}
		r.resolveVariableExprCELValues(ctx, env, varDef.Expr)
		if varDef.Name != "" && varDef.Expr.Type != nil {
			newEnv, err := env.Extend(cel.Variable(varDef.Name, ToCELType(varDef.Expr.Type)))
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`failed to extend cel.Env from variables of messages: %s`, err.Error()),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, idx).Location(),
					),
				)
				continue
			}
			env = newEnv
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
						source.NewLocationBuilder(msg.File.Name).
							WithMessage(msg.Name).
							WithField(field.Name).
							WithOption().WithBy().Location(),
					),
				)
			}
		}
		if field.Rule.Oneof != nil {
			fieldEnv, _ := env.Extend()
			oneof := field.Rule.Oneof
			if oneof.If != nil {
				if err := r.resolveCELValue(ctx, fieldEnv, oneof.If); err != nil {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewLocationBuilder(msg.File.Name).
								WithMessage(msg.Name).
								WithField(field.Name).
								WithOption().
								WithOneOf().WithIf().Location(),
						),
					)
				}
				if oneof.If.Out != nil {
					if oneof.If.Out.Kind != types.Bool {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`return value of "if" must be bool type but got %s type`, oneof.If.Out.Kind.ToString()),
								source.NewLocationBuilder(msg.File.Name).
									WithMessage(msg.Name).
									WithField(field.Name).
									WithOption().
									WithOneOf().WithIf().Location(),
							),
						)
					}
				}
			}
			for idx, varDef := range oneof.VariableDefinitions {
				if varDef.Expr == nil {
					continue
				}
				r.resolveVariableExprCELValues(ctx, fieldEnv, varDef.Expr)
				if varDef.Name != "" && varDef.Expr.Type != nil {
					newEnv, err := fieldEnv.Extend(cel.Variable(varDef.Name, ToCELType(varDef.Expr.Type)))
					if err != nil {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`failed to extend cel.Env from variables of messages: %s`, err.Error()),
								source.NewLocationBuilder(msg.File.Name).
									WithMessage(msg.Name).
									WithField(field.Name).
									WithOption().
									WithOneOf().
									WithVariableDefinitions(idx).
									WithMessage().Location(),
							),
						)
						continue
					}
					fieldEnv = newEnv
				}
			}
			if err := r.resolveCELValue(ctx, fieldEnv, oneof.By); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewLocationBuilder(msg.File.Name).
							WithMessage(msg.Name).
							WithField(field.Name).
							WithOption().
							WithOneOf().WithBy().Location(),
					),
				)
			}
		}
	}

	r.resolveUsedNameReference(msg)
}

func (r *Resolver) resolveVariableExprCELValues(ctx *context, env *cel.Env, expr *VariableExpr) {
	switch {
	case expr.By != nil:
		if err := r.resolveCELValue(ctx, env, expr.By); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).WithBy().Location(),
				),
			)
		}
		expr.Type = expr.By.Out
	case expr.Map != nil:
		mapEnv := env
		iter := expr.Map.Iterator
		if iter != nil && iter.Name != "" && iter.Source != nil {
			if !iter.Source.Expr.Type.Repeated {
				ctx.addError(
					ErrWithLocation(
						`map iterator's src value type must be repeated type`,
						source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
							WithMap().WithIteratorSource().Location(),
					),
				)
				return
			}
			iterType := iter.Source.Expr.Type.Clone()
			iterType.Repeated = false
			newEnv, err := env.Extend(cel.Variable(iter.Name, ToCELType(iterType)))
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
							WithMap().WithIteratorSource().Location(),
					),
				)
			}
			mapEnv = newEnv
		}
		r.resolveMapIteratorExprCELValues(ctx, mapEnv, expr.Map.Expr)
		varType := expr.Map.Expr.Type.Clone()
		varType.Repeated = true
		expr.Type = varType
	case expr.Call != nil:
		if expr.Call.Request != nil {
			for idx, arg := range expr.Call.Request.Args {
				if arg.Value == nil {
					continue
				}
				if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
								WithCall().
								WithRequest(idx).WithBy().Location(),
						),
					)
				}
			}
		}
		expr.Type = NewMessageType(expr.Call.Method.Response, false)
	case expr.Message != nil:
		for argIdx, arg := range expr.Message.Args {
			if arg.Value == nil {
				continue
			}
			if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
				if arg.Value.Inline {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
								WithMessage().
								WithArgs(argIdx).
								WithInline().Location(),
						),
					)
				} else {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
								WithMessage().WithArgs(argIdx).WithBy().Location(),
						),
					)
				}
			}
		}
		expr.Type = NewMessageType(expr.Message.Message, false)
	case expr.Validation != nil:
		if expr.Validation.Error.If != nil {
			e := expr.Validation.Error
			if err := r.resolveCELValue(ctx, env, e.If); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
							WithValidation().WithIf().Location(),
					),
				)
				return
			}
			if e.If.Out.Kind != types.Bool {
				ctx.addError(
					ErrWithLocation(
						"if must always return a boolean value",
						source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
							WithValidation().WithIf().Location(),
					),
				)
			}
		}
		for detIdx, detail := range expr.Validation.Error.Details {
			r.resolveMessageValidationErrorDetailCELValues(ctx, env, ctx.msg, ctx.defIndex(), detIdx, detail)
		}
		// This is a dummy type since the output from the validation is not supposed to be used (at least for now)
		expr.Type = BoolType
	}
}

func (r *Resolver) resolveMapIteratorExprCELValues(ctx *context, env *cel.Env, expr *MapIteratorExpr) {
	switch {
	case expr.By != nil:
		if err := r.resolveCELValue(ctx, env, expr.By); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
						WithMap().WithBy().Location(),
				),
			)
		}
		expr.Type = expr.By.Out
	case expr.Message != nil:
		for argIdx, arg := range expr.Message.Args {
			if arg.Value == nil {
				continue
			}
			if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
				if arg.Value.Inline {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
								WithMap().
								WithMessage().
								WithArgs(argIdx).WithInline().Location(),
						),
					)
				} else {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							source.NewMsgVarDefOptionBuilder(ctx.fileName(), ctx.messageName(), ctx.defIndex()).
								WithMap().
								WithMessage().
								WithArgs(argIdx).WithBy().Location(),
						),
					)
				}
			}
		}
		expr.Type = NewMessageType(expr.Message.Message, false)
	}
}

func (r *Resolver) resolveMessageValidationErrorDetailCELValues(ctx *context, env *cel.Env, msg *Message, vIdx, dIdx int, detail *ValidationErrorDetail) {
	if err := r.resolveCELValue(ctx, env, detail.If); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
					WithValidation().
					WithDetail(dIdx).WithIf().Location(),
			),
		)
	}
	if detail.If.Out != nil && detail.If.Out.Kind != types.Bool {
		ctx.addError(
			ErrWithLocation(
				"if must always return a boolean value",
				source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
					WithValidation().
					WithDetail(dIdx).WithIf().Location(),
			),
		)
	}
	for _, message := range detail.Messages {
		r.resolveVariableExprCELValues(ctx, env, message.Expr)
	}
	for fIdx, failure := range detail.PreconditionFailures {
		for fvIdx, violation := range failure.Violations {
			if err := r.resolveCELValue(ctx, env, violation.Type); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "type").Location(),
					),
				)
			}
			if violation.Type.Out != nil && violation.Type.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"type must always return a string value",
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "type").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Subject); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "subject").Location(),
					),
				)
			}
			if violation.Subject.Out != nil && violation.Subject.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"subject must always return a string value",
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "subject").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Description); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "description").Location(),
					),
				)
			}
			if violation.Description.Out != nil && violation.Description.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"description must always return a string value",
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithPreconditionFailure(fIdx, fvIdx, "description").Location(),
					),
				)
			}
		}
	}

	for bIdx, badRequest := range detail.BadRequests {
		for fvIdx, violation := range badRequest.FieldViolations {
			if err := r.resolveCELValue(ctx, env, violation.Field); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithBadRequest(bIdx, fvIdx, "field").Location(),
					),
				)
			}
			if violation.Field.Out != nil && violation.Field.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"field must always return a string value",
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithBadRequest(bIdx, fvIdx, "field").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Description); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithBadRequest(bIdx, fvIdx, "description").Location(),
					),
				)
			}
			if violation.Description.Out != nil && violation.Description.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"description must always return a string value",
						source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
							WithValidation().
							WithDetail(dIdx).WithBadRequest(bIdx, fvIdx, "description").Location(),
					),
				)
			}
		}
	}

	for lIdx, message := range detail.LocalizedMessages {
		if err := r.resolveCELValue(ctx, env, message.Message); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
						WithValidation().
						WithDetail(dIdx).WithLocalizedMessage(lIdx, "message").Location(),
				),
			)
		}
		if message.Message.Out != nil && message.Message.Out.Kind != types.String {
			ctx.addError(
				ErrWithLocation(
					"message must always return a string value",
					source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, vIdx).
						WithValidation().
						WithDetail(dIdx).WithLocalizedMessage(lIdx, "message").Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveUsedNameReference(msg *Message) {
	if msg.Rule == nil {
		return
	}
	nameMap := make(map[string]struct{})
	for _, name := range msg.ReferenceNames() {
		nameMap[name] = struct{}{}
	}
	for _, varDef := range msg.Rule.VariableDefinitions {
		if _, exists := nameMap[varDef.Name]; exists {
			varDef.Used = true
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
		for _, varDef := range oneof.VariableDefinitions {
			if _, exists := oneofNameMap[varDef.Name]; exists {
				varDef.Used = true
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
	r.celRegistry.clearErrors()
	ast, issues := env.Compile(expr)
	if issues.Err() != nil {
		return errors.Join(append(r.celRegistry.errors(), issues.Err())...)
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
	envOpts := []cel.EnvOption{
		cel.StdLib(),
		cel.Lib(grpcfedcel.NewLibrary()),
		cel.Lib(grpcfedcel.NewContextualLibrary(gocontext.Background())),
		cel.CrossTypeNumericComparisons(true),
		cel.CustomTypeAdapter(r.celRegistry),
		cel.CustomTypeProvider(r.celRegistry),
	}
	for _, plugin := range r.celPluginMap {
		envOpts = append(envOpts, cel.Lib(plugin))
	}

	if msg.Rule != nil && msg.Rule.MessageArgument != nil {
		envOpts = append(envOpts, cel.Variable(federation.MessageArgumentVariableName, cel.ObjectType(msg.Rule.MessageArgument.FQDN())))
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
			return &Type{Kind: types.Enum, Enum: enum}, nil
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
		if grpcfedcel.IsStandardLibraryType(typ.TypeName()) {
			pkgAndName := strings.TrimPrefix(strings.TrimPrefix(typ.TypeName(), "."), "grpc.federation.")
			names := strings.Split(pkgAndName, ".")
			if len(names) <= 1 {
				return nil, fmt.Errorf(`unexpected package name %q`, pkgAndName)
			}
			pkgName := names[0]
			msgName := names[1]
			return NewCELStandardLibraryMessageType(pkgName, msgName), nil
		}
		return r.resolveType(
			ctx,
			typ.TypeName(),
			types.Message,
			descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL,
		)
	case celtypes.OpaqueKind:
		return r.fromCELType(ctx, typ.Parameters()[0])
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
		if field.Type.Message != nil {
			typeName = field.Type.Message.FQDN()
		}
		if field.Type.Enum != nil {
			typeName = field.Type.Enum.FQDN()
		}
		label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
		if field.Type.Repeated {
			label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		}
		kind := descriptorpb.FieldDescriptorProto_Type(field.Type.Kind)
		msg.Field = append(msg.Field, &descriptorpb.FieldDescriptorProto{
			Name:     proto.String(field.Name),
			Number:   proto.Int32(int32(idx) + 1),
			Type:     &kind,
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

func (r *Resolver) resolveAutoBind(ctx *context, files []*File) {
	msgs := r.allMessages(files)
	for _, msg := range msgs {
		ctx := ctx.withFile(msg.File).withMessage(msg)
		r.resolveAutoBindFields(ctx, msg)
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
			msg.Rule.VariableDefinitionGroups = graph.VariableDefinitionGroups(ctx)
		}
		for defIdx, def := range msg.Rule.VariableDefinitions {
			if def.Expr == nil {
				continue
			}
			if def.Expr.Validation == nil || def.Expr.Validation.Error == nil {
				continue
			}
			ctx := ctx.withDefIndex(defIdx)
			for detIdx, detail := range def.Expr.Validation.Error.Details {
				ctx := ctx.withErrDetailIndex(detIdx)
				if graph := CreateMessageDependencyGraphByValidationErrorDetailMessages(msg, detail.Messages); graph != nil {
					detail.DependencyGraph = graph
					detail.VariableDefinitionGroups = graph.VariableDefinitionGroups(ctx)
				}
			}
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
				field.Rule.Oneof.VariableDefinitionGroups = graph.VariableDefinitionGroups(ctx)
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
		value = &Value{Inline: true, CEL: &CELValue{Expr: def.GetInline()}}
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
			} else if typ.Message != t.Message {
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

// ReferenceNames returns all the unique reference names in the error definition.
func (v *ValidationError) ReferenceNames() []string {
	nameSet := make(map[string]struct{})
	register := func(names []string) {
		for _, name := range names {
			nameSet[name] = struct{}{}
		}
	}
	register(v.If.ReferenceNames())
	for _, detail := range v.Details {
		register(detail.If.ReferenceNames())
		for _, message := range detail.Messages {
			register(message.ReferenceNames())
		}
		for _, failure := range detail.PreconditionFailures {
			for _, violation := range failure.Violations {
				register(violation.Type.ReferenceNames())
				register(violation.Subject.ReferenceNames())
				register(violation.Description.ReferenceNames())
			}
		}
		for _, req := range detail.BadRequests {
			for _, violation := range req.FieldViolations {
				register(violation.Field.ReferenceNames())
				register(violation.Description.ReferenceNames())
			}
		}
		for _, msg := range detail.LocalizedMessages {
			register(msg.Message.ReferenceNames())
		}
	}
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}
	return names
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
