package generator_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// assertPrecompileCoversCacheIndexes checks that every cache index used by the generated code of a service
// has an entry in the precompile table of the same service, otherwise the expression would silently be compiled lazily.
func assertPrecompileCoversCacheIndexes(t *testing.T, generated string) {
	t.Helper()

	problems, err := precompileProblems(generated)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range problems {
		t.Error(problem)
	}
}

// precompileProblems parses the generated Go code and returns what is wrong with the precompile tables.
// A service is identified by the receiver of the methods, and its table is its precompileCEL method.
//
// An index is used by a composite literal that also holds the text of the expression, in a key named after the
// index key: ByCacheIndex goes with By, IfCacheIndex with If. CacheIndex goes with whichever of expressionKeys
// the literal has. The text must be the one of the table entry with that index.
func precompileProblems(generated string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", generated, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the generated code: %w", err)
	}

	type use struct {
		key   string
		index int
		expr  string
		found bool // whether the literal has the text of the expression
	}
	used := map[string][]use{}
	entries := map[string]map[int][]string{} // service -> index -> expressions of the entries
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		service := receiverName(fn)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			fields := map[string]ast.Expr{}
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if key, ok := kv.Key.(*ast.Ident); ok {
					fields[key.Name] = kv.Value
				}
			}
			if index, ok := intLiteral(fields["Index"]); ok && fn.Name.Name == "precompileCEL" {
				if entries[service] == nil {
					entries[service] = map[int][]string{}
				}
				expr, _ := stringLiteral(fields["Expr"])
				entries[service][index] = append(entries[service][index], expr)
			}
			for _, key := range sortedKeys(fields) {
				if !strings.HasSuffix(key, "CacheIndex") {
					continue
				}
				index, ok := intLiteral(fields[key])
				if !ok {
					continue
				}
				u := use{key: key, index: index}
				for _, textKey := range textKeys(key, fields) {
					if expr, ok := stringLiteral(fields[textKey]); ok {
						u.expr, u.found = expr, true
						break
					}
				}
				used[service] = append(used[service], u)
			}
			return true
		})
	}

	var problems []string
	for _, service := range sortedKeys(used) {
		for _, u := range used[service] {
			exprs := entries[service][u.index]
			switch {
			case len(exprs) == 0:
				problems = append(problems, fmt.Sprintf("%s: cache index %d is used but is not precompiled", serviceLabel(service), u.index))
			case !u.found:
				problems = append(problems, fmt.Sprintf("%s: %s %d has no expression text next to it", serviceLabel(service), u.key, u.index))
			case !slices.Contains(exprs, u.expr):
				problems = append(problems, fmt.Sprintf("%s: cache index %d is used by %q but is precompiled as %q", serviceLabel(service), u.index, u.expr, exprs))
			}
		}
	}
	for _, service := range sortedKeys(entries) {
		for index, exprs := range entries[service] {
			if len(exprs) > 1 {
				problems = append(problems, fmt.Sprintf("%s: cache index %d has %d precompile entries", serviceLabel(service), index, len(exprs)))
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
}

// expressionKeys are the keys that hold the text of the expression next to a plain CacheIndex.
// MessageValueName is the name of the variable that error detail messages are evaluated by.
var expressionKeys = []string{"Expr", "If", "Message", "MessageValueName"}

// textKeys returns the keys that can hold the text of the expression that an index key refers to.
func textKeys(indexKey string, fields map[string]ast.Expr) []string {
	prefix := strings.TrimSuffix(indexKey, "CacheIndex")
	if prefix != "" {
		return []string{prefix}
	}
	var keys []string
	for _, key := range expressionKeys {
		if _, ok := fields[key]; ok {
			keys = append(keys, key)
		}
	}
	return keys
}

func receiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	typ := fn.Recv.List[0].Type
	if star, ok := typ.(*ast.StarExpr); ok {
		typ = star.X
	}
	if ident, ok := typ.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return v, true
}

func intLiteral(expr ast.Expr) (int, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return 0, false
	}
	v, err := strconv.Atoi(lit.Value)
	if err != nil {
		return 0, false
	}
	return v, true
}

func serviceLabel(service string) string {
	if service == "" {
		return "(a function without a receiver)"
	}
	return service
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestPrecompileProblems(t *testing.T) {
	t.Parallel()

	const header = "package federation\n\nimport \"context\"\n\n"
	for _, test := range []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "every used index has an entry",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: "a", CacheIndex: 1, By: "b", ByCacheIndex: 2}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}, {Index: 2, Expr: "b"}}
	return nil
}`,
		},
		{
			name: "keys aligned by gofmt are found",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{
		Expr:           "a",
		CacheIndex:     1,
		Type:           "t",
		TypeCacheIndex: 2,
	}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}}
	return nil
}`,
			want: []string{"A: cache index 2 is used but is not precompiled"},
		},
		{
			name: "an index that only another service has an entry for is missing",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: "a", CacheIndex: 1}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 5, Expr: "a"}}
	return nil
}
func (s *B) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}}
	return nil
}`,
			want: []string{"A: cache index 1 is used but is not precompiled"},
		},
		{
			name: "a duplicated entry",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: "a", CacheIndex: 1}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}, {Index: 1, Expr: "b"}}
	return nil
}`,
			want: []string{"A: cache index 1 has 2 precompile entries"},
		},
		{
			name: "an index used outside of a service",
			src: header + `
func helper() {
	_ = p{Expr: "a", CacheIndex: 3}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 3, Expr: "a"}}
	return nil
}`,
			want: []string{"(a function without a receiver): cache index 3 is used but is not precompiled"},
		},
		{
			name: "other keys that end differently are ignored",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{MessageIndex: 7, Index: 8}
}
func (s *A) precompileCEL(ctx context.Context) error {
	return nil
}`,
		},
		{
			name: "an index used for another expression than the entry has",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: "b", CacheIndex: 1}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}}
	return nil
}`,
			want: []string{`A: cache index 1 is used by "b" but is precompiled as ["a"]`},
		},
		{
			name: "the text of each index key is the key it is named after",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: "x", If: "a", IfCacheIndex: 1, By: "b", ByCacheIndex: 2}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}, {Index: 2, Expr: "a"}}
	return nil
}`,
			want: []string{`A: cache index 2 is used by "b" but is precompiled as ["a"]`},
		},
		{
			name: "the expression of a plain index can be in any of the expression keys",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{If: "a", CacheIndex: 1}
	_ = p{Message: "b", Locale: "en", CacheIndex: 2}
	_ = p{MessageValueName: "c", CacheIndex: 3}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}, {Index: 2, Expr: "b"}, {Index: 3, Expr: "c"}}
	return nil
}`,
		},
		{
			name: "an index without the text of the expression",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{CacheIndex: 1}
	_ = p{ByCacheIndex: 1, Expr: "a"}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: "a"}}
	return nil
}`,
			want: []string{"A: ByCacheIndex 1 has no expression text next to it", "A: CacheIndex 1 has no expression text next to it"},
		},
		{
			name: "raw and quoted strings are compared by value",
			src: header + `
func (s *A) resolve(ctx context.Context) {
	_ = p{Expr: ` + "`a\\nb`" + `, CacheIndex: 1}
}
func (s *A) precompileCEL(ctx context.Context) error {
	_ = []e{{Index: 1, Expr: ` + "`a\\nb`" + `}}
	return nil
}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := precompileProblems(test.src)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(got, "\n") != strings.Join(test.want, "\n") {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
