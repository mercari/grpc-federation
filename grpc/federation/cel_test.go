package federation_test

import (
	"context"
	"testing"

	"github.com/google/cel-go/cel"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

func TestPrecompileCEL(t *testing.T) {
	t.Parallel()

	celHelper := grpcfed.NewCELTypeHelper("test", grpcfed.CELTypeHelperFieldMap{})
	envOpts := grpcfed.NewDefaultEnvOptions(celHelper)

	t.Run("compile all entries", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		entries := []*grpcfed.CELPrecompileEntry{
			{Index: 1, Expr: `1 + 2`},
			{
				Index:     2,
				Expr:      `x * 2`,
				Variables: []cel.EnvOption{cel.Variable("x", cel.IntType)},
			},
			{
				Index:     3,
				Expr:      `$ == 'foo'`,
				Variables: []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.StringType)},
			},
		}
		if err := grpcfed.PrecompileCEL(ctx, envOpts, entries); err != nil {
			t.Fatal(err)
		}
		// A precompiled expression is evaluated without compiling it again,
		// so evaluating it with an env that lacks its variable must still work.
		value := grpcfed.NewLocalValue(ctx, nil, "dummy", nil)
		got, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
			Value:      value,
			Expr:       `1 + 2`,
			CacheIndex: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != int64(3) {
			t.Fatalf("unexpected result: %v", got)
		}
	})

	t.Run("compile error", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		err := grpcfed.PrecompileCEL(ctx, envOpts, []*grpcfed.CELPrecompileEntry{
			{Index: 1, Expr: `undefined_variable + 1`},
		})
		if err == nil {
			t.Fatal("expected compile error")
		}
	})

	t.Run("invalid cache index", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		err := grpcfed.PrecompileCEL(ctx, envOpts, []*grpcfed.CELPrecompileEntry{
			{Index: 0, Expr: `1`},
		})
		if err == nil {
			t.Fatal("expected error for cache index 0")
		}
	})

	t.Run("cache map is required", func(t *testing.T) {
		t.Parallel()
		err := grpcfed.PrecompileCEL(context.Background(), envOpts, []*grpcfed.CELPrecompileEntry{
			{Index: 1, Expr: `1`},
		})
		if err == nil {
			t.Fatal("expected error when cache map is missing")
		}
	})
}
