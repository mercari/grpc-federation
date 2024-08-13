package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"log"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
	"github.com/mercari/grpc-federation/util"
)

type CodeGenerator struct {
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (g *CodeGenerator) Generate(file *resolver.File, enums []*resolver.Enum) ([]byte, error) {
	tmpl, err := loadTemplate()
	if err != nil {
		return nil, err
	}
	return generateGoContent(tmpl, NewFile(file, enums))
}

type File struct {
	*resolver.File
	enums  []*resolver.Enum
	pkgMap map[*resolver.GoPackage]struct{}
}

func NewFile(file *resolver.File, enums []*resolver.Enum) *File {
	return &File{
		File:   file,
		pkgMap: make(map[*resolver.GoPackage]struct{}),
		enums:  enums,
	}
}

func (f *File) Version() string {
	return grpcfed.Version
}

func (f *File) Source() string {
	return f.Name
}

func (f *File) Services() []*Service {
	ret := make([]*Service, 0, len(f.File.Services))
	for _, svc := range f.File.Services {
		ret = append(ret, newService(svc, f))
	}
	return ret
}

type CELPlugin struct {
	file *File
	*resolver.CELPlugin
}

func (p *CELPlugin) FederationVersion() string {
	return grpcfed.Version
}

func (p *CELPlugin) FieldName() string {
	return util.ToPublicGoVariable(p.Name)
}

func (p *CELPlugin) PluginName() string {
	return fmt.Sprintf("%sPlugin", util.ToPublicGoVariable(p.Name))
}

func (p *CELPlugin) Functions() []*CELFunction {
	ret := make([]*CELFunction, 0, len(p.CELPlugin.Functions))
	for _, fn := range p.CELPlugin.Functions {
		ret = append(ret, &CELFunction{
			file:        p.file,
			CELFunction: fn,
		})
	}
	return ret
}

type CELFunction struct {
	file *File
	*resolver.CELFunction
	overloadIdx int
}

func (f *CELFunction) GoName() string {
	funcName := f.CELFunction.Name
	if f.overloadIdx != 0 {
		funcName += fmt.Sprint(f.overloadIdx + 1)
	}
	if f.Receiver != nil {
		return protoFQDNToPublicGoName(fmt.Sprintf("%s.%s", f.Receiver.FQDN(), funcName))
	}
	return protoFQDNToPublicGoName(funcName)
}

func (f *CELFunction) Name() string {
	funcName := f.CELFunction.Name
	if f.Receiver != nil {
		return protoFQDNToPublicGoName(fmt.Sprintf("%s.%s", f.Receiver.FQDN(), funcName))
	}
	return protoFQDNToPublicGoName(funcName)
}

func (f *CELFunction) IsMethod() bool {
	return f.CELFunction.Receiver != nil
}

func (f *CELFunction) ExportName() string {
	return f.ID
}

func (f *CELFunction) Args() []*CELFunctionArgument {
	ret := make([]*CELFunctionArgument, 0, len(f.CELFunction.Args))
	for _, arg := range f.CELFunction.Args {
		arg := arg
		ret = append(ret, &CELFunctionArgument{
			typeFunc: func() string {
				return f.file.toTypeText(arg)
			},
			file: f.file,
			arg:  arg,
		})
	}
	return ret
}

func (f *CELFunction) Return() *CELFunctionReturn {
	return &CELFunctionReturn{
		typeFunc: func() string {
			return f.file.toTypeText(f.CELFunction.Return)
		},
		ret:  f.CELFunction.Return,
		file: f.file,
	}
}

type CELFunctionArgument struct {
	typeFunc func() string
	file     *File
	arg      *resolver.Type
}

func (f *CELFunctionArgument) Type() string {
	return f.typeFunc()
}

func (f *CELFunctionArgument) CELType() string {
	return toCELTypeDeclare(f.arg)
}

func (f *CELFunctionArgument) Converter() string {
	switch f.arg.Kind {
	case types.Double:
		return "ToFloat64"
	case types.Float:
		return "ToFloat32"
	case types.Int32, types.Sint32, types.Sfixed32:
		return "ToInt32"
	case types.Int64, types.Sint64, types.Sfixed64:
		return "ToInt64"
	case types.Uint32, types.Fixed32:
		return "ToUint32"
	case types.Uint64, types.Fixed64:
		return "ToUint64"
	case types.Bool:
		return "ToBool"
	case types.String:
		return "ToString"
	case types.Bytes:
		return "ToBytes"
	case types.Enum:
		enum := f.file.toTypeText(f.arg)
		return fmt.Sprintf("ToEnum[%s]", enum)
	case types.Message:
		msg := f.file.toTypeText(f.arg)
		return fmt.Sprintf("ToMessage[%s]", msg)
	}
	return ""
}

type CELFunctionReturn struct {
	typeFunc func() string
	file     *File
	ret      *resolver.Type
}

func (r *CELFunctionReturn) Type() string {
	return r.typeFunc()
}

func (r *CELFunctionReturn) CELType() string {
	return toCELTypeDeclare(r.ret)
}

func (r *CELFunctionReturn) FuncName() string {
	return util.ToPublicGoVariable(r.ret.Kind.ToString())
}

func (r *CELFunctionReturn) Converter() string {
	switch r.ret.Kind {
	case types.Double:
		return "ToFloat64CELPluginResponse"
	case types.Float:
		return "ToFloat32CELPluginResponse"
	case types.Int32, types.Sint32, types.Sfixed32, types.Enum:
		return "ToInt32CELPluginResponse"
	case types.Int64, types.Sint64, types.Sfixed64:
		return "ToInt64CELPluginResponse"
	case types.Uint32, types.Fixed32:
		return "ToUint32CELPluginResponse"
	case types.Uint64, types.Fixed64:
		return "ToUint64CELPluginResponse"
	case types.Bool:
		return "ToBoolCELPluginResponse"
	case types.String:
		return "ToStringCELPluginResponse"
	case types.Bytes:
		return "ToBytesCELPluginResponse"
	case types.Message:
		msg := r.file.toTypeText(r.ret)
		return fmt.Sprintf("ToMessageCELPluginResponse[%s]", msg)
	}
	return ""
}

func (p *CELPlugin) PluginFunctions() []*CELFunction {
	ret := make([]*CELFunction, 0, len(p.CELPlugin.Functions))
	overloadCount := make(map[string]int)
	for _, fn := range p.CELPlugin.Functions {
		ret = append(ret, &CELFunction{
			file:        p.file,
			CELFunction: fn,
			overloadIdx: overloadCount[fn.Name],
		})
		overloadCount[fn.Name]++
	}
	return ret
}

func (f *File) CELPlugins() []*CELPlugin {
	ret := make([]*CELPlugin, 0, len(f.File.CELPlugins))
	for _, plug := range f.File.CELPlugins {
		ret = append(ret, &CELPlugin{
			file:      f,
			CELPlugin: plug,
		})
	}
	return ret
}

type Service struct {
	*resolver.Service
	file              *File
	nameToLogValueMap map[string]*LogValue
	celCacheIndex     int
}

func newService(svc *resolver.Service, file *File) *Service {
	return &Service{
		Service:           svc,
		file:              file,
		nameToLogValueMap: make(map[string]*LogValue),
	}
}

func (s *Service) CELCacheIndex() int {
	s.celCacheIndex++
	return s.celCacheIndex
}

func (s *Service) Env() *Env {
	if s.Rule == nil {
		return nil
	}
	if s.Rule.Env == nil {
		return nil
	}
	return &Env{
		Env:  s.Rule.Env,
		file: s.file,
	}
}

type Env struct {
	*resolver.Env
	file *File
}

func (e *Env) Vars() []*EnvVar {
	ret := make([]*EnvVar, 0, len(e.Env.Vars))
	for _, v := range e.Env.Vars {
		ret = append(ret, &EnvVar{
			EnvVar: v,
			file:   e.file,
		})
	}
	return ret
}

type EnvVar struct {
	*resolver.EnvVar
	file *File
}

func (v *EnvVar) Name() string {
	return util.ToPublicGoVariable(v.EnvVar.Name)
}

func (v *EnvVar) ProtoName() string {
	return v.EnvVar.Name
}

const durationPkg = "google.golang.org/protobuf/types/known/durationpb"

func (v *EnvVar) CELType() string {
	return toCELTypeDeclare(v.EnvVar.Type)
}

func (v *EnvVar) Type() string {
	var existsDurationPkg bool
	for pkg := range v.file.pkgMap {
		if pkg.ImportPath == durationPkg {
			existsDurationPkg = true
			break
		}
	}

	text := v.file.toTypeText(v.EnvVar.Type)

	// Since it is not possible to read env values directly into *durationpb.Duration,
	// replace it with the grpcfed.Duration type.
	defer func() {
		// If the duration type was not used previously, delete it to maintain consistency.
		if !existsDurationPkg {
			for pkg := range v.file.pkgMap {
				if pkg.ImportPath == durationPkg {
					delete(v.file.pkgMap, pkg)
					break
				}
			}
		}
	}()
	return strings.ReplaceAll(text, "*durationpb.", "grpcfed.")
}

func (v *EnvVar) Tag() string {
	if v.EnvVar.Option == nil {
		return ""
	}
	var opts []string
	if v.EnvVar.Option.Alternate != "" {
		opts = append(opts, fmt.Sprintf("envconfig:%q", v.EnvVar.Option.Alternate))
	}
	if v.EnvVar.Option.Default != "" {
		opts = append(opts, fmt.Sprintf("default:%q", v.EnvVar.Option.Default))
	}
	if v.EnvVar.Option.Required {
		opts = append(opts, `required:"true"`)
	}
	if v.EnvVar.Option.Ignored {
		opts = append(opts, `ignored:"true"`)
	}
	return strings.Join(opts, " ")
}

type Import struct {
	Path  string
	Alias string
	Used  bool
}

func (f *File) StandardImports() []*Import {
	var (
		existsServiceDef bool
		existsPluginDef  bool
	)
	if len(f.File.Services) != 0 {
		existsServiceDef = true
	}
	if len(f.File.CELPlugins) != 0 {
		existsPluginDef = true
	}

	pkgs := []*Import{
		{Path: "bufio", Used: existsPluginDef},
		{Path: "context", Used: existsServiceDef || existsPluginDef},
		{Path: "fmt", Used: existsPluginDef},
		{Path: "io", Used: existsServiceDef},
		{Path: "os", Used: existsPluginDef},
		{Path: "log/slog", Used: existsServiceDef},
		{Path: "reflect", Used: true},
	}
	usedPkgs := make([]*Import, 0, len(pkgs))
	for _, pkg := range pkgs {
		if !pkg.Used {
			continue
		}
		usedPkgs = append(usedPkgs, pkg)
	}
	return usedPkgs
}

func (f *File) DefaultImports() []*Import {
	var (
		existsServiceDef bool
		existsPluginDef  bool
	)
	if len(f.File.Services) != 0 {
		existsServiceDef = true
	}
	if len(f.File.CELPlugins) != 0 {
		existsPluginDef = true
	}
	pkgs := []*Import{
		{Alias: "grpcfed", Path: "github.com/mercari/grpc-federation/grpc/federation", Used: existsServiceDef || existsPluginDef},
		{Alias: "grpcfedcel", Path: "github.com/mercari/grpc-federation/grpc/federation/cel", Used: existsServiceDef},
		{Path: "go.opentelemetry.io/otel", Used: existsServiceDef},
		{Path: "go.opentelemetry.io/otel/trace", Used: existsServiceDef},
		{Path: "google.golang.org/grpc/metadata", Used: existsPluginDef},

		{Path: "google.golang.org/protobuf/types/descriptorpb"},
		{Path: "google.golang.org/protobuf/types/known/dynamicpb"},
		{Path: "google.golang.org/protobuf/types/known/anypb"},
		{Path: "google.golang.org/protobuf/types/known/durationpb"},
		{Path: "google.golang.org/protobuf/types/known/emptypb"},
		{Path: "google.golang.org/protobuf/types/known/timestamppb"},
	}
	importMap := make(map[string]*Import)
	for _, pkg := range pkgs {
		importMap[pkg.Path] = pkg
	}

	for pkg := range f.pkgMap {
		if imprt, exists := importMap[pkg.ImportPath]; exists {
			imprt.Used = true
		}
	}
	usedPkgs := make([]*Import, 0, len(pkgs))
	for _, pkg := range pkgs {
		if !pkg.Used {
			continue
		}
		usedPkgs = append(usedPkgs, pkg)
	}
	return usedPkgs
}

func (f *File) Imports() []*Import {
	defaultImportMap := make(map[string]struct{})
	for _, imprt := range f.DefaultImports() {
		defaultImportMap[imprt.Path] = struct{}{}
	}
	curImportPath := f.GoPackage.ImportPath

	imports := make([]*Import, 0, len(f.pkgMap))
	addImport := func(pkg *resolver.GoPackage) {
		// ignore standard library's enum.
		if strings.HasPrefix(pkg.Name, "google.") {
			return
		}
		if pkg.ImportPath == curImportPath {
			return
		}
		if _, exists := defaultImportMap[pkg.ImportPath]; exists {
			return
		}
		imports = append(imports, &Import{
			Path:  pkg.ImportPath,
			Alias: pkg.Name,
			Used:  true,
		})
	}
	for pkg := range f.pkgMap {
		addImport(pkg)
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path < imports[j].Path
	})
	return imports
}

