package testutil

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

type BuilderReferenceManager struct {
	builders []*FileBuilder
}

func NewBuilderReferenceManager(builders ...*FileBuilder) *BuilderReferenceManager {
	return &BuilderReferenceManager{
		builders: builders,
	}
}

func (m *BuilderReferenceManager) Type(t *testing.T, pkg, typ string) *resolver.Type {
	t.Helper()
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkg {
			continue
		}
		return builder.Type(t, typ)
	}
	t.Fatalf("failed to find %s.%s type", pkg, typ)
	return nil
}

func (m *BuilderReferenceManager) RepeatedType(t *testing.T, pkgName, typeName string) *resolver.Type {
	t.Helper()
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkgName {
			continue
		}
		return builder.RepeatedType(t, typeName)
	}
	t.Fatalf("failed to find %s.%s type", pkgName, typeName)
	return nil
}

func (m *BuilderReferenceManager) Service(t *testing.T, pkgName, svcName string) *resolver.Service {
	t.Helper()
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkgName {
			continue
		}
		for _, svc := range builder.file.Services {
			if svc.Name != svcName {
				continue
			}
			return svc
		}
	}
	t.Fatalf("failed to find %s.%s service", pkgName, svcName)
	return nil
}

func (m *BuilderReferenceManager) Method(t *testing.T, pkgName, svcName, mtdName string) *resolver.Method {
	t.Helper()
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkgName {
			continue
		}
		for _, svc := range builder.file.Services {
			if svc.Name != svcName {
				continue
			}
			for _, mtd := range svc.Methods {
				if mtd.Name != mtdName {
					continue
				}
				return mtd
			}
		}
	}
	t.Fatalf("failed to find %s.%s.%s method", pkgName, svcName, mtdName)
	return nil
}

func (m *BuilderReferenceManager) Message(t *testing.T, pkgName, msgName string) *resolver.Message {
	t.Helper()
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkgName {
			continue
		}
		for _, msg := range builder.file.Messages {
			foundMessage := m.message(msg, pkgName, msgName)
			if foundMessage != nil {
				return foundMessage
			}
		}
	}
	t.Fatalf("failed to find %s.%s message", pkgName, msgName)
	return nil
}

func (m *BuilderReferenceManager) message(msg *resolver.Message, pkgName, msgName string) *resolver.Message {
	if msg.FQDN() == fmt.Sprintf("%s.%s", pkgName, msgName) {
		return msg
	}
	for _, nested := range msg.NestedMessages {
		if msg := m.message(nested, pkgName, msgName); msg != nil {
			return msg
		}
	}
	return nil
}

func (m *BuilderReferenceManager) Field(t *testing.T, pkgName, msgName, fieldName string) *resolver.Field {
	t.Helper()
	field := m.Message(t, pkgName, msgName).Field(fieldName)
	if field == nil {
		t.Fatalf("failed to find %s field from %s.%s message", fieldName, pkgName, msgName)
	}
	return field
}

func (m *BuilderReferenceManager) Enum(t *testing.T, pkgName, enumName string) *resolver.Enum {
	t.Helper()
	enumFQDN := fmt.Sprintf("%s.%s", pkgName, enumName)
	for _, builder := range m.builders {
		if builder.file.Package.Name != pkgName {
			continue
		}
		for _, enum := range builder.file.Enums {
			if enum.FQDN() == enumFQDN {
				return enum
			}
		}
		for _, msg := range builder.file.Messages {
			for _, enum := range msg.Enums {
				if enum.FQDN() == enumFQDN {
					return enum
				}
			}
		}
	}
	t.Fatalf("failed to find %s enum", enumFQDN)
	return nil
}

func (m *BuilderReferenceManager) EnumValue(t *testing.T, pkgName, enumName, valueName string) *resolver.EnumValue {
	t.Helper()
	value := m.Enum(t, pkgName, enumName).Value(valueName)
	if value == nil {
		t.Fatalf("failed to find %s value from %s.%s", valueName, pkgName, enumName)
	}
	return value
}

type FileBuilder struct {
	file            *resolver.File
	typeMap         map[string]*resolver.Type
	repeatedTypeMap map[string]*resolver.Type
}

func NewFileBuilder(fileName string) *FileBuilder {
	return &FileBuilder{
		file:            &resolver.File{Name: fileName},
		typeMap:         make(map[string]*resolver.Type),
		repeatedTypeMap: make(map[string]*resolver.Type),
	}
}

func (b *FileBuilder) SetPackage(name string) *FileBuilder {
	b.file.Package = &resolver.Package{Name: name, Files: []*resolver.File{b.file}}
	return b
}

func (b *FileBuilder) SetGoPackage(importPath, name string) *FileBuilder {
	b.file.GoPackage = &resolver.GoPackage{Name: name, ImportPath: importPath}
	return b
}

func (b *FileBuilder) Type(t *testing.T, name string) *resolver.Type {
	t.Helper()
	typ, exists := b.typeMap[name]
	if !exists {
		t.Fatalf("failed to find %s type", name)
	}
	return typ
}

func (b *FileBuilder) RepeatedType(t *testing.T, name string) *resolver.Type {
	t.Helper()
	typ, exists := b.repeatedTypeMap[name]
	if !exists {
		t.Fatalf("failed to find %s repeated type", name)
	}
	return typ
}

func (b *FileBuilder) addType(name string, typ *resolver.Type) {
	b.typeMap[name] = typ
	b.repeatedTypeMap[name] = &resolver.Type{
		Kind:       typ.Kind,
		Message:    typ.Message,
		Enum:       typ.Enum,
		OneofField: typ.OneofField,
		Repeated:   true,
	}
}

func (b *FileBuilder) AddService(svc *resolver.Service) *FileBuilder {
	svc.File = b.file
	b.file.Services = append(b.file.Services, svc)
	return b
}

func (b *FileBuilder) AddEnum(enum *resolver.Enum) *FileBuilder {
	enum.File = b.file
	typ := &resolver.Type{Kind: types.Enum, Enum: enum}
	var enumName string
	if enum.Message != nil {
		enumName = strings.Join(
			append(enum.Message.ParentMessageNames(), enum.Message.Name, enum.Name),
			".",
		)
	} else {
		enumName = enum.Name
	}
	b.addType(enumName, typ)
	b.file.Enums = append(b.file.Enums, enum)
	return b
}

func (b *FileBuilder) AddMessage(msg *resolver.Message) *FileBuilder {
	typ := resolver.NewMessageType(msg, false)
	msgName := strings.Join(append(msg.ParentMessageNames(), msg.Name), ".")
	b.addType(msgName, typ)
	msg.File = b.file
	for _, enum := range msg.Enums {
		enum.File = b.file
		b.AddEnum(enum)
	}
	for _, m := range msg.NestedMessages {
		m.File = b.file
		b.AddMessage(m)
	}
	b.file.Messages = append(b.file.Messages, msg)
	return b
}

