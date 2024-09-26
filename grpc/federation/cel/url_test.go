package cel_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestURLFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(ref.Val) error
	}{
		{
			name: "join path",
			expr: `grpc.federation.url.joinPath('https://example.com/path?query=1#fragment', ['/new'])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := url.JoinPath("https://example.com/path?query=1#fragment", "/new")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "path escape",
			expr: `grpc.federation.url.pathEscape('あ /')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := url.PathEscape("あ /")
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "path unescape",
			expr: `grpc.federation.url.pathUnescape('%E3%81%82%20%2F')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := url.PathUnescape("%E3%81%82%20%2F")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "query escape",
			expr: `grpc.federation.url.queryEscape('あ /')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := url.QueryEscape("あ /")
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "query unescape",
			expr: `grpc.federation.url.queryUnescape('%E3%81%82%20%2F')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := url.QueryUnescape("%E3%81%82%20%2F")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// url functions
		{
			name: "parse",
			expr: "grpc.federation.url.parse('https://example.com/path?query=1#fragment')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				expected, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse request uri",
			expr: `grpc.federation.url.parseRequestURI('https://example.com/path?query=1#fragment')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				expected, err := url.ParseRequestURI("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse scheme",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').scheme()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Scheme
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse opaque",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').opaque()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Opaque
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse user",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').userinfo()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.Userinfo)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotUser, err := gotV.GoUserinfo()
				if err != nil {
					return err
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.User
				if diff := userinfoDiff(gotUser, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse host",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').host()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Host
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse path",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').path()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Path
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse raw path",
			expr: `grpc.federation.url.parse('https://example.com/%E3%81%82%20%2Fpath?query=1#fragment').rawPath()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/%E3%81%82%20%2Fpath?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.RawPath
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse force query",
			expr: `grpc.federation.url.parse('https://example.com/path?').forceQuery()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?")
				if err != nil {
					return err
				}
				parse.ForceQuery = true
				expected := parse.String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse raw query",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').rawQuery()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.RawQuery
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse fragment",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').fragment()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Fragment
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse raw fragment",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#frag%20ment').rawFragment()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#frag%20ment")
				if err != nil {
					return err
				}
				expected := parse.RawFragment
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse escaped fragment",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').escapedFragment()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.EscapedFragment()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse escaped path",
			expr: `grpc.federation.url.parse('https://example.com/pa th?query=1#fragment').escapedPath()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/pa th?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.EscapedPath()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse hostname",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').hostname()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Hostname()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "is abs",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').isAbs()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.IsAbs()
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "url join path",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').joinPath(['/new'])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.JoinPath("/new")
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "marshal binary",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').marshalBinary()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected, err := parse.MarshalBinary()
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Value(), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "url parse",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').parse('/relativePath')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected, err := parse.Parse("/relativePath")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "port",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').port()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Port()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "query",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').query()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotQueryMap := make(map[string][]string)
				it := gotV.Iterator()
				for it.HasNext() == types.True {
					key := it.Next()
					keyStr := string(key.(types.String))
					valueList := gotV.Get(key).(traits.Lister)

					var values []string
					for i := int64(0); i < int64(valueList.Size().(types.Int)); i++ {
						values = append(values, string(valueList.Get(types.Int(i)).(types.String)))
					}
					gotQueryMap[keyStr] = values
				}

				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := map[string][]string(parse.Query())

				if diff := cmp.Diff(gotQueryMap, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "redacted",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').redacted()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.Redacted()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "request uri",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').requestURI()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.RequestURI()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "resolve reference",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').resolveReference(grpc.federation.url.parse('/relativePath'))`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				ref, err := url.Parse("/relativePath")
				if err != nil {
					return err
				}
				expected := parse.ResolveReference(ref)
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "string",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').string()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse.String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "unmarshal binary",
			expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').unmarshalBinary(grpc.federation.url.parse('https://example.com/path?query=1#fragment').marshalBinary())`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.URL)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotURL, err := gotV.GoURL()
				if err != nil {
					return err
				}
				parse, err := url.Parse("https://example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := parse
				if diff := cmp.Diff(gotURL, *expected, cmpopts.IgnoreUnexported(url.URL{})); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// userinfo tests
		{
			name: "user",
			expr: `grpc.federation.url.user('username')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.Userinfo)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotUser, err := gotV.GoUserinfo()
				if err != nil {
					return err
				}
				expected := url.User("username")
				if diff := userinfoDiff(gotUser, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "user password to user info",
			expr: `grpc.federation.url.userPassword('username', 'password')`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.Value().(*cellib.Userinfo)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				gotUser, err := gotV.GoUserinfo()
				if err != nil {
					return err
				}
				expected := url.UserPassword("username", "password")
				if diff := userinfoDiff(gotUser, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "user username",
			expr: `grpc.federation.url.parse('https://username:password@example.com/path?query=1#fragment').userinfo().username()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				user, err := url.Parse("https://username:password@example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := user.User.Username()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "user password",
			expr: `grpc.federation.url.parse('https://username:password@example.com/path?query=1#fragment').userinfo().password()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				user, err := url.Parse("https://username:password@example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected, _ := user.User.Password()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "user password set",
			expr: `grpc.federation.url.parse('https://username:password@example.com/path?query=1#fragment').userinfo().passwordSet()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				user, err := url.Parse("https://username:password@example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				_, expected := user.User.Password()
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "user string",
			expr: `grpc.federation.url.parse('https://username:password@example.com/path?query=1#fragment').userinfo().string()`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				user, err := url.Parse("https://username:password@example.com/path?query=1#fragment")
				if err != nil {
					return err
				}
				expected := user.User.String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}

	reg, err := types.NewRegistry(new(cellib.URL), new(cellib.Userinfo))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Lib(cellib.NewURLLibrary(reg)),
			)
			if err != nil {
				t.Fatal(err)
			}
			ast, iss := env.Compile(test.expr)
			if iss.Err() != nil {
				t.Fatal(iss.Err())
			}
			program, err := env.Program(ast)
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]any{cellib.ContextVariableName: cellib.NewContextValue(context.Background())}
			for k, v := range test.args {
				args[k] = v
			}
			out, _, err := program.Eval(args)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// UserinfoDiff compares two Userinfo structs and returns a diff string similar to cmp.Diff.
func userinfoDiff(got, expected *url.Userinfo) string {
	var diffs []string

	if got.Username() != expected.Username() {
		diffs = append(diffs, fmt.Sprintf("username: got %q, want %q", got.Username(), expected.Username()))
	}

	gotPassword, gotPasswordSet := got.Password()
	expectedPassword, expectedPasswordSet := expected.Password()
	if gotPasswordSet != expectedPasswordSet {
		diffs = append(diffs, fmt.Sprintf("password set: got %t, want %t", gotPasswordSet, expectedPasswordSet))
	}
	if gotPassword != expectedPassword {
		diffs = append(diffs, fmt.Sprintf("password: got %q, want %q", gotPassword, expectedPassword))
	}

	if len(diffs) == 0 {
		return ""
	}

	return "(-got, +want)\n" + strings.Join(diffs, "\n")
}
