package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
		Type:     typ.Type,
		Ref:      typ.Ref,
		Enum:     typ.Enum,
		Repeated: true,
	}
}

func (b *FileBuilder) AddService(svc *resolver.Service) *FileBuilder {
	svc.File = b.file
	b.file.Services = append(b.file.Services, svc)
	return b
}

func (b *FileBuilder) AddEnum(enum *resolver.Enum) *FileBuilder {
	enum.File = b.file
	typ := &resolver.Type{Type: types.Enum, Enum: enum}
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
	typ := &resolver.Type{Type: types.Message, Ref: msg}
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
			Name:   name,
			Fields: []*resolver.Field{},
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

func (b *MessageBuilder) AddOneof(oneof *resolver.Oneof) *MessageBuilder {
	for idx, oneofField := range oneof.Fields {
		field := b.msg.Field(oneofField.Name)
		oneof.Fields[idx] = field
		field.Oneof = oneof
		field.Type.OneofField = &resolver.OneofField{Field: field}
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
		return &resolver.Type{Type: types.Message, Ref: b.msg}
	}
	for _, enum := range b.msg.Enums {
		if enum.Name == name {
			return &resolver.Type{Type: types.Enum, Enum: enum}
		}
	}
	for _, msg := range b.msg.NestedMessages {
		if msg.Name == name {
			return &resolver.Type{Type: types.Message, Ref: msg}
		}
	}
	t.Fatalf("failed to find %s type in %s message", name, b.msg.Name)
	return nil
}

func (b *MessageBuilder) AddField(name string, typ *resolver.Type) *MessageBuilder {
	b.msg.Fields = append(b.msg.Fields, &resolver.Field{Name: name, Type: typ})
	return b
}

func (b *MessageBuilder) AddFieldWithSelfType(name string, isRepeated bool) *MessageBuilder {
	typ := &resolver.Type{Type: types.Message, Ref: b.msg, Repeated: isRepeated}
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

func (b *MessageBuilder) AddFieldWithAlias(name string, typ *resolver.Type, field *resolver.Field) *MessageBuilder {
	b.msg.Fields = append(
		b.msg.Fields,
		&resolver.Field{
			Name: name,
			Type: typ,
			Rule: &resolver.FieldRule{
				Alias: field,
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
	b.msg.Fields = append(b.msg.Fields, &resolver.Field{Name: name, Type: typ, Rule: rule})
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
			Aliases: []*resolver.EnumValue{},
		},
		Enum: b.enum,
	})
	return b
}

func (b *EnumBuilder) AddValueWithAlias(value string, alias ...*resolver.EnumValue) *EnumBuilder {
	b.enum.Values = append(b.enum.Values, &resolver.EnumValue{
		Value: value,
		Rule: &resolver.EnumValueRule{
			Aliases: alias,
		},
		Enum: b.enum,
	})
	return b
}

func (b *EnumBuilder) WithAlias(alias *resolver.Enum) *EnumBuilder {
	b.enum.Rule = &resolver.EnumRule{
		Alias: alias,
	}
	return b
}

func (b *EnumBuilder) Build(t *testing.T) *resolver.Enum {
	t.Helper()
	return b.enum
}

type ServiceBuilder struct {
	svc *resolver.Service
}

func NewServiceBuilder(name string) *ServiceBuilder {
	return &ServiceBuilder{
		svc: &resolver.Service{Name: name},
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
}

func NewMessageRuleBuilder() *MessageRuleBuilder {
	return &MessageRuleBuilder{
		rule: &resolver.MessageRule{},
	}
}

func (b *MessageRuleBuilder) SetMethodCall(call *resolver.MethodCall) *MessageRuleBuilder {
	b.rule.MethodCall = call
	return b
}

func (b *MessageRuleBuilder) Method() *resolver.MethodCall {
	return b.rule.MethodCall
}

func (b *MessageRuleBuilder) AddMessageDependency(name string, msg *resolver.Message, args []*resolver.Argument, autobind, used bool) *MessageRuleBuilder {
	dep := &resolver.MessageDependency{
		Name:     name,
		Message:  msg,
		Args:     args,
		AutoBind: autobind,
		Used:     used,
	}
	b.rule.MessageDependencies = append(b.rule.MessageDependencies, dep)
	return b
}

func (b *MessageRuleBuilder) SetCustomResolver(v bool) *MessageRuleBuilder {
	b.rule.CustomResolver = v
	return b
}

func (b *MessageRuleBuilder) SetAlias(alias *resolver.Message) *MessageRuleBuilder {
	b.rule.Alias = alias
	return b
}

func (b *MessageRuleBuilder) SetMessageArgument(msg *resolver.Message) *MessageRuleBuilder {
	b.rule.MessageArgument = msg
	return b
}

func (b *MessageRuleBuilder) SetDependencyGraph(graph *resolver.MessageRuleDependencyGraph) *MessageRuleBuilder {
	b.rule.DependencyGraph = graph
	return b
}

func (b *MessageRuleBuilder) AddResolver(group resolver.MessageResolverGroup) *MessageRuleBuilder {
	b.rule.Resolvers = append(b.rule.Resolvers, group)
	return b
}

func (b *MessageRuleBuilder) Build(t *testing.T) *resolver.MessageRule {
	t.Helper()
	return b.rule
}

func NewMessageResolver(name string) *resolver.MessageResolver {
	return &resolver.MessageResolver{Name: name}
}

type MessageResolverGroupBuilder struct {
	starts []resolver.MessageResolverGroup
	end    *resolver.MessageResolver
}

func NewMessageResolverGroupByName(name string) *resolver.SequentialMessageResolverGroup {
	return &resolver.SequentialMessageResolverGroup{
		End: &resolver.MessageResolver{Name: name},
	}
}

func NewMessageResolverGroupBuilder() *MessageResolverGroupBuilder {
	return &MessageResolverGroupBuilder{}
}

func (b *MessageResolverGroupBuilder) AddStart(start resolver.MessageResolverGroup) *MessageResolverGroupBuilder {
	b.starts = append(b.starts, start)
	return b
}

func (b *MessageResolverGroupBuilder) SetEnd(end *resolver.MessageResolver) *MessageResolverGroupBuilder {
	b.end = end
	return b
}

func (b *MessageResolverGroupBuilder) Build(t *testing.T) resolver.MessageResolverGroup {
	t.Helper()
	if len(b.starts) > 1 {
		return &resolver.ConcurrentMessageResolverGroup{
			Starts: b.starts,
			End:    b.end,
		}
	}
	var start resolver.MessageResolverGroup
	if len(b.starts) == 1 {
		start = b.starts[0]
	}
	return &resolver.SequentialMessageResolverGroup{
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

func (b *MethodCallBuilder) SetResponse(res *resolver.Response) *MethodCallBuilder {
	res.Type = b.call.Method.Response
	b.call.Response = res
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

func (b *RequestBuilder) Build(t *testing.T) *resolver.Request {
	t.Helper()
	return b.req
}

type ResponseBuilder struct {
	res *resolver.Response
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		res: &resolver.Response{},
	}
}

func (b *ResponseBuilder) AddField(name, field string, typ *resolver.Type, autobind, used bool) *ResponseBuilder {
	b.res.Fields = append(b.res.Fields, &resolver.ResponseField{
		Name:      name,
		FieldName: field,
		Type:      typ,
		AutoBind:  autobind,
		Used:      used,
	})
	return b
}

func (b *ResponseBuilder) Build(t *testing.T) *resolver.Response {
	t.Helper()
	return b.res
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
	value.Inline = true
	b.args = append(b.args, &resolver.Argument{Value: value})
	return b
}

func (b *MessageDependencyArgumentBuilder) Build(t *testing.T) []*resolver.Argument {
	t.Helper()
	return b.args
}

type MessageArgumentValueBuilder struct {
	pathBuilder *resolver.PathBuilder
	value       *resolver.Value
}

func NewMessageArgumentValueBuilder(ref, filtered *resolver.Type, path string) *MessageArgumentValueBuilder {
	return &MessageArgumentValueBuilder{
		pathBuilder: resolver.NewPathBuilder(path),
		value: &resolver.Value{
			PathType: resolver.MessageArgumentPathType,
			Ref:      ref,
			Filtered: filtered,
		},
	}
}

func (b *MessageArgumentValueBuilder) Build(t *testing.T) *resolver.Value {
	t.Helper()
	p, err := b.pathBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	b.value.Path = &resolver.PathRoot{PathBase: &resolver.PathBase{Child: p}}
	return b.value
}

type NameReferenceValueBuilder struct {
	pathBuilder *resolver.PathBuilder
	value       *resolver.Value
}

func NewNameReferenceValueBuilder(ref, filtered *resolver.Type, path string) *NameReferenceValueBuilder {
	return &NameReferenceValueBuilder{
		pathBuilder: resolver.NewPathBuilder(path),
		value: &resolver.Value{
			PathType: resolver.NameReferencePathType,
			Ref:      ref,
			Filtered: filtered,
		},
	}
}

func (b *NameReferenceValueBuilder) Build(t *testing.T) *resolver.Value {
	t.Helper()
	p, err := b.pathBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	b.value.Path = p
	return b.value
}

type FieldRuleBuilder struct {
	rule *resolver.FieldRule
}

func NewFieldRuleBuilder(value *resolver.Value) *FieldRuleBuilder {
	return &FieldRuleBuilder{rule: &resolver.FieldRule{Value: value}}
}

func (b *FieldRuleBuilder) SetCustomResolver(v bool) *FieldRuleBuilder {
	b.rule.CustomResolver = v
	return b
}

func (b *FieldRuleBuilder) SetMessageCustomResolver(v bool) *FieldRuleBuilder {
	b.rule.MessageCustomResolver = v
	return b
}

func (b *FieldRuleBuilder) SetAlias(v *resolver.Field) *FieldRuleBuilder {
	b.rule.Alias = v
	return b
}

func (b *FieldRuleBuilder) Build(t *testing.T) *resolver.FieldRule {
	t.Helper()
	return b.rule
}

type DependencyGraphBuilder struct {
	graph *resolver.MessageRuleDependencyGraph
}

func NewDependencyGraphBuilder() *DependencyGraphBuilder {
	return &DependencyGraphBuilder{
		graph: &resolver.MessageRuleDependencyGraph{},
	}
}

func (b *DependencyGraphBuilder) Add(parent *resolver.Message, children ...*resolver.Message) *DependencyGraphBuilder {
	nodes := make([]*resolver.MessageRuleDependencyGraphNode, 0, len(children))
	for _, child := range children {
		nodes = append(nodes, &resolver.MessageRuleDependencyGraphNode{Message: child})
	}
	b.graph.Roots = append(b.graph.Roots, &resolver.MessageRuleDependencyGraphNode{
		Message:  parent,
		Children: nodes,
	})
	return b
}

func (b *DependencyGraphBuilder) Build(t *testing.T) *resolver.MessageRuleDependencyGraph {
	t.Helper()
	return b.graph
}

type ServiceRuleBuilder struct {
	rule *resolver.ServiceRule
}

func NewServiceRuleBuilder() *ServiceRuleBuilder {
	return &ServiceRuleBuilder{
		rule: &resolver.ServiceRule{
			Dependencies: []*resolver.ServiceDependency{},
		},
	}
}

func (b *ServiceRuleBuilder) AddDependency(name string, svc *resolver.Service) *ServiceRuleBuilder {
	b.rule.Dependencies = append(b.rule.Dependencies, &resolver.ServiceDependency{
		Name:    name,
		Service: svc,
	})
	return b
}

func (b *ServiceRuleBuilder) Build(t *testing.T) *resolver.ServiceRule {
	t.Helper()
	return b.rule
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
