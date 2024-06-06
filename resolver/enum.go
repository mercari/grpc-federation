package resolver

import "strings"

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
