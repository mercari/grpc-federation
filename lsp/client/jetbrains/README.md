# gRPC Federation Language Server for JetBrains

<!-- Plugin description -->
## Features

### Goto Definition

It supports jumps to the following definition sources.

- Imported file name
    - e.g.) import "<u>path/to/foo.proto</u>"

- Type of field definition
    - e.g.) <u>Foo</u> foo = 1

- gRPC method name
    - e.g.) def { call { method: "<u>foopkg.FooService/BarMethod</u>" } }

- Alias name
    - e.g.) option (grpc.federation.message).alias = "<u>foopkg.Bar</u>"

### Diagnostics
### Code Completion
<!-- Plugin description end -->

## Installation
To use the plugin, `grpc-federation-language-server` must be present in the path. If Go is available, run the following command to install it.

```console
$ go install github.com/mercari/grpc-federation/cmd/grpc-federation-language-server@latest
```

Then install the plugin from the Marketplace.

`Settings/Preferences` > `Plugins` > `Marketplace` > `Search for "gRPC Federation"` > `Install`

