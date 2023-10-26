package compiler_test

import (
	"context"
	"fmt"
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
	protos, err := c.Compile(ctx, file, compiler.AutoImportOption())
	if err != nil {
		t.Fatal(err)
	}

	for _, proto := range protos {
		fmt.Println(*proto.Name)
	}

	const expectedProtoNum = 8 // service.proto, post.proto, user.proto, federation.proto, google/protobuf/descriptor.proto, google/protobuf/duration.proto, google/rpc/error_details.proto, google/rpc/code.proto
	if len(protos) != expectedProtoNum {
		t.Fatalf("failed to get protos. expected %d but got %d", expectedProtoNum, len(protos))
	}
}
