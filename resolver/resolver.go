package resolver

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	fileNameToDefMap           map[string]*descriptorpb.FileDescriptorProto
	protoPackageNameToFileDefs map[string][]*descriptorpb.FileDescriptorProto
	protoPackageNameToPackage  map[string]*Package
	celPluginMap               map[string]*CELPlugin
	ctxOverloadIDPrefixes      []string

	serviceToRuleMap   map[*Service]*federation.ServiceRule
	methodToRuleMap    map[*Method]*federation.MethodRule
	messageToRuleMap   map[*Message]*federation.MessageRule
	enumToRuleMap      map[*Enum]*federation.EnumRule
	enumValueToRuleMap map[*EnumValue]*federation.EnumValueRule
	fieldToRuleMap     map[*Field]*federation.FieldRule
	oneofToRuleMap     map[*Oneof]*federation.OneofRule

	cachedMessageMap   map[string]*Message
	cachedEnumMap      map[string]*Enum
	cachedEnumValueMap map[string]*EnumValue
	cachedMethodMap    map[string]*Method
	cachedServiceMap   map[string]*Service
}

func New(files []*descriptorpb.FileDescriptorProto) *Resolver {
	msgMap := make(map[string]*Message)
	enumValueMap := make(map[string]*EnumValue)
	celRegistry := newCELRegistry(msgMap, enumValueMap)
	return &Resolver{
		files:                      files,
		celRegistry:                celRegistry,
		defToFileMap:               make(map[*descriptorpb.FileDescriptorProto]*File),
		fileNameToDefMap:           make(map[string]*descriptorpb.FileDescriptorProto),
		protoPackageNameToFileDefs: make(map[string][]*descriptorpb.FileDescriptorProto),
		protoPackageNameToPackage:  make(map[string]*Package),
		celPluginMap:               make(map[string]*CELPlugin),
		ctxOverloadIDPrefixes:      grpcfedcel.NewLibrary(celRegistry).ContextOverloadIDPrefixes(),

		serviceToRuleMap:   make(map[*Service]*federation.ServiceRule),
		methodToRuleMap:    make(map[*Method]*federation.MethodRule),
		messageToRuleMap:   make(map[*Message]*federation.MessageRule),
		enumToRuleMap:      make(map[*Enum]*federation.EnumRule),
		enumValueToRuleMap: make(map[*EnumValue]*federation.EnumValueRule),
		fieldToRuleMap:     make(map[*Field]*federation.FieldRule),
		oneofToRuleMap:     make(map[*Oneof]*federation.OneofRule),

		cachedMessageMap:   msgMap,
		cachedEnumMap:      make(map[string]*Enum),
		cachedEnumValueMap: enumValueMap,
		cachedMethodMap:    make(map[string]*Method),
		cachedServiceMap:   make(map[string]*Service),
	}
}

// Result of resolver processing.
type Result struct {
	// Files list of files with services with the grpc.federation.service option.
	Files []*File
	// Enums list of all enum definition.
	Enums []*Enum
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
		files = append(files, r.resolveFile(ctx, fileDef, source.NewLocationBuilder(fileDef.GetName())))
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

	resultFiles := r.resultFiles(files)
	return &Result{
		Files:    resultFiles,
		Enums:    r.allEnums(resultFiles),
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
		r.fileNameToDefMap[fileDef.GetName()] = fileDef
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
		files = append(files, r.resolveFile(ctx, fileDef, source.NewLocationBuilder(fileDef.GetName())))
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
		for _, msg := range file.Messages {
			msgs = append(msgs, msg.AllMessages()...)
		}
	}
	return msgs
}

func (r *Resolver) validateServiceFromFiles(ctx *context, files []*File) {
	for _, file := range files {
		ctx := ctx.withFile(file)
		for _, svc := range file.Services {
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

func (r *Resolver) allEnums(files []*File) []*Enum {
	var enums []*Enum
	for _, file := range files {
		enums = append(enums, file.AllEnums()...)
		for _, importFile := range file.ImportFiles {
			enums = append(enums, importFile.AllEnums()...)
		}
	}
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].FQDN() < enums[j].FQDN()
	})
	return enums
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
	r.validateMethodResponse(ctx, svc)
}

func (r *Resolver) validateMethodResponse(ctx *context, service *Service) {
	for _, method := range service.Methods {
		response := method.Response
		if response.Rule == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`"%s.%s" message needs to specify "grpc.federation.message" option`, response.PackageName(), response.Name),
					source.NewMessageBuilder(ctx.fileName(), response.Name).Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveFile(ctx *context, def *descriptorpb.FileDescriptorProto, builder *source.LocationBuilder) *File {
	file := r.defToFileMap[def]
	ctx = ctx.withFile(file)
	fileDef, err := getExtensionRule[*federation.FileRule](def.GetOptions(), federation.E_File)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
	}
	if fileDef != nil {
		for _, export := range fileDef.GetPlugin().GetExport() {
			if plugin := r.resolveCELPlugin(ctx, def, export, builder.WithExport(export.GetName())); plugin != nil {
				file.CELPlugins = append(file.CELPlugins, plugin)
			}
		}
	}

	for _, depFileName := range def.GetDependency() {
		depDef, exists := r.fileNameToDefMap[depFileName]
		if !exists {
			continue
		}
		file.ImportFiles = append(file.ImportFiles, r.defToFileMap[depDef])
	}
	for _, serviceDef := range def.GetService() {
		name := serviceDef.GetName()
		service := r.resolveService(ctx, file.Package, name, builder.WithService(name))
		if service == nil {
			continue
		}
		file.Services = append(file.Services, service)
	}
	for _, msgDef := range def.GetMessageType() {
		name := msgDef.GetName()
		msg := r.resolveMessage(ctx, file.Package, name, builder.WithMessage(name))
		if msg == nil {
			continue
		}
		file.Messages = append(file.Messages, msg)
	}
	for _, enumDef := range def.GetEnumType() {
		name := enumDef.GetName()
		file.Enums = append(file.Enums, r.resolveEnum(ctx, file.Package, name, builder.WithEnum(name)))
	}
	return file
}

func (r *Resolver) resolveCELPlugin(ctx *context, fileDef *descriptorpb.FileDescriptorProto, def *federation.CELPluginExport, builder *source.ExportBuilder) *CELPlugin {
	if def == nil {
		return nil
	}
	pkgName := fileDef.GetPackage()
	plugin := &CELPlugin{Name: def.GetName()}
	ctx = ctx.withPlugin(plugin)
	for idx, fn := range def.GetFunctions() {
		pluginFunc := r.resolvePluginGlobalFunction(ctx, pkgName, fn, builder.WithFunctions(idx))
		if pluginFunc == nil {
			continue
		}
		plugin.Functions = append(plugin.Functions, pluginFunc)
	}
	for idx, msgType := range def.GetTypes() {
		builder := builder.WithTypes(idx)
		msg, err := r.resolveMessageByName(ctx, msgType.GetName(), source.ToLazyMessageBuilder(builder, msgType.GetName()))
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithName().Location(),
				),
			)
			continue
		}
		for idx, fn := range msgType.GetMethods() {
			pluginFunc := r.resolvePluginMethod(ctx, msg, fn, builder.WithMethods(idx))
			if pluginFunc == nil {
				continue
			}
			plugin.Functions = append(plugin.Functions, pluginFunc)
		}
	}
	r.celPluginMap[def.GetName()] = plugin
	return plugin
}

func (r *Resolver) resolvePluginMethod(ctx *context, msg *Message, fn *federation.CELFunction, builder *source.PluginFunctionBuilder) *CELFunction {
	msgType := NewMessageType(msg, false)
	pluginFunc := &CELFunction{
		Name:     fn.GetName(),
		Receiver: msg,
		Args:     []*Type{msgType},
	}
	args, ret := r.resolvePluginFunctionArgumentsAndReturn(ctx, fn.GetArgs(), fn.GetReturn(), builder)
	pluginFunc.Args = append(pluginFunc.Args, args...)
	pluginFunc.Return = ret
	pluginFunc.ID = r.toPluginFunctionID(fmt.Sprintf("%s_%s", msg.FQDN(), fn.GetName()), append(pluginFunc.Args, pluginFunc.Return))
	return pluginFunc
}

func (r *Resolver) resolvePluginGlobalFunction(ctx *context, pkgName string, fn *federation.CELFunction, builder *source.PluginFunctionBuilder) *CELFunction {
	pluginFunc := &CELFunction{
		Name: fmt.Sprintf("%s.%s", pkgName, fn.GetName()),
	}
	args, ret := r.resolvePluginFunctionArgumentsAndReturn(ctx, fn.GetArgs(), fn.GetReturn(), builder)
	pluginFunc.Args = append(pluginFunc.Args, args...)
	pluginFunc.Return = ret
	pluginFunc.ID = r.toPluginFunctionID(fmt.Sprintf("%s_%s", pkgName, fn.GetName()), append(pluginFunc.Args, pluginFunc.Return))
	return pluginFunc
}

func (r *Resolver) resolvePluginFunctionArgumentsAndReturn(ctx *context, args []*federation.CELFunctionArgument, ret *federation.CELType, builder *source.PluginFunctionBuilder) ([]*Type, *Type) {
	argTypes := r.resolvePluginFunctionArguments(ctx, args, builder)
	if ret == nil {
		return argTypes, nil
	}
	retType, err := r.resolvePluginType(ctx, ret, false)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithReturnType().Location(),
			),
		)
	}
	return argTypes, retType
}

func (r *Resolver) resolvePluginFunctionArguments(ctx *context, args []*federation.CELFunctionArgument, builder *source.PluginFunctionBuilder) []*Type {
	var ret []*Type
	for argIdx, arg := range args {
		typ, err := r.resolvePluginType(ctx, arg.GetType(), false)
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithArgs(argIdx).Location(),
				),
			)
			continue
		}
		ret = append(ret, typ)
	}
	return ret
}

func (r *Resolver) resolvePluginType(ctx *context, typ *federation.CELType, repeated bool) (*Type, error) {
	var label descriptorpb.FieldDescriptorProto_Label
	if repeated {
		label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	} else {
		label = descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
	}

	switch typ.Type.(type) {
	case *federation.CELType_Kind:
		switch typ.GetKind() {
		case federation.TypeKind_STRING:
			if repeated {
				return StringRepeatedType, nil
			}
			return StringType, nil
		case federation.TypeKind_BOOL:
			if repeated {
				return BoolRepeatedType, nil
			}
			return BoolType, nil
		case federation.TypeKind_INT64:
			if repeated {
				return Int64RepeatedType, nil
			}
			return Int64Type, nil
		case federation.TypeKind_UINT64:
			if repeated {
				return Uint64RepeatedType, nil
			}
			return Uint64Type, nil
		case federation.TypeKind_DOUBLE:
			if repeated {
				return DoubleRepeatedType, nil
			}
			return DoubleType, nil
		case federation.TypeKind_DURATION:
			if repeated {
				return DurationRepeatedType, nil
			}
			return DurationType, nil
		}
	case *federation.CELType_Repeated:
		return r.resolvePluginType(ctx, typ.GetRepeated(), true)
	case *federation.CELType_Map:
		mapType := typ.GetMap()
		key, err := r.resolvePluginType(ctx, mapType.GetKey(), false)
		if err != nil {
			return nil, err
		}
		value, err := r.resolvePluginType(ctx, mapType.GetValue(), false)
		if err != nil {
			return nil, err
		}
		return NewMapType(key, value), nil
	case *federation.CELType_Message:
		ctx := newContext().withFile(ctx.file())
		return r.resolveType(ctx, typ.GetMessage(), types.Message, label)
	case *federation.CELType_Enum:
		ctx := newContext().withFile(ctx.file())
		return r.resolveType(ctx, typ.GetEnum(), types.Enum, label)
	}
	return nil, fmt.Errorf("failed to resolve plugin type")
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

func (r *Resolver) resolveService(ctx *context, pkg *Package, name string, builder *source.ServiceBuilder) *Service {
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
				builder.Location(),
			),
		)
		return nil
	}
	ruleDef, err := getExtensionRule[*federation.ServiceRule](serviceDef.GetOptions(), federation.E_Service)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
		return nil
	}
	var plugins []*CELPlugin
	if len(r.celPluginMap) != 0 {
		plugins = make([]*CELPlugin, 0, len(r.celPluginMap))
		for _, plugin := range r.celPluginMap {
			plugins = append(plugins, plugin)
		}
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Name < plugins[j].Name
		})
	}
	service := &Service{
		File:       file,
		Name:       name,
		Methods:    make([]*Method, 0, len(serviceDef.GetMethod())),
		CELPlugins: plugins,
	}
	r.serviceToRuleMap[service] = ruleDef
	for _, methodDef := range serviceDef.GetMethod() {
		method := r.resolveMethod(ctx, service, methodDef, builder.WithMethod(methodDef.GetName()))
		if method == nil {
			continue
		}
		service.Methods = append(service.Methods, method)
	}
	r.cachedServiceMap[fqdn] = service
	return service
}

