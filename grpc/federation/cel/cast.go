package cel

import (
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const CastPackageName = "cast"

const (
	CastNullValueFunc = "grpc.federation.cast.null_value"
)

type CastLibrary struct{}

func (lib *CastLibrary) LibraryName() string {
	return CastPackageName
}

func (lib *CastLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.Function(CastNullValueFunc,
			cel.Overload("grpc_federation_cast_null_value",
				[]*cel.Type{cel.DynType}, cel.DynType,
				cel.UnaryBinding(lib.nullValue),
			),
		),
	}
	return opts
}

func (lib *CastLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (lib *CastLibrary) nullValue(val ref.Val) ref.Val {
	rv := reflect.ValueOf(val.Value())
	if val.Value() == nil || val.Equal(types.NullValue) == types.True || rv.Kind() == reflect.Ptr && rv.IsNil() {
		return types.NullValue
	}
	return val
}
