package cel_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/grpc/federation/cel/testdata/testpb"
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
		{
			name: "sortAsc",
			expr: "[4, 1, 3, 2].sortAsc(v, v)",
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []int64{1, 2, 3, 4}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]int64{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sortDesc",
			expr: "[4, 1, 3, 2].sortDesc(v, v)",
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []int64{4, 3, 2, 1}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]int64{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sortAsc field selection",
			expr: `[
grpc.federation.cel.test.Message{id: "a", inner: grpc.federation.cel.test.InnerMessage{id: "a2"}}, 
grpc.federation.cel.test.Message{id: "c", inner: grpc.federation.cel.test.InnerMessage{id: "c2"}}, 
grpc.federation.cel.test.Message{id: "b", inner: grpc.federation.cel.test.InnerMessage{id: "b2"}}
].sortAsc(v, v.inner.id)`,
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []*testpb.Message{
					{
						Id: "a",
						Inner: &testpb.InnerMessage{
							Id: "a2",
						},
					},
					{
						Id: "b",
						Inner: &testpb.InnerMessage{
							Id: "b2",
						},
					},
					{
						Id: "c",
						Inner: &testpb.InnerMessage{
							Id: "c2",
						},
					},
				}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]*testpb.Message{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected, cmpopts.IgnoreUnexported(testpb.Message{}, testpb.InnerMessage{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sortDesc field selection",
			expr: `[
grpc.federation.cel.test.Message{id: "a"}, 
grpc.federation.cel.test.Message{id: "c"}, 
grpc.federation.cel.test.Message{id: "b"}
].sortDesc(v, v.id)`,
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []*testpb.Message{
					{
						Id: "c",
					},
					{
						Id: "b",
					},
					{
						Id: "a",
					},
				}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]*testpb.Message{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected, cmpopts.IgnoreUnexported(testpb.Message{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sortStableAsc",
			expr: `[
grpc.federation.cel.test.Message{id:"A", num:25}, 
grpc.federation.cel.test.Message{id:"E", num:75}, 
grpc.federation.cel.test.Message{id:"A", num:75}, 
grpc.federation.cel.test.Message{id:"B", num:75},
grpc.federation.cel.test.Message{id:"A", num:75},
grpc.federation.cel.test.Message{id:"B", num:25},
grpc.federation.cel.test.Message{id:"C", num:25},
grpc.federation.cel.test.Message{id:"E", num:25},
].sortStableAsc(v, v.id).sortStableAsc(v, v.num)`,
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []*testpb.Message{
					{
						Id:  "A",
						Num: 25,
					},
					{
						Id:  "B",
						Num: 25,
					},
					{
						Id:  "C",
						Num: 25,
					},
					{
						Id:  "E",
						Num: 25,
					},
					{
						Id:  "A",
						Num: 75,
					},
					{
						Id:  "A",
						Num: 75,
					},
					{
						Id:  "B",
						Num: 75,
					},
					{
						Id:  "E",
						Num: 75,
					},
				}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]*testpb.Message{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected, cmpopts.IgnoreUnexported(testpb.Message{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sortStableDesc",
			expr: `[
grpc.federation.cel.test.Message{id:"A", num:25}, 
grpc.federation.cel.test.Message{id:"E", num:75}, 
grpc.federation.cel.test.Message{id:"A", num:75}, 
grpc.federation.cel.test.Message{id:"B", num:75},
grpc.federation.cel.test.Message{id:"A", num:75},
grpc.federation.cel.test.Message{id:"B", num:25},
grpc.federation.cel.test.Message{id:"C", num:25},
grpc.federation.cel.test.Message{id:"E", num:25},
].sortStableDesc(v, v.id).sortStableDesc(v, v.num)`,
			cmp: func(got any) error {
				lister, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := []*testpb.Message{
					{
						Id:  "E",
						Num: 75,
					},
					{
						Id:  "B",
						Num: 75,
					},
					{
						Id:  "A",
						Num: 75,
					},
					{
						Id:  "A",
						Num: 75,
					},
					{
						Id:  "E",
						Num: 25,
					},
					{
						Id:  "C",
						Num: 25,
					},
					{
						Id:  "B",
						Num: 25,
					},
					{
						Id:  "A",
						Num: 25,
					},
				}
				if lister.Size().(types.Int) != types.Int(len(expected)) {
					return fmt.Errorf("invalid size")
				}

				gotV, err := lister.ConvertToNative(reflect.TypeOf([]*testpb.Message{}))
				if err != nil {
					return fmt.Errorf("failed to convert to native: %w", err)
				}
				if diff := cmp.Diff(gotV, expected, cmpopts.IgnoreUnexported(testpb.Message{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Lib(cellib.NewListLibrary(types.DefaultTypeAdapter)),
				cel.Types(&testpb.Message{}, &testpb.InnerMessage{}),
				cel.ASTValidators(cellib.NewListValidator()),
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

func TestListValidator(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "sort int",
			expr:     `[1, 2].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)`,
			expected: ``,
		},
		{
			name:     "sort string",
			expr:     `['a', 'b'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)`,
			expected: ``,
		},
		{
			name: "sort timestamp",
			expr: `[
google.protobuf.Timestamp{seconds: 1}, 
google.protobuf.Timestamp{seconds: 2},
].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)`,
			expected: ``,
		},
		{
			name: "sort uncomparable dynamic type",
			expr: `[1, 'a'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)`,
			expected: `ERROR: <input>:1:17: list(dyn) is not comparable
 | [1, 'a'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ................^
ERROR: <input>:1:32: list(dyn) is not comparable
 | [1, 'a'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ...............................^
ERROR: <input>:1:52: list(dyn) is not comparable
 | [1, 'a'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ...................................................^
ERROR: <input>:1:73: list(dyn) is not comparable
 | [1, 'a'].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ........................................................................^`,
		},
		{
			name: "sort uncomparable struct type",
			expr: `[
grpc.federation.cel.test.Message{id: "a"}, 
grpc.federation.cel.test.Message{id: "b"}
].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)`,
			expected: `ERROR: <input>:4:10: list(grpc.federation.cel.test.Message) is not comparable
 | ].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | .........^
ERROR: <input>:4:25: list(grpc.federation.cel.test.Message) is not comparable
 | ].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ........................^
ERROR: <input>:4:45: list(grpc.federation.cel.test.Message) is not comparable
 | ].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | ............................................^
ERROR: <input>:4:66: list(grpc.federation.cel.test.Message) is not comparable
 | ].sortAsc(v, v).sortDesc(v, v).sortStableAsc(v, v).sortStableDesc(v, v)
 | .................................................................^`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Lib(cellib.NewListLibrary(types.DefaultTypeAdapter)),
				cel.Types(&testpb.Message{}, &testpb.InnerMessage{}),
				cel.ASTValidators(cellib.NewListValidator()),
			)
			if err != nil {
				t.Fatal(err)
			}
			_, iss := env.Compile(test.expr)
			if test.expected == "" {
				if iss.Err() != nil {
					t.Errorf("expected no error but got: %v", iss.Err())
				}
				return
			}
			if diff := cmp.Diff(iss.Err().Error(), test.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
