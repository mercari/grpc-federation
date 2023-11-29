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

func (def *VariableDefinition) MessageExprs() []*MessageExpr {
	if def.Expr == nil {
		return nil
	}
	expr := def.Expr
	switch {
	case expr.Map != nil && expr.Map.Expr != nil:
		return []*MessageExpr{expr.Map.Expr.Message}
	case expr.Message != nil:
		return []*MessageExpr{expr.Message}
	case expr.Validation != nil && expr.Validation.Error != nil:
		var ret []*MessageExpr
		for _, detail := range expr.Validation.Error.Details {
			for _, msg := range detail.Messages {
				ret = append(ret, msg.Expr.Message)
			}
		}
		return ret
	}
	return nil
}
