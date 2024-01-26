package cel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"unsafe"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	ErrOutOfRangeAccessWasmMemory = errors.New(
		`grpc-federation: detected out of range access`,
	)
	ErrWasmContentMismatch = errors.New(
		`grpc-federation: wasm file content mismatch`,
	)
)

type WasmPlugin struct {
	File   string
	r      wazero.Runtime
	mod    api.Module
	malloc api.Function
	free   api.Function
}

func NewWasmPlugin(ctx context.Context, wasmCfg WasmConfig) (*WasmPlugin, error) {
	wasmFile, err := os.ReadFile(wasmCfg.Path)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(wasmFile)
	gotHash := hex.EncodeToString(hash[:])
	if wasmCfg.Sha256 != gotHash {
		return nil, fmt.Errorf(`expected [%s] but got [%s]: %w`, wasmCfg.Sha256, gotHash, ErrWasmContentMismatch)
	}
	cfg := wazero.NewRuntimeConfig()
	r := wazero.NewRuntimeWithConfig(ctx, cfg)

	if _, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(wasmDebugLog).Export("grpc_federation_log").
		Instantiate(ctx); err != nil {
		return nil, err
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	mod, err := r.Instantiate(ctx, wasmFile)
	if err != nil {
		return nil, err
	}
	return &WasmPlugin{
		File:   wasmCfg.Path,
		r:      r,
		mod:    mod,
		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),
	}, nil
}

func (p *WasmPlugin) Call(ctx context.Context, fn *CELFunction, md []byte, args ...ref.Val) ref.Val {
	f := p.mod.ExportedFunction(fn.ID)
	if f == nil {
		return types.NewErr(fmt.Sprintf("grpc-federation: failed to find exported function %s in %s", fn.Name, p.File))
	}

	ptr, size, err := p.stringToPtr(ctx, string(md))
	if err != nil {
		return types.NewErr(fmt.Sprintf("grpc-federation: failed to encode metadata: %s", err.Error()))
	}
	wasmArgs := []uint64{api.EncodeU32(ptr), api.EncodeU32(size)}
	for idx, arg := range args {
		wasmArg, err := p.refToWASMType(ctx, fn.Args[idx], arg)
		if err != nil {
			return types.NewErr(err.Error())
		}
		wasmArgs = append(wasmArgs, wasmArg...)
	}
	result, err := f.Call(ctx, wasmArgs...)
	if err != nil {
		return types.NewErr(err.Error())
	}
	ret := ReturnValue(result[0])
	if err := p.returnValueToError(ret); err != nil {
		return types.NewErr(err.Error())
	}
	return p.returnValueToCELValue(fn, fn.Return, ret)
}

func (p *WasmPlugin) refToWASMType(ctx context.Context, typ *cel.Type, v ref.Val) ([]uint64, error) {
	switch typ.Kind() {
	case types.BoolKind:
		vv := v.(types.Bool)
		if vv {
			return []uint64{uint64(1)}, nil
		}
		return []uint64{uint64(0)}, nil
	case types.DoubleKind:
		vv := v.(types.Double)
		return []uint64{api.EncodeF64(float64(vv))}, nil
	case types.IntKind:
		vv := v.(types.Int)
		return []uint64{api.EncodeI64(int64(vv))}, nil
	case types.BytesKind:
		vv := v.(types.Bytes)
		ptr, size, err := p.stringToPtr(ctx, string(vv))
		if err != nil {
			return nil, err
		}
		return []uint64{api.EncodeU32(ptr), api.EncodeU32(size)}, nil
	case types.StringKind:
		vv := v.(types.String)
		ptr, size, err := p.stringToPtr(ctx, string(vv))
		if err != nil {
			return nil, err
		}
		return []uint64{api.EncodeU32(ptr), api.EncodeU32(size)}, nil
	case types.UintKind:
		vv := v.(types.Uint)
		return []uint64{uint64(vv)}, nil
	case types.StructKind:
		switch obj := v.Value().(type) {
		case *objectValueWrapper:
			return []uint64{obj.wasmValue()}, nil
		default:
			return nil, fmt.Errorf(
				`grpc-federation: currently unsupported native proto message "%T"`,
				obj,
			)
		}
	}
	return nil, fmt.Errorf(`grpc-federation: found unexpected cel function's argument type "%s"`, typ)
}

