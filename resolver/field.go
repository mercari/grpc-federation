package resolver

func (f *Field) HasRule() bool {
	return f.Rule != nil
}

func (f *Field) HasCustomResolver() bool {
	return f.HasRule() && f.Rule.CustomResolver
}

func (f *Field) HasMessageCustomResolver() bool {
	return f.HasRule() && f.Rule.MessageCustomResolver
}

func (f *Field) TypeConversionDecls() []*TypeConversionDecl {
	return f.typeConversionDecls(make(map[string]struct{}))
}

func (f *Field) typeConversionDecls(convertedFQDNMap map[string]struct{}) []*TypeConversionDecl {
	if !f.RequiredTypeConversion() {
		return nil
	}
	fromType := f.SourceType()
	toType := f.Type
	return uniqueTypeConversionDecls(typeConversionDecls(fromType, toType, convertedFQDNMap))
}

func (f *Field) RequiredTypeConversion() bool {
	if !f.HasRule() {
		return false
	}
	if f.HasCustomResolver() {
		return false
	}
	return requiredTypeConversion(f.SourceType(), f.Type)
}

func (f *Field) SourceType() *Type {
	if !f.HasRule() {
		return nil
	}
	rule := f.Rule
	switch {
	case rule.Value != nil:
		return rule.Value.Type()
	case rule.Alias != nil:
		return rule.Alias.Type
	case rule.AutoBindField != nil:
		return rule.AutoBindField.Field.Type
	}
	return nil
}