type Enum struct {
	ProtoName string
	GoName    string
}

func (f *File) Enums() []*Enum {
	ret := make([]*Enum, 0, len(f.enums))
	for _, enum := range f.enums {
		protoName := enum.FQDN()
		// ignore standard library's enum.
		if strings.HasPrefix(protoName, "google.") {
			continue
		}
		if strings.HasPrefix(protoName, "grpc.federation.") {
			continue
		}
		// f.enums contain all the enums defined in the package
		// Currently Enums are used only from a File contains Services.
		if len(f.File.Services) != 0 {
			f.pkgMap[enum.GoPackage()] = struct{}{}
		}
		ret = append(ret, &Enum{
			ProtoName: protoName,
			GoName:    f.enumTypeToText(enum),
		})
	}
	return ret
}

func newServiceTypeDeclares(file *File, msgs []*resolver.Message) []*Type {
	var ret []*Type
	for _, msg := range msgs {
		ret = append(ret, newTypeDeclaresWithMessage(file, msg)...)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret
}

func newTypeDeclaresWithMessage(file *File, msg *resolver.Message) []*Type {
	if !msg.HasRule() {
		return nil
	}
	arg := msg.Rule.MessageArgument
	if arg == nil {
		return nil
	}
	msgName := fullMessageName(msg)
	argName := fmt.Sprintf("%sArgument", msgName)
	msgFQDN := fmt.Sprintf(`%s.%s`, msg.PackageName(), msg.Name)
	typ := &Type{
		Name:      argName,
		Desc:      fmt.Sprintf(`%s is argument for %q message`, argName, msgFQDN),
		ProtoFQDN: arg.FQDN(),
	}

	genMsg := &Message{Message: msg, file: file}
	fieldNameMap := make(map[string]struct{})
	for _, group := range genMsg.VariableDefinitionGroups() {
		for _, def := range group.VariableDefinitions() {
			if def == nil {
				continue
			}
			if !def.Used {
				continue
			}
			fieldName := util.ToPublicGoVariable(def.Name)
			if typ.HasField(fieldName) {
				continue
			}
			if _, exists := fieldNameMap[fieldName]; exists {
				continue
			}
			typ.Fields = append(typ.Fields, &Field{
				Name: fieldName,
				Type: file.toTypeText(def.Expr.Type),
			})
			fieldNameMap[fieldName] = struct{}{}
		}
	}
	for _, field := range arg.Fields {
		typ.ProtoFields = append(typ.ProtoFields, &ProtoField{Field: field})
		fieldName := util.ToPublicGoVariable(field.Name)
		if _, exists := fieldNameMap[fieldName]; exists {
			continue
		}
		var typText string
		if field.Type.OneofField != nil {
			t := field.Type.OneofField.Type.Clone()
			t.OneofField = nil
			typText = file.toTypeText(t)
		} else {
			typText = file.toTypeText(field.Type)
		}
		typ.Fields = append(typ.Fields, &Field{
			Name: fieldName,
			Type: typText,
		})
		fieldNameMap[fieldName] = struct{}{}
	}
	sort.Slice(typ.Fields, func(i, j int) bool {
		return typ.Fields[i].Name < typ.Fields[j].Name
	})

	declTypes := []*Type{typ}
	for _, field := range msg.CustomResolverFields() {
		typeName := fmt.Sprintf("%s_%sArgument", msgName, util.ToPublicGoVariable(field.Name))
		fields := []*Field{{Type: argName}}
		if msg.HasCustomResolver() {
			fields = append(fields, &Field{
				Name: msgName,
				Type: file.toTypeText(resolver.NewMessageType(msg, false)),
			})
		}
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Name < fields[j].Name
		})
		declTypes = append(declTypes, &Type{
			Name:   typeName,
			Fields: fields,
			Desc: fmt.Sprintf(
				`%s is custom resolver's argument for %q field of %q message`,
				typeName,
				field.Name,
				msgFQDN,
			),
		})
	}
	return declTypes
}

type OneofType struct {
	Name             string
	FieldName        string
	MessageProtoFQDN string
	TypeDeclare      string
	FieldZeroValues  []string
	FieldGetterNames []string
	ReturnZeroValue  string
}

func (s *Service) CELPlugins() []*CELPlugin {
	ret := make([]*CELPlugin, 0, len(s.Service.CELPlugins))
	for _, plugin := range s.Service.CELPlugins {
		ret = append(ret, &CELPlugin{
			CELPlugin: plugin,
			file:      s.file,
		})
	}
	return ret
}

func (s *Service) Types() Types {
	return newServiceTypeDeclares(s.file, s.Service.Messages)
}

func (s *Service) OneofTypes() []*OneofType {
	var ret []*OneofType
	svc := s.Service
	for _, msg := range svc.Messages {
		for _, oneof := range msg.Oneofs {
			if !oneof.IsSameType() {
				continue
			}
			fieldZeroValues := make([]string, 0, len(oneof.Fields))
			fieldGetterNames := make([]string, 0, len(oneof.Fields))
			for _, field := range oneof.Fields {
				fieldZeroValues = append(
					fieldZeroValues,
					toMakeZeroValue(s.file, field.Type),
				)
				fieldGetterNames = append(
					fieldGetterNames,
					fmt.Sprintf("Get%s", util.ToPublicGoVariable(field.Name)),
				)
			}
			cloned := oneof.Fields[0].Type.Clone()
			cloned.OneofField = nil
			returnZeroValue := toMakeZeroValue(s.file, cloned)
			ret = append(ret, &OneofType{
				Name:             oneof.Name,
				FieldName:        util.ToPublicGoVariable(oneof.Name),
				MessageProtoFQDN: oneof.Message.FQDN(),
				TypeDeclare:      toCELTypeDeclare(oneof.Fields[0].Type),
				FieldZeroValues:  fieldZeroValues,
				FieldGetterNames: fieldGetterNames,
				ReturnZeroValue:  returnZeroValue,
			})
		}
	}
	return ret
}

func (s *Service) ServiceName() string {
	return util.ToPublicGoVariable(s.Name)
}

func (s *Service) PackageName() string {
	return s.GoPackage().Name
}

type ServiceDependency struct {
	*resolver.ServiceDependency
}

func (dep *ServiceDependency) ServiceName() string {
	return fmt.Sprintf("%s.%s", dep.Service.PackageName(), dep.Service.Name)
}

func (dep *ServiceDependency) NameConfig() string {
	return dep.Name
}

func (dep *ServiceDependency) ClientName() string {
	svcName := fullServiceName(dep.Service)
	return fmt.Sprintf("%sClient", svcName)
}

func (dep *ServiceDependency) PrivateClientName() string {
	svcName := fullServiceName(dep.Service)
	return fmt.Sprintf("%sClient", util.ToPrivateGoVariable(svcName))
}

func (dep *ServiceDependency) ClientType() string {
	return fmt.Sprintf(
		"%s.%sClient",
		dep.Service.GoPackage().Name,
		dep.Service.Name,
	)
}

func (dep *ServiceDependency) ClientConstructor() string {
	return fmt.Sprintf(
		"%s.New%sClient",
		dep.Service.GoPackage().Name,
		dep.Service.Name,
	)
}

func (s *Service) ServiceDependencies() []*ServiceDependency {
	deps := s.Service.ServiceDependencies()
	ret := make([]*ServiceDependency, 0, len(deps))
	for _, dep := range deps {
		ret = append(ret, &ServiceDependency{dep})
	}
	return ret
}

type CustomResolver struct {
	*resolver.CustomResolver
	Service *resolver.Service
	file    *File
}

func (r *CustomResolver) Name() string {
	return fmt.Sprintf("Resolve_%s", r.fullName())
}

func (r *CustomResolver) RequestType() string {
	return fmt.Sprintf("%sArgument", r.fullName())
}

func (r *CustomResolver) fullName() string {
	ur := r.CustomResolver
	msg := fullMessageName(ur.Message)
	if ur.Field != nil {
		return fmt.Sprintf("%s_%s", msg, util.ToPublicGoVariable(ur.Field.Name))
	}
	return msg
}

func (r *CustomResolver) ProtoFQDN() string {
	fqdn := r.Message.FQDN()
	if r.Field != nil {
		return fmt.Sprintf("%s.%s", fqdn, r.Field.Name)
	}
	return fqdn
}

func (r *CustomResolver) ReturnType() string {
	ur := r.CustomResolver
	if ur.Field != nil {
		// field resolver
		return r.file.toTypeText(ur.Field.Type)
	}
	// message resolver
	return r.file.toTypeText(resolver.NewMessageType(ur.Message, false))
}

func (s *Service) CustomResolvers() []*CustomResolver {
	resolvers := s.Service.CustomResolvers()
	ret := make([]*CustomResolver, 0, len(resolvers))
	for _, resolver := range resolvers {
		ret = append(ret, &CustomResolver{
			CustomResolver: resolver,
			Service:        s.Service,
			file:           s.file,
		})
	}
	return ret
}

type Types []*Type

func (t Types) HasProtoFields() bool {
	for _, tt := range t {
		if len(tt.ProtoFields) != 0 {
			return true
		}
	}
	return false
}

type Type struct {
	Name        string
	Fields      []*Field
	ProtoFields []*ProtoField
	Desc        string
	ProtoFQDN   string
}

func (t *Type) HasField(fieldName string) bool {
	for _, field := range t.Fields {
		if field.Name == fieldName {
			return true
		}
	}
	return false
}

type ProtoField struct {
	*resolver.Field
}

func (f *ProtoField) FieldName() string {
	return util.ToPublicGoVariable(f.Field.Name)
}

func (f *ProtoField) Name() string {
	return f.Field.Name
}

func (f *ProtoField) TypeDeclare() string {
	return toCELTypeDeclare(f.Field.Type)
}

type Field struct {
	Name string
	Type string
}

func toMakeZeroValue(file *File, t *resolver.Type) string {
	text := file.toTypeText(t)
	if t.Repeated || t.Kind == types.Bytes {
		return fmt.Sprintf("%s(nil)", text)
	}
	if t.OneofField != nil {
		return fmt.Sprintf("(%s)(nil)", text)
	}
	if t.IsNumber() {
		return fmt.Sprintf("%s(0)", text)
	}
	switch t.Kind {
	case types.Bool:
		return `false`
	case types.String:
		return `""`
	}
	return fmt.Sprintf("(%s)(nil)", text)
}

func toCELTypeDeclare(t *resolver.Type) string {
	if t.Repeated {
		cloned := t.Clone()
		cloned.Repeated = false
		return fmt.Sprintf("grpcfed.NewCELListType(%s)", toCELTypeDeclare(cloned))
	}
	switch t.Kind {
	case types.Double, types.Float:
		return "grpcfed.CELDoubleType"
	case types.Int32, types.Int64, types.Sint32, types.Sint64, types.Sfixed32, types.Sfixed64, types.Enum:
		return "grpcfed.CELIntType"
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return "grpcfed.CELUintType"
	case types.Bool:
		return "grpcfed.CELBoolType"
	case types.String:
		return "grpcfed.CELStringType"
	case types.Bytes:
		return "grpcfed.CELBytesType"
	case types.Message:
		if t.Message != nil && t.Message.IsMapEntry {
			key := toCELTypeDeclare(t.Message.Fields[0].Type)
			value := toCELTypeDeclare(t.Message.Fields[1].Type)
			return fmt.Sprintf(`grpcfed.NewCELMapType(%s, %s)`, key, value)
		}
		if t == resolver.DurationType {
			return "grpcfed.CELDurationType"
		}
		return fmt.Sprintf(`grpcfed.NewCELObjectType(%q)`, t.Message.FQDN())
	}
	return ""
}

