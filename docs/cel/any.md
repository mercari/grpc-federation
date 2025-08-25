# grpc.federation.any APIs

# Index

- [`new`](#new)

# Functions

## new

`new` explicitly create an instance of the `google.protobuf.Any` type. The argument must be a Protocol Buffers message type.

### Parameters

`new(message) google.protobuf.Any`

### Examples

```cel
grpc.federation.any.new(google.protobuf.Empty{})
```