func (r *Resolver) resolveMethod(ctx *context, service *Service, methodDef *descriptorpb.MethodDescriptorProto, builder *source.MethodBuilder) *Method {
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
				builder.Location(),
			),
		)
	}
	resPkg, err := r.lookupPackage(methodDef.GetOutputType())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
	}
	var (
		req *Message
		res *Message
	)
	if reqPkg != nil {
		reqType := r.trimPackage(reqPkg, methodDef.GetInputType())
		req = r.resolveMessage(ctx, reqPkg, reqType, source.NewMessageBuilder(ctx.fileName(), reqType))
	}
	if resPkg != nil {
		resType := r.trimPackage(resPkg, methodDef.GetOutputType())
		res = r.resolveMessage(ctx, resPkg, resType, source.NewMessageBuilder(ctx.fileName(), resType))
	}
	ruleDef, err := getExtensionRule[*federation.MethodRule](methodDef.GetOptions(), federation.E_Method)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithOption().Location(),
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

func (r *Resolver) resolveMessageByName(ctx *context, name string, builder *source.MessageBuilder) (*Message, error) {
	if strings.Contains(name, ".") {
		pkg, err := r.lookupPackage(name)
		if err != nil {
			// attempt to resolve the message because of a possible name specified as a nested message.
			if msg := r.resolveMessage(ctx, ctx.file().Package, name, builder); msg != nil {
				return msg, nil
			}
			return nil, err
		}
		msgName := r.trimPackage(pkg, name)
		return r.resolveMessage(ctx, pkg, msgName, source.ToLazyMessageBuilder(builder, msgName)), nil
	}
	return r.resolveMessage(ctx, ctx.file().Package, name, builder), nil
}

func (r *Resolver) resolveMessage(ctx *context, pkg *Package, name string, builder *source.MessageBuilder) *Message {
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
				builder.Location(),
			),
		)
		return nil
	}
	msg := &Message{File: file, Name: msgDef.GetName()}
	for _, nestedMsgDef := range msgDef.GetNestedType() {
		nestedMsg := r.resolveMessage(ctx, pkg, fmt.Sprintf("%s.%s", name, nestedMsgDef.GetName()), builder.WithMessage(name))
		if nestedMsg == nil {
			continue
		}
		nestedMsg.ParentMessage = msg
		msg.NestedMessages = append(msg.NestedMessages, nestedMsg)
	}
	for _, enumDef := range msgDef.GetEnumType() {
		enum := r.resolveEnum(ctx, pkg, fmt.Sprintf("%s.%s", name, enumDef.GetName()), builder.WithEnum(name))
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
				builder.WithOption().Location(),
			),
		)
	}
	r.cachedMessageMap[fqdn] = msg
	r.messageToRuleMap[msg] = rule
	ctx = ctx.withMessage(msg)
	var oneofs []*Oneof
	for _, oneofDef := range msgDef.GetOneofDecl() {
		oneof := r.resolveOneof(ctx, oneofDef, builder.WithOneof(oneofDef.GetName()))
		oneof.Message = msg
		oneofs = append(oneofs, oneof)
	}
	msg.Fields = r.resolveFields(ctx, msgDef.GetField(), oneofs, builder)
	msg.Oneofs = oneofs
	return msg
}

func (r *Resolver) resolveOneof(ctx *context, def *descriptorpb.OneofDescriptorProto, builder *source.OneofBuilder) *Oneof {
	rule, err := getExtensionRule[*federation.OneofRule](def.GetOptions(), federation.E_Oneof)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
	}
	oneof := &Oneof{
		Name: def.GetName(),
	}
	r.oneofToRuleMap[oneof] = rule
	return oneof
}

func (r *Resolver) resolveEnum(ctx *context, pkg *Package, name string, builder *source.EnumBuilder) *Enum {
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
				builder.Location(),
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
					builder.WithValue(valueName).Location(),
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
				builder.Location(),
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
		r.resolveMessageRules(ctx, file.Messages, func(name string) *source.MessageBuilder {
			return source.NewMessageBuilder(file.Name, name)
		})
		r.resolveEnumRules(ctx, file.Enums)
		r.resolveServiceRules(ctx, file.Services)
	}
}

