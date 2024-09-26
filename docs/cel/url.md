# grpc.federation.url APIs

The API for this package was created based on Go's [`net/url`](https://pkg.go.dev/net/url) package.

# Index

## Functions

- [`joinPath`](#joinPath)
- [`pathEscape`](#pathEscape)
- [`pathUnescape`](#pathUnescape)
- [`queryEscape`](#queryEscape)
- [`queryUnescape`](#queryUnescape)

## URL

- [`parse`](#parse)
- [`parseRequestURI`](#parseRequestURI)
- [`URL.scheme`](#urlscheme)
- [`URL.opaque`](#urlopaque)
- [`URL.userinfo`](#urluserinfo)
- [`URL.host`](#urlhost)
- [`URL.path`](#urlpath)
- [`URL.rawPath`](#urlrawpath)
- [`URL.forceQuery`](#urlforceQuery)
- [`URL.rawQuery`](#urlrawQuery)
- [`URL.fragment`](#urlfragment)
- [`URL.rawFragment`](#urlrawFragment)
- [`URL.escapedFragment`](#urlescapedFragment)
- [`URL.escapedPath`](#urlescapedPath)
- [`URL.hostname`](#urlhostname)
- [`URL.isAbs`](#urlisAbs)
- [`URL.joinPath`](#urljoinPath)
- [`URL.marshalBinary`](#urlmarshalBinary)
- [`URL.parse`](#urlparse)
- [`URL.port`](#urlport)
- [`URL.query`](#urlquery)
- [`URL.redacted`](#urlredacted)
- [`URL.requestURI`](#urlrequestURI)
- [`URL.resolveReference`](#urlresolveReference)
- [`URL.string`](#urlstring)
- [`URL.unmarshalBinary`](#urlunmarshalBinary)

## Userinfo

- [`user`](#user)
- [`userPassword`](#userPassword)
- [`Userinfo.username`](#userinfousername)
- [`Userinfo.password`](#userinfopassword)
- [`Userinfo.passwordSet`](#userinfopasswordSet)
- [`Userinfo.string`](#userinfostring)

# Functions

## joinPath

`joinPath` concatenates a base URL with additional path elements to form a complete URL.

### Parameters

`joinPath(baseURL string, elements []string) string`

- `baseURL`: the base URL to start with.
- `elements`: an array of path segments to join with the base URL.

### Example

```cel
grpc.federation.url.joinPath("https://example.com", ["path", "to", "resource"]) //=> "https://example.com/path/to/resource"
```

## pathEscape

`pathEscape` escapes the string so it can be safely placed inside a URL path segment.

### Parameters

`pathEscape(path string) string`

- `path`: the path segment to escape.

### Examples

```cel
grpc.federation.url.pathEscape("/path with spaces/")
```

## pathUnescape

`pathUnescape` unescapes a path segment.

### Parameters

`pathUnescape(path string) string`

- `path`: the path segment to unescape.

### Examples

```cel
grpc.federation.url.pathUnescape("/path%20with%20spaces/")
```

## queryEscape

`queryEscape` escapes the string so it can be safely placed inside a URL query.

### Parameters

`queryEscape(query string) string`

- `query`: the query string to escape.

### Examples

```cel
grpc.federation.url.queryEscape("query with spaces")
```

## queryUnescape

`queryUnescape` unescapes a query string.

### Parameters

`queryUnescape(query string) string`

- `query`: the query string to unescape.

### Examples

```cel
grpc.federation.url.queryUnescape("query%20with%20spaces")
```

# URL Methods

## parse

`parse` parses a URL and returns a URL struct.

### Parameters

`parse(url string) URL`

- `url`: the URL string to parse.

### Examples

```cel
grpc.federation.url.parse("https://www.example.com/path?query=value#fragment")
```

## parseRequestURI

`parseRequestURI` parses a URL in request context and returns a URL struct.

### Parameters

`parseRequestURI(requestURI string) URL`

- `requestURI`: the URI for a request.

### Examples

```cel
grpc.federation.url.parseRequestURI("/path?query=value#fragment")
```

## URL.scheme

`scheme` returns the scheme of the URL.

### Parameters

`URL.scheme() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com").scheme() //=> "https"
```

## URL.opaque

`opaque` returns the opaque part of the URL.

### Parameters

`URL.opaque() string`

### Examples

```cel
grpc.federation.url.parse("mailto:John.Doe@example.com").opaque() //=> "John.Doe@example.com"
```

## URL.userinfo

`userinfo` returns the user information in the URL.

### Parameters

`URL.userinfo() Userinfo`

### Examples

```cel
grpc.federation.url.parse("https://user:pass@example.com").userinfo() //=> grpc.federation.url.Userinfo{username: "user", password: "pass"}
```

## URL.host

`host` returns the host of the URL.

### Parameters

`URL.host() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com").host() //=> "example.com"
```

## URL.path

`path` returns the path of the URL.

### Parameters

`URL.path() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com/path").path() //=> "/path"
```

## URL.rawPath

`rawPath` returns the raw path of the URL.

### Parameters

`URL.rawPath() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com/%2Fpath%2F").rawPath() //=> "/%2Fpath%2F"
```

## URL.forceQuery

`forceQuery` returns the URL string with the ForceQuery field set to true.

### Parameters

`URL.forceQuery() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com").forceQuery()
```
## URL.rawQuery

`rawQuery` returns the raw query string of the URL.

### Parameters

`URL.rawQuery() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com?query=value").rawQuery() //=> "query=value"
```

## URL.fragment

`fragment` returns the fragment of the URL.

### Parameters

`URL.fragment() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com#fragment").fragment() //=> "fragment"
```

## URL.rawFragment

`rawFragment` returns the raw fragment of the URL.

### Parameters

`URL.rawFragment() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com#fragment").rawFragment() //=> "fragment"
```

## URL.escapedFragment

`escapedFragment` returns the escaped fragment of the URL.

### Parameters

`URL.escapedFragment() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com#frag ment").escapedFragment() //=> "frag%20ment"
```

## URL.escapedPath

`escapedPath` returns the escaped path of the URL.

### Parameters

`URL.escapedPath() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com/path with spaces").escapedPath() //=> "/path%20with%20spaces"
```

## URL.hostname

`hostname` returns the host name of the URL.

### Parameters

`URL.hostname() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com:8080").hostname() //=> "example.com"
```

## URL.isAbs

`isAbs` reports whether the URL is absolute.

### Parameters

`URL.isAbs() bool`

### Examples

```cel
grpc.federation.url.parse("https://example.com").isAbs() //=> true
grpc.federation.url.parse("/path").isAbs() //=> false
```

## URL.joinPath

`joinPath` joins the path elements with the URL path.

### Parameters

`URL.joinPath(elements []string) URL`

- `elements`: list of path elements to join with the URL path.

### Examples

```cel
grpc.federation.url.parse("https://example.com/path").joinPath(["to", "resource"]) //=> "https://example.com/path/to/resource"
```

## URL.marshalBinary

`marshalBinary` returns the URL in a binary format.

### Parameters

`URL.marshalBinary() bytes`

### Examples

```cel
grpc.federation.url.parse("https://example.com").marshalBinary()
```

## URL.parse

`parse` parses a URL string in the context of the current URL.

### Parameters

`URL.parse(ref string) URL`

- `ref`: the URL reference string to parse.

### Examples

```cel
grpc.federation.url.parse("https://example.com/base").parse("/resource") //=> "https://example.com/resource"
```

## URL.port

`port` returns the port of the URL.

### Parameters

`URL.port() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com:8080").port() //=> "8080"
```

## URL.query

`query` returns the parsed query parameters of the URL.

### Parameters

`URL.query() map<string, [string]>`

### Examples

```cel
grpc.federation.url.parse("https://example.com?key=value&key=another").query() //=> {"key": ["value", "another"]}
```

## URL.redacted

`redacted` returns the URL with password part redacted.

### Parameters

`URL.redacted() string`

### Examples

```cel
grpc.federation.url.parse("https://user:password@example.com").redacted() //=> "https://user:xxxxx@example.com"
```

## URL.requestURI

`requestURI` returns the request URI of the URL.

### Parameters

`URL.requestURI() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com/path?query=value").requestURI() //=> "/path?query=value"
```

## URL.resolveReference

`resolveReference` resolves a reference URL against the base URL.

### Parameters

`URL.resolveReference(ref URL) URL`

- `ref`: the reference URL to resolve.

### Examples

```cel
base := grpc.federation.url.parse("https://example.com/base")
ref := grpc.federation.url.parse("/resource")
base.resolveReference(ref) //=> "https://example.com/resource"
```

## URL.string

`string` returns the URL as a string.

### Parameters

`URL.string() string`

### Examples

```cel
grpc.federation.url.parse("https://example.com").string()
```

## URL.unmarshalBinary

`unmarshalBinary` unmarshals a binary format URL into a URL struct.

### Parameters

`URL.unmarshalBinary(data bytes) URL`

- `data`: binary format URL data.

### Examples

```cel
url := grpc.federation.url.parse("https://example.com")
binaryData := url.marshalBinary()
grpc.federation.url.parse("https://example.com").unmarshalBinary(binaryData)
```

# Userinfo Methods

## user

`user` returns a Userinfo object with the provided username.

### Parameters

`user(username string) Userinfo`

- `username`: the username for the userinfo.

### Examples

```cel
grpc.federation.url.user("user")
```

## userPassword

`userPassword` returns a Userinfo object with the provided username and password.

### Parameters

`userPassword(username string, password string) Userinfo`

- `username`: the username for the userinfo.
- `password`: the password for the userinfo.

### Examples

```cel
grpc.federation.url.userPassword("user", "password")
```

## Userinfo.username

`username` returns the username of the Userinfo.

### Parameters

`Userinfo.username() string`

### Examples

```cel
grpc.federation.url.user("user").username() //=> "user"
```

## Userinfo.password

`password` returns the password of the Userinfo, if set.

### Parameters

`Userinfo.password() string`

### Examples

```cel
grpc.federation.url.userPassword("user", "password").password() //=> "password"
```

## Userinfo.passwordSet

`passwordSet` returns true if the password is set for the Userinfo.

### Parameters

`Userinfo.passwordSet() bool`

### Examples

```cel
grpc.federation.url.userPassword("user", "password").passwordSet() //=> true
grpc.federation.url.user("user").passwordSet() //=> false
```

## Userinfo.string

`string` returns the Userinfo as a string.

### Parameters

`Userinfo.string() string`

### Examples

```cel
grpc.federation.url.userPassword("user", "password").string() //=> "user:password"
grpc.federation.url.user("user").string() //=> "user"
```
