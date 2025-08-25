package cel

import (
	"context"
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const AnyPackageName = "any"

type Any struct {
	*anypb.Any
}

func (any *Any) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return any.Any, nil
}

func (any *Any) ConvertToType(typeValue ref.Type) ref.Val {
	return nil
}

func (any *Any) Equal(other ref.Val) ref.Val {
	return other
}

func (any *Any) Type() ref.Type {
	return cel.AnyType
}

func (any *Any) Value() any {
	return any.Any
}

func NewAnyLibrary(typeAdapter types.Adapter) *AnyLibrary {
	return &AnyLibrary{
		typeAdapter: typeAdapter,
	}
}

type AnyLibrary struct {
	typeAdapter types.Adapter
}

func (lib *AnyLibrary) LibraryName() string {
	return packageName(AnyPackageName)
}

func createAny(name string) string {
	return createName(AnyPackageName, name)
}

func createAnyID(name string) string {
	return createID(AnyPackageName, name)
}

func (lib *AnyLibrary) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createAny("new"),
			OverloadFunc(createAnyID("new_dyn_any"),
				[]*cel.Type{cel.DynType}, cel.AnyType,
				func(ctx context.Context, args ...ref.Val) ref.Val {
					var msg proto.Message
					switch v := args[0].Value().(type) {
					case proto.Message:
						msg = v
					case time.Time:
						// google.protobuf.Timestamppb
						msg = timestamppb.New(v)
					default:
						return types.NewErr("specified type as an argument must be a Protocol Buffers message type")
					}
					anymsg, err := anypb.New(msg)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &Any{Any: anymsg}
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *AnyLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