func (r *Resolver) resolveServiceRules(ctx *context, svcs []*Service) {
	pkgToSvcs := make(map[*Package][]*Service)
	for _, svc := range svcs {
		builder := source.NewServiceBuilder(ctx.fileName(), svc.Name)
		svc.Rule = r.resolveServiceRule(ctx, r.serviceToRuleMap[svc], builder.WithOption())
		r.resolveMethodRules(ctx, svc.Methods, builder)
		pkgToSvcs[svc.Package()] = append(pkgToSvcs[svc.Package()], svc)
	}
	for pkg, svcs := range pkgToSvcs {
		var envSvcs []*Service
		for _, svc := range svcs {
			if svc.Rule == nil {
				continue
			}
			if svc.Rule.Env == nil {
				continue
			}
			envSvcs = append(envSvcs, svc)
		}
		if len(envSvcs) <= 1 {
			continue
		}
		for _, envSvc := range envSvcs {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`%q: multiple services within the same package (%q) cannot use the env option`,
						envSvc.FQDN(),
						pkg.Name,
					),
					source.NewServiceBuilder(envSvc.File.Name, envSvc.Name).WithOption().WithEnv().Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveMethodRules(ctx *context, mtds []*Method, builder *source.ServiceBuilder) {
	for _, mtd := range mtds {
		mtd.Rule = r.resolveMethodRule(ctx, r.methodToRuleMap[mtd], builder.WithMethod(mtd.Name))
	}
}

func (r *Resolver) resolveMessageRules(ctx *context, msgs []*Message, builder func(name string) *source.MessageBuilder) {
	for _, msg := range msgs {
		ctx := ctx.withMessage(msg)
		mb := builder(msg.Name)
		r.resolveMessageRule(ctx, msg, r.messageToRuleMap[msg], mb.WithOption())
		r.resolveFieldRules(ctx, msg, mb)
		r.resolveEnumRules(ctx, msg.Enums)
		if msg.HasCustomResolver() || msg.HasCustomResolverFields() {
			// If using custom resolver, set the `Used` flag true
			// because all dependency message references are passed as arguments for custom resolver.
			msg.UseAllNameReference()
		}
		r.resolveMessageRules(ctx, msg.NestedMessages, func(name string) *source.MessageBuilder {
			return mb.WithMessage(name)
		})
	}
}

func (r *Resolver) resolveFieldRules(ctx *context, msg *Message, builder *source.MessageBuilder) {
	for _, field := range msg.Fields {
		field.Rule = r.resolveFieldRule(ctx, msg, field, r.fieldToRuleMap[field], builder.WithField(field.Name))
		if msg.Rule == nil && field.Rule != nil {
			msg.Rule = &MessageRule{DefSet: &VariableDefinitionSet{}}
		}
	}
	r.validateFieldsOneofRule(ctx, msg, builder)
}

func (r *Resolver) validateFieldsOneofRule(ctx *context, msg *Message, builder *source.MessageBuilder) {
	var usedDefault bool
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		oneof := field.Rule.Oneof
		if oneof == nil {
			continue
		}
		builder := builder.WithField(field.Name).WithOption().WithOneOf()
		if oneof.Default {
			if usedDefault {
				ctx.addError(
					ErrWithLocation(
						`"default" found multiple times in the "grpc.federation.field.oneof". "default" can only be specified once per oneof`,
						builder.WithDefault().Location(),
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
					builder.Location(),
				),
			)
		}
		if oneof.By == nil {
			ctx.addError(
				ErrWithLocation(
					`"by" must be specified in "grpc.federation.field.oneof"`,
					builder.Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveAutoBindFields(ctx *context, msg *Message, builder *source.MessageBuilder) {
	if msg.Rule == nil {
		return
	}
	if msg.HasCustomResolver() {
		return
	}
	rule := msg.Rule
	autobindFieldMap := make(map[string][]*AutoBindField)
	for _, varDef := range rule.DefSet.Definitions() {
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
		builder := builder.WithField(field.Name)
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
					builder.Location(),
				),
			)
			continue
		}
		autoBindField := autoBindFields[0]
		if autoBindField.Field.Type == nil || field.Type == nil {
			continue
		}
		if isDifferentType(autoBindField.Field.Type, field.Type) {
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

var (
	nameRe             = regexp.MustCompile(namePattern)
	reservedKeywordMap = map[string]struct{}{
		"error": {},
	}
)

func (r *Resolver) validateName(name string) error {
	if name == "" {
		return nil
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf(`%q is invalid name. name should be in the following pattern: %s`, name, namePattern)
	}
	if _, exists := reservedKeywordMap[name]; exists {
		return fmt.Errorf(`%q is the reserved keyword. this name is not available`, name)
	}
	return nil
}

func (r *Resolver) validateMessages(ctx *context, msgs []*Message) {
	for _, msg := range msgs {
		ctx := ctx.withFile(msg.File).withMessage(msg)
		mb := newMessageBuilderFromMessage(msg)
		r.validateMessageFields(ctx, msg, mb)
		// Don't have to check msg.NestedMessages since r.allMessages(files) already includes nested messages
	}
}

func (r *Resolver) validateMessageFields(ctx *context, msg *Message, builder *source.MessageBuilder) {
	if msg.Rule == nil {
		return
	}
	for _, field := range msg.Fields {
		builder := builder.WithField(field.Name)
		if !field.HasRule() {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q field in %q message needs to specify "grpc.federation.field" option`, field.Name, msg.FQDN()),
					builder.Location(),
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
		rule := field.Rule
		if rule.Value != nil {
			// If you are explicitly using CEL for field binding and the binding target uses multiple enum aliases,
			// we must bind the value of the EnumSelector.
			// Otherwise, an int type might be directly specified, making it unclear which enum it should be linked to.
			if err := r.validateEnumMultipleAliases(rule.Value.Type(), field); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.Location(),
					),
				)
			} else {
				r.validateBindFieldType(ctx, rule.Value.Type(), field, builder)
			}
		}
		for _, alias := range rule.Aliases {
			r.validateBindFieldType(ctx, alias.Type, field, builder)
		}
		if rule.AutoBindField != nil {
			r.validateBindFieldType(ctx, rule.AutoBindField.Field.Type, field, builder)
		}
	}
}

func (r *Resolver) validateRequestFieldType(ctx *context, fromType *Type, toField *Field, builder *source.RequestOptionBuilder) {
	if fromType == nil || toField == nil {
		return
	}
	toType := toField.Type
	if toType.Kind == types.Message {
		if fromType.Message == nil || toType.Message == nil {
			return
		}
		if fromType.IsNumberWrapper() && toType.IsNumberWrapper() {
			// If both are of the number type from google.protobuf.wrappers, they can be mutually converted.
			return
		}
		if fromType.Message.IsMapEntry && toType.Message.IsMapEntry {
			for _, name := range []string{"key", "value"} {
				fromMapType := fromType.Message.Field(name).Type
				toMapType := toType.Message.Field(name).Type
				if isDifferentType(fromMapType, toMapType) {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(
								`cannot convert type automatically: map %s type is %q but specified map %s type is %q`,
								name, toMapType.Kind.ToString(), name, fromMapType.Kind.ToString(),
							),
							builder.Location(),
						),
					)
				}
			}
			return
		}
		fromMessage := fromType.Message
		fromMessageName := fromType.Message.FQDN()
		toMessageName := toType.Message.FQDN()
		if fromMessageName == toMessageName {
			// assignment of the same type is okay.
			return
		}
		if fromMessage.Rule == nil || len(fromMessage.Rule.Aliases) == 0 {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the %q field`,
						toMessageName, fromMessageName, toField.FQDN(),
					),
					builder.Location(),
				),
			)
			return
		}
		var found bool
		for _, alias := range fromMessage.Rule.Aliases {
			fromMessageAliasName := alias.FQDN()
			if fromMessageAliasName == toMessageName {
				found = true
				break
			}
		}
		if !found {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the %q field`,
						toMessageName, fromMessageName, toField.FQDN(),
					),
					source.NewMessageBuilder(fromMessage.File.Name, fromMessage.Name).
						WithOption().WithAlias().Location(),
				),
			)
			return
		}
	}
	if toType.Kind == types.Enum {
		if fromType.Enum == nil || toType.Enum == nil {
			return
		}
		fromEnum := fromType.Enum
		fromEnumName := fromEnum.FQDN()
		toEnumName := toType.Enum.FQDN()
		if fromEnumName == toEnumName {
			// assignment of the same type is okay.
			return
		}
		var fromEnumMessageName string
		if fromEnum.Message != nil {
			fromEnumMessageName = fromEnum.Message.Name
		}
		if fromEnum.Rule == nil || len(fromEnum.Rule.Aliases) == 0 {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.enum option for the %q type to automatically assign a value to the %q field`,
						toEnumName, fromEnumName, toField.FQDN(),
					),
					source.NewEnumBuilder(ctx.fileName(), fromEnumMessageName, fromEnum.Name).Location(),
				),
			)
			return
		}
		var found bool
		for _, alias := range fromEnum.Rule.Aliases {
			fromEnumAliasName := alias.FQDN()
			if fromEnumAliasName == toEnumName {
				found = true
				break
			}
		}
		if !found {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.enum option for the %q type to automatically assign a value to the %q field`,
						toEnumName, fromEnumName, toField.FQDN(),
					),
					source.NewEnumBuilder(ctx.fileName(), fromEnumMessageName, fromEnum.Name).WithOption().Location(),
				),
			)
			return
		}
	}
	if isDifferentType(fromType, toField.Type) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`cannot convert type automatically: field type is %q but specified value type is %q`,
					toField.Type.Kind.ToString(), fromType.Kind.ToString(),
				),
				builder.Location(),
			),
		)
		return
	}
}

func (r *Resolver) validateBindFieldEnumSelectorType(ctx *context, enumSelector *Message, toField *Field, builder *source.FieldBuilder) {
	for _, field := range enumSelector.Fields {
		if field.Type.Kind == types.Message && field.Type.Message.IsEnumSelector() {
			r.validateBindFieldEnumSelectorType(ctx, field.Type.Message, toField, builder)
		} else {
			r.validateBindFieldType(ctx, field.Type, toField, builder)
		}
	}
}

func (r *Resolver) validateBindFieldType(ctx *context, fromType *Type, toField *Field, builder *source.FieldBuilder) {
	if fromType == nil || toField == nil {
		return
	}
	toType := toField.Type
	if fromType.Kind == types.Message {
		if fromType.Message.IsEnumSelector() && toType.Kind == types.Enum {
			r.validateBindFieldEnumSelectorType(ctx, fromType.Message, toField, builder)
			return
		}
		if toType.Kind != types.Message {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`cannot convert message to %q`, toType.Kind.ToString()),
					builder.Location(),
				),
			)
			return
		}
		if fromType.Message == nil || toType.Message == nil {
			return
		}
		if fromType.IsNumberWrapper() && toType.IsNumberWrapper() {
			// If both are of the number type from google.protobuf.wrappers, they can be mutually converted.
			return
		}
		if fromType.Message.IsMapEntry && toType.Message.IsMapEntry {
			for _, name := range []string{"key", "value"} {
				fromMapType := fromType.Message.Field(name).Type
				toMapType := toType.Message.Field(name).Type
				if isDifferentType(fromMapType, toMapType) {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(
								`cannot convert type automatically: map %s type is %q but specified map %s type is %q`,
								name, toMapType.Kind.ToString(), name, fromMapType.Kind.ToString(),
							),
							builder.Location(),
						),
					)
				}
			}
			return
		}
		fromMessageName := fromType.Message.FQDN()
		toMessage := toType.Message
		toMessageName := toMessage.FQDN()
		if fromMessageName == toMessageName {
			// assignment of the same type is okay.
			return
		}
		if toMessage.Rule == nil || len(toMessage.Rule.Aliases) == 0 {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromMessageName, toMessageName, ctx.messageName(), toField.Name,
					),
					builder.Location(),
				),
			)
			return
		}
		var found bool
		for _, alias := range toMessage.Rule.Aliases {
			toMessageAliasName := alias.FQDN()
			if toMessageAliasName == fromMessageName {
				found = true
				break
			}
		}
		if !found {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.message option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromMessageName, toMessageName, ctx.messageName(), toField.Name,
					),
					source.NewMessageBuilder(toMessage.File.Name, toMessage.Name).
						WithOption().WithAlias().Location(),
				),
			)
			return
		}
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
		if toEnum.Rule == nil || len(toEnum.Rule.Aliases) == 0 {
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
		var found bool
		for _, alias := range toEnum.Rule.Aliases {
			toEnumAliasName := alias.FQDN()
			if toEnumAliasName == fromEnumName {
				found = true
				break
			}
		}
		if !found {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(
						`required specify alias = %q in grpc.federation.enum option for the %q type to automatically assign a value to the "%s.%s" field via autobind`,
						fromEnumName, toEnumName, ctx.messageName(), toField.Name,
					),
					source.NewEnumBuilder(ctx.fileName(), toEnumMessageName, toEnum.Name).WithOption().Location(),
				),
			)
			return
		}
	}
	if isDifferentType(fromType, toField.Type) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`cannot convert type automatically: field type is %q but specified value type is %q`,
					toField.Type.Kind.ToString(), fromType.Kind.ToString(),
				),
				builder.Location(),
			),
		)
		return
	}
}

func (r *Resolver) validateEnumMultipleAliases(fromType *Type, toField *Field) error {
	toType := toField.Type
	if toType.Kind != types.Enum {
		return nil
	}
	if toType.Enum == nil || toType.Enum.Rule == nil {
		return nil
	}
	if len(toType.Enum.Rule.Aliases) <= 1 {
		return nil
	}

	if fromType == nil {
		return nil
	}
	if fromType.Kind == types.Message && fromType.Message != nil && fromType.Message.IsEnumSelector() {
		return nil
	}
	return errors.New(`if multiple aliases are specified, you must use grpc.federation.enum.select function to bind`)
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

func (r *Resolver) resolveServiceRule(ctx *context, def *federation.ServiceRule, builder *source.ServiceOptionBuilder) *ServiceRule {
	if def == nil {
		return nil
	}
	return &ServiceRule{
		Env: r.resolveEnv(ctx, def.GetEnv(), builder.WithEnv()),
	}
}

func (r *Resolver) resolveEnv(ctx *context, def *federation.Env, builder *source.EnvBuilder) *Env {
	if def == nil {
		return nil
	}
	if def.GetMessage() != "" && len(def.GetVar()) != 0 {
		ctx.addError(
			ErrWithLocation(
				`"message" and "var" cannot be used simultaneously`,
				builder.Location(),
			),
		)
		return nil
	}

	var vars []*EnvVar
	if msgName := def.GetMessage(); msgName != "" {
		msgBuilder := builder.WithMessage()
		var msg *Message
		if strings.Contains(msgName, ".") {
			pkg, err := r.lookupPackage(msgName)
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						msgBuilder.Location(),
					),
				)
				return nil
			}
			name := r.trimPackage(pkg, msgName)
			msg = r.resolveMessage(ctx, pkg, name, source.ToLazyMessageBuilder(msgBuilder, name))
		} else {
			msg = r.resolveMessage(ctx, ctx.file().Package, msgName, source.ToLazyMessageBuilder(msgBuilder, msgName))
		}
		if msg == nil {
			return nil
		}
		for _, field := range msg.Fields {
			vars = append(vars, r.resolveEnvVarWithField(field))
		}
	} else {
		for idx, v := range def.GetVar() {
			ev := r.resolveEnvVar(ctx, v, builder.WithVar(idx))
			if ev == nil {
				continue
			}
			vars = append(vars, ev)
		}
	}
	return &Env{
		Vars: vars,
	}
}

func (r *Resolver) resolveEnvVar(ctx *context, def *federation.EnvVar, builder *source.EnvVarBuilder) *EnvVar {
	name := def.GetName()
	if name == "" {
		ctx.addError(
			ErrWithLocation(
				`"name" is required`,
				builder.WithName().Location(),
			),
		)
		return nil
	}
	if err := r.validateName(name); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithName().Location(),
			),
		)
		return nil
	}
	envType := def.GetType()
	if envType == nil {
		ctx.addError(
			ErrWithLocation(
				`"type" is required`,
				builder.WithType().Location(),
			),
		)
		return nil
	}
	typ, err := r.resolveEnvType(envType, false)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithType().Location(),
			),
		)
		return nil
	}
	if typ.Message != nil && typ.Message.IsMapEntry {
		mp := typ.Message
		mp.Name = cases.Title(language.Und).String(name) + "Entry"
		file := ctx.file()
		copied := *file
		copied.Package = &Package{Name: federation.PrivatePackageName}
		mp.File = &copied
		r.cachedMessageMap[mp.FQDN()] = mp
	}
	return &EnvVar{
		Name:   name,
		Type:   typ,
		Option: r.resolveEnvVarOption(def.GetOption()),
	}
}

func (r *Resolver) resolveEnvVarWithField(field *Field) *EnvVar {
	var opt *EnvVarOption
	if field.Rule != nil {
		opt = field.Rule.Env
	}
	return &EnvVar{
		Name:   field.Name,
		Type:   field.Type,
		Option: opt,
	}
}

func (r *Resolver) resolveEnvType(def *federation.EnvType, repeated bool) (*Type, error) {
	switch def.Type.(type) {
	case *federation.EnvType_Kind:
		switch def.GetKind() {
		case federation.TypeKind_STRING:
			if repeated {
				return StringRepeatedType, nil
			}
			return StringType, nil
		case federation.TypeKind_BOOL:
			if repeated {
				return BoolRepeatedType, nil
			}
			return BoolType, nil
		case federation.TypeKind_INT64:
			if repeated {
				return Int64RepeatedType, nil
			}
			return Int64Type, nil
		case federation.TypeKind_UINT64:
			if repeated {
				return Uint64RepeatedType, nil
			}
			return Uint64Type, nil
		case federation.TypeKind_DOUBLE:
			if repeated {
				return DoubleRepeatedType, nil
			}
			return DoubleType, nil
		case federation.TypeKind_DURATION:
			if repeated {
				return DurationRepeatedType, nil
			}
			return DurationType, nil
		}
	case *federation.EnvType_Repeated:
		return r.resolveEnvType(def.GetRepeated(), true)
	case *federation.EnvType_Map:
		mapType := def.GetMap()
		key, err := r.resolveEnvType(mapType.GetKey(), false)
		if err != nil {
			return nil, err
		}
		value, err := r.resolveEnvType(mapType.GetValue(), false)
		if err != nil {
			return nil, err
		}
		return NewMapType(key, value), nil
	}
	return nil, fmt.Errorf("failed to resolve env type")
}

func (r *Resolver) resolveEnvVarOption(def *federation.EnvVarOption) *EnvVarOption {
	if def == nil {
		return nil
	}
	return &EnvVarOption{
		Alternate: def.GetAlternate(),
		Default:   def.GetDefault(),
		Required:  def.GetRequired(),
		Ignored:   def.GetIgnored(),
	}
}

func (r *Resolver) resolveMethodRule(ctx *context, def *federation.MethodRule, builder *source.MethodBuilder) *MethodRule {
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
					builder.WithOption().WithTimeout().Location(),
				),
			)
		} else {
			rule.Timeout = &duration
		}
	}
	return rule
}

func (r *Resolver) resolveMessageRule(ctx *context, msg *Message, ruleDef *federation.MessageRule, builder *source.MessageOptionBuilder) {
	if ruleDef == nil {
		return
	}
	msg.Rule = &MessageRule{
		DefSet: &VariableDefinitionSet{
			Defs: r.resolveVariableDefinitions(ctx, ruleDef.GetDef(), func(idx int) *source.VariableDefinitionOptionBuilder {
				return builder.WithDef(idx)
			}),
		},
		CustomResolver: ruleDef.GetCustomResolver(),
		Aliases:        r.resolveMessageAliases(ctx, ruleDef.GetAlias(), builder),
	}
}

func (r *Resolver) resolveMessageAliases(ctx *context, aliasNames []string, builder *source.MessageOptionBuilder) []*Message {
	if len(aliasNames) == 0 {
		return nil
	}
	var ret []*Message
	for _, aliasName := range aliasNames {
		if strings.Contains(aliasName, ".") {
			pkg, err := r.lookupPackage(aliasName)
			if err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.WithAlias().Location(),
					),
				)
				return nil
			}
			name := r.trimPackage(pkg, aliasName)
			if alias := r.resolveMessage(ctx, pkg, name, source.ToLazyMessageBuilder(builder, name)); alias != nil {
				ret = append(ret, alias)
			}
		} else {
			if alias := r.resolveMessage(ctx, ctx.file().Package, aliasName, source.ToLazyMessageBuilder(builder, aliasName)); alias != nil {
				ret = append(ret, alias)
			}
		}
	}
	return ret
}

func (r *Resolver) resolveVariableDefinitions(ctx *context, varDefs []*federation.VariableDefinition, builderFn func(idx int) *source.VariableDefinitionOptionBuilder) []*VariableDefinition {
	ctx.clearVariableDefinitions()

	var ret []*VariableDefinition
	for idx, varDef := range varDefs {
		ctx := ctx.withDefIndex(idx)
		vd := r.resolveVariableDefinition(ctx, varDef, builderFn(idx))
		ctx.addVariableDefinition(vd)
		ret = append(ret, vd)
	}
	return ret
}

func (r *Resolver) resolveVariableDefinition(ctx *context, varDef *federation.VariableDefinition, builder *source.VariableDefinitionOptionBuilder) *VariableDefinition {
	var ifValue *CELValue
	if varDef.GetIf() != "" {
		ifValue = &CELValue{Expr: varDef.GetIf()}
	}
	return &VariableDefinition{
		Idx:      ctx.defIndex(),
		Name:     r.resolveVariableName(ctx, varDef.GetName(), builder),
		If:       ifValue,
		AutoBind: varDef.GetAutobind(),
		Expr:     r.resolveVariableExpr(ctx, varDef, builder),
		builder:  builder,
	}
}

func (r *Resolver) resolveVariableName(ctx *context, name string, builder *source.VariableDefinitionOptionBuilder) string {
	if !ctx.ignoreNameValidation {
		if err := r.validateName(name); err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				builder.WithName().Location(),
			))
			return ""
		}
	}
	if name != "" {
		return name
	}
	return fmt.Sprintf("_def%d", ctx.defIndex())
}

func (r *Resolver) resolveVariableExpr(ctx *context, varDef *federation.VariableDefinition, builder *source.VariableDefinitionOptionBuilder) *VariableExpr {
	switch varDef.GetExpr().(type) {
	case *federation.VariableDefinition_By:
		return &VariableExpr{By: &CELValue{Expr: varDef.GetBy()}}
	case *federation.VariableDefinition_Map:
		return &VariableExpr{Map: r.resolveMapExpr(ctx, varDef.GetMap(), builder.WithMap())}
	case *federation.VariableDefinition_Message:
		return &VariableExpr{Message: r.resolveMessageExpr(ctx, varDef.GetMessage(), builder.WithMessage())}
	case *federation.VariableDefinition_Call:
		return &VariableExpr{Call: r.resolveCallExpr(ctx, varDef.GetCall(), builder.WithCall())}
	case *federation.VariableDefinition_Validation:
		return &VariableExpr{Validation: r.resolveValidationExpr(ctx, varDef.GetValidation(), builder.WithValidation())}
	}
	return nil
}

func (r *Resolver) resolveMapExpr(ctx *context, def *federation.MapExpr, builder *source.MapExprOptionBuilder) *MapExpr {
	if def == nil {
		return nil
	}
	return &MapExpr{
		Iterator: r.resolveMapIterator(ctx, def.GetIterator(), builder),
		Expr:     r.resolveMapIteratorExpr(ctx, def, builder),
	}
}

func (r *Resolver) resolveMapIterator(ctx *context, def *federation.Iterator, builder *source.MapExprOptionBuilder) *Iterator {
	name := def.GetName()
	if name == "" {
		ctx.addError(
			ErrWithLocation(
				"map iterator name must be specified",
				builder.WithIteratorName().Location(),
			),
		)
	}
	var srcDef *VariableDefinition
	if src := def.GetSrc(); src == "" {
		ctx.addError(
			ErrWithLocation(
				"map iterator src must be specified",
				builder.WithIteratorSource().Location(),
			),
		)
	} else {
		srcDef = ctx.variableDef(src)
		if srcDef == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q variable is not defined`, src),
					builder.WithIteratorSource().Location(),
				),
			)
		}
	}
	return &Iterator{
		Name:   name,
		Source: srcDef,
	}
}

