package cel

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/grpc/metadata"
)

const MetadataPackageName = "metadata"

func NewMetadataLibrary() *MetadataLibrary {
	return &MetadataLibrary{}
}

type MetadataLibrary struct {
	ctx context.Context
}

func (lib *MetadataLibrary) LibraryName() string {
	return packageName(MetadataPackageName)
}

func (lib *MetadataLibrary) Initialize(ctx context.Context) {
	lib.ctx = ctx
}

func createMetadata(name string) string {
	return createName(MetadataPackageName, name)
}

func createMetadataID(name string) string {
	return createID(MetadataPackageName, name)
}

func (lib *MetadataLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.Function(
			createMetadata("incoming"),
			cel.Overload(createMetadataID("incoming_map_string_list_string"), []*cel.Type{}, types.NewMapType(types.StringType, types.NewListType(types.StringType)),
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					refMap := make(map[ref.Val]ref.Val)
					md, ok := metadata.FromIncomingContext(lib.ctx)
					if !ok {
						return types.NewRefValMap(types.DefaultTypeAdapter, refMap)
					}

					for k, vs := range md {
						refMap[types.String(k)] = types.NewStringList(types.DefaultTypeAdapter, vs)
					}
					return types.NewRefValMap(types.DefaultTypeAdapter, refMap)
				}),
			),
		),
	}
	return opts
}

func (lib *MetadataLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
