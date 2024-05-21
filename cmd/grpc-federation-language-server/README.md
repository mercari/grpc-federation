# gRPC Federation Language Server

<img width="500px" src="https://github.com/mercari/grpc-federation/blob/main/images/icon.png?raw=true"/>

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

## Installation

```console
$ go install github.com/mercari/grpc-federation/cmd/grpc-federation-language-server@latest
```

## Usage

```console
Usage:
  grpc-federation-language-server [OPTIONS]

Application Options:
      --log-file=    specify the log file path for debugging
  -I, --import-path= specify the import path for loading proto file

Help Options:
  -h, --help         Show this help message
```

## Clients

The list of extensions or plugins for the IDE.

### Visual Studio Code

https://marketplace.visualstudio.com/items?itemName=Mercari.grpc-federation