type objectValueWrapper struct {
	typ *cel.Type
	ptr unsafe.Pointer
}

func (w *objectValueWrapper) wasmValue() uint64 {
	return uint64(uintptr(w.ptr))
}

func (w *objectValueWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) { return w, nil }

func (w *objectValueWrapper) ConvertToType(typeValue ref.Type) ref.Val { return w }

func (w *objectValueWrapper) Equal(other ref.Val) ref.Val { return nil }

func (w *objectValueWrapper) Type() ref.Type { return w.typ }

func (w *objectValueWrapper) Value() any { return w }

func (p *WasmPlugin) returnValueToCELValue(fn *CELFunction, typ *cel.Type, v ReturnValue) ref.Val {
	switch typ.Kind() {
	case types.BoolKind:
		return types.Bool(p.returnValueToBool(v))
	case types.BytesKind:
		return types.Bytes(p.returnValueToBytes(v))
	case types.DoubleKind:
		return types.Double(p.returnValueToFloat64(v))
	case types.ErrorKind:
		return types.NewErr(p.returnValueToString(v))
	case types.IntKind:
		return types.Int(p.returnValueToInt64(v))
	case types.StringKind:
		return types.String(p.returnValueToString(v))
	case types.UintKind:
		return types.Uint(p.returnValueToUint64(v))
	case types.StructKind:
		return &objectValueWrapper{
			typ: typ,
			ptr: p.returnValueToPtr(v),
		}
	}
	return types.NewErr("grpc-federation: unknown result type %s from %s function", typ, fn.Name)
}

type ReturnValue uint64

func (p *WasmPlugin) returnValueToPtr(v ReturnValue) unsafe.Pointer {
	return unsafe.Pointer(uintptr(v >> 32))
}

func (p *WasmPlugin) returnValueToBool(v ReturnValue) bool {
	return uint32(v>>32) == 1
}

func (p *WasmPlugin) returnValueToInt64(v ReturnValue) int64 {
	return int64(v >> 32)
}

func (p *WasmPlugin) returnValueToUint64(v ReturnValue) uint64 {
	return uint64(v) >> 32
}

func (p *WasmPlugin) returnValueToFloat64(v ReturnValue) float64 {
	return api.DecodeF64(uint64(v) >> 32)
}

func (p *WasmPlugin) returnValueToString(v ReturnValue) string {
	ptr := uint32(v >> 32)
	size := uint32(v)
	return p.ptrToString(ptr, size)
}

func (p *WasmPlugin) returnValueToBytes(v ReturnValue) []byte {
	return []byte(p.returnValueToString(v))
}

func (p *WasmPlugin) returnValueToError(v ReturnValue) error {
	ptr := uint32(v >> 32)
	size := uint32(v)

	if (size & (1 << 31)) <= 0 {
		return nil
	}

	size &^= (1 << 31)
	bytes, ok := p.mod.Memory().Read(ptr, size)
	if !ok {
		return fmt.Errorf(
			`failed to read wasm memory: (ptr, size) = (%d, %d) and memory size is %d: %w`,
			ptr, size, p.mod.Memory().Size(),
			ErrOutOfRangeAccessWasmMemory,
		)
	}
	return errors.New(string(bytes))
}

func (p *WasmPlugin) stringToPtr(ctx context.Context, s string) (uint32, uint32, error) {
	results, err := p.malloc.Call(ctx, uint64(len(s)))
	if err != nil {
		return 0, 0, err
	}
	ptr := uint32(results[0])
	size := uint32(len(s))
	if !p.mod.Memory().Write(ptr, []byte(s)) {
		return 0, 0, fmt.Errorf(
			`failed to write wasm memory: (ptr, size) = (%d, %d) and memory size is %d: %w`,
			ptr, size, p.mod.Memory().Size(),
			ErrOutOfRangeAccessWasmMemory,
		)
	}
	return ptr, size, nil
}

func (p *WasmPlugin) ptrToString(ptr, size uint32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func wasmDebugLog(_ context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		log.Panicf(
			`grpc-federation: failed to output debug log: detected out of range access. (ptr, size) = (%d, %d)`,
			offset, byteCount,
		)
	}
	fmt.Fprintln(os.Stdout, string(buf))
}
