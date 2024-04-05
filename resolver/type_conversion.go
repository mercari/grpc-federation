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
	if !requiredTypeConversion(fromType, toType) {
		return nil
	}
	decl := &TypeConversionDecl{From: fromType, To: toType}
	fqdn := decl.FQDN()
	if _, exists := convertedFQDNMap[fqdn]; exists {
		return nil
	}
	convertedFQDNMap[fqdn] = struct{}{}
	if fromType.Kind == types.Message && fromType.Message != nil && toType.Message != nil && fromType.Message.IsMapEntry {
		// map type
		fromMap := fromType.Message
		toMap := toType.Message
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
	case fromType.Kind == types.Message:
		for _, field := range toType.Message.Fields {
			fromField := fromType.Message.Field(field.Name)
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
	if fromType.Kind == types.Message {
		if fromType.Message.IsMapEntry {
			fromKey := fromType.Message.Field("key")
			fromValue := fromType.Message.Field("value")
			toKey := toType.Message.Field("key")
			toValue := toType.Message.Field("value")
			if fromKey == nil || fromValue == nil || toKey == nil || toValue == nil {
				return false
			}
			return requiredTypeConversion(fromKey.Type, toKey.Type) || requiredTypeConversion(fromValue.Type, toValue.Type)
		}
		return fromType.Message != toType.Message
	}
	if fromType.Kind == types.Enum {
		return fromType.Enum != toType.Enum
	}
	return fromType.Kind != toType.Kind
}
