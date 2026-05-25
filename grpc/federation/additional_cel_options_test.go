package federation_test

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

// TestAdditionalCELOptions_OverridesBindingByOverloadID verifies the design
// assumption underlying the AdditionalCELOptions hook on the generated service
// config: when a CEL EnvOption registers a cel.Function with the same name and
// the same overload ID and matching signature as a previously registered
// function, the later binding replaces the earlier one.
//
// In the generator, AdditionalCELOptions are appended to celEnvOpts AFTER any
// CEL plugin (WASM) bindings, so callers can inject native Go bindings to
// override plugin-registered functions for performance-critical CEL functions
// (e.g. pure in-memory operations that should not pay the WASM dispatch +
// per-process mutex serialization cost in CELPluginInstance.Call).
//
// This test exercises that semantics with cel-go directly so that we have a
// regression guard independent of the WASM plugin runtime.
func TestAdditionalCELOptions_OverridesBindingByOverloadID(t *testing.T) {
	t.Parallel()

	const (
		fnName     = "example_fn"
		overloadID = "example_fn_string"
	)

	// First binding ("plugin"): returns "from_plugin".
	pluginOpt := cel.Function(fnName,
		cel.Overload(overloadID, []*cel.Type{}, cel.StringType,
			cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
				return types.String("from_plugin")
			}),
		),
	)

	// Second binding ("native override"): same overload ID + signature, returns "from_native".
	nativeOpt := cel.Function(fnName,
		cel.Overload(overloadID, []*cel.Type{}, cel.StringType,
			cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
				return types.String("from_native")
			}),
		),
	)

	// Build env in the same order generator emits: plugin first, then AdditionalCELOptions.
	envOpts := []grpcfed.CELEnvOption{pluginOpt, nativeOpt}
	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		t.Fatalf("failed to build cel env: %v", err)
	}

	ast, iss := env.Compile(fnName + "()")
	if iss != nil && iss.Err() != nil {
		t.Fatalf("compile failed: %v", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatalf("program failed: %v", err)
	}
	out, _, err := prg.Eval(map[string]any{})
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	got, ok := out.Value().(string)
	if !ok {
		t.Fatalf("unexpected result type: %T", out.Value())
	}
	if got != "from_native" {
		t.Fatalf("expected later AddOverload binding to win (from_native) but got %q", got)
	}
}

// TestAdditionalCELOptions_AppendOrderEquivalence verifies that appending
// AdditionalCELOptions to a pre-built celEnvOpts slice (the operation the
// generator emits in NewXxxService) preserves the override-by-later-binding
// behavior relied on by callers.
func TestAdditionalCELOptions_AppendOrderEquivalence(t *testing.T) {
	t.Parallel()

	const (
		fnName     = "example_fn2"
		overloadID = "example_fn2_string"
	)
	pluginOpt := cel.Function(fnName,
		cel.Overload(overloadID, []*cel.Type{}, cel.StringType,
			cel.FunctionBinding(func(_ ...ref.Val) ref.Val { return types.String("plugin") }),
		),
	)
	nativeOpt := cel.Function(fnName,
		cel.Overload(overloadID, []*cel.Type{}, cel.StringType,
			cel.FunctionBinding(func(_ ...ref.Val) ref.Val { return types.String("native") }),
		),
	)

	// Mimic generator: build base celEnvOpts, then append AdditionalCELOptions.
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, pluginOpt)
	additional := []grpcfed.CELEnvOption{nativeOpt}
	celEnvOpts = append(celEnvOpts, additional...)

	env, err := cel.NewEnv(celEnvOpts...)
	if err != nil {
		t.Fatalf("failed to build cel env: %v", err)
	}
	ast, iss := env.Compile(fnName + "()")
	if iss != nil && iss.Err() != nil {
		t.Fatalf("compile failed: %v", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatalf("program failed: %v", err)
	}
	out, _, err := prg.Eval(map[string]any{})
	if err != nil {
		t.Fatalf("eval failed: %v", err)
	}
	if got := out.Value().(string); got != "native" {
		t.Fatalf("expected appended option to override base binding (native), got %q", got)
	}

	// Also verify the no-op case: nil AdditionalCELOptions preserves base binding.
	var baseOnly []grpcfed.CELEnvOption
	baseOnly = append(baseOnly, pluginOpt)
	var nilAdditional []grpcfed.CELEnvOption
	baseOnly = append(baseOnly, nilAdditional...)
	env2, err := cel.NewEnv(baseOnly...)
	if err != nil {
		t.Fatalf("failed to build base-only env: %v", err)
	}
	ast2, iss := env2.Compile(fnName + "()")
	if iss != nil && iss.Err() != nil {
		t.Fatalf("compile (base-only) failed: %v", iss.Err())
	}
	prg2, err := env2.Program(ast2)
	if err != nil {
		t.Fatalf("program (base-only) failed: %v", err)
	}
	out2, _, err := prg2.Eval(map[string]any{})
	if err != nil {
		t.Fatalf("eval (base-only) failed: %v", err)
	}
	if got := out2.Value().(string); got != "plugin" {
		t.Fatalf("expected base binding (plugin) when AdditionalCELOptions is nil, got %q", got)
	}
}