func (r *Resolver) resolveMapIteratorExpr(ctx *context, def *federation.MapExpr, builder *source.MapExprOptionBuilder) *MapIteratorExpr {
	switch def.GetExpr().(type) {
	case *federation.MapExpr_By:
		return &MapIteratorExpr{By: &CELValue{Expr: def.GetBy()}}
	case *federation.MapExpr_Message:
		return &MapIteratorExpr{Message: r.resolveMessageExpr(ctx, def.GetMessage(), builder.WithMessage())}
	}
	return nil
}

func (r *Resolver) resolveMessageExpr(ctx *context, def *federation.MessageExpr, builder *source.MessageExprOptionBuilder) *MessageExpr {
	if def == nil {
		return nil
	}
	msg, err := r.resolveMessageByName(ctx, def.GetName(), source.ToLazyMessageBuilder(builder, def.GetName()))
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithName().Location(),
			),
		)
	}
	if ctx.msg == msg {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`recursive definition: %q is own message name`, msg.Name),
				builder.WithName().Location(),
			),
		)
	}
	args := make([]*Argument, 0, len(def.GetArgs()))
	for idx, argDef := range def.GetArgs() {
		args = append(args, r.resolveMessageExprArgument(ctx, argDef, builder.WithArgs(idx)))
	}
	return &MessageExpr{
		Message: msg,
		Args:    args,
	}
}

func (r *Resolver) resolveCallExpr(ctx *context, def *federation.CallExpr, builder *source.CallExprOptionBuilder) *CallExpr {
	if def == nil {
		return nil
	}

	pkgName, serviceName, methodName, err := r.splitMethodFullName(ctx.file().Package, def.GetMethod())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithMethod().Location(),
			),
		)
		return nil
	}
	pkg, exists := r.protoPackageNameToPackage[pkgName]
	if !exists {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q package does not exist`, pkgName),
				builder.WithMethod().Location(),
			),
		)
		return nil
	}
	service := r.resolveService(ctx, pkg, serviceName, source.NewServiceBuilder(ctx.fileName(), serviceName))
	if service == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`cannot find %q method because the service to which the method belongs does not exist`, methodName),
				builder.WithMethod().Location(),
			),
		)
		return nil
	}

	method := service.Method(methodName)
	if method == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q method does not exist in %s service`, methodName, service.Name),
				builder.WithMethod().Location(),
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
					builder.WithTimeout().Location(),
				),
			)
		} else {
			timeout = &duration
		}
	}

	var grpcErrs []*GRPCError
	for idx, grpcErr := range def.GetError() {
		grpcErrs = append(grpcErrs, r.resolveGRPCError(ctx, grpcErr, builder.WithError(idx)))
	}
	return &CallExpr{
		Method:  method,
		Request: r.resolveRequest(ctx, method, def.GetRequest(), builder),
		Timeout: timeout,
		Retry:   r.resolveRetry(ctx, def.GetRetry(), timeout, builder.WithRetry()),
		Errors:  grpcErrs,
	}
}

func (r *Resolver) resolveValidationExpr(ctx *context, def *federation.ValidationExpr, builder *source.ValidationExprOptionBuilder) *ValidationExpr {
	return &ValidationExpr{
		Name:  def.GetName(),
		Error: r.resolveGRPCError(ctx, def.GetError(), builder.WithError()),
	}
}

func (r *Resolver) resolveGRPCError(ctx *context, def *federation.GRPCError, builder *source.GRPCErrorOptionBuilder) *GRPCError {
	var (
		msg               *CELValue
		ignoreAndResponse *CELValue
	)
	if m := def.GetMessage(); m != "" {
		msg = &CELValue{Expr: m}
	}
	if res := def.GetIgnoreAndResponse(); res != "" {
		if def.GetIgnore() {
			ctx.addError(
				ErrWithLocation(
					`cannot set both "ignore" and "ignore_and_response"`,
					builder.WithIgnore().Location(),
				),
			)
			ctx.addError(
				ErrWithLocation(
					`cannot set both "ignore" and "ignore_and_response"`,
					builder.WithIgnoreAndResponse().Location(),
				),
			)
		}
		ignoreAndResponse = &CELValue{Expr: res}
	}
	return &GRPCError{
		DefSet: &VariableDefinitionSet{
			Defs: r.resolveVariableDefinitions(ctx, def.GetDef(), func(idx int) *source.VariableDefinitionOptionBuilder {
				return builder.WithDef(idx)
			}),
		},
		If:                r.resolveGRPCErrorIf(def.GetIf()),
		Code:              def.GetCode(),
		Message:           msg,
		Details:           r.resolveGRPCErrorDetails(ctx, def.GetDetails(), builder),
		Ignore:            def.GetIgnore(),
		IgnoreAndResponse: ignoreAndResponse,
	}
}

