package resolver_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
)

// shoutLibrary registers org.example.shout(string) string as a Go-native
// CEL library, modeled on grpc/federation/cel/list.go.
type shoutLibrary struct{}

func (shoutLibrary) LibraryName() string { return "org.example.shout" }

func (shoutLibrary) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("org.example.shout",
			cel.Overload("org_example_shout_string_string",
				[]*cel.Type{cel.StringType}, cel.StringType,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					s, ok := arg.Value().(string)
					if !ok {
						return types.NewErr("shout: expected string, got %T", arg.Value())
					}
					return types.String(strings.ToUpper(s))
				}),
			),
		),
	}
}

func (shoutLibrary) ProgramOptions() []cel.ProgramOption { return nil }

// TestCELLibrariesOption asserts that a library registered via
// resolver.CELLibrariesOption is visible to the codegen-time CEL env, so a
// DSL expression that calls into the library passes type-checking.
func TestCELLibrariesOption(t *testing.T) {
	t.Parallel()
	fileName := filepath.Join(testutil.RepoRoot(), "resolver", "testdata", "cel_library.proto")
	files := testutil.Compile(t, fileName)

	r := resolver.New(files, resolver.CELLibrariesOption(shoutLibrary{}))
	result, err := r.Resolve()
	if err != nil {
		t.Fatalf("expected resolve to succeed with library registered, got: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}
}
