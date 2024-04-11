<!-- Plugin description -->
# gRPC Federation Language Server for JetBrains
The gRPC Federation Language Server for JetBrains is a plugin designed for JetBrains IDEs to integrate with the gRPC Federation language server, offering various useful functionalities such as code completion.

## Features

Currently, the following features are supported:

- Goto Definition
- Diagnostics
- Code Completion

## Installation
To utilize the plugin, `grpc-federation-language-server` must be accessible in the path. If Go is available, execute the following command to install it:

```console
$ go install github.com/mercari/grpc-federation/cmd/grpc-federation-language-server@latest
```

Afterward, install the plugin from the Marketplace.

`Settings/Preferences` > `Plugins` > `Marketplace` > `Search for "gRPC Federation"` > `Install`

## Configuration
For `grpc-federation-language-server` to function properly, provide proto import paths to import dependent proto files.

`Settings/Preferences` > `Tools` > `gRPC Federation` > `Proto Import Paths`
<!-- Plugin description end -->