func (f *File) toTypeText(t *resolver.Type) string {
	if t == nil || t.IsNull {
		return "any"
	}
	if t.OneofField != nil {
		typ := f.oneofTypeToText(t.OneofField)
		if t.Repeated {
			return "[]" + typ
		}
		return typ
	}

	var typ string
	switch t.Kind {
	case types.Double:
		typ = "float64"
	case types.Float:
		typ = "float32"
	case types.Int32:
		typ = "int32"
	case types.Int64:
		typ = "int64"
	case types.Uint32:
		typ = "uint32"
	case types.Uint64:
		typ = "uint64"
	case types.Sint32:
		typ = "int32"
	case types.Sint64:
		typ = "int64"
	case types.Fixed32:
		typ = "uint32"
	case types.Fixed64:
		typ = "uint64"
	case types.Sfixed32:
		typ = "int32"
	case types.Sfixed64:
		typ = "int64"
	case types.Bool:
		typ = "bool"
	case types.String:
		typ = "string"
	case types.Bytes:
		typ = "[]byte"
	case types.Enum:
		f.pkgMap[t.Enum.GoPackage()] = struct{}{}
		typ = f.enumTypeToText(t.Enum)
	case types.Message:
		if t.Message.IsMapEntry {
			return fmt.Sprintf("map[%s]%s", f.toTypeText(t.Message.Fields[0].Type), f.toTypeText(t.Message.Fields[1].Type))
		}
		typ = f.messageTypeToText(t.Message)
	default:
		log.Fatalf("grpc-federation: specified unsupported type value %s", t.Kind.ToString())
	}
	if t.Repeated {
		return "[]" + typ
	}
	return typ
}

func (f *File) oneofTypeToText(oneofField *resolver.OneofField) string {
	msg := f.messageTypeToText(oneofField.Oneof.Message)
	oneof := fmt.Sprintf("%s_%s", msg, util.ToPublicGoVariable(oneofField.Field.Name))
	if oneofField.IsConflict() {
		oneof += "_"
	}
	return oneof
}

func (f *File) messageTypeToText(msg *resolver.Message) string {
	f.pkgMap[msg.GoPackage()] = struct{}{}
	if msg.PackageName() == "google.protobuf" {
		switch msg.Name {
		case "Any":
			return "*anypb.Any"
		case "Timestamp":
			return "*timestamppb.Timestamp"
		case "Duration":
			return "*durationpb.Duration"
		case "Empty":
			return "*emptypb.Empty"
		}
	}
	if msg.IsMapEntry {
		key := f.toTypeText(msg.Field("key").Type)
		value := f.toTypeText(msg.Field("value").Type)
		return fmt.Sprintf("map[%s]%s", key, value)
	}
	var names []string
	for _, name := range append(msg.ParentMessageNames(), msg.Name) {
		names = append(names, util.ToPublicGoVariable(name))
	}
	name := strings.Join(names, "_")
	if f.GoPackage.ImportPath == msg.GoPackage().ImportPath {
		return fmt.Sprintf("*%s", name)
	}
	return fmt.Sprintf("*%s.%s", msg.GoPackage().Name, name)
}

func (f *File) enumTypeToText(enum *resolver.Enum) string {
	var name string
	if enum.Message != nil {
		var names []string
		for _, n := range append(enum.Message.ParentMessageNames(), enum.Message.Name, enum.Name) {
			names = append(names, util.ToPublicGoVariable(n))
		}
		name = strings.Join(names, "_")
	} else {
		name = util.ToPublicGoVariable(enum.Name)
	}
	if f.GoPackage.ImportPath == enum.GoPackage().ImportPath {
		return name
	}
	return fmt.Sprintf("%s.%s", enum.GoPackage().Name, name)
}

type LogValue struct {
	Name      string
	ValueType string
	Attrs     []*LogValueAttr
	Type      *resolver.Type
	Value     string // for repeated type
}

func (v *LogValue) IsRepeated() bool {
	if v.IsMap() {
		return false
	}
	return v.Type.Repeated
}

func (v *LogValue) IsMessage() bool {
	return v.Type.Kind == types.Message
}

func (v *LogValue) IsMap() bool {
	if !v.IsMessage() {
		return false
	}
	return v.Type.Message.IsMapEntry
}

type LogValueAttr struct {
	Type  string
	Key   string
	Value string
}

func (s *Service) LogValues() []*LogValue {
	msgs := s.Service.Messages
	for _, dep := range s.Service.ServiceDependencies() {
		for _, method := range dep.Service.Methods {
			msgs = append(msgs, method.Request)
		}
	}

	for _, msg := range msgs {
		if msg.Rule != nil {
			s.setLogValueByMessageArgument(msg)
		}
		s.setLogValueByMessage(msg)
	}
	logValues := make([]*LogValue, 0, len(s.nameToLogValueMap))
	for _, logValue := range s.nameToLogValueMap {
		logValues = append(logValues, logValue)
	}
	sort.Slice(logValues, func(i, j int) bool {
		return logValues[i].Name < logValues[j].Name
	})
	return logValues
}

func (s *Service) setLogValueByMessage(msg *resolver.Message) {
	if msg.IsMapEntry {
		s.setLogValueByMapMessage(msg)
		return
	}
	name := s.messageToLogValueName(msg)
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	msgType := resolver.NewMessageType(msg, false)
	logValue := &LogValue{
		Name:      name,
		ValueType: s.file.toTypeText(msgType),
		Attrs:     make([]*LogValueAttr, 0, len(msg.Fields)),
		Type:      msgType,
	}
	s.nameToLogValueMap[name] = logValue
	for _, field := range msg.Fields {
		logValue.Attrs = append(logValue.Attrs, &LogValueAttr{
			Type:  s.logType(field.Type),
			Key:   field.Name,
			Value: s.logValue(field),
		})
		s.setLogValueByType(field.Type)
	}
	for _, msg := range msg.NestedMessages {
		s.setLogValueByMessage(msg)
	}
	for _, enum := range msg.Enums {
		s.setLogValueByEnum(enum)
	}
}

func (s *Service) setLogValueByMapMessage(msg *resolver.Message) {
	name := s.messageToLogValueName(msg)
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	value := msg.Field("value")
	msgType := resolver.NewMessageType(msg, false)
	logValue := &LogValue{
		Name:      name,
		ValueType: s.file.toTypeText(msgType),
		Type:      msgType,
		Value:     s.logMapValue(value),
	}
	s.nameToLogValueMap[name] = logValue
	s.setLogValueByType(value.Type)
}

func (s *Service) setLogValueByType(typ *resolver.Type) {
	isMap := typ.Message != nil && typ.Message.IsMapEntry
	if typ.Message != nil {
		s.setLogValueByMessage(typ.Message)
	}
	if typ.Enum != nil {
		s.setLogValueByEnum(typ.Enum)
	}
	if typ.Repeated && !isMap {
		s.setLogValueByRepeatedType(typ)
	}
}

func (s *Service) setLogValueByMessageArgument(msg *resolver.Message) {
	arg := msg.Rule.MessageArgument

	// The message argument belongs to "grpc.federation.private" package,
	// but we want to change the package name temporarily
	// because we want to use the name of the package to which the message belongs as log value's function name.
	pkg := arg.File.Package
	defer func() {
		// restore the original package name.
		arg.File.Package = pkg
	}()
	// replace package name to message's package name.
	arg.File.Package = msg.File.Package

	name := s.messageToLogValueName(arg)
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	logValue := &LogValue{
		Name:      name,
		ValueType: "*" + s.ServiceName() + "_" + protoFQDNToPublicGoName(arg.FQDN()),
		Attrs:     make([]*LogValueAttr, 0, len(arg.Fields)),
		Type:      resolver.NewMessageType(arg, false),
	}
	s.nameToLogValueMap[name] = logValue
	for _, field := range arg.Fields {
		logValue.Attrs = append(logValue.Attrs, &LogValueAttr{
			Type:  s.logType(field.Type),
			Key:   field.Name,
			Value: s.msgArgumentLogValue(field),
		})
		s.setLogValueByType(field.Type)
	}
}

func (s *Service) setLogValueByEnum(enum *resolver.Enum) {
	name := s.enumToLogValueName(enum)
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	enumType := &resolver.Type{Kind: types.Enum, Enum: enum}
	logValue := &LogValue{
		Name:      name,
		ValueType: s.file.toTypeText(enumType),
		Attrs:     make([]*LogValueAttr, 0, len(enum.Values)),
		Type:      enumType,
	}
	for _, value := range enum.Values {
		logValue.Attrs = append(logValue.Attrs, &LogValueAttr{
			Key:   value.Value,
			Value: toEnumValueText(toEnumValuePrefix(s.file, enumType), value.Value),
		})
	}
	s.nameToLogValueMap[name] = logValue
}

func (s *Service) setLogValueByRepeatedType(typ *resolver.Type) {
	if typ.Kind != types.Message && typ.Kind != types.Enum {
		return
	}
	var (
		name  string
		value string
	)
	if typ.Kind == types.Message {
		name = s.repeatedMessageToLogValueName(typ.Message)
		value = s.messageToLogValueName(typ.Message)
	} else {
		name = s.repeatedEnumToLogValueName(typ.Enum)
		value = s.enumToLogValueName(typ.Enum)
	}
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	logValue := &LogValue{
		Name:      name,
		ValueType: s.file.toTypeText(typ),
		Type:      typ,
		Value:     value,
	}
	s.nameToLogValueMap[name] = logValue
}

func (s *Service) messageToLogValueName(msg *resolver.Message) string {
	return fmt.Sprintf("logvalue_%s", protoFQDNToPublicGoName(msg.FQDN()))
}

func (s *Service) enumToLogValueName(enum *resolver.Enum) string {
	return fmt.Sprintf("logvalue_%s", protoFQDNToPublicGoName(enum.FQDN()))
}

func (s *Service) repeatedMessageToLogValueName(msg *resolver.Message) string {
	return fmt.Sprintf("logvalue_repeated_%s", protoFQDNToPublicGoName(msg.FQDN()))
}

func (s *Service) repeatedEnumToLogValueName(enum *resolver.Enum) string {
	return fmt.Sprintf("logvalue_repeated_%s", protoFQDNToPublicGoName(enum.FQDN()))
}

func (s *Service) logType(typ *resolver.Type) string {
	if typ.Repeated {
		return "Any"
	}
	switch typ.Kind {
	case types.Double, types.Float:
		return "Float64"
	case types.Int32, types.Int64, types.Sint32, types.Sint64, types.Sfixed32, types.Sfixed64:
		return "Int64"
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return "Uint64"
	case types.Bool:
		return "Bool"
	case types.String, types.Bytes:
		return "String"
	case types.Enum:
		return "String"
	case types.Group, types.Message:
		return "Any"
	}
	log.Fatalf("grpc-federation: specified unknown type value %s", typ.Kind.ToString())
	return ""
}

func (s *Service) logValue(field *resolver.Field) string {
	typ := field.Type
	base := fmt.Sprintf("v.Get%s()", util.ToPublicGoVariable(field.Name))
	if typ.Kind == types.Message {
		return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
	} else if typ.Kind == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
		}
		return fmt.Sprintf("%s(%s).String()", s.logValueFuncName(typ), base)
	}
	if field.Type.Repeated {
		return base
	}
	switch field.Type.Kind {
	case types.Int32, types.Sint32, types.Sfixed32:
		base = fmt.Sprintf("int64(%s)", base)
	case types.Uint32, types.Fixed32:
		base = fmt.Sprintf("uint64(%s)", base)
	case types.Float:
		base = fmt.Sprintf("float64(%s)", base)
	case types.Bytes:
		base = fmt.Sprintf("string(%s)", base)
	}
	return base
}

func (s *Service) logMapValue(value *resolver.Field) string {
	typ := value.Type
	if typ.Kind == types.Message {
		return fmt.Sprintf("%s(value)", s.logValueFuncName(typ))
	} else if typ.Kind == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(value)", s.logValueFuncName(typ))
		}
		return fmt.Sprintf("slog.StringValue(%s(value).String())", s.logValueFuncName(typ))
	} else if typ.Kind == types.Bytes {
		return "slog.StringValue(string(value))"
	}
	return "slog.AnyValue(value)"
}

