# gRPC Federation

gRPC Federation is a mechanism to automatically generate a BFF (Backend for frontend) server that aggregates and returns the results of gRPC protocol based microservices by writing Protocol Buffer's option.


*NOTE*: gRPC Federation is still in alpha version. Please do not use it in production, as the design may change significantly.

## Motivation

Consider a system with a backend consisting of multiple microservices.
In this case, when returning responses to requests from client applications, instead of communicating directly with each microservice. It is known that it is better to have a dedicated service that accepts requests from clients and aggregates necessary information from various microservices and returns it to the client application. This service is called a BFF ( Backend for Frontned ) service.  

However, as the system grows, when this BFF service becomes large, the question arises as to which team is responsible for maintaining the BFF service, since the BFF depends on a variety of services and the responsibilities tend to be ambiguous. To solve this problem, a Federated Architecture may be used.  

A well-known example of Federated Architecture is the GraphQL ( Apollo ) Federation.  

Apollo Federation assumes that each microservice has its own GraphQL server, and by limiting the work performed on the BFF to tasks such as aggregating acquired resources, it is possible to keep the BFF service itself in a simple state.

However, it is not easy to add a new GraphQL server to an already huge system. In addition, there are various problems that must be solved in order to introduce GraphQL to a system that has been developed using the gRPC protocol.

We thought that if we could automatically generate a BFF service using the gRPC protocol, we could continue to use gRPC while reducing the cost of maintaining the BFF service.
Since the schema information for the BFF service is already defined using protocol buffers, we are trying to accomplish this by giving the minimum required implementation information using custom options.

## Features

gRPC Federation automatically generates a gRPC server by writing a custom option in Protocol Buffers.  
To generate a gRPC server, you can use libraries for the Go language as well as the protoc plugin.  
We also provide a linter to verify that option descriptions are correct, and a language server to assist with option descriptions.  

- `protoc-gen-grpc-federation`: protoc's plugin for gRPC Federation
- `grpc-federation-linter`: linter for gRPC Federation
- `grpc-federation-language-server`: language server program for gRPC Federation

## Why use this

### Can eliminate the typical work that must be done in building the BFF

Consider defining a BFF service with a proto, assuming that the BFF service proxies some microservices calls and aggregates the results. In this case, we need to define a message to store the responses of the microservices that depend on the BFF side. If we were to use the messages defined in different packages as is, BFF's client need to be aware of the different packages. This is not a good idea, so the message must be redefined on the BFF side.

This means that a large number of messages almost identical to the microservice-side messages on which the BFF depends will be created on the BFF side.  
This causes a lot of type conversion work for the same message between different packages.

With gRPC Federation, there is no need to go through the tedious process of type conversion. Just define the simple option on proto and it will automatically convert the value.

This is just one example of how the gRPC Federation can make your work easier, but if you can create a BFF service with only a small amount of definitions on a proto, you will be able to lower the cost of developing a BFF service up to now.

### Can make dependencies on services explicit

By using the gRPC Federation, it is possible to know which services the BFF depends on just by reading the proto file. In some cases, you can get a complete picture of which microservice methods a given method of the BFF depends on.  

In addition, since the gRPC Federation provides the functionality to obtain dependencies as a library in Go, it is possible to automatically obtain service dependencies simply by statically analyzing proto file. This is very useful for various types of analysis and automation.

## Philosophy

### gRPC method call exists to create a response message

Consider that the gRPC method call exists to create a response message.
In other words, instead of "a response message is obtained as a result of a gRPC method call" consider that "a gRPC method call is made to create a response message".

In this way of thinking, the response message to the gRPC method is what is noteworthy.
Therefore, in the gRPC Federation, the Protocol Buffers option is written while considering how to construct the response message.

### For every message, there is always a method that returns that message

Every message defined in Protocol Buffers has a method to return it. Therefore, it is possible to link message and method.


## Usage

## Usage of Protocol Buffers option

We will explain using the `PostService` and `UserService`.

post.proto
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

user.proto
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

The `FederationService` service that implements an API that aggregates and returns the results of `PostService` and `UserService`.
We will explain how to implement the `FederationService` by using gRPC Federation.

federation.proto
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

The relationship between services can be illustrated as follows.

