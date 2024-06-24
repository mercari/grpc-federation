# grpc.federation.enum APIs

# Index

- [`select`](#select)

# Functions

## select

`select` selects the first enum value if the specified condition is `true`, and the second enum value if the condition is `false`. It is used when performing field binding for enums that map multiple enums with an alias.

If you try to return an enum value without using this function, an error will occur due to type mismatch.

e.g.) `true ? pkg.EnumType.Enum_VALUE_A : pkgv2.EnumType.ENUM_VALUE_B`

```
grpc-federation: ERROR: <input>:1:6: found no matching overload for '_?_:_' applied to '(bool, pkg.EnumType(int), pkgv2.EnumType(int))'
 | true ? pkg.EnumType.value('ENUM_VALUE_A') : pkgv2.EnumType.value('ENUM_VALUE_B')
 | .....^
```

### Parameters

`select(cond bool, <first-enum-value>, <second-enum-value>) grpc.federation.private.EnumSelector`

- `cond`: condition to select enum value
- `first-enum-value`: first enum value. this is selected if the condition is true.
- `second-enum-value`: second enum value. this is selected if the condition is false.

### Examples

```cel
grpc.federation.enum.select(true, pkg.EnumType.ENUM_VALUE_A, pkgv2.EnumType.ENUM_VALUE_B)
```