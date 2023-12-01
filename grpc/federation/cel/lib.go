package cel

import "github.com/google/cel-go/cel"

type Library struct {
	subLibs []cel.SingletonLibrary
}

func NewLibrary() *Library {
	return &Library{
		subLibs: []cel.SingletonLibrary{
			new(TimeLibrary),
		},
	}
}

func (lib *Library) LibraryName() string {
	return "grpc.federation"
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