func (b *FileBuilder) Build(t *testing.T) *resolver.File {
	t.Helper()
	return b.file
}

type MessageBuilder struct {
	msg *resolver.Message
}

func NewMessageBuilder(name string) *MessageBuilder {
	return &MessageBuilder{
		msg: &resolver.Message{
			Name: name,
		},
	}
}

func (b *MessageBuilder) SetIsMapEntry(v bool) *MessageBuilder {
	b.msg.IsMapEntry = v
	return b
}

func (b *MessageBuilder) AddMessage(msg *resolver.Message) *MessageBuilder {
	msg.ParentMessage = b.msg
	b.msg.NestedMessages = append(b.msg.NestedMessages, msg)
	return b
}

var addOneofMu sync.Mutex

func (b *MessageBuilder) AddOneof(oneof *resolver.Oneof) *MessageBuilder {
	for idx, oneofField := range oneof.Fields {
		field := b.msg.Field(oneofField.Name)
		oneof.Fields[idx] = field
		field.Oneof = oneof
		addOneofMu.Lock()
		field.Type.OneofField = &resolver.OneofField{Field: field}
		addOneofMu.Unlock()
	}
	b.msg.Oneofs = append(b.msg.Oneofs, oneof)
	oneof.Message = b.msg
	return b
}

func (b *MessageBuilder) AddEnum(enum *resolver.Enum) *MessageBuilder {
	enum.Message = b.msg
	b.msg.Enums = append(b.msg.Enums, enum)
	return b
}

func (b *MessageBuilder) getType(t *testing.T, name string) *resolver.Type {
	t.Helper()
	if b.msg.Name == name {
		return resolver.NewMessageType(b.msg, false)
	}
	for _, enum := range b.msg.Enums {
		if enum.Name == name {
			return &resolver.Type{Kind: types.Enum, Enum: enum}
		}
	}
	for _, msg := range b.msg.NestedMessages {
		if msg.Name == name {
			return resolver.NewMessageType(msg, false)
		}
	}
	t.Fatalf("failed to find %s type in %s message", name, b.msg.Name)
	return nil
}

func (b *MessageBuilder) AddField(name string, typ *resolver.Type) *MessageBuilder {
	b.msg.Fields = append(b.msg.Fields, &resolver.Field{Name: name, Type: typ})
	return b
}

func (b *MessageBuilder) AddFieldWithOneof(name string, typ *resolver.Type, oneof *resolver.Oneof) *MessageBuilder {
	b.msg.Fields = append(b.msg.Fields, &resolver.Field{Name: name, Type: typ, Oneof: oneof})
	return b
}

func (b *MessageBuilder) AddFieldWithSelfType(name string, isRepeated bool) *MessageBuilder {
	typ := resolver.NewMessageType(b.msg, isRepeated)
	b.msg.Fields = append(b.msg.Fields, &resolver.Field{Name: name, Type: typ})
	return b
}

func (b *MessageBuilder) AddFieldWithTypeName(t *testing.T, name, typeName string, isRepeated bool) *MessageBuilder {
	t.Helper()
	typ := b.getType(t, typeName)
	if isRepeated {
		typ.Repeated = true
	}
	return b.AddField(name, typ)
}

func (b *MessageBuilder) AddFieldWithTypeNameAndAlias(t *testing.T, name, typeName string, isRepeated bool, field *resolver.Field) *MessageBuilder {
	t.Helper()
	typ := b.getType(t, typeName)
	if isRepeated {
		typ.Repeated = true
	}
	return b.AddFieldWithAlias(name, typ, field)
}

func (b *MessageBuilder) AddFieldWithTypeNameAndAutoBind(t *testing.T, name, typeName string, isRepeated bool, field *resolver.Field) *MessageBuilder {
	t.Helper()
	typ := b.getType(t, typeName)
	if isRepeated {
		typ.Repeated = true
	}
	b.msg.Fields = append(
		b.msg.Fields,
		&resolver.Field{
			Name: name,
			Type: typ,
			Rule: &resolver.FieldRule{
				AutoBindField: &resolver.AutoBindField{
					Field: field,
				},
			},
		},
	)
	return b
}

func (b *MessageBuilder) AddFieldWithAutoBind(name string, typ *resolver.Type, field *resolver.Field) *MessageBuilder {
	b.msg.Fields = append(
		b.msg.Fields,
		&resolver.Field{
			Name: name,
			Type: typ,
			Rule: &resolver.FieldRule{
				AutoBindField: &resolver.AutoBindField{
					Field: field,
				},
			},
		},
	)
	return b
}

func (b *MessageBuilder) AddFieldWithAlias(name string, typ *resolver.Type, fields ...*resolver.Field) *MessageBuilder {
	b.msg.Fields = append(
		b.msg.Fields,
		&resolver.Field{
			Name: name,
			Type: typ,
			Rule: &resolver.FieldRule{
				Aliases: fields,
			},
		},
	)
	return b
}

func (b *MessageBuilder) AddFieldWithTypeNameAndRule(t *testing.T, name, typeName string, isRepeated bool, rule *resolver.FieldRule) *MessageBuilder {
	t.Helper()
	typ := b.getType(t, typeName)
	if isRepeated {
		typ.Repeated = true
	}
	return b.AddFieldWithRule(name, typ, rule)
}

func (b *MessageBuilder) AddFieldWithRule(name string, typ *resolver.Type, rule *resolver.FieldRule) *MessageBuilder {
	field := &resolver.Field{Name: name, Type: typ, Rule: rule}
	b.msg.Fields = append(b.msg.Fields, field)
	return b
}

func (b *MessageBuilder) SetRule(rule *resolver.MessageRule) *MessageBuilder {
	b.msg.Rule = rule
	return b
}

func (b *MessageBuilder) Build(t *testing.T) *resolver.Message {
	t.Helper()
	return b.msg
}

type OneofBuilder struct {
	oneof *resolver.Oneof
}

func NewOneofBuilder(name string) *OneofBuilder {
	return &OneofBuilder{
		oneof: &resolver.Oneof{
			Name: name,
		},
	}
}

func (b *OneofBuilder) AddFieldNames(fields ...string) *OneofBuilder {
	for _, field := range fields {
		b.oneof.Fields = append(b.oneof.Fields, &resolver.Field{Name: field})
	}
	return b
}

