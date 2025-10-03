//go:build wasip1

package federation

import (
	"unsafe"
)

//go:wasmimport grpcfederation write
func grpc_federation_write(ptr, size uint32)

//go:wasmimport grpcfederation read_length
func grpc_federation_read_length() uint32

//go:wasmimport grpcfederation read
func grpc_federation_read(uint32)

func WritePluginContent(content []byte) {
	grpc_federation_write(
		uint32(uintptr(unsafe.Pointer(&content[0]))),
		uint32(len(content)),
	)
}

func ReadPluginContent() string {
	length := grpc_federation_read_length()
	if length == 0 {
		return ""
	}
	buf := make([]byte, length)
	grpc_federation_read(
		uint32(uintptr(unsafe.Pointer(&buf[0]))),
	)
	return string(buf)
}
