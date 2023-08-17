package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/source"
)

func RepoRoot() string {
	p, _ := filepath.Abs(filepath.Join(curDir(), "..", ".."))
	return p
}

func Compile(t *testing.T, path string) []*descriptorpb.FileDescriptorProto {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		t.Fatal(err)
	}

	opt := compiler.ImportPathOption(filepath.Join(RepoRoot(), "proto"))
	desc, err := compiler.New().Compile(context.Background(), file, opt)
	if err != nil {
		t.Fatal(err)
	}
	return desc
}

func curDir() string {
	_, file, _, _ := runtime.Caller(0) //nolint:dogsled
	return filepath.Dir(file)
}
