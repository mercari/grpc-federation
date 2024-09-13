package cel

import (
	"context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"net/url"
)

const URLPackageName = "url"

var (
	URLType      = cel.ObjectType("grpc.federation.url.URL")
	UserinfoType = cel.ObjectType("grpc.federation.url.Userinfo")
)

func (x *URL) GoURL() (url.URL, error) {
	var user *url.Userinfo
	if u := x.GetUser(); u != nil {
		if u2, err := u.GoUserinfo(); err != nil {
			return url.URL{}, err
		} else {
			user = u2
		}
	}

	return url.URL{
		Scheme:     x.GetScheme(),
		Opaque:     x.GetOpaque(),
		User:       user,
		Host:       x.GetHost(),
		Path:       x.GetPath(),
		RawPath:    x.GetRawPath(),
		ForceQuery: x.GetForceQuery(),
		RawQuery:   x.GetRawQuery(),
		Fragment:   x.GetFragment(),
	}, nil
}

func (x *Userinfo) GoUserinfo() (*url.Userinfo, error) {
	if x.GetPasswordSet() {
		return url.UserPassword(x.GetUsername(), x.GetPassword()), nil
	} else {
		return url.User(x.GetUsername()), nil
	}
}

var _ cel.SingletonLibrary = new(URLLibrary)

type URLLibrary struct {
	typeAdapter types.Adapter
}

func NewURLLibrary(typeAdapter types.Adapter) *URLLibrary {
	return &URLLibrary{
		typeAdapter: typeAdapter,
	}
}

func (lib *URLLibrary) LibraryName() string {
	return packageName(URLPackageName)
}

func createURLName(name string) string {
	return createName(URLPackageName, name)
}

func createURLID(name string) string {
	return createID(URLPackageName, name)
}

func (lib *URLLibrary) refToGoURLValue(v ref.Val) (url.URL, error) {
	return v.Value().(*URL).GoURL()
}

func (lib *URLLibrary) toURLValue(v url.URL) ref.Val {
	var userinfo *Userinfo
	if v.User != nil {
		password, hasPassword := v.User.Password()
		userinfo = &Userinfo{
			Username:    v.User.Username(),
			Password:    password,
			PasswordSet: hasPassword,
		}
	}

	return lib.typeAdapter.NativeToValue(&URL{
		Scheme:     v.Scheme,
		Opaque:     v.Opaque,
		User:       userinfo,
		Host:       v.Host,
		Path:       v.Path,
		RawPath:    v.RawPath,
		ForceQuery: v.ForceQuery,
		RawQuery:   v.RawQuery,
		Fragment:   v.Fragment,
	})
}

func (lib *URLLibrary) toUserinfoValue(username, password string, passwordSet bool) ref.Val {
	return lib.typeAdapter.NativeToValue(&Userinfo{Username: username, Password: password, PasswordSet: passwordSet})
}