func (b *OneofBuilder) Build(t *testing.T) *resolver.Oneof {
	t.Helper()
	return b.oneof
}

type EnumBuilder struct {
	enum *resolver.Enum
}

func NewEnumBuilder(name string) *EnumBuilder {
	return &EnumBuilder{
		enum: &resolver.Enum{Name: name},
	}
}

func (b *EnumBuilder) AddValue(value string) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, &resolver.EnumValue{Value: value, Enum: b.enum})
	return b
}

func (b *EnumBuilder) AddValueWithDefault(value string) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, &resolver.EnumValue{
		Value: value,
		Rule: &resolver.EnumValueRule{
			Default: true,
		},
		Enum: b.enum,
	})
	return b
}

func (b *EnumBuilder) AddValueWithAlias(value string, aliases ...*resolver.EnumValue) *EnumBuilder {
	var enumValueAliases []*resolver.EnumValueAlias
	for _, alias := range aliases {
		enumValueAliases = append(enumValueAliases, &resolver.EnumValueAlias{
			EnumAlias: alias.Enum,
			Aliases:   []*resolver.EnumValue{alias},
		})
	}
	b.enum.Values = append(b.enum.Values, &resolver.EnumValue{
		Value: value,
		Rule: &resolver.EnumValueRule{
			Aliases: enumValueAliases,
		},
		Enum: b.enum,
	})
	return b
}

func (b *EnumBuilder) AddValueWithRule(value string, rule *resolver.EnumValueRule) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, &resolver.EnumValue{
		Value: value,
		Rule:  rule,
		Enum:  b.enum,
	})
	return b
}

func (b *EnumBuilder) SetAlias(aliases ...*resolver.Enum) *EnumBuilder {
	b.enum.Rule = &resolver.EnumRule{
		Aliases: aliases,
	}
	return b
}

func (b *EnumBuilder) Build(t *testing.T) *resolver.Enum {
	t.Helper()
	return b.enum
}

type EnumValueRuleBuilder struct {
	rule *resolver.EnumValueRule
}

func NewEnumValueRuleBuilder() *EnumValueRuleBuilder {
	return &EnumValueRuleBuilder{
		rule: &resolver.EnumValueRule{},
	}
}

func (b *EnumValueRuleBuilder) SetAlias(aliases ...*resolver.EnumValue) *EnumValueRuleBuilder {
	enumValueAliases := make([]*resolver.EnumValueAlias, 0, len(aliases))
	for _, alias := range aliases {
		enumValueAliases = append(enumValueAliases, &resolver.EnumValueAlias{
			EnumAlias: alias.Enum,
			Aliases:   []*resolver.EnumValue{alias},
		})
	}
	b.rule.Aliases = enumValueAliases
	return b
}

func (b *EnumValueRuleBuilder) SetDefault() *EnumValueRuleBuilder {
	b.rule.Default = true
	return b
}

func (b *EnumValueRuleBuilder) SetAttr(attrs ...*resolver.EnumValueAttribute) *EnumValueRuleBuilder {
	b.rule.Attrs = attrs
	return b
}

func (b *EnumValueRuleBuilder) Build(t *testing.T) *resolver.EnumValueRule {
	t.Helper()
	return b.rule
}

type ServiceBuilder struct {
	svc *resolver.Service
}

func NewServiceBuilder(name string) *ServiceBuilder {
	return &ServiceBuilder{
		svc: &resolver.Service{
			Name:    name,
			Methods: []*resolver.Method{},
		},
	}
}

func (b *ServiceBuilder) AddMethod(name string, req, res *resolver.Message, rule *resolver.MethodRule) *ServiceBuilder {
	b.svc.Methods = append(b.svc.Methods, &resolver.Method{
		Service:  b.svc,
		Name:     name,
		Request:  req,
		Response: res,
		Rule:     rule,
	})
	return b
}

func (b *ServiceBuilder) AddMessage(msg, arg *resolver.Message) *ServiceBuilder {
	if msg != nil {
		b.svc.Messages = append(b.svc.Messages, msg)
	}
	if arg != nil {
		b.svc.MessageArgs = append(b.svc.MessageArgs, arg)
	}
	return b
}

func (b *ServiceBuilder) SetRule(rule *resolver.ServiceRule) *ServiceBuilder {
	b.svc.Rule = rule
	return b
}

func (b *ServiceBuilder) Build(t *testing.T) *resolver.Service {
	t.Helper()
	return b.svc
}

type MessageRuleBuilder struct {
	rule *resolver.MessageRule
	errs []error
}

func NewMessageRuleBuilder() *MessageRuleBuilder {
	return &MessageRuleBuilder{
		rule: &resolver.MessageRule{
			DefSet: &resolver.VariableDefinitionSet{},
		},
	}
}

func (b *MessageRuleBuilder) SetCustomResolver(v bool) *MessageRuleBuilder {
	b.rule.CustomResolver = v
	return b
}

func (b *MessageRuleBuilder) SetAlias(aliases ...*resolver.Message) *MessageRuleBuilder {
	b.rule.Aliases = aliases
	return b
}

func (b *MessageRuleBuilder) SetMessageArgument(msg *resolver.Message) *MessageRuleBuilder {
	b.rule.MessageArgument = msg
	return b
}

func (b *MessageRuleBuilder) SetDependencyGraph(graph *resolver.MessageDependencyGraph) *MessageRuleBuilder {
	b.rule.DefSet.Graph = graph
	return b
}

func (b *MessageRuleBuilder) AddVariableDefinitionGroup(group resolver.VariableDefinitionGroup) *MessageRuleBuilder {
	b.rule.DefSet.Groups = append(b.rule.DefSet.Groups, group)
	return b
}

func (b *MessageRuleBuilder) AddVariableDefinition(def *resolver.VariableDefinition) *MessageRuleBuilder {
	if def.Expr != nil && def.Expr.Map != nil && def.Expr.Map.Iterator != nil {
		name := def.Expr.Map.Iterator.Source.Name
		var found bool
		for _, varDef := range b.rule.DefSet.Definitions() {
			if varDef.Name == name {
				def.Expr.Map.Iterator.Source = varDef
				found = true
				break
			}
		}
		if !found {
			b.errs = append(b.errs, fmt.Errorf("%s variable name is not found", name))
		}
	}
	def.Idx = len(b.rule.DefSet.Definitions())
	b.rule.DefSet.Defs = append(b.rule.DefSet.Defs, def)
	return b
}

func (b *MessageRuleBuilder) Build(t *testing.T) *resolver.MessageRule {
	t.Helper()
	if len(b.errs) != 0 {
		t.Fatal(errors.Join(b.errs...))
	}
	return b.rule
}

