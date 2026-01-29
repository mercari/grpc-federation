# gRPC Federation Feature References

# grpc.federation.service

The `grpc.federation.service` option is **always required** to enable gRPC Federation. Please specify as follows.

```proto
service MyService {
  option (grpc.federation.service) = {};
}
```

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`env`](#grpcfederationserviceenv) | Env | optional |
| [`var`](#grpcfederationservicevar) | repeated ServiceVariable | optional |

## (grpc.federation.service).env

The definition of environment variables used by the service.
The environment variables defined here can be accessed through the `Get<service-name>Env()` function in the custom resolver. They can also be used by accessing the `grpc.federation.env` variable in CEL expressions.
When using `grpc.federation.env` to access environment variables, the message using the option must exist in the same package as the service that defines the `env` option.

Keep in mind that multiple services within the same package can use the `env` option. However, if a message is employed by multiple services with differing `env` options, that message can only access the set of environment variables that are common to those services.
For instance, suppose ServiceA has two environment variables, `ENV_VAR1` and `ENV_VAR2`, while ServiceB has just one, `ENV_VAR1`. If both services reference `Message1`, users can only access `ENV_VAR1` from `Message1`, as ServiceB does not define `ENV_VAR2` in its environment options.

`message` and `var` cannot be used simultaneously.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`message`](#grpcfederationserviceenvmessage) | string | optional |
| [`var`](#grpcfederationserviceenvvar) | repeated EnvVar | optional |

## (grpc.federation.service).env.message

If there is a message that represents a collection of environment variables, you can specify that message.

### Example

If the following environment variables exist, they can be interpreted and handled as an `Env` message.

- `FOO=hello`
- `BAR=1`

```proto
service MyService {
  option (grpc.federation.service) = {
    env { message: "Env" }
  };
}

message Env {
  string foo = 1;
  int64 bar = 2;
}
```

## (grpc.federation.service).env.var

The definition of the environment variable.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`name`](#grpcfederationserviceenvvarname) | string | required |
| [`type`](#grpcfederationserviceenvvartype) | EnvType | required |
| [`option`](#grpcfederationserviceenvvaroption) | EnvVarOption | optional |

## (grpc.federation.service).env.var.name

The name of the environment variable. It is case insensitive.

## (grpc.federation.service).env.var.type

This represents the types of environment variables.
Available types include primitive types, their repeated types, and map types.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`kind`](#grpcfederationserviceenvvartypekind) | EnvKind | optional |
| [`repeated`](#grpcfederationserviceenvvartyperepeated) | EnvType | optional |
| [`map`](#grpcfederationserviceenvvartypemap) | EnvMapType | optional |

## (grpc.federation.service).env.var.type.kind

It is used to represent the following primitive types.

- `string`: `STRING`
- `bool`: `BOOL`
- `int64`: `INT64`
- `uint64`: `UINT64`
- `double`: `DOUBLE`
- `google.protobuf.Duration`: `DURATION`

### Example

The following example represents the equivalent of the `int64` type.

```proto
service MyService {
  option (grpc.federation.service) = {
    env {
      var {
        name: "i64"
        type { kind: INT64 }
      }
    }
  };
}
```

## (grpc.federation.service).env.var.type.repeated

It is used to represent the `repeated` type.

### Example

The following example represents the equivalent of the `repeated string` type.

```proto
service MyService {
  option (grpc.federation.service) = {
    env {
      var {
        name: "repeated_value"
        type {
          repeated {
            kind: STRING
          }
        }
      }
    }
  };
}
```

## (grpc.federation.service).env.var.type.map

It is used to represent a map type, specifying `EnvType` values for both key and value.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`key`](#grpcfederationserviceenvvartype) | EnvType | required |
| [`value`](#grpcfederationserviceenvvartype) | EnvType | required |

### Example

The following example represents the equivalent of the `map<string, int64>` type.

```proto
service MyService {
  option (grpc.federation.service) = {
    env {
      var {
        name: "map"
        type {
          map {
            key { kind: STRING }
            value { kind: INT64 }
          }
        }
      }
    }
  };
}
```

## (grpc.federation.service).env.var.option

These are options for modifying the behavior when reading environment variables.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`alternate`](#grpcfederationserviceenvvaroptionalternate) | string | optional |
| [`default`](#grpcfederationserviceenvvaroptiondefault) | string | optional |
| [`required`](#grpcfederationserviceenvvaroptionrequired) | bool | optional |
| [`ignored`](#grpcfederationserviceenvvaroptionignored) | bool | optional |

## (grpc.federation.service).env.var.option.alternate

If you want to refer to an environment variable by a different name than specified by `var.name`, use this option.

## (grpc.federation.service).env.var.option.default

If you have a value you want to specify in case the environment variable does not exist, use this option.

## (grpc.federation.service).env.var.option.required

Make the existence of the environment variable mandatory.

## (grpc.federation.service).env.var.option.ignored

No action is taken even if the environment variable exists.

## (grpc.federation.service).var

This option defines variables at the service level.
This definition is executed at server startup, after the initialization of environment variables.
The defined variables can be used across all messages that the service depends on.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`name`](#grpcfederationservicevarname) | string | optional |
| [`if`](#grpcfederationservicevarif) | [CEL](./cel.md) | optional |
| [`by`](#grpcfederationservicevarby) | [CEL](./cel.md) | optional |
| [`map`](#grpcfederationservicevarmap) | MapExpr | optional |
| [`message`](#grpcfederationservicevarmessage) | MessageExpr | optional |
| [`enum`](#grpcfederationservicevarenum) | EnumExpr | optional |
| [`switch`](#grpcfederationservicevarswitch) | SwitchExpr | optional |
| [`validation`](#grpcfederationservicevarvalidation) | ServiceVariableValidationExpr | optional |

### Example

```proto
service MyService {
  option (grpc.federation.service) = {
    env {
      var {
        // version is '1.0.0'.
        name: "version"
        type { kind: STRING }
      }
    }
    var {
      // modified_version is 'v1.0.0'.
      name: "modified_version"
      by: "'v' + grpc.federation.env.version"
    }
  };
  rpc Get(GetRequest) returns (GetResponse) {};
}

message GetResponse {
  option (grpc.federation.message) = {
    def {
      // refer modified_version variable with `grpc.federation.var.` prefix.
      if: "grpc.federation.var.modified_version == 'v1.0.0'"
    }
  };
}
```

## (grpc.federation.service).var.name

The `name` is a variable name.
This name can be referenced in all CELs related to the service by using `grpc.federation.var.` prefix.

## (grpc.federation.service).var.if

If `if` is defined, then it is evaluated only if the condition is matched.
If does not evaluated, a default value of the type obtained when evaluating is assigned.

## (grpc.federation.service).var.by

Binds the result of evaluating the [CEL](./cel.md) defined in `by` to the variable defined in name.

## (grpc.federation.service).var.map

Please see [(grpc.federation.message).def.map](#grpcfederationmessagedefmap).

## (grpc.federation.service).var.message

Please see [(grpc.federation.message).def.message](#grpcfederationmessagedefmessage).

## (grpc.federation.service).var.enum

Please see [(grpc.federation.message).def.enum](#grpcfederationmessagedefenum).

## (grpc.federation.service).var.switch

Please see [(grpc.federation.message).def.switch](#grpcfederationmessagedefswitch).

## (grpc.federation.service).var.validation

A validation rule against variables defined within the current scope.

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`if`](#grpcfederationservicevarvalidationif) | [CEL](./cel.md) | required |
| [`message`](#grpcfederationservicevarvalidationmessage) | [CEL](./cel.md) | required |

## (grpc.federation.service).var.validation.if

A validation rule in [CEL](./cel.md). If the condition is true, the validation returns the error.
The return value must always be of type boolean.

## (grpc.federation.service).var.validation.message

A gRPC status message in case of validation error.

# grpc.federation.method

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`timeout`](#grpcfederationmethodtimeout) | string | optional |
| [`response`](#grpcfederationmethodresponse) | string | optional |

## (grpc.federation.method).timeout

the time to timeout. If the specified time period elapses, DEADLINE_EXCEEDED status is returned.  
If you want to handle this error, you need to implement a custom error handler in Go.  
The format is the same as Go's time.Duration format. See https://pkg.go.dev/time#ParseDuration.

### Example

```proto
service MyService {
  option (grpc.federation.service) = {};
  rpc Get(GetRequest) returns (GetResponse) {
    option (grpc.federation.method).timeout = "1m";
  };
}
```

## (grpc.federation.method).response

`response` specify the name of the message you want to use to create the response value.
If you specify a reserved type like `google.protobuf.Empty` as the response, you cannot define gRPC Federation options.
In such cases, you can specify a separate message to create the response value.
The specified response message must contain fields with the same names and types as all the fields in the original response.

### Example

```proto
service MyService {
  option (grpc.federation.service) = {};
  rpc Update(UpdateRequest) returns (google.protobuf.Empty) {
    option (grpc.federation.method).response = "UpdateResponse";
  };
}

message UpdateRequest {
  string id = 1;
}

message UpdateResponse {
  option (grpc.federation.message) = {
    def {
      call {
        method: "pkg.PostService/UpdatePost"
        request { field: "id" by: "$.id" }
      }
    }
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

### Example

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
| [`noalias`](#grpcfederationenum_valuenoalias) | bool | optional |
| [`attr`](#grpcfederationenum_valueattr) | repeated EnumValueAttribute | optional |

## (grpc.federation.enum_value).default

Specifies the default value of the enum.
All values other than those specified in alias will be default values.

### Example

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

### Example

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

## (grpc.federation.enum_value).noalias

`noalias` exclude from the target of `alias`.
This option cannot be specified simultaneously with `default` or `alias`.

### Example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

enum FooEnum {
  option (grpc.federation.enum).alias = "foopkg.FooEnum";

  ENUM_VALUE_1 = 0 [(grpc.federation.enum_value).alias = "FOO_VALUE_1"];
  ENUM_VALUE_2 = 1 [(grpc.federation.enum_value).noalias = true];
}
```

## (grpc.federation.enum_value).attr

`attr` is used to hold multiple name-value pairs corresponding to an enum value.
The values specified by the name must be consistently specified for all enum values.
The values stored using this feature can be retrieved using the [`attr()`](./cel/enum.md#enum-fqdnattr) method of the enum API.

### Example

```proto
package mypkg;

message Foo {
  option (grpc.federation.message) = {
    def { name: "v" by: "Type.value('VALUE_1')" }
  };
  string text = 1 [(grpc.federation.field).by = "Type.attr(v, 'attr_x')"];
}

enum Type {
  VALUE_1 = 1 [(grpc.federation.enum_value).attr = {
    name: "attr_x"
    value: "foo"
  }];
  VALUE_2 = 2 [(grpc.federation.enum_value).attr = {
    name: "attr_x"
    value: "bar"
  }];
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
| [`enum`](#grpcfederationmessagedefenum) | EnumExpr | optional |
| [`map`](#grpcfederationmessagedefmap) | MapExpr | optional |
| [`switch`](#grpcfederationmessagedefswitch) | SwitchExpr | optional |
| [`validation`](#grpcfederationmessagedefvalidation) | ValidationExpr | optional |


## (grpc.federation.message).def.autobind

If the defined value is a message type and the field of that message type exists in the message with the same name and type,
the field binding is automatically performed.
If multiple autobinds are used at the same message,
you must explicitly use the `grpc.federation.field` option to do the binding yourself, since duplicate field names cannot be correctly determined as one.

### Example

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    def {
      call { method: "foopkg.FooService/GetFoo" }

      // The value returned by `call` is `foopkg.GetFooReply` message.
      // In this case, if a field with the same name and type as the field in the `GetFooReply` message exists in `MyMessage`,
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

### Example

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

### Example

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

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      // The type of "res" variable is `foopkg.FooService/GetFoo` method's response type.
      name: "res"
      call {
        method: "foopkg.FooService/GetFoo"
        request { field: "field1", by: "'abcd'" }
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
| [`error`](#grpcfederationmessagedefcallerror) | repeated GRPCError | optional |

## (grpc.federation.message).def.call.method

Specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`.

### Example

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
The `field` corresponds to the field name in the request message, and the `by` specifies which value is associated with the field.

### Example

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
          { field: "field1", by: "'hello'" }
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

## (grpc.federation.message).def.call.request.field

The field name of the request message.

## (grpc.federation.message).def.call.request.by

`by` used to refer to a variable or message argument defined in a `grpc.federation.message` option by [CEL](./cel.md).
You need to use `$.` to refer to the message argument.

## (grpc.federation.message).def.call.timeout

The time to timeout. If the specified time period elapses, DEADLINE_EXCEEDED status is returned.  
If you want to handle this error, you need to implement a custom error handler in Go.  
The format is the same as Go's time.Duration format. See https://pkg.go.dev/time#ParseDuration.   

### Example

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
| [`if`](#grpcfederationmessagedefcallretryif) | [CEL](./cel.md) | optional |
| [`constant`](#grpcfederationmessagedefcallretryconstant) | RetryPolicyConstant | optional |
| [`exponential`](#grpcfederationmessagedefcallretryexponential) | RetryPolicyExponential | optional |

## (grpc.federation.message).def.call.retry.if

`if` specifies condition in CEL. If the condition is `true`, run the retry process according to the policy.
If this field is omitted, it is always treated as `true` and run the retry process.
The return value must always be of type `boolean`.

> [!NOTE]
> Within the `if`, the `error` variable can be used as a gRPC error variable when evaluating CEL.
> [A detailed description of this variable is here](./cel.md#error)

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"
        retry {
          if: "error.code != google.rpc.Code.UNIMPLEMENTED"
          constant {
            interval: "10s"
          }
        }
      }
    }
  };
}
```

## (grpc.federation.message).def.call.retry.constant

Retry according to the "constant" policy.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`inerval`](#grpcfederationmessagedefcallretryconstantinterval) | string | optional |
| [`max_retries`](#grpcfederationmessagedefcallretryconstantmax_retries) | uint64 | optional |

## (grpc.federation.message).def.call.retry.constant.interval

Interval value. Default value is `1s`.

### Example

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

### Example

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

### Example

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

### Example

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

### Example

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

### Example

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

### Example

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

## (grpc.federation.message).def.call.error

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"

        // If an error is returned when the gRPC method is called, the error block is evaluated.
        // If there are multiple error blocks, they are processed in order from the top.
        // The first error matching the condition is returned.
        // If none of the conditions match, the original error is returned.
        error {
          // The variables can be defined for use in error blocks
          def { name: "a" by: "1" }

          if: "a == 1" // if true, returns the error.
          code: FAILED_PRECONDITION // gRPC error code. this field is required.
          message: "'this is error message'" // gRPC error message. this field is required.

          // If you want to add the more information to the gRPC error, you can specify it in the details block.
          details {
            // The variables can be defined for use in details block.
            def { name: "b" by: "2" }

            if: "b == 2" // if true, add to the error details.

            // You can specify the value defined in `google/rpc/error_details.proto`.
            // FYI: https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto
            precondition_failure {
              violations {
                type: "'type value'"
                subject: "'subject value'"
                description: "'desc value'"
              }
            }

            // If you want to add your own message, you can add it using the message name and arguments as follows.
            message {
              name: "CustomMessage"
              args { name: "msg" by: "id" }
            }
          }
        }

        // This error block is evaluated if the above condition is false.
        error {
          // If `ignore: true` is specified, the error is ignored.
          ignore: true
        }
      }
    }
  };
}
```

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`def`](#grpcfederationmessagedef) | repeated VariableDefinition | optional |
| [`if`](#grpcfederationmessagedefcallerrorif) | [CEL](./cel.md) | optional |
| [`code`](#grpcfederationmessagedefcallerrorcode) | [google.rpc.Code](../proto_deps/google/rpc/code.proto) | required |
| [`message`](#grpcfederationmessagedefcallerrormessage) | [CEL](./cel.md) | required |
| `details` | repeated GRPCErrorDetail | optional |
| [`ignore`](#grpcfederationmessagedefcallerrorignore) | bool | optional |
| [`ignore_and_response`](#grpcfederationmessagedefcallerrorignore_and_response) | [CEL](./cel.md) | optional |

> [!NOTE]
> Within the error block, the `error` variable can be used as a special variable when evaluating CEL.
> [A detailed description of this variable is here](./cel.md#error)

## (grpc.federation.message).def.call.error.def
`def` defines a variable scoped to the entire error block.

 It is important to note that definitions scoped at the top level of the error block will be evaluated before `error.if`.

## (grpc.federation.message).def.call.error.if

`if` specifies condition in CEL. If the condition is `true`, it returns defined error information.
If this field is omitted, it is always treated as `true` and returns defined error information.
The return value must always be of type `boolean`.

## (grpc.federation.message).def.call.error.code

`code` is a gRPC status code.

## (grpc.federation.message).def.call.error.message

`message` is a gRPC status message.
If omitted, the message will be auto-generated from the configurations.

## (grpc.federation.message).def.call.error.ignore

`ignore` ignore the error if the condition in the `if` field is `true` and `ignore` field is set to `true`.
When an error is ignored, the returned response is always empty value.
If you want to return a response that is not empty, please use `ignore_and_response` feature.
Therefore, `ignore` and `ignore_and_response` cannot be specified same.

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"

        error {
          // dscribe the condition for ignoring error.
          if: "error.code == google.rpc.Code.UNAVAILABLE"
          ignore: true
        }
      }
    }
  };
}
```

If you want to define custom processing using the original error value when an error is ignored,
you can retrieve the contents of the error by calling the `ignoredError()` method on the return value.
Additionally, you can check whether `ignoredError()` can be used by calling `hasIgnoredError()`.

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      name: "res"
      call {
        method: "foopkg.FooService/GetFoo"

        error {
          // dscribe the condition for ignoring error.
          if: "error.code == google.rpc.Code.UNAVAILABLE"
          ignore: true
        }
      }
    }
    def {
      // It returns true if the error was being ignored.
      if: "res.hasIgnoredError()"

      // Binds error code to code variable ( google.rpc.Code.UNAVAILABLE )
      name: "code"

      // Get original error variable by ignoredError method.
      by: "res.ignoredError().code"
    }
  };
}
```

## (grpc.federation.message).def.call.error.ignore_and_response

`ignore_and_response` ignore the error if the condition in the `if` field is `true` and it returns response specified in CEL.
The evaluation value of CEL must always be the same as the response message type.
`ignore` and `ignore_and_response` cannot be specified same.

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      call {
        method: "foopkg.FooService/GetFoo"

        error {
          // dscribe the condition for ignoring error.
          if: "error.code == google.rpc.Code.UNAVAILABLE"

          // specify the response to return when errors are ignored.
          // The return value must always be a response message.
          ignore_and_response: "foopkg.Foo{x: 1}"
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

### Example

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

### Example

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
          { name: "x" by: "'xxxx'" },
          { name: "y" by: "true" }
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

## (grpc.federation.message).def.enum

`enum` is a feature designed to create enum values defined in the federation package. If you want to create enum values in the federation package based on enum values defined in a dependency package, this feature will be useful.

### Example

In this example, the value of `dep.Type.TYPE_1` is converted to the type `mypkg.MyType`. If the conversion fails, it is logged.

```proto
package mypkg;

message MyMessage {
  option (grpc.federation.message) = {
    def {
      name: "v"
      enum {
        name: "MyType"
        by: "dep.Type.TYPE_1"
      }
    }
    def {
      if: "v == MyType.TYPE_UNSPECIFIED"
      by: "grpc.federation.log.warn('got unexpected type')"
    }
  };
}

enum MyType {
  option (grpc.federation.enum).alias = "dep.Type";

  TYPE_UNSPECIFIED = 1 [(grpc.federation.enum_value).default = true];
  TYPE_1 = 2;
  TYPE_2 = 3;
}
```

```proto
package dep;

enum Type {
  TYPE_1 = 1;
  TYPE_2 = 2;
}
```

## (grpc.federation.message).def.map

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`iterator`](#grpcfederationmessagedefmapiterator) | Iterator | required |
| [`by`](#grpcfederationmessagedefmapby) | [CEL](./cel.md) | optional |
| [`message`](#grpcfederationmessagedefmapmessage) | MessageExpr | optional |
| [`enum`](#grpcfederationmessagedefmapenum) | EnumExpr | optional |

## (grpc.federation.message).def.map.iterator

Create the iterator variable.

- `src` must be a repeated type
- `name` defines the name of the iterator variable

### Example

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

## (grpc.federation.message).def.map.enum

Create map elements using `enum` value by referencing variables created with `iterator` section.

## (grpc.federation.message).def.switch

`switch` evaluates cases in order and returns the value from the first matching case, or the default case, if no case matches.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`case`](#grpcfederationmessagedefswitchcase) | repeated SwitchCaseExpr | optional |
| [`default`](#grpcfederationmessagedefswitchdefault) | SwitchDefaultExpr | required |

### Example

```proto
message MyMessage {
  option (.grpc.federation.message) = {
    def {
      name: "status"
      switch {
        case { 
          if: "$.activation_status == 'active'" 
          by: "'enabled'" 
        }
        case { 
          if: "$.activation_status == 'inactive'" 
          by: "'disabled'" 
        }
        default { by: "'unknown'" }
      }
    }
  };
  string status = 1 [(grpc.federation.field).by = "status"];
}
```

## (grpc.federation.message).def.switch.case

A single case in a `switch` expression. Cases are evaluated in order, and the first case whose `if` condition evaluates to `true` will have its `by` expression evaluated and the resulting value returned as the value of the `switch`.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`if`](#grpcfederationmessagedefswitchcaseif) | [CEL](./cel.md) | required |
| [`by`](#grpcfederationmessagedefswitchcaseby) | [CEL](./cel.md) | required |

## (grpc.federation.message).def.switch.case.if

A condition in [CEL](./cel.md) that determines whether this case matches. The return value must always be of type boolean. If the condition evaluates to `true`, the case matches and its `by` expression is evaluated.

## (grpc.federation.message).def.switch.case.by

A [CEL](./cel.md) expression that is evaluated when this case matches. The result of this expression is returned as the value of the `switch` expression.

## (grpc.federation.message).def.switch.default

The default case that is evaluated when none of the switch cases match.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`by`](#grpcfederationmessagedefswitchdefaultby) | [CEL](./cel.md) | required |

## (grpc.federation.message).def.switch.default.by

A [CEL](./cel.md) expression that is evaluated when no switch case matches. The result of this expression is returned as the value of the `switch` expression.

## (grpc.federation.message).def.validation

A validation rule against variables defined within the current scope.

| field                                               | type            | required or optional |
|-----------------------------------------------------|-----------------|----------------------|
| [`name`](#grpcfederationmessagedefvalidationname)   | string          | optional             |
| [`error`](#grpcfederationmessagedefvalidationerror) | ValidationError | required             |

## (grpc.federation.message).def.validation.name
A unique name for the validation.
If set, the validation error type will be <message-name><name>Error.
If omitted, the validation error type will be ValidationError.

## (grpc.federation.message).def.validation.error

A validation rule and validation error to be returned.

| field                                                        | type                           | required or optional |
|--------------------------------------------------------------|--------------------------------|----------------------|
| [`code`](#grpcfederationmessagedefvalidationerrorcode)       | [google.rpc.Code](../proto_deps/google/rpc/code.proto) | required |
| [`message`](#grpcfederationmessagedefvalidationerrormessage) | string                         | optional             |
| [`if`](#grpcfederationmessagedefvalidationerrorif)           | [CEL](./cel.md)                            | optional             |
| [`details`](#grpcfederationmessagedefvalidationerrordetails) | repeated ValidationErrorDetail | optional             |

## (grpc.federation.message).def.validation.error.code
A gRPC status code to be returned in case of validation error. For the available Enum codes, check [googleapis/google/rpc /code.proto](https://github.com/googleapis/googleapis/blob/89b562b76f5b215990a20d3ea08bc6e1c0377906/google/rpc/code.proto#L32-L186).

## (grpc.federation.message).def.validation.error.message
A gRPC status message in case of validation error. If omitted, the message will be auto-generated from the configurations.

## (grpc.federation.message).def.validation.error.if
A validation rule in [CEL](./cel.md). If the condition is true, the validation returns the error.
The return value must always be of type boolean. Either `if` or `details` must be specified.

### Example

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      validation {
        name: "myMessageValidation",
        error {
          code: FAILED_PRECONDITION
          message: "MyMessage validation failed",
          if: "$.id == 'wrong'"
        }
      }
    }
  };
  ...
}
```

## (grpc.federation.message).def.validation.error.details

`details` is a list of validation rules and error details. If the validation fails, the corresponding error details are set.
Either `if` or `details` must be specified. The other error detail types will be supported soon.

| field                                                                                        | type                                    | required or optional |
|----------------------------------------------------------------------------------------------|-----------------------------------------|----------------------|
| [`if`](#grpcfederationmessagedefvalidationerrordetailsif)                                    | [CEL](./cel.md)                                     | required             |
| [`by`](#grpcfederationmessagedefvalidationerrordetailsby)                                    | repeated [CEL](./cel.md)                         | optional             |
| [`message`](#grpcfederationmessagedefvalidationerrordetailsmessage)                          | repeated MessageExpr                    | optional             |
| [`precondition_failure`](#grpcfederationmessagedefvalidationerrordetailspreconditionfailure) | repeated google.rpc.PreconditionFailure | optional             |
| [`bad_request`](#grpcfederationmessagedefvalidationerrordetailsbadrequest)                   | repeated google.rpc.BadRequest          | optional             |
| [`localized_message`](#grpcfederationmessagedefvalidationerrordetailslocalizedmessage)       | repeated google.rpc.LocalizedMessage    | optional             |

```proto
message MyMessage {
  option (grpc.federation.message) = {
    def {
      validation {
        name: "myMessageValidation",
        error {
          code: FAILED_PRECONDITION
          message: "MyMessage validation failed",
          details {
            if: "$.id == 'wrong'",
            by: "pkg.Message{field: value}"
            message {
              name: "ErrorMessage",
              args {
                name: "message",
                by: "'some message'"
              }
            }
            precondition_failure {
              violations {
                type: "'some-type'"
                subject: "'some-subject'"
                description: "'some-desc'"
              }
            }
            bad_request {
              field_violations {
                field: "'some-field'"
                description: "'some-desc'"
              }
            }
            localized_message {
              locale: "en-US"
              message: "'some message'"
            }
          }
        }
      }
    }
  };
  ...
}
```

## (grpc.federation.message).def.validation.error.details.if

`if` specifies validation rule in [CEL](./cel.md). If the condition is true, the validation returns an error with the specified details.

## (grpc.federation.message).def.validation.error.details.by

`by` specify a message in [CEL](./cel.md) to express the details of the error.

## (grpc.federation.message).def.validation.error.details.message

`message` represents arbitrary proto messages to describe the error detail.

## (grpc.federation.message).def.validation.error.details.precondition_failure

`precondition_failure` describes what preconditions have failed. See [google.rpc.PreconditionFailure](https://github.com/googleapis/googleapis/blob/89b562b76f5b215990a20d3ea08bc6e1c0377906/google/rpc/error_details.proto#L144) for the details.

## (grpc.federation.message).def.validation.error.details.bad_request

`bad_request` describes violations in a client request. See [google.rpc.BadRequest](https://github.com/googleapis/googleapis/blob/89b562b76f5b215990a20d3ea08bc6e1c0377906/google/rpc/error_details.proto#L170) for the details.

## (grpc.federation.message).def.validation.error.details.localized_message

`localized_message` provides a localized error message that is safe to return to the user. See [google.rpc.BadRequest](https://github.com/googleapis/googleapis/blob/89b562b76f5b215990a20d3ea08bc6e1c0377906/google/rpc/error_details.proto#L277) for the details.

## (grpc.federation.message).custom_resolver

There is a limit to what can be expressed in proto, so if you want to execute a process that cannot be expressed in proto, you will need to implement it yourself in the Go language.  

If custom_resolver is true, the resolver for this message is implemented by Go.
If there are any values retrieved by `def`, they are passed as arguments for custom resolver.
Each field of the message returned by the custom resolver is automatically bound.
If you want to change the binding process for a particular field, set `custom_resolver=true` option for that field.

### Example

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

# Message Argument

Message argument is an argument passed when depends on other messages via the `def.message` feature. This parameter propagates the information necessary to resolve message dependencies.

Normally, only values explicitly specified in `args` can be referenced as message arguments.

However, exceptionally, a field in the gRPC method's request message can be referenced as a message argument to the method's response.

For example, consider the `GetPost` method of the `FederationService`: the `GetPostReply` message implicitly allows the fields of the `GetPostRequest` message to be used as message arguments.

In summary, it looks like this.

1. For the response message of rpc, the request message fields are the message arguments.
2. For other messages, the parameters explicitly passed in `def.message.args` option are the message arguments.
3. You can access to the message argument with `$.` prefix.
