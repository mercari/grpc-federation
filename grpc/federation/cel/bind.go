package cel

import (
	"context"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/types/ref"
)

type BindFunctionOpt struct {
	cel.FunctionOpt
}

type BindMemberFunctionOpt struct {
	cel.FunctionOpt
}

func BindFunction(name string, opts ...BindFunctionOpt) []cel.EnvOption {
	parts := strings.Split(name, ".")
	funcName := parts[len(parts)-1]
	celOpts := make([]cel.FunctionOpt, 0, len(opts))
	for _, opt := range opts {
		celOpts = append(celOpts, opt.FunctionOpt)
	}
	return []cel.EnvOption{
		cel.Macros(cel.ReceiverVarArgMacro(funcName, func(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
			sel := toSelectorName(target)
			fqdn := funcName
			if sel != "" {
				fqdn = sel + "." + funcName
			}
			var argsWithCtx []ast.Expr
			if len(args) > 0 && args[0].AsIdent() == ContextVariableName {
				argsWithCtx = args
			} else {
				argsWithCtx = append([]ast.Expr{mef.NewIdent(ContextVariableName)}, args...)
			}
			return mef.NewCall(fqdn, argsWithCtx...), nil
		})),
		cel.Function(name, celOpts...),
	}
}

func BindMemberFunction(name string, opts ...BindMemberFunctionOpt) []cel.EnvOption {
	celOpts := make([]cel.FunctionOpt, 0, len(opts))
	for _, opt := range opts {
		celOpts = append(celOpts, opt.FunctionOpt)
	}
	return []cel.EnvOption{
		cel.Macros(cel.ReceiverVarArgMacro(name, func(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
			var argsWithCtx []ast.Expr
			if len(args) > 0 && args[0].AsIdent() == ContextVariableName {
				argsWithCtx = args
			} else {
				argsWithCtx = append([]ast.Expr{mef.NewIdent(ContextVariableName)}, args...)
			}
			return mef.NewMemberCall(name, target, argsWithCtx...), nil
		})),
		cel.Function(name, celOpts...),
	}
}

func OverloadFunc(name string, args []*cel.Type, result *cel.Type, cb func(ctx context.Context, values ...ref.Val) ref.Val) BindFunctionOpt {
	return BindFunctionOpt{
		FunctionOpt: cel.Overload(name,
			append([]*cel.Type{cel.ObjectType(ContextTypeName)}, args...),
			result,
			cel.FunctionBinding(func(values ...ref.Val) ref.Val {
				ctx := values[0].(*ContextValue).Context
				return cb(ctx, values[1:]...)
			}),
		),
	}
}

func MemberOverloadFunc(name string, self *cel.Type, args []*cel.Type, result *cel.Type, cb func(ctx context.Context, self ref.Val, args ...ref.Val) ref.Val) BindMemberFunctionOpt {
	return BindMemberFunctionOpt{
		FunctionOpt: cel.MemberOverload(name,
			append([]*cel.Type{self, cel.ObjectType(ContextTypeName)}, args...),
			result,
			cel.FunctionBinding(func(values ...ref.Val) ref.Val {
				self := values[0]
				ctx := values[1].(*ContextValue).Context
				return cb(ctx, self, values[2:]...)
			}),
		),
	}
}
