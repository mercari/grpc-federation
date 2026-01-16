package cel

import (
	"context"
	"reflect"
	"regexp"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const RegexpPackageName = "regexp"

var (
	RegexpType = types.NewObjectType(createRegexpName("Regexp"))
)

type Regexp struct {
	regexp.Regexp
}

func (r *Regexp) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return r, nil
}

func (r *Regexp) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErrFromString("grpc.federation.regexp: regexp type conversion does not support")
}

func (r *Regexp) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Regexp); ok {
		return types.Bool(r.String() == o.String())
	}
	return types.False
}

func (r *Regexp) Type() ref.Type {
	return RegexpType
}

func (r *Regexp) Value() any {
	return r
}

type RegexpLibrary struct {
}

func (lib *RegexpLibrary) LibraryName() string {
	return packageName(RegexpPackageName)
}

func createRegexpName(name string) string {
	return createName(RegexpPackageName, name)
}

func createRegexpID(name string) string {
	return createID(RegexpPackageName, name)
}

func (lib *RegexpLibrary) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createRegexpName("compile"),
			OverloadFunc(createRegexpID("compile_string_regexp"), []*cel.Type{cel.StringType}, RegexpType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					re, err := regexp.Compile(string(args[0].(types.String)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return &Regexp{
						Regexp: *re,
					}
				},
			),
		),
		BindFunction(
			createRegexpName("mustCompile"),
			OverloadFunc(createRegexpID("mustCompile_string_regexp"), []*cel.Type{cel.StringType}, RegexpType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return &Regexp{
						Regexp: *regexp.MustCompile(string(args[0].(types.String))),
					}
				},
			),
		),
		BindFunction(
			createRegexpName("quoteMeta"),
			OverloadFunc(createRegexpID("quoteMeta_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(regexp.QuoteMeta(string(args[0].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"findStringSubmatch",
			MemberOverloadFunc(createRegexpID("findStringSubmatch_regexp_string_strings"), RegexpType, []*cel.Type{cel.StringType}, cel.ListType(cel.StringType),
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.NewStringList(types.DefaultTypeAdapter, self.(*Regexp).FindStringSubmatch(string(args[0].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"matchString",
			MemberOverloadFunc(createRegexpID("match_string_regexp_string"), RegexpType, []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(*Regexp).MatchString(string(args[0].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"replaceAllString",
			MemberOverloadFunc(createRegexpID("replaceAllString_regexp_string_string_string"), RegexpType, []*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(*Regexp).ReplaceAllString(string(args[0].(types.String)), string(args[1].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createRegexpID("string_regexp"), RegexpType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(*Regexp).String())
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *RegexpLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
