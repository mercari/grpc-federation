package cel_test

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/grpc/federation/cel/testdata/testpb"
)

func TestCast(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		msg      any
		expected ref.Val
	}{
		{
			name:     "cast nil",
			expr:     "grpc.federation.cast.null_value(msg) == null",
			msg:      nil,
			expected: types.True,
		},
		{
			name:     "cast typed-nil",
			expr:     "grpc.federation.cast.null_value(msg) == null",
			msg:      (*testpb.Message)(nil),
			expected: types.True,
		},
		{
			name:     "cast struct pointer",
			expr:     "grpc.federation.cast.null_value(msg) == null",
			msg:      &testpb.Message{},
			expected: types.False,
		},
		{
			name:     "cast lit",
			expr:     "grpc.federation.cast.null_value(msg) == null",
			msg:      100,
			expected: types.False,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Types(&testpb.Message{}),
				cel.Lib(new(cellib.CastLibrary)),
				cel.Variable("msg", cel.DynType),
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
			out, _, err := program.Eval(map[string]any{
				"msg": test.msg,
			})
			if err != nil {
				t.Fatal(err)
			}
			if out.Equal(test.expected) == types.False {
				t.Fatalf("unexpected value: want %v but got %v", test.expected, out)
			}
		})
	}
}
