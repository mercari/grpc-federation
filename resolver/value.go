package resolver

import (
	"github.com/mercari/grpc-federation/grpc/federation"
)

func (v *Value) Type() *Type {
	if v == nil {
		return nil
	}
	if v.CEL != nil {
		return v.CEL.Out
	}
	return nil
}

func (v *Value) ReferenceNames() []string {
	if v == nil {
		return nil
	}
	return v.CEL.ReferenceNames()
}

func (v *CELValue) ReferenceNames() []string {
	if v == nil {
		return nil
	}
	if v.CheckedExpr == nil {
		return nil
	}

	var refNames []string
	iterVarMap := make(map[string]struct{})
	if cmp := v.CheckedExpr.GetExpr().GetComprehensionExpr(); cmp != nil {
		if iterVar := cmp.GetIterVar(); iterVar != "" {
			iterVarMap[iterVar] = struct{}{}
		}
		if iterVar := cmp.GetIterVar2(); iterVar != "" {
			iterVarMap[iterVar] = struct{}{}
		}
	}
	for _, ref := range v.CheckedExpr.ReferenceMap {
		if ref.Name == federation.MessageArgumentVariableName {
			continue
		}
		// Name can be empty sting if the reference points to a function
		if ref.Name == "" {
			continue
		}
		if _, exists := iterVarMap[ref.Name]; exists {
			continue
		}
		refNames = append(refNames, ref.Name)
	}
	return refNames
}

type commonValueDef struct {
	CustomResolver *bool
	By             *string
	Inline         *string
	Alias          *string
}

func (def *commonValueDef) GetBy() string {
	if def != nil && def.By != nil {
		return *def.By
	}
	return ""
}

func (def *commonValueDef) GetInline() string {
	if def != nil && def.Inline != nil {
		return *def.Inline
	}
	return ""
}

func (def *commonValueDef) GetAlias() string {
	if def != nil && def.Alias != nil {
		return *def.Alias
	}
	return ""
}

func fieldRuleToCommonValueDef(def *federation.FieldRule) *commonValueDef {
	return &commonValueDef{
		CustomResolver: def.CustomResolver,
		By:             def.By,
		Alias:          def.Alias,
	}
}

func methodRequestToCommonValueDef(def *federation.MethodRequest) *commonValueDef {
	return &commonValueDef{
		By: def.By,
	}
}

func argumentToCommonValueDef(def *federation.Argument) *commonValueDef {
	return &commonValueDef{
		By:     def.By,
		Inline: def.Inline,
	}
}

func NewByValue(expr string, out *Type) *Value {
	return &Value{
		CEL: &CELValue{
			Expr: expr,
			Out:  out,
		},
	}
}
