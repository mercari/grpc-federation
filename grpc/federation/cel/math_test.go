package cel_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/go-cmp/cmp"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestMathFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(ref.Val) error
	}{
		// math package
		{
			name: "sqrt",
			expr: "grpc.federation.math.sqrt(3.0*3.0 + 4.0*4.0)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double).Value().(float64)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				const (
					a = 3.0
					b = 4.0
				)
				expected := math.Sqrt(a*a + b*b)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "pow",
			expr: "grpc.federation.math.pow(2.0, 3.0)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double).Value().(float64)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				const (
					a = 2.0
					b = 3.0
				)
				expected := math.Pow(a, b)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "floor",
			expr: "grpc.federation.math.floor(1.51)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double).Value().(float64)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				const (
					a = 1.51
				)
				expected := math.Floor(a)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.CrossTypeNumericComparisons(true),
				cel.Lib(cellib.NewMathLibrary()),
			)
			if err != nil {
				t.Fatal(err)
			}
			ast, iss := env.Compile(test.expr)
			if iss.Err() != nil {
				t.Fatal(iss.Err())
			}
			program, err := env.Program(ast)
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]any{cellib.ContextVariableName: cellib.NewContextValue(context.Background())}
			for k, v := range test.args {
				args[k] = v
			}
			out, _, err := program.Eval(args)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
