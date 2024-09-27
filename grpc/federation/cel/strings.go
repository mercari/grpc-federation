package cel

import (
	"context"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const StringsPackageName = "strings"

var _ cel.SingletonLibrary = new(StringsLibrary)

type StringsLibrary struct {
}

func NewStringsLibrary() *StringsLibrary {
	return &StringsLibrary{}
}

func (lib *StringsLibrary) LibraryName() string {
	return packageName(StringsPackageName)
}

func createStringsName(name string) string {
	return createName(StringsPackageName, name)
}

func createStringsID(name string) string {
	return createID(StringsPackageName, name)
}

func (lib *StringsLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{}

	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createStringsName("clone"),
			OverloadFunc(createStringsID("clone_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return args[0].(types.String)
				},
			),
		),
		BindFunction(
			createStringsName("compare"),
			OverloadFunc(createStringsID("compare_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return args[0].(types.String).Compare(args[1].(types.String))
				},
			),
		),
		BindFunction(
			createStringsName("contains"),
			OverloadFunc(createStringsID("contains_string_string_bool"), []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strings.Contains(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("containsAny"),
			OverloadFunc(createStringsID("containsAny_string_string_bool"), []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strings.ContainsAny(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("containsRune"),
			OverloadFunc(createStringsID("containsRune_string_int_bool"), []*cel.Type{cel.StringType, cel.IntType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					str := args[0].(types.String).Value().(string)
					r := rune(args[1].(types.Int).Value().(int64))
					return types.Bool(strings.ContainsRune(str, r))
				},
			),
		),
		BindFunction(
			createStringsName("count"),
			OverloadFunc(createStringsID("count_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.Count(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("cut"),
			OverloadFunc(createStringsID("cut_string_string_strings"), []*cel.Type{cel.StringType, cel.StringType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					str := args[0].(types.String).Value().(string)
					sep := args[1].(types.String).Value().(string)
					before, after, _ := strings.Cut(str, sep)
					return types.NewStringList(types.DefaultTypeAdapter, []string{before, after})
				},
			),
		),
		BindFunction(
			createStringsName("cutPrefix"),
			OverloadFunc(createStringsID("cutPrefix_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					str := args[0].(types.String).Value().(string)
					prefix := args[1].(types.String).Value().(string)
					after, _ := strings.CutPrefix(str, prefix)
					return types.String(after)
				},
			),
		),
		BindFunction(
			createStringsName("cutSuffix"),
			OverloadFunc(createStringsID("cutSuffix_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					str := args[0].(types.String).Value().(string)
					suffix := args[1].(types.String).Value().(string)
					before, _ := strings.CutSuffix(str, suffix)
					return types.String(before)
				},
			),
		),
		BindFunction(
			createStringsName("equalFold"),
			OverloadFunc(createStringsID("equalFold_string_string_bool"), []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strings.EqualFold(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("fields"),
			OverloadFunc(createStringsID("fields_string_strings"), []*cel.Type{cel.StringType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					adapter := types.DefaultTypeAdapter
					return types.NewStringList(adapter, strings.Fields(args[0].(types.String).Value().(string)))
				},
			),
		),
		// func FieldsFunc(s string, f func(rune) bool) []string is not implemented because it has a function argument.
		BindFunction(
			createStringsName("hasPrefix"),
			OverloadFunc(createStringsID("hasPrefix_string_string_bool"), []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strings.HasPrefix(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("hasSuffix"),
			OverloadFunc(createStringsID("hasSuffix_string_string_bool"), []*cel.Type{cel.StringType, cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strings.HasSuffix(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("index"),
			OverloadFunc(createStringsID("index_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.Index(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("indexAny"),
			OverloadFunc(createStringsID("indexAny_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.IndexAny(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("indexByte"),
			OverloadFunc(createStringsID("indexByte_string_byte_int"), []*cel.Type{cel.StringType, cel.BytesType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.IndexByte(args[0].(types.String).Value().(string), args[1].(types.Bytes).Value().(byte)))
				},
			),
			OverloadFunc(createStringsID("indexByte_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.IndexByte(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)[0]))
				},
			),
		),
		// func IndexFunc(s string, f func(rune) bool) int　is not implemented because it has a function argument.
		BindFunction(
			createStringsName("indexRune"),
			OverloadFunc(createStringsID("indexRune_string_int_int"), []*cel.Type{cel.StringType, cel.IntType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.IndexRune(args[0].(types.String).Value().(string), rune(args[1].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("join"),
			OverloadFunc(createStringsID("join_strings_string_string"), []*cel.Type{cel.ListType(cel.StringType), cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					elems := args[0].(traits.Lister)
					strs := make([]string, elems.Size().(types.Int))
					for i := types.IntZero; i < elems.Size().(types.Int); i++ {
						strs[i] = elems.Get(i).(types.String).Value().(string)
					}
					return types.String(strings.Join(strs, args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("lastIndex"),
			OverloadFunc(createStringsID("lastIndex_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.LastIndex(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("lastIndexAny"),
			OverloadFunc(createStringsID("lastIndexAny_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.LastIndexAny(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("lastIndexByte"),
			OverloadFunc(createStringsID("lastIndexByte_string_byte_int"), []*cel.Type{cel.StringType, cel.BytesType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.LastIndexByte(args[0].(types.String).Value().(string), args[1].(types.Bytes).Value().(byte)))
				},
			),
			OverloadFunc(createStringsID("lastIndexByte_string_string_int"), []*cel.Type{cel.StringType, cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Int(strings.LastIndexByte(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)[0]))
				},
			),
		),
		// func LastIndexFunc(s string, f func(rune) bool) int　is not implemented because it has a function argument.
		// func Map(mapping func(rune) rune, s string) string　is not implemented because it has a function argument.
		BindFunction(
			createStringsName("repeat"),
			OverloadFunc(createStringsID("repeat_string_int_string"), []*cel.Type{cel.StringType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.Repeat(args[0].(types.String).Value().(string), int(args[1].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("replace"),
			OverloadFunc(createStringsID("replace_string_string_string_int_string"), []*cel.Type{cel.StringType, cel.StringType, cel.StringType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.Replace(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string), args[2].(types.String).Value().(string), int(args[3].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("replaceAll"),
			OverloadFunc(createStringsID("replaceAll_string_string_string_string"), []*cel.Type{cel.StringType, cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.ReplaceAll(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string), args[2].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("split"),
			OverloadFunc(createStringsID("split_string_string_strings"), []*cel.Type{cel.StringType, cel.StringType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					adapter := types.DefaultTypeAdapter
					return types.NewStringList(adapter, strings.Split(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("splitAfter"),
			OverloadFunc(createStringsID("splitAfter_string_string_strings"), []*cel.Type{cel.StringType, cel.StringType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					adapter := types.DefaultTypeAdapter
					return types.NewStringList(adapter, strings.SplitAfter(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("splitAfterN"),
			OverloadFunc(createStringsID("splitAfterN_string_string_int_strings"), []*cel.Type{cel.StringType, cel.StringType, cel.IntType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					adapter := types.DefaultTypeAdapter
					return types.NewStringList(adapter, strings.SplitAfterN(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string), int(args[2].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("splitN"),
			OverloadFunc(createStringsID("splitN_string_string_int_strings"), []*cel.Type{cel.StringType, cel.StringType, cel.IntType}, cel.ListType(cel.StringType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					adapter := types.DefaultTypeAdapter
					return types.NewStringList(adapter, strings.SplitN(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string), int(args[2].(types.Int).Value().(int64))))
				},
			),
		),
		// strings.Title is deprecated. So, we use golang.org/x/text/cases.Title instead.
		BindFunction(
			createStringsName("title"),
			OverloadFunc(createStringsID("title_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					c := cases.Title(language.English)
					return types.String(c.String(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("toLower"),
			OverloadFunc(createStringsID("toLower_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.ToLower(args[0].(types.String).Value().(string)))
				},
			),
		),
		// func ToLowerSpecial(c unicode.SpecialCase, s string) string is not implemented because unicode.SpecialCase is not supported.
		BindFunction(
			createStringsName("toTitle"),
			OverloadFunc(createStringsID("toTitle_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.ToTitle(args[0].(types.String).Value().(string)))
				},
			),
		),
		// func ToTitleSpecial(c unicode.SpecialCase, s string) string is not implemented because unicode.SpecialCase is not supported.
		BindFunction(
			createStringsName("toUpper"),
			OverloadFunc(createStringsID("toUpper_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.ToUpper(args[0].(types.String).Value().(string)))
				},
			),
		),
		// func ToUpperSpecial(c unicode.SpecialCase, s string) string is not implemented because unicode.SpecialCase is not supported.
		BindFunction(
			createStringsName("toValidUTF8"),
			OverloadFunc(createStringsID("toValidUTF8_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.ToValidUTF8(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("trim"),
			OverloadFunc(createStringsID("trim_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.Trim(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		// func TrimFunc(s string, f func(rune) bool) string is not implemented because it has a function argument.
		BindFunction(
			createStringsName("trimLeft"),
			OverloadFunc(createStringsID("trimLeft_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.TrimLeft(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		// func TrimLeftFunc(s string, f func(rune) bool) string is not implemented because it has a function argument.
		BindFunction(
			createStringsName("trimPrefix"),
			OverloadFunc(createStringsID("trimPrefix_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.TrimPrefix(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("trimRight"),
			OverloadFunc(createStringsID("trimRight_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.TrimRight(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
		// func TrimRightFunc(s string, f func(rune) bool) string is not implemented because it has a function argument.
		BindFunction(
			createStringsName("trimSpace"),
			OverloadFunc(createStringsID("trimSpace_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.TrimSpace(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("trimSuffix"),
			OverloadFunc(createStringsID("trimSuffix_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strings.TrimSuffix(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)))
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}

	return opts
}

func (lib *StringsLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
