package cel_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestRegexp(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		{
			name: "compile",
			expr: "grpc.federation.regexp.compile('a+b')",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Regexp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.MustCompile("a+b")
				if diff := cmp.Diff(gotV.Regexp.String(), expected.String()); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "mustCompile",
			expr: "grpc.federation.regexp.mustCompile('a+b')",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Regexp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.MustCompile("a+b")
				if diff := cmp.Diff(gotV.Regexp.String(), expected.String()); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteMeta",
			expr: "grpc.federation.regexp.quoteMeta('[a-z]+\\\\d')",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.QuoteMeta("\\w+\\d[0-5]")
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "findStringSubmatch",
			expr: "grpc.federation.regexp.mustCompile('([a-z]+)\\\\d(\\\\d)').findStringSubmatch('abc123')",
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotV, err := lister.ConvertToNative(reflect.TypeOf([]string{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				expected := regexp.MustCompile(`(\w+)\d(\d)`).FindStringSubmatch("abc123")
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "matchString",
			expr: "grpc.federation.regexp.mustCompile('[a-z]+\\\\d\\\\d').matchString('abc123')",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.MustCompile(`\w+\d\d`).MatchString("abc123")
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "replaceAllString",
			expr: "grpc.federation.regexp.mustCompile('mackerel').replaceAllString('mackerel is tasty', 'salmon')",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.MustCompile(`mackerel`).ReplaceAllString("mackerel is tasty", "salmon")
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "string",
			expr: "grpc.federation.regexp.mustCompile('[a-z]+\\\\d\\\\d').string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := regexp.MustCompile(`\w+\d\d`).String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
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
				cel.Lib(new(cellib.RegexpLibrary)),
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
