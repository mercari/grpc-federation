package cel

import (
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
	return types.NewErr("grpc.federation.uuid: uuid type conversion does not support")
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
	opts := []cel.EnvOption{
		// UUID functions
		cel.Function(
			createUUIDName("new"),
			cel.Overload(createUUIDID("new_uuid"), []*cel.Type{}, UUIDType,
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					return &UUID{
						UUID: uuid.New(),
					}
				}),
			),
		),
		cel.Function(
			createUUIDName("newRandom"),
			cel.Overload(createUUIDID("new_random_uuid"), []*cel.Type{}, UUIDType,
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					id, err := uuid.NewRandom()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &UUID{UUID: id}
				}),
			),
		),
		cel.Function(
			createUUIDName("newRandomFromRand"),
			cel.Overload(createUUIDID("new_random_from_rand_uuid"), []*cel.Type{RandType}, UUIDType,
				cel.UnaryBinding(func(rand ref.Val) ref.Val {
					id, err := uuid.NewRandomFromReader(rand.(*Rand))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &UUID{UUID: id}
				}),
			),
		),
		cel.Function(
			"domain",
			cel.MemberOverload(createRandID("domain_uuid_string"), []*cel.Type{UUIDType}, cel.UintType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Uint(self.(*UUID).Domain())
				}),
			),
		),
		cel.Function(
			"id",
			cel.MemberOverload(createRandID("id_uuid_string"), []*cel.Type{UUIDType}, cel.UintType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Uint(self.(*UUID).ID())
				}),
			),
		),
		cel.Function(
			"time",
			cel.MemberOverload(createRandID("time_uuid_int"), []*cel.Type{UUIDType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(*UUID).Time())
				}),
			),
		),
		cel.Function(
			"urn",
			cel.MemberOverload(createRandID("urn_uuid_string"), []*cel.Type{UUIDType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.String(self.(*UUID).URN())
				}),
			),
		),
		cel.Function(
			"string",
			cel.MemberOverload(createRandID("string_uuid_string"), []*cel.Type{UUIDType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.String(self.(*UUID).String())
				}),
			),
		),
		cel.Function(
			"version",
			cel.MemberOverload(createRandID("version_uuid_uint"), []*cel.Type{UUIDType}, cel.UintType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Uint(self.(*UUID).Version())
				}),
			),
		),
	}
	return opts
}

func (lib *UUIDLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
