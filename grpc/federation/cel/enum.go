package cel

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const (
	EnumPackageName  = "enum"
	EnumSelectorFQDN = "grpc.federation.private.EnumSelector"
)

type EnumLibrary struct {
}

func (lib *EnumLibrary) LibraryName() string {
	return packageName(EnumPackageName)
}

func (lib *EnumLibrary) CompileOptions() []cel.EnvOption {
	typeT := celtypes.NewTypeParamType("T")
	typeU := celtypes.NewTypeParamType("U")
	enumSelector := celtypes.NewOpaqueType(EnumSelectorFQDN, typeT, typeU)
	opts := []cel.EnvOption{
		cel.Function("grpc.federation.enum.select",
			cel.Overload("grpc_federation_enum_select",
				[]*cel.Type{cel.BoolType, typeT, typeU}, enumSelector,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					ret := &EnumSelector{
						Cond: bool(values[0].(celtypes.Bool)),
					}
					if sel, ok := values[1].(*EnumSelector); ok {
						ret.True = &EnumSelector_TrueSelector{
							TrueSelector: sel,
						}
					} else {
						ret.True = &EnumSelector_TrueValue{
							TrueValue: int32(values[1].(celtypes.Int)),
						}
					}
					if sel, ok := values[2].(*EnumSelector); ok {
						ret.False = &EnumSelector_FalseSelector{
							FalseSelector: sel,
						}
					} else {
						ret.False = &EnumSelector_FalseValue{
							FalseValue: int32(values[2].(celtypes.Int)),
						}
					}
					return ret
				}),
			),
		),
	}
	return opts
}

func (lib *EnumLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (s *EnumSelector) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return s, nil
}

func (s *EnumSelector) ConvertToType(typeValue ref.Type) ref.Val {
	return celtypes.NewErr(fmt.Sprintf("%s: type conversion does not support", EnumSelectorFQDN))
}

func (s *EnumSelector) Equal(other ref.Val) ref.Val {
	if _, ok := other.(*EnumSelector); ok {
		return celtypes.True
	}
	return celtypes.False
}

func (s *EnumSelector) Type() ref.Type {
	return celtypes.NewObjectType(EnumSelectorFQDN)
}

func (s *EnumSelector) Value() any {
	return s
}

type enumValidator struct{}

func NewEnumValidator() cel.ASTValidator {
	return &enumValidator{}
}

func (v *enumValidator) Name() string {
	return "grpc.federation.enum.validator"
}

func (v *enumValidator) Validate(_ *cel.Env, _ cel.ValidatorConfig, a *ast.AST, iss *cel.Issues) {
	root := ast.NavigateAST(a)

	// Checks at compile time if the value types are int or opaque<int> or EnumSelector.
	dupErrs := map[string]struct{}{}
	funcName := "grpc.federation.enum.select"
	funcCalls := ast.MatchDescendants(root, ast.FunctionMatcher(funcName))

	candidateTypes := []*celtypes.Type{
		celtypes.NewOpaqueType(EnumSelectorFQDN),
		celtypes.IntType,
		celtypes.NewOpaqueType("enum", celtypes.IntType),
	}

	for _, call := range funcCalls {
		expr1 := v.getArgExpr(call, 1)
		expr2 := v.getArgExpr(call, 2)
		if expr1 == nil || expr2 == nil {
			continue
		}
		type1 := expr1.Type()
		type2 := expr2.Type()

		// If an error occurs when calling sort macros in a chain, the same error is reported multiple times.
		// To avoid this, check if the error already report and ignore it.
		dupKey1 := v.getKey(funcName, expr1, a)
		dupKey2 := v.getKey(funcName, expr2, a)
		if err := v.validateType(type1, candidateTypes...); err != nil {
			iss.ReportErrorAtID(expr1.ID(), err.Error())
			dupErrs[dupKey1] = struct{}{}
			continue
		}
		if err := v.validateType(type2, candidateTypes...); err != nil {
			iss.ReportErrorAtID(expr2.ID(), err.Error())
			dupErrs[dupKey2] = struct{}{}
			continue
		}
	}
}

func (v *enumValidator) getArgExpr(expr ast.Expr, argNum int) ast.NavigableExpr {
	args := expr.AsCall().Args()
	if len(args) <= argNum {
		return nil
	}
	arg := args[argNum]
	nav, ok := arg.(ast.NavigableExpr)
	if !ok {
		return nil
	}
	return nav
}

func (v *enumValidator) getKey(funcName string, expr ast.NavigableExpr, a *ast.AST) string {
	loc := a.SourceInfo().GetStartLocation(expr.ID())
	return fmt.Sprintf("%s:%d:%d:%s", funcName, loc.Column(), loc.Line(), expr.Type())
}

func (v *enumValidator) validateType(got *celtypes.Type, candidates ...*celtypes.Type) error {
	for _, candidate := range candidates {
		if got.Kind() != candidate.Kind() {
			continue
		}
		switch candidate.Kind() {
		case celtypes.StructKind:
			if got.TypeName() == candidate.TypeName() {
				return nil
			}
		case celtypes.OpaqueKind:
			if got.TypeName() == candidate.TypeName() {
				return nil
			}
			gotParams := got.Parameters()
			candidateParams := candidate.Parameters()
			if len(gotParams) != len(candidateParams) {
				continue
			}
			for i := 0; i < len(gotParams); i++ {
				if err := v.validateType(gotParams[i], candidateParams[i]); err == nil {
					return nil
				}
			}
		default:
			return nil
		}
	}
	return fmt.Errorf("%s type is unexpected", got.TypeName())
}
