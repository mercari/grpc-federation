package resolver

import "fmt"

func (e *Enum) FQDN() string {
	if e.Message != nil {
		return fmt.Sprintf("%s.%s", e.Message.FQDN(), e.Name)
	}
	return fmt.Sprintf("%s.%s", e.PackageName(), e.Name)
}

func (e *Enum) HasValue(name string) bool {
	return e.Value(name) != nil
}

func (e *Enum) Value(name string) *EnumValue {
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
