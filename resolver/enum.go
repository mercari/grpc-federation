package resolver

import (
	"strings"

	"github.com/mercari/grpc-federation/types"
)

func NewEnumType(enum *Enum, repeated bool) *Type {
	return &Type{
		Kind:     types.Enum,
		Enum:     enum,
		Repeated: repeated,
	}
}

func (e *Enum) HasValue(name string) bool {
	return e.Value(name) != nil
}

func (e *Enum) Value(name string) *EnumValue {
	if strings.Contains(name, ".") {
		enumFQDNPrefix := e.FQDN() + "."
		if !strings.HasPrefix(name, enumFQDNPrefix) {
			return nil
		}
		name = strings.TrimPrefix(name, enumFQDNPrefix)
	}
	for _, value := range e.Values {
		if value.Value == name {
			return value
		}
	}
	return nil
}

func (e *Enum) AttributeMap() map[string][]*EnumValue {
	attrMap := make(map[string][]*EnumValue)
	for _, value := range e.Values {
		if value.Rule == nil {
			continue
		}
		for _, attr := range value.Rule.Attrs {
			attrMap[attr.Name] = append(attrMap[attr.Name], value)
		}
	}
	return attrMap
}

func (e *Enum) Package() *Package {
	if e.File == nil {
		return nil
	}
	return e.File.Package
}

func (e *Enum) GoPackage() *GoPackage {
	if e.File == nil {
		return nil
	}
	return e.File.GoPackage
}

func (e *Enum) PackageName() string {
	pkg := e.Package()
	if pkg == nil {
		return ""
	}
	return pkg.Name
}

func (e *EnumExpr) ReferenceNames() []string {
	return e.By.ReferenceNames()
}