func (r *Resolver) resolveGRPCErrorIf(expr string) *CELValue {
	if expr == "" {
		return &CELValue{Expr: "true"}
	}
	return &CELValue{Expr: expr}
}

func (r *Resolver) resolveGRPCErrorDetails(ctx *context, details []*federation.GRPCErrorDetail, builder *source.GRPCErrorOptionBuilder) []*GRPCErrorDetail {
	if len(details) == 0 {
		return nil
	}
	result := make([]*GRPCErrorDetail, 0, len(details))
	for idx, detail := range details {
		ctx := ctx.withErrDetailIndex(idx)
		builder := builder.WithDetail(idx)
		result = append(result, &GRPCErrorDetail{
			If: r.resolveGRPCErrorIf(detail.GetIf()),
			DefSet: &VariableDefinitionSet{
				Defs: r.resolveVariableDefinitions(ctx, detail.GetDef(), func(idx int) *source.VariableDefinitionOptionBuilder {
					return builder.WithDef(idx)
				}),
			},
			Messages: &VariableDefinitionSet{
				Defs: r.resolveGRPCDetailMessages(ctx, detail.GetMessage(), func(idx int) *source.VariableDefinitionOptionBuilder {
					return builder.WithMessage(idx)
				}),
			},
			PreconditionFailures: r.resolvePreconditionFailures(detail.GetPreconditionFailure()),
			BadRequests:          r.resolveBadRequests(detail.GetBadRequest()),
			LocalizedMessages:    r.resolveLocalizedMessages(detail.GetLocalizedMessage()),
		})
	}
	return result
}

func (r *Resolver) resolveGRPCDetailMessages(ctx *context, messages []*federation.MessageExpr, builderFn func(int) *source.VariableDefinitionOptionBuilder) VariableDefinitions {
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
	ctx = ctx.withIgnoreNameValidation()
	defs := make([]*VariableDefinition, 0, len(msgs))
	for _, def := range r.resolveVariableDefinitions(ctx, msgs, builderFn) {
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

func (r *Resolver) resolveFieldRule(ctx *context, msg *Message, field *Field, ruleDef *federation.FieldRule, builder *source.FieldBuilder) *FieldRule {
	if ruleDef == nil {
		if msg.Rule == nil {
			return nil
		}
		if !msg.Rule.CustomResolver && len(msg.Rule.Aliases) == 0 {
			return nil
		}
		ret := &FieldRule{}
		if msg.Rule.CustomResolver {
			ret.MessageCustomResolver = true
		}
		for _, msgAlias := range msg.Rule.Aliases {
			fieldAlias := r.resolveFieldAlias(ctx, msg, field, "", msgAlias, builder)
			if fieldAlias == nil {
				return nil
			}
			ret.Aliases = append(ret.Aliases, fieldAlias)
		}
		return ret
	}
	oneof := r.resolveFieldOneofRule(ctx, field, ruleDef.GetOneof(), builder)
	envOpt := r.resolveEnvVarOption(ruleDef.GetEnv())
	var value *Value
	if oneof == nil && envOpt == nil {
		v, err := r.resolveValue(fieldRuleToCommonValueDef(ruleDef))
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.Location(),
				),
			)
			return nil
		}
		value = v
	}
	if ruleDef.GetAlias() != "" && (msg.Rule == nil || len(msg.Rule.Aliases) == 0) {
		ctx.addError(
			ErrWithLocation(
				`use "alias" in "grpc.federation.field" option, but "alias" is not defined in "grpc.federation.message" option`,
				builder.Location(),
			),
		)
	}

	var aliases []*Field
	if msg.Rule != nil {
		for _, msgAlias := range msg.Rule.Aliases {
			if alias := r.resolveFieldAlias(ctx, msg, field, ruleDef.GetAlias(), msgAlias, builder); alias != nil {
				aliases = append(aliases, alias)
			}
		}
	}
	return &FieldRule{
		Value:          value,
		CustomResolver: ruleDef.GetCustomResolver(),
		Aliases:        aliases,
		Oneof:          oneof,
		Env:            envOpt,
	}
}

func (r *Resolver) resolveFieldOneofRule(ctx *context, field *Field, def *federation.FieldOneof, builder *source.FieldBuilder) *FieldOneofRule {
	if def == nil {
		return nil
	}
	if field.Oneof == nil {
		ctx.addError(
			ErrWithLocation(
				`"oneof" feature can only be used for fields within oneof`,
				builder.Location(),
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
	return &FieldOneofRule{
		If:      ifValue,
		Default: def.GetDefault(),
		DefSet: &VariableDefinitionSet{
			Defs: r.resolveVariableDefinitions(ctx, def.GetDef(), func(idx int) *source.VariableDefinitionOptionBuilder {
				return builder.WithOption().WithOneOf().WithDef(idx)
			}),
		},
		By: by,
	}
}

func (r *Resolver) resolveFieldRuleByAutoAlias(ctx *context, msg *Message, field *Field, alias *Message, builder *source.FieldBuilder) *Field {
	if alias == nil {
		return nil
	}
	aliasField := alias.Field(field.Name)
	if aliasField == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`specified "alias" in grpc.federation.message option, but %q field does not exist in %q message`,
					field.Name, alias.FQDN(),
				),
				builder.Location(),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if isDifferentType(field.Type, aliasField.Type) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, field.Type.Kind.ToString(),
					aliasField.Message.FQDN(), aliasField.Type.Kind.ToString(),
				),
				builder.Location(),
			),
		)
		return nil
	}
	return aliasField
}

func (r *Resolver) resolveFieldAlias(ctx *context, msg *Message, field *Field, fieldAlias string, alias *Message, builder *source.FieldBuilder) *Field {
	if fieldAlias == "" {
		return r.resolveFieldRuleByAutoAlias(ctx, msg, field, alias, builder)
	}
	aliasField := alias.Field(fieldAlias)
	if aliasField == nil {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q field does not exist in %q message`, fieldAlias, alias.FQDN()),
				builder.Location(),
			),
		)
		return nil
	}
	if field.Type == nil || aliasField.Type == nil {
		return nil
	}
	if isDifferentType(field.Type, aliasField.Type) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`The types of %q's %q field (%q) and %q's field (%q) are different. This field cannot be resolved automatically, so you must use the "grpc.federation.field" option to bind it yourself`,
					msg.FQDN(), field.Name, field.Type.Kind.ToString(),
					alias.FQDN(), aliasField.Type.Kind.ToString(),
				),
				builder.Location(),
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
		Aliases: r.resolveEnumAliases(ctx, ruleDef.GetAlias()),
	}
}

func (r *Resolver) resolveEnumAliases(ctx *context, aliasNames []string) []*Enum {
	if len(aliasNames) == 0 {
		return nil
	}
	var ret []*Enum
	for _, aliasName := range aliasNames {
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
			if alias := r.resolveEnum(ctx, pkg, name, source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), name)); alias != nil {
				ret = append(ret, alias)
			}
		} else {
			if alias := r.resolveEnum(ctx, ctx.file().Package, aliasName, source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), aliasName)); alias != nil {
				ret = append(ret, alias)
			}
		}
	}
	return ret
}

func (r *Resolver) resolveEnumValueRule(ctx *context, enum *Enum, enumValue *EnumValue, ruleDef *federation.EnumValueRule) *EnumValueRule {
	if ruleDef == nil {
		if enum.Rule == nil {
			return nil
		}
		return &EnumValueRule{
			Aliases: r.resolveEnumValueAlias(ctx, enumValue.Value, "", enum.Rule.Aliases),
		}
	}
	if len(ruleDef.GetAlias()) != 0 && (enum.Rule == nil || len(enum.Rule.Aliases) == 0) {
		ctx.addError(
			ErrWithLocation(
				`use "alias" in "grpc.federation.enum_value" option, but "alias" is not defined in "grpc.federation.enum" option`,
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).
					WithValue(enumValue.Value).
					WithOption().WithAlias().Location(),
			),
		)
		return nil
	}

	defaultValue := ruleDef.GetDefault()
	var aliases []*EnumValueAlias
	if enum.Rule != nil && !defaultValue {
		for _, aliasName := range ruleDef.GetAlias() {
			aliases = append(aliases, r.resolveEnumValueAlias(ctx, enumValue.Value, aliasName, enum.Rule.Aliases)...)
		}
	}
	return &EnumValueRule{
		Default: defaultValue,
		Aliases: aliases,
	}
}

func (r *Resolver) resolveEnumValueAlias(ctx *context, enumValueName, enumValueAlias string, aliases []*Enum) []*EnumValueAlias {
	if enumValueAlias == "" {
		return r.resolveEnumValueRuleByAutoAlias(ctx, enumValueName, aliases)
	}
	aliasFQDNs := make([]string, 0, len(aliases))
	var enumValueAliases []*EnumValueAlias
	for _, alias := range aliases {
		if value := alias.Value(enumValueAlias); value != nil {
			enumValueAliases = append(enumValueAliases, &EnumValueAlias{
				EnumAlias: alias,
				Aliases:   []*EnumValue{value},
			})
		}
		aliasFQDNs = append(aliasFQDNs, fmt.Sprintf("%q", alias.FQDN()))
	}
	if len(enumValueAliases) == 0 {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q value does not exist in %s enum`, enumValueAlias, strings.Join(aliasFQDNs, ", ")),
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithValue(enumValueName).Location(),
			),
		)
		return nil
	}

	hasPrefix := strings.Contains(enumValueAlias, ".")
	if !hasPrefix && len(enumValueAliases) != len(aliases) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(`%q value must be present in all enums, but it is missing in %s enum`, enumValueAlias, strings.Join(aliasFQDNs, ", ")),
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithValue(enumValueName).Location(),
			),
		)
		return nil
	}
	return enumValueAliases
}

func (r *Resolver) resolveEnumValueRuleByAutoAlias(ctx *context, enumValueName string, aliases []*Enum) []*EnumValueAlias {
	aliasFQDNs := make([]string, 0, len(aliases))
	var enumValueAliases []*EnumValueAlias
	for _, alias := range aliases {
		if value := alias.Value(enumValueName); value != nil {
			enumValueAliases = append(enumValueAliases, &EnumValueAlias{
				EnumAlias: alias,
				Aliases:   []*EnumValue{value},
			})
		}
		aliasFQDNs = append(aliasFQDNs, fmt.Sprintf("%q", alias.FQDN()))
	}
	if len(enumValueAliases) != len(aliases) {
		ctx.addError(
			ErrWithLocation(
				fmt.Sprintf(
					`specified "alias" in grpc.federation.enum option, but %q value does not exist in %s enum`,
					enumValueName, strings.Join(aliasFQDNs, ", "),
				),
				source.NewEnumBuilder(ctx.fileName(), ctx.messageName(), ctx.enumName()).WithValue(enumValueName).Location(),
			),
		)
		return nil
	}
	return enumValueAliases
}