func NewVariableDefinition(name string) *resolver.VariableDefinition {
	return &resolver.VariableDefinition{Name: name}
}

type VariableDefinitionBuilder struct {
	def *resolver.VariableDefinition
}

func NewVariableDefinitionBuilder() *VariableDefinitionBuilder {
	return &VariableDefinitionBuilder{
		def: &resolver.VariableDefinition{},
	}
}

func (b *VariableDefinitionBuilder) SetIdx(idx int) *VariableDefinitionBuilder {
	b.def.Idx = idx
	return b
}

func (b *VariableDefinitionBuilder) SetName(v string) *VariableDefinitionBuilder {
	b.def.Name = v
	return b
}

func (b *VariableDefinitionBuilder) SetIf(v string) *VariableDefinitionBuilder {
	b.def.If = &resolver.CELValue{
		Expr: v,
		Out:  resolver.BoolType,
	}
	return b
}

func (b *VariableDefinitionBuilder) SetAutoBind(v bool) *VariableDefinitionBuilder {
	b.def.AutoBind = v
	return b
}

func (b *VariableDefinitionBuilder) SetUsed(v bool) *VariableDefinitionBuilder {
	b.def.Used = v
	return b
}

func (b *VariableDefinitionBuilder) SetBy(v *resolver.CELValue) *VariableDefinitionBuilder {
	b.def.Expr = &resolver.VariableExpr{
		By:   v,
		Type: v.Out,
	}
	return b
}

func (b *VariableDefinitionBuilder) SetMap(v *resolver.MapExpr) *VariableDefinitionBuilder {
	mapExprType := v.Expr.Type.Clone()
	mapExprType.Repeated = true
	b.def.Expr = &resolver.VariableExpr{
		Map:  v,
		Type: mapExprType,
	}
	return b
}

func (b *VariableDefinitionBuilder) SetCall(v *resolver.CallExpr) *VariableDefinitionBuilder {
	b.def.Expr = &resolver.VariableExpr{
		Call: v,
		Type: resolver.NewMessageType(v.Method.Response, false),
	}
	return b
}

func (b *VariableDefinitionBuilder) SetMessage(v *resolver.MessageExpr) *VariableDefinitionBuilder {
	b.def.Expr = &resolver.VariableExpr{
		Message: v,
		Type:    resolver.NewMessageType(v.Message, false),
	}
	return b
}

func (b *VariableDefinitionBuilder) SetEnum(v *resolver.EnumExpr) *VariableDefinitionBuilder {
	b.def.Expr = &resolver.VariableExpr{
		Enum: v,
		Type: resolver.NewEnumType(v.Enum, false),
	}
	return b
}

func (b *VariableDefinitionBuilder) SetValidation(v *resolver.ValidationExpr) *VariableDefinitionBuilder {
	b.def.Expr = &resolver.VariableExpr{
		Validation: v,
		Type:       resolver.BoolType,
	}
	return b
}

func (b *VariableDefinitionBuilder) Build(t *testing.T) *resolver.VariableDefinition {
	t.Helper()
	return b.def
}

type MapExprBuilder struct {
	expr *resolver.MapExpr
}

func NewMapExprBuilder() *MapExprBuilder {
	return &MapExprBuilder{
		expr: &resolver.MapExpr{},
	}
}

func (b *MapExprBuilder) SetIterator(v *resolver.Iterator) *MapExprBuilder {
	b.expr.Iterator = v
	return b
}

func (b *MapExprBuilder) SetExpr(v *resolver.MapIteratorExpr) *MapExprBuilder {
	b.expr.Expr = v
	return b
}

func (b *MapExprBuilder) Build(t *testing.T) *resolver.MapExpr {
	t.Helper()
	return b.expr
}

type IteratorBuilder struct {
	iter *resolver.Iterator
}

func NewIteratorBuilder() *IteratorBuilder {
	return &IteratorBuilder{
		iter: &resolver.Iterator{},
	}
}

func (b *IteratorBuilder) SetName(v string) *IteratorBuilder {
	b.iter.Name = v
	return b
}

func (b *IteratorBuilder) SetSource(v string) *IteratorBuilder {
	b.iter.Source = &resolver.VariableDefinition{
		Name: v,
	}
	return b
}

func (b *IteratorBuilder) Build(t *testing.T) *resolver.Iterator {
	t.Helper()
	return b.iter
}

type MapIteratorExprBuilder struct {
	expr *resolver.MapIteratorExpr
}

func NewMapIteratorExprBuilder() *MapIteratorExprBuilder {
	return &MapIteratorExprBuilder{
		expr: &resolver.MapIteratorExpr{},
	}
}

func (b *MapIteratorExprBuilder) SetBy(v *resolver.CELValue) *MapIteratorExprBuilder {
	b.expr.By = v
	b.expr.Type = v.Out
	return b
}

func (b *MapIteratorExprBuilder) SetMessage(v *resolver.MessageExpr) *MapIteratorExprBuilder {
	b.expr.Message = v
	b.expr.Type = resolver.NewMessageType(v.Message, false)
	return b
}

func (b *MapIteratorExprBuilder) SetEnum(v *resolver.EnumExpr) *MapIteratorExprBuilder {
	b.expr.Enum = v
	b.expr.Type = resolver.NewEnumType(v.Enum, false)
	return b
}

func (b *MapIteratorExprBuilder) Build(t *testing.T) *resolver.MapIteratorExpr {
	t.Helper()
	return b.expr
}

type CallExprBuilder struct {
	expr    *resolver.CallExpr
	timeout string
}

func NewCallExprBuilder() *CallExprBuilder {
	return &CallExprBuilder{
		expr: &resolver.CallExpr{},
	}
}

func (b *CallExprBuilder) SetMethod(v *resolver.Method) *CallExprBuilder {
	b.expr.Method = v
	return b
}

func (b *CallExprBuilder) SetRequest(v *resolver.Request) *CallExprBuilder {
	v.Type = b.expr.Method.Request
	b.expr.Request = v
	return b
}

func (b *CallExprBuilder) SetTimeout(v string) *CallExprBuilder {
	b.timeout = v
	return b
}

func (b *CallExprBuilder) SetRetryIf(expr string) *CallExprBuilder {
	ifValue := &resolver.CELValue{
		Expr: expr,
		Out:  resolver.BoolType,
	}
	if b.expr.Retry != nil {
		b.expr.Retry.If = ifValue
	} else {
		b.expr.Retry = &resolver.RetryPolicy{If: ifValue}
	}
	return b
}

