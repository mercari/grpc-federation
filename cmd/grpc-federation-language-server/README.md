# gRPC Federation Language Server

## Features

### Syntax Highlighting

Semantic Highlighting allows for accurate recognition of gRPC Federation options.
Syntax highlighting is available in quoted value. So, this is especially effective in CEL value.

<img width="800px" src="https://github.com/mercari/grpc-federation/blob/main/images/semantic_highlighting.png?raw=true"/>

### Jump to Definition

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