func (lib *URLLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{}

	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createURLName("joinPath"),
			OverloadFunc(createURLID("joinPath_string_strings_url"), []*cel.Type{cel.StringType, cel.ListType(cel.StringType)}, URLType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					base := string(args[0].(types.String))
					elems := args[1].(traits.Lister)
					var paths []string
					for i := types.Int(0); i < elems.Size().(types.Int); i++ {
						pathElem := elems.Get(i)
						paths = append(paths, string(pathElem.(types.String)))
					}

					result, err := url.JoinPath(base, paths...)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(result)
				},
			),
		),
		BindFunction(
			createURLName("pathEscape"),
			OverloadFunc(createURLID("pathEscape_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(url.PathEscape(string(args[0].(types.String))))
				},
			),
		),
		BindFunction(
			createURLName("pathUnescape"),
			OverloadFunc(createURLID("pathUnescape_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					result, err := url.PathUnescape(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(result)
				},
			),
		),
		BindFunction(
			createURLName("queryEscape"),
			OverloadFunc(createURLID("queryEscape_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.String(url.QueryEscape(string(args[0].(types.String))))
				},
			),
		),
		BindFunction(
			createURLName("queryUnescape"),
			OverloadFunc(createURLID("queryUnescape_string_string"), []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					result, err := url.QueryUnescape(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(result)
				},
			),
		),

		// URL functions
		BindFunction(
			createURLName("parse"),
			OverloadFunc(createURLID("parse_string_url"), []*cel.Type{cel.StringType}, URLType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					v, err := url.Parse(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toURLValue(*v)
				},
			),
		),
		BindFunction(
			createURLName("parseRequestURI"),
			OverloadFunc(createURLID("parseRequestURI_string_url"), []*cel.Type{cel.StringType}, URLType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					v, err := url.ParseRequestURI(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toURLValue(*v)
				},
			),
		),
		BindMemberFunction(
			"scheme",
			MemberOverloadFunc(createURLID("scheme_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoURLValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Scheme)
				},
			),
		),
		BindMemberFunction(
			"opaque",
			MemberOverloadFunc(createURLID("opaque_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Opaque)
				},
			),
		),
		BindMemberFunction(
			"user",
			MemberOverloadFunc(createURLID("user_url_userinfo"), URLType, []*cel.Type{}, UserinfoType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					password, hasPassword := v.User.Password()
					return lib.toUserinfoValue(v.User.Username(), password, hasPassword)
				},
			),
		),
		BindMemberFunction(
			"host",
			MemberOverloadFunc(createURLID("host_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Host)
				},
			),
		),
		BindMemberFunction(
			"path",
			MemberOverloadFunc(createURLID("path_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Path)
				},
			),
		),
		BindMemberFunction(
			"rawPath",
			MemberOverloadFunc(createURLID("rawPath_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RawPath)
				},
			),
		),
		// forceQuery is changed to return String
		BindMemberFunction(
			"forceQuery",
			MemberOverloadFunc(createURLID("forceQuery_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					v.ForceQuery = true
					return types.String(v.String())
				},
			),
		),
		BindMemberFunction(
			"rawQuery",
			MemberOverloadFunc(createURLID("rawQuery_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RawQuery)
				},
			),
		),
		BindMemberFunction(
			"fragment",
			MemberOverloadFunc(createURLID("fragment_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Fragment)
				},
			),
		),
		BindMemberFunction(
			"rawFragment",
			MemberOverloadFunc(createURLID("rawFragment_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RawFragment)
				},
			),
		),
		BindMemberFunction(
			"escapedFragment",
			MemberOverloadFunc(createURLID("escapedFragment_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.EscapedFragment())
				},
			),
		),
		BindMemberFunction(
			"escapedPath",
			MemberOverloadFunc(createURLID("escapedPath_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.EscapedPath())
				},
			),
		),
		BindMemberFunction(
			"hostname",
			MemberOverloadFunc(createURLID("hostname_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Hostname())
				},
			),
		),
		BindMemberFunction(
			"isAbs",
			MemberOverloadFunc(createURLID("isAbs_url_bool"), URLType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.IsAbs())
				},
			),
		),
		BindMemberFunction(
			"joinPath",
			MemberOverloadFunc(createURLID("joinPath_url_strings_url"), URLType, []*cel.Type{cel.ListType(cel.StringType)}, URLType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}

					elems := args[0].(traits.Lister)
					var paths []string
					for i := types.Int(0); i < elems.Size().(types.Int); i++ {
						pathElem := elems.Get(i)
						paths = append(paths, string(pathElem.(types.String)))
					}

					v.Path, err = url.JoinPath(v.Path, paths...)

					return lib.toURLValue(v)
				},
			),
		),
		BindMemberFunction(
			"marshalBinary",
			MemberOverloadFunc(createURLID("MarshalBinary_url_bytes"), URLType, []*cel.Type{}, cel.BytesType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}

					b, err := v.MarshalBinary()
					if err != nil {
						return types.NewErr(err.Error())
					}

					return types.Bytes(b)
				},
			),
		),
		BindMemberFunction(
			"parse",
			MemberOverloadFunc(createURLID("parse_url_string_url"), URLType, []*cel.Type{cel.StringType}, URLType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}

					u, err := v.Parse(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}

					return lib.toURLValue(*u)
				},
			),
		),
		BindMemberFunction(
			"port",
			MemberOverloadFunc(createURLID("port_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Port())
				},
			),
		),
		BindMemberFunction(
			"query",
			MemberOverloadFunc(createURLID("query_url_map"), URLType, []*cel.Type{}, cel.MapType(cel.StringType, cel.ListType(cel.StringType)),
				func(ctx context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}

					adapter := types.DefaultTypeAdapter
					queryParams := v.Query()
					queryMap := map[ref.Val]ref.Val{}
					for key, values := range queryParams {
						var valueList []ref.Val
						for _, val := range values {
							valueList = append(valueList, types.String(val))
						}
						queryMap[types.String(key)] = types.NewStringList(adapter, values)
					}

					return types.NewRefValMap(adapter, queryMap)
				},
			),
		),
		BindMemberFunction(
			"requestURI",
			MemberOverloadFunc(createURLID("requestURI_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RequestURI())
				},
			),
		),
		BindMemberFunction(
			"resolveReference",
			MemberOverloadFunc(createURLID("resolveReference_url_url_url"), URLType, []*cel.Type{URLType}, URLType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}

					r, err := args[0].Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					u := v.ResolveReference(&r)

					return lib.toURLValue(*u)
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createURLID("string_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.String())
				},
			),
		),
		BindMemberFunction(
			"unmarshalBinary",
			MemberOverloadFunc(createURLID("unmarshalBinary_url_bytes_url"), URLType, []*cel.Type{cel.BytesType}, URLType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					var u url.URL
					err := u.UnmarshalBinary(args[0].(types.Bytes))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toURLValue(u)
				},
			),
		),

		// Userinfo functions
		BindFunction(
			createURLName("user"),
			OverloadFunc(createURLID("user_string_userinfo"), []*cel.Type{cel.StringType}, UserinfoType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toUserinfoValue(string(args[0].(types.String)), "", false)
				},
			),
		),
		BindFunction(
			createURLName("userPassword"),
			OverloadFunc(createURLID("userPassword_string_string_userinfo"), []*cel.Type{cel.StringType, cel.StringType}, UserinfoType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toUserinfoValue(string(args[0].(types.String)), string(args[1].(types.String)), true)
				},
			),
		),
		BindMemberFunction(
			"username",
			MemberOverloadFunc(createURLID("username_userinfo_string"), UserinfoType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*Userinfo).GoUserinfo()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Username())
				},
			),
		),
		BindMemberFunction(
			"password",
			MemberOverloadFunc(createURLID("password_userinfo_string"), UserinfoType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*Userinfo).GoUserinfo()
					if err != nil {
						return types.NewErr(err.Error())
					}
					password, hasPassword := v.Password()
					if hasPassword {
						return types.String(password)
					} else {
						return types.String("")
					}
				},
			),
		),
		BindMemberFunction(
			"passwordSet",
			MemberOverloadFunc(createURLID("passwordSet_userinfo_bool"), UserinfoType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*Userinfo).GoUserinfo()
					if err != nil {
						return types.NewErr(err.Error())
					}
					_, hasPassword := v.Password()
					return types.Bool(hasPassword)
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createURLID("string_userinfo_string"), UserinfoType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*Userinfo).GoUserinfo()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.String())
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}

	return opts
}

func (lib *URLLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
