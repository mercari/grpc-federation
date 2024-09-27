# grpc.federation.strings APIs

The API for this package was created based on Go's [`strings`](https://pkg.go.dev/strings) and [`strconv`](https://pkg.go.dev/strconv) packages.

# Index

## Strings Functions

- [`clone`](#clone)
- [`compare`](#compare)
- [`contains`](#contains)
- [`containsAny`](#containsAny)
- [`containsRune`](#containsRune)
- [`count`](#count)
- [`cut`](#cut)
- [`cutPrefix`](#cutPrefix)
- [`cutSuffix`](#cutSuffix)
- [`equalFold`](#equalFold)
- [`fields`](#fields)
- [`hasPrefix`](#hasPrefix)
- [`hasSuffix`](#hasSuffix)
- [`index`](#index)
- [`indexAny`](#indexAny)
- [`indexByte`](#indexByte)
- [`indexRune`](#indexRune)
- [`join`](#join)
- [`lastIndex`](#lastIndex)
- [`lastIndexAny`](#lastIndexAny)
- [`lastIndexByte`](#lastIndexByte)
- [`repeat`](#repeat)
- [`replace`](#replace)
- [`replaceAll`](#replaceAll)
- [`split`](#split)
- [`splitAfter`](#splitAfter)
- [`splitAfterN`](#splitAfterN)
- [`splitN`](#splitN)
- [`title`](#title)
- [`toLower`](#toLower)
- [`toTitle`](#toTitle)
- [`toUpper`](#toUpper)
- [`toValidUTF8`](#toValidUTF8)
- [`trim`](#trim)
- [`trimLeft`](#trimLeft)
- [`trimPrefix`](#trimPrefix)
- [`trimRight`](#trimRight)
- [`trimSpace`](#trimSpace)
- [`trimSuffix`](#trimSuffix)

## Strconv Functions

- [`appendBool`](#appendBool)
- [`appendFloat`](#appendFloat)
- [`appendInt`](#appendInt)
- [`appendQuote`](#appendQuote)
- [`appendQuoteRune`](#appendQuoteRune)
- [`appendQuoteToASCII`](#appendQuoteToASCII)
- [`appendUint`](#appendUint)
- [`atoi`](#atoi)
- [`canBackquote`](#canBackquote)
- [`formatBool`](#formatBool)
- [`formatComplex`](#formatComplex)
- [`formatFloat`](#formatFloat)
- [`formatInt`](#formatInt)
- [`formatUint`](#formatUint)
- [`isGraphic`](#isGraphic)
- [`isPrint`](#isPrint)
- [`itoa`](#itoa)
- [`parseBool`](#parseBool)
- [`parseComplex`](#parseComplex)
- [`parseFloat`](#parseFloat)
- [`parseInt`](#parseInt)
- [`parseUint`](#parseUint)
- [`quote`](#quote)
- [`quoteRune`](#quoteRune)
- [`quoteToASCII`](#quoteToASCII)
- [`quotedPrefix`](#quotedPrefix)
- [`unquote`](#unquote)
- [`unquoteChar`](#unquoteChar)

# Strings Functions

## clone

`clone` returns a copy of the input string.

### Parameters

`clone(s string) string`

- `s`: the input string to be cloned.

### Examples

```cel
grpc.federation.strings.clone("hello") //=> "hello"
```

## compare

`compare` compares two strings lexicographically. It returns an integer comparing `a` and `b` lexicographically.

### Parameters

`compare(a string, b string) int`

- `a`: first string.
- `b`: second string.

### Examples

```cel
grpc.federation.strings.compare("a", "b") //=> -1
grpc.federation.strings.compare("a", "a") //=> 0
grpc.federation.strings.compare("b", "a") //=> 1
```

## contains

`contains` reports whether `substr` is within `s`.

### Parameters

`contains(s string, substr string) bool`

- `s`: the string to search within.
- `substr`: the substring to search for.

### Examples

```cel
grpc.federation.strings.contains("hello", "ll") //=> true
grpc.federation.strings.contains("hello", "world") //=> false
```

## containsAny

`containsAny` reports whether any of the Unicode code points in `chars` are within `s`.

### Parameters

`containsAny(s string, chars string) bool`

- `s`: the string to search within.
- `chars`: the set of characters to search for.

### Examples

```cel
grpc.federation.strings.containsAny("hello", "xyz") //=> false
grpc.federation.strings.containsAny("hello", "e") //=> true
```

## containsRune

`containsRune` reports whether the Unicode code point `r` is within `s`.

### Parameters

`containsRune(s string, r int) bool`

- `s`: the string to search within.
- `r`: the rune (Unicode code point) to search for.

### Examples

```cel
grpc.federation.strings.containsRune("hello", 101) //=> true ('e')
grpc.federation.strings.containsRune("hello", 120) //=> false ('x')
```

## count

`count` returns the number of non-overlapping instances of `substr` in `s`.

### Parameters

`count(s string, substr string) int`

- `s`: the string to search within.
- `substr`: the substring to search for.

### Examples

```cel
grpc.federation.strings.count("cheese", "e") //=> 3
grpc.federation.strings.count("five", "ve") //=> 1
```

## cut

`cut` slices the input string `s` into two substrings, `before` and `after`, separated by the first occurrence of the separator `sep`. If `sep` is not found, `before` will be set to `s` and `after` will be empty.

### Parameters

`cut(s string, sep string) []string`

- `s`: the string to cut.
- `sep`: the separator string.

### Returns

A list of two strings:
- The part of `s` before the first occurrence of `sep`.
- The part of `s` after the first occurrence of `sep`.

### Examples

```cel
grpc.federation.strings.cut("gophers", "ph") //=> ["go", "ers"]
grpc.federation.strings.cut("gophers", "x") //=> ["gophers", ""]
```

## cutPrefix

`cutPrefix` returns the string `s` after removing the provided `prefix`. If the string doesn't start with `prefix`, `s` is returned unchanged.

### Parameters

`cutPrefix(s string, prefix string) string`

- `s`: the string to cut the prefix from.
- `prefix`: the prefix to cut.

### Examples

```cel
grpc.federation.strings.cutPrefix("hello", "he") //=> "llo"
grpc.federation.strings.cutPrefix("hello", "wo") //=> "hello"
```

## cutSuffix

`cutSuffix` returns the string `s` after removing the provided `suffix`. If the string doesn't end with `suffix`, `s` is returned unchanged.

### Parameters

`cutSuffix(s string, suffix string) string`

- `s`: the string to cut the suffix from.
- `suffix`: the suffix to cut.

### Examples

```cel
grpc.federation.strings.cutSuffix("hello", "lo") //=> "hel"
grpc.federation.strings.cutSuffix("hello", "wo") //=> "hello"
```

## equalFold

`equalFold` reports whether `s` and `t` are equal under Unicode case-folding.

### Parameters

`equalFold(s string, t string) bool`

- `s`: first string.
- `t`: second string.

### Examples

```cel
grpc.federation.strings.equalFold("Go", "go") //=> true
grpc.federation.strings.equalFold("Go", "Java") //=> false
```

## fields

`fields` splits the string `s` around each instance of one or more consecutive white space characters, returning a slice of substrings.

### Parameters

`fields(s string) []string`

- `s`: the string to split into fields.

### Examples

```cel
grpc.federation.strings.fields("hello world") //=> ["hello", "world"]
grpc.federation.strings.fields("  leading spaces") //=> ["leading", "spaces"]
```

## hasPrefix

`hasPrefix` tests whether the string `s` begins with `prefix`.

### Parameters

`hasPrefix(s string, prefix string) bool`

- `s`: the string to check.
- `prefix`: the prefix to check for.

### Examples

```cel
grpc.federation.strings.hasPrefix("hello", "he") //=> true
grpc.federation.strings.hasPrefix("hello", "wo") //=> false
```

## hasSuffix

`hasSuffix` tests whether the string `s` ends with `suffix`.

### Parameters

`hasSuffix(s string, suffix string) bool`

- `s`: the string to check.
- `suffix`: the suffix to check for.

### Examples

```cel
grpc.federation.strings.hasSuffix("hello", "lo") //=> true
grpc.federation.strings.hasSuffix("hello", "he") //=> false
```

## index

`index` returns the index of the first instance of `substr` in `s`, or `-1` if `substr` is not present in `s`.

### Parameters

`index(s string, substr string) int`

- `s`: the string to search within.
- `substr`: the substring to search for.

### Examples

```cel
grpc.federation.strings.index("hello", "ll") //=> 2
grpc.federation.strings.index("hello", "xx") //=> -1
```

## indexAny

`indexAny` returns the index of the first instance of any Unicode code point from `chars` in `s`, or `-1` if no Unicode code point from `chars` is present in `s`.

### Parameters

`indexAny(s string, chars string) int`

- `s`: the string to search within.
- `chars`: the string containing characters to search for.

### Examples

```cel
grpc.federation.strings.indexAny("hello", "aeiou") //=> 1
grpc.federation.strings.indexAny("hello", "xyz") //=> -1
```

## indexByte

`indexByte` returns the index of the first instance of `byte` in `s`, or `-1` if `byte` is not present.

### Parameters

`indexByte(s string, b byte) int`

- `s`: the string to search within.
- `b`: the byte to search for.

### Examples

```cel
grpc.federation.strings.indexByte("hello", 'e') //=> 1
grpc.federation.strings.indexByte("hello", 'x') //=> -1
```

## indexRune

`indexRune` returns the index of the first instance of the rune `r` in `s`, or `-1` if `r` is not present.

### Parameters

`indexRune(s string, r int) int`

- `s`: the string to search within.
- `r`: the rune (Unicode code point) to search for.

### Examples

```cel
grpc.federation.strings.indexRune("hello", 101) //=> 1 ('e')
grpc.federation.strings.indexRune("hello", 120) //=> -1 ('x')
```

## join

`join` concatenates the elements of the list `elems` to create a single string. The separator string `sep` is placed between elements in the resulting string.

### Parameters

`join(elems []string, sep string) string`

- `elems`: the list of strings to join.
- `sep`: the separator string.

### Examples

```cel
grpc.federation.strings.join(["foo", "bar", "baz"], ", ") //=> "foo, bar, baz"
```

## lastIndex

`lastIndex` returns the index of the last instance of `substr` in `s`, or `-1` if `substr` is not present.

### Parameters

`lastIndex(s string, substr string) int`

- `s`: the string to search within.
- `substr`: the substring to search for.

### Examples

```cel
grpc.federation.strings.lastIndex("go gophers", "go") //=> 3
grpc.federation.strings.lastIndex("hello", "world") //=> -1
```

## lastIndexAny

`lastIndexAny` returns the index of the last instance of any Unicode code point from `chars` in `s`, or `-1` if no Unicode code point from `chars` is present in `s`.

### Parameters

`lastIndexAny(s string, chars string) int`

- `s`: the string to search within.
- `chars`: the string containing characters to search for.

### Examples

```cel
grpc.federation.strings.lastIndexAny("hello", "aeiou") //=> 4
grpc.federation.strings.lastIndexAny("hello", "xyz") //=> -1
```

## lastIndexByte

`lastIndexByte` returns the index of the last instance of `byte` in `s`, or `-1` if `byte` is not present.

### Parameters

`lastIndexByte(s string, b byte) int`

- `s`: the string to search within.
- `b`: the byte to search for.

### Examples

```cel
grpc.federation.strings.lastIndexByte("hello", 'e') //=> 1
grpc.federation.strings.lastIndexByte("hello", 'x') //=> -1
```

## repeat

`repeat` returns a new string consisting of `count` copies of the string `s`.

### Parameters

`repeat(s string, count int) string`

- `s`: the string to repeat.
- `count`: the number of times to repeat the string.

### Examples

```cel
grpc.federation.strings.repeat("ha", 3) //=> "hahaha"
grpc.federation.strings.repeat("ha", 0) //=> ""
```

## replace

`replace` returns a copy of the string `s` with the first `n` non-overlapping instances of `old` replaced by `new`. If `n` is negative, all instances are replaced.

### Parameters

`replace(s string, old string, new string, n int) string`

- `s`: the string to modify.
- `old`: the substring to replace.
- `new`: the replacement substring.
- `n`: the number of instances to replace (or -1 for all).

### Examples

```cel
grpc.federation.strings.replace("foo bar foo", "foo", "baz", 1) //=> "baz bar foo"
grpc.federation.strings.replace("foo bar foo", "foo", "baz", -1) //=> "baz bar baz"
```

## replaceAll

`replaceAll` returns a copy of the string `s` with all non-overlapping instances of `old` replaced by `new`.

### Parameters

`replaceAll(s string, old string, new string) string`

- `s`: the string to modify.
- `old`: the substring to replace.
- `new`: the replacement substring.

### Examples

```cel
grpc.federation.strings.replaceAll("foo bar foo", "foo", "baz") //=> "baz bar baz"
```

## split

`split` slices `s` into all substrings separated by `sep` and returns a slice of the substrings.

### Parameters

`split(s string, sep string) []string`

- `s`: the string to split.
- `sep`: the separator string.

### Examples

```cel
grpc.federation.strings.split("a,b,c", ",") //=> ["a", "b", "c"]
grpc.federation.strings.split("a b c", " ") //=> ["a", "b", "c"]
```

## splitAfter

`splitAfter` slices `s` into all substrings after each instance of `sep` and returns a slice of the substrings.

### Parameters

`splitAfter(s string, sep string) []string`

- `s`: the string to split.
- `sep`: the separator string.

### Examples

```cel
grpc.federation.strings.splitAfter("a,b,c", ",") //=> ["a,", "b,", "c"]
```

## splitAfterN

`splitAfterN` slices `s` into `n` substrings after each instance of `sep` and returns a slice of the substrings.

### Parameters

`splitAfterN(s string, sep string, n int) []string`

- `s`: the string to split.
- `sep`: the separator string.
- `n`: the maximum number of substrings to return.

### Examples

```cel
grpc.federation.strings.splitAfterN("a,b,c", ",", 2) //=> ["a,", "b,c"]
```

## splitN

`splitN` slices `s` into `n` substrings separated by `sep` and returns a slice of the substrings.

### Parameters

`splitN(s string, sep string, n int) []string`

- `s`: the string to split.
- `sep`: the separator string.
- `n`: the maximum number of substrings to return.

### Examples

```cel
grpc.federation.strings.splitN("a,b,c", ",", 2) //=> ["a", "b,c"]
```

## title

`title` returns a copy of the string `s` with all Unicode letters that begin words mapped to their title case.

### Parameters

`title(s string) string`

- `s`: the string to convert to title case.

### Examples

```cel
grpc.federation.strings.title("hello world") //=> "Hello World"
```

## toLower

`toLower` returns a copy of the string `s` with all Unicode letters mapped to their lower case.

### Parameters

`toLower(s string) string`

- `s`: the string to convert to lower case.

### Examples

```cel
grpc.federation.strings.toLower("HELLO") //=> "hello"
```

## toTitle

`toTitle` returns a copy of the string `s` with all Unicode letters mapped to their title case.

### Parameters

`toTitle(s string) string`

- `s`: the string to convert to title case.

### Examples

```cel
grpc.federation.strings.toTitle("hello") //=> "Hello"
```

## toUpper

`toUpper` returns a copy of the string `s` with all Unicode letters mapped to their upper case.

### Parameters

`toUpper(s string) string`

- `s`: the string to convert to upper case.

### Examples

```cel
grpc.federation.strings.toUpper("hello") //=> "HELLO"
```

## toValidUTF8

`toValidUTF8` returns a copy of the string `s` with each run of invalid UTF-8 byte sequences replaced by the replacement string, which may be empty.

### Parameters

`toValidUTF8(s string, replacement string) string`

- `s`: the string to check.
- `replacement`: the string to replace invalid sequences with.

### Examples

```cel
grpc.federation.strings.toValidUTF8("hello\x80world", "?") //=> "hello?world"
```

## trim

`trim` returns a slice of the string `s` with all leading and trailing Unicode code points contained in `cutset` removed.

### Parameters

`trim(s string, cutset string) string`

- `s`: the string to trim.
- `cutset`: the characters to remove from the string.

### Examples

```cel
grpc.federation.strings.trim("!!!hello!!!", "!") //=> "hello"
```

## trimLeft

`trimLeft` returns a slice of the string `s` with all leading Unicode code points contained in `cutset` removed.

### Parameters

`trimLeft(s string, cutset string) string`

- `s`: the string to trim.
- `cutset`: the characters to remove from the start of the string.

### Examples

```cel
grpc.federation.strings.trimLeft("!!!hello!!!", "!") //=> "hello!!!"
```

## trimPrefix

`trimPrefix` returns a slice of the string `s` without the provided leading `prefix`. If `s` doesn't start with `prefix`, it returns `s` unchanged.

### Parameters

`trimPrefix(s string, prefix string) string`

- `s`: the string to trim.
- `prefix`: the prefix to remove.

### Examples

```cel
grpc.federation.strings.trimPrefix("hello world", "hello") //=> " world"
grpc.federation.strings.trimPrefix("hello world", "world") //=> "hello world"
```

## trimRight

`trimRight` returns a slice of the string `s` with all trailing Unicode code points contained in `cutset` removed.

### Parameters

`trimRight(s string, cutset string) string`

- `s`: the string to trim.
- `cutset`: the characters to remove from the end of the string.

### Examples

```cel
grpc.federation.strings.trimRight("!!!hello!!!", "!") //=> "!!!hello"
```

## trimSpace

`trimSpace` returns a slice of the string `s` with all leading and trailing white space removed, as defined by Unicode.

### Parameters

`trimSpace(s string) string`

- `s`: the string to trim.

### Examples

```cel
grpc.federation.strings.trimSpace("  hello  ") //=> "hello"
```

## trimSuffix

`trimSuffix` returns a slice of the string `s` without the provided trailing `suffix`. If `s` doesn't end with `suffix`, it returns `s` unchanged.

### Parameters

`trimSuffix(s string, suffix string) string`

- `s`: the string to trim.
- `suffix`: the suffix to remove.

### Examples

```cel
grpc.federation.strings.trimSuffix("hello world", "world") //=> "hello "
grpc.federation.strings.trimSuffix("hello world", "hello") //=> "hello world"
```

# `strconv` Functions

## appendBool

Appends the string form of a boolean value to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendBool(b []byte, v bool) []byte
grpc.federation.strings.appendBool(s string, v bool) string
```

**Example:**
```cel
grpc.federation.strings.appendBool(b"hello ", true) //=> "hello true"
grpc.federation.strings.appendBool("hello ", true)  //=> "hello true"
```

## appendFloat

Appends the string form of a floating-point value to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendFloat(b []byte, f float64, fmt byte, prec int, bitSize int) []byte
grpc.federation.strings.appendFloat(s string, f float64, fmt byte, prec int, bitSize int) string
```

**Example:**
```cel
grpc.federation.strings.appendFloat(b"price: ", 1.23, 'f', 2, 64) //=> "price: 1.23"
grpc.federation.strings.appendFloat("price: ", 1.23, 'f', 2, 64)  //=> "price: 1.23"
```

## appendInt

Appends the string form of an integer value to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendInt(b []byte, i int64, base int) []byte
grpc.federation.strings.appendInt(s string, i int64, base int) string
```

**Example:**
```cel
grpc.federation.strings.appendInt(b"number: ", 42, 10) //=> "number: 42"
grpc.federation.strings.appendInt("number: ", 42, 10)  //=> "number: 42"
```

## appendQuote

Appends the quoted string form of `s` to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendQuote(b []byte, s string) []byte
grpc.federation.strings.appendQuote(s string, s string) string
```

**Example:**
```cel
grpc.federation.strings.appendQuote(b"quoted: ", "hello") //=> "quoted: \"hello\""
grpc.federation.strings.appendQuote("quoted: ", "hello")  //=> "quoted: \"hello\""
```

## appendQuoteRune

Appends the quoted rune form of `r` to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendQuoteRune(b []byte, r rune) []byte
grpc.federation.strings.appendQuoteRune(s string, r rune) string
```

**Example:**
```cel
grpc.federation.strings.appendQuoteRune(b"quoted: ", 'a') //=> "quoted: 'a'"
grpc.federation.strings.appendQuoteRune("quoted: ", 'a')  //=> "quoted: 'a'"
```

## appendQuoteToASCII

Appends the ASCII-quoted string form of `s` to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendQuoteToASCII(b []byte, s string) []byte
grpc.federation.strings.appendQuoteToASCII(s string, s string) string
```

**Example:**
```cel
grpc.federation.strings.appendQuoteToASCII(b"ascii: ", "abc") //=> "ascii: \"abc\""
grpc.federation.strings.appendQuoteToASCII("ascii: ", "abc")  //=> "ascii: \"abc\""
```

## appendUint

Appends the string form of an unsigned integer value to a byte slice.

**Signature:**
```cel
grpc.federation.strings.appendUint(b []byte, u uint64, base int) []byte
grpc.federation.strings.appendUint(s string, u uint64, base int) string
```

**Example:**
```cel
grpc.federation.strings.appendUint(b"number: ", 123, 10) //=> "number: 123"
grpc.federation.strings.appendUint("number: ", 123, 10)  //=> "number: 123"
```

## atoi

Parses a string and returns the integer it represents.

**Signature:**
```cel
grpc.federation.strings.atoi(s string) int
```

**Example:**
```cel
grpc.federation.strings.atoi("123") //=> 123
```

## canBackquote

Reports whether the string `s` can be represented unchanged as a single-line backquoted string.

**Signature:**
```cel
grpc.federation.strings.canBackquote(s string) bool
```

**Example:**
```cel
grpc.federation.strings.canBackquote("hello") //=> true
grpc.federation.strings.canBackquote("hello\nworld") //=> false
```

## formatBool

Returns the string representation of a boolean value.

**Signature:**
```cel
grpc.federation.strings.formatBool(v bool) string
```

**Example:**
```cel
grpc.federation.strings.formatBool(true) //=> "true"
grpc.federation.strings.formatBool(false) //=> "false"
```

## formatComplex

Returns the string representation of a complex number.

**Signature:**
```cel
grpc.federation.strings.formatComplex(c []float64, fmt byte, prec int, bitSize int) string
```

**Example:**
```cel
grpc.federation.strings.formatComplex([1.23, 4.56], 'f', 2, 64) //=> "(1.23+4.56i)"
```

## formatFloat

Returns the string representation of a floating-point number.

**Signature:**
```cel
grpc.federation.strings.formatFloat(f float64, fmt byte, prec int, bitSize int) string
```

**Example:**
```cel
grpc.federation.strings.formatFloat(1.23, 'f', 2, 64) //=> "1.23"
```

## formatInt

Returns the string representation of an integer.

**Signature:**
```cel
grpc.federation.strings.formatInt(i int64, base int) string
``

`

**Example:**
```cel
grpc.federation.strings.formatInt(42, 10) //=> "42"
```

## formatUint

Returns the string representation of an unsigned integer.

**Signature:**
```cel
grpc.federation.strings.formatUint(u uint64, base int) string
```

**Example:**
```cel
grpc.federation.strings.formatUint(123, 10) //=> "123"
```

## isGraphic

Returns `true` if the provided rune is a graphic, i.e., a printable character other than space.

**Signature:**
```cel
grpc.federation.strings.isGraphic(r rune) bool
```

**Example:**
```cel
grpc.federation.strings.isGraphic('a') //=> true
grpc.federation.strings.isGraphic(' ') //=> false
```

## isPrint

Returns `true` if the provided rune is printable, meaning it is either a letter, number, punctuation, space, or symbol.

**Signature:**
```cel
grpc.federation.strings.isPrint(r rune) bool
```

**Example:**
```cel
grpc.federation.strings.isPrint('a') //=> true
grpc.federation.strings.isPrint('\n') //=> false
```

## itoa

Converts an integer to its string representation.

**Signature:**
```cel
grpc.federation.strings.itoa(i int) string
```

**Example:**
```cel
grpc.federation.strings.itoa(123) //=> "123"
```

## parseBool

Parses a boolean value from its string representation.

**Signature:**
```cel
grpc.federation.strings.parseBool(s string) bool
```

**Example:**
```cel
grpc.federation.strings.parseBool("true")  //=> true
grpc.federation.strings.parseBool("false") //=> false
```

## parseComplex

Parses a complex number from a string representation.

**Signature:**
```cel
grpc.federation.strings.parseComplex(s string, bitSize int) complex128
```

**Example:**
```cel
grpc.federation.strings.parseComplex("(1.23+4.56i)", 128) //=> (1.23 + 4.56i)
```

## parseFloat

Parses a floating-point number from a string representation.

**Signature:**
```cel
grpc.federation.strings.parseFloat(s string, bitSize int) float64
```

**Example:**
```cel
grpc.federation.strings.parseFloat("1.23", 64) //=> 1.23
```

## parseInt

Parses an integer from a string representation, supporting base conversions.

**Signature:**
```cel
grpc.federation.strings.parseInt(s string, base int, bitSize int) int64
```

**Example:**
```cel
grpc.federation.strings.parseInt("42", 10, 64)   //=> 42
grpc.federation.strings.parseInt("101010", 2, 8) //=> 42
```

## parseUint

Parses an unsigned integer from a string representation, supporting base conversions.

**Signature:**
```cel
grpc.federation.strings.parseUint(s string, base int, bitSize int) uint64
```

**Example:**
```cel
grpc.federation.strings.parseUint("42", 10, 64)   //=> 42
grpc.federation.strings.parseUint("101010", 2, 8) //=> 42
```

## quote

Returns a double-quoted string with any special characters escaped.

**Signature:**
```cel
grpc.federation.strings.quote(s string) string
```

**Example:**
```cel
grpc.federation.strings.quote("hello") //=> "\"hello\""
grpc.federation.strings.quote("tab\t") //=> "\"tab\\t\""
```

## quoteRune

Returns a single-quoted string literal with the provided rune, escaping any special characters.

**Signature:**
```cel
grpc.federation.strings.quoteRune(r rune) string
```

**Example:**
```cel
grpc.federation.strings.quoteRune('a') //=> "'a'"
grpc.federation.strings.quoteRune('\n') //=> "'\\n'"
```

## quoteToASCII

Returns a double-quoted string with any non-ASCII characters escaped.

**Signature:**
```cel
grpc.federation.strings.quoteToASCII(s string) string
```

**Example:**
```cel
grpc.federation.strings.quoteToASCII("abc")     //=> "\"abc\""
grpc.federation.strings.quoteToASCII("こんにちは") //=> "\"\\u3053\\u3093\\u306b\\u3061\\u306f\""
```

## quotedPrefix

Parses a quoted prefix from the input string.

**Signature:**
```cel
grpc.federation.strings.quotedPrefix(s string) string
```

**Example:**
```cel
grpc.federation.strings.quotedPrefix("\"hello\" world") //=> "hello"
```

## unquote

Removes the surrounding quotes from a quoted string and unescapes any special characters.

**Signature:**
```cel
grpc.federation.strings.unquote(s string) string
```

**Example:**
```cel
grpc.federation.strings.unquote("\"hello\"") //=> "hello"
```

## unquoteChar

Decodes the next character or byte in the quoted string literal.

**Signature:**
```cel
grpc.federation.strings.unquoteChar(s string, quote byte) rune
```

**Example:**
```cel
grpc.federation.strings.unquoteChar("\\n", '"') //=> '\n'
```
