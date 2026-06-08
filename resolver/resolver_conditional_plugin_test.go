package resolver_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/internal/testutil"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

// compileWithTestdataImports compiles a single .proto whose imports may
// reference sibling files under resolver/testdata. testutil.Compile only adds
// the repo's proto/ dir to the import path, which is enough for the standard
// federation.proto import but not for the sibling-plugin scenario this file
// covers.
func compileWithTestdataImports(t *testing.T, path string) []*descriptorpb.FileDescriptorProto {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}
	opt := compiler.ImportPathOption(
		filepath.Join(testutil.RepoRoot(), "proto"),
		filepath.Join(testutil.RepoRoot(), "resolver", "testdata"),
	)
	desc, err := compiler.New().Compile(context.Background(), file, opt)
	if err != nil {
		t.Fatal(err)
	}
	return desc
}

// TestConditionalPluginImport_RegularImport asserts that a file whose only
// reference to a plugin .proto is through a regular protobuf `import "..."`
// statement does NOT have the plugin auto-registered for CEL type-checking.
// The plugin file is loaded for type resolution but its plugin.export block
// is ignored, leaving the function unavailable in the CEL env.
func TestConditionalPluginImport_RegularImport(t *testing.T) {
	t.Parallel()
	consumerFile := filepath.Join(testutil.RepoRoot(), "resolver", "testdata", "conditional_plugin_via_regular_import.proto")
	files := compileWithTestdataImports(t, consumerFile)

	result, err := resolver.New(files).Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	consumer := findFileByName(t, result.Files, "conditional_plugin_via_regular_import.proto")
	if got := consumer.AllCELPlugins(); len(got) != 0 {
		t.Errorf("AllCELPlugins via regular-import edge: got %d plugins, want 0; first plugin: %s",
			len(got), got[0].Name)
	}
}

// TestConditionalPluginImport_FederationImport asserts that a file that
// reaches a plugin .proto via an (grpc.federation.file).import edge DOES
// auto-register the plugin — preserving backward-compatible behavior for
// the federation-import path.
func TestConditionalPluginImport_FederationImport(t *testing.T) {
	t.Parallel()
	consumerFile := filepath.Join(testutil.RepoRoot(), "resolver", "testdata", "conditional_plugin_via_federation_import.proto")
	files := compileWithTestdataImports(t, consumerFile)

	result, err := resolver.New(files, resolver.ImportPathOption(
		filepath.Join(testutil.RepoRoot(), "proto"),
		filepath.Join(testutil.RepoRoot(), "resolver", "testdata"),
	)).Resolve()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	consumer := findFileByName(t, result.Files, "conditional_plugin_via_federation_import.proto")
	plugins := consumer.AllCELPlugins()
	if len(plugins) != 1 {
		t.Fatalf("AllCELPlugins via federation-import edge: got %d, want 1", len(plugins))
	}
	if plugins[0].Name != "testplugin" {
		t.Errorf("plugin name: got %q, want %q", plugins[0].Name, "testplugin")
	}
}

func findFileByName(t *testing.T, files []*resolver.File, name string) *resolver.File {
	t.Helper()
	for _, f := range files {
		if filepath.Base(f.Name) == name {
			return f
		}
	}
	t.Fatalf("no file named %q in result; have %d files", name, len(files))
	return nil
}
