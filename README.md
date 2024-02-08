# gRPC Federation

[![Test](https://github.com/mercari/grpc-federation/actions/workflows/test.yml/badge.svg)](https://github.com/mercari/grpc-federation/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/mercari/grpc-federation?status.svg)](https://pkg.go.dev/github.com/mercari/grpc-federation?tab=doc)
[![Buf](https://img.shields.io/badge/Buf-docs-blue)](https://buf.build/mercari/grpc-federation)

gRPC Federation automatically generate a BFF (Backend for frontend) server that aggregates and returns the results of gRPC protocol based microservices by writing Protocol Buffer's option.

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

# Features

## 1. Various tools are available to assist in code generation

gRPC Federation automatically generates a gRPC server by writing a custom option in Protocol Buffers. So, it supports the `protoc-gen-grpc-federation` CLI available from `protoc`. Various other tools exist to assist in code generation.

- `protoc-gen-grpc-federation`: protoc's plugin for gRPC Federation
- `grpc-federation-linter`: linter for gRPC Federation
- `grpc-federation-language-server`: language server program for gRPC Federation
- `grpc-federation-generator`: standalone code generation tool for monitoring proto changes and interactively performs code generation

## 2. Supports CEL API to represent complex operations

The gRPC Federation supports [CEL](https://github.com/google/cel-spec), allowing you to use the CEL API to represent complex operations you want to perform on your BFF.  

[gRPC Federation CEL API References](./docs/cel.md)

## 3. Extensible system with WebAssembly

The gRPC Federation has three extension points.

1. The code generation pipeline
2. The complex processes that cannot be expressed by Protocol Buffers
3. CEL API

We plan to make these extension points extensible with WebAssembly. Currently, only the CEL API can be extended.

### 3.1 The code generation pipeline

If you want to run your own auto-generated process using the results of the gRPC Federation, this feature is available.

[How to run your own code generation process](./docs/code_generation_plugin.md)

### 3.2 The complex processes that cannot be expressed by Protocol Buffers

The gRPC Federation uses a hybrid system in which logic that cannot be expressed in Protocol Buffers is developed in the Go language. Also, we plan to support WebAssembly in order to extend dedicated logic in the future.

[See here for features on extending with the Go language](./docs/references.md#grpcfederationmessagecustom_resolver)

### 3.3 CEL API

The gRPC Federation supports various CEL APIs by default. However, if you want to use internal domain logic as CEL API, you can use this functionality.

[How to extend the CEL API](./docs/cel_plugin.md)

# Installation

There are currently three ways to use gRPC Federation.

1. Use `buf generate`
2. Use `protoc-gen-grpc-federation`
3. Use `grpc-federation-generator`

For more information on each, please see here.

- [Installation](./docs/installation.md)

# Usage

- [`Getting Started`](./docs/getting_started.md)
- [`Examples`](./_examples/)

# Documents

- [Feature References](./docs/references.md)
- [CEL API References](./docs/cel.md)

# Contribution

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

# License

Copyright 2023 Mercari, Inc.

Licensed under the MIT License.
