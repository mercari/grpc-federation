# grpc.federation.regexp APIs

The API for this package was created based on Go's [`regexp`](https://pkg.go.dev/regexp) package.

# Index

## Regexp

- [`compile`](#compile)
- [`mustCompile`](#mustCompile)
- [`quoteMeta`](#quoteMeta)
- [`Regexp.findStringSubmatch`](#regexpfindStringSubmatch)
- [`Regexp.matchString`](#regexpmatchString)
- [`Regexp.replaceAllString`](#regexpreplaceAllString)
- [`Regexp.string`](#regexpstring)

# Types

## Regexp

A `Regexp` is a compiled regular expression.

FYI: https://pkg.go.dev/regexp#Regexp

# Functions

## compile

`compile` parses a regular expression and returns, if successful, a Regexp object that can be used to match against text.

FYI: https://pkg.go.dev/regexp#Compile

### Parameters

`compile(expr string) Regexp`

- `expr`: the regular expression pattern to compile.

### Examples

```cel
grpc.federation.regexp.compile("^[a-z]+$")
```

## mustCompile

`mustCompile` is like `compile` but panics if the expression cannot be parsed. It simplifies safe initialization of global variables holding compiled regular expressions.

FYI: https://pkg.go.dev/regexp#MustCompile

### Parameters

`mustCompile(expr string) Regexp`

- `expr`: the regular expression pattern to compile.

### Examples

```cel
grpc.federation.regexp.mustCompile("^[a-z]+$")
```

## quoteMeta

`quoteMeta` returns a string that escapes all regular expression metacharacters inside the argument text.

FYI: https://pkg.go.dev/regexp#QuoteMeta

### Parameters

`quoteMeta(s string) string`

- `s`: the string to escape.

### Examples

```cel
grpc.federation.regexp.quoteMeta("foo.bar") //=> "foo\\.bar"
```

# Regexp Methods

## Regexp.findStringSubmatch

`findStringSubmatch` returns a slice of strings holding the text of the leftmost match of the regular expression in s and the matches, if any, of its subexpressions.

FYI: https://pkg.go.dev/regexp#Regexp.FindStringSubmatch

### Parameters

`Regexp.findStringSubmatch(s string) []string`

- `s`: the string to search.

### Examples

```cel
re := grpc.federation.regexp.mustCompile("a(x*)b")
re.findStringSubmatch("-ab-") //=> ["ab", ""]
re.findStringSubmatch("-axxb-") //=> ["axxb", "xx"]
re.findStringSubmatch("-ab-") //=> null
```

## Regexp.matchString

`matchString` reports whether the string s contains any match of the regular expression.

FYI: https://pkg.go.dev/regexp#Regexp.MatchString

### Parameters

`Regexp.matchString(s string) bool`

- `s`: the string to check.

### Examples

```cel
re := grpc.federation.regexp.mustCompile("^[a-z]+$")
re.matchString("hello") //=> true
re.matchString("Hello") //=> false
```

## Regexp.replaceAllString

`replaceAllString` replaces matches of the regexp with a replacement string.

FYI: https://pkg.go.dev/regexp#Regexp.ReplaceAllString

### Parameters

`Regexp.replaceAllString(src string, repl string) string`

- `src`: the source string to search and replace.
- `repl`: the replacement string.

### Examples

```cel
re := grpc.federation.regexp.mustCompile("a(x*)b")
re.replaceAllString("-ab-axxb-", "T") //=> "-T-T-"
re.replaceAllString("-ab-axxb-", "$1") //=> "--xx-"
re.replaceAllString("-ab-axxb-", "$1W") //=> "--WxxW-"
```

## Regexp.string

`string` returns the source text used to compile the regular expression.

FYI: https://pkg.go.dev/regexp#Regexp.String

### Parameters

`Regexp.string() string`

### Examples

```cel
re := grpc.federation.regexp.mustCompile("^[a-z]+$")
re.string() //=> "^[a-z]+$"
```