func (b *CallExprBuilder) SetRetryPolicyConstant(constant *resolver.RetryPolicyConstant) *CallExprBuilder {
	defaultIfValue := &resolver.CELValue{
		Expr: "true",
		Out:  resolver.BoolType,
	}
	if b.expr.Retry != nil {
		b.expr.Retry.Constant = constant
		if b.expr.Retry.If == nil {
			b.expr.Retry.If = defaultIfValue
		}
	} else {
		b.expr.Retry = &resolver.RetryPolicy{
			If:       defaultIfValue,
			Constant: constant,
		}
	}
	return b
}

func (b *CallExprBuilder) SetRetryPolicyExponential(exp *resolver.RetryPolicyExponential) *CallExprBuilder {
	defaultIfValue := &resolver.CELValue{
		Expr: "true",
		Out:  resolver.BoolType,
	}
	if b.expr.Retry != nil {
		b.expr.Retry.Exponential = exp
		if b.expr.Retry.If == nil {
			b.expr.Retry.If = defaultIfValue
		}
	} else {
		b.expr.Retry = &resolver.RetryPolicy{
			If:          defaultIfValue,
			Exponential: exp,
		}
	}
	return b
}

func (b *CallExprBuilder) AddError(err *resolver.GRPCError) *CallExprBuilder {
	b.expr.Errors = append(b.expr.Errors, err)
	return b
}

func (b *CallExprBuilder) Build(t *testing.T) *resolver.CallExpr {
	t.Helper()
	if b.timeout != "" {
		timeout, err := time.ParseDuration(b.timeout)
		if err != nil {
			t.Fatal(err)
		}
		b.expr.Timeout = &timeout
		if b.expr.Retry != nil && b.expr.Retry.Exponential != nil {
			b.expr.Retry.Exponential.MaxElapsedTime = timeout
		}
	}
	return b.expr
}

type MessageExprBuilder struct {
	expr *resolver.MessageExpr
}

func NewMessageExprBuilder() *MessageExprBuilder {
	return &MessageExprBuilder{
		expr: &resolver.MessageExpr{
			Args: []*resolver.Argument{},
		},
	}
}

func (b *MessageExprBuilder) SetMessage(v *resolver.Message) *MessageExprBuilder {
	b.expr.Message = v
	return b
}

func (b *MessageExprBuilder) SetArgs(v []*resolver.Argument) *MessageExprBuilder {
	b.expr.Args = v
	return b
}

func (b *MessageExprBuilder) Build(t *testing.T) *resolver.MessageExpr {
	t.Helper()
	return b.expr
}

type EnumExprBuilder struct {
	expr *resolver.EnumExpr
}

func NewEnumExprBuilder() *EnumExprBuilder {
	return &EnumExprBuilder{
		expr: &resolver.EnumExpr{},
	}
}

func (b *EnumExprBuilder) SetEnum(v *resolver.Enum) *EnumExprBuilder {
	b.expr.Enum = v
	return b
}

func (b *EnumExprBuilder) SetBy(v *resolver.CELValue) *EnumExprBuilder {
	b.expr.By = v
	return b
}

func (b *EnumExprBuilder) Build(t *testing.T) *resolver.EnumExpr {
	t.Helper()
	return b.expr
}

type ValidationExprBuilder struct {
	expr *resolver.ValidationExpr
}

func NewValidationExprBuilder() *ValidationExprBuilder {
	return &ValidationExprBuilder{
		expr: &resolver.ValidationExpr{},
	}
}

func (b *ValidationExprBuilder) SetError(err *resolver.GRPCError) *ValidationExprBuilder {
	b.expr.Error = err
	return b
}

func (b *ValidationExprBuilder) Build(t *testing.T) *resolver.ValidationExpr {
	t.Helper()
	return b.expr
}

type GRPCErrorBuilder struct {
	err *resolver.GRPCError
}

func NewGRPCErrorBuilder() *GRPCErrorBuilder {
	return &GRPCErrorBuilder{
		err: &resolver.GRPCError{
			DefSet: &resolver.VariableDefinitionSet{},
			If: &resolver.CELValue{
				Expr: "true",
				Out:  resolver.BoolType,
			},
			LogLevel: slog.LevelError,
		},
	}
}

func (b *GRPCErrorBuilder) AddVariableDefinition(def *resolver.VariableDefinition) *GRPCErrorBuilder {
	b.err.DefSet.Defs = append(b.err.DefSet.Defs, def)
	return b
}

func (b *GRPCErrorBuilder) SetDependencyGraph(graph *resolver.MessageDependencyGraph) *GRPCErrorBuilder {
	b.err.DefSet.Graph = graph
	return b
}

func (b *GRPCErrorBuilder) AddVariableDefinitionGroup(group resolver.VariableDefinitionGroup) *GRPCErrorBuilder {
	b.err.DefSet.Groups = append(b.err.DefSet.Groups, group)
	return b
}

func (b *GRPCErrorBuilder) SetIf(expr string) *GRPCErrorBuilder {
	b.err.If = &resolver.CELValue{
		Expr: expr,
		Out:  resolver.BoolType,
	}
	return b
}

func (b *GRPCErrorBuilder) SetCode(v code.Code) *GRPCErrorBuilder {
	b.err.Code = &v
	return b
}

func (b *GRPCErrorBuilder) SetMessage(v string) *GRPCErrorBuilder {
	b.err.Message = &resolver.CELValue{
		Expr: v,
		Out:  resolver.StringType,
	}
	return b
}

func (b *GRPCErrorBuilder) SetIgnore(v bool) *GRPCErrorBuilder {
	b.err.Ignore = v
	return b
}

func (b *GRPCErrorBuilder) SetIgnoreAndResponse(v string, typ *resolver.Type) *GRPCErrorBuilder {
	b.err.IgnoreAndResponse = &resolver.CELValue{
		Expr: v,
		Out:  typ,
	}
	return b
}

func (b *GRPCErrorBuilder) AddDetail(v *resolver.GRPCErrorDetail) *GRPCErrorBuilder {
	b.err.Details = append(b.err.Details, v)
	return b
}

func (b *GRPCErrorBuilder) SetLogLevel(level slog.Level) *GRPCErrorBuilder {
	b.err.LogLevel = level
	return b
}

func (b *GRPCErrorBuilder) Build(t *testing.T) *resolver.GRPCError {
	t.Helper()
	return b.err
}

type GRPCErrorDetailBuilder struct {
	detail *resolver.GRPCErrorDetail
}

func NewGRPCErrorDetailBuilder() *GRPCErrorDetailBuilder {
	return &GRPCErrorDetailBuilder{
		detail: &resolver.GRPCErrorDetail{
			If: &resolver.CELValue{
				Expr: "true",
				Out:  resolver.BoolType,
			},
			DefSet:   &resolver.VariableDefinitionSet{},
			Messages: &resolver.VariableDefinitionSet{},
		},
	}
}

