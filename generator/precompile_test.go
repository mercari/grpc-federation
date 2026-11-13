package generator_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
func precompileProblems(generated string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", generated, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the generated code: %w", err)
	}

	used := map[string]map[int]struct{}{}
	entries := map[string]map[int]int{} // service -> index -> number of entries
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		service := receiverName(fn)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			kv, ok := n.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				return true
			}
			index, ok := intLiteral(kv.Value)
			if !ok {
				return true
			}
			switch {
			case fn.Name.Name == "precompileCEL" && key.Name == "Index":
				if entries[service] == nil {
					entries[service] = map[int]int{}
				}
				entries[service][index]++
			case strings.HasSuffix(key.Name, "CacheIndex"):
				if used[service] == nil {
					used[service] = map[int]struct{}{}
				}
				used[service][index] = struct{}{}
			}
			return true
		})
	}

	var problems []string
	for _, service := range sortedKeys(used) {
		for _, index := range sortedIndexes(used[service]) {
			if entries[service][index] == 0 {
				problems = append(problems, fmt.Sprintf("%s: cache index %d is used but is not precompiled", serviceLabel(service), index))
			}
		}
	}
	for _, service := range sortedKeys(entries) {
		for index, count := range entries[service] {
			if count > 1 {
				problems = append(problems, fmt.Sprintf("%s: cache index %d has %d precompile entries", serviceLabel(service), index, count))
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
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

func sortedIndexes(m map[int]struct{}) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
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
	_ = p{CacheIndex: 1, ByCacheIndex: 2}
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
		CacheIndex:     1,
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
	_ = p{CacheIndex: 1}
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
	_ = p{CacheIndex: 1}
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
	_ = p{CacheIndex: 3}
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
