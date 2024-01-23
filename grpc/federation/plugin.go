//go:build tinygo.wasm

package federation

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

// #include <stdlib.h>
import "C"

type ReturnValue uint64

func MessageToReturnValue[T any](v *T) ReturnValue {
	return ReturnValue(uint64(uintptr(unsafe.Pointer(v))) << uint64(32))
}

func StringToReturnValue(v string) ReturnValue {
	return BytesToReturnValue([]byte(v))
}

func BytesToReturnValue(v []byte) ReturnValue {
	ptr, size := BytesToPtrWithAllocation(v)
	return ReturnValue((uint64(ptr) << uint64(32)) | uint64(size))
}

func BoolToReturnValue(v bool) ReturnValue {
	if v {
		return ReturnValue(1 << uint64(32))
	}
	return ReturnValue(0 << uint64(32))
}

func Int32ToReturnValue(v int32) ReturnValue {
	return ReturnValue(api.EncodeI32(v) << uint64(32))
}

func Int64ToReturnValue(v int64) ReturnValue {
	return ReturnValue(api.EncodeI64(v) << uint64(32))
}

func Uint32ToReturnValue(v uint32) ReturnValue {
	return ReturnValue(api.EncodeU32(v) << uint64(32))
}

func Uint64ToReturnValue(v uint64) ReturnValue {
	return ReturnValue(v << uint64(32))
}

func Float32ToReturnValue(v float32) ReturnValue {
	return ReturnValue(api.EncodeF32(v) << uint64(32))
}

func Float64ToReturnValue(v float64) ReturnValue {
	return ReturnValue(api.EncodeF64(v) << uint64(32))
}

func ErrorToReturnValue(err error) ReturnValue {
	ptr, size := BytesToPtrWithAllocation([]byte(err.Error()))
	return ReturnValue((uint64(ptr) << uint64(32)) | uint64(size) | (1 << 31))
}

func StringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

func BytesToPtr(b []byte) (uint32, uint32) {
	return StringToPtr(string(b))
}

func StringToPtrWithAllocation(s string) (uint32, uint32) {
	return BytesToPtrWithAllocation([]byte(s))
}

func BytesToPtrWithAllocation(b []byte) (uint32, uint32) {
	if len(b) == 0 {
		return 0, 0
	}

	size := C.ulong(len(b))
	ptr := unsafe.Pointer(C.malloc(size))

	copy(unsafe.Slice((*byte)(ptr), size), b)

	return uint32(uintptr(ptr)), uint32(len(b))
}

func ToFloat64(v uint64) float64 {
	return api.DecodeF64(v)
}

func ToFloat32(v uint32) float32 {
	return api.DecodeF32(uint64(v))
}

func ToString(ptr, size uint32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func ToMessage[T any](ptr uint32) *T {
	return (*T)(unsafe.Pointer(uintptr(ptr)))
}

func ToBytes(ptr, size uint32) []byte {
	var b []byte
	s := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	s.Len = uintptr(size)
	s.Cap = uintptr(size)
	s.Data = uintptr(ptr)
	return b
}

func FreePtr(ptr uint32) {
	C.free(unsafe.Pointer(uintptr(ptr)))
}

func DebugLog(msg string) {
	ptr, size := StringToPtr(msg)
	grpc_federation_log(ptr, size)
	runtime.KeepAlive(msg)
}

//go:wasm-module env
//export grpc_federation_log
func grpc_federation_log(ptr, size uint32)
