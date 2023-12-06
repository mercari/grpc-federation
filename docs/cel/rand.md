# grpc.federation.rand APIs

The API for this package was created based on Go's [`math/rand`](https://pkg.go.dev/math/rand) package.

# Index

## Source

- [`newSource`](#newsource)
- [`Source.int63`](#sourceint63)
- [`Source.seed`](#sourceseed)

## Rand

- [`new`](#new)
- [`Rand.expFloat64`](#randexpfloat64)
- [`Rand.float32`](#randfloat32)
- [`Rand.float64`](#randfloat64)
- [`Rand.int`](#randint)
- [`Rand.int31`](#randint31)
- [`Rand.int31n`](#randint31n)
- [`Rand.int63`](#randint63)
- [`Rand.int63n`](#randint63n)
- [`Rand.intn`](#randintn)
- [`Rand.normFloat64`](#randnormfloat64)
- [`Rand.seed`](#randseed)
- [`Rand.uint32`](#randuint32)

# Types

## Source

A `Source` represents a source of uniformly-distributed pseudo-random int64 values in the range [0, 1<<63).  

FYI: https://pkg.go.dev/math/rand#Source

## Rand

A `Rand` is a source of random numbers.  

FYI: https://pkg.go.dev/math/rand#Rand

# Functions

## newSource

`newSource` returns a new pseudo-random `Source` seeded with the given value.  

FYI: https://pkg.go.dev/math/rand#NewSource

### Parameters

`newSource(seed int) Source`

- `seed`: seed value

### Examples

```cel
grpc.federation.rand.newSource(10)
```

## Source.int63

`int63` returns a non-negative pseudo-random 63-bit integer as an int64.

### Parameters

`Source.int63() int`

### Examples

```cel
grpc.federation.rand.newSource(10).int63()
```

## Source.seed

`seed` uses the provided seed value to initialize the generator to a deterministic state.

### Parameters

`Source.seed(seed int) bool`

- `seed`: the seed value

Return value is always `true`.

### Examples

```cel
grpc.federation.rand.newSource(10).seed(20)
```

## new

`new` returns a new `Rand` that uses random values from `src` to generate other random values.

FYI: https://pkg.go.dev/math/rand#New

### Parameters

`new(src Source) Rand`

- `src`: source value for generates random values

### Examples

```cel
grpc.federation.rand.new()
```

## Rand.expFloat64

`expFloat64` returns an exponentially distributed double type value in the range (0, `max float64 value`) with an exponential distribution whose rate parameter (lambda) is 1 and whose mean is 1/lambda (1). To produce a distribution with a different rate parameter, callers can adjust the output using:

```cel
sample = grpc.federation.rand.new().expFloat64() / desiredRateParameter
```

FYI: https://pkg.go.dev/math/rand#Rand.ExpFloat64

### Parameters

`Rand.expFloat64() double`

### Examples

```cel
grpc.federation.rand.new().expFloat64()
```

## Rand.float32

`float32` returns a value as float type, a pseudo-random number in the half-open interval (0.0,1.0).

FYI: https://pkg.go.dev/math/rand#Rand.Float32

### Parameters

`Rand.float32() float`

### Examples

```cel
grpc.federation.rand.new().float32()
```

## Rand.float64

`float64` returns a value as double type, a pseudo-random number in the half-open interval (0.0,1.0).

FYI: https://pkg.go.dev/math/rand#Rand.Float64

### Parameters

`Rand.float64() double`

### Examples

```cel
grpc.federation.rand.new().float64()
```

## Rand.int

`int` returns a non-negative pseudo-random int.

FYI: https://pkg.go.dev/math/rand#Rand.Int

### Parameters

`Rand.int() int`

### Examples

```cel
grpc.federation.rand.new().int()
```

## Rand.int31

`int31` returns a non-negative pseudo-random 31-bit integer as an int value.

FYI: https://pkg.go.dev/math/rand#Rand.Int31

### Parameters

`Rand.int31() int`

### Examples

```cel
grpc.federation.rand.new().int31()
```

## Rand.int31n

`int31n` returns as an int type, a non-negative pseudo-random number in the half-open interval (0,n). It returns error if n <= 0.

FYI: https://pkg.go.dev/math/rand#Rand.Int31n

### Parameters

`Rand.int31n(n int) int`

- `n`: non-negative value

### Examples

```cel
grpc.federation.rand.new().int31n(10)
```

## Rand.int63

`int63` returns a non-negative pseudo-random 63-bit integer as an int64.

FYI: https://pkg.go.dev/math/rand#Rand.Int63

### Parameters

`Rand.int63() int`

### Examples

```cel
grpc.federation.rand.new().int63()
```

## Rand.int63n

`int63n` returns as an int64 bit value, a non-negative pseudo-random number in the half-open interval [0,n). It panics if n <= 0.

FYI: https://pkg.go.dev/math/rand#Rand.Int63n

### Parameters

`Rand.int63n(n int) int`

### Examples

```cel
grpc.federation.rand.new().int63n(10)
```

## Rand.intn

`intn` returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n). It panics if n <= 0.

FYI: https://pkg.go.dev/math/rand#Rand.Intn

### Parameters

`Rand.intn(n int) int`

### Examples

```cel
grpc.federation.rand.new().intn(10)
```

## Rand.normFloat64

`normFloat64` returns a normally distributed double value in the range `-<max double value>` through `<max double value>` inclusive, with standard normal distribution (mean = 0, stddev = 1). To produce a different normal distribution, callers can adjust the output using:

```cel
sample = grpc.federation.rand.new().normFloat64() * desiredStdDev + desiredMean
```

FYI: https://pkg.go.dev/math/rand#Rand.NormFloat64

### Parameters

`Rand.normFloat64() double`

### Examples

```cel
grpc.federation.rand.new().normFloat64()
```

## Rand.seed

`seed` uses the provided seed value to initialize the generator to a deterministic state.

FYI: https://pkg.go.dev/math/rand#Rand.Seed

### Parameters

`Rand.seed(seed int) bool`

- `seed`: the seed value

### Examples

```cel
grpc.federation.rand.new().seed(10)
```

## Rand.uint32

`uint32` returns a pseudo-random 32-bit value as a uint32.

FYI: https://pkg.go.dev/math/rand#Rand.Uint32

### Parameters

`Rand.uint32() uint`

### Examples

```cel
grpc.federation.rand.new().uint32()
```