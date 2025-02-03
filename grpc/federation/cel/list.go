package cel

import (
	"fmt"
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
		cel.Function("flatten",
			cel.MemberOverload("list_flatten",
				[]*cel.Type{cel.ListType(listTypeO)}, listTypeO,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					list, ok := arg.(traits.Lister)
					if !ok {
						return types.ValOrErr(arg, "no such overload: %v.flatten()", arg.Type())
					}
					flatList, err := flatten(list, 1)
					if err != nil {
						return types.WrapErr(err)
					}
					return lib.typeAdapter.NativeToValue(flatList)
				}),
			),
		),
	}
	return opts
}

func flatten(list traits.Lister, depth int64) ([]ref.Val, error) {
	if depth < 0 {
		return nil, fmt.Errorf("level must be non-negative")
	}

	var newList []ref.Val
	iter := list.Iterator()

	for iter.HasNext() == types.True {
		val := iter.Next()
		nestedList, isList := val.(traits.Lister)

		if !isList || depth == 0 {
			newList = append(newList, val)
			continue
		} else {
			flattenedList, err := flatten(nestedList, depth-1)
			if err != nil {
				return nil, err
			}

			newList = append(newList, flattenedList...)
		}
	}

	return newList, nil
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
	// When sort macros are called in a chain (e.g., [1, 2].sortAsc(v, v).sortDesc(v, v)), the makeSort method is called two or more times.
	// Then, on the second and subsequent calls to the makeSort method, the call expression created by the previous makeSort method is passed to the target argument.
	// Passing it as is to the global sort function call (e.g., grpc.federaiton.list.@sortAsc) will cause an `incompatible type already exists for expression error`.
	// To avoid this, use mef.Copy for the target argument to assign a new set of identifiers.
	return mef.NewCall(function, mef.Copy(target), mp), nil
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
		vals = append(vals, &sortVal{
			OrigVal:     origVal,
			ExpandedVal: expVal,
		})
	}

	fn := sort.Slice
	if stable {
		fn = sort.SliceStable
	}
	fn(vals, func(i, j int) bool {
		cmp := vals[i].ExpandedVal.(traits.Comparer)
		out := cmp.Compare(vals[j].ExpandedVal)
		return out == direction
	})

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

type listValidator struct{}

func NewListValidator() cel.ASTValidator {
	return &listValidator{}
}

func (v *listValidator) Name() string {
	return "grpc.federation.list.validator"
}

func (v *listValidator) Validate(_ *cel.Env, _ cel.ValidatorConfig, a *ast.AST, iss *cel.Issues) {
	root := ast.NavigateAST(a)

	// Checks at compile time if the value types are comparable
	funcNames := []string{
		listSortAscFunc,
		listSortDescFunc,
		listSortStableAscFunc,
		listSortStableDescFunc,
	}
	dupErrs := map[string]struct{}{}
	for _, funcName := range funcNames {
		funcCalls := ast.MatchDescendants(root, ast.FunctionMatcher(funcName))
		for _, call := range funcCalls {
			arg1 := call.AsCall().Args()[1]
			expr, ok := arg1.AsComprehension().Result().(ast.NavigableExpr)
			if !ok {
				continue
			}
			params := expr.Type().Parameters()
			if len(params) == 0 {
				continue
			}
			// If an error occurs when calling sort macros in a chain, the same error is reported multiple times.
			// To avoid this, check if the error already report and ignore it.
			loc := a.SourceInfo().GetStartLocation(expr.ID())
			dupKey := fmt.Sprintf("%s:%d:%d:%s", funcName, loc.Column(), loc.Line(), expr.Type())
			if _, exists := dupErrs[dupKey]; !exists && !params[0].HasTrait(traits.ComparerType) {
				iss.ReportErrorAtID(expr.ID(), "%s is not comparable", expr.Type())
				dupErrs[dupKey] = struct{}{}
			}
		}
	}
}
