# CEL Plugin Example

## 1. Write plugin definition in ProtocolBuffers

The plugin feature allows you to define global functions, methods, and variables.

[plugin.proto](./proto/plugin/plugin.proto)

## 2. Run code generator

Run `make generate` to generate code by `protoc-gen-grpc-federation`.

## 3. Write plugin code

[cmd/plugin/main.go](./cmd/plugin/main.go) is the plugin source.

The helper code needed to write the plugin is already generated, so it is used.

[The helper is here](./plugin/plugin_grpc_federation.pb.go).

In `plugin/plugin_grpc_federation.pb.go`, there is a `RegisterRegexpPlugin` function required for plugin registration and a `RegexpPlugin` interface required for the plugin.
The code for the plugin is as follows.

```go
package main

import (
	pluginpb "example/plugin"
	"regexp"
	"unsafe"
)

type plugin struct{}

func (_ *plugin) Val() string {
	return "hello grpc-federation plugin"
}

func (_ *plugin) Example_Regexp_Compile(expr string) (*pluginpb.Regexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &pluginpb.Regexp{
		Ptr: uint64(uintptr(unsafe.Pointer(re))),
	}, nil
}

func (_ *plugin) Example_Regexp_Regexp_MatchString(re *pluginpb.Regexp, s string) (bool, error) {
	return (*regexp.Regexp)(unsafe.Pointer(uintptr(re.Ptr))).MatchString(s), nil
}

func main() {
	pluginpb.RegisterRegexpPlugin(&plugin{})
}
```

## 4. Compile a plugin to WebAssembly

Run `make build/wasm` to compile to WebAssembly.
( `regexp.wasm` is output )

## 5. Calculates sha256 value for the WebAssembly file

```console
$ sha256sum regexp.wasm
820f86011519c42da0fe9876bc2ca7fbee5df746acf104d9e2b9bba802ddd2b9  regexp.wasm
```

## 6. Load plugin ( WebAssembly ) file

Initialize the gRPC server with the path to the wasm file and the sha256 value of the file.

```go
federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
	CELPlugin: &federation.FederationServiceCELPluginConfig{
		Regexp: federation.FederationServiceCELPluginWasmConfig{
			Path:   "regexp.wasm",
			Sha256: "820f86011519c42da0fe9876bc2ca7fbee5df746acf104d9e2b9bba802ddd2b9",
		},
	},
})
```

## 7. Plugin usage

You can evaluate the plugin's API in the same way you would evaluate a regular usage (e.g. `by` feature).

```proto
message IsMatchResponse {
  option (grpc.federation.message) = {
    def {
      name: "matched"
      by: "example.regexp.compile($.expr).matchString($.target)"
    }
  };
  bool result = 1 [(grpc.federation.field).by = "matched"];
}
```
