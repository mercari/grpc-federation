# gRPC Federation CEL API References

In the gRPC Federation, you can use `by` to write [CEL(Common Expression Language)](https://github.com/google/cel-spec) expressions.  

For more information on CEL, [please see here](https://github.com/google/cel-spec/blob/master/doc/langdef.md).

[Here is a list of macros that CEL supports by default](https://github.com/google/cel-spec/blob/master/doc/langdef.md#macros).

In addition to the standard CEL operations, the gRPC Federation supports a number of its own APIs. This page introduces those APIs.

- [grpc.federation.list APIs](./cel/list.md)
- [grpc.federation.rand APIs](./cel/rand.md)
- [grpc.federation.time APIs](./cel/time.md)
- [grpc.federation.uuid APIs](./cel/uuid.md)

## Refer to the defined variable

If you have defined variables using [`def`](#grpcfederationmessagedef) feature, you can use them in CEL.  
Also, the message argument should be `$.` can be used to refer to them.

## error

`error` is a reserved word. The `error` variable is used to represent errors in gRPC method calls.
[The type of the `error` variable is defined here](../proto/grpc/federation/private.proto).

### Example

```proto
if: "error.precondition_failures[0].violations[0].subject == 'subject value'"
```

### Type Definition

| field | type |
| ----- | ---- |
| `code` | int32 |
| `message` | string |
| `details` | repeated google.protobuf.Any |
| `custom_messages` | repeated google.protobuf.Any |
| `error_info` | repeated google.rpc.ErrorInfo |
| `retry_info` | repeated google.rpc.RetryInfo |
| `debug_info` | repeated google.rpc.DebugInfo |
| `quota_failures` | repeated google.rpc.QuotaFailure |
| `precondition_failures` | repeated google.rpc.PreconditionFailure |
| `bad_requests` | repeated google.rpc.BadRequest |
| `request_info` | repeated google.rpc.RequestInfo |
| `resource_info` | repeated google.rpc.ResourceInfo |
| `helps` | repeated google.rpc.Help |
| `localized_messages` | repeated google.rpc.LocalizedMessage |