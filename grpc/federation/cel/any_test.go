package cel_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestAnyFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		err  bool
		cmp  func(ref.Val) error
	}{
		{
			name: "new_with_empty",
			expr: `grpc.federation.any.new(google.protobuf.Empty{})`,
			cmp: func(got ref.Val) error {
				expected, err := anypb.New(&emptypb.Empty{})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(
					got.Value(),
					expected,
					cmpopts.IgnoreUnexported(anypb.Any{}, emptypb.Empty{}),
				); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "new_with_timestamp",
			expr: `grpc.federation.any.new(google.protobuf.Timestamp{})`,
			cmp: func(got ref.Val) error {
				expected, err := anypb.New(&timestamppb.Timestamp{})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(
					got.Value(),
					expected,
					cmpopts.IgnoreUnexported(anypb.Any{}, timestamppb.Timestamp{}),
				); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "new_with_not_proto_message",
			expr: `grpc.federation.any.new(1)`,
			err:  true,
		},
	}

	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Lib(cellib.NewAnyLibrary(reg)),
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
			if test.err {
				if err == nil {
					t.Fatal("expected error")
				}
			} else if err != nil {
				t.Fatal(err)
			} else {
				if err := test.cmp(out); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
