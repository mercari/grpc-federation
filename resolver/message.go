package resolver

import (
	"strings"

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
	if m.HasCustomResolver() {
		return true
	}
	if len(m.Rule.Resolvers) != 0 {
		return true
	}
	if m.HasRuleEveryFields() {
		return true
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
	var decls []*TypeConversionDecl
	if m.Rule != nil && m.Rule.MethodCall != nil && m.Rule.MethodCall.Response != nil {
		request := m.Rule.MethodCall.Request
		if request != nil {
			for _, arg := range request.Args {
				if !request.Type.HasField(arg.Name) {
					continue
				}
				fromType := arg.Value.Filtered
				toType := request.Type.Field(arg.Name).Type
				decls = append(decls, typeConversionDecls(fromType, toType)...)
			}
		}
	}
	for _, field := range m.Fields {
		decls = append(decls, field.TypeConversionDecls()...)
	}
	uniqueDecls := uniqueTypeConversionDecls(decls)
	return sortTypeConversionDecls(uniqueDecls)
}
