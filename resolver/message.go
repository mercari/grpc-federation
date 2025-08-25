package resolver

import (
	"strings"

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
		Name:  file.PrivatePackageName(),
		Files: Files{&file},
	}
	return &Message{
		File: &file,
		Name: strings.Join(append(msg.ParentMessageNames(), msg.Name+"Argument"), "_"),
	}
}

func (m *Message) ParentMessageNames() []string {
	if m == nil {
		return nil
	}
	if m.ParentMessage == nil {
		return []string{}
	}
	return append(m.ParentMessage.ParentMessageNames(), m.ParentMessage.Name)
}

func (m *Message) Package() *Package {
	if m == nil {
		return nil
	}
	if m.File == nil {
		return nil
	}
	return m.File.Package
}

func (m *Message) HasRule() bool {
	if m == nil {
		return false
	}
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
	if m == nil {
		return false
	}
	return m.FQDN() == grpcfedcel.EnumSelectorFQDN
}

func (m *Message) HasResolvers() bool {
	if m == nil {
		return false
	}
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
	if m == nil {
		return nil
	}
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

func (m *Message) AllVariableDefinitions() VariableDefinitions {
	if m == nil {
		return nil
	}

	var defs VariableDefinitions
	for _, group := range m.VariableDefinitionGroups() {
		defs = append(defs, group.VariableDefinitions()...)
	}
	return defs
}

func (m *Message) HasCELValue() bool {
	if m == nil {
		return false
	}
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
	if m == nil {
		return false
	}

	return m.Rule != nil && m.Rule.CustomResolver
}

func (m *Message) HasRuleEveryFields() bool {
	if m == nil {
		return false
	}

	for _, field := range m.Fields {
		if !field.HasRule() {
			return false
		}
	}
	return true
}

func (m *Message) HasCustomResolverFields() bool {
	if m == nil {
		return false
	}

	return len(m.CustomResolverFields()) != 0
}

func (m *Message) UseAllNameReference() {
	if m == nil {
		return
	}

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
	if e == nil {
		return nil
	}

	var refNames []string
	for _, arg := range e.Args {
		refNames = append(refNames, arg.Value.ReferenceNames()...)
	}
	return refNames
}

func (m *Message) ReferenceNames() []string {
	if m == nil {
		return nil
	}

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
	if m == nil {
		return nil
	}

	fields := make([]*Field, 0, len(m.Fields))
	for _, field := range m.Fields {
		if field.HasCustomResolver() {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Message) GoPackage() *GoPackage {
	if m == nil {
		return nil
	}
	if m.File == nil {
		return nil
	}
	return m.File.GoPackage
}

func (m *Message) PackageName() string {
	if m == nil {
		return ""
	}
	pkg := m.Package()
	if pkg == nil {
		return ""
	}
	return pkg.Name
}

func (m *Message) FileName() string {
	if m == nil {
		return ""
	}
	if m.File == nil {
		return ""
	}
	return m.File.Name
}

func (m *Message) HasField(name string) bool {
	if m == nil {
		return false
	}

	return m.Field(name) != nil
}

func (m *Message) Field(name string) *Field {
	if m == nil {
		return nil
	}

	for _, field := range m.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (m *Message) Oneof(name string) *Oneof {
	if m == nil {
		return nil
	}

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
	if m == nil {
		return nil
	}

	ret := []*Message{m}
	for _, msg := range m.NestedMessages {
		ret = append(ret, msg.AllMessages()...)
	}
	return ret
}

func (m *Message) AllEnums() []*Enum {
	if m == nil {
		return nil
	}

	enums := m.Enums
	for _, msg := range m.NestedMessages {
		enums = append(enums, msg.AllEnums()...)
	}
	return enums
}

func (m *Message) HasFieldRule() bool {
	if m == nil {
		return false
	}

	for _, field := range m.Fields {
		if field.HasRule() {
			return true
		}
	}
	return false
}

func (m *Message) DependencyGraphTreeFormat() string {
	if m == nil {
		return ""
	}
	if m.Rule == nil {
		return ""
	}
	return DependencyGraphTreeFormat(m.Rule.DefSet.DefinitionGroups())
}

func (m *Message) TypeConversionDecls() []*TypeConversionDecl {
	if m == nil {
		return nil
	}

	convertedFQDNMap := make(map[string]struct{})
	var decls []*TypeConversionDecl
	for _, def := range m.AllVariableDefinitions() {
		if def.Expr == nil {
			continue
		}
		switch {
		case def.Expr.Call != nil:
			callExpr := def.Expr.Call
			if callExpr.Request != nil {
				request := callExpr.Request
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
			}
		case def.Expr.Message != nil:
			// For numeric types, the specification allows accepting them even if the type of the message argument differs.
			// In such cases, since the actual message argument type may differ, additional type conversion is necessary.
			msgExpr := def.Expr.Message
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
		case def.Expr.Enum != nil:
			enumExpr := def.Expr.Enum
			decls = append(decls, typeConversionDecls(enumExpr.By.Out, def.Expr.Type, convertedFQDNMap)...)
		case def.Expr.Map != nil:
			mapExpr := def.Expr.Map.Expr
			switch {
			case mapExpr.Message != nil:
				msgExpr := mapExpr.Message
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
			case mapExpr.Enum != nil:
				from := mapExpr.Enum.By.Out.Clone()
				from.Repeated = true
				decls = append(decls, typeConversionDecls(from, def.Expr.Type, convertedFQDNMap)...)
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
	if m == nil {
		return nil
	}

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
	if m.Rule != nil {
		for _, group := range m.Rule.DefSet.DefinitionGroups() {
			for _, def := range group.VariableDefinitions() {
				ret = append(ret, m.customResolvers(def)...)
			}
		}
	}
	return ret
}

func (m *Message) customResolvers(def *VariableDefinition) []*CustomResolver {
	if m == nil {
		return nil
	}

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
	if m == nil {
		return nil
	}

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
	if m == nil {
		return nil
	}

	var svcs []*Service
	for _, def := range m.AllVariableDefinitions() {
		svcs = append(svcs, dependServicesByDefinition(def, defMap)...)
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
