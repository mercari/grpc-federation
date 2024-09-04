package cel_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(ref.Val) error
	}{
		{
			name: "parse",
			expr: "grpc.federation.url.parse('https://example.com/path?query=1#fragment')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				expected, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotURL, expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}

	reg, err := types.NewRegistry(new(cellib.URL), new(cellib.Userinfo))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Lib(cellib.NewURLLibrary(reg)),
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