func (b *GRPCErrorDetailBuilder) SetIf(expr string) *GRPCErrorDetailBuilder {
	b.detail.If = &resolver.CELValue{
		Expr: expr,
		Out:  resolver.BoolType,
	}
	return b
}

func (b *GRPCErrorDetailBuilder) AddDef(v *resolver.VariableDefinition) *GRPCErrorDetailBuilder {
	b.detail.DefSet.Defs = append(b.detail.DefSet.Defs, v)
	return b
}

func (b *GRPCErrorDetailBuilder) AddBy(v *resolver.CELValue) *GRPCErrorDetailBuilder {
	b.detail.By = append(b.detail.By, v)
	return b
}

func (b *GRPCErrorDetailBuilder) AddMessage(v *resolver.VariableDefinition) *GRPCErrorDetailBuilder {
	b.detail.Messages.Defs = append(b.detail.Messages.Defs, v)
	return b
}

func (b *GRPCErrorDetailBuilder) AddPreconditionFailure(v *resolver.PreconditionFailure) *GRPCErrorDetailBuilder {
	b.detail.PreconditionFailures = append(b.detail.PreconditionFailures, v)
	return b
}

func (b *GRPCErrorDetailBuilder) AddBadRequest(v *resolver.BadRequest) *GRPCErrorDetailBuilder {
	b.detail.BadRequests = append(b.detail.BadRequests, v)
	return b
}

func (b *GRPCErrorDetailBuilder) AddLocalizedMessage(v *resolver.LocalizedMessage) *GRPCErrorDetailBuilder {
	b.detail.LocalizedMessages = append(b.detail.LocalizedMessages, v)
	return b
}

func (b *GRPCErrorDetailBuilder) Build(t *testing.T) *resolver.GRPCErrorDetail {
	t.Helper()
	return b.detail
}

type VariableDefinitionGroupBuilder struct {
	starts []resolver.VariableDefinitionGroup
	end    *resolver.VariableDefinition
}

func NewVariableDefinitionGroupByName(name string) *resolver.SequentialVariableDefinitionGroup {
	return &resolver.SequentialVariableDefinitionGroup{
		End: &resolver.VariableDefinition{Name: name},
	}
}

func NewVariableDefinitionGroupBuilder() *VariableDefinitionGroupBuilder {
	return &VariableDefinitionGroupBuilder{}
}

func (b *VariableDefinitionGroupBuilder) AddStart(start resolver.VariableDefinitionGroup) *VariableDefinitionGroupBuilder {
	b.starts = append(b.starts, start)
	return b
}

func (b *VariableDefinitionGroupBuilder) SetEnd(end *resolver.VariableDefinition) *VariableDefinitionGroupBuilder {
	b.end = end
	return b
}

func (b *VariableDefinitionGroupBuilder) Build(t *testing.T) resolver.VariableDefinitionGroup {
	t.Helper()
	if len(b.starts) > 1 {
		return &resolver.ConcurrentVariableDefinitionGroup{
			Starts: b.starts,
			End:    b.end,
		}
	}
	var start resolver.VariableDefinitionGroup
	if len(b.starts) == 1 {
		start = b.starts[0]
	}
	return &resolver.SequentialVariableDefinitionGroup{
		Start: start,
		End:   b.end,
	}
}

type MethodCallBuilder struct {
	call    *resolver.MethodCall
	timeout string
}

func NewMethodCallBuilder(mtd *resolver.Method) *MethodCallBuilder {
	return &MethodCallBuilder{
		call: &resolver.MethodCall{Method: mtd},
	}
}

func (b *MethodCallBuilder) SetRequest(req *resolver.Request) *MethodCallBuilder {
	req.Type = b.call.Method.Request
	b.call.Request = req
	return b
}

func (b *MethodCallBuilder) SetTimeout(timeout string) *MethodCallBuilder {
	b.timeout = timeout
	return b
}

func (b *MethodCallBuilder) SetRetryPolicyConstant(constant *resolver.RetryPolicyConstant) *MethodCallBuilder {
	b.call.Retry = &resolver.RetryPolicy{
		Constant: constant,
	}
	return b
}

func (b *MethodCallBuilder) SetRetryPolicyExponential(exp *resolver.RetryPolicyExponential) *MethodCallBuilder {
	b.call.Retry = &resolver.RetryPolicy{
		Exponential: exp,
	}
	return b
}

func (b *MethodCallBuilder) Build(t *testing.T) *resolver.MethodCall {
	t.Helper()
	if b.timeout != "" {
		timeout, err := time.ParseDuration(b.timeout)
		if err != nil {
			t.Fatal(err)
		}
		b.call.Timeout = &timeout
		if b.call.Retry != nil && b.call.Retry.Exponential != nil {
			b.call.Retry.Exponential.MaxElapsedTime = timeout
		}
	}
	return b.call
}

type RetryPolicyConstantBuilder struct {
	constant *resolver.RetryPolicyConstant
	interval string
}

func NewRetryPolicyConstantBuilder() *RetryPolicyConstantBuilder {
	return &RetryPolicyConstantBuilder{
		constant: &resolver.RetryPolicyConstant{
			Interval:   resolver.DefaultRetryConstantInterval,
			MaxRetries: resolver.DefaultRetryMaxRetryCount,
		},
	}
}

func (b *RetryPolicyConstantBuilder) SetInterval(interval string) *RetryPolicyConstantBuilder {
	b.interval = interval
	return b
}

func (b *RetryPolicyConstantBuilder) SetMaxRetries(maxRetries uint64) *RetryPolicyConstantBuilder {
	b.constant.MaxRetries = maxRetries
	return b
}

func (b *RetryPolicyConstantBuilder) Build(t *testing.T) *resolver.RetryPolicyConstant {
	t.Helper()
	if b.interval != "" {
		interval, err := time.ParseDuration(b.interval)
		if err != nil {
			t.Fatal(err)
		}
		b.constant.Interval = interval
	}
	return b.constant
}

type RetryPolicyExponentialBuilder struct {
	exp             *resolver.RetryPolicyExponential
	initialInterval string
	maxInterval     string
}

func NewRetryPolicyExponentialBuilder() *RetryPolicyExponentialBuilder {
	return &RetryPolicyExponentialBuilder{
		exp: &resolver.RetryPolicyExponential{
			InitialInterval:     resolver.DefaultRetryExponentialInitialInterval,
			RandomizationFactor: resolver.DefaultRetryExponentialRandomizationFactor,
			Multiplier:          resolver.DefaultRetryExponentialMultiplier,
			MaxRetries:          resolver.DefaultRetryMaxRetryCount,
		},
	}
}