func (s *Service) msgArgumentLogValue(field *resolver.Field) string {
	typ := field.Type
	base := fmt.Sprintf("v.%s", util.ToPublicGoVariable(field.Name))
	if typ.Kind == types.Message {
		return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
	} else if typ.Kind == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
		}
		return fmt.Sprintf("%s(%s).String()", s.logValueFuncName(typ), base)
	}
	if field.Type.Repeated {
		return base
	}
	switch field.Type.Kind {
	case types.Int32, types.Sint32, types.Sfixed32:
		base = fmt.Sprintf("int64(%s)", base)
	case types.Uint32, types.Fixed32:
		base = fmt.Sprintf("uint64(%s)", base)
	case types.Float:
		base = fmt.Sprintf("float64(%s)", base)
	case types.Bytes:
		base = fmt.Sprintf("string(%s)", base)
	}
	return base
}

func (s *Service) logValueFuncName(typ *resolver.Type) string {
	if typ.Kind == types.Message {
		isMap := typ.Message.IsMapEntry
		if typ.Repeated && !isMap {
			return fmt.Sprintf("s.%s", s.repeatedMessageToLogValueName(typ.Message))
		}
		return fmt.Sprintf("s.%s", s.messageToLogValueName(typ.Message))
	}
	if typ.Kind == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("s.%s", s.repeatedEnumToLogValueName(typ.Enum))
		}
		return fmt.Sprintf("s.%s", s.enumToLogValueName(typ.Enum))
	}
	return ""
}

type DependentMethod struct {
	Name string
	FQDN string
}

func (s *Service) DependentMethods() []*DependentMethod {
	methodMap := make(map[string]*DependentMethod)
	for _, method := range s.Service.Methods {
		for _, depMethod := range s.dependentMethods(method.Response) {
			methodMap[depMethod.Name] = depMethod
		}
	}
	depMethods := make([]*DependentMethod, 0, len(methodMap))
	for _, depMethod := range methodMap {
		depMethods = append(depMethods, depMethod)
	}
	sort.Slice(depMethods, func(i, j int) bool {
		return depMethods[i].Name < depMethods[j].Name
	})
	return depMethods
}

func (s *Service) dependentMethods(msg *resolver.Message) []*DependentMethod {
	if msg.Rule == nil {
		return nil
	}
	var ret []*DependentMethod
	for _, varDef := range msg.Rule.DefSet.Definitions() {
		if varDef.Expr == nil {
			continue
		}

		expr := varDef.Expr
		if expr.Call != nil {
			if expr.Call.Method != nil {
				method := expr.Call.Method
				ret = append(ret, &DependentMethod{
					Name: fmt.Sprintf("%s_%s", fullServiceName(method.Service), method.Name),
					FQDN: "/" + method.FQDN(),
				})
			}
			continue
		}

		for _, msgExpr := range varDef.MessageExprs() {
			ret = append(ret, s.dependentMethods(msgExpr.Message)...)
		}
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for _, varDef := range field.Rule.Oneof.DefSet.Definitions() {
			if varDef.Expr == nil {
				continue
			}
			expr := varDef.Expr
			if expr.Call != nil {
				if expr.Call.Method != nil {
					method := expr.Call.Method
					ret = append(ret, &DependentMethod{
						Name: fmt.Sprintf("%s_%s", fullServiceName(method.Service), method.Name),
						FQDN: "/" + method.FQDN(),
					})
				}
				continue
			}
			for _, msgExpr := range varDef.MessageExprs() {
				ret = append(ret, s.dependentMethods(msgExpr.Message)...)
			}
		}
	}
	return ret
}

type Method struct {
	*resolver.Method
	Service *Service
	file    *File
}

func (m *Method) Name() string {
	return util.ToPublicGoVariable(m.Method.Name)
}

func (m *Method) ProtoFQDN() string {
	return m.FQDN()
}

func (m *Method) UseTimeout() bool {
	if m.Rule == nil {
		return false
	}
	return m.Rule.Timeout != nil
}

func (m *Method) Timeout() string {
	return fmt.Sprintf("%[1]d/* %[1]s */", *m.Rule.Timeout)
}

func (m *Method) ResolverName() string {
	msg := fullMessageName(m.Response)
	if m.Response.Rule == nil {
		return fmt.Sprintf("resolver.Resolve_%s", msg)
	}
	return fmt.Sprintf("resolve_%s", msg)
}

func (m *Method) ArgumentName() string {
	msg := fullMessageName(m.Response)
	return fmt.Sprintf("%s_%sArgument", m.Service.ServiceName(), msg)
}

func (m *Method) RequestType() string {
	return m.file.toTypeText(resolver.NewMessageType(m.Request, false))
}

func (m *Method) ReturnType() string {
	return m.file.toTypeText(resolver.NewMessageType(m.Response, false))
}

func (m *Method) ReturnTypeWithoutPtr() string {
	return strings.TrimPrefix(m.ReturnType(), "*")
}

func (m *Method) ReturnTypeArguments() []string {
	var args []string
	for _, field := range m.Request.Fields {
		args = append(args, util.ToPublicGoVariable(field.Name))
	}
	return args
}

func (s *Service) Methods() []*Method {
	methods := make([]*Method, 0, len(s.Service.Methods))
	for _, method := range s.Service.Methods {
		methods = append(methods, &Method{
			Service: s,
			Method:  method,
			file:    s.file,
		})
	}
	return methods
}

type Message struct {
	*resolver.Message
	Service *Service
	file    *File
}

func (m *Message) CELCacheIndex() int {
	return m.Service.CELCacheIndex()
}

func (m *Message) ProtoFQDN() string {
	return m.FQDN()
}

func (m *Message) ResolverName() string {
	msg := fullMessageName(m.Message)
	return fmt.Sprintf("resolve_%s", msg)
}

func (m *Message) RequestType() string {
	msg := fullMessageName(m.Message)
	return fmt.Sprintf("%sArgument", msg)
}

func (m *Message) RequestProtoType() string {
	return m.Message.Rule.MessageArgument.FQDN()
}

func (m *Message) CustomResolverName() string {
	msg := fullMessageName(m.Message)
	return fmt.Sprintf("Resolve_%s", msg)
}

func (m *Message) ReturnType() string {
	return strings.TrimPrefix(m.file.toTypeText(resolver.NewMessageType(m.Message, false)), "*")
}

func (m *Message) LogValueReturnType() string {
	return fullMessageName(m.Message)
}

func (m *Message) VariableDefinitionSet() *VariableDefinitionSet {
	if m.Rule == nil {
		return nil
	}
	if m.Rule.DefSet == nil {
		return nil
	}
	if len(m.Rule.DefSet.Defs) == 0 {
		return nil
	}
	return &VariableDefinitionSet{
		VariableDefinitionSet: m.Rule.DefSet,
		msg:                   m,
	}
}

type VariableDefinitionSet struct {
	msg *Message
	*resolver.VariableDefinitionSet
}

func (set *VariableDefinitionSet) Definitions() []*VariableDefinition {
	ret := make([]*VariableDefinition, 0, len(set.VariableDefinitionSet.Definitions()))
	for _, def := range set.VariableDefinitionSet.Definitions() {
		ret = append(ret, &VariableDefinition{
			Service:            set.msg.Service,
			Message:            set.msg,
			VariableDefinition: def,
		})
	}
	return ret
}

func (set *VariableDefinitionSet) DependencyGraph() string {
	if set == nil {
		return ""
	}
	tree := resolver.DependencyGraphTreeFormat(set.DefinitionGroups())
	if !strings.HasSuffix(tree, "\n") {
		// If there is only one node, no newline code is added at the end.
		// In this case, do not display graphs.
		return ""
	}
	return tree
}

func (set *VariableDefinitionSet) VariableDefinitionGroups() []*VariableDefinitionGroup {
	if set == nil {
		return nil
	}
	var groups []*VariableDefinitionGroup
	for _, group := range set.DefinitionGroups() {
		groups = append(groups, &VariableDefinitionGroup{
			Service:                 set.msg.Service,
			Message:                 set.msg,
			VariableDefinitionGroup: group,
		})
	}
	return groups
}

func (m *Message) IsDeclVariables() bool {
	return m.HasCELValue() || m.Rule != nil && len(m.Rule.DefSet.DefinitionGroups()) != 0
}

type DeclVariable struct {
	Name string
	Type string
}

func (m *Message) DeclVariables() []*DeclVariable {
	if m.Rule == nil {
		return nil
	}
	if !m.HasResolvers() {
		return nil
	}
	valueMap := map[string]*DeclVariable{}
	for _, group := range m.Message.VariableDefinitionGroups() {
		for _, def := range group.VariableDefinitions() {
			if def == nil {
				continue
			}
			valueMap[def.Name] = &DeclVariable{
				Name: def.Name,
				Type: m.file.toTypeText(def.Expr.Type),
			}
		}
	}
	values := make([]*DeclVariable, 0, len(valueMap))
	for _, value := range valueMap {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Name < values[j].Name
	})
	return values
}

type ReturnField struct {
	CustomResolver *CustomResolverReturnField
	Oneof          *OneofReturnField
	CEL            *CELReturnField
	AutoBind       *AutoBindReturnField
}

type OneofReturnField struct {
	Name              string
	OneofCaseFields   []*OneofField
	OneofDefaultField *OneofField
}

type CELReturnField struct {
	CEL          *resolver.CELValue
	SetterParam  *SetterParam
	ProtoComment string
	Type         string
}

type SetterParam struct {
	Name         string
	Value        string
	RequiredCast bool
	CastFunc     string
	EnumSelector *EnumSelectorSetterParam
}

type EnumSelectorSetterParam struct {
	Type                  string
	RequiredCastTrueType  bool
	RequiredCastFalseType bool
	TrueType              string
	FalseType             string
	CastTrueTypeFunc      string
	CastFalseTypeFunc     string
	TrueEnumSelector      *EnumSelectorSetterParam
	FalseEnumSelector     *EnumSelectorSetterParam
}

type CustomResolverReturnField struct {
	Name                string
	ResolverName        string
	RequestType         string
	MessageArgumentName string
	MessageName         string
}

type AutoBindReturnField struct {
	Name         string
	Value        string
	RequiredCast bool
	CastFunc     string
	ProtoComment string
}

func (f *OneofReturnField) HasFieldOneofRule() bool {
	for _, field := range f.OneofCaseFields {
		if field.FieldOneofRule != nil {
			return true
		}
	}
	if f.OneofDefaultField != nil && f.OneofDefaultField.FieldOneofRule != nil {
		return true
	}
	return false
}

type OneofField struct {
	Expr           string
	By             string
	Type           string
	Condition      string
	Name           string
	Value          string
	CastValue      string
	Message        *Message
	FieldOneofRule *resolver.FieldOneofRule
	SetterParam    *SetterParam
}

func (oneof *OneofField) VariableDefinitionSet() *VariableDefinitionSet {
	if oneof.FieldOneofRule == nil {
		return nil
	}
	if oneof.FieldOneofRule.DefSet == nil {
		return nil
	}
	if len(oneof.FieldOneofRule.DefSet.Defs) == 0 {
		return nil
	}
	return &VariableDefinitionSet{
		VariableDefinitionSet: oneof.FieldOneofRule.DefSet,
		msg:                   oneof.Message,
	}
}

type CastField struct {
	Name     string
	service  *resolver.Service
	fromType *resolver.Type
	toType   *resolver.Type
	file     *File
}

func (f *CastField) RequestType() string {
	if f.fromType.OneofField != nil {
		typ := f.fromType.OneofField.Type.Clone()
		typ.OneofField = nil
		return f.file.toTypeText(typ)
	}
	return f.file.toTypeText(f.fromType)
}

func (f *CastField) ResponseType() string {
	return f.file.toTypeText(f.toType)
}

func (f *CastField) RequestProtoFQDN() string {
	return f.fromType.FQDN()
}

func (f *CastField) ResponseProtoFQDN() string {
	return f.toType.FQDN()
}

func (f *CastField) IsSlice() bool {
	return f.fromType.Repeated
}

func (f *CastField) IsOneof() bool {
	return f.fromType.OneofField != nil
}

func (f *CastField) IsStruct() bool {
	return f.fromType.Kind == types.Message
}

func (f *CastField) IsEnumToNumber() bool {
	return f.fromType.Kind == types.Enum && f.toType.Kind != types.Enum && f.toType.IsNumber()
}

