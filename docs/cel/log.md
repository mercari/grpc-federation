# grpc.federation.log APIs

This is a CEL library for manipulating slog.Logger on the proto side.

# Index

- [`debug`](#debug)
- [`info`](#info)
- [`warn`](#warn)
- [`error`](#error)
- [`add`](#add)

# Functions

## debug

`debug` logs message at DEBUG level.

### Parameters

`debug(msg, <expr>)`

- `msg`: string value
- `<expr>` : map value expression to add attributes to logs. this parameter is optional

### Examples

```cel
grpc.federation.log.debug('message') //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"DEBUG","msg":"message"}
grpc.federation.log.debug('message', {'foo': 'bar'}) //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"DEBUG","msg":"message","foo":"bar"}
```

## info

`info` logs message at INFO level.

### Parameters

`info(msg, <expr>)`

- `msg`: string value
- `<expr>` : map value expression to add attributes to logs. this parameter is optional

### Examples

```cel
grpc.federation.log.info('message') //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"INFO","msg":"message"}
grpc.federation.log.info('message', {'foo': 'bar'}) //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"INFO","msg":"message","foo":"bar"}
```

## warn

`warn` logs message at WARN level.

### Parameters

`warn(msg, <expr>)`

- `msg`: string value
- `<expr>` : map value expression to add attributes to logs. this parameter is optional

### Examples

```cel
grpc.federation.log.warn('message') //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"WARN","msg":"message"}
grpc.federation.log.warn('message', {'foo': 'bar'}) //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"WARN","msg":"message","foo":"bar"}
```

## error

`error` logs message at ERROR level.

### Parameters

`error(msg, <expr>)`

- `msg`: string value
- `<expr>` : map value expression to add attributes to logs. this parameter is optional

### Examples

```cel
grpc.federation.log.error('message') //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"ERROR","msg":"message"}
grpc.federation.log.error('message', {'foo': 'bar'}) //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"ERROR","msg":"message","foo":"bar"}
```

## add

`add` adds attributes to logs.

### Parameters

`add(<expr>)`

- `<expr>` : map value expression to add attributes to logs

### Examples

```cel
grpc.federation.log.add({'foo': 'bar'})
grpc.federation.log.info('message') //=> {"time":"2006-01-02T15:04:05.999999999Z07:00","level":"INFO","msg":"message","foo":"bar"}
```
