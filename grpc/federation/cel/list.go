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
	typeAdapter types.Adapter
}

func NewListLibrary(typeAdapter types.Adapter) cel.SingletonLibrary {
	return &ListLibrary{
		typeAdapter: typeAdapter,
	}
}

func (lib *ListLibrary) LibraryName() string {
	return packageName(ListPackageName)
}

func (lib *ListLibrary) CompileOptions() []cel.EnvOption {
	listTypeO := cel.ListType(cel.TypeParamType("O"))
	listTypeE := cel.ListType(cel.TypeParamType("E"))
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.Macros(
			// range.reduce(accumulator, current, <expr>, <init>)
			cel.ReceiverMacro("reduce", 4, lib.makeReduce),

			// range.first(var, <expr>)
			// first macro is the shorthand version of the following expression
			// <range>.filter(iter, expr)[?0]
			cel.ReceiverMacro("first", 2, lib.makeFirst),

			// <range>.sortAsc(var, <expr>)
			cel.ReceiverMacro("sortAsc", 2, lib.makeSortAsc),

			// <range>.sortDesc(var, <expr>)
			cel.ReceiverMacro("sortDesc", 2, lib.makeSortDesc),

			// <range>.sortStableAsc(var, <expr>)
			cel.ReceiverMacro("sortStableAsc", 2, lib.makeSortStableAsc),

			// <range>.sortStableDesc(var, <expr>)
			cel.ReceiverMacro("sortStableDesc", 2, lib.makeSortStableDesc),
		),
		cel.Function(listSortAscFunc,
			cel.Overload("grpc_federation_list_@sort_asc",
				[]*cel.Type{listTypeO, listTypeE}, listTypeO,
				cel.BinaryBinding(lib.sortAsc),
			),
		),
		cel.Function(listSortDescFunc,
			cel.Overload("grpc_federation_list_@sort_desc",
				[]*cel.Type{listTypeO, listTypeE}, listTypeO,
				cel.BinaryBinding(lib.sortDesc),
			),
		),
		cel.Function(listSortStableAscFunc,
			cel.Overload("grpc_federation_list_@sort_stable_asc",
				[]*cel.Type{listTypeO, listTypeE}, listTypeO,
				cel.BinaryBinding(lib.sortStableAsc),
			),
		),
		cel.Function(listSortStableDescFunc,
			cel.Overload("grpc_federation_list_@sort_stable_desc",
				[]*cel.Type{listTypeO, listTypeE}, listTypeO,
				cel.BinaryBinding(lib.sortStableDesc),
			),
		),
	}
	return opts
}

func (lib *ListLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (lib *ListLibrary) makeReduce(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
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

func (lib *ListLibrary) makeFirst(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	filter, err := parser.MakeFilter(mef, target, args)
	if err != nil {
		return nil, err
	}
	return mef.NewCall(operators.OptIndex, filter, mef.NewLiteral(types.Int(0))), nil
}

func (lib *ListLibrary) makeSortAsc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return lib.makeSort(listSortAscFunc, mef, target, args)
}

func (lib *ListLibrary) makeSortDesc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return lib.makeSort(listSortDescFunc, mef, target, args)
}

func (lib *ListLibrary) makeSortStableAsc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return lib.makeSort(listSortStableAscFunc, mef, target, args)
}

func (lib *ListLibrary) makeSortStableDesc(mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	return lib.makeSort(listSortStableDescFunc, mef, target, args)
}

func (lib *ListLibrary) makeSort(function string, mef cel.MacroExprFactory, target ast.Expr, args []ast.Expr) (ast.Expr, *cel.Error) {
	mp, err := parser.MakeMap(mef, target, args)
	if err != nil {
		return nil, err
	}
	// Avoid incompatible type already exists for expression error by copying a map expression with new set of identifiers
	return mef.NewCall(function, target, mef.Copy(mp)), nil
}

func (lib *ListLibrary) sortAsc(orig, expanded ref.Val) ref.Val {
	return lib.sortRefVal(orig, expanded, -types.IntOne, false)
}

func (lib *ListLibrary) sortDesc(orig, expanded ref.Val) ref.Val {
	return lib.sortRefVal(orig, expanded, types.IntOne, false)
}

func (lib *ListLibrary) sortStableAsc(orig, expanded ref.Val) ref.Val {
	return lib.sortRefVal(orig, expanded, -types.IntOne, true)
}

func (lib *ListLibrary) sortStableDesc(orig, expanded ref.Val) ref.Val {
	return lib.sortRefVal(orig, expanded, types.IntOne, true)
}

func (lib *ListLibrary) sortRefVal(orig, expanded ref.Val, direction types.Int, stable bool) ref.Val {
	origLister := orig.(traits.Lister)
	expandedLister := expanded.(traits.Lister)

	// The element being sorted must be values before expansion.
	type sortVal struct {
		OrigVal     ref.Val
		ExpandedVal ref.Val
	}
	vals := make([]*sortVal, 0, int64(origLister.Size().(types.Int)))
	for i := types.IntZero; i < origLister.Size().(types.Int); i++ {
		origVal := origLister.Get(i)
		expVal := expandedLister.Get(i)
		if _, ok := expVal.(traits.Comparer); !ok {
			return types.NewErr("%s of list[%d] is not comparable", expVal.Type(), i)
		}
		vals = append(vals, &sortVal{
			OrigVal:     origVal,
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

	resVals := make([]ref.Val, 0, len(vals))
	for _, v := range vals {
		resVals = append(resVals, v.OrigVal)
	}
	return lib.typeAdapter.NativeToValue(resVals)
}

func extractIdent(e ast.Expr) (string, bool) {
	switch e.Kind() {
	case ast.IdentKind:
		return e.AsIdent(), true
	}
	return "", false
}
