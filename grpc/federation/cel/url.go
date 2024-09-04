package cel

import (
	"context"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"net/url"
)

const URLPackageName = "url"

var (
	URLType      = cel.ObjectType("grpc.federation.url.URL")
	UserinfoType = cel.ObjectType("grpc.federation.url.Userinfo")
)

func (x *URL) GoURL() (url.URL, error) {
	user, err := x.GetUser().GoUserinfo()
	if err != nil {
		return url.URL{}, err
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
	userinfo := &Userinfo{}
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
		BindMemberFunction(
			"scheme",
			MemberOverloadFunc(createTimeID("scheme_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("opaque_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("user_url_userinfo"), URLType, []*cel.Type{}, UserinfoType,
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
			MemberOverloadFunc(createTimeID("host_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("path_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("rawPath_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RawPath)
				},
			),
		),
		BindMemberFunction(
			"forceQuery",
			MemberOverloadFunc(createTimeID("forceQuery_url_bool"), URLType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.ForceQuery)
				},
			),
		),
		BindMemberFunction(
			"rawQuery",
			MemberOverloadFunc(createTimeID("rawQuery_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("fragment_url_string"), URLType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("rawFragment_url_string"), URLType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := self.Value().(*URL).GoURL()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.RawFragment)
				},
			),
		),

		// Userinfo functions
		BindMemberFunction(
			"username",
			MemberOverloadFunc(createTimeID("username_userinfo_string"), UserinfoType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("password_userinfo_string"), UserinfoType, []*cel.Type{}, cel.StringType,
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
			MemberOverloadFunc(createTimeID("passwordSet_userinfo_bool"), UserinfoType, []*cel.Type{}, cel.BoolType,
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
	} {
		opts = append(opts, funcOpts...)
	}

	return opts
}

func (lib *URLLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
