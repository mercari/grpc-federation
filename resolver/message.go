package resolver

import (
	"strings"

	"github.com/mercari/grpc-federation/grpc/federation"
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/types"
)

func NewMessage(name string, fields []*Field) *Message {
	return &Message{Name: name, Fields: fields}
}

func NewMessageType(msg *Message, repeated bool) *Type {
	return &Type{
		Kind:     types.Message,
		Message:  msg,
		Repeated: repeated,
	}
}

func NewMapType(key, value *Type) *Type {
	return NewMessageType(&Message{
		IsMapEntry: true,
		Fields: []*Field{
			{Name: "key", Type: key},
			{Name: "value", Type: value},
		},
	}, false)
}

func NewMapTypeWithName(name string, key, value *Type) *Type {
	return NewMessageType(&Message{
		Name:       name,
		IsMapEntry: true,
		Fields: []*Field{
			{Name: "key", Type: key},
			{Name: "value", Type: value},
		},
	}, false)
}

func NewEnumSelectorType(trueType, falseType *Type) *Type {
	return NewMessageType(&Message{
		File: &File{
			Package: &Package{
				Name: "grpc.federation.private",
			},
			GoPackage: &GoPackage{
				Name:       "grpcfedcel",
				ImportPath: "github.com/mercari/grpc-federation/grpc/federation/cel",
				AliasName:  "grpcfedcel",
			},
		},
		Name: "EnumSelector",
		Fields: []*Field{
			{Name: "true", Type: trueType},
			{Name: "false", Type: falseType},
		},
	}, false)
}

