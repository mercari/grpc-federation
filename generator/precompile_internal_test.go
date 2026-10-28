package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

func TestReferencedCELVariables(t *testing.T) {
	t.Parallel()

	env, err := cel.NewEnv(
		ext.TwoVarComprehensions(),
		cel.Variable("a", cel.IntType),
		cel.Variable("b", cel.StringType),
		cel.Variable("xs", cel.ListType(cel.IntType)),
		cel.Variable("error", cel.StringType),
		cel.Variable("grpc.federation.var", cel.StringType),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		expr string
		want []string
	}{
		{name: "no variable", expr: `1 + 2`, want: nil},
		{name: "sorted by name", expr: `b + string(a)`, want: []string{"a:int", "b:string"}},
		{name: "duplicate reference", expr: `a + a`, want: []string{"a:int"}},
		{name: "comprehension variables are not declared", expr: `xs.filter(v, v > a)`, want: []string{"a:int", "xs:list(int)"}},
		{name: "two variable comprehension", expr: `xs.all(i, v, i < v)`, want: []string{"xs:list(int)"}},
		{name: "same name inside and outside of comprehension", expr: `a + xs.map(a, a)[0]`, want: []string{"a:int", "xs:list(int)"}},
		{name: "message literal is not a variable", expr: `google.protobuf.Int64Value{value: a}`, want: []string{"a:int"}},
		{name: "service-wide variables are excluded", expr: `error + grpc.federation.var + b`, want: []string{"b:string"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ast, issues := env.Compile(test.expr)
			if issues.Err() != nil {
				t.Fatal(issues.Err())
			}
			checked, err := cel.AstToCheckedExpr(ast)
			if err != nil {
				t.Fatal(err)
			}
			vars, err := referencedCELVariables(checked)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, v := range vars {
				got = append(got, v.Name+":"+v.Type.String())
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %v but want %v", got, test.want)
			}
			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("got %v but want %v", got, test.want)
				}
			}
		})
	}
}

func TestCELIndexerMissedExpression(t *testing.T) {
	t.Parallel()

	testdataDir := filepath.Join(testutil.RepoRoot(), "testdata")
	files := testutil.Compile(t, filepath.Join(testdataDir, "simple_aggregation.proto"))
	result, err := resolver.New(files, resolver.ImportPathOption(testdataDir)).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	file := result.Files[0]
	file.Name = filepath.Base(file.Name)

	generate := func(f *File) error {
		tmpl, err := loadTemplate()
		if err != nil {
			t.Fatal(err)
		}
		_, err = generateGoContent(tmpl, f)
		return err
	}

	t.Run("every expression is reached", func(t *testing.T) {
		t.Parallel()
		if err := generate(NewFile(file)); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("an expression the walker did not reach fails generation", func(t *testing.T) {
		t.Parallel()
		f := NewFile(file)
		indexer, err := f.celIndexer()
		if err != nil {
			t.Fatal(err)
		}
		// Forget the first expression, as if the walker had not visited it.
		var missed *resolver.CELValue
		for value, idx := range indexer.values {
			if idx == 1 {
				missed = value
			}
		}
		if missed == nil {
			t.Fatal("no expression was indexed")
		}
		delete(indexer.values, missed)

		err = generate(f)
		if err == nil {
			t.Fatal("expected generation to fail")
		}
		if !strings.Contains(err.Error(), "not reachable") || !strings.Contains(err.Error(), missed.Expr) {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
