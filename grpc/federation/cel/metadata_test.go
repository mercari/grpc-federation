package cel_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/metadata"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		name string
		ctx  func() context.Context
		expr string
		cmp  func(got ref.Val) error
	}{
		{
			name: "incoming with incoming context",
			ctx: func() context.Context {
				md := map[string]string{
					"key1": "value1",
					"key2": "value2",
				}
				return metadata.NewIncomingContext(context.Background(), metadata.New(md))
			},
			expr: "grpc.federation.metadata.incoming()",
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(map[ref.Val]ref.Val)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotMap := make(map[string][]string)
				for k, v := range gotV {
					gotMap[string(k.(types.String))] = v.Value().([]string)
				}
				if diff := cmp.Diff(gotMap, map[string][]string{
					"key1": {"value1"},
					"key2": {"value2"},
				}); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "incoming without incoming context",
			ctx:  context.Background,
			expr: "grpc.federation.metadata.incoming()",
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(map[ref.Val]ref.Val)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if len(gotV) != 0 {
					return fmt.Errorf("received an unexpected non-empty map: %v", gotV)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(cel.Lib(cellib.NewMetadataLibrary(test.ctx())))
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
			out, _, err := program.Eval(map[string]any{})
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