The original gRPC Federation's definitions are [here](./proto/grpc/federation/federation.proto).

### grpc.federation.service option

The `grpc.federation.service` option is used to specify a service to be automatically generated using gRPC Federation.  
Simply add this option and the gRPC server code will be automatically generated. However, this option alone does not provide implementation details, so you will have to add your own implementation in the Go language.

```diff
 service FederationService {
+  option (grpc.federation.service) = {};
   rpc GetPost (GetPostRequest) returns (GetPostReply) {}
 }
```

In the `grpc.federation.service` option, `dependencies` field is available to specify the services that depend on it.

`Federation` service depends on the `PostService` in `post` package and `UserService` in `user` package, write the following.

```proto
import "post.proto";
import "user.proto";

service FederationService {
  option (grpc.federation.service) = {
    dependencies: [
      { name: "post", service: "post.PostService" },
      { name: "user", service: "user.UserService" }
    ]
  };
  rpc GetPost (GetPostRequest) returns (GetPostReply) {}
}
```


| field | type | description |
| ----- | ---- | ----------- |
| `dependencies` | repeated ServiceDependency | dependent service information |

#### ServiceDependency

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | used when initializing the gRPC client | required |
| `service` | string | name of the dependent service | required |


### grpc.federation.message option

The message option is used for the following purposes.  
There must always be a method to retrieve it for a message. First, we link the message and its method. However, a parameter is required to call the method. Therefore, it is necessary to be able to define the parameters necessary to call the method in option.

In the `grpc.federation.message` option, the following fields are available to build itself.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `resolver` | Resolver | defines how to call method | optional |
| `messages` | repeated Message | defines a list of dependent messages | optional |


#### Resolver

The resolver is used to link message and method. Only one resolver can be defined per message.

If you call the `GetPost` method of `post.PostService` service, write the following.

```proto
message Post {
  option (grpc.federation.message) = {
    resolver {
      method: "post.PostService/GetPost"
      // assign "abcd" constant value to post_id field of GetPostRequest and call GetPost.
      request: [ { field: "post_id", value: { string: "abcd" } }]
    }
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
      // refer GetPostRequest message's id field.
      request: [ { field: "post_id", by: "$.id" } ]
    }
  };
}
```

If you want to use the result of the `GetPost` method, use the `response` field as follows.

```proto
message Post {
  option (grpc.federation.message) = {
    resolver {
      method: "post.PostService/GetPost"
      request: [ { field: "post_id", by: "$.id" } ]
      response: [{
        // specify a name to refer the value for the `post` field.
        name: "res"
        // select `post` field from `GetPostReply` message in `post` package.
        field: "post"
        // automatically binds the fields ( `id` or `title` or `content` ) of `post` field to its own fields.
        autobind: true
      }]
    }
  };
  string id = 1; // automatically binds the value of `res.id`.
  string title = 2; // automatically binds the value of `res.title`.
  string content = 3; // automatically binds the value of `res.content`.
}
```

The fields referenced by the options are shown in the figure below.

In this example, there is no `User user = 4;` field in the `Post` message. This field can be supported by the `Message` option.

The `resolver` has the following fields.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `method` | string | specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`. | required |
| `request` | repeated MethodRequest | request parameters for the gRPC method | optional |
| `response` | repeated MethodResponse | which value of the gRPC method response to use | optional |

#### MethodRequest

Refer to the request message of the method to be called and specify the necessary parameters. The `field` corresponds to the field name in the request message, and the `by` or `value` specifies which value is associated with the field.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `field` | string | field name of the request message | required |
| `by` | string | used to refer to a name or message argument defined in a message option, use `$.` to refer to the message argument | optional |
| `value` | Value | the referenced value directly in literal | optional |

#### MethodResponse

Describe the information necessary to bind the result of the method call to the message field. Select and name the required information from the response message, and use the defined name when referencing the information in the message option.

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | the unique name that can be used in a message option or field option for a specific field in the response | optional |
| `field` | string | field name in response message | optional |
| `autobind` | bool | if the value referenced by `field` is a message type, the value of a field with the same name and type as the field name of its own message is automatically assigned to the value of the field in the message | optional |

#### Message

Since only one `resolver` can be defined per message, use the `messages` field if multiple methods must be called to assign values to all fields.
`messages` is used when it is necessary to depend on other messages.  
When depends on a message, you can specify parameters to be passed to the depndent message. This parameter is called the `message argument`.

In the following example, We retrieve the `User` message and assign to the fourth field of Post message.

```proto
message Post {
  option (grpc.federation.message) = {
    resolver {
      method: "post.PostService/GetPost"
      request: [ { field: "post_id", by: "$.id" } ]
      response: [{ name: "res", field: "post", autobind: true }]
    }
    messages {
      // specify the name to refer the value for User message
      name: "u"
      // specify the name of the dependent message.
      message: "User"
      // specify the message arguments for resolving dependent message.
      args: [{
        // The name of the message argument. This name with the `$.` prefix can refer to this value.
        name: "uid"
        // specify the value.
        by: "res.user_id"
      }]
    }
  };
  string id = 1;
  string title = 2;
  string content = 3;
  // represents directly binding the value of the User message.
  User user = 4 [(grpc.federation.field).by = "u"];
}

