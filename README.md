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

```console
git clone https://github.com/mercari/grpc-federation.git
cd grpc-federation
cp -r proto /path/to/proto
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
protoc -I/path/to/proto --grpc-federation_out=. ./path/to/my_federation.proto
```

**NOTE**

In the future, we plan to make it easier to use by registering with the [Buf Schema Registry](https://buf.build/explore) !


## 2. Use `grpc-federation-generator`

`grpc-federation-generator` is an standalone generator that allows you to use the gRPC Federation code generation functionality without `protoc` plugin.
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

# specify a list of directories to be monitored for changes in proto files.
src:
  - proto

# specify the destination of gRPC Federation's code-generated output.
out: .

# specify options to be used during code generation of gRPC Federation. 
opt:
  paths: source_relative
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

**NOTE**

In the future, we plan to implement a mechanism for code generation by other plugins in combination with `buf.work.yaml` and `buf.gen.yaml` definitions.

# Usage

Consider the example of creating a BFF ( `FederationService` ) that combines and returns the results obtained using the `PostService` and `UserService`.

The definitions of `PostService` and `UserService` are as follows.

## PostService

`PostService` has `GetPost` gRPC method. It returns `Post` message.

- post.proto

```proto
package post;

service PostService {
  rpc GetPost (GetPostRequest) returns (GetPostReply) {}
}

message GetPostRequest {
  string post_id = 1;
}

message GetPostReply {
  Post post = 1;
}

message Post {
  string id = 1;
  string title = 2;
  string content = 3;
  string user_id = 4;
}
```

## UserService

`UserService` has `GetUser` gRPC method. It returns `User` message.

- user.proto

```proto
package user;

service UserService {
  rpc GetUser (GetUserRequest) returns (GetUserReply) {}
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserReply {
  User user = 1;
}

message User {
  string id = 1;
  string name = 2;
  int age = 3;
}
```

The `FederationService` implements an API that combines and returns the results obtained using the `PostService` and `UserService`.
We will explain how to implement the `FederationService` by using gRPC Federation.

## FederationService

`FederationService` has a `GetPost` method. It returns a response containing a `User` message in the `Post` message.

- federation.proto

```proto
package federation;

service FederationService {
  rpc GetPost (GetPostRequest) returns (GetPostReply) {}
}

message GetPostRequest {
  string id = 1;
}

message GetPostReply {
  Post post = 1;
}

message Post {
  string id = 1;
  string title = 2;
  string content = 3;
  User user = 4;
}

message User {
  string id = 1;
  string name = 3;
  int age = 3;  
}
```

In this example, we will use the following three types of options provided by the gRPC Federation for implementation.

- `grpc.federation.service` option
- `grpc.federation.message` option
- `grpc.federation.field` option

## grpc.federation.service option

The `grpc.federation.service` option is used to specify a service to be automatically generated using gRPC Federation.  
Therefore, your first step is to import the gRPC Federation proto file and add the service option.

```diff
+import "grpc/federation/federation.proto";

 service FederationService {
+  option (grpc.federation.service) = {};
   rpc GetPost (GetPostRequest) returns (GetPostReply) {}
 }
```

In the `grpc.federation.service` option, `dependencies` field is available to specify the services that depend on it.

`FederationService` depends on the `PostService` in `post` package and `UserService` in `user` package, write the following.

```diff
 import "grpc/federation/federation.proto";
+import "post.proto";
+import "user.proto";

 service FederationService {
   option (grpc.federation.service) = {
+    dependencies: [
+      { service: "post.PostService" },
+      { service: "user.UserService" }
+    ]
   };
   rpc GetPost (GetPostRequest) returns (GetPostReply) {}
 }
```

See the following table for a list of features available under the `grpc.federation.service` option.

| field | type | description |
| ----- | ---- | ----------- |
| `dependencies` | repeated ServiceDependency | dependent service information |

### ServiceDependency

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | used when initializing the gRPC client | optional |
| `service` | string | name of the dependent service | required |


## grpc.federation.message option

The message option is used for the following purposes.  
There must always be a method to retrieve it for a message. First, we link the message and its method. However, a parameter is required to call the method. Therefore, it is necessary to be able to define the parameters necessary to call the method in option.

In the `grpc.federation.message` option, the following fields are available to build itself.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `resolver` | Resolver | defines how to call method | optional |
| `messages` | repeated Message | defines a list of dependent messages | optional |


### Resolver

The resolver is used to link message and method. Only one resolver can be defined per message.

If you call the `GetPost` method of `post.PostService` service, write the following.

```diff
 message Post {
   option (grpc.federation.message) = {
+    resolver {
+      method: "post.PostService/GetPost"
+      // assign "abcd" constant value to post_id field of GetPostRequest and call GetPost.
+      request { field: "post_id", string: "abcd" }
+    }
   };
 }
```

You may have wondered about the use of constants as parameters in the request. This is for simplicity's sake, and the request is actually made using the value that exists in the `GetPostRequest` message of the `FederationService`.
To use the value of the `GetPostRequest` message, use the `by` feature and `$.` prefix. These will be explained later.

```proto
message Post {
  option (grpc.federation.message) = {
    resolver {
      method: "post.PostService/GetPost"
      // refer GetPostRequest message's id field by `$.id`.
      request { field: "post_id", by: "$.id" }
    }
  };
}
```

If you want to use the result of the `GetPost` method, use the `response` field as follows.

```diff
 message Post {
   option (grpc.federation.message) = {
     resolver {
       method: "post.PostService/GetPost"
       request { field: "post_id", by: "$.id" }
+      response {
+        // specify a name to refer the value for the `post` field.
+        name: "res"
+        // select `post` field from `GetPostReply` message in `post` package.
+        field: "post"
+        // automatically binds the fields ( `id` or `title` or `content` ) of `post` field to its own fields.
+        autobind: true
+      }
     }
   };
   string id = 1; // automatically binds the value of `res.id`.
   string title = 2; // automatically binds the value of `res.title`.
   string content = 3; // automatically binds the value of `res.content`.
 }
```

In this example, there is no `User user = 4;` field in the `Post` message. This field can be supported by the `Message` feature (explain later).

The `resolver` has the following fields.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `method` | string | specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`. | required |
| `request` | repeated MethodRequest | request parameters for the gRPC method | optional |
| `response` | repeated MethodResponse | which value of the gRPC method response to use | optional |

### MethodRequest

Refer to the request message of the method to be called and specify the necessary parameters. The `field` corresponds to the field name in the request message, and the `by` or `string` specifies which value is associated with the field.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `field` | string | field name of the request message | required |
| `by` | string | used to refer to a name or message argument defined in a message option, use `$.` to refer to the message argument | optional |
| `string` | string | the referenced value directly in literal (string) | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

### MethodResponse

Describe the information necessary to bind the result of the method call to the message field. Select and name the required information from the response message, and use the defined name when referencing the information in the message option.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | the unique name that can be used in a message option or field option for a specific field in the response | optional |
| `field` | string | field name in response message | optional |
| `autobind` | bool | if the value referenced by `field` is a message type, the value of a field with the same name and type as the field name of its own message is automatically assigned to the value of the field in the message | optional |

### Message

Since only one `resolver` can be defined per message, use the `messages` field if multiple methods must be called to assign values to all fields.
`messages` is used when it is necessary to depend on other messages.  
When depends on a message, you can specify parameters to be passed to the depndent message. This parameter is called the `message argument`.

In the following example, We retrieve the `User` message and assign to the fourth field of Post message.

```diff
 message Post {
   option (grpc.federation.message) = {
     resolver {
       method: "post.PostService/GetPost"
       request { field: "post_id", by: "$.id" }
       response { name: "res", field: "post", autobind: true }
     }
+    messages {
+      // specify the name to refer the value for User message
+      name: "u"
+      // specify the name of the dependent message.
+      message: "User"
+      // specify the message arguments for resolving dependent message.
+      args {
+        // The name of the message argument. This name with the `$.` prefix can refer to this value.
+        name: "uid"
+        // specify the value.
+        by: "res.user_id"
+      }
+    }
   };
   string id = 1;
   string title = 2;
   string content = 3;
+  // represents directly binding the value of the User message.
+  User user = 4 [(grpc.federation.field).by = "u"];
 }

 message User {
+  option (grpc.federation.message) = {
+    resolver {
+      method: "user.UserService/GetUser"
+      request { field: "user_id", by: "$.uid" }
+      response { field: "user", autobind: true }
+    }
+  };
   string id = 1;
   string name = 2;
   int age = 3;
 }
```

The referenced by the options are shown in the figure below.

The `messages` has the following fields.

| field | type | description | required or optional |
| ------| ---- | ----------- | -------------------- |
| `name` | string | a unique name for the dependent message | required |
| `message` | string | the message to be referred to by FQDN. format is `<package-name>.<message-name>`. The `<package-name>` can be omitted when referring to messages in the same package | required |
| `args` | repeated Argument | the parameters needed to retrieve the message. This is called the `message argument` | optional |

### Argument

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | name of the message argument. Use this name to refer to the message argument. For example, if `foo` is specified as the name, it is referenced by `$.foo` | required |
| `by` | string | refer to a name or message argument defined in message option, use `$.` to refer to the message argument | optional |
| `inline` | string | like by, it refers to the specified value and expands all fields beyond it. For this reason, the referenced value must always be of message type | optional |
| `string` | string | specify the referenced value directly in literal | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## grpc.federation.field option

The field option used to specify which value to bind to each field of a message.  
The only names that can be referenced by `by` are those specified in the message option of the message for which the field is defined.


| field | type | description | required or optional |
| ------| ---- | ----------- | -------------------- |
| `by` | string | refer to a name or message argument defined in message option, use `$.` to refer to the message argument | optional |
| `string` | string | specify the referenced value directly in literal (string) | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## Build response message

The purpose of gRPC Federation is to build a response messsage. Therefore, the examples described so far are not yet complete. To build a response message, write the following.

```diff
 message GetPostReply {
+  option (grpc.federation.message) = {
+    messages { name: "p", message: "Post", args { name: "id", by: "$.id" } }
+  };
+  Post post = 1 [(grpc.federation.field).by = "p"];
 }
```

`GetPostReply` corresponds to the root message. In this message, all fields of `GetPostRequest` can be referenced as message arguments. Thus, `"$.id"` refers to the `id` field of the `GetPostRequest` message.

In this example, we name the retrieved `Post` message `p`. Then, `grpc.federation.field` option is used to refer to `p` and bind it to a field.

Finally, the completed `FederationService` will look like this.

```proto
package federation;

import "grpc/federation/federation.proto";
import "post.proto";
import "user.proto";

service FederationService {
  option (grpc.federation.service) = {
    dependencies: [
     { service: "post.PostService" },
     { service: "user.UserService" }
    ]
  };
  rpc GetPost (GetPostRequest) returns (GetPostReply) {}
}

message GetPostRequest {
  string id = 1;
}

message GetPostReply {
  option (grpc.federation.message) = {
    messages { name: "p", message: "Post", args { name: "id", by: "$.id" } }
  };
  Post post = 1 [(grpc.federation.field).by = "p"];
}

message Post {
  option (grpc.federation.message) = {
    resolver {
      method: "post.PostService/GetPost"
      request { field: "post_id", by: "$.id" }
      response { name: "res", field: "post", autobind: true }
    }
    messages { name: "u", message: "User", args { name: "uid", by: "res.user_id" } }
  };

  string id = 1;
  string title = 2;
  string content = 3;
  User user = 4 [(grpc.federation.field).by = "u"];
}

message User {
  option (grpc.federation.message) = {
    resolver {
      method: "user.UserService/GetUser"
      request { field: "user_id", by: "$.uid" }
      response { field: "user", autobind: true }
    }
  };
  string id = 1;
  string name = 3;
  int age = 3;  
}
```

## Other Important Specs

### Order of method invocation

The order in which `resolver` and `messages` are called is automatically determined by the parameters on which they depend.  
If the request parameter of the `resolver` is made with reference to the message specified in `messages`, the message is obtained first. Conversely, if the response value obtained by the `resolver` is passed as a message parameter, the `resolver` is called first.  

If there are circular dependencies, etc., an error will occur during automatic generation.

### Message Argument

Message argument is an argument passed when depends on other messages via the `messages` option. This parameter propagates the information necessary to resolve message dependencies.

Normally, only values explicitly specified in `args` can be referenced as message arguments.

However, exceptionally, a field in the method's request message can be referenced as a message argument to the method's response.

For example, consider the GetPost method of the FederationService: the GetPostResponse message implicitly allows the fields of the GetPostRequest message to be used as message arguments.

In summary, it looks like this.

1. For the root message ( the response message of rpc ), the request message fields are the message arguments.
2. For other messages, the parameters explicitly passed in `messages[].args` option are the message arguments.
3. You can access to the message argument with `$.` prefix.

### Implements custom resolvers in Go

There is a limit to what can be expressed in proto, so if you want to execute a process that cannot be expressed in proto, you will need to implement it yourself in the Go language.  

In the services for the gRPC Federation, by setting `custom_resolver: true` in the message or field option, only that part can be implemented in Go. 

```proto
message Foo {
  option (grpc.federation.message).custom_resolver = true;
}
```

or

```proto
message Foo {
  User user = 1 [(grpc.federation.field).custom_resolver = true];
}
```

## Use generated server

Running code generation using the `federation.proto` described above will create a `federationservice_grpc_federation.go` file under the specified path.

In `federationservice_grpc_federation.go`, an initialization function ( `NewFederationService` ) for the `FederationService` is created. The server instance initialized using that function can be registered as a gRPC server using the `RegisterFederationService` function defined in `federation_grpc.pb.go` as is.

When initializing, you need to create a dedicated Config structure and pass.

```go
type ClientConfig struct{}

func (c *ClientConfig) Post_PostServiceClient(cfg federation.FederationServiceClientConfig) (post.PostServiceClient, error) {
  // create by post.NewPostServiceClient()
  ...
}

func (c *ClientConfig) User_UserServiceClient(cfg federation.FederationServiceClientConfig) (user.UserServiceClient, error) {
  // create by user.NewUserServiceClient()
  ...
}

federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
  // Client provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
  // If this interface is not provided, an error is returned during initialization.
  Client: new(ClientConfig),
})
if err != nil { ... }
grpcServer := grpc.NewServer()
federation.RegisterFederationServiceServer(grpcServer, federationServer)
```

`Config` (e.g. `federation.FederationServiceConfig` ) must always be passed a configuration to initialize the gRPC Client, which is needed to invoke the methods on which the federation service depends.

Also, there are settings for customizing error handling on method calls or logger, etc.

# Examples

Examples using gRPC Federation are located under the [_examples](./_examples) directory.


# Contribution

*NOTE*: Currently, we do not accept pull requests from anyone other than the development team. Please report feature requests or bug fix requests via Issues.

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

# License

Copyright 2023 Mercari, Inc.

Licensed under the MIT License.