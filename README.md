# gRPC Federation

[![Test](https://github.com/mercari/grpc-federation/actions/workflows/test.yml/badge.svg)](https://github.com/mercari/grpc-federation/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/mercari/grpc-federation?status.svg)](https://pkg.go.dev/github.com/mercari/grpc-federation?tab=doc)


gRPC Federation is a mechanism to automatically generate a BFF (Backend for frontend) server that aggregates and returns the results of gRPC protocol based microservices by writing Protocol Buffer's option.


*NOTE*: gRPC Federation is still in alpha version. Please do not use it in production, as the design may change significantly.

# Motivation

Consider a system with a backend consisting of multiple microservices.
In this case, it is known that instead of the client communicating directly with each microservice, it is better to prepare a dedicated service that accepts requests from the client, aggregates the necessary information from each microservice, and returns it to the client application.
This dedicated service is called a BFF ( Backend for Frontned ) service.  

However, as the system grows, the BFF becomes dependent on a variety of services. This raises the problem of which team is responsible for maintaining the BFF service ( responsibilities tend to ambiguity ). Federated Architecture can be used to solve this problem. 

A well-known example of Federated Architecture is the [GraphQL ( Apollo ) Federation](https://www.apollographql.com/docs/federation/).  

Apollo Federation assumes that each microservice has its own GraphQL server, and by limiting the work performed on the BFF to tasks such as aggregating acquired resources, it is possible to keep the BFF service itself in a simple state.

However, it is not easy to add a new GraphQL server to an already huge system. In addition, there are various problems that must be solved in order to introduce GraphQL to a system that has been developed using the gRPC protocol.

We thought that if we could automatically generate a BFF service using the gRPC protocol, we could continue to use gRPC while reducing the cost of maintaining the BFF service.
Since the schema information for the BFF service is already defined using Protocol Buffers, we are trying to accomplish this by giving the minimum required implementation information using custom options.

# Features

gRPC Federation automatically generates a gRPC server by writing a custom option in Protocol Buffers.  
To generate a gRPC server, you can use libraries for the Go language as well as the `protoc` plugin.  
We also provide a linter to verify that option descriptions are correct, and a language server to assist with option descriptions.  

- `protoc-gen-grpc-federation`: protoc's plugin for gRPC Federation
- `grpc-federation-linter`: linter for gRPC Federation
- `grpc-federation-language-server`: language server program for gRPC Federation
- `grpc-federation-generator`: monitors proto changes and interactively performs code generation

# Why use this

## 1. Reduce the typical implementation required to create a BFF

### 1.1. Many type conversion work for the same message between different packages

Consider defining a BFF service with a proto, assuming that the BFF service proxies some microservices calls and aggregates the results. In this case, we need to define a message to store the responses of the microservices that depend on the BFF side. If we were to use the messages defined in different packages as is, BFF's client need to be aware of the different packages. This is not a good idea, so the message must be redefined on the BFF side.

This means that a large number of messages almost identical to the microservice-side messages on which the BFF depends will be created on the BFF side.  
This causes a lot of type conversion work for the same message between different packages.

With gRPC Federation, there is no need to go through the tedious process of type conversion. Just define the simple option on proto and it will automatically convert the typed value.

### 1.2. Optimize gRPC method calls

When calling multiple gRPC methods, performance can be improved by analyzing the dependencies between method calls, and requests in parallel when possible. However, figuring out dependencies becomes more difficult as the number of method calls increases, and requests in parallel tends to become more complex implementation. 

gRPC Federation automatically analyzes the dependencies and generates code that makes the request in the most efficient way, so you don't have to think about them.

### 1.3. Retries and timeouts on gRPC method calls

When calling a gRPC method of a dependent service, you may set a timeout or a retry count. These implementations can be complex, but with gRPC Federation, they can be written declaratively.


These are just a few examples of how gRPC federation can make your work easier, but if you can create a BFF service with only a few definitions on the proto, you can lower the cost of developing BFF services to date.

## 2. Explicitly declare dependencies on services

By using the gRPC Federation, it is possible to know which services the BFF depends on just by reading the proto file. In some cases, you can get a complete picture of which microservice methods a given method of the BFF depends on.  

In addition, since the gRPC Federation provides the functionality to obtain dependencies as a library in Go, it is possible to automatically obtain service dependencies simply by statically analyzing proto file. This is very useful for various types of analysis and automation.

# Philosophy

## 1. gRPC method call exists to create a response message

Normally, the response message of a gRPC method exists as a result of the processing of the gRPC method, so the implementer's concern is "processing content" first and "processing result" second.

The gRPC Federation focuses on the "processing result" and considers that the gRPC method is called to create a response to the message.

Therefore, what you must do with gRPC Federation is to write dedicated options in those messages to create response messages for each of the Federation Service methods.

## 2. For every message, there is always a method that returns that message

Every message defined in Protocol Buffers has a method to return it. Therefore, it is possible to link message and method.


With these, we believe we can implemente the Federation Service by defining the message dependencies to create a response message and map them to the method definition.

# Installation

There are currently two ways to use gRPC Federation.

1. Use `protoc-gen-grpc-federation`
2. Use `grpc-federation-generator`

## 1. Use protoc-gen-grpc-federation

### 1.1. Install `protoc-gen-grpc-federation`

`protoc-gen-grpc-federation` is a `protoc` plugin made by Go. It can be installed with the following command.

```console
go install github.com/mercari/grpc-federation/cmd/protoc-gen-grpc-federation@latest
```

### 1.2. Put gRPC Federation proto file to the import path

Copy the gRPC Federation proto file to the import path.
gRPC Federation's proto file is [here](./proto/grpc/federation/federation.proto).

Also, gRPC Federation depends on the following proto file. These are located in the `proto_deps` directory and should be added to the import path if necessary.

- [`google/rpc/code.proto`](./proto_deps/google/rpc/code.proto)
- [`google/rpc/error_details.proto`](./proto_deps/google/rpc/error_details.proto)

```console
git clone https://github.com/mercari/grpc-federation.git
cd grpc-federation
cp -r proto /path/to/proto
cp -r proto_deps /path/to/proto_deps
```

### 1.3. Use gRPC Federation proto file in your proto file.

- my_federation.proto

```proto
syntax = "proto3";

package mypackage;

import "grpc/federation/federation.proto";
```

### 1.4. Run `protoc` command with gRPC Federation plugin

```console
protoc -I/path/to/proto -I/path/to/proto_deps --grpc-federation_out=. ./path/to/my_federation.proto
```

> [!NOTE]
> In the future, we plan to make it easier to use by registering with the [Buf Schema Registry](https://buf.build/explore) !


## 2. Use `grpc-federation-generator`

`grpc-federation-generator` is an standalone generator that allows you to use the gRPC Federation code generation functionality without `protoc` plugin.
Also, the code generated by gRPC Federation depends on the code generated by `protoc-gen-go` and `protoc-gen-go-grpc`, so it generates `file.pb.go` and `file_grpc.pb.go` files at the same time without that plugins. ( This behavior can be optionally disabled )


`grpc-federation-generator` also has the ability to monitor changes in proto files and automatically generate them, allowing you to perform automatic generation interactively. Combined with the existing hot-reloader tools, it allows for rapid Go application development.
The version of the gRPC Federation proto is the version when `grpc-federation-generator` is built.

### 1. Install `grpc-federation-generator`

`grpc-federation-generator` made by Go. It can be installed with the following command.

```console
go install github.com/mercari/grpc-federation/cmd/grpc-federation-generator@latest
```

### 2. Write configuration file for `grpc-federation-generator`

Parameters to be specified when running `protoc` are described in `grpc-federation.yaml`

- grpc-federation.yaml

```yaml
# specify import paths.
imports:
  - proto

# specify the directory where your proto files are located.
# In watch mode, files under these directories are recompiled and regenerated immediately when changes are detected.
src:
  - proto

# specify the destination of gRPC Federation's code-generated output.
out: .

# specify plugin options to be used during code generation of gRPC Federation.
# The format of this plugins section is same as [buf.gen.yaml](https://buf.build/docs/configuration/v1/buf-gen-yaml#plugins).
# Of course, other plugins can be specified, such as `buf.gen.yaml`.
plugins:
  - plugin: go
    opt: paths=source_relative
  - plugin: go-grpc
    opt: paths=source_relative
  - plugin: grpc-federation
    opt: paths=source_relative
```

### 3. Run `grpc-federation-generator`

If a proto file is specified, code generation is performed on that file.

```console
grpc-federation-generator ./path/to/my_federation.proto
```

When invoked with the `-w` option, monitors changes to the proto file for directories specified in `grpc-federation.yaml` and performs interactive compilation and code generation.

```console
grpc-federation-generator -w
```

> [!NOTE]
> In the future, we plan to implement a mechanism for code generation by other plugins in combination with `buf.work.yaml` and `buf.gen.yaml` definitions.

# Usage

- [`Getting Started`](./docs/getting_started.md)
- [`Examples`](./_examples/)

# Contribution

*NOTE*: Currently, we do not accept pull requests from anyone other than the development team. Please report feature requests or bug fix requests via Issues.

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

# License

Copyright 2023 Mercari, Inc.

Licensed under the MIT License.