message User {
  option (grpc.federation.message) = {
    resolver {
      method: "user.UserService/GetUser"
      request: [ { field: "user_id", by: "$.uid" } ]
      response: [ { name: "res", field: "user", autobind: true } ]
    }
  };
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

#### Argument

| field | type | description | required or optional |
| ----- | ---- | ----------- | -------------------- |
| `name` | string | name of the message argument. Use this name to refer to the message argument. For example, if `foo` is specified as the name, it is referenced by `$.foo` | required |
| `by` | string | refer to a name or message argument defined in message option, use `$.` to refer to the message argument | optional |
| `inline` | string | like by, it refers to the specified value and expands all fields beyond it. For this reason, the referenced value must always be of message type | optional |
| `value` | Value | specify the referenced value directly in literal | optional |

### grpc.federation.field option

The field option used to specify which value to bind to each field of a message.  
The only names that can be referenced by `by` are those specified in the message option of the message for which the field is defined.


| field | type | description | required or optional |
| ------| ---- | ----------- | -------------------- |
| `by` | string | refer to a name or message argument defined in message option, use `$.` to refer to the message argument | optional |
| `value` | Value | specify the referenced value directly in literal | optional |

### Build response message

The purpose of gRPC Federation is to build a response messsage. Therefore, the examples described so far are not yet complete. To build a response message, write the following.

```proto
message GetPostReply {
  option (grpc.federation.message) = {
    messages { name: "p", message: "Post", args { name: "id", by: "$.id" } }
  };
  Post post = 1 [(grpc.federation.field).by = "p"];
}
```

`GetPostReply` corresponds to the root message. In this message, all fields of `GetPostRequest` can be referenced as message arguments. Thus, `"$.id"` refers to the `id` field of the `GetPostRequest` message.

In this example, we name the retrieved `Post` message `p`. Then, `grpc.federation.field` option is used to refer to `p` and bind it to a field.

See the aforementioned sample for how to build a Post message.

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

In the services for the gRPC Federation, omitting the message option and field option generates code that assumes the creation of a custom resolver.

## Usage of generated server

Normally, using the `protoc` compiler to generate code for the Go language from `federation.proto` will generate `federation.pb.go` ( by `protoc-gen-go` ) and `federation_grpc.pb.go` ( by `protoc-gen-grpc-go` ). When using gRPC Federation, `federation_grpc_federation.go` is generated in the same location.

In `federation_grpc_federation.go`, an initialization function ( `NewFederationService` ) for the federation service is created. The server instance initialized using that function can be registered as a gRPC server using the `RegisterFederationService` function defined in `federation_grpc.pb.go` as is.

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
Also, there are settings for customizing error handling on method calls, etc.

# CONTRIBUTION

*NOTE*: Currently, we do not accept pull requests from anyone other than the development team. Please report feature requests or bug fix requests via Issues.

Please read the CLA carefully before submitting your contribution to Mercari. Under any circumstances, by submitting your contribution, you are deemed to accept and agree to be bound by the terms and conditions of the CLA.

https://www.mercari.com/cla/

# LICENSE

Copyright 2023 Mercari, Inc.

Licensed under the MIT License.