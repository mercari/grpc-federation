package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/parser"
)

const ListPackageName = "list"

type ListLibrary struct {
}

func (lib *ListLibrary) LibraryName() string {
	return packageName(ListPackageName)
}

func (lib *ListLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.Macros(
			// range.reduce(accumulator, current, <expr>, <init>)
			cel.ReceiverMacro("reduce", 4, makeReduce),

			// range.first(var, <expr>)
			// first macro is the shorthand version of the following expression
			// <range>.filter(iter, expr)[?0]
			cel.ReceiverMacro("first", 2, makeFirst),
		),
	}
	return opts
}

func (lib *ListLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func makeReduce(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	accum, found := extractIdent(args[0])
	if !found {
		return nil, mef.NewError(args[0].ID(), "argument is not an identifier")
	}
	cur, found := extractIdent(args[1])
	if !found {
		return nil, mef.NewError(args[1].ID(), "argument is not an identifier")
	}
	reduce := args[2]
	init := args[3]
	condition := mef.NewLiteral(types.True)
	accuExpr := mef.NewIdent(accum)
	return mef.NewComprehension(target, cur, accum, init, condition, reduce, accuExpr), nil
}

func makeFirst(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	filter, err := parser.MakeFilter(mef, target, args)
	if err != nil {
		return nil, err
	}
	return mef.NewCall(operators.OptIndex, filter, mef.NewLiteral(types.Int(0))), nil
}

func extractIdent(e ast.Expr) (string, bool) {
	switch e.Kind() {
	case ast.IdentKind:
		return e.AsIdent(), true
	}
	return "", false
}
