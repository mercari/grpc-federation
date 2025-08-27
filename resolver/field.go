package resolver

func (f *Field) HasRule() bool {
	if f == nil {
		return false
	}
	return f.Rule != nil
}

func (f *Field) HasCustomResolver() bool {
	if f == nil {
		return false
	}
	return f.HasRule() && f.Rule.CustomResolver
}

func (f *Field) HasMessageCustomResolver() bool {
	if f == nil {
		return false
	}
	return f.HasRule() && f.Rule.MessageCustomResolver
}

func (f *Field) TypeConversionDecls() []*TypeConversionDecl {
	if f == nil {
		return nil
	}
	return f.typeConversionDecls(make(map[string]struct{}))
}

func (f *Field) typeConversionDecls(convertedFQDNMap map[string]struct{}) []*TypeConversionDecl {
	if f == nil {
		return nil
	}
	if !f.RequiredTypeConversion() {
		return nil
	}
	toType := f.Type
	var decls []*TypeConversionDecl
	for _, fromType := range f.SourceTypes() {
		decls = append(decls, typeConversionDecls(fromType, toType, convertedFQDNMap)...)
	}
	return uniqueTypeConversionDecls(decls)
}

func (f *Field) RequiredTypeConversion() bool {
	if f == nil {
		return false
	}
	if !f.HasRule() {
		return false
	}
	if f.HasCustomResolver() {
		return false
	}
	for _, fromType := range f.SourceTypes() {
		if requiredTypeConversion(fromType, f.Type) {
			return true
		}
	}
	return false
}

func (f *Field) SourceTypes() []*Type {
	if f == nil {
		return nil
	}
	if !f.HasRule() {
		return nil
	}
	rule := f.Rule
	var ret []*Type
	if rule.Value != nil {
		ret = append(ret, rule.Value.Type())
	}
	if len(rule.Aliases) != 0 {
		for _, alias := range rule.Aliases {
			ret = append(ret, alias.Type)
		}
	}
	if rule.AutoBindField != nil {
		ret = append(ret, rule.AutoBindField.Field.Type)
	}
	if rule.Oneof != nil {
		if rule.Oneof.By != nil {
			ret = append(ret, rule.Oneof.By.Out)
		}
	}
	return ret
}
