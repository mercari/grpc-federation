//go:build wasip1

package federation

import "unsafe"

//go:wasmimport grpcfederation grpc_federation_write
func grpc_federation_write(ptr, size uint32)

func WritePluginContent(content []byte) {
	grpc_federation_write(
		uint32(uintptr(unsafe.Pointer(&content[0]))),
		uint32(len(content)),
	)
}
