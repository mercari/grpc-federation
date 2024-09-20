package cel_test

import (
	"context"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"
	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
	"strings"
	"testing"
)

func TestStringsFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(ref.Val) error
	}{
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
			name: "cutPrefix",
			expr: "grpc.federation.strings.cutPrefix('abc', 'a')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimPrefix("abc", "a"))
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
				expected := types.String(strings.TrimSuffix("abc", "c"))
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
				expected := types.String(strings.Title("her royal highness"))
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
			expr: "grpc.federation.strings.trimRight('¡¡¡Hello, Gophers!!!', '¡¡¡')",
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.String(strings.TrimRight("¡¡¡Hello, Gophers!!!", "¡¡¡"))
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
	}

	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Lib(cellib.NewStringsLibrary(reg)),
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
