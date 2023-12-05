package cel_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestUUID(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		// UUID functions
		{
			name: "new",
			expr: "grpc.federation.uuid.new()",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.UUID) //nolint: staticcheck
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if len(gotV.UUID) == 0 {
					return fmt.Errorf("failed to get uuid")
				}
				return nil
			},
		},
		{
			name: "newRandom",
			expr: "grpc.federation.uuid.newRandom()",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.UUID) //nolint: staticcheck
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if len(gotV.UUID) == 0 {
					return fmt.Errorf("failed to get uuid")
				}
				return nil
			},
		},
		{
			name: "newRandomFromRand",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix())))",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.UUID)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.String()
				if diff := cmp.Diff(gotV.UUID.String(), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "domain",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))).domain()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.Domain()
				if diff := cmp.Diff(uuid.Domain(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "id",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))).id()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.ID()
				if diff := cmp.Diff(uint32(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "time",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))).time()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.Time()
				if diff := cmp.Diff(uuid.Time(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "urn",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))).urn()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.URN()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "string",
			expr: "grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(grpc.federation.time.date(2023, 12, 25, 12, 0, 0, 0, grpc.federation.time.UTC).unix()))).string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				r := rand.New(rand.NewSource(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC).Unix())) //nolint: gosec
				id, err := uuid.NewRandomFromReader(r)
				if err != nil {
					return err
				}
				expected := id.String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "version",
			expr: "grpc.federation.uuid.new().version()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := uuid.New().Version()
				if diff := cmp.Diff(uuid.Version(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Lib(new(cellib.UUIDLibrary)),
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
