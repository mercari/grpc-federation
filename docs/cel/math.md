# grpc.federation.math APIs

The API for this package was created based on Go's [`math`](https://pkg.go.dev/math) package.

# Index

- [`sqrt`](#sqrt)
- [`pow`](#pow)
- [`floor`](#floor)
- [`round`](#round)

# Functions

## sqrt

`sqrt` returns the square root of x.

FYI: https://pkg.go.dev/math#Sqrt

### Parameters

`sqrt(x double) double`
`sqrt(x int) double`

- `x`: a number

### Examples

```cel
grpc.federation.math.sqrt(25.0)  // returns 5.0
grpc.federation.math.sqrt(25)    // returns 5.0
grpc.federation.math.sqrt(3.0*3.0 + 4.0*4.0)  // returns 5.0 (Pythagorean theorem)
```

## pow

`pow` returns x**y, the base-x exponential of y.

FYI: https://pkg.go.dev/math#Pow

### Parameters

`pow(x double, y double) double`

- `x`: the base
- `y`: the exponent

### Examples

```cel
grpc.federation.math.pow(2.0, 3.0)   // returns 8.0 (2^3)
grpc.federation.math.pow(10.0, 2.0)  // returns 100.0 (10^2)
```

## floor

`floor` returns the greatest integer value less than or equal to x.

FYI: https://pkg.go.dev/math#Floor

### Parameters

`floor(x double) double`

- `x`: a number

### Examples

```cel
grpc.federation.math.floor(1.51)   // returns 1.0
grpc.federation.math.floor(-1.51)  // returns -2.0
grpc.federation.math.floor(3.0)    // returns 3.0
```

## round

`round` returns the nearest integer, rounding half away from zero.

FYI: https://pkg.go.dev/math#Round

### Parameters

`round(x double) double`

- `x`: a number

### Examples

```cel
grpc.federation.math.round(1.51)   // returns 2.0
grpc.federation.math.round(1.49)   // returns 1.0
grpc.federation.math.round(-1.51)  // returns -2.0
grpc.federation.math.round(1.5)    // returns 2.0
``` 