func (b *RetryPolicyExponentialBuilder) SetInitialInterval(interval string) *RetryPolicyExponentialBuilder {
	b.initialInterval = interval
	return b
}

func (b *RetryPolicyExponentialBuilder) SetRandomizationFactor(factor float64) *RetryPolicyExponentialBuilder {
	b.exp.RandomizationFactor = factor
	return b
}

func (b *RetryPolicyExponentialBuilder) SetMultiplier(multiplier float64) *RetryPolicyExponentialBuilder {
	b.exp.Multiplier = multiplier
	return b
}

func (b *RetryPolicyExponentialBuilder) SetMaxInterval(interval string) *RetryPolicyExponentialBuilder {
	b.maxInterval = interval
	return b
}

func (b *RetryPolicyExponentialBuilder) SetMaxRetries(maxRetries uint64) *RetryPolicyExponentialBuilder {
	b.exp.MaxRetries = maxRetries
	return b
}

func (b *RetryPolicyExponentialBuilder) Build(t *testing.T) *resolver.RetryPolicyExponential {
	t.Helper()
	if b.initialInterval != "" {
		interval, err := time.ParseDuration(b.initialInterval)
		if err != nil {
			t.Fatal(err)
		}
		b.exp.InitialInterval = interval
	}
	if b.maxInterval != "" {
		interval, err := time.ParseDuration(b.maxInterval)
		if err != nil {
			t.Fatal(err)
		}
		b.exp.MaxInterval = interval
	}
	return b.exp
}

type RequestBuilder struct {
	req *resolver.Request
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{req: &resolver.Request{}}
}

func (b *RequestBuilder) AddField(name string, typ *resolver.Type, value *resolver.Value) *RequestBuilder {
	b.req.Args = append(b.req.Args, &resolver.Argument{
		Name:  name,
		Type:  typ,
		Value: value,
	})
	return b
}

func (b *RequestBuilder) AddFieldWithIf(name string, typ *resolver.Type, value *resolver.Value, ifValue string) *RequestBuilder {
	b.req.Args = append(b.req.Args, &resolver.Argument{
		Name:  name,
		Type:  typ,
		Value: value,
		If: &resolver.CELValue{
			Expr: ifValue,
			Out:  resolver.BoolType,
		},
	})
	return b
}

func (b *RequestBuilder) Build(t *testing.T) *resolver.Request {
	t.Helper()
	return b.req
}

type MessageDependencyArgumentBuilder struct {
	args []*resolver.Argument
}

func NewMessageDependencyArgumentBuilder() *MessageDependencyArgumentBuilder {
	return &MessageDependencyArgumentBuilder{args: []*resolver.Argument{}}
}

func (b *MessageDependencyArgumentBuilder) Add(name string, value *resolver.Value) *MessageDependencyArgumentBuilder {
	b.args = append(b.args, &resolver.Argument{
		Name:  name,
		Value: value,
	})
	return b
}

func (b *MessageDependencyArgumentBuilder) Inline(value *resolver.Value) *MessageDependencyArgumentBuilder {
	if value.CEL == nil {
		value.CEL = &resolver.CELValue{}
	}
	value.Inline = true
	b.args = append(b.args, &resolver.Argument{Value: value})
	return b
}

func (b *MessageDependencyArgumentBuilder) Build(t *testing.T) []*resolver.Argument {
	t.Helper()
	return b.args
}

type MessageArgumentValueBuilder struct {
	value *resolver.Value
}

func NewMessageArgumentValueBuilder(ref, filtered *resolver.Type, expr string) *MessageArgumentValueBuilder {
	return &MessageArgumentValueBuilder{
		value: &resolver.Value{
			CEL: &resolver.CELValue{
				Expr: expr,
				Out:  filtered,
			},
		},
	}
}

func (b *MessageArgumentValueBuilder) Build(t *testing.T) *resolver.Value {
	t.Helper()
	return b.value
}

type CELValueBuilder struct {
	cel *resolver.CELValue
}

func NewCELValueBuilder(expr string, out *resolver.Type) *CELValueBuilder {
	return &CELValueBuilder{
		cel: &resolver.CELValue{
			Expr: expr,
			Out:  out,
		},
	}
}

func (b *CELValueBuilder) Build(t *testing.T) *resolver.CELValue {
	t.Helper()
	return b.cel
}

type NameReferenceValueBuilder struct {
	value *resolver.Value
}

func NewNameReferenceValueBuilder(ref, filtered *resolver.Type, expr string) *NameReferenceValueBuilder {
	return &NameReferenceValueBuilder{
		value: &resolver.Value{
			CEL: &resolver.CELValue{
				Expr: expr,
				Out:  filtered,
			},
		},
	}
}

func (b *NameReferenceValueBuilder) Build(t *testing.T) *resolver.Value {
	t.Helper()
	return b.value
}

type FieldRuleBuilder struct {
	rule *resolver.FieldRule
}

func NewFieldRuleBuilder(value *resolver.Value) *FieldRuleBuilder {
	return &FieldRuleBuilder{rule: &resolver.FieldRule{Value: value}}
}

func (b *FieldRuleBuilder) SetAutoBind(field *resolver.Field) *FieldRuleBuilder {
	b.rule.AutoBindField = &resolver.AutoBindField{
		Field: field,
	}
	return b
}

func (b *FieldRuleBuilder) SetCustomResolver(v bool) *FieldRuleBuilder {
	b.rule.CustomResolver = v
	return b
}

func (b *FieldRuleBuilder) SetMessageCustomResolver(v bool) *FieldRuleBuilder {
	b.rule.MessageCustomResolver = v
	return b
}

func (b *FieldRuleBuilder) SetAlias(v ...*resolver.Field) *FieldRuleBuilder {
	b.rule.Aliases = v
	return b
}

func (b *FieldRuleBuilder) SetOneof(v *resolver.FieldOneofRule) *FieldRuleBuilder {
	b.rule.Oneof = v
	return b
}

func (b *FieldRuleBuilder) Build(t *testing.T) *resolver.FieldRule {
	t.Helper()
	return b.rule
}

type DependencyGraphBuilder struct {
	graph *resolver.MessageDependencyGraph
}

func NewDependencyGraphBuilder() *DependencyGraphBuilder {
	return &DependencyGraphBuilder{
		graph: &resolver.MessageDependencyGraph{},
	}
}

