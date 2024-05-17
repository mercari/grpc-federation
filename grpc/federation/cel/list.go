package cel

import (
	"sort"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/cel-go/parser"
)

const ListPackageName = "list"

const (
	listSortAscFunc        = "grpc.federaiton.list.@sortAsc"
	listSortDescFunc       = "grpc.federaiton.list.@sortDesc"
	listSortStableAscFunc  = "grpc.federaiton.list.@sortStableAsc"
	listSortStableDescFunc = "grpc.federaiton.list.@sortStableDesc"
)

type ListLibrary struct {
}

func (lib *ListLibrary) LibraryName() string {
	return packageName(ListPackageName)
}

func (lib *ListLibrary) CompileOptions() []cel.EnvOption {
	listTypeT := cel.ListType(cel.TypeParamType("T"))
	listTypeE := cel.ListType(cel.TypeParamType("E"))
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.Macros(
			// range.reduce(accumulator, current, <expr>, <init>)
			cel.ReceiverMacro("reduce", 4, makeReduce),

			// range.first(var, <expr>)
			// first macro is the shorthand version of the following expression
			// <range>.filter(iter, expr)[?0]
			cel.ReceiverMacro("first", 2, makeFirst),

			// <range>.sortAsc(var, <expr>)
			cel.ReceiverMacro("sortAsc", 2, makeSortAsc),

			// <range>.sortDesc(var, <expr>)
			cel.ReceiverMacro("sortDesc", 2, makeSortDesc),

			// <range>.sortStableAsc(var, <expr>)
			cel.ReceiverMacro("sortStableAsc", 2, makeSortStableAsc),

			// <range>.sortStableDesc(var, <expr>)
			cel.ReceiverMacro("sortStableDesc", 2, makeSortStableDesc),
		),
		cel.Function(listSortAscFunc,
			cel.Overload("grpc_federation_list_@sort_asc",
				[]*cel.Type{listTypeT, listTypeE}, listTypeT,
				cel.BinaryBinding(sortAsc),
			),
		),
		cel.Function(listSortDescFunc,
			cel.Overload("grpc_federation_list_@sort_desc",
				[]*cel.Type{listTypeT, listTypeE}, listTypeT,
				cel.BinaryBinding(sortDesc),
			),
		),
		cel.Function(listSortStableAscFunc,
			cel.Overload("grpc_federation_list_@sort_stable_asc",
				[]*cel.Type{listTypeT, listTypeE}, listTypeT,
				cel.BinaryBinding(sortStableAsc),
			),
		),
		cel.Function(listSortStableDescFunc,
			cel.Overload("grpc_federation_list_@sort_stable_desc",
				[]*cel.Type{listTypeT, listTypeE}, listTypeT,
				cel.BinaryBinding(sortStableDesc),
			),
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

func makeSortAsc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return makeSort(listSortAscFunc, mef, target, args)
}

func makeSortDesc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return makeSort(listSortDescFunc, mef, target, args)
}

func makeSortStableAsc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return makeSort(listSortStableAscFunc, mef, target, args)
}

func makeSortStableDesc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return makeSort(listSortStableDescFunc, mef, target, args)
}

func makeSort(function string, mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	mp, err := parser.MakeMap(mef, target, args)
	if err != nil {
		return nil, err
	}

	return mef.NewCall(function, target, mp), nil
}

func sortAsc(target, expanded ref.Val) ref.Val {
	return sortByOrder(target, expanded, -types.IntOne, false)
}

func sortDesc(target, expanded ref.Val) ref.Val {
	return sortByOrder(target, expanded, types.IntOne, false)
}

func sortStableAsc(target, expanded ref.Val) ref.Val {
	return sortByOrder(target, expanded, -types.IntOne, true)
}

func sortStableDesc(target, expanded ref.Val) ref.Val {
	return sortByOrder(target, expanded, types.IntOne, true)
}

func sortByOrder(target, expanded ref.Val, direction types.Int, stable bool) ref.Val {
	targetLister := target.(traits.Lister)
	expandedLister := expanded.(traits.Lister)
	if targetLister.Size().(types.Int) != expandedLister.Size().(types.Int) {
		return types.NewErr("size of list of target and expanded does not match")
	}

	type sortVal struct {
		TargetVal   ref.Val
		ExpandedVal ref.Val
	}
	vals := make([]*sortVal, 0, int64(targetLister.Size().(types.Int)))
	for i := types.IntZero; i < targetLister.Size().(types.Int); i++ {
		tgtVal := targetLister.Get(i)
		expVal := expandedLister.Get(i)
		if _, ok := expVal.(traits.Comparer); !ok {
			return types.NewErr("%s is not comparable", expVal.Type())
		}
		vals = append(vals, &sortVal{
			TargetVal:   tgtVal,
			ExpandedVal: expVal,
		})
	}

	fn := sort.Slice
	if stable {
		fn = sort.SliceStable
	}

	var hasErr bool
	fn(vals, func(i, j int) bool {
		cmp := vals[i].ExpandedVal.(traits.Comparer)
		out := cmp.Compare(vals[j].ExpandedVal)
		if types.IsUnknownOrError(out) {
			hasErr = true
		}
		return out == direction
	})
	if hasErr {
		return types.NewErr("cannot sort because some elements of list are not comparable")
	}

	resVal := make([]ref.Val, 0, len(vals))
	for _, v := range vals {
		resVal = append(resVal, v.TargetVal)
	}
	return types.DefaultTypeAdapter.NativeToValue(resVal)
}

func extractIdent(e ast.Expr) (string, bool) {
	switch e.Kind() {
	case ast.IdentKind:
		return e.AsIdent(), true
	}
	return "", false
}
