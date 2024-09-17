package cel_test

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp/cmpopts"
	"net/url"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/go-cmp/cmp"
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
		//TODO user
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
			name: "parse force query with empty query",
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
			name: "parse raw fragment with escaped fragment",
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
			name: "join path",
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
		//{
		//	name: "query",
		//	expr: `grpc.federation.url.parse('https://example.com/path?query=1#fragment').query()`,
		//	cmp: func(got ref.Val) error {
		//		gotV, ok := got.(types.baseMap)
		//		if !ok {
		//			return fmt.Errorf("invalid result type: %T", got)
		//		}
		//		parse, err := url.Parse("https://example.com/path?query=1#fragment")
		//		if err != nil {
		//			return err
		//		}
		//		expected := parse.Query()
		//		if diff := cmp.Diff(gotV, expected); diff != "" {
		//			return fmt.Errorf("(-got, +want)\n%s", diff)
		//		}
		//		return nil
		//	},
		//},
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
		//TODO unmarshalBinary
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
