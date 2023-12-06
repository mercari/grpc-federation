# gRPC Federation Feature References

# grpc.federation.service

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`dependencies`](#grpcfederationservicedependencies) | repeated ServiceDependency | optional |

## (grpc.federation.service).dependencies

`dependencies` defines a unique name for all services on which federation service depends.
The name will be used when creating the gRPC client.

### example

```proto
service MyService {
  option (grpc.federation.service).dependencies = [
    { name: "foo", service: "foopkg.FooService" },
    { name: "bar", service: "barpkg.BarService" }
  ]
}
```

### reference

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`name`](#grpcfederationservicedependenciesname) | string | optional |
| [`service`](#grpcfederationservicedependenciesservice) | string | required |

## (grpc.federation.service).dependencies.name

Name to be used when initializing the gRPC client.

### example

```proto
service MyService {
  option (grpc.federation.service).dependencies = {
    name: "foo"
    service: "foopkg.FooService"
  };
}
```

## (grpc.federation.service).dependencies.service

Service is the name of the dependent service.

### example

```proto
service MyService {
  option (grpc.federation.service).dependencies = {
    service: "foopkg.FooService"
  };
}
```

# grpc.federation.method

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`timeout`](#grpcfederationmethodtimeout) | string | optional |

## (grpc.federation.method).timeout

the time to timeout. If the specified time period elapses, DEADLINE_EXCEEDED status is returned.  
If you want to handle this error, you need to implement a custom error handler in Go.  
The format is the same as Go's time.Duration format. See https://pkg.go.dev/time#ParseDuration.

### example

```proto
service MyService {
  option (grpc.federation.service) = {};
  rpc Get(GetRequest) returns (GetResponse) {
    option (grpc.federation.method).timeout = "1m";
  };
}
```

# grpc.federation.enum

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`alias`](#grpcfederationenumalias) | string | optional |

## (grpc.federation.enum).alias

`alias` mapping between enums defined in other packages and enums defined on the federation service side.
The `alias` is the FQDN ( `<package-name>.<enum-name>` ) to the enum.
If this definition exists, type conversion is automatically performed before the enum value assignment operation.
If a enum with this option has a value that is not present in the enum specified by alias, and the alias option is not specified for that value, an error is occurred.

### example

```proto
package mypkg;

enum FooEnum {
  option (grpc.federation.enum).alias = "foopkg.FooEnum";

  ENUM_VALUE_1 = 0;
  ENUM_VALUE_2 = 1;
}
```

# grpc.federation.enum_value

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`default`](#grpcfederationenum_valuedefault) | bool | optional |
| [`alias`](#grpcfederationenum_valuealias) | repeated string | optional |

## (grpc.federation.enum_value).default

Specifies the default value of the enum.
All values other than those specified in alias will be default values.

### example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

enum FooEnum {
  option (grpc.federation.enum).alias = "foopkg.FooEnum";

  ENUM_VALUE_UNKNOWN = 0 [(grpc.federation.enum_value).default = true];
  ENUM_VALUE_1 = 1 [(grpc.federation.enum_value).alias = "FOO_VALUE_1"];
  ENUM_VALUE_2 = 2 [(grpc.federation.enum_value).alias = "FOO_VALUE_2"];
}
```

- foo.proto

```proto
package foopkg;

enum FooEnum {
  FOO_VALUE_1 = 0;
  FOO_VALUE_2 = 1;
  FOO_VALUE_3 = 2;
}
```

## (grpc.federation.enum_value).alias

`alias` can be used when alias is specified in `grpc.federation.enum` option,
and specifies the value name to be referenced among the enums specified in alias of enum option.
multiple value names can be specified for `alias`.

### example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

enum FooEnum {
  option (grpc.federation.enum).alias = "foopkg.FooEnum";

  ENUM_VALUE_1 = 0 [(grpc.federation.enum_value).alias = "FOO_VALUE_1"];
  ENUM_VALUE_2 = 1 [(grpc.federation.enum_value).alias = "FOO_VALUE_2"];
}
```

- foo.proto

```proto
package foopkg;

enum FooEnum {
  FOO_VALUE_1 = 0;
  FOO_VALUE_2 = 1;
}
```

# grpc.federation.message

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`def`](#grpcfederationmessagedef) | repeated VariableDefinition | optional |
| [`custom_resolver`](#grpcfederationmessagecustom_resolver) | bool | optional |
| [`alias`](#grpcfederationmessagealias) | string | optional |

## (grpc.federation.message).def

`def` can define variables.  
The defined variables can be referenced in subsequent [CEL](./cel.md) field.
The `name` is a variable name and `autobind` can be used when the variable is `message` type.
It can be used for automatic field binding.

If `if` is defined, then it is evaluated only if the condition is matched.
If does not evaluated, a default value of the type obtained when evaluating is assigned.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessagedefname) | string | optional |
| [`autobind`](#grpcfederationmessagedefautobind) | bool | optional |
| [`if`](#grpcfederationmessagedefif) | [CEL](./cel.md) | optional |

The value to be assigned to a variable can be created with the following features.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`by`](#grpcfederationmessagedefby) | [CEL](./cel.md) | optional |
| [`call`](#grpcfederationmessagedefcall) | CallExpr | optional |
| [`message`](#grpcfederationmessagedefmessage) | MessageExpr | optional |
| [`map`](#grpcfederationmessagedefmap) | MapExpr | optional |
| [`validation`](#grpcfederationmessagedefvalidation) | ValidationExpr | optional |


## (grpc.federation.message).def.autobind

If the defined value is a message type and the field of that message type exists in the message with the same name and type,
the field binding is automatically performed.
If multiple autobinds are used at the same message,
you must explicitly use the `grpc.federation.field` option to do the binding yourself, since duplicate field names cannot be correctly determined as one.

### example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    def {
      call { method: "foopkg.FooService/GetFoo" }

      // The value returned by `call` is `foopkg.GetReply` message.
      // In this case, if a field with the same name and type as the field in the `GetReply` message exists in `MyMessage`,
      // field binding is automatically performed.
      autobind: true
    }
  };

  string field1 = 1; // autobound from `foopkg.GetFooReply.field1`
  bool field2 = 2; // autobound from `foopkg.GetFooReply.field2`
}
```

- foo.proto

```proto
package foopkg;

service FooService {
  rpc GetFoo(GetFooRequest) returns (GetFooReply) {};
}

message GetFooReply {
  string field1 = 1;
  bool field2 = 2;
}
```

## (grpc.federation.message).def.if

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    // `x` is assigned `10`
    def {
      name: "x"
      if: "true"
      by: "10"
    }

    // `y` is assigned `0`
    def {
      name: "y"
      if: "false"
      by: "10"
    }
  };
}
```

## (grpc.federation.message).def.by

Binds the result of evaluating the [CEL](./cel.md) defined in `by` to the variable defined in name.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      // variable `v` is assigned the value `6`.
      name: "v"
      by: "1 + 2 + 3"
    }
   };
 }
```

## (grpc.federation.message).def.call

`call` is used to call the gRPC method and assign the result to a variable.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      // The type of "res" variable is `foopkg.FooService/GetFoo` method's response type.
      name: "res"
      call {
        method: "foopkg.FooService/GetFoo"
        request { field: "field1", string: "abcd" }
      }
    }
    // The type of "foo" variable is foo's field type of GetFoo's response type.
    // If the variable type is a message type, `autobind` feature can be used.
    def { name: "foo" by: "res.foo" autobind: true }
   };
 }
```

### reference

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`method`](#grpcfederationmessagedefcallmethod) | string | required |
| [`request`](#grpcfederationmessagedefcallrequest) | repeated MethodRequest | optional |
| [`timeout`](#grpcfederationmessagedefcalltimeout) | string | optional |
| [`retry`](#grpcfederationmessagedefcallretry) | RetryPolicy | optional |

## (grpc.federation.message).def.call.method

Specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`.

### example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
      }
    }
  };
}
```

- foo.proto

```proto
package foopkg;

service FooService {
  rpc GetFoo(GetFooRequest) returns (GetFooReply) {};
}
```

## (grpc.federation.message).def.call.request

Specify the request parameters for the gRPC method.
The `field` corresponds to the field name in the request message, and the `by` or `string` specifies which value is associated with the field.

### example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        request: [
          { field: "field1", string: "hello" }
        ]
      }
    }
  };
}
```

- foo.proto

```proto
package foopkg;

service FooService {
  rpc GetFoo(GetFooRequest) returns (GetFooReply) {};
}

message GetFooRequest {
  string field1 = 1;
}
```

### reference

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`field`](#grpcfederationmessagedefcallrequestfield) | string | required |
| [`by`](#grpcfederationmessagedefcallrequestby) | [CEL](./cel.md) | optional |
| [`string`](#grpcfederationmessagedefcallrequeststring) | string | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## (grpc.federation.message).def.call.request.field

The field name of the request message.

## (grpc.federation.message).def.call.request.by

`by` used to refer to a variable or message argument defined in a `grpc.federation.message` option by [CEL](./cel.md).
You need to use `$.` to refer to the message argument.

## (grpc.federation.message).def.call.request.string

Use constant string value.

## (grpc.federation.message).def.call.timeout

The time to timeout. If the specified time period elapses, DEADLINE_EXCEEDED status is returned.  
If you want to handle this error, you need to implement a custom error handler in Go.  
The format is the same as Go's time.Duration format. See https://pkg.go.dev/time#ParseDuration.   

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        timeout: "1m"
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry

Specify the retry policy if the method call fails.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`constant`](#grpcfederationmessagedefcallretryconstant) | RetryPolicyConstant | optional |
| [`exponential`](#grpcfederationmessagedefcallretryexponential) | RetryPolicyExponential | optional |

## (grpc.federation.message).def.call.retry.constant

Retry according to the "constant" policy.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`inerval`](#grpcfederationmessagedefcallretryconstantinterval) | string | optional |
| [`max_retries`](#grpcfederationmessagedefcallretryconstantmax_retries) | uint64 | optional |

## (grpc.federation.message).def.call.retry.constant.interval

Interval value. Default value is `1s`.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          constant {
            interval: "10s"
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.constant.max_retries

Max retry count. Default value is `5`. If `0` is specified, it never stops.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          constant {
            max_retries: 3
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.exponential

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`initial_interval`](#grpcfederationmessagedefcallretryexponentialinitial_interval) | string | optional |
| [`randomization_factor`](#grpcfederationmessagedefcallretryexponentialrandomization_factor) | double | optional |
| [`multiplier`](#grpcfederationmessagedefcallretryexponentialmultiplier) | double | optional |
| [`max_interval`](#grpcfederationmessagedefcallretryexponentialmax_interval) | string | optional |
| [`max_retries`](#grpcfederationmessagedefcallretryexponentialmax_retries) | uint64 | optional |

## (grpc.federation.message).def.call.retry.exponential.initial_interval

Initial interval value. Default value is `500ms`.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          exponential {
            initial_interval: "1s"
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.exponential.randomization_factor

Randomization factor value. Default value is `0.5`.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          exponential {
            randomization_factor: 1.0
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.exponential.multiplier

Multiplier. Default value is `1.5`.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          exponential {
            multiplier: 1.0
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.exponential.max_interval

Max interval value. Default value is `60s`.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          exponential {
            max_interval: "30s"
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.exponential.max_retries

Max retry count. Default value is `5`. If `0` is specified, it never stops.

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          exponential {
            max_retries: 3
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.message

`message` used to refer to other message.  
The parameters necessary to obtain a message must be passed as message arguments by `args`.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessagedefmessagename) | string | required |
| [`args`](#grpcfederationmessagedefmessageargs) | repeated Argument | optional |

## (grpc.federation.message).def.message.name

Specify the message name with FQDN. format is `<package-name>.<message-name>`.  
`<package-name>` can be omitted when referring to messages in the same package.

### example

```proto
package mypkg;

message MyMessage {
  option (grpc.federation.message) = {
    // get the `Foo` message and assign it to the `foo` variable.
    def {
      name: "foo"
      message { name: "Foo" }
    }
  };
}

message Foo {
  option (grpc.federation.message).custom_resolver = true;

  string x = 1;
}
```

## (grpc.federation.message).def.message.args

Specify the parameters needed to retrieve the message. This is called the message argument.

### example

```proto
package mypkg;

message MyMessage {
  option (grpc.federation.message) = {
    def {
      name: "foo"
      message {
        name: "Foo"
        // pass arguments to retrieve Foo message.
        args [
          { name: "x" string: "xxxx" },
          { name: "y" bool: true }
        ]
      }
    }
  };
}

message Foo {
  // $.x is `xxxx` value
  string x = 1 [(grpc.federation.field).by = "$.x"];

  // $.y is `true` value
  bool y = 2 [(grpc.federation.field).by = "$.y"];
}
```

### reference

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessagedefmessageargsname) | string | required |
| [`by`](#grpcfederationmessagedefmessageargsby) | [CEL](./cel.md) | optional |
| [`inline`](#grpcfederationmessagedefmessageargsinline) | [CEL](./cel.md) | optional |
| [`string`](#grpcfederationmessagedefmessageargsstring) | string | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## (grpc.federation.message).def.message.args.name

Name of the message argument.
Use this name to refer to the message argument.
For example, if `foo` is specified as the name, it is referenced by `$.foo` in dependent message.

## (grpc.federation.message).def.message.args.by

`by` used to refer to a variable or message argument defined in a `grpc.federation.message` option by [CEL](./cel.md).
You need to use `$.` to refer to the message argument.

## (grpc.federation.message).def.message.args.inline

`inline` like `by`, it refers to the specified value and expands all fields beyond it.
For this reason, the referenced value must always be of message type.

## (grpc.federation.message).def.message.args.string

`string` is constant string value.

## (grpc.federation.message).def.map

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`iterator`](#grpcfederationmessagedefmapiterator) | Iterator | required |
| [`by`](#grpcfederationmessagedefmapby) | [CEL](./cel.md) | optional |
| [`message`](#grpcfederationmessagedefmapmessage) | MessageExpr | optional |

## (grpc.federation.message).def.map.iterator

Create the iterator variable.

- `src` must be a repeated type
- `name` defines the name of the iterator variable

### example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      name: "foo_list"
      message { name: "FooList" }
    }
    def {
      // `list_x` value is a list of Foo.x. The type is repeated string.
      name: "list_x"
      map {
        iterator {
          // iterator variable name is `iter`.
          name: "iter"

          // refer to FooList message's `foos` field.
          src: "foo_list.foos"
        }

        // refer to `iter` variable and returns x field value.
        by: "iter.x"
      }
    }
  };
}

message FooList {
  option (grpc.federation.message).custom_resolver = true;

  repeated Foo foos = 1;
}

message Foo {
  string x = 1;
}
```

### reference

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessagedefmapiteratorname) | string | required |
| [`src`](#grpcfederationmessagedefmapiteratorsrc) | [CEL](./cel.md) | required |

## (grpc.federation.message).def.map.iterator.name

The name of the iterator variable.

## (grpc.federation.message).def.map.iterator.src

The variable of repeated type that form the basis of iterator.

## (grpc.federation.message).def.map.by

Create map elements using [`CEL`](./cel.md) by referencing variables created with `iterator` section.

## (grpc.federation.message).def.map.message

Create map elements using `message` value by referencing variables created with `iterator` section.

## (grpc.federation.message).def.validation

TODO..

## (grpc.federation.message).custom_resolver

There is a limit to what can be expressed in proto, so if you want to execute a process that cannot be expressed in proto, you will need to implement it yourself in the Go language.  

If custom_resolver is true, the resolver for this message is implemented by Go.
If there are any values retrieved by `def`, they are passed as arguments for custom resolver.
Each field of the message returned by the custom resolver is automatically bound.
If you want to change the binding process for a particular field, set `custom_resolver=true` option for that field.

### example

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

## (grpc.federation.message).alias

`alias` mapping between messages defined in other packages and messages defined on the federation service side.
The `alias` is the FQDN ( `<package-name>.<message-name>` ) to the message.
If this definition exists, type conversion is automatically performed before the field assignment operation.
If a message with this option has a field that is not present in the message specified by alias, and the alias option is not specified for that field, an error is occurred.

# grpc.federation.field

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`by`](#grpcfederationfieldby) | string | optional |
| [`oneof`](#grpcfederationfieldoneof) | FieldOneof | optional |
| [`custom_resolver`](#grpcfederationfieldcustom_resolver) | bool | optional |
| [`alias`](#grpcfederationfieldalias) | string | optional |
| [`string`](#grpcfederationfieldstring) | string | optional |

## (grpc.federation.field).by

`by` used to refer to a variable or message argument defined in a `grpc.federation.message` option by [CEL](./cel.md).
You need to use `$.` to refer to the message argument.

## (grpc.federation.field).oneof

`oneof` used to directly assign a value to a field defined by `oneof`.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`if`](#grpcfederationfieldoneofif) | [CEL](./cel.md) | optional |
| [`default`](#grpcfederationfieldoneofdefault) | bool | optional |
| [`def`](#grpcfederationfieldoneofdef) | repeated VariableDefinition | optional |
| [`by`](#grpcfederationfieldoneofby) | [CEL](./cel.md) | optional |

## (grpc.federation.field).oneof.if

`if` is the condition to be assigned to oneof field.
The return value must be of bool type.

## (grpc.federation.field).oneof.default

`default` used to assign a value when none of the other fields match any of the specified expressions.
Only one value can be defined per oneof.

## (grpc.federation.field).oneof.def

`def` defines the list of variables to which the oneof field refers.

## (grpc.federation.field).oneof.by

`by` used to refer to a variable or message argument defined in a `grpc.federation.message` and `grpc.federation.field.oneof` by [CEL](./cel.md).
You need to use `$.` to refer to the message argument.

## (grpc.federation.field).custom_resolver

If custom_resolver is true, the field binding process is to be implemented in Go.
If there are any values retrieved by `grpc.federation.message` option, they are passed as arguments for custom resolver.

## (grpc.federation.field).alias

`alias` can be used when `alias` is specified in `grpc.federation.message` option,
and specifies the field name to be referenced among the messages specified in `alias` of message option.
If the specified field has the same type or can be converted automatically, its value is assigned.

## (grpc.federation.field).string

# Message Argument

Message argument is an argument passed when depends on other messages via the `def.message` feature. This parameter propagates the information necessary to resolve message dependencies.

Normally, only values explicitly specified in `args` can be referenced as message arguments.

However, exceptionally, a field in the gRPC method's request message can be referenced as a message argument to the method's response.

For example, consider the `GetPost` method of the `FederationService`: the `GetPostReply` message implicitly allows the fields of the `GetPostRequest` message to be used as message arguments.

In summary, it looks like this.

1. For the response message of rpc, the request message fields are the message arguments.
2. For other messages, the parameters explicitly passed in `def.message.args` option are the message arguments.
3. You can access to the message argument with `$.` prefix.