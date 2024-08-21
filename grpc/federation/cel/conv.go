package cel

import "github.com/google/cel-go/common/ast"

func ToIdentifiers(expr ast.Expr) []string {
	var idents []string
	switch expr.Kind() {
	case ast.CallKind:
		call := expr.AsCall()
		idents = append(idents, ToIdentifiers(call.Target())...)
		for _, arg := range call.Args() {
			idents = append(idents, ToIdentifiers(arg)...)
		}
	case ast.IdentKind:
		idents = append(idents, expr.AsIdent())
	case ast.ListKind:
		l := expr.AsList()
		for _, e := range l.Elements() {
			idents = append(idents, ToIdentifiers(e)...)
		}
	case ast.MapKind:
		m := expr.AsMap()
		for _, entry := range m.Entries() {
			idents = append(idents, toEntryNames(entry)...)
		}
	case ast.SelectKind:
		idents = append(idents, toSelectorName(expr))
	case ast.StructKind:
		idents = append(idents, expr.AsStruct().TypeName())
		for _, field := range expr.AsStruct().Fields() {
			idents = append(idents, toEntryNames(field)...)
		}
	}
	return idents
}

func toEntryNames(entry ast.EntryExpr) []string {
	var ident []string
	switch entry.Kind() {
	case ast.MapEntryKind:
		ident = append(ident, ToIdentifiers(entry.AsMapEntry().Key())...)
		ident = append(ident, ToIdentifiers(entry.AsMapEntry().Value())...)
	case ast.StructFieldKind:
		ident = append(ident, ToIdentifiers(entry.AsStructField().Value())...)
	}
	return ident
}

func toSelectorName(v ast.Expr) string {
	switch v.Kind() {
	case ast.SelectKind:
		sel := v.AsSelect()
		parent := toSelectorName(sel.Operand())
		if parent != "" {
			return parent + "." + sel.FieldName()
		}
		return sel.FieldName()
	case ast.IdentKind:
		return v.AsIdent()
	default:
		return ""
	}
}
