package resolver

import "sort"

func (e *MapExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}

	refNameMap := make(map[string]struct{})
	if e.Iterator != nil && e.Iterator.Name != "" {
		refNameMap[e.Iterator.Name] = struct{}{}
	}
	for _, name := range e.Expr.ReferenceNames() {
		refNameMap[name] = struct{}{}
	}

	refNames := make([]string, 0, len(refNameMap))
	for name := range refNameMap {
		refNames = append(refNames, name)
	}
	sort.Strings(refNames)
	return refNames
}

func (e *MapIteratorExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}

	refNameMap := make(map[string]struct{})
	switch {
	case e.By != nil:
		for _, name := range e.By.ReferenceNames() {
			refNameMap[name] = struct{}{}
		}
	case e.Message != nil:
		for _, name := range e.Message.ReferenceNames() {
			refNameMap[name] = struct{}{}
		}
	}

	refNames := make([]string, 0, len(refNameMap))
	for name := range refNameMap {
		refNames = append(refNames, name)
	}
	sort.Strings(refNames)
	return refNames
}