func (f *CastField) IsNumberToEnum() bool {
	return f.fromType.Kind != types.Enum && f.fromType.IsNumber() && f.toType.Kind == types.Enum
}

func (f *CastField) IsEnum() bool {
	return f.fromType.Kind == types.Enum && f.toType.Kind == types.Enum
}

func (f *CastField) IsMap() bool {
	return f.fromType.Kind == types.Message && f.fromType.Message.IsMapEntry &&
		f.toType.Kind == types.Message && f.toType.Message.IsMapEntry
}

var (
	castValidationNumberMap = map[string]bool{
		types.Int64.ToString() + types.Int32.ToString():   true,
		types.Int64.ToString() + types.Uint32.ToString():  true,
		types.Int64.ToString() + types.Uint64.ToString():  true,
		types.Int32.ToString() + types.Uint32.ToString():  true,
		types.Int32.ToString() + types.Uint64.ToString():  true,
		types.Uint64.ToString() + types.Int32.ToString():  true,
		types.Uint64.ToString() + types.Int64.ToString():  true,
		types.Uint64.ToString() + types.Uint32.ToString(): true,
		types.Uint32.ToString() + types.Int32.ToString():  true,
	}
)

func (f *CastField) IsRequiredValidationNumber() bool {
	return castValidationNumberMap[f.fromType.Kind.ToString()+f.toType.Kind.ToString()]
}

func (f *CastField) CastWithValidationName() string {
	return fmt.Sprintf("%sTo%s", util.ToPublicVariable(f.RequestType()), util.ToPublicVariable(f.ResponseType()))
}

type CastEnum struct {
	FromValues   []*CastEnumValue
	DefaultValue string
}

type CastEnumValue struct {
	FromValue string
	ToValue   string
}

func (f *CastField) ToEnum() (*CastEnum, error) {
	toEnum := f.toType.Enum
	if toEnum.Rule != nil && len(toEnum.Rule.Aliases) != 0 {
		return f.toEnumFromDepServiceToService(toEnum, f.fromType.Enum), nil
	}
	fromEnum := f.fromType.Enum
	if fromEnum.Rule != nil && len(fromEnum.Rule.Aliases) != 0 {
		// the type conversion is performed at the time of gRPC method call.
		return f.toEnumFromServiceToDepService(fromEnum, f.toType.Enum)
	}
	toEnumName := toEnumValuePrefix(f.file, f.toType)
	fromEnumName := toEnumValuePrefix(f.file, f.fromType)
	var enumValues []*CastEnumValue
	for _, toValue := range toEnum.Values {
		if toValue.Rule == nil {
			continue
		}
		fromValue := fromEnum.Value(toValue.Value)
		if fromValue == nil {
			continue
		}
		enumValues = append(enumValues, &CastEnumValue{
			FromValue: toEnumValueText(fromEnumName, fromValue.Value),
			ToValue:   toEnumValueText(toEnumName, toValue.Value),
		})
	}
	return &CastEnum{
		FromValues:   enumValues,
		DefaultValue: "0",
	}, nil
}

// toEnumFromDepServiceToService converts dependent service's enum type to our service's enum type.
// alias is defined for dependent service.
func (f *CastField) toEnumFromDepServiceToService(enum, depSvcEnum *resolver.Enum) *CastEnum {
	var (
		enumValues   []*CastEnumValue
		defaultValue = "0"
	)
	svcEnumName := toEnumValuePrefix(f.file, f.toType)
	depSvcEnumName := toEnumValuePrefix(f.file, f.fromType)
	for _, value := range enum.Values {
		if value.Rule == nil {
			continue
		}
		for _, enumValueAlias := range value.Rule.Aliases {
			if enumValueAlias.EnumAlias != depSvcEnum {
				continue
			}
			for _, alias := range enumValueAlias.Aliases {
				enumValues = append(enumValues, &CastEnumValue{
					FromValue: toEnumValueText(depSvcEnumName, alias.Value),
					ToValue:   toEnumValueText(svcEnumName, value.Value),
				})
			}
		}
		if value.Rule.Default {
			defaultValue = toEnumValueText(svcEnumName, value.Value)
		}
	}
	return &CastEnum{
		FromValues:   enumValues,
		DefaultValue: defaultValue,
	}
}

// toEnumFromServiceToDepService converts our service's enum type to dependent service's enum type.
// alias is defined for dependent service.
func (f *CastField) toEnumFromServiceToDepService(enum, depSvcEnum *resolver.Enum) (*CastEnum, error) {
	var (
		enumValues   []*CastEnumValue
		defaultValue = "0"
	)
	depSvcEnumName := toEnumValuePrefix(f.file, f.toType)
	svcEnumName := toEnumValuePrefix(f.file, f.fromType)
	for _, value := range enum.Values {
		if value.Rule == nil {
			continue
		}
		for _, enumValueAlias := range value.Rule.Aliases {
			if enumValueAlias.EnumAlias != depSvcEnum {
				continue
			}
			if len(enumValueAlias.Aliases) != 1 {
				return nil, fmt.Errorf("found multiple aliases are set. the conversion destination cannot be uniquely determined")
			}
			depEnumValue := enumValueAlias.Aliases[0]
			enumValues = append(enumValues, &CastEnumValue{
				FromValue: toEnumValueText(svcEnumName, value.Value),
				ToValue:   toEnumValueText(depSvcEnumName, depEnumValue.Value),
			})
		}
		if value.Rule.Default {
			return nil, fmt.Errorf("found default alias is set. the conversion destination cannot be uniquely determined")
		}
	}
	return &CastEnum{
		FromValues:   enumValues,
		DefaultValue: defaultValue,
	}, nil
}

type CastSlice struct {
	ResponseType     string
	ElemRequiredCast bool
	ElemCastName     string
}

func (f *CastField) ToSlice() *CastSlice {
	fromElemType := f.fromType.Clone()
	toElemType := f.toType.Clone()
	fromElemType.Repeated = false
	toElemType.Repeated = false
	return &CastSlice{
		ResponseType:     f.ResponseType(),
		ElemRequiredCast: requiredCast(fromElemType, toElemType),
		ElemCastName:     castFuncName(fromElemType, toElemType),
	}
}

type CastStruct struct {
	Name   string
	Fields []*CastStructField
	Oneofs []*CastOneofStruct
}

type CastOneofStruct struct {
	Name   string
	Fields []*CastStructField
}

type CastStructField struct {
	ToFieldName   string
	FromFieldName string
	CastName      string
	RequiredCast  bool
}

func (f *CastField) ToStruct() *CastStruct {
	toMsg := f.toType.Message
	fromMsg := f.fromType.Message

	var castFields []*CastStructField
	for _, toField := range toMsg.Fields {
		if toField.Type.OneofField != nil {
			continue
		}
		field := f.toStructField(toField, fromMsg)
		if field == nil {
			continue
		}
		castFields = append(castFields, field)
	}

	var castOneofStructs []*CastOneofStruct
	for _, oneof := range toMsg.Oneofs {
		var fields []*CastStructField
		for _, toField := range oneof.Fields {
			if toField.Type.OneofField == nil {
				continue
			}
			field := f.toStructField(toField, fromMsg)
			if field == nil {
				continue
			}
			fields = append(fields, field)
		}
		castOneofStructs = append(castOneofStructs, &CastOneofStruct{
			Name:   util.ToPublicGoVariable(oneof.Name),
			Fields: fields,
		})
	}

	var names []string
	for _, n := range append(toMsg.ParentMessageNames(), toMsg.Name) {
		names = append(names, util.ToPublicGoVariable(n))
	}
	name := strings.Join(names, "_")
	if f.service.GoPackage().ImportPath != toMsg.GoPackage().ImportPath {
		name = fmt.Sprintf("%s.%s", toMsg.GoPackage().Name, name)
	}
	return &CastStruct{
		Name:   name,
		Fields: castFields,
		Oneofs: castOneofStructs,
	}
}

func (f *CastField) toStructField(toField *resolver.Field, fromMsg *resolver.Message) *CastStructField {
	if toField.HasRule() && len(toField.Rule.Aliases) != 0 {
		for _, alias := range toField.Rule.Aliases {
			if fromMsg != alias.Message {
				continue
			}
			fromField := alias
			fromType := fromField.Type
			toType := toField.Type
			requiredCast := requiredCast(fromType, toType)
			return &CastStructField{
				ToFieldName:   util.ToPublicGoVariable(toField.Name),
				FromFieldName: util.ToPublicGoVariable(fromField.Name),
				RequiredCast:  requiredCast,
				CastName:      castFuncName(fromType, toType),
			}
		}
	}
	fromField := fromMsg.Field(toField.Name)
	if fromField == nil {
		return nil
	}
	fromType := fromField.Type
	toType := toField.Type
	requiredCast := requiredCast(fromType, toType)
	return &CastStructField{
		ToFieldName:   util.ToPublicGoVariable(toField.Name),
		FromFieldName: util.ToPublicGoVariable(fromField.Name),
		RequiredCast:  requiredCast,
		CastName:      castFuncName(fromType, toType),
	}
}

type CastOneof struct {
	Name         string
	FieldName    string
	CastName     string
	RequiredCast bool
}

func (f *CastField) ToOneof() *CastOneof {
	toField := f.toType.OneofField
	msg := f.toType.OneofField.Oneof.Message

	fromField := f.fromType.OneofField
	fromType := fromField.Type.Clone()
	toType := toField.Type.Clone()
	fromType.OneofField = nil
	toType.OneofField = nil
	requiredCast := requiredCast(fromType, toType)
	var names []string
	for _, n := range append(msg.ParentMessageNames(), msg.Name, toField.Name) {
		names = append(names, util.ToPublicGoVariable(n))
	}
	name := strings.Join(names, "_")
	if toField.IsConflict() {
		name += "_"
	}
	if f.service.GoPackage().ImportPath != msg.GoPackage().ImportPath {
		name = fmt.Sprintf("%s.%s", msg.GoPackage().Name, name)
	}
	return &CastOneof{
		Name:         name,
		FieldName:    util.ToPublicGoVariable(toField.Name),
		RequiredCast: requiredCast,
		CastName:     castFuncName(fromType, toType),
	}
}

type CastMap struct {
	ResponseType      string
	KeyRequiredCast   bool
	KeyCastName       string
	ValueRequiredCast bool
	ValueCastName     string
}

func (f *CastField) ToMap() *CastMap {
	fromKeyType := f.fromType.Message.Fields[0].Type
	toKeyType := f.toType.Message.Fields[0].Type
	fromValueType := f.fromType.Message.Fields[1].Type
	toValueType := f.toType.Message.Fields[1].Type
	return &CastMap{
		ResponseType:      f.ResponseType(),
		KeyRequiredCast:   requiredCast(fromKeyType, toKeyType),
		KeyCastName:       castFuncName(fromKeyType, toKeyType),
		ValueRequiredCast: requiredCast(fromValueType, toValueType),
		ValueCastName:     castFuncName(fromValueType, toValueType),
	}
}

func (m *Message) CustomResolverArguments() []*Argument {
	args := []*Argument{}
	argNameMap := make(map[string]struct{})
	for _, group := range m.VariableDefinitionGroups() {
		for _, def := range group.VariableDefinitions() {
			if def == nil || !def.Used {
				continue
			}
			name := def.Name
			if _, exists := argNameMap[name]; exists {
				continue
			}
			args = append(args, &Argument{
				Name:  util.ToPublicGoVariable(name),
				Value: toUserDefinedVariable(name),
			})
			argNameMap[name] = struct{}{}
		}
	}
	sort.Slice(args, func(i, j int) bool {
		return args[i].Name < args[j].Name
	})
	return args
}

func (m *Message) ReturnFields() ([]*ReturnField, error) {
	var returnFields []*ReturnField
	for _, field := range m.Fields {
		if field.Oneof != nil {
			continue
		}
		if !field.HasRule() {
			continue
		}
		rule := field.Rule
		switch {
		case rule.CustomResolver:
			returnFields = append(returnFields, &ReturnField{
				CustomResolver: m.customResolverToReturnField(field),
			})
		case rule.Value != nil:
			value := rule.Value
			returnFields = append(returnFields, &ReturnField{
				CEL: m.celValueToReturnField(field, value.CEL),
			})
		case rule.AutoBindField != nil:
			autoBind, err := m.autoBindFieldToReturnField(field, rule.AutoBindField)
			if err != nil {
				return nil, err
			}
			returnFields = append(returnFields, &ReturnField{
				AutoBind: autoBind,
			})
		}
	}
	for _, oneof := range m.Oneofs {
		returnField, err := m.oneofValueToReturnField(oneof)
		if err != nil {
			return nil, err
		}
		returnFields = append(returnFields, &ReturnField{
			Oneof: returnField,
		})
	}
	return returnFields, nil
}

