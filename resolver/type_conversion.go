package resolver

import (
	"sort"

	"github.com/mercari/grpc-federation/types"
)

func (decl *TypeConversionDecl) FQDN() string {
	return decl.From.FQDN() + decl.To.FQDN()
}

func typeConversionDecls(fromType, toType *Type, convertedFQDNMap map[string]struct{}) []*TypeConversionDecl {
	if fromType == nil || toType == nil {
		return nil
	}
	if fromType.Type != toType.Type {
		return nil
	}
	if !requiredTypeConversion(fromType, toType) {
		return nil
	}
	decl := &TypeConversionDecl{From: fromType, To: toType}
	fqdn := decl.FQDN()
	if _, exists := convertedFQDNMap[fqdn]; exists {
		return nil
	}
	convertedFQDNMap[fqdn] = struct{}{}
	if fromType.Type == types.Message && fromType.Ref != nil && toType.Ref != nil && fromType.Ref.IsMapEntry {
		// map type
		fromMap := fromType.Ref
		toMap := toType.Ref
		fromKey := fromMap.Field("key")
		toKey := toMap.Field("key")
		fromValue := fromMap.Field("value")
		toValue := toMap.Field("value")
		var decls []*TypeConversionDecl
		if fromKey != nil && toKey != nil {
			decls = append(decls, typeConversionDecls(fromKey.Type, toKey.Type, convertedFQDNMap)...)
		}
		if fromValue != nil && toValue != nil {
			decls = append(decls, typeConversionDecls(fromValue.Type, toValue.Type, convertedFQDNMap)...)
		}
		return decls
	}
	decls := []*TypeConversionDecl{{From: fromType, To: toType}}
	switch {
	case fromType.Repeated:
		ft := fromType.Clone()
		tt := toType.Clone()
		ft.Repeated = false
		tt.Repeated = false
		decls = append(decls, typeConversionDecls(ft, tt, convertedFQDNMap)...)
	case fromType.OneofField != nil:
		ft := fromType.Clone()
		tt := toType.Clone()
		ft.OneofField = nil
		tt.OneofField = nil
		decls = append(decls, typeConversionDecls(ft, tt, convertedFQDNMap)...)
	case fromType.Type == types.Message:
		for _, field := range toType.Ref.Fields {
			fromField := fromType.Ref.Field(field.Name)
			if fromField == nil {
				continue
			}
			decls = append(decls, typeConversionDecls(fromField.Type, field.Type, convertedFQDNMap)...)
		}
	}
	return decls
}

func uniqueTypeConversionDecls(decls []*TypeConversionDecl) []*TypeConversionDecl {
	declMap := make(map[string]*TypeConversionDecl)
	for _, decl := range decls {
		declMap[decl.FQDN()] = decl
	}
	ret := make([]*TypeConversionDecl, 0, len(declMap))
	for _, decl := range declMap {
		ret = append(ret, decl)
	}
	return ret
}

func sortTypeConversionDecls(decls []*TypeConversionDecl) []*TypeConversionDecl {
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FQDN() < decls[j].FQDN()
	})
	return decls
}

func requiredTypeConversion(fromType, toType *Type) bool {
	if fromType == nil || toType == nil {
		return false
	}
	if fromType.Type == types.Message {
		if fromType.Ref.IsMapEntry {
			fromKey := fromType.Ref.Field("key")
			fromValue := fromType.Ref.Field("value")
			toKey := toType.Ref.Field("key")
			toValue := toType.Ref.Field("value")
			if fromKey == nil || fromValue == nil || toKey == nil || toValue == nil {
				return false
			}
			return requiredTypeConversion(fromKey.Type, toKey.Type) || requiredTypeConversion(fromValue.Type, toValue.Type)
		}
		return fromType.Ref != toType.Ref
	}
	if fromType.Type == types.Enum {
		return fromType.Enum != toType.Enum
	}
	return fromType.Type != toType.Type
}
