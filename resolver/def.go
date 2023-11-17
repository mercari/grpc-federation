package resolver

func (def *VariableDefinition) ReferenceNames() []string {
	if def.Expr == nil {
		return nil
	}
	expr := def.Expr
	switch {
	case expr.By != nil:
		return expr.By.ReferenceNames()
	case expr.Map != nil:
		return expr.Map.ReferenceNames()
	case expr.Call != nil:
		return expr.Call.ReferenceNames()
	case expr.Message != nil:
		return expr.Message.ReferenceNames()
	case expr.Validation != nil:
		return expr.Validation.Error.ReferenceNames()
	}
	return nil
}
