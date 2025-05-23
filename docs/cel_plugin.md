# How to extend the CEL API

This document explains how you can add your own logic as a CEL API and make it available to the gRPC Federation.

We call this mechanism the CEL Plugin.

This document is based on [_examples/15_cel_plugin](./../_examples/15_cel_plugin/). If you would like to see a working sample, please refer to that document.

## 1. Write API definitions in ProtocolBuffers

First, define in ProtocolBuffers the types, functions you want to provide.

```proto
syntax = "proto3";

package example.regexp;

import "grpc/federation/federation.proto";

option go_package = "example/plugin;pluginpb";

message Regexp {
  uint64 ptr = 1; // store raw pointer value.
}

option (grpc.federation.file).plugin.export = {
  name: "regexp"
  types: [
    {
      name: "Regexp"
      methods: [
        {
          name: "matchString"
          args { type { kind: STRING } }
          return { kind: BOOL }
        }
      ]
    }
  ]
  functions: [
    {
      name: "compile"
      args { type { kind: STRING } }
      return { message: "Regexp" }
    }
  ]
};
```

`(grpc.federation.file).plugin.export` option is used to define the API. In this example, the plugin is named `regexp`.

The `regexp` plugin belongs to the `example.regexp` package. Also, this provides the `Regexp` message type and makes `matchString` available as a method on the `Regexp` message, and `compile` function is also added.

In summary, it contains the following definitions.

- `example.regexp.Regexp` message
- `example.regexp.Regexp.matchString(string) bool` method
- `example.regexp.compile(string) example.regexp.Regexp` function

Put this definition in `plugin/plugin.proto` .

## 2. Use the defined CEL API in your BFF

Next, write the code to use the defined API.  
The following definition is a sample of adding a gRPC method that compiles the `expr` specified by the `example.regexp.compile` function and returns whether the contents of `target` matches the result.

```proto
syntax = "proto3";

package org.federation;

import "grpc/federation/federation.proto";
import "plugin/plugin.proto"; // your CEL API definition

option go_package = "example/federation;federation";

service FederationService {
  option (grpc.federation.service) = {};
  rpc IsMatch(IsMatchRequest) returns (IsMatchResponse) {};
}

message IsMatchRequest {
  string expr = 1;
  string target = 2;
}
    
message IsMatchResponse {
  option (grpc.federation.message) = {
    def {
      name: "matched"

      // Use the defined CEL API.
      by: "example.regexp.compile($.expr).matchString($.target)"
    }
  };
  bool result = 1 [(grpc.federation.field).by = "matched"];
}
```

## 3. Run code generator

Run the gRPC Federation code generator to generate Go language code. At this time, CEL API's schema is verified by the gRPC Federation Compiler, and an error is output if it is used incorrectly.

## 4. Write CEL Plugin code

By the code generation, the library code is generated to write the plugin. In this example, the output is in `plugin/plugin_grpc_federation.pb.go`.

Using that library, write a plugin as follows.

```go
package main

import (
	pluginpb "example/plugin"
	"regexp"
	"unsafe"
)

var _ pluginpb.RegexpPlugin = new(plugin)

type plugin struct{}

// For example.regexp.compile function.
func (_ *plugin) Example_Regexp_Compile(ctx context.Context, expr string) (*pluginpb.Regexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &pluginpb.Regexp{
		Ptr: uint32(uintptr(unsafe.Pointer(re))),
	}, nil
}

// For example.regexp.Regexp.matchString method.
func (_ *plugin) Example_Regexp_Regexp_MatchString(ctx context.Context, re *pluginpb.Regexp, s string) (bool, error) {
	return (*regexp.Regexp)(unsafe.Pointer(uintptr(re.Ptr))).MatchString(s), nil
}

func main() {
	pluginpb.RegisterRegexpPlugin(new(plugin))
}
```

`pluginpb.RegisterRegexpPlugin` function for registering developed plugins. Also, `pluginpb.RegexpPlugin` is interface type for plugin.

## 5. Compile plugin to WebAssembly

```console
$ GOOS=wasip1 GOARCH=wasm go build -o regexp.wasm ./cmd/plugin
```

## 6. Calculates sha256 value for the WebAssembly file

```console
$ sha256sum regexp.wasm
820f86011519c42da0fe9876bc2ca7fbee5df746acf104d9e2b9bba802ddd2b9  regexp.wasm
```

## 7. Load plugin ( WebAssembly ) file

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

## Resource Capability

By default, plugins do not have access to Environment variables, File System, or Network. This helps to run plugins securely, but there may be cases where you want to allow access to these resources.
In such cases, you can enable access to each resource by configuring as follows:

```proto
option (grpc.federation.file).plugin.export = {
  name: "regexp"
  capability {
    network: {} // enable network access via HTTP/HTTPS
    env { names: ["foo"] } // enable access to `FOO` environment variable
    file_system { mount_path: "/" } // enable access to file system from mount point `/`
  }
}
```

For more details, please refer to the sample that uses this configuration in [_examples/21_wasm_net](../_examples/21_wasm_net/).

# How it works

Host and wasm plugin using stdio to exchange data.
The exchanged data types are `CELPluginRequest` and `CELPluginResponse` defined on [plugin.proto](https://github.com/mercari/grpc-federation/blob/main/proto/grpc/federation/plugin.proto). Actually, the data is converted to json and then exchanged.