func newMessageArgument(msg *Message) *Message {
	file := *msg.File
	file.Package = &Package{
		Name:  federation.PrivatePackageName,
		Files: Files{&file},
	}
	return &Message{
		File: &file,
		Name: strings.Join(append(msg.ParentMessageNames(), msg.Name+"Argument"), "_"),
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

func (m *Message) IsEnumSelector() bool {
	return m.FQDN() == grpcfedcel.EnumSelectorFQDN
}

func (m *Message) HasResolvers() bool {
	if m.Rule == nil {
		return false
	}
	if len(m.Rule.DefSet.DefinitionGroups()) != 0 {
		return true
	}
	if len(m.Rule.DefSet.Definitions()) != 0 {
		return true
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		if len(field.Rule.Oneof.DefSet.DefinitionGroups()) != 0 {
			return true
		}
	}
	return false
}

func (m *Message) VariableDefinitionGroups() []VariableDefinitionGroup {
	if m.Rule == nil {
		return nil
	}

	ret := m.Rule.DefSet.DefinitionGroups()
	for _, def := range m.Rule.DefSet.Definitions() {
		if def.Expr == nil {
			continue
		}
		switch {
		case def.Expr.Call != nil:
			for _, err := range def.Expr.Call.Errors {
				ret = append(ret, err.DefinitionGroups()...)
			}
		case def.Expr.Validation != nil:
			if def.Expr.Validation.Error != nil {
				ret = append(ret, def.Expr.Validation.Error.DefinitionGroups()...)
			}
		}
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		ret = append(ret, field.Rule.Oneof.DefSet.DefinitionGroups()...)
	}
	return ret
}

func (m *Message) HasCELValue() bool {
	if m.Rule == nil {
		return false
	}
	for _, varDef := range m.Rule.DefSet.Definitions() {
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
		case expr.Validation != nil:
			return true
		}
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		value := field.Rule.Value
		if value != nil && value.CEL != nil {
			return true
		}
		if field.Rule.Oneof != nil && field.Rule.Oneof.If != nil {
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
	for _, varDef := range m.Rule.DefSet.Definitions() {
		if varDef.Name == "" {
			continue
		}
		// Validation results won't be referenced
		if varDef.Expr.Validation != nil {
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

func (e *MessageExpr) HasContextCELLibrary() bool {
	for _, arg := range e.Args {
		if arg.Value.UseContextCELLibrary() {
			return true
		}
	}
	return false
}

func (m *Message) ReferenceNames() []string {
	if m.Rule == nil {
		return nil
	}

	rule := m.Rule
	refNames := rule.DefSet.ReferenceNames()
	for _, field := range m.Fields {
		if !field.HasRule() {
			continue
		}
		refNames = append(refNames, field.Rule.Value.ReferenceNames()...)
		if field.Rule.Oneof != nil {
			refNames = append(refNames, field.Rule.Oneof.If.ReferenceNames()...)
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

func (m *Message) AllMessages() []*Message {
	ret := []*Message{m}
	for _, msg := range m.NestedMessages {
		ret = append(ret, msg.AllMessages()...)
	}
	return ret
}

func (m *Message) AllEnums() []*Enum {
	enums := m.Enums
	for _, msg := range m.NestedMessages {
		enums = append(enums, msg.AllEnums()...)
	}
	return enums
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
	return DependencyGraphTreeFormat(m.Rule.DefSet.DefinitionGroups())
}

func (m *Message) TypeConversionDecls() []*TypeConversionDecl {
	convertedFQDNMap := make(map[string]struct{})
	var decls []*TypeConversionDecl
	if m.Rule != nil && len(m.Rule.DefSet.Definitions()) != 0 {
		for _, varDef := range m.Rule.DefSet.Definitions() {
			if varDef.Expr == nil {
				continue
			}
			switch {
			case varDef.Expr.Call != nil && varDef.Expr.Call.Request != nil:
				request := varDef.Expr.Call.Request
				for _, arg := range request.Args {
					if !request.Type.HasField(arg.Name) {
						continue
					}
					fromType := arg.Value.Type()
					field := request.Type.Field(arg.Name)
					toType := field.Type
					if toType.OneofField != nil {
						toType = field.Type.OneofField.Type.Clone()
						toType.OneofField = nil
					}
					decls = append(decls, typeConversionDecls(fromType, toType, convertedFQDNMap)...)
				}
			case varDef.Expr.Message != nil:
				// For numeric types, the specification allows accepting them even if the type of the message argument differs.
				// In such cases, since the actual message argument type may differ, additional type conversion is necessary.
				msgExpr := varDef.Expr.Message
				if msgExpr.Message != nil && msgExpr.Message.Rule != nil {
					msgArg := msgExpr.Message.Rule.MessageArgument
					for _, arg := range msgExpr.Args {
						switch {
						case arg.Name != "":
							msgArgField := msgArg.Field(arg.Name)
							decls = append(decls, typeConversionDecls(arg.Value.Type(), msgArgField.Type, convertedFQDNMap)...)
						case arg.Value != nil && arg.Value.Inline:
							for _, field := range arg.Value.CEL.Out.Message.Fields {
								msgArgField := msgArg.Field(field.Name)
								decls = append(decls, typeConversionDecls(field.Type, msgArgField.Type, convertedFQDNMap)...)
							}
						}
					}
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

func (m *Message) CustomResolvers() []*CustomResolver {
	var ret []*CustomResolver
	if m.HasCustomResolver() {
		ret = append(ret, &CustomResolver{Message: m})
	}
	for _, field := range m.Fields {
		if field.HasCustomResolver() {
			ret = append(ret, &CustomResolver{
				Message: m,
				Field:   field,
			})
		}
	}
	for _, group := range m.Rule.DefSet.DefinitionGroups() {
		for _, def := range group.VariableDefinitions() {
			ret = append(ret, m.customResolvers(def)...)
		}
	}
	return ret
}

func (m *Message) customResolvers(def *VariableDefinition) []*CustomResolver {
	var ret []*CustomResolver
	if def != nil {
		for _, expr := range def.MessageExprs() {
			if expr.Message == nil {
				continue
			}
			ret = append(ret, expr.Message.CustomResolvers()...)
		}
	}
	return ret
}

func (m *Message) GoPackageDependencies() []*GoPackage {
	pkgMap := map[*GoPackage]struct{}{}
	gopkg := m.GoPackage()
	pkgMap[gopkg] = struct{}{}
	for _, svc := range m.DependServices() {
		pkgMap[svc.GoPackage()] = struct{}{}
	}
	seenMsgMap := make(map[*Message]struct{})
	for _, field := range m.Fields {
		if field.Type.Message == nil {
			continue
		}
		for _, gopkg := range getGoPackageDependencies(gopkg, field.Type.Message, seenMsgMap) {
			pkgMap[gopkg] = struct{}{}
		}
	}
	pkgs := make([]*GoPackage, 0, len(pkgMap))
	for pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func getGoPackageDependencies(base *GoPackage, msg *Message, seenMsgMap map[*Message]struct{}) []*GoPackage {
	var ret []*GoPackage
	if base != msg.GoPackage() {
		ret = append(ret, msg.GoPackage())
	}
	seenMsgMap[msg] = struct{}{}
	for _, field := range msg.Fields {
		if field.Type.Message == nil {
			continue
		}
		if _, exists := seenMsgMap[field.Type.Message]; exists {
			continue
		}
		ret = append(ret, getGoPackageDependencies(base, field.Type.Message, seenMsgMap)...)
	}
	for _, m := range msg.NestedMessages {
		ret = append(ret, getGoPackageDependencies(base, m, seenMsgMap)...)
	}
	return ret
}

func (m *Message) DependServices() []*Service {
	if m == nil {
		return nil
	}

	return m.dependServices(make(map[*VariableDefinition]struct{}))
}

func (m *Message) dependServices(defMap map[*VariableDefinition]struct{}) []*Service {
	var svcs []*Service
	if m.Rule != nil {
		for _, group := range m.Rule.DefSet.DefinitionGroups() {
			for _, def := range group.VariableDefinitions() {
				svcs = append(svcs, dependServicesByDefinition(def, defMap)...)
			}
		}
		for _, def := range m.Rule.DefSet.Definitions() {
			if def.Expr == nil {
				continue
			}
			if def.Expr.Validation == nil || def.Expr.Validation.Error == nil {
				continue
			}
			for _, detail := range def.Expr.Validation.Error.Details {
				for _, group := range detail.DefSet.DefinitionGroups() {
					for _, def := range group.VariableDefinitions() {
						svcs = append(svcs, dependServicesByDefinition(def, defMap)...)
					}
				}
			}
		}
	}
	for _, field := range m.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for _, group := range field.Rule.Oneof.DefSet.DefinitionGroups() {
			for _, def := range group.VariableDefinitions() {
				svcs = append(svcs, dependServicesByDefinition(def, defMap)...)
			}
		}
	}
	return svcs
}

func dependServicesByDefinition(def *VariableDefinition, defMap map[*VariableDefinition]struct{}) []*Service {
	if _, found := defMap[def]; found {
		return nil
	}
	defMap[def] = struct{}{}

	if def == nil {
		return nil
	}

	expr := def.Expr
	if expr == nil {
		return nil
	}
	if expr.Call != nil {
		return []*Service{expr.Call.Method.Service}
	}
	var ret []*Service
	for _, msgExpr := range def.MessageExprs() {
		if msgExpr.Message != nil {
			ret = append(ret, msgExpr.Message.dependServices(defMap)...)
		}
	}
	return ret
}
