package compiler_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/source"
)

func TestCompiler(t *testing.T) {
	ctx := context.Background()

	path := filepath.Join("testdata", "service.proto")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}
	c := compiler.New()
	protos, err := c.Compile(ctx, file)
	if err != nil {
		t.Fatal(err)
	}

	const expectedProtoNum = 5 // service.proto, post.proto, user.proto, federation.proto, google/protobuf/descriptor.proto
	if len(protos) != expectedProtoNum {
		t.Fatalf("failed to get protos. expected %d but got %d", expectedProtoNum, len(protos))
	}
}
