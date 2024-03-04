# gRPC Federation

[![Test](https://github.com/mercari/grpc-federation/actions/workflows/test.yml/badge.svg)](https://github.com/mercari/grpc-federation/actions/workflows/test.yml)
[![GoDoc](https://godoc.org/github.com/mercari/grpc-federation?status.svg)](https://pkg.go.dev/github.com/mercari/grpc-federation?tab=doc)
[![Buf](https://img.shields.io/badge/Buf-docs-blue)](https://buf.build/mercari/grpc-federation)

gRPC Federation auto-generates a BFF (Backend for Frontend) server, aggregating and returning results from microservices using the gRPC protocol, configured through Protocol Buffer options.

# Motivation

Imagine a system with a backend of multiple microservices. Instead of the client directly communicating with each microservice, it's more efficient to use a dedicated service (BFF - Backend for Frontend) to aggregate information and send it back to the client. However, as the system grows, determining which team should handle the BFF service becomes unclear. This is where Federated Architecture, like [GraphQL (Apollo) Federation](https://www.apollographql.com/docs/federation/), comes in.

With GraphQL Federation, each microservice has its own GraphQL server, and the BFF focuses on aggregating resources. However, integrating a new GraphQL server into a large system is challenging, especially in systems developed with the gRPC protocol.

To streamline, we're exploring the idea of automatically generating a BFF using the gRPC protocol with simple custom options. By leveraging existing Protocol Buffers schema for the BFF, we aim to keep using gRPC while drastically reducing maintenance costs.

# Why Choose gRPC Federation?

## 1. Reduce the Boilerplate Implementation Required to Create a BFF

### 1.1. Automate Type Conversions for the Same Message Across Different Packages

Consider defining a BFF using a proto, assuming the BFF acts as a proxy calling multiple microservices and aggregating results. In this scenario, it's necessary to redefine the same messages to handle microservices responses on the BFF. Without defining these messages on the BFF, the BFF client must be aware of different packages, which is not ideal.

Redefining messages on the BFF results in a considerable number of type conversions for the same messages across different packages.

Using gRPC Federation removes the need for tedious type conversions. Just define custom options in the proto, and gRPC Federation will automatically handle these conversions.

### 1.2. Optimize gRPC Method Calls

When making multiple gRPC method calls, it is crucial to enhance performance by analyzing dependencies between method calls and processing requests in parallel whenever possible. However, identifying dependencies becomes more challenging as the number of method calls increases.

gRPC Federation simplifies this process by automatically analyzing dependencies and generating code to optimize request flows. It removes the need for manual consideration.

### 1.3. Simplify Retries and Timeouts on gRPC Method Calls

Setting timeouts or retry counts for gRPC method calls in dependent services can be complicated. However, gRPC Federation allows for a declarative approach to implementing retries and timeouts.

These are just a few examples of how gRPC Federation simplifies work. By creating a BFF with minimal proto definitions, you can significantly reduce the development cost of BFFs.

## 2. Explicitly Declare Dependencies on Services

Using gRPC Federation enables a clear understanding of the services the BFF depends on by reading the proto file. It uncovers a relationship between each microservice and BFF.

Furthermore, gRPC Federation offers functionality to extract dependencies as a Go library, enabling the extraction of service dependencies through static analysis of the proto file. This capability proves valuable for various types of analysis and automation.

# Features

## 1. Code Generation Assistance Tools

gRPC Federation automatically generates a gRPC server by adding custom options to Protocol Buffers. It supports the `protoc-gen-grpc-federation` CLI, accessible through `protoc`. Various other tools are available to assist in code generation:

- `protoc-gen-grpc-federation`: protoc's plugin for gRPC Federation
- `grpc-federation-linter`: Linter for gRPC Federation
- `grpc-federation-language-server`: Language server for gRPC Federation
- `grpc-federation-generator`: Standalone code generation tool, which monitors proto changes and interactively generates codes

## 2. CEL Support for Complex Operations

gRPC Federation integrates [CEL](https://github.com/google/cel-spec) to enable support for more advanced operations on BFFs.  

[gRPC Federation CEL API References](./docs/cel.md)

## 3. Extensible System with WebAssembly

gRPC Federation features three extension points:

1. Code generation pipeline
2. Complex processes not expressible by Protocol Buffers
3. CEL API

We plan to make all three extension points available through WebAssembly in the future. Currently, the code generation pipeline and the CEL API are extensible.

### 3.1 Code Generation Pipeline

This feature allows you to run a custom auto-generated process using the results from gRPC Federation.

[How to Run Your Own Code Generation Process](./docs/code_generation_plugin.md)

### 3.2 Complex Processes Not Expressible by Protocol Buffers

gRPC Federation uses a hybrid system, with logic in Go managing operations that can't be expressed in Protocol Buffers. We plan to add support for WebAssembly in the future.

[How to Extend the API with the Go](./docs/references.md#grpcfederationmessagecustom_resolver)

### 3.3 CEL API

While gRPC Federation inherently supports various CEL APIs, users can use this functionality to use internal domain logic as a CEL API.

[How to Extend the CEL API](./docs/cel_plugin.md)

# Installation

Currently, there are three ways to use gRPC Federation:

1. Use `buf generate`
2. Use `protoc-gen-grpc-federation`
3. Use `grpc-federation-generator`

For detailed information on each method, please refer to [Installation Documentation](./docs/installation.md).

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
