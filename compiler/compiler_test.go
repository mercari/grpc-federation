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

	const expectedProtoNum = 10 // service.proto, post.proto, user.proto, federation.proto, private.proto, google/protobuf/descriptor.proto, google/protobuf/duration.proto, google/rpc/error_details.proto, google/rpc/code.proto
	if len(protos) != expectedProtoNum {
		t.Fatalf("failed to get protos. expected %d but got %d", expectedProtoNum, len(protos))
	}
}
