//go:build !tinygo.wasm

package federation

type ReturnValue uint64

// These are dummy APIs to compile plugin SDK at host context.
func MessageToReturnValue[T any](v *T) ReturnValue        { return 0 }
func StringToReturnValue(v string) ReturnValue            { return 0 }
func BytesToReturnValue(v []byte) ReturnValue             { return 0 }
func BoolToReturnValue(v bool) ReturnValue                { return 0 }
func Int32ToReturnValue(v int32) ReturnValue              { return 0 }
func Int64ToReturnValue(v int64) ReturnValue              { return 0 }
func Uint32ToReturnValue(v uint32) ReturnValue            { return 0 }
func Uint64ToReturnValue(v uint64) ReturnValue            { return 0 }
func Float32ToReturnValue(v float32) ReturnValue          { return 0 }
func Float64ToReturnValue(v float64) ReturnValue          { return 0 }
func ErrorToReturnValue(err error) ReturnValue            { return 0 }
func StringToPtr(s string) (uint32, uint32)               { return 0, 0 }
func BytesToPtr(b []byte) (uint32, uint32)                { return 0, 0 }
func StringToPtrWithAllocation(s string) (uint32, uint32) { return 0, 0 }
func BytesToPtrWithAllocation(b []byte) (uint32, uint32)  { return 0, 0 }
func ToFloat64(v uint64) float64                          { return 0 }
func ToFloat32(v uint32) float32                          { return 0 }
func ToString(ptr, size uint32) string                    { return "" }
func ToMessage[T any](ptr uint32) *T                      { return nil }
func ToBytes(ptr, size uint32) []byte                     { return nil }
func FreePtr(ptr uint32)                                  {}
func DebugLog(msg string)                                 {}