func (m *Message) autoBindFieldToReturnField(field *resolver.Field, autoBindField *resolver.AutoBindField) (*AutoBindReturnField, error) {
	var name string
	if autoBindField.VariableDefinition != nil {
		name = autoBindField.VariableDefinition.Name
	}

	fieldName := util.ToPublicGoVariable(field.Name)

	var (
		value    = fmt.Sprintf("value.vars.%s.Get%s()", name, fieldName)
		castFunc string
	)
	if field.RequiredTypeConversion() {
		var fromType *resolver.Type
		for _, sourceType := range field.SourceTypes() {
			if typ := m.autoBindSourceType(autoBindField, sourceType); typ != nil {
				fromType = typ
				break
			}
		}
		if fromType == nil {
			return nil, fmt.Errorf(
				"failed to autobind: expected message type is %q but it is not found from %q field",
				autoBindField.Field.Message.FQDN(),
				field.FQDN(),
			)
		}
		castFunc = castFuncName(fromType, field.Type)
	}
	return &AutoBindReturnField{
		Name:         fieldName,
		Value:        value,
		RequiredCast: field.RequiredTypeConversion(),
		CastFunc:     castFunc,
		ProtoComment: fmt.Sprintf(`// { name: %q, autobind: true }`, name),
	}, nil
}

func (m *Message) autoBindSourceType(autoBindField *resolver.AutoBindField, candidateType *resolver.Type) *resolver.Type {
	switch autoBindField.Field.Type.Kind {
	case types.Message:
		if candidateType.Message == nil {
			return nil
		}
		if autoBindField.Field.Type.Message != candidateType.Message {
			return nil
		}
		return candidateType
	case types.Enum:
		if candidateType.Enum == nil {
			return nil
		}
		if autoBindField.Field.Type.Enum != candidateType.Enum {
			return nil
		}
		return candidateType
	default:
		return candidateType
	}
}

func (m *Message) celValueToReturnField(field *resolver.Field, value *resolver.CELValue) *CELReturnField {
	toType := field.Type
	fromType := value.Out

	enumSelectorSetterParam := m.createEnumSelectorParam(fromType.Message, toType)

	toText := m.file.toTypeText(toType)
	fromText := m.file.toTypeText(fromType)

	var (
		typ          string
		requiredCast bool
	)
	switch fromType.Kind {
	case types.Message:
		typ = fromText
		requiredCast = field.RequiredTypeConversion()
	case types.Enum:
		typ = fromText
		requiredCast = field.RequiredTypeConversion()
	default:
		// Since fromType is a primitive type, type conversion is possible on the CEL side.
		typ = toText
	}
	return &CELReturnField{
		CEL: value,
		ProtoComment: field.Rule.ProtoFormat(&resolver.ProtoFormatOption{
			Prefix:         "// ",
			IndentSpaceNum: 2,
		}),
		Type: typ,
		SetterParam: &SetterParam{
			Name:         util.ToPublicGoVariable(field.Name),
			Value:        "v",
			RequiredCast: requiredCast,
			EnumSelector: enumSelectorSetterParam,
			CastFunc:     castFuncName(fromType, toType),
		},
	}
}

func (m *Message) createEnumSelectorParam(msg *resolver.Message, toType *resolver.Type) *EnumSelectorSetterParam {
	if msg == nil {
		return nil
	}
	if !msg.IsEnumSelector() {
		return nil
	}
	ret := &EnumSelectorSetterParam{
		Type: m.file.toTypeText(toType),
	}
	trueType := msg.Fields[0].Type
	if trueType.Message != nil && trueType.Message.IsEnumSelector() {
		ret.TrueEnumSelector = m.createEnumSelectorParam(trueType.Message, toType)
	} else {
		ret.RequiredCastTrueType = requiredCast(trueType, toType)
		ret.TrueType = m.file.toTypeText(trueType)
		ret.CastTrueTypeFunc = castFuncName(trueType, toType)
	}
	falseType := msg.Fields[1].Type
	if falseType.Message != nil && falseType.Message.IsEnumSelector() {
		ret.FalseEnumSelector = m.createEnumSelectorParam(falseType.Message, toType)
	} else {
		ret.RequiredCastFalseType = requiredCast(falseType, toType)
		ret.FalseType = m.file.toTypeText(falseType)
		ret.CastFalseTypeFunc = castFuncName(falseType, toType)
	}
	return ret
}

func (m *Message) customResolverToReturnField(field *resolver.Field) *CustomResolverReturnField {
	msgName := fullMessageName(m.Message)
	resolverName := fmt.Sprintf("Resolve_%s_%s", msgName, util.ToPublicGoVariable(field.Name))
	requestType := fmt.Sprintf("%s_%sArgument", msgName, util.ToPublicGoVariable(field.Name))
	return &CustomResolverReturnField{
		Name:                util.ToPublicGoVariable(field.Name),
		ResolverName:        resolverName,
		RequestType:         requestType,
		MessageArgumentName: fmt.Sprintf("%sArgument", msgName),
		MessageName:         msgName,
	}
}

func (m *Message) oneofValueToReturnField(oneof *resolver.Oneof) (*OneofReturnField, error) {
	var (
		caseFields   []*OneofField
		defaultField *OneofField
	)
	for _, field := range oneof.Fields {
		if !field.HasRule() {
			continue
		}
		rule := field.Rule
		switch {
		case rule.AutoBindField != nil:
			autoBind, err := m.autoBindFieldToReturnField(field, rule.AutoBindField)
			if err != nil {
				return nil, err
			}
			var castValue string
			if autoBind.CastFunc != "" {
				castValue = fmt.Sprintf("s.%s(%s)", autoBind.CastFunc, autoBind.Value)
			}
			caseFields = append(caseFields, &OneofField{
				Condition: fmt.Sprintf("%s != nil", autoBind.Value),
				Name:      autoBind.Name,
				Value:     autoBind.Value,
				CastValue: castValue,
			})
		case rule.Oneof != nil:
			// explicit binding with FieldOneofRule.
			fromType := rule.Oneof.By.Out
			toType := field.Type

			fieldName := util.ToPublicGoVariable(field.Name)
			oneofTypeName := m.Message.Name + "_" + fieldName
			if toType.OneofField.IsConflict() {
				oneofTypeName += "_"
			}
			var (
				typ       string
				value     string
				castValue string
				castFunc  string
			)
			switch fromType.Kind {
			case types.Message:
				typ = m.file.toTypeText(fromType)
				value = fmt.Sprintf("&%s{%s: v}", oneofTypeName, fieldName)
				if requiredCast(fromType, toType) {
					castFunc = castFuncName(fromType, toType)
					castValue = fmt.Sprintf("&%s{%s: s.%s(v)}", oneofTypeName, fieldName, castFunc)
				}
			case types.Enum:
				typ = "int64"
				value = fmt.Sprintf("&%s{%s: %s(v)}", oneofTypeName, fieldName, m.file.toTypeText(fromType))
				if requiredCast(fromType, toType) {
					castFunc = castFuncName(fromType, toType)
					castValue = fmt.Sprintf("&%s{%s: s.%s(v)}", oneofTypeName, fieldName, castFunc)
				}
			default:
				// Since fromType is a primitive type, type conversion is possible on the CEL side.
				deletedOneofFieldType := toType.Clone()
				deletedOneofFieldType.OneofField = nil
				typ = m.file.toTypeText(deletedOneofFieldType)
				value = fmt.Sprintf("&%s{%s: v}", oneofTypeName, fieldName)
			}
			if rule.Oneof.Default {
				defaultField = &OneofField{
					By:             rule.Oneof.By.Expr,
					Type:           typ,
					Condition:      fmt.Sprintf(`oneof_%s.(bool)`, fieldName),
					Name:           fieldName,
					Value:          value,
					CastValue:      castValue,
					FieldOneofRule: rule.Oneof,
					Message:        m,
					SetterParam: &SetterParam{
						Name:         util.ToPublicGoVariable(oneof.Name),
						Value:        value,
						EnumSelector: nil,
						RequiredCast: castFunc != "",
						CastFunc:     castFunc,
					},
				}
			} else {
				caseFields = append(caseFields, &OneofField{
					Expr:           rule.Oneof.If.Expr,
					By:             rule.Oneof.By.Expr,
					Type:           typ,
					Condition:      fmt.Sprintf(`oneof_%s.(bool)`, fieldName),
					Name:           fieldName,
					Value:          value,
					CastValue:      castValue,
					FieldOneofRule: rule.Oneof,
					Message:        m,
					SetterParam: &SetterParam{
						Name:         util.ToPublicGoVariable(oneof.Name),
						Value:        value,
						EnumSelector: nil,
						RequiredCast: castFunc != "",
						CastFunc:     castFunc,
					},
				})
			}
		}
	}
	return &OneofReturnField{
		Name:              util.ToPublicGoVariable(oneof.Name),
		OneofCaseFields:   caseFields,
		OneofDefaultField: defaultField,
	}, nil
}

type VariableDefinitionGroup struct {
	Service *Service
	Message *Message
	resolver.VariableDefinitionGroup
}

func (g *VariableDefinitionGroup) IsConcurrent() bool {
	return g.Type() == resolver.ConcurrentVariableDefinitionGroupType
}

func (g *VariableDefinitionGroup) ExistsStart() bool {
	rg, ok := g.VariableDefinitionGroup.(*resolver.SequentialVariableDefinitionGroup)
	if !ok {
		return false
	}
	return rg.Start != nil
}

func (g *VariableDefinitionGroup) ExistsEnd() bool {
	switch rg := g.VariableDefinitionGroup.(type) {
	case *resolver.SequentialVariableDefinitionGroup:
		return rg.End != nil
	case *resolver.ConcurrentVariableDefinitionGroup:
		return rg.End != nil
	}
	return false
}

func (g *VariableDefinitionGroup) Starts() []*VariableDefinitionGroup {
	rg, ok := g.VariableDefinitionGroup.(*resolver.ConcurrentVariableDefinitionGroup)
	if !ok {
		return nil
	}
	var starts []*VariableDefinitionGroup
	for _, start := range rg.Starts {
		starts = append(starts, &VariableDefinitionGroup{
			Service:                 g.Service,
			Message:                 g.Message,
			VariableDefinitionGroup: start,
		})
	}
	return starts
}

func (g *VariableDefinitionGroup) Start() *VariableDefinitionGroup {
	rg, ok := g.VariableDefinitionGroup.(*resolver.SequentialVariableDefinitionGroup)
	if !ok {
		return nil
	}
	return &VariableDefinitionGroup{
		Service:                 g.Service,
		Message:                 g.Message,
		VariableDefinitionGroup: rg.Start,
	}
}

func (g *VariableDefinitionGroup) End() *VariableDefinition {
	switch rg := g.VariableDefinitionGroup.(type) {
	case *resolver.SequentialVariableDefinitionGroup:
		return &VariableDefinition{
			Service:            g.Service,
			Message:            g.Message,
			VariableDefinition: rg.End,
		}
	case *resolver.ConcurrentVariableDefinitionGroup:
		return &VariableDefinition{
			Service:            g.Service,
			Message:            g.Message,
			VariableDefinition: rg.End,
		}
	}
	return nil
}

type VariableDefinition struct {
	Service *Service
	Message *Message
	*resolver.VariableDefinition
}

func (d *VariableDefinition) CELCacheIndex() int {
	return d.Service.CELCacheIndex()
}

func (d *VariableDefinition) Key() string {
	return d.VariableDefinition.Name
}

func (d *VariableDefinition) UseIf() bool {
	return d.VariableDefinition.If != nil
}

func (d *VariableDefinition) If() string {
	return d.VariableDefinition.If.Expr
}

func (d *VariableDefinition) UseTimeout() bool {
	expr := d.VariableDefinition.Expr
	return expr.Call != nil && expr.Call.Timeout != nil
}

func (d *VariableDefinition) UseRetry() bool {
	expr := d.VariableDefinition.Expr
	return expr.Call != nil && expr.Call.Retry != nil
}

