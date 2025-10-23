package cel

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type BindFunctionOpt struct {
	cel.FunctionOpt
}

type BindMemberFunctionOpt struct {
	cel.FunctionOpt
}

var (
	funcNameMap = make(map[string]struct{})
	funcNameMu  sync.RWMutex
)

func BindFunction(name string, opts ...BindFunctionOpt) []cel.EnvOption {
	parts := strings.Split(name, ".")
	funcName := parts[len(parts)-1]
	celOpts := make([]cel.FunctionOpt, 0, len(opts))
	for _, opt := range opts {
		celOpts = append(celOpts, opt.FunctionOpt)
	}
	funcNameMu.Lock()
	defer funcNameMu.Unlock()
	funcNameMap[name] = struct{}{}

	return []cel.EnvOption{
		createMacro(funcName),
		cel.Function(name, celOpts...),
	}
}

func createMacro(funcName string) cel.EnvOption {
	return cel.Macros(cel.ReceiverVarArgMacro(funcName, func(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
		sel := toSelectorName(target)
		fqdn := funcName
		if sel != "" {
			fqdn = sel + "." + funcName
		}
		fqdn = strings.TrimPrefix(fqdn, ".")
		funcNameMu.RLock()
		defer funcNameMu.RUnlock()

		var argsWithCtx []ast.Expr
		if len(args) > 0 && args[0].AsIdent() == ContextVariableName {
			argsWithCtx = args
		} else {
			argsWithCtx = append([]ast.Expr{mef.NewIdent(ContextVariableName)}, args...)
		}
		if _, exists := funcNameMap[fqdn]; !exists {
			return mef.NewMemberCall(funcName, target, argsWithCtx...), nil
		}
		return mef.NewCall(fqdn, argsWithCtx...), nil
	}))
}

func BindMemberFunction(name string, opts ...BindMemberFunctionOpt) []cel.EnvOption {
	celOpts := make([]cel.FunctionOpt, 0, len(opts))
	for _, opt := range opts {
		celOpts = append(celOpts, opt.FunctionOpt)
	}
	return []cel.EnvOption{
		createMacro(name),
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

func BindExtFunction(extLib cel.EnvOption, name, signature string, args []*cel.Type, result *cel.Type) []cel.EnvOption {
	decls := []cel.EnvOption{}
	argNames := make([]string, 0, len(args))
	for idx, typ := range args {
		argName := fmt.Sprintf("arg%d", idx)
		decls = append(decls, cel.Variable(argName, typ))
		argNames = append(argNames, argName)
	}
	prg, prgErr := compileExt(name, argNames, append([]cel.EnvOption{extLib}, decls...))
	return BindFunction(
		name,
		BindFunctionOpt{
			FunctionOpt: cel.Overload(signature,
				append([]*cel.Type{cel.ObjectType(ContextTypeName)}, args...),
				result,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					if prgErr != nil {
						return types.NewErrFromString(prgErr.Error())
					}
					args := map[string]any{}
					for idx, value := range values[1:] {
						args[fmt.Sprintf("arg%d", idx)] = value
					}
					out, _, err := prg.Eval(args)
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return out
				}),
			),
		},
	)
}

func BindExtMemberFunction(extLib cel.EnvOption, name, signature string, self *cel.Type, args []*cel.Type, result *cel.Type) []cel.EnvOption {
	decls := []cel.EnvOption{cel.Variable("self", self)}
	argNames := make([]string, 0, len(args))
	for idx, typ := range args {
		argName := fmt.Sprintf("arg%d", idx)
		decls = append(decls, cel.Variable(argName, typ))
		argNames = append(argNames, argName)
	}
	prg, prgErr := compileMemberExt(name, argNames, append([]cel.EnvOption{extLib}, decls...))
	return BindMemberFunction(
		name,
		BindMemberFunctionOpt{
			FunctionOpt: cel.MemberOverload(signature,
				append([]*cel.Type{self, cel.ObjectType(ContextTypeName)}, args...),
				result,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					if prgErr != nil {
						return types.NewErrFromString(prgErr.Error())
					}
					self := values[0]
					args := map[string]any{"self": self}
					for idx, value := range values[2:] {
						args[fmt.Sprintf("arg%d", idx)] = value
					}
					out, _, err := prg.Eval(args)
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return out
				}),
			),
		},
	)
}

func compileExt(name string, argNames []string, opts []cel.EnvOption) (cel.Program, error) {
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, err
	}
	ast, iss := env.Compile(fmt.Sprintf("%s(%s)", name, strings.Join(argNames, ",")))
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return prg, nil
}

func compileMemberExt(name string, argNames []string, opts []cel.EnvOption) (cel.Program, error) {
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, err
	}
	ast, iss := env.Compile(fmt.Sprintf("self.%s(%s)", name, strings.Join(argNames, ",")))
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	return prg, nil
}
