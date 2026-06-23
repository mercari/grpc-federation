// Command lint validates federation.proto with the external library
// extlib.NewLibrary() provided via validator.CELLibrariesOption. This is
// the lint-time analog of cmd/codegen: a CEL expression that uses the
// library-provided function (`example.ext.upper($.name)`) type-checks
// correctly only when the library is registered. The shared binary
// grpc-federation-linter cannot load CEL libraries itself, so
// an additional linter is required.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"

	"example/extlib"
)

func main() {
	const protoPath = "proto/federation/federation.proto"

	content, err := os.ReadFile(protoPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	srcFile, err := source.NewFile(protoPath, content)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	outs := validator.New().Validate(context.Background(), srcFile,
		validator.ImportPathOption("proto", "proto_deps"),
		validator.CELLibrariesOption(extlib.NewLibrary()),
	)
	if len(outs) == 0 {
		return
	}
	fmt.Println(validator.Format(outs))
	if validator.ExistsError(outs) {
		os.Exit(1)
	}
}
