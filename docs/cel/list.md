# grpc.federation.list APIs

# Index

- [`reduce`](#reduce)
- [`first`](#first)

# Macros

## reduce

`reduce` executes a user-supplied "reducer" expression on each element of the repeated type, in order, passing in the return value from the calculation on the preceding element. The final result of running the reducer across all elements of the repeated type is a single value.

### Parameters

`range.reduce(accumulator, current, <expr>, <init>)`

- `range`: repeated type value
- `accumulator` : The value resulting from the previous  `<expr>` expression. On the first call, its value is the result of `<init>` expression.
- `current`: current iteration value
- `<expr>`: expression for reduce operation
- `<init>`: expression for initial value

### Examples

```cel
[2, 3, 4].reduce(accum, cur, accum + cur, 1) //=> 10
```

## first

Returns the element that evaluates to `<expr>` when the result is true. If all elements are not matched, `optional.none` is returned. Thus, the return value of `first` will always be of optional type. The optional value can be used as is for field binding.

`first` is equivalent to the following expression.

```cel
range.filter(var, <expr>)[?0]
```

### Parameters

`range.first(var, <expr>)`

- `range`: repeated type value
- `var`: current iteration value
- `<expr>`: Write a conditional expression to return the first match. Must return a boolean value.

### Examples

```cel
[1, 2, 3, 4].first(v, v % 2 == 0) //=> optional.of(2)
```

# Functions