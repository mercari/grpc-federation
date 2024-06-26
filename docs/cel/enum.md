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

## <enum-fqdn>.name

For all enum types, you can use the `name` method to obtain the name of the enum value as a string.

### Parameters

`pkg.EnumType.name(<enum-value>) string`

- `<enum-value>`: int or enum typed value

### Examples

In the following case, `"ENUM_VALUE_1"` ( string typed value) is returned.

```cel
foo.EnumType.name(foo.EnumType.ENUM_VALUE_1)
```

```proto
package foo;

enum EnumType {
  ENUM_VALUE_UNKNOWN = 0;
  ENUM_VALUE_1 = 1;
}
```

## <enum-fqdn>.value

For all enum types, you can use the `value` method to obtain the enum typed value from name.

### Parameters

`pkg.EnumType.value(enumValueName string) EnumValue`

- `enumValueName`: name of the enum value
- `EnumValue`: enum typed value. This value can be used as an argument for the `select` function

### Examples

In the following case, `foo.EnumType.ENUM_VALUE_1` ( enum typed value) is returned.

```cel
foo.EnumType.value('ENUM_VALUE_1')
```

```proto
package foo;

enum EnumType {
  ENUM_VALUE_UNKNOWN = 0;
  ENUM_VALUE_1 = 1;
}
```

## <enum-fqdn>.from

For all enum types, you can use the `from` method to obtain the enum typed value from int value.

### Parameters

`pkg.EnumType.from(enumValue int) EnumValue`

- `enumValue`: int typed value of the enum value
- `EnumValue`: enum typed value. This value can be used as an argument for the `select` function

### Examples

In the following case, `foo.EnumType.ENUM_VALUE_1` ( enum typed value) is returned.

> [!NOTE]
> Here, you might be confused by the difference between `foo.EnumType.ENUM_VALUE_1` appearing in the expression and the returned `foo.EnumType.ENUM_VALUE_1` value.
> As a basic principle, CEL treats enum values as int types when specified directly.
> The `from` method is used to convert this to a typed enum value.
> Although the result looks the same, it is actually typed and can be used as an argument for the `select` function.


```cel
foo.EnumType.from(foo.EnumType.ENUM_VALUE_1)
```

```proto
package foo;

enum EnumType {
  ENUM_VALUE_UNKNOWN = 0;
  ENUM_VALUE_1 = 1;
}
```