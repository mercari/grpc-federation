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
	if !f.RequiredTypeConversion(FieldConversionKindAll) {
		return nil
	}
	toType := f.Type
	var decls []*TypeConversionDecl
	for _, fromType := range f.SourceTypes(FieldConversionKindAll) {
		decls = append(decls, typeConversionDecls(fromType, toType, convertedFQDNMap)...)
	}
	return uniqueTypeConversionDecls(decls)
}

type FieldConversionKind int

const (
	FieldConversionKindUnknown FieldConversionKind = 1 << iota
	FieldConversionKindValue
	FieldConversionKindAutoBind
	FieldConversionKindAlias
	FieldConversionKindOneof
)

var FieldConversionKindAll = FieldConversionKindValue | FieldConversionKindAutoBind | FieldConversionKindAlias | FieldConversionKindOneof

func (f *Field) RequiredTypeConversion(kind FieldConversionKind) bool {
	if f == nil {
		return false
	}
	if !f.HasRule() {
		return false
	}
	if f.HasCustomResolver() {
		return false
	}
	for _, fromType := range f.SourceTypes(kind) {
		if requiredTypeConversion(fromType, f.Type) {
			return true
		}
	}
	return false
}

func (f *Field) SourceTypes(kind FieldConversionKind) []*Type {
	if f == nil {
		return nil
	}
	if !f.HasRule() {
		return nil
	}
	rule := f.Rule
	var ret []*Type
	if kind&FieldConversionKindValue != 0 && rule.Value != nil {
		ret = append(ret, rule.Value.Type())
	}
	if kind&FieldConversionKindAlias != 0 && len(rule.Aliases) != 0 {
		for _, alias := range rule.Aliases {
			ret = append(ret, alias.Type)
		}
	}
	if kind&FieldConversionKindAutoBind != 0 && rule.AutoBindField != nil {
		ret = append(ret, rule.AutoBindField.Field.Type)
	}
	if kind&FieldConversionKindOneof != 0 && rule.Oneof != nil {
		if rule.Oneof.By != nil {
			ret = append(ret, rule.Oneof.By.Out)
		}
	}
	return ret
}
