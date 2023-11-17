package resolver

import (
	"fmt"
	"strings"

	"github.com/mercari/grpc-federation/grpc/federation"
	"github.com/mercari/grpc-federation/types"
)

func NewMessage(name string, fields []*Field) *Message {
	return &Message{Name: name, Fields: fields}
}

func NewMessageType(msg *Message, repeated bool) *Type {
	return &Type{
		Type:     types.Message,
		Ref:      msg,
		Repeated: repeated,
	}
}

func newMessageArgument(msg *Message) *Message {
	file := *msg.File
	file.Package = &Package{
		Name:  federation.PrivatePackageName,
		Files: Files{&file},
	}
	return &Message{
		File:   &file,
		Name:   fmt.Sprintf("%sArgument", msg.Name),
		Fields: []*Field{},
	}
}

func (m *Message) ParentMessageNames() []string {
	if m.ParentMessage == nil {
		return []string{}
	}
	return append(m.ParentMessage.ParentMessageNames(), m.ParentMessage.Name)
}

func (m *Message) Package() *Package {
	if m.File == nil {
		return nil
	}
	return m.File.Package
}

func (m *Message) FQDN() string {
	return strings.Join(
		append(append([]string{m.PackageName()}, m.ParentMessageNames()...), m.Name),
		".",
	)
}

func (m *Message) HasRule() bool {
	if m.Rule == nil {
		return false
	}
	if m.HasResolvers() {
		return true
	}
	if m.HasCustomResolver() {
		return true
	}
	if m.HasRuleEveryFields() {
		return true
	}
	return false
}

func (m *Message) HasResolvers() bool {
	if m.Rule == nil {
		return false
	}
	if len(m.Rule.Resolvers) != 0 {
		return true
	}
	if len(m.Rule.VariableDefinitions) != 0 {
		return true
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		if len(field.Rule.Oneof.Resolvers) != 0 {
			return true
		}
	}
	return false
}

func (m *Message) MessageResolvers() []MessageResolverGroup {
	if m.Rule == nil {
		return nil
	}

	ret := m.Rule.Resolvers
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		ret = append(ret, field.Rule.Oneof.Resolvers...)
	}
	return ret
}

func (m *Message) HasCELValue() bool {
	if m.Rule == nil {
		return false
	}
	methodCall := m.Rule.MethodCall
	if methodCall != nil {
		if methodCall.Request != nil {
			for _, arg := range methodCall.Request.Args {
				if arg.Value != nil && arg.Value.CEL != nil {
					return true
				}
			}
		}
	}
	for _, dep := range m.Rule.MessageDependencies {
		for _, arg := range dep.Args {
			if arg.Value != nil && arg.Value.CEL != nil {
				return true
			}
		}
	}
	for _, varDef := range m.Rule.VariableDefinitions {
		if varDef.Expr == nil {
			continue
		}
		expr := varDef.Expr
		switch {
		case expr.By != nil:
			return true
		case expr.Call != nil:
			if expr.Call.Request != nil {
				for _, arg := range expr.Call.Request.Args {
					if arg.Value != nil && arg.Value.CEL != nil {
						return true
					}
				}
			}
		case expr.Message != nil:
			for _, arg := range expr.Message.Args {
				if arg.Value != nil && arg.Value.CEL != nil {
					return true
				}
			}
		}
	}
	if len(m.Rule.Validations) != 0 {
		return true
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		value := field.Rule.Value
		if value != nil && value.CEL != nil {
			return true
		}
		if field.Rule.Oneof != nil && field.Rule.Oneof.Expr != nil {
			return true
		}
		if field.Rule.Oneof != nil && field.Rule.Oneof.By != nil {
			return true
		}
	}
	return false
}

func (m *Message) HasCustomResolver() bool {
	return m.Rule != nil && m.Rule.CustomResolver
}

func (m *Message) HasRuleEveryFields() bool {
	for _, field := range m.Fields {
		if !field.HasRule() {
			return false
		}
	}
	return true
}

func (m *Message) HasCustomResolverFields() bool {
	return len(m.CustomResolverFields()) != 0
}

func (m *Message) UseAllNameReference() {
	if m.Rule == nil {
		return
	}
	if m.Rule.MethodCall != nil && m.Rule.MethodCall.Response != nil {
		for _, field := range m.Rule.MethodCall.Response.Fields {
			field.Used = true
		}
	}
	for _, msg := range m.Rule.MessageDependencies {
		if msg.Name == "" {
			continue
		}
		msg.Used = true
	}
	for _, varDef := range m.Rule.VariableDefinitions {
		if varDef.Name == "" {
			continue
		}
		varDef.Used = true
	}
}