func (b *DependencyGraphBuilder) Add(parent *resolver.Message, children ...*resolver.Message) *DependencyGraphBuilder {
	nodes := make([]*resolver.MessageDependencyGraphNode, 0, len(children))
	for _, child := range children {
		nodes = append(nodes, &resolver.MessageDependencyGraphNode{BaseMessage: child})
	}
	b.graph.Roots = append(b.graph.Roots, &resolver.MessageDependencyGraphNode{
		BaseMessage: parent,
		Children:    nodes,
	})
	return b
}

func (b *DependencyGraphBuilder) Build(t *testing.T) *resolver.MessageDependencyGraph {
	t.Helper()
	return b.graph
}

type ServiceRuleBuilder struct {
	rule *resolver.ServiceRule
}

func NewServiceRuleBuilder() *ServiceRuleBuilder {
	return &ServiceRuleBuilder{
		rule: &resolver.ServiceRule{},
	}
}

func (b *ServiceRuleBuilder) SetEnv(env *resolver.Env) *ServiceRuleBuilder {
	b.rule.Env = env
	return b
}

func (b *ServiceRuleBuilder) Build(t *testing.T) *resolver.ServiceRule {
	t.Helper()
	return b.rule
}

type EnvBuilder struct {
	env *resolver.Env
}

func NewEnvBuilder() *EnvBuilder {
	return &EnvBuilder{
		env: &resolver.Env{},
	}
}

func (b *EnvBuilder) AddVar(v *resolver.EnvVar) *EnvBuilder {
	b.env.Vars = append(b.env.Vars, v)
	return b
}

func (b *EnvBuilder) Build(t *testing.T) *resolver.Env {
	t.Helper()
	return b.env
}

type EnvVarBuilder struct {
	v *resolver.EnvVar
}

func NewEnvVarBuilder() *EnvVarBuilder {
	return &EnvVarBuilder{
		v: &resolver.EnvVar{},
	}
}

func (b *EnvVarBuilder) SetName(name string) *EnvVarBuilder {
	b.v.Name = name
	return b
}

func (b *EnvVarBuilder) SetType(typ *resolver.Type) *EnvVarBuilder {
	b.v.Type = typ
	return b
}

func (b *EnvVarBuilder) SetOption(opt *resolver.EnvVarOption) *EnvVarBuilder {
	b.v.Option = opt
	return b
}

func (b *EnvVarBuilder) Build(t *testing.T) *resolver.EnvVar {
	t.Helper()
	return b.v
}

type EnvVarOptionBuilder struct {
	opt *resolver.EnvVarOption
}

func NewEnvVarOptionBuilder() *EnvVarOptionBuilder {
	return &EnvVarOptionBuilder{
		opt: &resolver.EnvVarOption{},
	}
}

func (b *EnvVarOptionBuilder) SetDefault(v string) *EnvVarOptionBuilder {
	b.opt.Default = v
	return b
}

func (b *EnvVarOptionBuilder) SetAlternate(v string) *EnvVarOptionBuilder {
	b.opt.Alternate = v
	return b
}

func (b *EnvVarOptionBuilder) SetRequired(v bool) *EnvVarOptionBuilder {
	b.opt.Required = v
	return b
}

func (b *EnvVarOptionBuilder) SetIgnored(v bool) *EnvVarOptionBuilder {
	b.opt.Ignored = v
	return b
}

func (b *EnvVarOptionBuilder) Build(t *testing.T) *resolver.EnvVarOption {
	t.Helper()
	return b.opt
}

type MethodRuleBuilder struct {
	duration string
	rule     *resolver.MethodRule
}

func NewMethodRuleBuilder() *MethodRuleBuilder {
	return &MethodRuleBuilder{
		rule: &resolver.MethodRule{},
	}
}

func (b *MethodRuleBuilder) Timeout(duration string) *MethodRuleBuilder {
	b.duration = duration
	return b
}

func (b *MethodRuleBuilder) Response(msg *resolver.Message) *MethodRuleBuilder {
	b.rule.Response = msg
	return b
}

func (b *MethodRuleBuilder) Build(t *testing.T) *resolver.MethodRule {
	t.Helper()
	if b.duration != "" {
		duration, err := time.ParseDuration(b.duration)
		if err != nil {
			t.Fatal(err)
		}
		b.rule.Timeout = &duration
	}
	return b.rule
}

type FieldOneofRuleBuilder struct {
	errs []error
	rule *resolver.FieldOneofRule
}

func NewFieldOneofRuleBuilder() *FieldOneofRuleBuilder {
	return &FieldOneofRuleBuilder{
		rule: &resolver.FieldOneofRule{
			DefSet: &resolver.VariableDefinitionSet{},
		},
	}
}

func (b *FieldOneofRuleBuilder) SetIf(v string, out *resolver.Type) *FieldOneofRuleBuilder {
	b.rule.If = &resolver.CELValue{
		Expr: v,
		Out:  out,
	}
	return b
}

func (b *FieldOneofRuleBuilder) SetDefault(v bool) *FieldOneofRuleBuilder {
	b.rule.Default = true
	return b
}

func (b *FieldOneofRuleBuilder) AddVariableDefinition(def *resolver.VariableDefinition) *FieldOneofRuleBuilder {
	if def.Expr != nil && def.Expr.Map != nil && def.Expr.Map.Iterator != nil {
		name := def.Expr.Map.Iterator.Source.Name
		var found bool
		for _, varDef := range b.rule.DefSet.Definitions() {
			if varDef.Name == name {
				def.Expr.Map.Iterator.Source = varDef
				found = true
				break
			}
		}
		if !found {
			b.errs = append(b.errs, fmt.Errorf("%s variable name is not found", name))
		}
	}
	def.Idx = len(b.rule.DefSet.Definitions())
	b.rule.DefSet.Defs = append(b.rule.DefSet.Defs, def)
	return b
}

func (b *FieldOneofRuleBuilder) SetBy(expr string, out *resolver.Type) *FieldOneofRuleBuilder {
	b.rule.By = &resolver.CELValue{
		Expr: expr,
		Out:  out,
	}
	return b
}

func (b *FieldOneofRuleBuilder) SetDependencyGraph(graph *resolver.MessageDependencyGraph) *FieldOneofRuleBuilder {
	b.rule.DefSet.Graph = graph
	return b
}

func (b *FieldOneofRuleBuilder) AddVariableDefinitionGroup(group resolver.VariableDefinitionGroup) *FieldOneofRuleBuilder {
	b.rule.DefSet.Groups = append(b.rule.DefSet.Groups, group)
	return b
}

func (b *FieldOneofRuleBuilder) Build(t *testing.T) *resolver.FieldOneofRule {
	t.Helper()
	if len(b.errs) != 0 {
		t.Fatal(errors.Join(b.errs...))
	}
	return b.rule
}