func (r *Resolver) resolveFields(ctx *context, fieldsDef []*descriptorpb.FieldDescriptorProto, oneofs []*Oneof, builder *source.MessageBuilder) []*Field {
	if len(fieldsDef) == 0 {
		return nil
	}
	fields := make([]*Field, 0, len(fieldsDef))
	for _, fieldDef := range fieldsDef {
		field := r.resolveField(ctx, fieldDef, oneofs, builder.WithField(fieldDef.GetName()))
		if field == nil {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func (r *Resolver) resolveField(ctx *context, fieldDef *descriptorpb.FieldDescriptorProto, oneofs []*Oneof, builder *source.FieldBuilder) *Field {
	typ, err := r.resolveType(ctx, fieldDef.GetTypeName(), types.Kind(fieldDef.GetType()), fieldDef.GetLabel())
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
		return nil
	}
	field := &Field{Name: fieldDef.GetName(), Type: typ, Message: ctx.msg}
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
				builder.Location(),
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
		msg = r.resolveMessage(ctx, pkg, name, source.NewMessageBuilder(ctx.fileName(), typeName))
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
		enum = r.resolveEnum(ctx, pkg, name, source.NewEnumBuilder(ctx.fileName(), "", name))
	}
	repeated := label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	if msg != nil && msg.IsMapEntry {
		repeated = false
	}
	return &Type{
		Kind:     kind,
		Repeated: repeated,
		Message:  msg,
		Enum:     enum,
	}, nil
}

func (r *Resolver) resolveRetry(ctx *context, def *federation.RetryPolicy, timeout *time.Duration, builder *source.RetryOptionBuilder) *RetryPolicy {
	if def == nil {
		return nil
	}
	ifExpr := "true"
	if cond := def.GetIf(); cond != "" {
		ifExpr = cond
	}
	return &RetryPolicy{
		If:          &CELValue{Expr: ifExpr},
		Constant:    r.resolveRetryConstant(ctx, def.GetConstant(), builder),
		Exponential: r.resolveRetryExponential(ctx, def.GetExponential(), timeout, builder),
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

func (r *Resolver) resolveRetryConstant(ctx *context, def *federation.RetryPolicyConstant, builder *source.RetryOptionBuilder) *RetryPolicyConstant {
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
					builder.WithConstantInterval().Location(),
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

func (r *Resolver) resolveRetryExponential(ctx *context, def *federation.RetryPolicyExponential, timeout *time.Duration, builder *source.RetryOptionBuilder) *RetryPolicyExponential {
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
					builder.WithExponentialInitialInterval().Location(),
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
					builder.WithExponentialMaxInterval().Location(),
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

func (r *Resolver) resolveRequest(ctx *context, method *Method, requestDef []*federation.MethodRequest, builder *source.CallExprOptionBuilder) *Request {
	reqMsg := method.Request
	args := make([]*Argument, 0, len(requestDef))
	for idx, req := range requestDef {
		fieldName := req.GetField()
		var argType *Type
		if !reqMsg.HasField(fieldName) {
			ctx.addError(ErrWithLocation(
				fmt.Sprintf(`%q field does not exist in %q message for method request`, fieldName, reqMsg.FQDN()),
				builder.WithRequest(idx).WithField().Location(),
			))
		} else {
			argType = reqMsg.Field(fieldName).Type
		}
		value, err := r.resolveValue(methodRequestToCommonValueDef(req))
		if err != nil {
			ctx.addError(ErrWithLocation(
				err.Error(),
				builder.WithRequest(idx).WithBy().Location(),
			))
		}
		var ifValue *CELValue
		if req.GetIf() != "" {
			ifValue = &CELValue{Expr: req.GetIf()}
		}
		if argType != nil && argType.OneofField != nil && ifValue == nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`%q field is a oneof field, so you need to specify an "if" expression`, fieldName),
					builder.WithRequest(idx).WithField().Location(),
				),
			)
		}
		args = append(args, &Argument{
			Name:  fieldName,
			Type:  argType,
			Value: value,
			If:    ifValue,
		})
	}
	return &Request{Args: args, Type: reqMsg}
}

func (r *Resolver) resolveMessageExprArgument(ctx *context, argDef *federation.Argument, builder *source.ArgumentOptionBuilder) *Argument {
	value, err := r.resolveValue(argumentToCommonValueDef(argDef))
	if err != nil {
		switch {
		case argDef.GetBy() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithBy().Location(),
				),
			)
		case argDef.GetInline() != "":
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithInline().Location(),
				),
			)
		}
	}
	name := argDef.GetName()
	if err := r.validateName(name); err != nil {
		ctx.addError(ErrWithLocation(
			err.Error(),
			builder.WithName().Location(),
		))
	}
	return &Argument{
		Name:  name,
		Value: value,
	}
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
	builder := newMessageBuilderFromMessage(msg)
	arg := msg.Rule.MessageArgument
	fileDesc := messageArgumentFileDescriptor(arg)
	if err := r.celRegistry.RegisterFiles(append(r.files, fileDesc)...); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
		return nil
	}
	env, err := r.createCELEnv(msg)
	if err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.Location(),
			),
		)
		return nil
	}
	r.resolveMessageCELValues(ctx.withFile(msg.File).withMessage(msg), env, msg, builder)

	msgs := []*Message{msg}
	msgToDefsMap := msg.Rule.DefSet.MessageToDefsMap()
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for k, v := range field.Rule.Oneof.DefSet.MessageToDefsMap() {
			msgToDefsMap[k] = append(msgToDefsMap[k], v...)
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
		}
		defs := msgToDefsMap[depMsg]
		depMsgArg.Fields = append(depMsgArg.Fields, r.resolveMessageArgumentFields(ctx, depMsgArg, defs)...)
		for _, field := range depMsgArg.Fields {
			field.Message = depMsgArg
		}
		m := r.resolveMessageArgumentRecursive(ctx, child)
		msgs = append(msgs, m...)
	}
	return msgs
}

func (r *Resolver) resolveMessageArgumentFields(ctx *context, arg *Message, defs []*VariableDefinition) []*Field {
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

	evaluatedArgNameMap := make(map[string]*Type)
	// First, evaluate the fields that are already registered in the message argument.
	for _, field := range arg.Fields {
		evaluatedArgNameMap[field.Name] = field.Type
		argNameMap[field.Name] = struct{}{}
	}
	var fields []*Field
	for _, varDef := range defs {
		for _, msgExpr := range varDef.MessageExprs() {
			r.validateMessageDependencyArgumentName(ctx, argNameMap, varDef)
			for argIdx, arg := range msgExpr.Args {
				if arg.Value == nil {
					continue
				}
				fieldType := arg.Value.Type()
				if fieldType == nil {
					continue
				}
				if typ, exists := evaluatedArgNameMap[arg.Name]; exists {
					if isDifferentType(typ, fieldType) {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(
									"%q argument name is declared with a different type kind. found %q and %q type",
									arg.Name,
									typ.Kind.ToString(),
									fieldType.Kind.ToString(),
								),
								varDef.builder.WithMessage().
									WithArgs(argIdx).Location(),
							),
						)
					}
					continue
				}
				if arg.Value.CEL != nil && arg.Value.Inline {
					if fieldType.Kind != types.Message {
						ctx.addError(
							ErrWithLocation(
								"inline value is not message type",
								varDef.builder.WithMessage().
									WithArgs(argIdx).
									WithInline().Location(),
							),
						)
						continue
					}
					for _, field := range fieldType.Message.Fields {
						if typ, exists := evaluatedArgNameMap[field.Name]; exists {
							if isDifferentType(typ, field.Type) {
								ctx.addError(
									ErrWithLocation(
										fmt.Sprintf(
											"%q argument name is declared with a different type kind. found %q and %q type",
											field.Name,
											typ.Kind.ToString(),
											field.Type.Kind.ToString(),
										),
										varDef.builder.WithMessage().
											WithArgs(argIdx).Location(),
									),
								)
							}
							continue
						}
						fields = append(fields, field)
						evaluatedArgNameMap[field.Name] = field.Type
					}
				} else {
					fields = append(fields, &Field{
						Name: arg.Name,
						Type: fieldType,
					})
					evaluatedArgNameMap[arg.Name] = fieldType
				}
			}
		}
	}
	return fields
}

func (r *Resolver) validateMessageDependencyArgumentName(ctx *context, argNameMap map[string]struct{}, def *VariableDefinition) {
	for _, msgExpr := range def.MessageExprs() {
		curDepArgNameMap := make(map[string]struct{})
		for _, arg := range msgExpr.Args {
			if arg.Name != "" {
				curDepArgNameMap[arg.Name] = struct{}{}
			}
			if arg.Value.CEL != nil && arg.Value.Inline {
				fieldType := arg.Value.Type()
				if fieldType != nil && fieldType.Message != nil {
					for _, field := range fieldType.Message.Fields {
						curDepArgNameMap[field.Name] = struct{}{}
					}
				}
			}
		}
		for name := range argNameMap {
			if _, exists := curDepArgNameMap[name]; exists {
				continue
			}
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf("%q argument is defined in other message dependency arguments, but not in this context", name),
					def.builder.WithMessage().WithArgs(0).Location(),
				),
			)
		}
	}
}

func (r *Resolver) resolveMessageCELValues(ctx *context, env *cel.Env, msg *Message, builder *source.MessageBuilder) {
	if msg.Rule == nil {
		return
	}
	for idx, def := range msg.Rule.DefSet.Definitions() {
		ctx := ctx.withDefIndex(idx)
		env = r.resolveVariableDefinitionCELValues(ctx, env, def, def.builder)
	}
	for _, field := range msg.Fields {
		if !field.HasRule() {
			continue
		}

		fieldBuilder := builder.WithField(field.Name).WithOption()

		if field.Rule.Value != nil {
			if err := r.resolveCELValue(ctx, env, field.Rule.Value.CEL); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						fieldBuilder.WithBy().Location(),
					),
				)
			}
		}
		if field.Rule.Oneof != nil {
			fieldEnv, _ := env.Extend()
			oneof := field.Rule.Oneof
			oneofBuilder := fieldBuilder.WithOneOf()
			if oneof.If != nil {
				if err := r.resolveCELValue(ctx, fieldEnv, oneof.If); err != nil {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							oneofBuilder.WithIf().Location(),
						),
					)
				}
				if oneof.If.Out != nil {
					if oneof.If.Out.Kind != types.Bool {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`return value of "if" must be bool type but got %s type`, oneof.If.Out.Kind.ToString()),
								oneofBuilder.WithIf().Location(),
							),
						)
					}
				}
			}
			for idx, varDef := range oneof.DefSet.Definitions() {
				fieldEnv = r.resolveVariableDefinitionCELValues(ctx, fieldEnv, varDef, oneofBuilder.WithDef(idx))
			}
			if err := r.resolveCELValue(ctx, fieldEnv, oneof.By); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						oneofBuilder.WithBy().Location(),
					),
				)
			}
		}
	}

	r.resolveUsedNameReference(msg)
}

func (r *Resolver) resolveVariableDefinitionCELValues(ctx *context, env *cel.Env, def *VariableDefinition, builder *source.VariableDefinitionOptionBuilder) *cel.Env {
	if def.If != nil {
		if err := r.resolveCELValue(ctx, env, def.If); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithIf().Location(),
				),
			)
		}
		if def.If.Out != nil {
			if def.If.Out.Kind != types.Bool {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`return value of "if" must be bool type but got %s type`, def.If.Out.Kind.ToString()),
						builder.WithIf().Location(),
					),
				)
			}
		}
	}
	if def.Expr == nil {
		return env
	}
	r.resolveVariableExprCELValues(ctx, env, def.Expr, builder)
	if def.Name != "" && def.Expr.Type != nil {
		newEnv, err := env.Extend(cel.Variable(def.Name, ToCELType(def.Expr.Type)))
		if err != nil {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`failed to extend cel.Env from variables of messages: %s`, err.Error()),
					builder.Location(),
				),
			)
			return env
		}
		return newEnv
	}
	return env
}

