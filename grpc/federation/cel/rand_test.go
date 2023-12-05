package cel_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestRand(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		// Source functions
		{
			name: "newSource",
			expr: "grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Source)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix()).Int63()
				if diff := cmp.Diff(gotV.Int63(), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int63",
			expr: "grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()).int63()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix()).Int63()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "seed",
			expr: "grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()).seed(10)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(bool(gotV), true); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// Rand functions
		{
			name: "new",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Rand)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int()
				if diff := cmp.Diff(int(gotV.Int()), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "expFloat64",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).expFloat64()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).ExpFloat64()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "float32",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).float32()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Float32()
				if diff := cmp.Diff(float32(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "float64",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).float64()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Float64()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).int()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int31",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).int31()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int31()
				if diff := cmp.Diff(int32(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int31n",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).int31n(10)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int31n(10)
				if diff := cmp.Diff(int32(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "rand_int63",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).int63()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int63()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int63n",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).int63n(10)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Int63n(10)
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "intn",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).intn(10)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Intn(10)
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "normFloat64",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).normFloat64()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).NormFloat64()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "uint32",
			expr: "grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())).uint32()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())).Uint32()
				if diff := cmp.Diff(uint32(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Lib(new(cellib.RandLibrary)),
				cel.Lib(new(cellib.TimeLibrary)),
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
