package resolver

func (e *CallExpr) ReferenceNames() []string {
	if e.Request == nil {
		return nil
	}

	var refNames []string
	for _, arg := range e.Request.Args {
		refNames = append(refNames, arg.Value.ReferenceNames()...)
	}
	return refNames
}
