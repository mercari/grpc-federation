package cel_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestStringsFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(ref.Val) error
	}{
		// strings package
		{
			name: "clone",
			expr: "grpc.federation.strings.clone('abc')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String("abc")
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "compare",
			expr: "grpc.federation.strings.compare('abc', 'abd')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.Compare("abc", "abd"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "contains",
			expr: "grpc.federation.strings.contains('abc', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := strings.Contains("abc", "a")
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "containsAny",
			expr: "grpc.federation.strings.containsAny('fail', 'ui')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := strings.ContainsAny("fail", "ui")
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "containsRune",
			expr: "grpc.federation.strings.containsRune('aardvark', 97)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := strings.ContainsRune("aardvark", 97)
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "count",
			expr: "grpc.federation.strings.count('cheese', 'e')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.Count("cheese", "e"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "cut",
			expr: "grpc.federation.strings.cut('abc', 'b')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				before, after, _ := strings.Cut("abc", "b")
				expected := []string{before, after}
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "cutPrefix",
			expr: "grpc.federation.strings.cutPrefix('abc', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				c, _ := strings.CutPrefix("abc", "a")
				expected := types.String(c)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "cutSuffix",
			expr: "grpc.federation.strings.cutSuffix('abc', 'c')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				c, _ := strings.CutSuffix("abc", "c")
				expected := types.String(c)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "equalFold",
			expr: "grpc.federation.strings.equalFold('Go', 'go')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Bool(strings.EqualFold("Go", "go"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "fields",
			expr: "grpc.federation.strings.fields('a b c')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				expected := strings.Fields("a b c")
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "hasPrefix",
			expr: "grpc.federation.strings.hasPrefix('abc', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Bool(strings.HasPrefix("abc", "a"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "hasSuffix",
			expr: "grpc.federation.strings.hasSuffix('abc', 'c')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Bool(strings.HasSuffix("abc", "c"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "index",
			expr: "grpc.federation.strings.index('chicken', 'ken')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.Index("chicken", "ken"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "indexAny",
			expr: "grpc.federation.strings.indexAny('chicken', 'aeiou')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.IndexAny("chicken", "aeiou"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "indexByte",
			expr: "grpc.federation.strings.indexByte('golang', 'g')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.IndexByte("golang", 'g'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "indexRune",
			expr: "grpc.federation.strings.indexRune('golang', 111)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.IndexRune("golang", 111))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "join",
			expr: "grpc.federation.strings.join(['a', 'b', 'c'], ',')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.Join([]string{"a", "b", "c"}, ","))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "lastIndex",
			expr: "grpc.federation.strings.lastIndex('go gopher', 'go')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.LastIndex("go gopher", "go"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "lastIndexAny",
			expr: "grpc.federation.strings.lastIndexAny('go gopher', 'go')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.LastIndexAny("go gopher", "go"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "lastIndexByte",
			expr: "grpc.federation.strings.lastIndexByte('golang', 'g')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Int(strings.LastIndexByte("golang", 'g'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "repeat",
			expr: "grpc.federation.strings.repeat('a', 5)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.Repeat("a", 5))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "replace",
			expr: "grpc.federation.strings.replace('oink oink oink', 'k', 'ky', 2)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.Replace("oink oink oink", "k", "ky", 2))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "replaceAll",
			expr: "grpc.federation.strings.replaceAll('oink oink oink', 'k', 'ky')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.ReplaceAll("oink oink oink", "k", "ky"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "split",
			expr: "grpc.federation.strings.split('a,b,c', ',')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				expected := strings.Split("a,b,c", ",")
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "splitAfter",
			expr: "grpc.federation.strings.splitAfter('a,b,c', ',')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				expected := strings.SplitAfter("a,b,c", ",")
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "splitAfterN",
			expr: "grpc.federation.strings.splitAfterN('a,b,c', ',', 2)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				expected := strings.SplitAfterN("a,b,c", ",", 2)
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "splitN",
			expr: "grpc.federation.strings.splitN('a,b,c', ',', 2)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				strs := make([]string, gotV.Size().(types.Int))
				for i := 0; i < int(gotV.Size().(types.Int)); i++ {
					strs[i] = gotV.Get(types.Int(i)).(types.String).Value().(string)
				}
				expected := strings.SplitN("a,b,c", ",", 2)
				if diff := cmp.Diff(strs, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "title",
			expr: "grpc.federation.strings.title('her royal highness')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				c := cases.Title(language.English)
				expected := types.String(c.String("her royal highness"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "toLower",
			expr: "grpc.federation.strings.toLower('Gopher')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.ToLower("Gopher"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "toTitle",
			expr: "grpc.federation.strings.toTitle('loud noises')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.ToTitle("loud noises"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "toUpper",
			expr: "grpc.federation.strings.toUpper('loud noises')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.ToUpper("loud noises"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "toValidUTF8",
			expr: "grpc.federation.strings.toValidUTF8('abc', '\uFFFD')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.ToValidUTF8("abc", "\uFFFD"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trim",
			expr: "grpc.federation.strings.trim('¡¡¡Hello, Gophers!!!', '¡')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.Trim("¡¡¡Hello, Gophers!!!", "¡"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trimLeft",
			expr: "grpc.federation.strings.trimLeft('¡¡¡Hello, Gophers!!!', '¡')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimLeft("¡¡¡Hello, Gophers!!!", "¡"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trimPrefix",
			expr: "grpc.federation.strings.trimPrefix('¡¡¡Hello, Gophers!!!', '¡¡¡')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimPrefix("¡¡¡Hello, Gophers!!!", "¡¡¡"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trimRight",
			expr: "grpc.federation.strings.trimRight('¡¡¡Hello, Gophers!!!', '¡')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimRight("¡¡¡Hello, Gophers!!!", "¡"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trimSpace",
			expr: "grpc.federation.strings.trimSpace(' \t Hello, Gophers \t\t ')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimSpace(" \t Hello, Gophers \t\t "))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "trimSuffix",
			expr: "grpc.federation.strings.trimSuffix('¡¡¡Hello, Gophers!!!', '!!!')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimSuffix("¡¡¡Hello, Gophers!!!", "!!!"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// strconv package
		{
			name: "appendBool(bytes)",
			expr: "grpc.federation.strings.appendBool(b\"abc\", true)",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendBool([]byte("abc"), true))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendBool(string)",
			expr: "grpc.federation.strings.appendBool('ab', true)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendBool([]byte("ab"), true))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendFloat(bytes)",
			expr: "grpc.federation.strings.appendFloat(b\"abc\", 1.23, 'f', 2, 64)",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendFloat([]byte("abc"), 1.23, 'f', 2, 64))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendFloat(string)",
			expr: "grpc.federation.strings.appendFloat('ab', 1.23, 'f', 2, 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendFloat([]byte("ab"), 1.23, 'f', 2, 64))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendInt(bytes)",
			expr: "grpc.federation.strings.appendInt(b\"abc\", 123, 10)",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendInt([]byte("abc"), 123, 10))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendInt(string)",
			expr: "grpc.federation.strings.appendInt('ab', 123, 10)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendInt([]byte("ab"), 123, 10))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuote(bytes)",
			expr: "grpc.federation.strings.appendQuote(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuote([]byte("abc"), "a"))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuote(string)",
			expr: "grpc.federation.strings.appendQuote('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuote([]byte("ab"), "a"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRune(bytes)",
			expr: "grpc.federation.strings.appendQuoteRune(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuoteRune([]byte("abc"), 'a'))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRune(string)",
			expr: "grpc.federation.strings.appendQuoteRune('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuoteRune([]byte("ab"), 'a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRuneToASCII(bytes)",
			expr: "grpc.federation.strings.appendQuoteRuneToASCII(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuoteRuneToASCII([]byte("abc"), 'a'))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRuneToASCII(string)",
			expr: "grpc.federation.strings.appendQuoteRuneToASCII('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuoteRuneToASCII([]byte("ab"), 'a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRuneToGraphic(bytes)",
			expr: "grpc.federation.strings.appendQuoteRuneToGraphic(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuoteRuneToGraphic([]byte("abc"), 'a'))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteRuneToGraphic(string)",
			expr: "grpc.federation.strings.appendQuoteRuneToGraphic('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuoteRuneToGraphic([]byte("ab"), 'a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteToASCII(bytes)",
			expr: "grpc.federation.strings.appendQuoteToASCII(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuoteToASCII([]byte("abc"), "a"))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteToASCII(string)",
			expr: "grpc.federation.strings.appendQuoteToASCII('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuoteToASCII([]byte("ab"), "a"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteToGraphic(bytes)",
			expr: "grpc.federation.strings.appendQuoteToGraphic(b\"abc\", 'a')",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendQuoteToGraphic([]byte("abc"), "a"))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendQuoteToGraphic(string)",
			expr: "grpc.federation.strings.appendQuoteToGraphic('ab', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendQuoteToGraphic([]byte("ab"), "a"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendUint(bytes)",
			expr: "grpc.federation.strings.appendUint(b\"abc\", uint(123), 10)",
			cmp: func(got ref.Val) error {
				_, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bytes(strconv.AppendUint([]byte("abc"), 123, 10))
				if diff := cmp.Diff(got, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendUint(string)",
			expr: "grpc.federation.strings.appendUint('ab', uint(123), 10)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.AppendUint([]byte("ab"), 123, 10))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "atoi",
			expr: "grpc.federation.strings.atoi('123')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				i, err := strconv.Atoi("123")
				if err != nil {
					return err
				}
				expected := types.Int(i)
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "canBackquote",
			expr: "grpc.federation.strings.canBackquote('abc')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Bool(strconv.CanBackquote("abc"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "formatBool",
			expr: "grpc.federation.strings.formatBool(true)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strconv.FormatBool(true))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "formatComplex",
			expr: "grpc.federation.strings.formatComplex([1.23, 4.56], 'f', 2, 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.FormatComplex(1.23+4.56i, 'f', 2, 64))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "formatFloat",
			expr: "grpc.federation.strings.formatFloat(1.23, 'f', 2, 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.FormatFloat(1.23, 'f', 2, 64))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "formatInt",
			expr: "grpc.federation.strings.formatInt(123, 10)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.FormatInt(123, 10))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "formatUint",
			expr: "grpc.federation.strings.formatUint(uint(123), 10)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.FormatUint(123, 10))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "isGraphic",
			expr: "grpc.federation.strings.isGraphic('a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bool(strconv.IsGraphic('a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "isPrint",
			expr: "grpc.federation.strings.isPrint('a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.Bool(strconv.IsPrint('a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "itoa",
			expr: "grpc.federation.strings.itoa(123)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.Itoa(123))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parseBool",
			expr: "grpc.federation.strings.parseBool('true')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.ParseBool("true")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV, types.Bool(expected)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parseComplex",
			expr: "grpc.federation.strings.parseComplex('1.23', 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				c, err := strconv.ParseComplex("1.23", 64)
				if err != nil {
					return err
				}
				re, im := real(c), imag(c)
				expected := []ref.Val{types.Double(re), types.Double(im)}
				if diff := cmp.Diff(gotV.Get(types.Int(0)), expected[0]); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				if diff := cmp.Diff(gotV.Get(types.Int(1)), expected[1]); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parseFloat",
			expr: "grpc.federation.strings.parseFloat('1.23', 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.ParseFloat("1.23", 64)
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV, types.Double(expected)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parseInt",
			expr: "grpc.federation.strings.parseInt('123', 10, 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.ParseInt("123", 10, 64)
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV, types.Int(expected)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parseUint",
			expr: "grpc.federation.strings.parseUint('123', 10, 64)",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.ParseUint("123", 10, 64)
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV, types.Uint(expected)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quote",
			expr: "grpc.federation.strings.quote('abc')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.Quote("abc"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteRune",
			expr: "grpc.federation.strings.quoteRune('a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.QuoteRune('a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteRuneToASCII",
			expr: "grpc.federation.strings.quoteRuneToASCII('a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.QuoteRuneToASCII('a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteRuneToGraphic",
			expr: "grpc.federation.strings.quoteRuneToGraphic('a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.QuoteRuneToGraphic('a'))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteToASCII",
			expr: "grpc.federation.strings.quoteToASCII('abc')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.QuoteToASCII("abc"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quoteToGraphic",
			expr: "grpc.federation.strings.quoteToGraphic('abc')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected := types.String(strconv.QuoteToGraphic("abc"))
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "quotedPrefix",
			expr: "grpc.federation.strings.quotedPrefix('`or backquoted` with more trailing text')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.QuotedPrefix("`or backquoted` with more trailing text")
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
			name: "unquote",
			expr: "grpc.federation.strings.unquote('\"abc\"')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}

				expected, err := strconv.Unquote("\"abc\"")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV, types.String(expected)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "unquoteChar",
			expr: "grpc.federation.strings.unquoteChar('`a`', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				s := rune(gotV.Get(types.Int(0)).Value().(string)[0])
				b := gotV.Get(types.Int(1)).Value().(bool)
				t2 := gotV.Get(types.Int(2)).Value().(string)
				es, eb, et, err := strconv.UnquoteChar("`a`", 'a')
				if err != nil {
					return err
				}
				if diff := cmp.Diff(s, es); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				if diff := cmp.Diff(b, eb); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				if diff := cmp.Diff(t2, et); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Lib(cellib.NewStringsLibrary()),
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
