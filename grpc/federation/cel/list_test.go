package cel_test

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestList(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		{
			name: "reduce",
			expr: "[2, 3, 4].reduce(accum, cur, accum + cur, 1)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(int(gotV), 10); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "first match",
			expr: "[1, 2, 3, 4].first(v, v % 2 == 0)",
			cmp: func(got any) error {
				opt, ok := got.(*types.Optional)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotV, ok := opt.GetValue().(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", opt)
				}
				if diff := cmp.Diff(int(gotV), 2); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "first not match",
			expr: "[1, 2, 3, 4].first(v, v == 5)",
			cmp: func(got any) error {
				gotV, ok := got.(*types.Optional)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV != types.OptionalNone {
					return fmt.Errorf("invalid optional type: %v", gotV)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Lib(new(cellib.ListLibrary)),
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
			out, _, err := program.Eval(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
