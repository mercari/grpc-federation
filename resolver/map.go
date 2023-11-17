package resolver

import "sort"

func (e *MapExpr) ReferenceNames() []string {
	expr := e.Expr
	if expr == nil {
		return nil
	}
	refNameMap := make(map[string]struct{})
	switch {
	case expr.By != nil:
		for _, name := range expr.By.ReferenceNames() {
			refNameMap[name] = struct{}{}
		}
	case expr.Message != nil:
		for _, name := range expr.Message.ReferenceNames() {
			refNameMap[name] = struct{}{}
		}
	}
	delete(refNameMap, e.Iterator.Name)

	refNames := make([]string, 0, len(refNameMap))
	for name := range refNameMap {
		refNames = append(refNames, name)
	}
	sort.Strings(refNames)
	return refNames
}
