package resolver

import "sort"

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

func (r *EnumRule) AliasValues() []*EnumValue {
	if len(r.Aliases) == 0 {
		return nil
	}
	if len(r.Aliases) == 1 {
		return r.Aliases[0].Values
	}

	type ValueWithCount struct {
		value *EnumValue
		count int
	}

	valueMap := make(map[string]*ValueWithCount)
	for _, alias := range r.Aliases {
		for _, value := range alias.Values {
			valueWithCount := valueMap[value.Value]
			if valueWithCount == nil {
				valueWithCount = &ValueWithCount{value: value}
				valueMap[value.Value] = valueWithCount
			}
			valueWithCount.count++
		}
	}

	var ret []*EnumValue
	for _, valueWithCount := range valueMap {
		if valueWithCount.count == len(r.Aliases) {
			ret = append(ret, valueWithCount.value)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Value < ret[j].Value
	})
	return ret
}

func (r *EnumRule) AliasValue(valueName string) *EnumValue {
	for _, value := range r.AliasValues() {
		if value.Value == valueName {
			return value
		}
	}
	return nil
}
