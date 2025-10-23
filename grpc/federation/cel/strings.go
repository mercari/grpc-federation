package cel

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/cel-go/ext"
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

	extLib := ext.Strings()
	for _, funcOpts := range [][]cel.EnvOption{
		// strings package functions

		// define ext.Strings apis.
		BindExtMemberFunction(extLib, "charAt", "string_char_at_int", cel.StringType, []*cel.Type{cel.IntType}, cel.StringType),
		BindExtMemberFunction(extLib, "indexOf", "string_index_of_string", cel.StringType, []*cel.Type{cel.StringType}, cel.IntType),
		BindExtMemberFunction(extLib, "indexOf", "string_index_of_string_int", cel.StringType, []*cel.Type{cel.StringType, cel.IntType}, cel.IntType),
		BindExtMemberFunction(extLib, "lastIndexOf", "string_last_index_of_string", cel.StringType, []*cel.Type{cel.StringType}, cel.IntType),
		BindExtMemberFunction(extLib, "lastIndexOf", "string_last_index_of_string_int", cel.StringType, []*cel.Type{cel.StringType, cel.IntType}, cel.IntType),
		BindExtMemberFunction(extLib, "lowerAscii", "string_lower_ascii", cel.StringType, []*cel.Type{}, cel.StringType),
		BindExtMemberFunction(extLib, "replace", "string_replace_string_string", cel.StringType, []*cel.Type{cel.StringType, cel.StringType}, cel.StringType),
		BindExtMemberFunction(extLib, "replace", "string_replace_string_string_int", cel.StringType, []*cel.Type{cel.StringType, cel.StringType, cel.IntType}, cel.StringType),
		BindExtMemberFunction(extLib, "split", "string_split_string", cel.StringType, []*cel.Type{cel.StringType}, cel.ListType(cel.StringType)),
		BindExtMemberFunction(extLib, "split", "string_split_string_int", cel.StringType, []*cel.Type{cel.StringType, cel.IntType}, cel.ListType(cel.StringType)),
		BindExtMemberFunction(extLib, "substring", "string_substring_int", cel.StringType, []*cel.Type{cel.IntType}, cel.StringType),
		BindExtMemberFunction(extLib, "substring", "string_substring_int_int", cel.StringType, []*cel.Type{cel.IntType, cel.IntType}, cel.StringType),
		BindExtMemberFunction(extLib, "trim", "string_trim", cel.StringType, []*cel.Type{}, cel.StringType),
		BindExtMemberFunction(extLib, "upperAscii", "string_upper_ascii", cel.StringType, []*cel.Type{}, cel.StringType),
		BindExtMemberFunction(extLib, "format", "string_format_list_string", cel.StringType, []*cel.Type{cel.ListType(cel.DynType)}, cel.StringType),
		BindExtMemberFunction(extLib, "join", "list_join", cel.ListType(cel.StringType), []*cel.Type{}, cel.StringType),
		BindExtMemberFunction(extLib, "join", "list_join_string", cel.ListType(cel.StringType), []*cel.Type{cel.StringType}, cel.StringType),
		BindExtMemberFunction(extLib, "reverse", "string_reverse", cel.StringType, []*cel.Type{}, cel.StringType),
		BindExtFunction(extLib, "strings.quote", "strings_quote", []*cel.Type{cel.StringType}, cel.StringType),

		// add gRPC Federation standard apis.
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

		// strconv package functions
		BindFunction(
			createStringsName("appendBool"),
			OverloadFunc(createStringsID("appendBool_bytes_bool_bytes"), []*cel.Type{cel.BytesType, cel.BoolType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendBool(args[0].(types.Bytes).Value().([]byte), args[1].(types.Bool).Value().(bool)))
				},
			),
			OverloadFunc(createStringsID("appendBool_string_bool_string"), []*cel.Type{cel.StringType, cel.BoolType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendBool([]byte(args[0].(types.String).Value().(string)), args[1].(types.Bool).Value().(bool)))
				},
			),
		),
		BindFunction(
			createStringsName("appendFloat"),
			OverloadFunc(createStringsID("appendFloat_bytes_float64_string_int_int_bytes"), []*cel.Type{cel.BytesType, cel.DoubleType, cel.StringType, cel.IntType, cel.IntType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendFloat(args[0].(types.Bytes).Value().([]byte), args[1].(types.Double).Value().(float64), args[2].(types.String).Value().(string)[0], int(args[3].(types.Int).Value().(int64)), int(args[4].(types.Int).Value().(int64))))
				},
			),
			OverloadFunc(createStringsID("appendFloat_string_float64_int_string_int_string"), []*cel.Type{cel.StringType, cel.DoubleType, cel.StringType, cel.IntType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendFloat([]byte(args[0].(types.String).Value().(string)), args[1].(types.Double).Value().(float64), args[2].(types.String).Value().(string)[0], int(args[3].(types.Int).Value().(int64)), int(args[4].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("appendInt"),
			OverloadFunc(createStringsID("appendInt_bytes_int_int_bytes"), []*cel.Type{cel.BytesType, cel.IntType, cel.IntType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendInt(args[0].(types.Bytes).Value().([]byte), args[1].(types.Int).Value().(int64), int(args[2].(types.Int).Value().(int64))))
				},
			),
			OverloadFunc(createStringsID("appendInt_string_int_int_string"), []*cel.Type{cel.StringType, cel.IntType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendInt([]byte(args[0].(types.String).Value().(string)), args[1].(types.Int).Value().(int64), int(args[2].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuote"),
			OverloadFunc(createStringsID("appendQuote_bytes_string_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuote(args[0].(types.Bytes).Value().([]byte), args[1].(types.String).Value().(string)))
				},
			),
			OverloadFunc(createStringsID("appendQuote_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuote([]byte(args[0].(types.String).Value().(string)), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuoteRune"),
			OverloadFunc(createStringsID("appendQuoteRune_bytes_int_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuoteRune(args[0].(types.Bytes).Value().([]byte), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
			OverloadFunc(createStringsID("appendQuoteRune_string_int_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuoteRune([]byte(args[0].(types.String).Value().(string)), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuoteRuneToASCII"),
			OverloadFunc(createStringsID("appendQuoteRuneToASCII_bytes_string_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuoteRuneToASCII(args[0].(types.Bytes).Value().([]byte), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
			OverloadFunc(createStringsID("appendQuoteRuneToASCII_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuoteRuneToASCII([]byte(args[0].(types.String).Value().(string)), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuoteRuneToGraphic"),
			OverloadFunc(createStringsID("appendQuoteRuneToGraphic_bytes_string_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuoteRuneToGraphic(args[0].(types.Bytes).Value().([]byte), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
			OverloadFunc(createStringsID("appendQuoteRuneToGraphic_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuoteRuneToGraphic([]byte(args[0].(types.String).Value().(string)), rune(args[1].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuoteToASCII"),
			OverloadFunc(createStringsID("appendQuoteToASCII_bytes_string_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuoteToASCII(args[0].(types.Bytes).Value().([]byte), args[1].(types.String).Value().(string)))
				},
			),
			OverloadFunc(createStringsID("appendQuoteToASCII_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuoteToASCII([]byte(args[0].(types.String).Value().(string)), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("appendQuoteToGraphic"),
			OverloadFunc(createStringsID("appendQuoteToGraphic_bytes_string_bytes"), []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendQuoteToGraphic(args[0].(types.Bytes).Value().([]byte), args[1].(types.String).Value().(string)))
				},
			),
			OverloadFunc(createStringsID("appendQuoteToGraphic_string_string_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendQuoteToGraphic([]byte(args[0].(types.String).Value().(string)), args[1].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("appendUint"),
			OverloadFunc(createStringsID("appendUint_bytes_uint_int_bytes"), []*cel.Type{cel.BytesType, cel.UintType, cel.IntType}, cel.BytesType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bytes(strconv.AppendUint(args[0].(types.Bytes).Value().([]byte), args[1].(types.Uint).Value().(uint64), int(args[2].(types.Int).Value().(int64))))
				},
			),
			OverloadFunc(createStringsID("appendUint_string_uint_int_string"), []*cel.Type{cel.StringType, cel.UintType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.AppendUint([]byte(args[0].(types.String).Value().(string)), args[1].(types.Uint).Value().(uint64), int(args[2].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("atoi"),
			OverloadFunc(createStringsID("atoi_string_int"), []*cel.Type{cel.StringType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					i, err := strconv.Atoi(args[0].(types.String).Value().(string))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.Int(i)
				},
			),
		),
		BindFunction(
			createStringsName("canBackquote"),
			OverloadFunc(createStringsID("canBackquote_string_bool"), []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strconv.CanBackquote(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("formatBool"),
			OverloadFunc(createStringsID("formatBool_bool_string"), []*cel.Type{cel.BoolType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.FormatBool(args[0].(types.Bool).Value().(bool)))
				},
			),
		),
		BindFunction(
			createStringsName("formatComplex"),
			OverloadFunc(createStringsID("formatComplex_complex128_string_int_int_string"), []*cel.Type{cel.ListType(cel.DoubleType), cel.StringType, cel.IntType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					cList := args[0].(traits.Lister)
					c := complex(cList.Get(types.IntZero).(types.Double).Value().(float64), cList.Get(types.IntOne).(types.Double).Value().(float64))
					return types.String(strconv.FormatComplex(c, args[1].(types.String).Value().(string)[0], int(args[2].(types.Int).Value().(int64)), int(args[3].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("formatFloat"),
			OverloadFunc(createStringsID("formatFloat_float64_string_int_int_string"), []*cel.Type{cel.DoubleType, cel.StringType, cel.IntType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.FormatFloat(args[0].(types.Double).Value().(float64), args[1].(types.String).Value().(string)[0], int(args[2].(types.Int).Value().(int64)), int(args[3].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("formatInt"),
			OverloadFunc(createStringsID("formatInt_int_int_string"), []*cel.Type{cel.IntType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.FormatInt(args[0].(types.Int).Value().(int64), int(args[1].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("formatUint"),
			OverloadFunc(createStringsID("formatUint_uint_int_string"), []*cel.Type{cel.UintType, cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.FormatUint(args[0].(types.Uint).Value().(uint64), int(args[1].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("isGraphic"),
			OverloadFunc(createStringsID("isGraphic_byte_bool"), []*cel.Type{cel.BytesType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strconv.IsGraphic(rune(args[0].(types.Bytes).Value().(byte))))
				},
			),
			OverloadFunc(createStringsID("isGraphic_string_bool"), []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strconv.IsGraphic(rune(args[0].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("isPrint"),
			OverloadFunc(createStringsID("isPrint_byte_bool"), []*cel.Type{cel.BytesType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strconv.IsPrint(rune(args[0].(types.Bytes).Value().(byte))))
				},
			),
			OverloadFunc(createStringsID("isPrint_string_bool"), []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(strconv.IsPrint(rune(args[0].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("itoa"),
			OverloadFunc(createStringsID("itoa_int_string"), []*cel.Type{cel.IntType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.Itoa(int(args[0].(types.Int).Value().(int64))))
				},
			),
		),
		BindFunction(
			createStringsName("parseBool"),
			OverloadFunc(createStringsID("parseBool_string_bool"), []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					b, err := strconv.ParseBool(args[0].(types.String).Value().(string))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.Bool(b)
				},
			),
		),
		BindFunction(
			createStringsName("parseComplex"),
			OverloadFunc(createStringsID("parseComplex_string_int_complex128"), []*cel.Type{cel.StringType, cel.IntType}, cel.ListType(cel.DoubleType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					c, err := strconv.ParseComplex(args[0].(types.String).Value().(string), int(args[1].(types.Int).Value().(int64)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.NewDynamicList(types.DefaultTypeAdapter, []ref.Val{types.Double(real(c)), types.Double(imag(c))})
				},
			),
		),
		BindFunction(
			createStringsName("parseFloat"),
			OverloadFunc(createStringsID("parseFloat_string_int_float64"), []*cel.Type{cel.StringType, cel.IntType}, cel.DoubleType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					f, err := strconv.ParseFloat(args[0].(types.String).Value().(string), int(args[1].(types.Int).Value().(int64)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.Double(f)
				},
			),
		),
		BindFunction(
			createStringsName("parseInt"),
			OverloadFunc(createStringsID("parseInt_string_int_int_int"), []*cel.Type{cel.StringType, cel.IntType, cel.IntType}, cel.IntType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					i, err := strconv.ParseInt(args[0].(types.String).Value().(string), int(args[1].(types.Int).Value().(int64)), int(args[2].(types.Int).Value().(int64)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.Int(i)
				},
			),
		),
		BindFunction(
			createStringsName("parseUint"),
			OverloadFunc(createStringsID("parseUint_string_int_int__uint"), []*cel.Type{cel.StringType, cel.IntType, cel.IntType}, cel.UintType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					u, err := strconv.ParseUint(args[0].(types.String).Value().(string), int(args[1].(types.Int).Value().(int64)), int(args[2].(types.Int).Value().(int64)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.Uint(u)
				},
			),
		),
		BindFunction(
			createStringsName("quote"),
			OverloadFunc(createStringsID("quote_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.Quote(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("quoteRune"),
			OverloadFunc(createStringsID("quoteRune_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.QuoteRune(rune(args[0].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("quoteRuneToASCII"),
			OverloadFunc(createStringsID("quoteRuneToASCII_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.QuoteRuneToASCII(rune(args[0].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("quoteRuneToGraphic"),
			OverloadFunc(createStringsID("quoteRuneToGraphic_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.QuoteRuneToGraphic(rune(args[0].(types.String).Value().(string)[0])))
				},
			),
		),
		BindFunction(
			createStringsName("quoteToASCII"),
			OverloadFunc(createStringsID("quoteToASCII_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.QuoteToASCII(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("quoteToGraphic"),
			OverloadFunc(createStringsID("quoteToGraphic_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(strconv.QuoteToGraphic(args[0].(types.String).Value().(string)))
				},
			),
		),
		BindFunction(
			createStringsName("quotedPrefix"),
			OverloadFunc(createStringsID("quotedPrefix_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					s, err := strconv.QuotedPrefix(args[0].(types.String).Value().(string))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.String(s)
				},
			),
		),
		BindFunction(
			createStringsName("unquote"),
			OverloadFunc(createStringsID("unquote_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					s, err := strconv.Unquote(args[0].(types.String).Value().(string))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.String(s)
				},
			),
		),
		BindFunction(
			createStringsName("unquoteChar"),
			OverloadFunc(createStringsID("unquoteChar_string_byte_string_bool_string"), []*cel.Type{cel.StringType, cel.StringType}, cel.ListType(cel.AnyType),
				func(_ context.Context, args ...ref.Val) ref.Val {
					s, b, t, err := strconv.UnquoteChar(args[0].(types.String).Value().(string), args[1].(types.String).Value().(string)[0])
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return types.NewDynamicList(types.DefaultTypeAdapter, []ref.Val{types.String(s), types.Bool(b), types.String(t)})
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