func (d *VariableDefinition) Retry() *resolver.RetryPolicy {
	return d.VariableDefinition.Expr.Call.Retry
}

func (d *VariableDefinition) MethodFQDN() string {
	expr := d.VariableDefinition.Expr
	if expr.Call != nil {
		return expr.Call.Method.FQDN()
	}
	return ""
}

func (d *VariableDefinition) RequestTypeFQDN() string {
	expr := d.VariableDefinition.Expr
	if expr.Call != nil {
		return expr.Call.Request.Type.FQDN()
	}
	return ""
}

func (d *VariableDefinition) Timeout() string {
	expr := d.VariableDefinition.Expr
	if expr.Call != nil {
		return fmt.Sprintf("%[1]d/* %[1]s */", *expr.Call.Timeout)
	}
	return ""
}

func (d *VariableDefinition) UseArgs() bool {
	expr := d.VariableDefinition.Expr
	switch {
	case expr.By != nil:
		return false
	case expr.Map != nil:
		return false
	}
	return true
}

func (d *VariableDefinition) Caller() string {
	expr := d.VariableDefinition.Expr
	switch {
	case expr.Call != nil:
		method := expr.Call.Method
		methodName := method.Name
		svcName := fullServiceName(method.Service)
		return fmt.Sprintf("client.%sClient.%s", svcName, methodName)
	case expr.Message != nil:
		msgName := fullMessageName(expr.Message.Message)
		if expr.Message.Message.HasRule() {
			return fmt.Sprintf("resolve_%s", msgName)
		}
		return fmt.Sprintf("resolver.Resolve_%s", msgName)
	}
	return ""
}

func (d *VariableDefinition) HasErrorHandler() bool {
	return d.VariableDefinition.Expr.Call != nil
}

func (d *VariableDefinition) GRPCErrors() []*GRPCError {
	callExpr := d.VariableDefinition.Expr.Call
	if callExpr == nil {
		return nil
	}
	ret := make([]*GRPCError, 0, len(callExpr.Errors))
	for _, grpcErr := range callExpr.Errors {
		ret = append(ret, &GRPCError{
			GRPCError: grpcErr,
			msg:       d.Message,
		})
	}
	return ret
}

type GRPCError struct {
	*resolver.GRPCError
	msg *Message
}

func (e *GRPCError) CELCacheIndex() int {
	return e.msg.Service.CELCacheIndex()
}

// GoGRPCStatusCode converts a gRPC status code to a corresponding Go const name
// e.g. FAILED_PRECONDITION -> FailedPrecondition.
func (e *GRPCError) GoGRPCStatusCode() string {
	strCode := e.Code.String()
	if strCode == "OK" {
		// The only exception that the second character is in capital case as well
		return "OKCode"
	}

	parts := strings.Split(strCode, "_")
	titles := make([]string, 0, len(parts))
	for _, part := range parts {
		titles = append(titles, cases.Title(language.Und).String(part))
	}
	return strings.Join(titles, "") + "Code"
}

func (e *GRPCError) VariableDefinitionSet() *VariableDefinitionSet {
	if e.DefSet == nil {
		return nil
	}
	if len(e.DefSet.Defs) == 0 {
		return nil
	}
	return &VariableDefinitionSet{
		VariableDefinitionSet: e.DefSet,
		msg:                   e.msg,
	}
}

func (e *GRPCError) Details() []*GRPCErrorDetail {
	ret := make([]*GRPCErrorDetail, 0, len(e.GRPCError.Details))
	for _, detail := range e.GRPCError.Details {
		ret = append(ret, &GRPCErrorDetail{
			GRPCErrorDetail: detail,
			msg:             e.msg,
		})
	}
	return ret
}

type GRPCErrorDetail struct {
	*resolver.GRPCErrorDetail
	msg *Message
}

func (detail *GRPCErrorDetail) VariableDefinitionSet() *VariableDefinitionSet {
	if detail == nil {
		return nil
	}
	if detail.DefSet == nil {
		return nil
	}
	if len(detail.DefSet.Defs) == 0 {
		return nil
	}
	return &VariableDefinitionSet{
		VariableDefinitionSet: detail.DefSet,
		msg:                   detail.msg,
	}
}

func (detail *GRPCErrorDetail) MessageSet() *VariableDefinitionSet {
	if detail == nil {
		return nil
	}
	if detail.Messages == nil {
		return nil
	}
	if len(detail.Messages.Defs) == 0 {
		return nil
	}
	return &VariableDefinitionSet{
		VariableDefinitionSet: detail.Messages,
		msg:                   detail.msg,
	}
}

type GRPCErrorDetailBy struct {
	Expr string
	Type string
}

func (detail *GRPCErrorDetail) By() []*GRPCErrorDetailBy {
	ret := make([]*GRPCErrorDetailBy, 0, len(detail.GRPCErrorDetail.By))
	for _, by := range detail.GRPCErrorDetail.By {
		ret = append(ret, &GRPCErrorDetailBy{
			Expr: by.Expr,
			Type: toMakeZeroValue(detail.msg.file, by.Out),
		})
	}
	return ret
}

func (detail *GRPCErrorDetail) PreconditionFailures() []*PreconditionFailure {
	ret := make([]*PreconditionFailure, 0, len(detail.GRPCErrorDetail.PreconditionFailures))
	for _, pf := range detail.GRPCErrorDetail.PreconditionFailures {
		ret = append(ret, &PreconditionFailure{
			PreconditionFailure: pf,
			msg:                 detail.msg,
		})
	}
	return ret
}

func (detail *GRPCErrorDetail) BadRequests() []*BadRequest {
	ret := make([]*BadRequest, 0, len(detail.GRPCErrorDetail.BadRequests))
	for _, b := range detail.GRPCErrorDetail.BadRequests {
		ret = append(ret, &BadRequest{
			BadRequest: b,
			msg:        detail.msg,
		})
	}
	return ret
}

func (detail *GRPCErrorDetail) LocalizedMessages() []*LocalizedMessage {
	ret := make([]*LocalizedMessage, 0, len(detail.GRPCErrorDetail.LocalizedMessages))
	for _, m := range detail.GRPCErrorDetail.LocalizedMessages {
		ret = append(ret, &LocalizedMessage{
			LocalizedMessage: m,
			msg:              detail.msg,
		})
	}
	return ret
}

type PreconditionFailure struct {
	*resolver.PreconditionFailure
	msg *Message
}

type PreconditionFailureViolation struct {
	*resolver.PreconditionFailureViolation
	msg *Message
}

func (pf *PreconditionFailure) Violations() []*PreconditionFailureViolation {
	ret := make([]*PreconditionFailureViolation, 0, len(pf.PreconditionFailure.Violations))
	for _, v := range pf.PreconditionFailure.Violations {
		ret = append(ret, &PreconditionFailureViolation{
			PreconditionFailureViolation: v,
			msg:                          pf.msg,
		})
	}
	return ret
}

func (v *PreconditionFailureViolation) CELCacheIndex() int {
	return v.msg.Service.CELCacheIndex()
}

type BadRequest struct {
	*resolver.BadRequest
	msg *Message
}

type BadRequestFieldViolation struct {
	*resolver.BadRequestFieldViolation
	msg *Message
}

func (b *BadRequest) FieldViolations() []*BadRequestFieldViolation {
	ret := make([]*BadRequestFieldViolation, 0, len(b.BadRequest.FieldViolations))
	for _, v := range b.BadRequest.FieldViolations {
		ret = append(ret, &BadRequestFieldViolation{
			BadRequestFieldViolation: v,
			msg:                      b.msg,
		})
	}
	return ret
}

func (v *BadRequestFieldViolation) CELCacheIndex() int {
	return v.msg.Service.CELCacheIndex()
}

type LocalizedMessage struct {
	msg *Message
	*resolver.LocalizedMessage
}

func (m *LocalizedMessage) CELCacheIndex() int {
	return m.msg.Service.CELCacheIndex()
}

func (d *VariableDefinition) ServiceName() string {
	return util.ToPublicGoVariable(d.Service.Name)
}

func (d *VariableDefinition) DependentMethodName() string {
	method := d.VariableDefinition.Expr.Call.Method
	return fmt.Sprintf("%s_%s", fullServiceName(method.Service), util.ToPublicGoVariable(method.Name))
}

func (d *VariableDefinition) RequestType() string {
	expr := d.VariableDefinition.Expr
	switch {
	case expr.Call != nil:
		request := expr.Call.Request
		return fmt.Sprintf("%s.%s",
			request.Type.GoPackage().Name,
			util.ToPublicGoVariable(request.Type.Name),
		)
	case expr.Message != nil:
		msgName := fullMessageName(expr.Message.Message)
		return fmt.Sprintf("%s_%sArgument", d.ServiceName(), msgName)
	}
	return ""
}

func (d *VariableDefinition) LogValueRequestType() string {
	expr := d.VariableDefinition.Expr
	if expr.Call != nil {
		return fullMessageName(expr.Call.Request.Type)
	}
	return ""
}

func (d *VariableDefinition) ReturnType() string {
	expr := d.VariableDefinition.Expr
	switch {
	case expr.Call != nil:
		response := expr.Call.Method.Response
		return fmt.Sprintf("%s.%s",
			response.GoPackage().Name,
			response.Name,
		)
	case expr.Message != nil:
		return expr.Message.Message.Name
	}
	return ""
}

func (d *VariableDefinition) UseResponseVariable() bool {
	return d.VariableDefinition.Used
}

func (d *VariableDefinition) IsBy() bool {
	return d.VariableDefinition.Expr.By != nil
}

func (d *VariableDefinition) IsMap() bool {
	return d.VariableDefinition.Expr.Map != nil
}

func (d *VariableDefinition) IsCall() bool {
	return d.VariableDefinition.Expr.Call != nil
}

func (d *VariableDefinition) MapResolver() *MapResolver {
	return &MapResolver{
		Service: d.Service,
		MapExpr: d.VariableDefinition.Expr.Map,
		file:    d.Message.file,
	}
}

func (d *VariableDefinition) IsValidation() bool {
	return d.VariableDefinition.Expr.Validation != nil
}

func (d *VariableDefinition) By() *resolver.CELValue {
	return d.VariableDefinition.Expr.By
}

func (d *VariableDefinition) ZeroValue() string {
	return toMakeZeroValue(d.Message.file, d.VariableDefinition.Expr.Type)
}

func (d *VariableDefinition) ProtoComment() string {
	opt := &resolver.ProtoFormatOption{
		Prefix:         "",
		IndentSpaceNum: 2,
	}
	return d.VariableDefinition.ProtoFormat(opt)
}

func (d *VariableDefinition) Type() string {
	return d.Message.file.toTypeText(d.VariableDefinition.Expr.Type)
}

func (d *VariableDefinition) CELType() string {
	return toCELNativeType(d.VariableDefinition.Expr.Type)
}

type ResponseVariable struct {
	Name    string
	CELExpr string
	CELType string
}

func (d *VariableDefinition) ResponseVariable() *ResponseVariable {
	if !d.VariableDefinition.Used {
		return nil
	}

	return &ResponseVariable{
		Name:    toUserDefinedVariable(d.VariableDefinition.Name),
		CELExpr: d.VariableDefinition.Name,
		CELType: toCELNativeType(d.VariableDefinition.Expr.Type),
	}
}

type MapResolver struct {
	Service *Service
	MapExpr *resolver.MapExpr
	file    *File
}

func (r *MapResolver) IteratorName() string {
	return r.MapExpr.Iterator.Name
}

func (r *MapResolver) IteratorCELType() string {
	iterType := r.MapExpr.Iterator.Source.Expr.Type.Clone()
	iterType.Repeated = false
	return toCELNativeType(iterType)
}

func (r *MapResolver) IteratorType() string {
	iterType := r.MapExpr.Expr.Type.Clone()
	iterType.Repeated = false
	return r.file.toTypeText(iterType)
}

func (r *MapResolver) IteratorZeroValue() string {
	iterType := r.MapExpr.Expr.Type.Clone()
	iterType.Repeated = false
	return toMakeZeroValue(r.file, iterType)
}

func (r *MapResolver) IteratorSource() string {
	return toUserDefinedVariable(r.MapExpr.Iterator.Source.Name)
}

func (r *MapResolver) IteratorSourceType() string {
	cloned := r.MapExpr.Iterator.Source.Expr.Type.Clone()
	cloned.Repeated = false
	return r.file.toTypeText(cloned)
}

