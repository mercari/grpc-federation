# How to run your own code generation process

gRPC Federation supports code generation plugin system for running your own code generation process.

This document is based on [_examples/16_code_gen_plugin](./../_examples/16_code_gen_plugin/). If you would like to see a working sample, please refer it.

## 1. Write code generation process

```go
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/mercari/grpc-federation/grpc/federation/generator"
)

//go:embed resolver.go.tmpl
var tmpl []byte

func main() {
	req, err := generator.ToCodeGeneratorRequest(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range req.GRPCFederationFiles {
		for _, svc := range file.Services {
			fmt.Println("service name", svc.Name)
			for _, method := range svc.Methods {
				fmt.Println("method name", method.Name)
				if method.Rule != nil {
					fmt.Println("method timeout", method.Rule.Timeout)
				}
			}
		}
		for _, msg := range file.Messages {
			fmt.Println("msg name", msg.Name)
		}
	}
	if err := os.WriteFile("resolver_test.go", tmpl, 0o600); err != nil {
		log.Fatal(err)
	}
}
```

You can use `embed` library as you would normally write a Go program. Also, since `os.Stdin` is the value of encode [generator.proto](../proto/grpc/federation/generator.proto), you can use `generator.ToCodeGeneratorRequest` API to decode it.

`generator.CodeGeneratorRequest` contains information from the gRPC Federation's analyzing results of the proto file. Therefore, it can be used to implement code generation process you want.

> [!NOTE]
> The relative path is the directory where the code generation was performed.

## 2. Compile plugin to WebAssembly

```console
$ GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm ./cmd/plugin
```

## 3. Calculates sha256 value for the WebAssembly file

```console
$ sha256sum plugin.wasm
d010eb2cb5ad1c95d428bfea50dcce918619f4760230588e8c3df95635c992fe  plugin.wasm
```

## 4. Add `plugins` option

If you are using `Buf`, add an option in the format `plugins=<url>:<sha256>` to the `grpc-federation` plugin option.

- `<url>`: `http://`, `https://`, `file://` schemes are supported.
  - `http://` and `https://`: Download and use the specified wasm file.
  - `file://`: Refer to a file on the local file system. If you start with `/`, the path expected an absolute path.

- `<sha256>`: The content hash of the wasm file. It is used for validation of the wasm file.

- buf.gen.yaml

```yaml
version: v1
managed:
  enabled: true
plugins:
  - plugin: go
    out: .
    opt: paths=source_relative
  - plugin: go-grpc
    out: .
    opt: paths=source_relative
  - plugin: grpc-federation
    out: .
    opt:
      - paths=source_relative
      - plugins=file://plugin.wasm:d010eb2cb5ad1c95d428bfea50dcce918619f4760230588e8c3df95635c992fe
```

If you use `grpc-federation.yaml` for code generation, please describe it in the same way.

- grpc-federation.yaml

```yaml
imports:
  - proto
src:
  - proto
out: .
plugins:
  - plugin: go
    opt: paths=source_relative
  - plugin: go-grpc
    opt: paths=source_relative
  - plugin: grpc-federation
    opt:
      - paths=source_relative
      - plugins=file://plugin.wasm:d010eb2cb5ad1c95d428bfea50dcce918619f4760230588e8c3df95635c992fe
```

## 5. Run code generator

Run the gRPC Federation code generator.

```console
buf generate
```

or 

```console
grpc-federation-generator ./proto/federation/federation.proto
```

## 6. Run your code generator

The plugin is executed and the `resover_test.go` file is created in the current working directory.