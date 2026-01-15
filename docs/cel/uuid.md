# grpc.federation.uuid APIs

The API for this package was created based on Go's [`google/uuid`](https://pkg.go.dev/github.com/google/uuid) package.

# Index

## UUID

- [`new`](#new)
- [`newRandom`](#newrandom)
- [`newRandomFromRand`](#newrandomfromrand)
- [`parse`](#parse)
- [`validate`](#validate)
- [`UUID.domain`](#uuiddomain)
- [`UUID.id`](#uuidid)
- [`UUID.time`](#uuidtime)
- [`UUID.urn`](#uuidurn)
- [`UUID.string`](#uuidstring)
- [`UUID.version`](#uuidversion)

# Types

## UUID

A `UUID` is a 128 bit (16 byte) Universal Unique IDentifier as defined in [RFC 4122](https://www.rfc-editor.org/rfc/rfc4122.html).

FYI: https://pkg.go.dev/github.com/google/uuid#UUID

# Functions

## new

`new` creates a new random `UUID`. `new` is equivalent to the expression.

FYI: https://pkg.go.dev/github.com/google/uuid#New

### Parameters

`new() UUID`

### Examples

```cel
grpc.federation.uuid.new()
```

## newRandom

`newRandom` returns a Random (Version 4) UUID.

FYI: https://pkg.go.dev/github.com/google/uuid#NewRandom

### Parameters

`newRandom() UUID`

### Examples

```cel
grpc.federation.uuid.newRandom()
```

## newRandomFromRand

`newRandomFromRand` returns a `UUID` based on [`grpc.federation.rand.Rand`](./rand.md#rand) instance.

FYI: https://pkg.go.dev/github.com/google/uuid#NewRandomFromReader

### Parameters

`newRandomFromRand(rand grpc.federation.rand.Rand) UUID`

- `rand`: [`grpc.federation.rand.Rand`](./rand.md#rand) instance

### Examples

Create UUID from fixed seed value.

```cel
grpc.federation.uuid.newRandomFromRand(grpc.federation.rand.new(grpc.federation.rand.newSource(10)))
```

## parse

`parse` returns a `UUID` parsed from x.

FYI: https://pkg.go.dev/github.com/google/uuid#Parse

### Parameters

`parse(x string) UUID`

- `x`: a string

### Examples

```cel
grpc.federation.uuid.parse('daa4728d-159f-4fc2-82cf-cae915d54e08')
gprc.federation.uuid.parse(gprc.federation.uuid.new().string())
```

## validate

`validate` returns a bool indicating whether the string is a valid UUID.

FYI: https://pkg.go.dev/github.com/google/uuid#Validate

### Parameters

`validate(x string) bool`

- `x`: a string

### Examples

```cel
grpc.federation.uuid.validate('daa4728d-159f-4fc2-82cf-cae915d54e08')
grpc.federation.uuid.validate('invalid-uuid')
```

## UUID.domain

`domain` returns the domain for a Version 2 UUID. Domains are only defined for Version 2 UUIDs.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.Domain

### Parameters

`UUID.domain() uint`

### Examples

```cel
grpc.federation.uuid.new().domain()
```

## UUID.id

`id` returns the id for a Version 2 UUID. IDs are only defined for Version 2 UUIDs.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.ID

### Parameters

`UUID.id() uint`

### Examples

```cel
grpc.federation.uuid.new().id()
```

## UUID.time

`time` returns the time in 100s of nanoseconds since 15 Oct 1582 encoded in uuid. The time is only defined for version 1 and 2 UUIDs.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.Time

### Parameters

`UUID.time() int`

### Examples

```cel
grpc.federation.uuid.new().time()
```

## UUID.urn

URN returns the [RFC 2141](https://www.rfc-editor.org/rfc/rfc2141.html) URN form of uuid, urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx, or "" if uuid is invalid.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.URN

### Parameters

`UUID.urn() string`

### Examples

```cel
grpc.federation.uuid.new().urn()
```

## UUID.string

`string` returns the string form of uuid, xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx , or "" if uuid is invalid.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.String

### Parameters

`UUID.string() string`

### Examples

```cel
grpc.federation.uuid.new().string()
```

## UUID.version

`version` returns the version of uuid.

FYI: https://pkg.go.dev/github.com/google/uuid#UUID.Version

### Parameters

`UUID.version() uint`

### Examples

```cel
grpc.federation.uuid.new().version()
```
