# gRPC Federation Feature References

# grpc.federation.service

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`dependencies`](#grpcfederationservicedependencies) | repeated ServiceDependency | optional |

## (grpc.federation.service).dependencies

`dependencies` defines a unique name for all services on which federation service depends.
The name will be used when creating the gRPC client.

```proto
service MyService {
  option (grpc.federation.service).dependencies = [
    { name: "foo", service: "foopkg.FooService" },
    { name: "bar", service: "barpkg.BarService" }
  ]
}
```

| field | type | required or optional |
| ----- | ---- | ------------------- |
| [`name`](#grpcfederationservicedependenciesname) | string | optional |
| [`service`](#grpcfederationservicedependenciesservice) | string | required |

## (grpc.federation.service).dependencies.name

Name to be used when initializing the gRPC client.

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
| [`resolver`](#grpcfederationmessageresolver) | Resolver | optional |
| [`messages`](#grpcfederationmessagemessages) | repeated Message | optional |
| [`custom_resolver`](#grpcfederationmessagecustom_resolver) | bool | optional |
| [`alias`](#grpcfederationmessagealias) | string | optional |

## (grpc.federation.message).resolver

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      request { field: "field1", string: "abcd" }
      response { name: "res", autobind: true }
    }
   };
 }
```

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`method`](#grpcfederationmessageresolvermethod) | string | required |
| [`request`](#grpcfederationmessageresolverrequest) | repeated MethodRequest | optional |
| [`response`](#grpcfederationmessageresolverresponse) | repeated MethodResponse | optional |
| [`timeout`](#grpcfederationmessageresolvertimeout) | string | optional |
| [`retry`](#grpcfederationmessageresolverretry) | RetryPolicy | optional |

## (grpc.federation.message).resolver.method

Specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`.

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
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

## (grpc.federation.message).resolver.request

Specify the request parameters for the gRPC method.
The `field` corresponds to the field name in the request message, and the `by` or `string` specifies which value is associated with the field.

- myservice.proto

```proto
package mypkg;

import "foo.proto";

message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      request: [
        { field: "field1", string: "hello" }
      ]
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

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`field`](#grpcfederationmessageresolverrequestfield) | string | required |
| [`by`](#grpcfederationmessageresolverrequestby) | string | optional |
| [`string`](#grpcfederationmessageresolverrequeststring) | string | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## (grpc.federation.message).resolver.request.field

The field name of the request message.

## (grpc.federation.message).resolver.request.by

`by` used to refer to a name or message argument defined in a MessageRule, use `$.` to refer to the message argument.

## (grpc.federation.message).resolver.request.string

## (grpc.federation.message).resolver.response

Describe the information necessary to bind the result of the method call to the message field. Select and name the required information from the response message, and use the defined name when referencing the information in the message option.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessageresolverresponsename) | string | optional |
| [`field`](#grpcfederationmessageresolverresponsefield) | string | optional |
| [`autobind`](#grpcfederationmessageresolverresponseautobind) | bool | optional |

## (grpc.federation.message).resolver.response.name

`name` specify the unique name that can be used in `grpc.federation.message` or `grpc.federation.field` option for the same message for a specific field in the response.

## (grpc.federation.message).resolver.response.field

The field name in response message.

## (grpc.federation.message).resolver.response.autobind

autobind if the value referenced by `field` is a message type,
the value of a field with the same name and type as the field name of its own message is automatically assigned to the value of the field in the message.
If multiple autobinds are used at the same message,
you must explicitly use the `grpc.federation.field` option to do the binding yourself, since duplicate field names cannot be correctly determined as one.

## (grpc.federation.message).resolver.timeout

The time to timeout. If the specified time period elapses, DEADLINE_EXCEEDED status is returned.  
If you want to handle this error, you need to implement a custom error handler in Go.  
The format is the same as Go's time.Duration format. See https://pkg.go.dev/time#ParseDuration.   

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      timeout: "1m"
    }
  };
}
```

## (grpc.federation.message).resolver.retry

Specify the retry policy if the method call fails.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`constant`](#grpcfederationmessageresolverretryconstant) | RetryPolicyConstant | optional |
| [`exponential`](#grpcfederationmessageresolverretryexponential) | RetryPolicyExponential | optional |

## (grpc.federation.message).resolver.retry.constant

Retry according to the "constant" policy.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`inerval`](#grpcfederationmessageresolverretryconstantinterval) | string | optional |
| [`max_retries`](#grpcfederationmessageresolverretryconstantmax_retries) | uint64 | optional |

## (grpc.federation.message).resolver.retry.constant.interval

Interval value. Default value is `1s`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        constant {
          interval: "10s"
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.constant.max_retries

Max retry count. Default value is `5`. If `0` is specified, it never stops.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        constant {
          max_retries: 3
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.exponential

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`initial_interval`](#grpcfederationmessageresolverretryexponentialinitial_interval) | string | optional |
| [`randomization_factor`](#grpcfederationmessageresolverretryexponentialrandomization_factor) | double | optional |
| [`multiplier`](#grpcfederationmessageresolverretryexponentialmultiplier) | double | optional |
| [`max_interval`](#grpcfederationmessageresolverretryexponentialmax_interval) | string | optional |
| [`max_retries`](#grpcfederationmessageresolverretryexponentialmax_retries) | uint64 | optional |

## (grpc.federation.message).resolver.retry.exponential.initial_interval

Initial interval value. Default value is `500ms`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        exponential {
          initial_interval: "1s"
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.exponential.randomization_factor

Randomization factor value. Default value is `0.5`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        exponential {
          randomization_factor: 1.0
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.exponential.multiplier

Multiplier. Default value is `1.5`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        exponential {
          multiplier: 1.0
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.exponential.max_interval

Max interval value. Default value is `60s`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        exponential {
          max_interval: "30s"
        }
      }
    }
  };
}
```

## (grpc.federation.message).resolver.retry.exponential.max_retries

Max retry count. Default value is `5`. If `0` is specified, it never stops.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
      retry {
        exponential {
          max_retries: 3
        }
      }
    }
  };
}
```

## (grpc.federation.message).messages

| field | type | required or optional |
| ----- | ---- | -------------------- | ---- |
| [`name`](#grpcfederationmessagemessagesname) | string | optional |
| [`message`](#grpcfederationmessagemessagesmessage) | string | required |
| [`args`](#grpcfederationmessagemessagesargs) | repeated Argument | optional |
| [`autobind`](#grpcfederationmessagemessagesautobind) | bool | optional |

## (grpc.federation.message).messages.name

Specify a unique name for the dependent message.

## (grpc.federation.message).messages.message

Specify the message to be referred to by FQDN. format is `<package-name>.<message-name>`.  
`<package-name>` can be omitted when referring to messages in the same package.

## (grpc.federation.message).messages.args

Specify the parameters needed to retrieve the message. This is called the message argument.

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`name`](#grpcfederationmessagemessagesargsname) | string | required |
| [`by`](#grpcfederationmessagemessagesargsby) | string | optional |
| [`inline`](#grpcfederationmessagemessagesargsinline) | string | optional |
| [`string`](#grpcfederationmessagemessagesargsstring) | string | optional |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## (grpc.federation.message).messages.args.name

Name of the message argument.
Use this name to refer to the message argument.
For example, if `foo` is specified as the name, it is referenced by `$.foo` in dependent message.

## (grpc.federation.message).messages.args.by

`by` used to refer to a name or message argument defined in a MessageRule, use `$.` to refer to the message argument.

## (grpc.federation.message).messages.args.inline

inline like by, it refers to the specified value and expands all fields beyond it.
For this reason, the referenced value must always be of message type.

## (grpc.federation.message).messages.args.string

## (grpc.federation.message).messages.autobind

`autobind` the value of a field with the same name and type as the field name of this message is automatically assigned to the field value in the message.
If multiple autobinds are used at the same message, you must explicitly use the `grpc.federation.field` option to do the binding yourself, since duplicate field names cannot be correctly determined as one.

## (grpc.federation.message).custom_resolver

If custom_resolver is true, the resolver for this message is implemented by Go.
If there are any values retrieved by resolver or messages, they are passed as arguments for custom resolver.
Each field of the message returned by the custom resolver is automatically bound.
If you want to change the binding process for a particular field, set `custom_resolver=true` option for that field.

## (grpc.federation.message).alias

`alias` mapping between messages defined in other packages and messages defined on the federation service side.
The `alias` is the FQDN ( `<package-name>.<message-name>` ) to the message.
If this definition exists, type conversion is automatically performed before the field assignment operation.
If a message with this option has a field that is not present in the message specified by alias, and the alias option is not specified for that field, an error is occurred.

# grpc.federation.field

| field | type | required or optional |
| ----- | ---- | -------------------- |
| [`by`](#grpcfederationfieldby) | string | optional |
| [`custom_resolver`](#grpcfederationfieldcustom_resolver) | bool | optional |
| [`alias`](#grpcfederationfieldalias) | string | optional |
| [`string`](#grpcfederationfieldstring) | string | optional |

## (grpc.federation.field).by

`by` used to refer to a name or message argument defined in a MessageRule, use `$.` to refer to the message argument.

## (grpc.federation.field).custom_resolver

If custom_resolver is true, the field binding process is to be implemented in Go.
If there are any values retrieved by `grpc.federation.message` option, they are passed as arguments for custom resolver.

## (grpc.federation.field).alias

`alias` can be used when `alias` is specified in `grpc.federation.message` option,
and specifies the field name to be referenced among the messages specified in `alias` of message option.
If the specified field has the same type or can be converted automatically, its value is assigned.

## (grpc.federation.field).string