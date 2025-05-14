# grpc.federation.metadata APIs

The API for this package was created based on Go's [`google.golang.org/grpc/metadata`](https://pkg.go.dev/google.golang.org/grpc/metadata) package.

# Index

- [`new`](#new)
- [`incoming`](#incoming)

# Functions

## new

`new` creates a new metadata instance.
The created type is `map(string, list(string))`.

### Parameters

`new() map(string, list(string))`

### Examples

```cel
grpc.federation.metadata.new()
```

## incoming

`incoming` returns the incoming metadata if it exists.

FYI: https://pkg.go.dev/google.golang.org/grpc/metadata#FromIncomingContext

### Parameters

`incoming() map(string, list(string))`

### Examples

```cel
grpc.federation.metadata.incoming()
```
