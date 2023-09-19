# gRPC Federation Feature References

# grpc.federation.service

| field | type | required or optional | link |
| ----- | ---- | ------------------- | ---- |
| `dependencies` | repeated ServiceDependency | optional | [(grpc.federation.service).dependencies](##(grpc.federation.service).dependencies) |

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

| field | type | required or optional | link |
| ----- | ---- | ------------------- | ---- |
| `name` | string | optional | [(grpc.federation.service).dependencies.name](##(grpc.federation.service).dependencies.name) |
| `service` | string | required | [(grpc.federation.service).dependencies.service](##(grpc.federation.service).dependencies.service) |

## (grpc.federation.service).dependencies.name

Name to be used when initializing the gRPC client.

## (grpc.federation.service).dependencies.service

Service is the name of the dependent service.

# grpc.federation.method

| field | type | required or optional | link |
| ----- | ---- | ------------------- | ---- |
| `timeout` | string | optional | [(grpc.federation.method).timeout](##(grpc.federation.method).timeout) |

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

| field | type | required or optional | link |
| ----- | ---- | ------------------- | ---- |
| `alias` | string | optional | [grpc.federation.enum.alias](##(grpc.federation.enum).alias) |

## (grpc.federation.enum).alias

`alias` mapping between enums defined in other packages and enums defined on the federation service side.
The `alias` is the FQDN ( `<package-name>.<enum-name>` ) to the enum.
If this definition exists, type conversion is automatically performed before the enum value assignment operation.
If a enum with this option has a value that is not present in the enum specified by alias, and the alias option is not specified for that value, an error is occurred.

# grpc.federation.enum_value

| field | type | required or optional | link |
| ----- | ---- | ------------------- | ---- |
| `default` | bool | optional | |
| `alias` | repeated string | optional | |

## (grpc.federation.enum_value).default

Specifies the default value of the enum.
All values other than those specified in alias will be default values.

## (grpc.federation.enum_value).alias

`alias` can be used when alias is specified in `grpc.federation.enum` option,
and specifies the value name to be referenced among the enums specified in alias of enum option.
multiple value names can be specified for `alias`.

# grpc.federation.message

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `resolver` | Resolver | optional | [(grpc.federation.message).resolver](##(grpc.federation.message).resolver) |
| `messages` | repeated Message | optional | [(grpc.federation.message).messages](##(grpc.federation.message).messages) |
| `custom_resolver` | bool | optional | [(grpc.federation.message).custom_resolver](##(grpc.federation.message).custom_resolver) |
| `alias` | string | optional | [(grpc.federation.message).alias](##(grpc.federation.message).alias) |

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

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `method` | string | required | [(grpc.federation.message).resolver.method](##(grpc.federation.message).resolver.method) |
| `request` | repeated MethodRequest | optional | [(grpc.federation.message).resolver.request](##(grpc.federation.message).resolver.request) |
| `response` | repeated MethodResponse | optional | [(grpc.federation.message).resolver.response](##(grpc.federation.message).resolver.response) |
| `timeout` | string | optional | [(grpc.federation.message).resolver.timeout](##(grpc.federation.message).resolver.timeout) |
| `retry` | RetryPolicy | optional | [(grpc.federation.message).resolver.retry](##(grpc.federation.message).resolver.retry) |

## (grpc.federation.message).resolver.method

Specify the FQDN for the gRPC method. format is `<package-name>.<service-name>/<method-name>`.

```proto
message MyMessage {
  option (grpc.federation.message) = {
    resolver {
      method: "foopkg.FooService/GetFoo"
    }
  };
}
```

## (grpc.federation.message).resolver.request

Specify the request parameters for the gRPC method.
The `field` corresponds to the field name in the request message, and the `by` or `string` specifies which value is associated with the field.

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `field` | string | required | [(grpc.federation.message).resolver.request.field](##(grpc.federation.message).resolver.request.field) |
| `by` | string | optional | [(grpc.federation.message).resolver.request.by](##(grpc.federation.message).resolver.request.by) |
| `string` | string | optional | [(grpc.federation.message).resolver.request.string](##(grpc.federation.message).resolver.request.string) |

In addition to `string`, you can write literals for any data type supported by proto, such as `int64` or `bool` .

## (grpc.federation.message).resolver.request.field

The field name of the request message.

## (grpc.federation.message).resolver.request.by

`by` used to refer to a name or message argument defined in a MessageRule, use `$.` to refer to the message argument.

## (grpc.federation.message).resolver.request.string

## (grpc.federation.message).resolver.response

Describe the information necessary to bind the result of the method call to the message field. Select and name the required information from the response message, and use the defined name when referencing the information in the message option.

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `name` | string | optional | [(grpc.federation.message).resolver.response.name](##(grpc.federation.message).resolver.response.name) |
| `field` | string | optional | [(grpc.federation.message).resolver.response.field](##(grpc.federation.message).resolver.response.field) |
| `autobind` | bool | optional | [(grpc.federation.message).resolver.response.autobind](##(grpc.federation.message).resolver.response.autobind) |

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

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `constant` | RetryPolicyConstant | optional | [(grpc.federation.message).resolver.retry.constant](##(grpc.federation.message).resolver.retry.constant) |
| `exponential` | RetryPolicyExponential | optional | [(grpc.federation.message).resolver.retry.exponential](##(grpc.federation.message).resolver.retry.exponential) |

## (grpc.federation.message).resolver.retry.constant

Retry according to the "constant" policy.

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `inerval` | string | optional | [(grpc.federation.message).resolver.retry.constant.interval](##(grpc.federation.message).resolver.retry.constant.interval) |
| `max_retries` | uint64 | optional | [(grpc.federation.message).resolver.retry.constant.max_retries](##(grpc.federation.message).resolver.retry.constant.max_retries) |

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

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `initial_interval` | string | optional | [(grpc.federation.message).resolver.retry.exponential.initial_interval](##(grpc.federation.message).resolver.retry.exponential.initial_interval) |
| `randomization_factor` | double | optional | [(grpc.federation.message).resolver.retry.exponential.randomization_factor](##(grpc.federation.message).resolver.retry.exponential.randomization_factor) |
| `multiplier` | double | optional | [(grpc.federation.message).resolver.retry.exponential.multiplier](##(grpc.federation.message).resolver.retry.exponential.multiplier) |
| `max_interval` | string | optional | [(grpc.federation.message).resolver.retry.exponential.max_interval](##(grpc.federation.message).resolver.retry.exponential.max_interval) |
| `max_retries` | uint64 | optional | [(grpc.federation.message).resolver.retry.exponential.max_retries](##(grpc.federation.message).resolver.retry.exponential.max_retries) |

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

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `name` | string | optional | [(grpc.federation.message).messages.name](##(grpc.federation.message).messages.name) |
| `message` | string | required | [(grpc.federation.message).messages.message](##(grpc.federation.message).messages.message) |
| `args` | repeated Argument | optional | [[(grpc.federation.message).messages.args](##(grpc.federation.message).messages.args) |
| `autobind` | bool | optional | [(grpc.federation.message).messages.autobind](##(grpc.federation.message).messages.autobind) |

## (grpc.federation.message).messages.name

Specify a unique name for the dependent message.

## (grpc.federation.message).messages.message

Specify the message to be referred to by FQDN. format is `<package-name>.<message-name>`.  
`<package-name>` can be omitted when referring to messages in the same package.

## (grpc.federation.message).messages.args

Specify the parameters needed to retrieve the message. This is called the message argument.

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `name` | string | required | [(grpc.federation.message).messages.args.name](##(grpc.federation.message).messages.args.name) |
| `by` | string | optional | [(grpc.federation.message).messages.args.by](##(grpc.federation.message).messages.args.by) |
| `inline` | string | optional | [(grpc.federation.message).messages.args.inline](##(grpc.federation.message).messages.args.inline) |
| `string` | string | optional | [(grpc.federation.message).messages.args.string](##(grpc.federation.message).messages.args.string) |

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

| field | type | required or optional | link |
| ----- | ---- | -------------------- | ---- |
| `by` | string | optional | [(grpc.federation.field).by](##(grpc.federation.field).by) |
| `custom_resolver` | bool | optional | [(grpc.federation.field).custom_resolver](##(grpc.federation.field).custom_resolver) |
| `alias` | string | optional | [(grpc.federation.field).alias](##(grpc.federation.field).alias) |
| `string` | string | optional | [(grpc.federation.field).string](##(grpc.federation.field).string) |

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