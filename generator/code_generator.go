package generator

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
	"github.com/mercari/grpc-federation/util"
)

type CodeGenerator struct {
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (g *CodeGenerator) Generate(file *resolver.File) ([]byte, error) {
	tmpl, err := loadTemplate("templates/server.go.tmpl")
	if err != nil {
		return nil, err
	}
	return generateGoContent(tmpl, &File{
		File: file,
	})
}

type File struct {
	*resolver.File
}

func (f *File) Services() []*Service {
	ret := make([]*Service, 0, len(f.File.Services))
	for _, svc := range f.File.Services {
		ret = append(ret, &Service{
			Service:           svc,
			nameToLogValueMap: make(map[string]*LogValue),
		})
	}
	return ret
}

type Service struct {
	*resolver.Service
	nameToLogValueMap map[string]*LogValue
}

type Import struct {
	Path  string
	Alias string
}

func (f *File) Imports() []*Import {
	depMap := make(map[string]*resolver.GoPackage)
	for _, svc := range f.File.Services {
		for _, dep := range svc.GoPackageDependencies() {
			depMap[dep.ImportPath] = dep
		}
	}
	imprts := make([]*Import, 0, len(depMap))
	for _, dep := range depMap {
		imprts = append(imprts, &Import{
			Path:  dep.ImportPath,
			Alias: dep.Name,
		})
	}
	sort.Slice(imprts, func(i, j int) bool {
		return imprts[i].Path < imprts[j].Path
	})
	return imprts
}

func (f *File) Types() Types {
	if len(f.File.Services) == 0 {
		return nil
	}
	return newTypeDeclares(f.Messages, f.File.Services[0])
}

func newTypeDeclares(msgs []*resolver.Message, svc *resolver.Service) []*Type {
	declTypes := make(Types, 0, len(msgs))
	typeNameMap := make(map[string]struct{})
	for _, msg := range msgs {
		if !msg.HasRule() {
			continue
		}
		arg := msg.Rule.MessageArgument
		if arg == nil {
			continue
		}
		msgName := fullMessageName(msg)
		argName := fmt.Sprintf("%sArgument", msgName)
		if _, exists := typeNameMap[argName]; exists {
			continue
		}
		msgFQDN := fmt.Sprintf(`%s.%s`, msg.PackageName(), msg.Name)
		typ := &Type{
			Name:      argName,
			Desc:      fmt.Sprintf(`%s is argument for %q message`, argName, msgFQDN),
			ProtoFQDN: arg.FQDN(),
		}

		genMsg := &Message{Message: msg, Service: svc}
		for _, group := range genMsg.MessageResolvers() {
			for _, msgResolver := range group.Resolvers() {
				if msgResolver.MethodCall != nil {
					if msgResolver.MethodCall.Response != nil {
						for _, responseField := range msgResolver.MethodCall.Response.Fields {
							if !responseField.Used {
								continue
							}
							fieldName := util.ToPublicGoVariable(responseField.Name)
							if typ.HasField(fieldName) {
								continue
							}
							typ.Fields = append(typ.Fields, &Field{
								Name: fieldName,
								Type: toTypeText(svc, responseField.Type),
							})
						}
					}
				} else if depMessage := msgResolver.MessageDependency; depMessage != nil && depMessage.Used {
					fieldName := util.ToPublicGoVariable(depMessage.Name)
					if typ.HasField(fieldName) {
						continue
					}
					typ.Fields = append(typ.Fields, &Field{
						Name: fieldName,
						Type: toTypeText(svc, &resolver.Type{
							Type: types.Message,
							Ref:  depMessage.Message,
						}),
					})
				}
			}
		}
		for _, field := range arg.Fields {
			typ.Fields = append(typ.Fields, &Field{
				Name: util.ToPublicGoVariable(field.Name),
				Type: toTypeText(svc, field.Type),
			})
			typ.ProtoFields = append(typ.ProtoFields, &ProtoField{Field: field})
		}
		sort.Slice(typ.Fields, func(i, j int) bool {
			return typ.Fields[i].Name < typ.Fields[j].Name
		})
		typeNameMap[argName] = struct{}{}
		declTypes = append(declTypes, typ)
		for _, field := range msg.CustomResolverFields() {
			typeName := fmt.Sprintf("%s_%sArgument", msgName, util.ToPublicGoVariable(field.Name))
			fields := []*Field{{Type: fmt.Sprintf("*%s", argName)}}
			if msg.HasCustomResolver() {
				fields = append(fields, &Field{
					Name: msgName,
					Type: toTypeText(svc, resolver.NewMessageType(msg, false)),
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
	}
	sort.Slice(declTypes, func(i, j int) bool {
		return declTypes[i].Name < declTypes[j].Name
	})
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

func (s *Service) Types() Types {
	return newTypeDeclares(s.Service.Messages, s.Service)
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
					toMakeZeroValue(svc, field.Type),
				)
				fieldGetterNames = append(
					fieldGetterNames,
					fmt.Sprintf("Get%s", util.ToPublicGoVariable(field.Name)),
				)
			}
			cloned := oneof.Fields[0].Type.Clone()
			cloned.OneofField = nil
			returnZeroValue := toMakeZeroValue(svc, cloned)
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
	return s.Name
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
		return toTypeText(r.Service, ur.Field.Type)
	}
	// message resolver
	return toTypeText(r.Service, &resolver.Type{
		Type: types.Message,
		Ref:  ur.Message,
	})
}

func (s *Service) CustomResolvers() []*CustomResolver {
	resolvers := s.Service.CustomResolvers()
	ret := make([]*CustomResolver, 0, len(resolvers))
	for _, resolver := range resolvers {
		ret = append(ret, &CustomResolver{
			CustomResolver: resolver,
			Service:        s.Service,
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

func toMakeZeroValue(svc *resolver.Service, t *resolver.Type) string {
	text := toTypeText(svc, t)
	if t.Repeated || t.Type == types.Bytes {
		return fmt.Sprintf("%s(nil)", text)
	}
	if t.IsNumber() {
		return fmt.Sprintf("%s(0)", text)
	}
	switch t.Type {
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
		return fmt.Sprintf("celtypes.NewListType(%s)", toCELTypeDeclare(cloned))
	}
	switch t.Type {
	case types.Double, types.Float:
		return "celtypes.DoubleType"
	case types.Int32, types.Int64, types.Sint32, types.Sint64, types.Sfixed32, types.Sfixed64, types.Enum:
		return "celtypes.IntType"
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return "celtypes.UintType"
	case types.Bool:
		return "celtypes.BoolType"
	case types.String:
		return "celtypes.StringType"
	case types.Bytes:
		return "celtypes.BytesType"
	case types.Message:
		return fmt.Sprintf(`celtypes.NewObjectType(%q)`, t.Ref.FQDN())
	}
	return ""
}

func toTypeText(baseSvc *resolver.Service, t *resolver.Type) string {
	if t == nil {
		return "any"
	}
	var typ string
	switch t.Type {
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
		typ = enumTypeToText(baseSvc, t.Enum)
	case types.Message:
		if t.OneofField != nil {
			typ = oneofTypeToText(baseSvc, t.OneofField)
		} else {
			typ = messageTypeToText(baseSvc, t.Ref)
		}
	default:
		log.Fatalf("grpc-federation: specified unsupported type value %s", t.Type)
	}
	if t.Repeated {
		return "[]" + typ
	}
	return typ
}

func oneofTypeToText(baseSvc *resolver.Service, oneofField *resolver.OneofField) string {
	msg := messageTypeToText(baseSvc, oneofField.Oneof.Message)
	oneof := fmt.Sprintf("%s_%s", msg, util.ToPublicGoVariable(oneofField.Field.Name))
	if oneofField.IsConflict() {
		oneof += "_"
	}
	return oneof
}

func messageTypeToText(baseSvc *resolver.Service, msg *resolver.Message) string {
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
		key := toTypeText(baseSvc, msg.Field("key").Type)
		value := toTypeText(baseSvc, msg.Field("value").Type)
		return fmt.Sprintf("map[%s]%s", key, value)
	}
	name := strings.Join(append(msg.ParentMessageNames(), msg.Name), "_")
	if baseSvc.GoPackage().ImportPath == msg.GoPackage().ImportPath {
		return fmt.Sprintf("*%s", name)
	}
	return fmt.Sprintf("*%s.%s", msg.GoPackage().Name, name)
}

func enumTypeToText(baseSvc *resolver.Service, enum *resolver.Enum) string {
	var name string
	if enum.Message != nil {
		name = fmt.Sprintf("%s_%s", enum.Message.Name, enum.Name)
	} else {
		name = enum.Name
	}
	if baseSvc.GoPackage().ImportPath == enum.GoPackage().ImportPath {
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
	return v.Type.Type == types.Message
}

func (v *LogValue) IsMap() bool {
	if !v.IsMessage() {
		return false
	}
	return v.Type.Ref.IsMapEntry
}

type LogValueAttr struct {
	Type  string
	Key   string
	Value string
}

func (s *Service) LogValues() []*LogValue {
	for _, msg := range s.Service.Messages {
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
	msgType := &resolver.Type{Type: types.Message, Ref: msg}
	logValue := &LogValue{
		Name:      name,
		ValueType: toTypeText(s.Service, msgType),
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
	msgType := &resolver.Type{Type: types.Message, Ref: msg}
	logValue := &LogValue{
		Name:      name,
		ValueType: toTypeText(s.Service, msgType),
		Type:      msgType,
		Value:     s.logMapValue(value),
	}
	s.nameToLogValueMap[name] = logValue
	s.setLogValueByType(value.Type)
}

func (s *Service) setLogValueByType(typ *resolver.Type) {
	isMap := typ.Ref != nil && typ.Ref.IsMapEntry
	if typ.Ref != nil {
		s.setLogValueByMessage(typ.Ref)
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
	generics := fmt.Sprintf("[*%sDependentClientSet]", s.Name)
	logValue := &LogValue{
		Name:      name,
		ValueType: "*" + protoFQDNToPublicGoName(arg.FQDN()) + generics,
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
	enumType := &resolver.Type{Type: types.Enum, Enum: enum}
	logValue := &LogValue{
		Name:      name,
		ValueType: toTypeText(s.Service, enumType),
		Attrs:     make([]*LogValueAttr, 0, len(enum.Values)),
		Type:      enumType,
	}
	for _, value := range enum.Values {
		logValue.Attrs = append(logValue.Attrs, &LogValueAttr{
			Key:   value.Value,
			Value: toEnumValueText(toEnumValuePrefix(s.Service, enumType), value.Value),
		})
	}
	s.nameToLogValueMap[name] = logValue
}

func (s *Service) setLogValueByRepeatedType(typ *resolver.Type) {
	if typ.Type != types.Message && typ.Type != types.Enum {
		return
	}
	var (
		name  string
		value string
	)
	if typ.Type == types.Message {
		name = s.repeatedMessageToLogValueName(typ.Ref)
		value = s.messageToLogValueName(typ.Ref)
	} else {
		name = s.repeatedEnumToLogValueName(typ.Enum)
		value = s.enumToLogValueName(typ.Enum)
	}
	if _, exists := s.nameToLogValueMap[name]; exists {
		return
	}
	logValue := &LogValue{
		Name:      name,
		ValueType: toTypeText(s.Service, typ),
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
	switch typ.Type {
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
	log.Fatalf("grpc-federation: specified unknown type value %s", typ.Type)
	return ""
}

func (s *Service) logValue(field *resolver.Field) string {
	typ := field.Type
	base := fmt.Sprintf("v.Get%s()", util.ToPublicGoVariable(field.Name))
	if typ.Type == types.Message {
		return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
	} else if typ.Type == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
		}
		return fmt.Sprintf("%s(%s).String()", s.logValueFuncName(typ), base)
	}
	if field.Type.Repeated {
		return base
	}
	switch field.Type.Type {
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
	if typ.Type == types.Message {
		return fmt.Sprintf("%s(value)", s.logValueFuncName(typ))
	} else if typ.Type == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(value)", s.logValueFuncName(typ))
		}
		return fmt.Sprintf("slog.StringValue(%s(value).String())", s.logValueFuncName(typ))
	} else if typ.Type == types.Bytes {
		return "slog.StringValue(string(value))"
	}
	return "slog.AnyValue(value)"
}

func (s *Service) msgArgumentLogValue(field *resolver.Field) string {
	typ := field.Type
	base := fmt.Sprintf("v.%s", util.ToPublicGoVariable(field.Name))
	if typ.Type == types.Message {
		return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
	} else if typ.Type == types.Enum {
		if typ.Repeated {
			return fmt.Sprintf("%s(%s)", s.logValueFuncName(typ), base)
		}
		return fmt.Sprintf("%s(%s).String()", s.logValueFuncName(typ), base)
	}
	if field.Type.Repeated {
		return base
	}
	switch field.Type.Type {
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
	if typ.Type == types.Message {
		isMap := typ.Ref.IsMapEntry
		if typ.Repeated && !isMap {
			return fmt.Sprintf("s.%s", s.repeatedMessageToLogValueName(typ.Ref))
		}
		return fmt.Sprintf("s.%s", s.messageToLogValueName(typ.Ref))
	}
	if typ.Type == types.Enum {
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
	if msg.Rule.MethodCall != nil && msg.Rule.MethodCall.Method != nil {
		method := msg.Rule.MethodCall.Method
		ret = append(ret, &DependentMethod{
			Name: fmt.Sprintf("%s_%s", fullServiceName(method.Service), method.Name),
			FQDN: "/" + method.FQDN(),
		})
	}
	for _, dep := range msg.Rule.MessageDependencies {
		ret = append(ret, s.dependentMethods(dep.Message)...)
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for _, dep := range field.Rule.Oneof.MessageDependencies {
			ret = append(ret, s.dependentMethods(dep.Message)...)
		}
	}
	return ret
}

type Method struct {
	*resolver.Method
	Service *resolver.Service
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
	return fmt.Sprintf("%sArgument", msg)
}

func (m *Method) RequestType() string {
	return toTypeText(m.Service, &resolver.Type{
		Type: types.Message,
		Ref:  m.Request,
	})
}

func (m *Method) ReturnType() string {
	return toTypeText(m.Service, &resolver.Type{
		Type: types.Message,
		Ref:  m.Response,
	})
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
		methods = append(methods, &Method{Service: s.Service, Method: method})
	}
	return methods
}

type Message struct {
	*resolver.Message
	Service *resolver.Service
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
	if m.Service.GoPackage().ImportPath == m.GoPackage().ImportPath {
		return m.Name
	}
	return fmt.Sprintf("%s.%s", m.GoPackage().Name, m.Name)
}

func (m *Message) LogValueReturnType() string {
	return fullMessageName(m.Message)
}

func (m *Message) MessageResolvers() []*MessageResolverGroup {
	if m.Rule == nil {
		return nil
	}
	var groups []*MessageResolverGroup
	for _, group := range m.Rule.Resolvers {
		groups = append(groups, &MessageResolverGroup{
			Service:              m.Service,
			Message:              m,
			MessageResolverGroup: group,
		})
	}
	return groups
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
	valueMap := map[string]*DeclVariable{
		"sg":      {Name: "sg", Type: "singleflight.Group"},
		"valueMu": {Name: "valueMu", Type: "sync.RWMutex"},
	}
	for _, group := range m.Message.MessageResolvers() {
		for _, r := range group.Resolvers() {
			if r.MethodCall != nil {
				if r.MethodCall.Response == nil {
					continue
				}
				for _, field := range r.MethodCall.Response.Fields {
					if !field.Used {
						continue
					}
					valueMap[field.Name] = &DeclVariable{
						Name: toUserDefinedVariable(field.Name),
						Type: toTypeText(m.Service, field.Type),
					}
				}
			} else if r.MessageDependency != nil {
				if !r.MessageDependency.Used {
					continue
				}
				valueMap[r.MessageDependency.Name] = &DeclVariable{
					Name: toUserDefinedVariable(r.MessageDependency.Name),
					Type: toTypeText(m.Service, &resolver.Type{Type: types.Message, Ref: r.MessageDependency.Message}),
				}
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

func (m *Message) DependencyGraph() string {
	format := m.DependencyGraphTreeFormat()
	if !strings.HasSuffix(format, "\n") {
		// If there is only one node, no newline code is added at the end.
		// In this case, do not display graphs.
		return ""
	}
	return format
}

type ReturnField struct {
	Name                  string
	Value                 string
	IsCustomResolverField bool
	IsOneofField          bool
	CEL                   *resolver.CELValue
	OneofCaseFields       []*OneofField
	OneofDefaultField     *OneofField
	ResolverName          string
	RequestType           string
	MessageName           string
	MessageArgumentName   string
	ProtoComment          string
	ZeroValue             string
}

func (r *ReturnField) HasFieldOneofRule() bool {
	for _, field := range r.OneofCaseFields {
		if field.FieldOneofRule != nil {
			return true
		}
	}
	if r.OneofDefaultField != nil && r.OneofDefaultField.FieldOneofRule != nil {
		return true
	}
	return false
}

type OneofField struct {
	Expr           string
	By             string
	OutType        string
	Condition      string
	Name           string
	Value          string
	Message        *Message
	FieldOneofRule *resolver.FieldOneofRule
}

func (oneof *OneofField) MessageResolvers() []*MessageResolverGroup {
	if oneof.FieldOneofRule == nil {
		return nil
	}
	var groups []*MessageResolverGroup
	for _, group := range oneof.FieldOneofRule.Resolvers {
		groups = append(groups, &MessageResolverGroup{
			Service:              oneof.Message.Service,
			Message:              oneof.Message,
			MessageResolverGroup: group,
		})
	}
	return groups
}

type CastField struct {
	Name     string
	service  *resolver.Service
	fromType *resolver.Type
	toType   *resolver.Type
}

func (f *CastField) RequestType() string {
	if f.fromType.OneofField != nil {
		typ := f.fromType.OneofField.Type.Clone()
		typ.OneofField = nil
		return toTypeText(f.service, typ)
	}
	return toTypeText(f.service, f.fromType)
}

func (f *CastField) ResponseType() string {
	return toTypeText(f.service, f.toType)
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
	return f.fromType.Type == types.Message
}

func (f *CastField) IsEnum() bool {
	return f.fromType.Type == types.Enum
}

type CastEnum struct {
	FromValues   []*CastEnumValue
	DefaultValue string
}

type CastEnumValue struct {
	FromValue string
	ToValue   string
}

func (f *CastField) ToEnum() *CastEnum {
	toEnum := f.toType.Enum
	if toEnum.Rule != nil && toEnum.Rule.Alias != nil {
		return f.toEnumByAlias()
	}
	fromEnum := f.fromType.Enum
	toEnumName := toEnumValuePrefix(f.service, f.toType)
	fromEnumName := toEnumValuePrefix(f.service, f.fromType)
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
	}
}

func (f *CastField) toEnumByAlias() *CastEnum {
	toEnum := f.toType.Enum
	toEnumName := toEnumValuePrefix(f.service, f.toType)
	fromEnumName := toEnumValuePrefix(f.service, f.fromType)
	var (
		enumValues   []*CastEnumValue
		defaultValue = "0"
	)
	for _, toValue := range toEnum.Values {
		if toValue.Rule == nil {
			continue
		}
		for _, alias := range toValue.Rule.Aliases {
			enumValues = append(enumValues, &CastEnumValue{
				FromValue: toEnumValueText(fromEnumName, alias.Value),
				ToValue:   toEnumValueText(toEnumName, toValue.Value),
			})
		}
		if toValue.Rule.Default {
			defaultValue = toEnumValueText(toEnumName, toValue.Value)
		}
	}
	return &CastEnum{
		FromValues:   enumValues,
		DefaultValue: defaultValue,
	}
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
	toMsg := f.toType.Ref
	fromMsg := f.fromType.Ref

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

	name := strings.Join(append(toMsg.ParentMessageNames(), toMsg.Name), "_")
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
	var (
		fromField *resolver.Field
		toType    *resolver.Type
		fromType  *resolver.Type
	)
	if toField.HasRule() && toField.Rule.Alias != nil {
		fromField = toField.Rule.Alias
		fromType = fromField.Type
		toType = toField.Type
	} else {
		fromField = fromMsg.Field(toField.Name)
		if fromField == nil {
			return nil
		}
		fromType = fromField.Type
		toType = toField.Type
	}
	if fromType.Type != toType.Type {
		return nil
	}
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
	names := append(
		msg.ParentMessageNames(),
		msg.Name,
		util.ToPublicGoVariable(toField.Name),
	)
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

func (m *Message) CustomResolverArguments() []*Argument {
	args := []*Argument{}
	argNameMap := make(map[string]struct{})
	for _, group := range m.MessageResolvers() {
		for _, msgResolver := range group.Resolvers() {
			if msgResolver.MethodCall != nil {
				if msgResolver.MethodCall.Response != nil {
					for _, field := range msgResolver.MethodCall.Response.Fields {
						if !field.Used {
							continue
						}
						name := field.Name
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
			} else if depMessage := msgResolver.MessageDependency; depMessage != nil && depMessage.Used {
				name := msgResolver.MessageDependency.Name
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
	}
	sort.Slice(args, func(i, j int) bool {
		return args[i].Name < args[j].Name
	})
	return args
}

func (m *Message) ReturnFields() []*ReturnField {
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
			returnFields = append(returnFields, m.customResolverToReturnField(field))
		case rule.Value != nil:
			value := rule.Value
			if value.Const != nil {
				returnFields = append(returnFields, m.constValueToReturnField(field, value.Const))
			} else {
				returnFields = append(returnFields, m.celValueToReturnField(field, value.CEL))
			}
		case rule.AutoBindField != nil:
			returnFields = append(returnFields, m.autoBindFieldToReturnField(field, rule.AutoBindField))
		}
	}
	for _, oneof := range m.Oneofs {
		returnFields = append(returnFields, m.oneofValueToReturnField(oneof))
	}
	return returnFields
}

func (m *Message) autoBindFieldToReturnField(field *resolver.Field, autoBindField *resolver.AutoBindField) *ReturnField {
	var name string
	switch {
	case autoBindField.ResponseField != nil:
		name = autoBindField.ResponseField.Name
	case autoBindField.MessageDependency != nil:
		name = autoBindField.MessageDependency.Name
	}

	valueName := toUserDefinedVariable(name)
	fieldName := util.ToPublicGoVariable(field.Name)

	var returnFieldValue string
	if field.RequiredTypeConversion() {
		fromType := field.SourceType()
		toType := field.Type
		castFuncName := castFuncName(fromType, toType)
		returnFieldValue = fmt.Sprintf("s.%s(%s.Get%s())", castFuncName, valueName, fieldName)
	} else {
		returnFieldValue = fmt.Sprintf("%s.Get%s()", valueName, fieldName)
	}
	return &ReturnField{
		Name:         fieldName,
		Value:        returnFieldValue,
		ProtoComment: fmt.Sprintf(`// { name: %q, autobind: true }`, name),
	}
}

func (m *Message) constValueToReturnField(field *resolver.Field, constValue *resolver.ConstValue) *ReturnField {
	return &ReturnField{
		Name:  util.ToPublicGoVariable(field.Name),
		Value: toGoConstValue(m.Service, constValue.Type, constValue.Value),
		ProtoComment: field.Rule.ProtoFormat(&resolver.ProtoFormatOption{
			Prefix:         "// ",
			IndentSpaceNum: 2,
		}),
	}
}

func (m *Message) celValueToReturnField(field *resolver.Field, value *resolver.CELValue) *ReturnField {
	toType := field.Type
	fromType := value.Out

	toText := toTypeText(m.Service, toType)
	fromText := toTypeText(m.Service, fromType)

	var (
		returnFieldValue string
		zeroValue        string
	)
	switch fromType.Type {
	case types.Message:
		zeroValue = toMakeZeroValue(m.Service, fromType)
		returnFieldValue = fmt.Sprintf("_value.(%s)", fromText)
		if field.RequiredTypeConversion() {
			castFuncName := castFuncName(fromType, toType)
			returnFieldValue = fmt.Sprintf("s.%s(%s)", castFuncName, returnFieldValue)
		}
	case types.Enum:
		zeroValue = toMakeZeroValue(m.Service, fromType)
		returnFieldValue = fmt.Sprintf("_value.(%s)", fromText)
		if field.RequiredTypeConversion() {
			castFuncName := castFuncName(fromType, toType)
			returnFieldValue = fmt.Sprintf("s.%s(%s)", castFuncName, returnFieldValue)
		}
	default:
		// Since fromType is a primitive type, type conversion is possible on the CEL side.
		zeroValue = toMakeZeroValue(m.Service, toType)
		returnFieldValue = fmt.Sprintf("_value.(%s)", toText)
	}
	return &ReturnField{
		Name:  util.ToPublicGoVariable(field.Name),
		Value: returnFieldValue,
		CEL:   value,
		ProtoComment: field.Rule.ProtoFormat(&resolver.ProtoFormatOption{
			Prefix:         "// ",
			IndentSpaceNum: 2,
		}),
		ZeroValue: zeroValue,
	}
}

func (m *Message) customResolverToReturnField(field *resolver.Field) *ReturnField {
	msgName := fullMessageName(m.Message)
	resolverName := fmt.Sprintf("Resolve_%s_%s", msgName, util.ToPublicGoVariable(field.Name))
	requestType := fmt.Sprintf("%s_%sArgument", msgName, util.ToPublicGoVariable(field.Name))
	return &ReturnField{
		Name:                  util.ToPublicGoVariable(field.Name),
		IsCustomResolverField: true,
		ResolverName:          resolverName,
		RequestType:           requestType,
		MessageArgumentName:   fmt.Sprintf("%sArgument", msgName),
		MessageName:           msgName,
	}
}

func (m *Message) oneofValueToReturnField(oneof *resolver.Oneof) *ReturnField {
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
			returnField := m.autoBindFieldToReturnField(field, rule.AutoBindField)
			caseFields = append(caseFields, &OneofField{
				Condition: fmt.Sprintf("%s != nil", returnField.Value),
				Name:      returnField.Name,
				Value:     returnField.Value,
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
				outType  string
				argValue string
			)
			switch fromType.Type {
			case types.Message:
				outType = "nil"
				argValue = fmt.Sprintf("_value.(%s)", toTypeText(m.Service, fromType))
				if requiredCast(fromType, toType) {
					castFuncName := castFuncName(fromType, toType)
					argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
				}
			case types.Enum:
				outType = "nil"
				argValue = fmt.Sprintf("%s(_value.(int64))", toTypeText(m.Service, fromType))
				if requiredCast(fromType, toType) {
					castFuncName := castFuncName(fromType, toType)
					argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
				}
			default:
				// Since fromType is a primitive type, type conversion is possible on the CEL side.
				outType = fmt.Sprintf("reflect.TypeOf(new(%s).Get%s())", oneofTypeName, fieldName)
				argValue = fmt.Sprintf("_value.(%s)", toTypeText(m.Service, toType))
			}
			if rule.Oneof.Default {
				defaultField = &OneofField{
					By:             rule.Oneof.By.Expr,
					OutType:        outType,
					Condition:      fmt.Sprintf(`oneof_%s.(bool)`, fieldName),
					Name:           fieldName,
					Value:          fmt.Sprintf("&%s{%s: %s}", oneofTypeName, fieldName, argValue),
					FieldOneofRule: rule.Oneof,
					Message:        m,
				}
			} else {
				caseFields = append(caseFields, &OneofField{
					Expr:           rule.Oneof.Expr.Expr,
					By:             rule.Oneof.By.Expr,
					OutType:        outType,
					Condition:      fmt.Sprintf(`oneof_%s.(bool)`, fieldName),
					Name:           fieldName,
					Value:          fmt.Sprintf("&%s{%s: %s}", oneofTypeName, fieldName, argValue),
					FieldOneofRule: rule.Oneof,
					Message:        m,
				})
			}
		}
	}
	return &ReturnField{
		Name:              util.ToPublicGoVariable(oneof.Name),
		IsOneofField:      true,
		OneofCaseFields:   caseFields,
		OneofDefaultField: defaultField,
	}
}

type MessageResolverGroup struct {
	Service *resolver.Service
	Message *Message
	resolver.MessageResolverGroup
}

func (g *MessageResolverGroup) IsConcurrent() bool {
	return g.Type() == resolver.ConcurrentMessageResolverGroupType
}

func (g *MessageResolverGroup) ExistsStart() bool {
	rg, ok := g.MessageResolverGroup.(*resolver.SequentialMessageResolverGroup)
	if !ok {
		return false
	}
	return rg.Start != nil
}

func (g *MessageResolverGroup) ExistsEnd() bool {
	switch rg := g.MessageResolverGroup.(type) {
	case *resolver.SequentialMessageResolverGroup:
		return rg.End != nil
	case *resolver.ConcurrentMessageResolverGroup:
		return rg.End != nil
	}
	return false
}

func (g *MessageResolverGroup) Starts() []*MessageResolverGroup {
	rg, ok := g.MessageResolverGroup.(*resolver.ConcurrentMessageResolverGroup)
	if !ok {
		return nil
	}
	var starts []*MessageResolverGroup
	for _, start := range rg.Starts {
		starts = append(starts, &MessageResolverGroup{
			Service:              g.Service,
			Message:              g.Message,
			MessageResolverGroup: start,
		})
	}
	return starts
}

func (g *MessageResolverGroup) Start() *MessageResolverGroup {
	rg, ok := g.MessageResolverGroup.(*resolver.SequentialMessageResolverGroup)
	if !ok {
		return nil
	}
	return &MessageResolverGroup{
		Service:              g.Service,
		Message:              g.Message,
		MessageResolverGroup: rg.Start,
	}
}

func (g *MessageResolverGroup) End() *MessageResolver {
	switch rg := g.MessageResolverGroup.(type) {
	case *resolver.SequentialMessageResolverGroup:
		return &MessageResolver{
			Service:         g.Service,
			Message:         g.Message,
			MessageResolver: rg.End,
		}
	case *resolver.ConcurrentMessageResolverGroup:
		return &MessageResolver{
			Service:         g.Service,
			Message:         g.Message,
			MessageResolver: rg.End,
		}
	}
	return nil
}

type MessageResolver struct {
	Service *resolver.Service
	Message *Message
	*resolver.MessageResolver
}

func (r *MessageResolver) Key() string {
	if r.MethodCall != nil {
		return r.MethodCall.Method.FQDN()
	}
	return fmt.Sprintf("%s_%s", r.MessageDependency.Name, r.MessageDependency.Message.FQDN())
}

func (r *MessageResolver) UseTimeout() bool {
	return r.MethodCall != nil && r.MethodCall.Timeout != nil
}

func (r *MessageResolver) UseRetry() bool {
	return r.MethodCall != nil && r.MethodCall.Retry != nil
}

func (r *MessageResolver) MethodFQDN() string {
	return r.MethodCall.Method.FQDN()
}

func (r *MessageResolver) Timeout() string {
	return fmt.Sprintf("%[1]d/* %[1]s */", *r.MethodCall.Timeout)
}

func (r *MessageResolver) Caller() string {
	if r.MethodCall != nil {
		methodCall := r.MethodCall
		methodName := methodCall.Method.Name
		svcName := fullServiceName(methodCall.Method.Service)
		return fmt.Sprintf("client.%sClient.%s", svcName, methodName)
	}
	msgName := fullMessageName(r.MessageDependency.Message)
	if r.MessageDependency.Message.HasRule() {
		return fmt.Sprintf("resolve_%s", msgName)
	}
	return fmt.Sprintf("resolver.Resolve_%s", msgName)
}

func (r *MessageResolver) HasErrorHandler() bool {
	return r.MethodCall != nil
}

func (r *MessageResolver) ServiceName() string {
	return r.Service.Name
}

func (r *MessageResolver) DependentMethodName() string {
	method := r.MethodCall.Method
	return fmt.Sprintf("%s_%s", fullServiceName(method.Service), method.Name)
}

func (r *MessageResolver) RequestType() string {
	if r.MethodCall != nil {
		request := r.MethodCall.Request
		return fmt.Sprintf("%s.%s",
			request.Type.GoPackage().Name,
			request.Type.Name,
		)
	}
	msgName := fullMessageName(r.MessageDependency.Message)
	return fmt.Sprintf("%sArgument[*%sDependentClientSet]", msgName, r.Service.Name)
}

func (r *MessageResolver) ReturnType() string {
	if r.MethodCall != nil {
		response := r.MethodCall.Response
		return fmt.Sprintf("%s.%s",
			response.Type.GoPackage().Name,
			response.Type.Name,
		)
	}
	return r.MessageDependency.Message.Name
}

func (r *MessageResolver) UseResponseVariable() bool {
	if r.MethodCall != nil {
		if r.MethodCall.Response != nil {
			for _, field := range r.MethodCall.Response.Fields {
				if field.Used {
					return true
				}
			}
		}
	}
	if r.MessageDependency == nil {
		return false
	}
	return r.MessageDependency.Used
}

func (r *MessageResolver) ProtoComment() string {
	opt := &resolver.ProtoFormatOption{
		Prefix:         "",
		IndentSpaceNum: 2,
	}
	if r.MethodCall != nil {
		return r.MethodCall.ProtoFormat(opt)
	}
	if r.MessageDependency != nil {
		return r.MessageDependency.ProtoFormat(opt)
	}
	return r.Validation.ProtoFormat(opt)
}

type ResponseVariable struct {
	UseName      bool
	Type         string
	Name         string
	Selector     string
	ProtoComment string
	CELExpr      string
	CELType      string
}

func (r *MessageResolver) ResponseVariable() string {
	if r.MethodCall != nil {
		return fmt.Sprintf("res%s", r.MethodCall.Response.Type.Name)
	}
	return fmt.Sprintf("res%s", r.MessageDependency.Message.Name)
}

func (r *MessageResolver) Type() string {
	if r.MethodCall != nil {
		msg := r.MethodCall.Response.Type
		return toTypeText(r.Service, &resolver.Type{Type: types.Message, Ref: msg})
	}
	msg := r.MessageDependency.Message
	return toTypeText(r.Service, &resolver.Type{Type: types.Message, Ref: msg})
}

func (r *MessageResolver) ResponseVariables() []*ResponseVariable {
	var values []*ResponseVariable
	if r.MethodCall != nil {
		for _, field := range r.MethodCall.Response.Fields {
			var (
				selector string
				typ      *resolver.Type
			)
			if field.FieldName == "" {
				selector = r.ResponseVariable()
				typ = resolver.NewMessageType(r.MethodCall.Response.Type, false)
			} else {
				selector = r.ResponseVariable() + "." + fmt.Sprintf("Get%s()", util.ToPublicGoVariable(field.FieldName))
				typ = r.MethodCall.Response.Type.Field(field.FieldName).Type
			}
			format := field.ProtoFormat(resolver.DefaultProtoFormatOption)
			if format != "" {
				format = "// " + format
			}
			values = append(values, &ResponseVariable{
				UseName:      field.Used,
				Name:         toUserDefinedVariable(field.Name),
				Selector:     selector,
				ProtoComment: format,
				CELExpr:      field.Name,
				CELType:      toCELNativeType(typ),
			})
		}
	} else {
		values = append(values, &ResponseVariable{
			UseName:      r.MessageDependency.Used,
			Name:         toUserDefinedVariable(r.MessageDependency.Name),
			Selector:     r.ResponseVariable(),
			ProtoComment: fmt.Sprintf(`// { name: %q, message: %q ... }`, r.MessageDependency.Name, r.MessageDependency.Message.Name),
			CELExpr:      r.MessageDependency.Name,
			CELType:      toCELNativeType(resolver.NewMessageType(r.MessageDependency.Message, false)),
		})
	}
	return values
}

func toCELNativeType(t *resolver.Type) string {
	if t.Repeated {
		cloned := t.Clone()
		cloned.Repeated = false
		return fmt.Sprintf("cel.ListType(%s)", toCELNativeType(cloned))
	}
	switch t.Type {
	case types.Double, types.Float:
		return "celtypes.DoubleType"
	case types.Int32, types.Int64, types.Sint32, types.Sint64, types.Sfixed32, types.Sfixed64, types.Enum:
		return "celtypes.IntType"
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return "celtypes.UintType"
	case types.Bool:
		return "celtypes.BoolType"
	case types.String:
		return "celtypes.StringType"
	case types.Bytes:
		return "celtypes.BytesType"
	case types.Message:
		return fmt.Sprintf("cel.ObjectType(%q)", t.Ref.FQDN())
	default:
		log.Fatalf("grpc-federation: specified unsupported type value %s", t.Type)
	}
	return ""
}

type ValidationRule struct {
	Name  string
	Error *ValidationError
}

type ValidationError struct {
	Code    code.Code
	Rule    string
	Details []*ValidationErrorDetail
}

// HasRule checks if it has rule or not.
func (v *ValidationError) HasRule() bool {
	return v.Rule != ""
}

// GoGRPCStatusCode converts a gRPC status code to a corresponding Go const name
// e.g. FAILED_PRECONDITION -> FailedPrecondition.
func (v *ValidationError) GoGRPCStatusCode() string {
	strCode := v.Code.String()
	if strCode == "OK" {
		// The only exception that the second character is in capital case as well
		return "OK"
	}

	parts := strings.Split(strCode, "_")
	titles := make([]string, 0, len(parts))
	for _, part := range parts {
		titles = append(titles, cases.Title(language.Und).String(part))
	}
	return strings.Join(titles, "")
}

type ValidationErrorDetail struct {
	Rule                 string
	PreconditionFailures []*PreconditionFailure
	BadRequests          []*BadRequest
	LocalizedMessages    []*LocalizedMessage
}

type PreconditionFailure struct {
	Violations []*PreconditionFailureViolation
}

type PreconditionFailureViolation struct {
	Type        string
	Subject     string
	Description string
}

type BadRequest struct {
	FieldViolations []*BadRequestFieldViolation
}

type BadRequestFieldViolation struct {
	Field       string
	Description string
}

type LocalizedMessage struct {
	Locale  string
	Message string
}

func (r *MessageResolver) MessageValidation() *ValidationRule {
	validation := r.MessageResolver.Validation
	vr := &ValidationRule{
		Name: validation.Name,
		Error: &ValidationError{
			Code:    validation.Error.Code,
			Details: make([]*ValidationErrorDetail, 0, len(validation.Error.Details)),
		},
	}
	if r := validation.Error.Rule; r != nil {
		vr.Error.Rule = r.Expr
	}
	for _, detail := range validation.Error.Details {
		ved := &ValidationErrorDetail{
			Rule: detail.Rule.Expr,
		}
		for _, failure := range detail.PreconditionFailures {
			vs := make([]*PreconditionFailureViolation, 0, len(failure.Violations))
			for _, v := range failure.Violations {
				vs = append(vs, &PreconditionFailureViolation{
					Type:        v.Type.Expr,
					Subject:     v.Subject.Expr,
					Description: v.Description.Expr,
				})
			}
			ved.PreconditionFailures = append(ved.PreconditionFailures, &PreconditionFailure{
				Violations: vs,
			})
		}
		for _, req := range detail.BadRequests {
			vs := make([]*BadRequestFieldViolation, 0, len(req.FieldViolations))
			for _, v := range req.FieldViolations {
				vs = append(vs, &BadRequestFieldViolation{
					Field:       v.Field.Expr,
					Description: v.Description.Expr,
				})
			}
			ved.BadRequests = append(ved.BadRequests, &BadRequest{
				FieldViolations: vs,
			})
		}
		for _, msg := range detail.LocalizedMessages {
			ved.LocalizedMessages = append(ved.LocalizedMessages, &LocalizedMessage{
				Locale:  msg.Locale,
				Message: msg.Message.Expr,
			})
		}
		vr.Error.Details = append(vr.Error.Details, ved)
	}
	return vr
}

type Argument struct {
	Name         string
	Value        string
	CEL          *resolver.CELValue
	InlineFields []*Argument
	ProtoComment string
	ZeroValue    string
}

func (r *MessageResolver) Arguments() []*Argument {
	var args []*resolver.Argument
	if r.MethodCall != nil {
		args = r.MethodCall.Request.Args
	} else {
		args = r.MessageDependency.Args
	}
	var generateArgs []*Argument
	isRequestArgument := r.MethodCall != nil
	if !isRequestArgument {
		generateArgs = append(generateArgs, &Argument{
			Name:  "Client",
			Value: "s.client",
		})
	}
	for _, arg := range args {
		for _, generatedArg := range r.argument(arg.Name, arg.Type, arg.Value) {
			format := arg.ProtoFormat(resolver.DefaultProtoFormatOption, isRequestArgument)
			if format != "" {
				generatedArg.ProtoComment = "// " + format
			}
			generateArgs = append(generateArgs, generatedArg)
		}
	}
	return generateArgs
}

func toValue(svc *resolver.Service, typ *resolver.Type, value *resolver.Value) string {
	if value.Const != nil {
		return toGoConstValue(svc, value.Const.Type, value.Const.Value)
	}
	var selectors []string
	selectors = append(selectors, "foo")
	fromType := value.CEL.Out
	toType := typ
	if !requiredCast(fromType, toType) {
		return strings.Join(selectors, ".")
	}
	castFuncName := castFuncName(fromType, toType)
	return fmt.Sprintf("s.%s(%s)", castFuncName, strings.Join(selectors, "."))
}

func (r *MessageResolver) argument(name string, typ *resolver.Type, value *resolver.Value) []*Argument {
	if value.Const != nil {
		return []*Argument{
			{
				Name:  util.ToPublicGoVariable(name),
				Value: toGoConstValue(r.Service, value.Const.Type, value.Const.Value),
			},
		}
	}
	if value.CEL == nil {
		return nil
	}
	var inlineFields []*Argument
	if value.Inline {
		for _, field := range value.CEL.Out.Ref.Fields {
			inlineFields = append(inlineFields, &Argument{
				Name:  util.ToPublicGoVariable(field.Name),
				Value: fmt.Sprintf("_inlineValue.Get%s()", util.ToPublicGoVariable(field.Name)),
			})
		}
	}
	fromType := value.CEL.Out
	var toType *resolver.Type
	if typ != nil {
		toType = typ
	} else {
		toType = fromType
	}

	toText := toTypeText(r.Service, toType)
	fromText := toTypeText(r.Service, fromType)

	var (
		argValue  string
		zeroValue string
	)
	switch fromType.Type {
	case types.Message:
		zeroValue = toMakeZeroValue(r.Service, fromType)
		argValue = fmt.Sprintf("_value.(%s)", fromText)
		if requiredCast(fromType, toType) {
			castFuncName := castFuncName(fromType, toType)
			argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
		}
	case types.Enum:
		zeroValue = toMakeZeroValue(r.Service, fromType)
		argValue = fmt.Sprintf("_value.(%s)", fromText)
		if requiredCast(fromType, toType) {
			castFuncName := castFuncName(fromType, toType)
			argValue = fmt.Sprintf("s.%s(%s)", castFuncName, argValue)
		}
	default:
		// Since fromType is a primitive type, type conversion is possible on the CEL side.
		zeroValue = toMakeZeroValue(r.Service, toType)
		argValue = fmt.Sprintf("_value.(%s)", toText)
	}
	return []*Argument{
		{
			Name:         util.ToPublicGoVariable(name),
			Value:        argValue,
			CEL:          value.CEL,
			InlineFields: inlineFields,
			ZeroValue:    zeroValue,
		},
	}
}

func toGoConstValue(svc *resolver.Service, typ *resolver.Type, value any) string {
	if typ.Repeated {
		rv := reflect.ValueOf(value)
		length := rv.Len()
		copied := *typ
		copied.Repeated = false
		var values []string
		for i := 0; i < length; i++ {
			values = append(values, toGoConstValue(svc, &copied, rv.Index(i).Interface()))
		}
		return fmt.Sprintf("%s{%s}", toTypeText(svc, typ), strings.Join(values, ","))
	}
	switch typ.Type {
	case types.Bool:
		return fmt.Sprintf("%t", value)
	case types.String:
		if envKey, ok := value.(resolver.EnvKey); ok {
			return fmt.Sprintf(`os.Getenv(%q)`, envKey)
		}
		return strconv.Quote(value.(string))
	case types.Bytes:
		b := value.([]byte)
		byts := make([]string, 0, len(b))
		for _, bb := range b {
			byts = append(byts, fmt.Sprint(bb))
		}
		return fmt.Sprintf("[]byte{%s}", strings.Join(byts, ","))
	case types.Enum:
		enumValue := value.(*resolver.EnumValue)
		prefix := toEnumValuePrefix(svc, &resolver.Type{
			Type: types.Enum,
			Enum: enumValue.Enum,
		})
		return toEnumValueText(prefix, enumValue.Value)
	case types.Message:
		mapV, ok := value.(map[string]*resolver.Value)
		if !ok {
			log.Fatalf("message const value must be map[string]*resolver.Value type. but %T", value)
		}
		if typ.Ref == nil {
			log.Fatal("message reference required")
		}
		var fields []string
		for _, field := range typ.Ref.Fields {
			v, exists := mapV[field.Name]
			if !exists {
				continue
			}
			fields = append(
				fields,
				fmt.Sprintf(
					"%s: %s",
					util.ToPublicGoVariable(field.Name),
					toValue(svc, field.Type, v),
				),
			)
		}
		if svc.GoPackage().ImportPath == typ.Ref.GoPackage().ImportPath {
			return fmt.Sprintf("&%s{%s}", typ.Ref.Name, strings.Join(fields, ","))
		}
		return fmt.Sprintf("&%s.%s{%s}", typ.Ref.GoPackage().Name, typ.Ref.Name, strings.Join(fields, ","))
	default:
		// number value
		return fmt.Sprint(value)
	}
}

func (s *Service) Messages() []*Message {
	msgs := make([]*Message, 0, len(s.Service.Messages))
	for _, msg := range s.Service.Messages {
		msgs = append(msgs, &Message{Message: msg, Service: s.Service})
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

func loadTemplate(path string) (*template.Template, error) {
	tmplContent, err := templates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	tmpl, err := template.New("").Funcs(
		map[string]any{
			"add":       Add,
			"map":       CreateMap,
			"parentCtx": ParentCtx,
		},
	).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", path, err)
	}
	return tmpl, nil
}

func generateGoContent(tmpl *template.Template, params any) ([]byte, error) {
	var b bytes.Buffer
	if err := tmpl.Execute(&b, params); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	buf, err := imports.Process("", b.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to format %s: %w", b.String(), err)
	}
	return buf, nil
}

//go:embed templates/*.tmpl
var templates embed.FS

func toUserDefinedVariable(name string) string {
	return util.ToPrivateGoVariable(fmt.Sprintf("value_%s", name))
}

func toEnumValuePrefix(svc *resolver.Service, typ *resolver.Type) string {
	enum := typ.Enum
	var name string
	if enum.Message != nil {
		name = strings.Join(append(enum.Message.ParentMessageNames(), enum.Message.Name), "_")
	} else {
		name = enum.Name
	}
	if svc.GoPackage().ImportPath == enum.GoPackage().ImportPath {
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
	if from.Type == types.Message {
		return from.Ref != to.Ref
	}
	if from.Type == types.Enum {
		return from.Enum != to.Enum
	}
	return from.Type != to.Type
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
	case typ.Type == types.Message:
		ret += fullMessageName(typ.Ref)
	case typ.Type == types.Enum:
		ret += fullEnumName(typ.Enum)
	default:
		ret += toTypeText(nil, &resolver.Type{Type: typ.Type})
	}
	return ret
}