func (r *Resolver) resolveVariableExprCELValues(ctx *context, env *cel.Env, expr *VariableExpr, builder *source.VariableDefinitionOptionBuilder) {
	switch {
	case expr.By != nil:
		if err := r.resolveCELValue(ctx, env, expr.By); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithBy().Location(),
				),
			)
		}
		expr.Type = expr.By.Out
	case expr.Map != nil:
		mapEnv := env
		iter := expr.Map.Iterator
		mapBuilder := builder.WithMap()
		if iter != nil && iter.Name != "" && iter.Source != nil {
			if !iter.Source.Expr.Type.Repeated {
				ctx.addError(
					ErrWithLocation(
						`map iterator's src value type must be repeated type`,
						mapBuilder.WithIteratorSource().Location(),
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
						mapBuilder.WithIteratorSource().Location(),
					),
				)
			}
			mapEnv = newEnv
		}
		r.resolveMapIteratorExprCELValues(ctx, mapEnv, expr.Map.Expr, mapBuilder)
		varType := expr.Map.Expr.Type.Clone()
		varType.Repeated = true
		expr.Type = varType
	case expr.Call != nil:
		callBuilder := builder.WithCall()
		if expr.Call.Request != nil {
			for idx, arg := range expr.Call.Request.Args {
				if arg.Value == nil {
					continue
				}
				if err := r.resolveCELValue(ctx, env, arg.Value.CEL); err != nil {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							callBuilder.WithRequest(idx).WithBy().Location(),
						),
					)
				}
				field := expr.Call.Request.Type.Field(arg.Name)
				if field != nil && arg.Value.CEL != nil && arg.Value.CEL.Out != nil {
					r.validateRequestFieldType(ctx, arg.Value.CEL.Out, field, callBuilder.WithRequest(idx))
				}
				if arg.If != nil {
					if err := r.resolveCELValue(ctx, env, arg.If); err != nil {
						ctx.addError(
							ErrWithLocation(
								err.Error(),
								callBuilder.WithRequest(idx).WithIf().Location(),
							),
						)
					}
					if arg.If.Out != nil && arg.If.Out.Kind != types.Bool {
						ctx.addError(
							ErrWithLocation(
								"if must always return a boolean value",
								callBuilder.WithRequest(idx).WithIf().Location(),
							),
						)
					}
				}
			}
		}
		grpcErrEnv, _ := env.Extend(
			cel.Variable("error", cel.ObjectType("grpc.federation.private.Error")),
		)
		if expr.Call.Retry != nil {
			retryBuilder := callBuilder.WithRetry()
			retry := expr.Call.Retry
			if err := r.resolveCELValue(ctx, grpcErrEnv, retry.If); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						retryBuilder.WithIf().Location(),
					),
				)
			}
			if retry.If.Out != nil && retry.If.Out.Kind != types.Bool {
				ctx.addError(
					ErrWithLocation(
						"if must always return a boolean value",
						retryBuilder.WithIf().Location(),
					),
				)
			}
		}
		for idx, grpcErr := range expr.Call.Errors {
			r.resolveGRPCErrorCELValues(ctx, grpcErrEnv, grpcErr, expr.Call.Method.Response, callBuilder.WithError(idx))
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
							builder.WithMessage().WithArgs(argIdx).WithInline().Location(),
						),
					)
				} else {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							builder.WithMessage().WithArgs(argIdx).WithBy().Location(),
						),
					)
				}
			}
		}
		expr.Type = NewMessageType(expr.Message.Message, false)
	case expr.Validation != nil:
		validationBuilder := builder.WithValidation()
		r.resolveGRPCErrorCELValues(ctx, env, expr.Validation.Error, nil, validationBuilder.WithError())
		// This is a dummy type since the output from the validation is not supposed to be used (at least for now)
		expr.Type = BoolType
	}
}

func (r *Resolver) resolveMapIteratorExprCELValues(ctx *context, env *cel.Env, expr *MapIteratorExpr, mapBuilder *source.MapExprOptionBuilder) {
	switch {
	case expr.By != nil:
		if err := r.resolveCELValue(ctx, env, expr.By); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					mapBuilder.WithBy().Location(),
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
							mapBuilder.WithMessage().WithArgs(argIdx).WithInline().Location(),
						),
					)
				} else {
					ctx.addError(
						ErrWithLocation(
							err.Error(),
							mapBuilder.WithMessage().WithArgs(argIdx).WithBy().Location(),
						),
					)
				}
			}
		}
		expr.Type = NewMessageType(expr.Message.Message, false)
	}
}

func (r *Resolver) resolveGRPCErrorCELValues(ctx *context, env *cel.Env, grpcErr *GRPCError, response *Message, builder *source.GRPCErrorOptionBuilder) {
	if grpcErr == nil {
		return
	}
	for idx, def := range grpcErr.DefSet.Definitions() {
		env = r.resolveVariableDefinitionCELValues(ctx, env, def, builder.WithDef(idx))
	}
	if err := r.resolveCELValue(ctx, env, grpcErr.If); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithIf().Location(),
			),
		)
	}
	if grpcErr.If.Out != nil && grpcErr.If.Out.Kind != types.Bool {
		ctx.addError(
			ErrWithLocation(
				"if must always return a boolean value",
				builder.WithIf().Location(),
			),
		)
	}
	if grpcErr.IgnoreAndResponse != nil {
		if err := r.resolveCELValue(ctx, env, grpcErr.IgnoreAndResponse); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithIgnoreAndResponse().Location(),
				),
			)
		}
		if grpcErr.IgnoreAndResponse.Out != nil && response != nil {
			msg := grpcErr.IgnoreAndResponse.Out.Message
			if msg == nil || msg != response {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`value must be %q type`, response.FQDN()),
						builder.WithIgnoreAndResponse().Location(),
					),
				)
			}
		}
	}
	if grpcErr.Message != nil {
		if err := r.resolveCELValue(ctx, env, grpcErr.Message); err != nil {
			ctx.addError(
				ErrWithLocation(
					err.Error(),
					builder.WithMessage().Location(),
				),
			)
		}
		if grpcErr.Message.Out != nil && grpcErr.Message.Out.Kind != types.String {
			ctx.addError(
				ErrWithLocation(
					"message must always return a string value",
					builder.WithMessage().Location(),
				),
			)
		}
	}
	for idx, detail := range grpcErr.Details {
		r.resolveGRPCErrorDetailCELValues(ctx, env, detail, builder.WithDetail(idx))
	}
}

func (r *Resolver) resolveGRPCErrorDetailCELValues(ctx *context, env *cel.Env, detail *GRPCErrorDetail, builder *source.GRPCErrorDetailOptionBuilder) {
	for idx, def := range detail.DefSet.Definitions() {
		env = r.resolveVariableDefinitionCELValues(ctx, env, def, builder.WithDef(idx))
	}

	if err := r.resolveCELValue(ctx, env, detail.If); err != nil {
		ctx.addError(
			ErrWithLocation(
				err.Error(),
				builder.WithIf().Location(),
			),
		)
	}
	if detail.If.Out != nil && detail.If.Out.Kind != types.Bool {
		ctx.addError(
			ErrWithLocation(
				"if must always return a boolean value",
				builder.WithIf().Location(),
			),
		)
	}
	for idx, def := range detail.Messages.Definitions() {
		env = r.resolveVariableDefinitionCELValues(ctx, env, def, builder.WithDef(idx))
	}

	for fIdx, failure := range detail.PreconditionFailures {
		for fvIdx, violation := range failure.Violations {
			if err := r.resolveCELValue(ctx, env, violation.Type); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.WithPreconditionFailure(fIdx, fvIdx, "type").Location(),
					),
				)
			}
			if violation.Type.Out != nil && violation.Type.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"type must always return a string value",
						builder.WithPreconditionFailure(fIdx, fvIdx, "type").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Subject); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.WithPreconditionFailure(fIdx, fvIdx, "subject").Location(),
					),
				)
			}
			if violation.Subject.Out != nil && violation.Subject.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"subject must always return a string value",
						builder.WithPreconditionFailure(fIdx, fvIdx, "subject").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Description); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.WithPreconditionFailure(fIdx, fvIdx, "description").Location(),
					),
				)
			}
			if violation.Description.Out != nil && violation.Description.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"description must always return a string value",
						builder.WithPreconditionFailure(fIdx, fvIdx, "description").Location(),
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
						builder.WithBadRequest(bIdx, fvIdx, "field").Location(),
					),
				)
			}
			if violation.Field.Out != nil && violation.Field.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"field must always return a string value",
						builder.WithBadRequest(bIdx, fvIdx, "field").Location(),
					),
				)
			}
			if err := r.resolveCELValue(ctx, env, violation.Description); err != nil {
				ctx.addError(
					ErrWithLocation(
						err.Error(),
						builder.WithBadRequest(bIdx, fvIdx, "description").Location(),
					),
				)
			}
			if violation.Description.Out != nil && violation.Description.Out.Kind != types.String {
				ctx.addError(
					ErrWithLocation(
						"description must always return a string value",
						builder.WithBadRequest(bIdx, fvIdx, "description").Location(),
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
					builder.WithLocalizedMessage(lIdx, "message").Location(),
				),
			)
		}
		if message.Message.Out != nil && message.Message.Out.Kind != types.String {
			ctx.addError(
				ErrWithLocation(
					"message must always return a string value",
					builder.WithLocalizedMessage(lIdx, "message").Location(),
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
	msg.Rule.DefSet.MarkUsed(nameMap)
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
		oneof.DefSet.MarkUsed(oneofNameMap)
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
	r.celRegistry.clear()
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

	var useContextLib bool
	for _, ref := range checkedExpr.GetReferenceMap() {
		for _, overloadID := range ref.GetOverloadId() {
			for _, prefix := range r.ctxOverloadIDPrefixes {
				if strings.HasPrefix(overloadID, prefix) {
					useContextLib = true
				}
			}
			for _, plugin := range r.celPluginMap {
				for _, fn := range plugin.Functions {
					if strings.HasPrefix(overloadID, fn.ID) {
						useContextLib = true
					}
				}
			}
		}
	}
	value.UseContextLibrary = useContextLib
	value.Out = out
	value.CheckedExpr = checkedExpr
	return nil
}

func (r *Resolver) createCELEnv(msg *Message) (*cel.Env, error) {
	envOpts := []cel.EnvOption{
		cel.StdLib(),
		cel.Lib(grpcfedcel.NewLibrary(r.celRegistry)),
		cel.CrossTypeNumericComparisons(true),
		cel.CustomTypeAdapter(r.celRegistry),
		cel.CustomTypeProvider(r.celRegistry),
		cel.ASTValidators(grpcfedcel.NewASTValidators()...),
	}
	if envMsg := r.getEnvMessage(msg); envMsg != nil {
		fileDesc := envFileDescriptor(envMsg)
		if err := r.celRegistry.RegisterFiles(append(r.files, fileDesc)...); err != nil {
			return nil, err
		}
		envOpts = append(envOpts, cel.Variable("grpc.federation.env", cel.ObjectType(envMsg.FQDN())))
	}
	envOpts = append(envOpts, r.enumAccessors()...)
	envOpts = append(envOpts, r.enumOperators()...)
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

func (r *Resolver) getEnvMessage(msg *Message) *Message {
	if msg.File == nil {
		return nil
	}
	if msg.File.Package == nil {
		return nil
	}
	var envSvcs []*Service
	for _, file := range msg.File.Package.Files {
		for _, svc := range file.Services {
			if svc.Rule == nil {
				continue
			}
			if svc.Rule.Env == nil {
				continue
			}
			envSvcs = append(envSvcs, svc)
		}
	}

	// If there are multiple services with env, we do not have a method to select one from them.
	// Therefore, we proceed only when there is a single candidate service.
	// Since validation for cases with multiple services is done elsewhere, we don't need to create an error here.
	if len(envSvcs) != 1 {
		return nil
	}
	return envToMessage(envSvcs[0].File, envSvcs[0].Rule.Env)
}

func (r *Resolver) enumAccessors() []cel.EnvOption {
	var ret []cel.EnvOption
	for _, enum := range r.cachedEnumMap {
		ret = append(ret,
			cel.Function(
				fmt.Sprintf("%s.name", enum.FQDN()),
				cel.Overload(fmt.Sprintf("%s_name_int_string", enum.FQDN()), []*cel.Type{cel.IntType}, cel.StringType,
					cel.UnaryBinding(func(self ref.Val) ref.Val { return nil }),
				),
				cel.Overload(fmt.Sprintf("%s_name_enum_string", enum.FQDN()), []*cel.Type{celtypes.NewOpaqueType(enum.FQDN(), cel.IntType)}, cel.StringType,
					cel.UnaryBinding(func(self ref.Val) ref.Val { return nil }),
				),
			),
			cel.Function(
				fmt.Sprintf("%s.value", enum.FQDN()),
				cel.Overload(fmt.Sprintf("%s_value_string_enum", enum.FQDN()), []*cel.Type{cel.StringType}, celtypes.NewOpaqueType(enum.FQDN(), cel.IntType),
					cel.UnaryBinding(func(self ref.Val) ref.Val { return nil }),
				),
			),
			cel.Function(
				fmt.Sprintf("%s.from", enum.FQDN()),
				cel.Overload(fmt.Sprintf("%s_from_int_enum", enum.FQDN()), []*cel.Type{cel.IntType}, celtypes.NewOpaqueType(enum.FQDN(), cel.IntType),
					cel.UnaryBinding(func(self ref.Val) ref.Val { return nil }),
				),
			),
		)
		for _, value := range enum.Values {
			r.cachedEnumValueMap[value.FQDN()] = value
		}
	}
	return ret
}

// enumOperators an enum may be treated as an `opaque<int>` or as an `int`.
// In this case, the default `equal` and `not-equal` operators cannot be used, so operators are registered so that different types can be compared.
func (r *Resolver) enumOperators() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function(operators.Equals,
			cel.Overload(overloads.Equals, []*cel.Type{celtypes.NewTypeParamType("A"), celtypes.NewTypeParamType("B")}, cel.BoolType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val { return nil }),
			),
		),
		cel.Function(operators.NotEquals,
			cel.Overload(overloads.NotEquals, []*cel.Type{celtypes.NewTypeParamType("A"), celtypes.NewTypeParamType("B")}, cel.BoolType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val { return nil }),
			),
		),
	}
}

