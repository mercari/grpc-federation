// Command codegen emits the grpc-federation-generated *.pb.go file for
// federation.proto with the external library extlib.NewLibrary() provided for
// the codegen-time resolver via resolver.CELLibrariesOption. This ensures that
// a CEL expression that uses the library-provided function (e.g.
// `example.ext.upper($.name)` type-checks correctly.
//
// protoc-gen-go and protoc-gen-go-grpc are run separately by buf — see
// buf.gen.yaml — because they need no library awareness.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/generator"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"

	"example/extlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	const protoPath = "proto/federation/federation.proto"

	content, err := os.ReadFile(protoPath)
	if err != nil {
		return err
	}
	srcFile, err := source.NewFile(protoPath, content)
	if err != nil {
		return err
	}
	files, err := compiler.New().Compile(
		context.Background(),
		srcFile,
		compiler.ImportPathOption("proto", "proto_deps"),
	)
	if err != nil {
		return err
	}

	r := resolver.New(files,
		resolver.ImportPathOption("proto"),
		resolver.CELLibrariesOption(extlib.NewLibrary()),
	)
	result, err := r.Resolve()
	if err != nil {
		return err
	}
	var targetFile *resolver.File
	for _, file := range result.Files {
		if len(file.Services) != 0 {
			targetFile = file
			break
		}
	}
	if targetFile == nil {
		return fmt.Errorf("federation service file not found")
	}

	out, err := generator.NewCodeGenerator().Generate(targetFile)
	if err != nil {
		return err
	}

	outPath := filepath.Join("federation", "federation_grpc_federation.pb.go")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, out, 0o600)
}
