package resolver

import (
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"

	"github.com/mercari/grpc-federation/types"
)

func TestCELScopeReferenced(t *testing.T) {
	t.Parallel()

	typ := &Type{Kind: types.Int64}
	names := func(vars []*CELVariable) string {
		ret := make([]string, 0, len(vars))
		for _, v := range vars {
			ret = append(ret, v.Name)
		}
		return strings.Join(ret, ",")
	}
	env, err := cel.NewEnv(ext.TwoVarComprehensions())
	if err != nil {
		t.Fatal(err)
	}
	scope := &celScope{env: env}
	for _, name := range []string{"a", "b", "c"} {
		scope, err = scope.extend(name, typ)
		if err != nil {
			t.Fatal(err)
		}
	}
	// The comprehension tests need a list, so they use an env with one more variable that is not in the scope.
	listEnv, err := scope.env.Extend(cel.Variable("l", cel.ListType(cel.IntType)))
	if err != nil {
		t.Fatal(err)
	}
	listScope := &celScope{env: listEnv, vars: scope.vars}

	check := func(t *testing.T, s *celScope, expr string) []*CELVariable {
		t.Helper()
		ast, issues := s.env.Compile(expr)
		if issues.Err() != nil {
			t.Fatal(issues.Err())
		}
		checked, err := cel.AstToCheckedExpr(ast)
		if err != nil {
			t.Fatal(err)
		}
		return s.referenced(checked)
	}

	for _, test := range []struct {
		name  string
		scope *celScope
		expr  string
		want  string
	}{
		{name: "no variable", scope: scope, expr: "1 + 2", want: ""},
		{name: "only the referenced variables, in declaration order", scope: scope, expr: "c + a", want: "a,c"},
		{name: "duplicate reference", scope: scope, expr: "a + a", want: "a"},
		{name: "variable used inside a comprehension", scope: listScope, expr: "l.map(x, x + b)", want: "b"},
		{name: "same name as a comprehension variable", scope: listScope, expr: "l.exists(a, a > c)", want: "a,c"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := names(check(t, test.scope, test.expr)); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}

	t.Run("the scope keeps the variables the expression does not reference", func(t *testing.T) {
		t.Parallel()
		// An expression that references only a and c must not remove b from the scope of what follows.
		if got := names(check(t, scope, "a + c")); got != "a,c" {
			t.Fatalf("got %q", got)
		}
		if got := names(scope.vars); got != "a,b,c" {
			t.Fatalf("the scope changed: %q", got)
		}
		next, err := scope.extend("d", typ)
		if err != nil {
			t.Fatal(err)
		}
		if got := names(check(t, next, "d + b")); got != "b,d" {
			t.Fatalf("got %q, want b,d", got)
		}
	})
}
