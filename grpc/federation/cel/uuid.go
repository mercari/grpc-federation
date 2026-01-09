package cel

import (
	"context"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/uuid"
)

const UUIDPackageName = "uuid"

var (
	UUIDType = types.NewObjectType(createUUIDName("UUID"))
)

type UUID struct {
	uuid.UUID
}

func (u *UUID) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return u, nil
}

func (u *UUID) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErrFromString("grpc.federation.uuid: uuid type conversion does not support")
}

func (u *UUID) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*UUID); ok {
		return types.Bool(u.String() == o.String())
	}
	return types.False
}

func (u *UUID) Type() ref.Type {
	return UUIDType
}

func (u *UUID) Value() any {
	return u
}

type UUIDLibrary struct {
}

func (lib *UUIDLibrary) LibraryName() string {
	return packageName(UUIDPackageName)
}

func createUUIDName(name string) string {
	return createName(UUIDPackageName, name)
}

func createUUIDID(name string) string {
	return createID(UUIDPackageName, name)
}

func (lib *UUIDLibrary) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		// UUID functions
		BindFunction(
			createUUIDName("new"),
			OverloadFunc(createUUIDID("new_uuid"), []*cel.Type{}, UUIDType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return &UUID{
						UUID: uuid.New(),
					}
				},
			),
		),
		BindFunction(
			createUUIDName("newRandom"),
			OverloadFunc(createUUIDID("new_random_uuid"), []*cel.Type{}, UUIDType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					id, err := uuid.NewRandom()
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return &UUID{UUID: id}
				},
			),
		),
		BindFunction(
			createUUIDName("newRandomFromRand"),
			OverloadFunc(createUUIDID("new_random_from_rand_uuid"), []*cel.Type{RandType}, UUIDType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					id, err := uuid.NewRandomFromReader(args[0].(*Rand))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return &UUID{UUID: id}
				},
			),
		),
		BindFunction(
			createUUIDName("parse"),
			OverloadFunc(createUUIDID("parse"), []*cel.Type{cel.StringType}, UUIDType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					id, err := uuid.Parse(string(args[0].(types.String)))
					if err != nil {
						return types.NewErrFromString(err.Error())
					}
					return &UUID{UUID: id}
				},
			),
		),
		BindFunction(
			createUUIDName("validate"),
			OverloadFunc(createUUIDID("validate"), []*cel.Type{cel.StringType}, cel.BoolType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Bool(uuid.Validate(string(args[0].(types.String))) == nil)
				},
			),
		),
		BindMemberFunction(
			"domain",
			MemberOverloadFunc(createRandID("domain_uuid_string"), UUIDType, []*cel.Type{}, cel.UintType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Uint(self.(*UUID).Domain())
				},
			),
		),
		BindMemberFunction(
			"id",
			MemberOverloadFunc(createRandID("id_uuid_string"), UUIDType, []*cel.Type{}, cel.UintType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Uint(self.(*UUID).ID())
				},
			),
		),
		BindMemberFunction(
			"time",
			MemberOverloadFunc(createRandID("time_uuid_int"), UUIDType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*UUID).Time())
				},
			),
		),
		BindMemberFunction(
			"urn",
			MemberOverloadFunc(createRandID("urn_uuid_string"), UUIDType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(*UUID).URN())
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createRandID("string_uuid_string"), UUIDType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(*UUID).String())
				},
			),
		),
		BindMemberFunction(
			"version",
			MemberOverloadFunc(createRandID("version_uuid_uint"), UUIDType, []*cel.Type{}, cel.UintType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Uint(self.(*UUID).Version())
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *UUIDLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