func (r *MapResolver) IsBy() bool {
	return r.MapExpr.Expr.By != nil
}

func (r *MapResolver) IsMessage() bool {
	return r.MapExpr.Expr.Message != nil
}

func (r *MapResolver) Arguments() []*Argument {
	return arguments(r.file, r.MapExpr.Expr.ToVariableExpr())
}

func (r *MapResolver) MapOutType() string {
	cloned := r.MapExpr.Expr.Type.Clone()
	cloned.Repeated = true
	return r.file.toTypeText(cloned)
}

func (r *MapResolver) Caller() string {
	msgName := fullMessageName(r.MapExpr.Expr.Message.Message)
	if r.MapExpr.Expr.Message.Message.HasRule() {
		return fmt.Sprintf("resolve_%s", msgName)
	}
	return fmt.Sprintf("resolver.Resolve_%s", msgName)
}

func (r *MapResolver) RequestType() string {
	expr := r.MapExpr.Expr
	switch {
	case expr.Message != nil:
		msgName := fullMessageName(expr.Message.Message)
		return fmt.Sprintf("%sArgument", msgName)
	}
	return ""
}

func toCELNativeType(t *resolver.Type) string {
	if t.Repeated {
		cloned := t.Clone()
		cloned.Repeated = false
		return fmt.Sprintf("grpcfed.CELListType(%s)", toCELNativeType(cloned))
	}
	if t.IsNull {
		return "grpcfed.CELNullType"
	}
	switch t.Kind {
	case types.Double, types.Float:
		return "grpcfed.CELDoubleType"
	case types.Int32, types.Int64, types.Sint32, types.Sint64, types.Sfixed32, types.Sfixed64, types.Enum:
		return "grpcfed.CELIntType"
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return "grpcfed.CELUintType"
	case types.Bool:
		return "grpcfed.CELBoolType"
	case types.String:
		return "grpcfed.CELStringType"
	case types.Bytes:
		return "grpcfed.CELBytesType"
	case types.Message:
		if t.Message.IsMapEntry {
			return fmt.Sprintf("grpcfed.NewCELMapType(%s, %s)", toCELNativeType(t.Message.Field("key").Type), toCELNativeType(t.Message.Field("value").Type))
		}
		return fmt.Sprintf("grpcfed.CELObjectType(%q)", t.Message.FQDN())
	default:
		log.Fatalf("grpc-federation: specified unsupported type value %s", t.Kind.ToString())
	}
	return ""
}

func (d *VariableDefinition) ValidationError() *GRPCError {
	if d.Expr == nil {
		return nil
	}
	if d.Expr.Validation == nil {
		return nil
	}
	if d.Expr.Validation.Error == nil {
		return nil
	}
	return &GRPCError{
		GRPCError: d.Expr.Validation.Error,
		msg:       d.Message,
	}
}

type Argument struct {
	Name           string
	Value          string
	CEL            *resolver.CELValue
	If             *resolver.CELValue
	InlineFields   []*Argument
	ProtoComment   string
	ZeroValue      string
	Type           string
	OneofName      string
	OneofFieldName string
	RequiredCast   bool
}

func (d *VariableDefinition) Arguments() []*Argument {
	return arguments(d.Message.file, d.Expr)
}

func arguments(file *File, expr *resolver.VariableExpr) []*Argument {
	var (
		isRequestArgument bool
		msgArg            *resolver.Message
		args              []*resolver.Argument
	)
	switch {
	case expr.Call != nil:
		isRequestArgument = true
		args = expr.Call.Request.Args
	case expr.Message != nil:
		msg := expr.Message.Message
		if msg.Rule != nil {
			msgArg = msg.Rule.MessageArgument
		}
		args = expr.Message.Args
	case expr.By != nil:
		return nil
	}
	var generateArgs []*Argument
	for _, arg := range args {
		for _, generatedArg := range argument(file, msgArg, arg) {
			protofmt := arg.ProtoFormat(resolver.DefaultProtoFormatOption, isRequestArgument)
			if protofmt != "" {
				generatedArg.ProtoComment = "// " + protofmt
			}
			generateArgs = append(generateArgs, generatedArg)
		}
	}
	return generateArgs
}

func argument(file *File, msgArg *resolver.Message, arg *resolver.Argument) []*Argument {
	if arg.Value.CEL == nil {
		return nil
	}
	var (
		oneofName      string
		oneofFieldName string
	)
	if arg.Type != nil && arg.Type.OneofField != nil {
		oneofName = util.ToPublicGoVariable(arg.Type.OneofField.Oneof.Name)
		oneofFieldName = strings.TrimPrefix(file.oneofTypeToText(arg.Type.OneofField), "*")
	}
	var inlineFields []*Argument
	if arg.Value.Inline {
		for _, field := range arg.Value.CEL.Out.Message.Fields {
			inlineFields = append(inlineFields, &Argument{
				Name:           util.ToPublicGoVariable(field.Name),
				Value:          fmt.Sprintf("v.Get%s()", util.ToPublicGoVariable(field.Name)),
				OneofName:      oneofName,
				OneofFieldName: oneofFieldName,
				If:             arg.If,
			})
		}
	}
	fromType := arg.Value.Type()
	var toType *resolver.Type
	if arg.Type != nil {
		toType = arg.Type
	} else {
		toType = fromType
	}

	toText := file.toTypeText(toType)
	fromText := file.toTypeText(fromType)

	var (
		argValue       = "v"
		zeroValue      string
		argType        string
		isRequiredCast bool
	)
	switch fromType.Kind {
	case types.Message:
		zeroValue = toMakeZeroValue(file, fromType)
		argType = fromText
		isRequiredCast = requiredCast(fromType, toType)
		if isRequiredCast {
			t := toType.Clone()
			if oneofName != "" {
				t.OneofField = nil
			}
			castFuncName := castFuncName(fromType, t)
			argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
		}
	case types.Enum:
		zeroValue = toMakeZeroValue(file, fromType)
		argType = fromText
		if msgArg != nil && arg.Name != "" {
			msgArgField := msgArg.Field(arg.Name)
			isRequiredCast = msgArgField != nil && msgArgField.Type.Kind != toType.Kind
			if isRequiredCast {
				castFuncName := castFuncName(fromType, msgArgField.Type)
				argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
			}
		}
		if !isRequiredCast {
			isRequiredCast = requiredCast(fromType, toType)
			if isRequiredCast {
				castFuncName := castFuncName(fromType, toType)
				argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
			}
		}
	default:
		// Since fromType is a primitive type, type conversion is possible on the CEL side.
		zeroValue = toMakeZeroValue(file, toType)
		if oneofName != "" {
			t := toType.Clone()
			t.OneofField = nil
			argType = file.toTypeText(t)
		} else {
			argType = toText
		}
		if msgArg != nil && arg.Name != "" {
			msgArgField := msgArg.Field(arg.Name)
			isRequiredCast = msgArgField != nil && msgArgField.Type.Kind != toType.Kind
			if isRequiredCast {
				castFuncName := castFuncName(toType, msgArgField.Type)
				argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
			}
		}
	}
	return []*Argument{
		{
			Name:           util.ToPublicGoVariable(arg.Name),
			Value:          argValue,
			CEL:            arg.Value.CEL,
			InlineFields:   inlineFields,
			ZeroValue:      zeroValue,
			Type:           argType,
			OneofName:      oneofName,
			OneofFieldName: oneofFieldName,
			If:             arg.If,
			RequiredCast:   isRequiredCast,
		},
	}
}

func (s *Service) Messages() []*Message {
	msgs := make([]*Message, 0, len(s.Service.Messages))
	for _, msg := range s.Service.Messages {
		msgs = append(msgs, &Message{
			Message: msg,
			Service: s,
			file:    s.file,
		})
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].ResolverName() < msgs[j].ResolverName()
	})
	return msgs
}

func (s *Service) CastFields() []*CastField {
	castFieldMap := make(map[string]*CastField)
	for _, msg := range s.Service.Messages {
		for _, decl := range msg.TypeConversionDecls() {
			fnName := castFuncName(decl.From, decl.To)
			castFieldMap[fnName] = &CastField{
				Name:     fnName,
				service:  s.Service,
				fromType: decl.From,
				toType:   decl.To,
				file:     s.file,
			}
		}
	}
	ret := make([]*CastField, 0, len(castFieldMap))
	for _, field := range castFieldMap {
		ret = append(ret, field)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret
}

func loadTemplate() (*template.Template, error) {
	tmpl, err := template.New("server.go.tmpl").Funcs(
		map[string]any{
			"add":             Add,
			"map":             CreateMap,
			"parentCtx":       ParentCtx,
			"toLocalVariable": LocalVariable,
		},
	).ParseFS(tmpls, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return tmpl, nil
}

func generateGoContent(tmpl *template.Template, f *File) ([]byte, error) {
	// TODO: Change to evaluate only once.
	var b bytes.Buffer

	// Evaluate template once to create File.pkgMap.
	// pkgMap is created when evaluating typeToText, so all values must be evaluated once.
	if err := tmpl.Execute(&b, f); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	b.Reset()

	// Evaluate the value of File.pkgMap to make the import statement correct and then evaluate it.
	if err := tmpl.Execute(&b, f); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	buf, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format %s: %w", b.String(), err)
	}
	return buf, nil
}

//go:embed templates
var tmpls embed.FS

func toUserDefinedVariable(name string) string {
	return "value.vars." + name
}

func toEnumValuePrefix(file *File, typ *resolver.Type) string {
	enum := typ.Enum
	var name string
	if enum.Message != nil {
		var names []string
		for _, n := range append(enum.Message.ParentMessageNames(), enum.Message.Name) {
			names = append(names, util.ToPublicGoVariable(n))
		}
		name = strings.Join(names, "_")
	} else {
		name = util.ToPublicGoVariable(enum.Name)
	}
	if file.GoPackage.ImportPath == enum.GoPackage().ImportPath {
		return name
	}
	return fmt.Sprintf("%s.%s", enum.GoPackage().Name, name)
}

func toEnumValueText(enumValuePrefix string, value string) string {
	return fmt.Sprintf("%s_%s", enumValuePrefix, value)
}

func protoFQDNToPublicGoName(fqdn string) string {
	names := strings.Split(fqdn, ".")
	formattedNames := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.Replace(name, "-", "_", -1)
		formattedNames = append(formattedNames, util.ToPublicGoVariable(name))
	}
	return strings.Join(formattedNames, "_")
}

func fullServiceName(svc *resolver.Service) string {
	return protoFQDNToPublicGoName(svc.FQDN())
}

func fullOneofName(oneofField *resolver.OneofField) string {
	goName := protoFQDNToPublicGoName(oneofField.FQDN())
	if oneofField.IsConflict() {
		goName += "_"
	}
	return goName
}

func fullMessageName(msg *resolver.Message) string {
	return protoFQDNToPublicGoName(msg.FQDN())
}

func fullEnumName(enum *resolver.Enum) string {
	return protoFQDNToPublicGoName(enum.FQDN())
}

func requiredCast(from, to *resolver.Type) bool {
	if from == nil || to == nil {
		return false
	}
	if from.Kind == types.Message {
		if from.Message.IsMapEntry && to.Message.IsMapEntry {
			return requiredCast(from.Message.Fields[0].Type, to.Message.Fields[0].Type) || requiredCast(from.Message.Fields[1].Type, to.Message.Fields[1].Type)
		}
		return from.Message != to.Message
	}
	if from.Kind == types.Enum {
		return from.Enum != to.Enum
	}
	return from.Kind != to.Kind
}

func castFuncName(from, to *resolver.Type) string {
	return fmt.Sprintf("cast_%s__to__%s", castName(from), castName(to))
}

func castName(typ *resolver.Type) string {
	var ret string
	if typ.Repeated {
		ret = "repeated_"
	}
	switch {
	case typ.OneofField != nil:
		ret += fullOneofName(typ.OneofField)
	case typ.Kind == types.Message:
		if typ.Message.IsMapEntry {
			ret = "map_" + castName(typ.Message.Fields[0].Type) + "_" + castName(typ.Message.Fields[1].Type)
		} else {
			ret += fullMessageName(typ.Message)
		}
	case typ.Kind == types.Enum:
		ret += fullEnumName(typ.Enum)
	default:
		ret += new(File).toTypeText(&resolver.Type{Kind: typ.Kind, IsNull: typ.IsNull})
	}
	return ret
}