func (e *MessageExpr) ReferenceNames() []string {
	var refNames []string
	for _, arg := range e.Args {
		refNames = append(refNames, arg.Value.ReferenceNames()...)
	}
	return refNames
}

func (m *Message) ReferenceNames() []string {
	if m.Rule == nil {
		return nil
	}

	rule := m.Rule
	var refNames []string
	methodCall := rule.MethodCall
	if methodCall != nil && methodCall.Request != nil {
		for _, arg := range methodCall.Request.Args {
			refNames = append(refNames, arg.Value.ReferenceNames()...)
		}
	}
	for _, depMessage := range rule.MessageDependencies {
		for _, arg := range depMessage.Args {
			refNames = append(refNames, arg.Value.ReferenceNames()...)
		}
	}
	for _, varDef := range rule.VariableDefinitions {
		refNames = append(refNames, varDef.ReferenceNames()...)
	}
	for _, field := range m.Fields {
		if !field.HasRule() {
			continue
		}
		refNames = append(refNames, field.Rule.Value.ReferenceNames()...)
		if field.Rule.Oneof != nil {
			refNames = append(refNames, field.Rule.Oneof.Expr.ReferenceNames()...)
		}
	}
	return refNames
}

func (m *Message) CustomResolverFields() []*Field {
	fields := make([]*Field, 0, len(m.Fields))
	for _, field := range m.Fields {
		if field.HasCustomResolver() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Message) GoPackage() *GoPackage {
	if m.File == nil {
		return nil
	}
	return m.File.GoPackage
}

func (m *Message) PackageName() string {
	pkg := m.Package()
	if pkg == nil {
		return ""
	}
	return pkg.Name
}

func (m *Message) FileName() string {
	if m.File == nil {
		return ""
	}
	return m.File.Name
}

func (m *Message) HasField(name string) bool {
	return m.Field(name) != nil
}

func (m *Message) Field(name string) *Field {
	for _, field := range m.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (m *Message) Oneof(name string) *Oneof {
	for _, field := range m.Fields {
		if field.Oneof == nil {
			continue
		}
		if field.Oneof.Name == name {
			return field.Oneof
		}
	}
	return nil
}

func (m *Message) HasFieldRule() bool {
	for _, field := range m.Fields {
		if field.HasRule() {
			return true
		}
	}
	return false
}

func (m *Message) DependencyGraphTreeFormat() string {
	if m.Rule == nil {
		return ""
	}
	return DependencyGraphTreeFormat(m.Rule.Resolvers)
}

func (m *Message) TypeConversionDecls() []*TypeConversionDecl {
	convertedFQDNMap := make(map[string]struct{})
	var decls []*TypeConversionDecl
	if m.Rule != nil && m.Rule.MethodCall != nil && m.Rule.MethodCall.Response != nil {
		request := m.Rule.MethodCall.Request
		if request != nil {
			for _, arg := range request.Args {
				if !request.Type.HasField(arg.Name) {
					continue
				}
				fromType := arg.Value.Type()
				field := request.Type.Field(arg.Name)
				toType := field.Type
				decls = append(decls, typeConversionDecls(fromType, toType, convertedFQDNMap)...)
			}
		}
	}
	if m.Rule != nil && len(m.Rule.VariableDefinitions) != 0 {
		for _, varDef := range m.Rule.VariableDefinitions {
			if varDef.Expr == nil {
				continue
			}
			if varDef.Expr.Call != nil && varDef.Expr.Call.Request != nil {
				request := varDef.Expr.Call.Request
				for _, arg := range request.Args {
					if !request.Type.HasField(arg.Name) {
						continue
					}
					fromType := arg.Value.Type()
					field := request.Type.Field(arg.Name)
					toType := field.Type
					decls = append(decls, typeConversionDecls(fromType, toType, convertedFQDNMap)...)
				}
			}
		}
	}

	for _, field := range m.Fields {
		decls = append(decls, field.typeConversionDecls(convertedFQDNMap)...)
	}
	uniqueDecls := uniqueTypeConversionDecls(decls)
	return sortTypeConversionDecls(uniqueDecls)
}
