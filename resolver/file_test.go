package resolver_test

import (
	"testing"

	"github.com/mercari/grpc-federation/resolver"
)

func TestOutputFileResolver(t *testing.T) {
	t.Run("paths=import", func(t *testing.T) {
		r := resolver.NewOutputFilePathResolver(resolver.OutputFilePathConfig{
			Mode:        resolver.ImportMode,
			ImportPaths: []string{"proto"},
			FilePath:    "proto/buzz/buzz.proto",
		})
		path, err := r.OutputPath(&resolver.File{
			Name: "buzz.proto",
			GoPackage: &resolver.GoPackage{
				ImportPath: "example.com/project/proto/fizz",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if path != "example.com/project/proto/fizz/buzz_grpc_federation.pb.go" {
			t.Fatalf("unexpected path: %s", path)
		}
	})
	t.Run("paths=source_relative", func(t *testing.T) {
		r := resolver.NewOutputFilePathResolver(resolver.OutputFilePathConfig{
			Mode:        resolver.SourceRelativeMode,
			ImportPaths: []string{"proto"},
			FilePath:    "proto/buzz/buzz.proto",
		})
		path, err := r.OutputPath(&resolver.File{
			Name: "buzz.proto",
		})
		if err != nil {
			t.Fatal(err)
		}
		if path != "buzz/buzz_grpc_federation.pb.go" {
			t.Fatalf("unexpected path: %s", path)
		}
	})
	t.Run("module=prefix", func(t *testing.T) {
		r := resolver.NewOutputFilePathResolver(resolver.OutputFilePathConfig{
			Mode:        resolver.ModulePrefixMode,
			ImportPaths: []string{"proto"},
			FilePath:    "proto/buzz/buzz.proto",
			Prefix:      "example.com/project",
		})
		path, err := r.OutputPath(&resolver.File{
			Name: "buzz.proto",
			GoPackage: &resolver.GoPackage{
				ImportPath: "example.com/project/proto/fizz",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if path != "proto/fizz/buzz_grpc_federation.pb.go" {
			t.Fatalf("unexpected path: %s", path)
		}
	})
}
