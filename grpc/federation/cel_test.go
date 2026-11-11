package federation_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

	t.Run("cache index 0 is rejected wherever it is", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		err := grpcfed.PrecompileCEL(ctx, envOpts, []*grpcfed.CELPrecompileEntry{
			{Index: 1, Expr: `1`},
			{Index: 2, Expr: `undefined_variable`},
			{Index: 0, Expr: `1`},
		})
		if !errors.Is(err, grpcfed.ErrCELCacheIndex) {
			t.Fatalf("expected ErrCELCacheIndex, got %v", err)
		}
	})

	t.Run("compile many entries concurrently", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		const n = 200
		entries := make([]*grpcfed.CELPrecompileEntry, 0, n)
		for i := 1; i <= n; i++ {
			entries = append(entries, &grpcfed.CELPrecompileEntry{
				Index:     i,
				Expr:      fmt.Sprintf(`%d`, i),
				Variables: []cel.EnvOption{cel.Variable("x", cel.IntType)},
			})
		}
		if err := grpcfed.PrecompileCEL(ctx, envOpts, entries); err != nil {
			t.Fatal(err)
		}
		// Every index holds the program of its own expression.
		value := grpcfed.NewLocalValue(ctx, nil, "dummy", nil)
		for i := 1; i <= n; i++ {
			got, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
				Value:      value,
				Expr:       fmt.Sprintf(`%d`, i),
				CacheIndex: i,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got != int64(i) {
				t.Fatalf("index %d: unexpected result %v", i, got)
			}
		}
	})

	t.Run("entries without variables do not share an env", func(t *testing.T) {
		t.Parallel()
		// Naming a message type makes the type checker record the type in the env it checks against,
		// which is a data race if the entries are checked against the same env.
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		entries := make([]*grpcfed.CELPrecompileEntry, 0, 200)
		for i := 1; i <= 200; i++ {
			entries = append(entries, &grpcfed.CELPrecompileEntry{
				Index: i,
				Expr:  fmt.Sprintf(`google.rpc.Status{code: %d}.code > 0 && [google.rpc.Status{code: %d}].exists(s, s.code > 0)`, i, i),
			})
		}
		if err := grpcfed.PrecompileCEL(ctx, envOpts, entries); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("the error of the first failing entry is returned", func(t *testing.T) {
		t.Parallel()
		ctx := grpcfed.WithCELCacheMap(context.Background(), grpcfed.NewCELCacheMap())
		entries := make([]*grpcfed.CELPrecompileEntry, 0, 100)
		for i := 1; i <= 100; i++ {
			expr := `1`
			if i == 30 || i == 31 || i == 90 {
				expr = fmt.Sprintf(`undefined_%d`, i)
			}
			entries = append(entries, &grpcfed.CELPrecompileEntry{Index: i, Expr: expr})
		}
		// Which goroutine fails first varies, the error must not.
		for range 20 {
			err := grpcfed.PrecompileCEL(ctx, envOpts, entries)
			if err == nil || !strings.Contains(err.Error(), "undefined_30") {
				t.Fatalf("expected the error of undefined_30, got %v", err)
			}
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
