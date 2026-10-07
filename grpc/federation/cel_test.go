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
		cacheMap := grpcfed.NewCELCacheMap()
		ctx := grpcfed.WithCELCacheMap(context.Background(), cacheMap)
		entries := []*grpcfed.CELPrecompileEntry{
			{CacheIndex: 1, Expr: `1 + 2`},
			{
				CacheIndex: 2,
				Expr:       `x * 2`,
				Variables:  []cel.EnvOption{cel.Variable("x", cel.IntType)},
			},
			{
				CacheIndex: 3,
				Expr:       `name == 'foo'`,
				Variables:  []cel.EnvOption{cel.Variable("name", cel.StringType)},
			},
		}
		if err := grpcfed.PrecompileCEL(ctx, envOpts, entries); err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !cacheMap.Has(entry.CacheIndex) {
				t.Errorf("cache index %d is not precompiled", entry.CacheIndex)
			}
		}
		if cacheMap.Has(4) {
			t.Error("unexpected cache entry")
		}
	})

	t.Run("compile error", func(t *testing.T) {
		t.Parallel()
		cacheMap := grpcfed.NewCELCacheMap()
		ctx := grpcfed.WithCELCacheMap(context.Background(), cacheMap)
		err := grpcfed.PrecompileCEL(ctx, envOpts, []*grpcfed.CELPrecompileEntry{
			{CacheIndex: 1, Expr: `undefined_variable + 1`},
		})
		if err == nil {
			t.Fatal("expected compile error")
		}
	})

	t.Run("invalid cache index", func(t *testing.T) {
		t.Parallel()
		cacheMap := grpcfed.NewCELCacheMap()
		ctx := grpcfed.WithCELCacheMap(context.Background(), cacheMap)
		err := grpcfed.PrecompileCEL(ctx, envOpts, []*grpcfed.CELPrecompileEntry{
			{CacheIndex: 0, Expr: `1`},
		})
		if err == nil {
			t.Fatal("expected error for cache index 0")
		}
	})

	t.Run("cache map is required", func(t *testing.T) {
		t.Parallel()
		err := grpcfed.PrecompileCEL(context.Background(), envOpts, []*grpcfed.CELPrecompileEntry{
			{CacheIndex: 1, Expr: `1`},
		})
		if err == nil {
			t.Fatal("expected error when cache map is missing")
		}
	})
}
