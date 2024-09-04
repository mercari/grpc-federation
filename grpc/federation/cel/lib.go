package cel

import (
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

type Library struct {
	name    string
	subLibs []cel.SingletonLibrary
}

func NewLibrary(typeAdapter types.Adapter) *Library {
	mdLib := NewMetadataLibrary()
	logLib := NewLogLibrary()
	return &Library{
		name: "grpc.federation.static",
		subLibs: []cel.SingletonLibrary{
			NewURLLibrary(typeAdapter),
			NewStringsLibrary(),
			NewTimeLibrary(typeAdapter),
			NewListLibrary(typeAdapter),
			new(RandLibrary),
			new(UUIDLibrary),
			new(EnumLibrary),
			mdLib,
			logLib,
			new(CastLibrary),
		},
	}
}

func NewASTValidators() []cel.ASTValidator {
	return []cel.ASTValidator{
		NewListValidator(),
		NewEnumValidator(),
	}
}

func IsStandardLibraryType(typeName string) bool {
	return strings.HasPrefix(strings.TrimPrefix(typeName, "."), "grpc.federation.")
}

func (lib *Library) LibraryName() string {
	return lib.name
}

func (lib *Library) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, sub := range lib.subLibs {
		opts = append(opts, sub.CompileOptions()...)
	}
	return opts
}

func (lib *Library) ProgramOptions() []cel.ProgramOption {
	var opts []cel.ProgramOption
	for _, sub := range lib.subLibs {
		opts = append(opts, sub.ProgramOptions()...)
	}
	return opts
}

func packageName(subpkg string) string {
	return "grpc.federation." + subpkg
}

func packageNameID(subpkg string) string {
	return "grpc_federation_" + subpkg
}

func createName(subpkg, name string) string {
	return packageName(subpkg) + "." + name
}

func createID(subpkg, name string) string {
	return packageNameID(subpkg) + "_" + name
}
