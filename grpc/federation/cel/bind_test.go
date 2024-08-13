package cel

import (
	"context"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type testBindLib struct{}

func (lib *testBindLib) LibraryName() string {
	return "grpc.federation.test"
}

func (lib *testBindLib) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			"grpc.federation.test.foo",
			OverloadFunc("grpc_federation_test_foo_string_string",
				[]*cel.Type{cel.StringType}, cel.StringType,
				func(ctx context.Context, args ...ref.Val) ref.Val {
					return types.String("grpc_federation_test_foo_string_string")
				},
			),
			OverloadFunc("grpc_federation_test_foo_int_string",
				[]*cel.Type{cel.IntType}, cel.StringType,
				func(ctx context.Context, args ...ref.Val) ref.Val {
					return types.String("grpc_federation_test_foo_int_string")
				},
			),
		),
		BindFunction(
			"grpc.federation.test2.foo",
			OverloadFunc("grpc_federation_test2_foo_string_string",
				[]*cel.Type{cel.StringType}, cel.StringType,
				func(ctx context.Context, args ...ref.Val) ref.Val {
					return types.String("grpc_federation_test2_foo_string_string")
				},
			),
			OverloadFunc("grpc_federation_test2_foo_int_string",
				[]*cel.Type{cel.IntType}, cel.BoolType,
				func(ctx context.Context, args ...ref.Val) ref.Val {
					return types.String("grpc_federation_test2_foo_int_string")
				},
			),
		),
		BindMemberFunction(
			"foo",
			MemberOverloadFunc("foo_string_int_string",
				cel.StringType, []*cel.Type{cel.IntType}, cel.StringType,
				func(ctx context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String("foo_string_int_string")
				},
			),
			MemberOverloadFunc("foo_int_string_string",
				cel.IntType, []*cel.Type{cel.StringType}, cel.StringType,
				func(ctx context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String("foo_int_string_string")
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *testBindLib) ProgramOptions() []cel.ProgramOption {
	return nil
}

func TestBind(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		args     map[string]any
		expected string
	}{
		{
			name:     "grpc.federation.test.foo('s')",
			expr:     "grpc.federation.test.foo('s')",
			expected: "grpc_federation_test_foo_string_string",
		},
		{
			name:     "grpc.federation.test.foo(1)",
			expr:     "grpc.federation.test.foo(1)",
			expected: "grpc_federation_test_foo_int_string",
		},
		{
			name:     "grpc.federation.test2.foo('s')",
			expr:     "grpc.federation.test2.foo('s')",
			expected: "grpc_federation_test2_foo_string_string",
		},
		{
			name:     "grpc.federation.test2.foo(1)",
			expr:     "grpc.federation.test2.foo(1)",
			expected: "grpc_federation_test2_foo_int_string",
		},
		{
			name:     "10.foo('s')",
			expr:     "10.foo('s')",
			expected: "foo_int_string_string",
		},
		{
			name:     "'s'.foo(1)",
			expr:     "'s'.foo(1)",
			expected: "foo_string_int_string",
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(ContextVariableName, cel.ObjectType(ContextTypeName)),
				cel.Lib(new(testBindLib)),
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
			args := map[string]any{ContextVariableName: NewContextValue(ctx)}
			for k, v := range test.args {
				args[k] = v
			}
			out, _, err := program.Eval(args)
			if err != nil {
				t.Fatal(err)
			}
			s, ok := out.(types.String)
			if !ok {
				t.Fatalf("failed to get result: %v", out)
			}
			if string(s) != test.expected {
				t.Fatalf("failed to call function: expected %s but call %s", test.expected, string(s))
			}
		})
	}

}
