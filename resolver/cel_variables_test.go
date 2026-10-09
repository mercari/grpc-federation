package resolver

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
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