func (r *Resolver) fromCELType(ctx *context, typ *cel.Type) (*Type, error) {
	declTypeName := typ.DeclaredTypeName()
	switch typ.Kind() {
	case celtypes.BoolKind:
		if declTypeName == "wrapper(bool)" {
			return BoolValueType, nil
		}
		return BoolType, nil
	case celtypes.BytesKind:
		if declTypeName == "wrapper(bytes)" {
			return BytesValueType, nil
		}
		return BytesType, nil
	case celtypes.DoubleKind:
		if declTypeName == "wrapper(double)" {
			return DoubleValueType, nil
		}
		return DoubleType, nil
	case celtypes.IntKind:
		if declTypeName == "wrapper(int)" {
			return Int64ValueType, nil
		}
		if enum, found := r.celRegistry.LookupEnum(typ); found {
			return &Type{Kind: types.Enum, Enum: enum}, nil
		}
		return Int64Type, nil
	case celtypes.UintKind:
		if declTypeName == "wrapper(uint)" {
			return Uint64ValueType, nil
		}
		return Uint64Type, nil
	case celtypes.StringKind:
		if declTypeName == "wrapper(string)" {
			return StringValueType, nil
		}
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
		return NewMapType(mapKey, mapValue), nil
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
		param := typ.Parameters()[0]
		enum, ok := r.cachedEnumMap[typ.TypeName()]
		if ok && param.Kind() == celtypes.IntKind {
			return &Type{Kind: types.Enum, Enum: enum}, nil
		}
		if typ.TypeName() == grpcfedcel.EnumSelectorFQDN {
			trueType, err := r.fromCELType(ctx, typ.Parameters()[0])
			if err != nil {
				return nil, err
			}
			falseType, err := r.fromCELType(ctx, typ.Parameters()[1])
			if err != nil {
				return nil, err
			}
			if (trueType.Enum == nil && trueType.FQDN() != grpcfedcel.EnumSelectorFQDN) || (falseType.Enum == nil && falseType.FQDN() != grpcfedcel.EnumSelectorFQDN) {
				return nil, fmt.Errorf("cannot find enum type from enum selector")
			}
			return NewEnumSelectorType(trueType, falseType), nil
		}
		return r.fromCELType(ctx, param)
	case celtypes.NullTypeKind:
		return NullType, nil
	case celtypes.DynKind:
		return nil, fmt.Errorf("dyn type is unsupported")
	}

	return nil, fmt.Errorf("unknown type %s is required", typ.TypeName())
}

const (
	privateProtoFile  = "grpc/federation/private.proto"
	durationProtoFile = "google/protobuf/duration.proto"
)

func messageArgumentFileDescriptor(arg *Message) *descriptorpb.FileDescriptorProto {
	desc := arg.File.Desc
	msg := messageToDescriptor(arg)
	var importedPrivateFile bool
	for _, dep := range desc.GetDependency() {
		if dep == privateProtoFile {
			importedPrivateFile = true
			break
		}
	}
	deps := append(desc.GetDependency(), arg.File.Name)
	if !importedPrivateFile {
		deps = append(deps, privateProtoFile)
	}
	return &descriptorpb.FileDescriptorProto{
		Name:             proto.String(arg.Name),
		Package:          proto.String(federation.PrivatePackageName),
		Dependency:       deps,
		PublicDependency: desc.PublicDependency,
		WeakDependency:   desc.WeakDependency,
		MessageType:      []*descriptorpb.DescriptorProto{msg},
	}
}

func envToMessage(file *File, env *Env) *Message {
	copied := *file
	copied.Package = &Package{
		Name: federation.PrivatePackageName,
	}
	msg := &Message{
		File: &copied,
		Name: "Env",
	}
	for _, v := range env.Vars {
		msg.Fields = append(msg.Fields, &Field{
			Name: v.Name,
			Type: v.Type,
		})
	}
	return msg
}

func envFileDescriptor(envMsg *Message) *descriptorpb.FileDescriptorProto {
	msg := messageToDescriptor(envMsg)
	desc := envMsg.File.Desc
	var (
		importedPrivateFile  bool
		importedDurationFile bool
	)
	for _, dep := range desc.GetDependency() {
		if dep == privateProtoFile {
			importedPrivateFile = true
		}
		if dep == durationProtoFile {
			importedDurationFile = true
		}
	}
	deps := append(desc.GetDependency(), envMsg.File.Name)
	if !importedPrivateFile {
		deps = append(deps, privateProtoFile)
	}
	if !importedDurationFile {
		deps = append(deps, durationProtoFile)
	}
	return &descriptorpb.FileDescriptorProto{
		Name:             proto.String("env"),
		Package:          proto.String(federation.PrivatePackageName),
		Dependency:       deps,
		PublicDependency: desc.PublicDependency,
		WeakDependency:   desc.WeakDependency,
		MessageType:      []*descriptorpb.DescriptorProto{msg},
	}
}

func messageToDescriptor(m *Message) *descriptorpb.DescriptorProto {
	msg := &descriptorpb.DescriptorProto{
		Name: proto.String(m.Name),
		Options: &descriptorpb.MessageOptions{
			MapEntry: proto.Bool(m.IsMapEntry),
		},
	}
	for idx, field := range m.Fields {
		var (
			typeName string
			label    = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
		)
		if field.Type.Message != nil && field.Type.Message.IsMapEntry {
			mp := messageToDescriptor(field.Type.Message)
			if mp.GetName() == "" {
				mp.Name = proto.String(cases.Title(language.Und).String(field.Name) + "Entry")
			}
			msg.NestedType = append(msg.NestedType, mp)
			typeName = m.FQDN() + "." + mp.GetName()
			label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		} else if field.Type.Message != nil {
			typeName = field.Type.Message.FQDN()
		}
		if field.Type.Enum != nil {
			typeName = field.Type.Enum.FQDN()
		}
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
	return msg
}

func (r *Resolver) resolveAutoBind(ctx *context, files []*File) {
	msgs := r.allMessages(files)
	for _, msg := range msgs {
		ctx := ctx.withFile(msg.File).withMessage(msg)
		r.resolveAutoBindFields(ctx, msg, source.NewMessageBuilder(msg.File.Name, msg.Name))
	}
}

// resolveMessageRuleDependencies resolve dependencies for each message.
func (r *Resolver) resolveMessageDependencies(ctx *context, files []*File) {
	msgs := r.allMessages(files)
	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}
		setupVariableDefinitionSet(ctx, msg, msg.Rule.DefSet)
		for defIdx, def := range msg.Rule.DefSet.Definitions() {
			ctx := ctx.withDefIndex(defIdx)
			expr := def.Expr
			if expr == nil {
				continue
			}
			switch {
			case expr.Call != nil:
				for _, grpcErr := range expr.Call.Errors {
					r.resolveGRPCErrorMessageDependencies(ctx, msg, grpcErr)
				}
			case expr.Validation != nil:
				r.resolveGRPCErrorMessageDependencies(ctx, msg, expr.Validation.Error)
			}
		}
		for _, field := range msg.Fields {
			if field.Rule == nil {
				continue
			}
			if field.Rule.Oneof == nil {
				continue
			}
			setupVariableDefinitionSet(ctx, msg, field.Rule.Oneof.DefSet)
		}
	}
	r.validateMessages(ctx, msgs)
}

func (r *Resolver) resolveGRPCErrorMessageDependencies(ctx *context, msg *Message, grpcErr *GRPCError) {
	if grpcErr == nil {
		return
	}
	setupVariableDefinitionSet(ctx, msg, grpcErr.DefSet)
	for detIdx, detail := range grpcErr.Details {
		ctx := ctx.withErrDetailIndex(detIdx)
		setupVariableDefinitionSet(ctx, msg, detail.DefSet)
		setupVariableDefinitionSet(ctx, msg, detail.Messages)
	}
}

func (r *Resolver) resolveValue(def *commonValueDef) (*Value, error) {
	const (
		customResolverOpt = "custom_resolver"
		aliasOpt          = "alias"
		byOpt             = "by"
		inlineOpt         = "inline"
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

func isDifferentType(from, to *Type) bool {
	if from == nil || to == nil {
		return false
	}
	if from.IsNumber() && to.IsNumber() {
		return false
	}
	if from.IsNull && (to.Repeated || to.Kind == types.Message || to.Kind == types.Bytes) {
		return false
	}
	return from.Kind != to.Kind
}

func newMessageBuilderFromMessage(message *Message) *source.MessageBuilder {
	if message.ParentMessage == nil {
		return source.NewMessageBuilder(message.FileName(), message.Name)
	}
	builder := newMessageBuilderFromMessage(message.ParentMessage)
	return builder.WithMessage(message.Name)
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
