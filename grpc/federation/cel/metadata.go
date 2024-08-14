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

type MetadataLibrary struct{}

func (lib *MetadataLibrary) LibraryName() string {
	return packageName(MetadataPackageName)
}

func createMetadata(name string) string {
	return createName(MetadataPackageName, name)
}

func createMetadataID(name string) string {
	return createID(MetadataPackageName, name)
}

func (lib *MetadataLibrary) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createMetadata("incoming"),
			OverloadFunc(createMetadataID("incoming_map_string_list_string"),
				[]*cel.Type{}, types.NewMapType(types.StringType, types.NewListType(types.StringType)),
				func(ctx context.Context, _ ...ref.Val) ref.Val {
					refMap := make(map[ref.Val]ref.Val)
					md, ok := metadata.FromIncomingContext(ctx)
					if !ok {
						return types.NewRefValMap(types.DefaultTypeAdapter, refMap)
					}

					for k, vs := range md {
						refMap[types.String(k)] = types.NewStringList(types.DefaultTypeAdapter, vs)
					}
					return types.NewRefValMap(types.DefaultTypeAdapter, refMap)
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *MetadataLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
