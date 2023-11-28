# Getting Started

Consider the example of creating a BFF ( `FederationService` ) that combines and returns the results obtained using the `PostService` and `UserService`.

This example’s architecture is following.

![arch](https://github.com/mercari/grpc-federation/assets/209884/dbd4a32d-5f63-49eb-b32b-36600c09968e)

1. End-User sends a gRPC request to `FederationService` with `post id`
2. `FederationService` sends a gRPC request to `PostService`  microservice with `post id` to get `Post` message
3. `FederationService` sends a gRPC request to `UserService` microservice with the `user_id` present in the `Post` message and retrieves the `User` message
4. `FederationService` aggregates `Post` and `User` messages and returns it to the End-User as a single message

The Protocol Buffers file definitions of `PostService` and `UserService` and `FederationService` are as follows.

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

## FederationService

`FederationService` has a `GetPost` method that aggregates the `Post` and `User` messages retrieved from `PostService` and `UserService` and returns it as a single message.

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

## 1. Add `grpc.federation.service` option

`grpc.federation.service` option is used to specify a service to generated target using gRPC Federation.  
Therefore, your first step is to import the gRPC Federation proto file and add the service option.

```diff
+import "grpc/federation/federation.proto";

 service FederationService {
+  option (grpc.federation.service) = {};
   rpc GetPost (GetPostRequest) returns (GetPostReply) {}
 }
```

## 2. Add `grpc.federation.message` option to the response message

gRPC Federation focuses on the response message of the gRPC method.
Therefore, add an option to the `GetPostReply` message, which is the response message of the `GetPost` method, to describe how to construct the response message.

- federation.proto

```diff
 message GetPostReply {
+  option (grpc.federation.message) = {};
   Post post = 1;
 }
```

## 3. Use `def` feature of `grpc.federation.message` option

In the gRPC Federation, `grpc.federation.message` option creates a variable and `grpc.federation.field` option refers to that variable and assigns a value to the field.

So first, we use [`def`](https://github.com/mercari/grpc-federation/blob/main/docs/references.md#grpcfederationmessagedef) to define variables.

- federation.proto

```diff
 message GetPostReply {
   option (grpc.federation.message) = {
+    def {
+      name: "p"
+      message {
+        name: "Post"
+        args { name: "pid", by: "$.post_id" }
+      }
+    }
   };
   Post post = 1;
 }
```

The above definition is equivalent to the following pseudo Go code.

```go
// getGetPostReply returns GetPostReply message by GetPostRequest message.
func getGetPostReply(req *pb.GetPostRequest) *pb.GetPostReply {
    p := getPost(&PostArgument{Pid: req.GetPostId()})
    ...
}

// getPost returns Post message by PostArgument.
func getPost(arg *PostArgument) *pb.Post {
    postID := arg.Pid
    ...
}
```

- `name: "p"`: It means to create a variable named `p`
- `message {}`: It menas to get message instance
- `name: "Post"`: It means to get `Post` message in `federation` package.
- `args: {name: "pid", by: "$.post_id"}`: It means to retrieve the Post message, to pass as an argument a value whose name is `pid` and whose value is `$.post_id`.

`$.post_id` indicates a reference to a message argument. The message argument for a `GetPostReply` message is a `GetPostRequest` message. Therefore, the `"$."` can be used to refer to each field of `GetPostRequest` message.

For more information on each feature, please refer to the [API Reference](./references.md)

## 4. Add `grpc.federation.field` option to the response message

Assigns the value using `grpc.federation.field` option to a field ( field binding ).
`p` variable type is a `Post` message type and `Post post = 1` field also `Post` message type. Therefore, it can be assigned as is without type conversion.

```diff
 message GetPostReply {
   option (grpc.federation.message) = {
     def {
       name: "p"
       message {
         name: "Post"
         args { name: "pid", by: "$.post_id" }
       }
     }
   };
+  Post post = 1 [(grpc.federation.field).by = "p"];
 }
```

## 5. Add `grpc.federation.message` option to the Post message

To create `GetPostReply` message, `Post` message is required. Therefore, it is necessary to define how to create `Post` message, by adding an gRPC Federation's option to `Post` message as in `GetPostReply` message.


```proto
message GetPostReply {
  option (grpc.federation.message) = {
    def {
      name: "p"
      message {
        name: "Post"
        args { name: "pid", by: "$.post_id" }
      }
    }
  };
  Post post = 1 [(grpc.federation.field).by = "p"];
}

message Post {
  option (grpc.federation.message) = {
    // call post.PostService/GetPost method with post_id and binds the response message to `res` variable
    def {
      name: "res"
      call {
        method: "post.PostService/GetPost"
        request { field: "post_id", by: "$.pid" }
      }
    }

    // refer to `res` variable and access post field.
    // Use autobind feature with the retrieved value.
    def {
      by: "res.post"
      autobind: true
    }
  };

  string id = 1; // binds the value of `res.post.id` by autobind feature
  string title = 2; // binds the value of `res.post.title` by autobind feature
  string content = 3; // binds the value of `res.post.content` by autobind feature

  User user = 4; // TODO
}
```

In `def`, besides getting the message, you can call the gRPC method and assign the result to a variable, or get another value from the value of a variable and assign it to a new variable.  

The first `def` in the `Post` message calls `post.PostService`'s `GetPost` method and assigns the result to the `res` variable.  

The second `def` in the `Post` message access `post` field of `res` variable and use `autobind` feature for easy field binding.

> [!TIP]
> [`autobind`](https://github.com/mercari/grpc-federation/blob/main/docs/references.md#grpcfederationmessagedefautobind)
>
> If the defined value is a message type and the field of that message type exists in the message with the same name and type, the field binding is automatically performed. If multiple autobinds are used at the same message, you must explicitly use the grpc.federation.field option to do the binding yourself, since duplicate field names cannot be correctly determined as one.

## 6. Add gRPC Federation option to the User message

Finally, since `Post` message depends on `User` message, add an option to `User` message.

```proto
message Post {
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "post.PostService/GetPost"
        request { field: "post_id", by: "$.pid" }
      }
    }
    def {
      // the value of `res.post` assigns to `post` variable.
      name: "post"
      by: "res.post"
      autobind: true
    }

    // get User message and assign it to `u` variable.
    // The `post` variable is referenced to retrieve the `user_id`, and the value is named `uid` as an argument for User message.
    def {
      name: "u"
      message {
        name: "User"
        args { name: "uid", by: "post.user_id" }
      }
    }
  };

  string id = 1;
  string title = 2;
  string content = 3;

  // binds `u` variable to user field.
  User user = 4 [(grpc.federation.field).by = "u"];
}

message User {
  option (grpc.federation.message) = {
    def [
      {
        name: "res"
        call {
          method: "user.UserService/GetUser"
          // refer to message arguments with `$.uid`
          request { field: "user_id", by: "$.uid" }
        }
      },
      {
        by: "res.user"
        autobind: true
      }
    ]
  };

  string id = 1;
  string name = 3;
  int age = 3;  
}
```

## 7. Run code generator

The final completed proto definition will look like this.

- federation.proto

```proto
package federation;

import "grpc/federation/federation.proto";
import "post.proto";
import "user.proto";

service FederationService {
  option (grpc.federation.service) = {};
  rpc GetPost (GetPostRequest) returns (GetPostReply) {}
}

message GetPostRequest {
  string id = 1;
}

message GetPostReply {
  option (grpc.federation.message) = {
    def {
      name: "p"
      message {
        name: "Post"
        args { name: "pid", by: "$.post_id" }
      }
    }
  };
  Post post = 1 [(grpc.federation.field).by = "p"];
}

message Post {
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "post.PostService/GetPost"
        request { field: "post_id", by: "$.pid" }
      }
    }
    def {
      name: "post"
      by: "res.post"
      autobind: true
    }
    def {
      name: "u"
      message {
        name: "User"
        args { name: "uid", by: "post.user_id" }
      }
    }
  };

  string id = 1;
  string title = 2;
  string content = 3;
  User user = 4 [(grpc.federation.field).by = "u"];
}

message User {
  option (grpc.federation.message) = {
    def [
      {
        name: "res"
        call {
          method: "user.UserService/GetUser"
          request { field: "user_id", by: "$.uid" }
        }
      },
      {
        by: "res.user"
        autobind: true
      }
    ]
  };

  string id = 1;
  string name = 3;
  int age = 3;  
}
```

Next, generates gRPC server codes using by this `federation.proto`.

First, install `grpc-federation-generator`

```console
go install github.com/mercari/grpc-federation/cmd/grpc-federation-generator@latest
```

Puts `federation.proto`, `post.proto`, `user.proto` files to under the `proto` directory.
Also, write `grpc-federation.yaml` file to run generator.

- grpc-federation.yaml

```yaml
imports:
  - proto
src:
  - proto
out: .
```

Run code generator by the following command.

```console
grpc-federation-generator ./proto/federation.proto
```

## 8. Use generated server

Running code generation using the `federation.proto` will create a `federation_grpc_federation.pb.go` file under the output path.

In `federation_grpc_federation.pb.go`, an initialization function ( `NewFederationService` ) for the `FederationService` is created. The server instance initialized using that function can be registered as a gRPC server using the `RegisterFederationService` function defined in `federation_grpc.pb.go` as is.

When initializing, you need to create a dedicated `Config` structure and pass.

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

## 9. Other Examples

[A sample that actually works can be found here](../_examples/)





