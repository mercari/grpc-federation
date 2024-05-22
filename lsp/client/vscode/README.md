# gRPC Federation Language Server for VSCode

<img width="500px" src="https://github.com/mercari/grpc-federation/blob/main/images/logo.png?raw=true"/>

## Features

### Syntax Highlighting

Semantic Highlighting allows for accurate recognition of gRPC Federation options.
Syntax highlighting is available in quoted value. So, this is especially effective in CEL value.

<img width="800px" src="https://github.com/mercari/grpc-federation/blob/main/images/semantic_highlighting.png?raw=true"/>

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

## Install

To use this extension, `grpc-federation-language-server` is required.

If already installed Go in your local environment, please run the following command.

```console
$ go install github.com/mercari/grpc-federation/cmd/grpc-federation-language-server@latest
```

## Settings

The following options can be set in `.vscode/settings.json` .

The example settings is here.

```json
{
    "grpc-federation": {
        "path": "/path/to/grpc-federation-language-server",
        "import-paths": [
          "./proto"
        ]
    }
}
```

### path

Specify the path to the location where `grpc-federation-language-server` is installed.
If the installation location has already been added to your `PATH` environment variable, you do not need to specify this.

### import-paths

Specifies the path to search for proto files.


