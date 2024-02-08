//go:build tinygo.wasm

package cel

import (
	"context"

	"github.com/google/cel-go/common/types/ref"
)

type WasmPlugin struct {
	File string
}

func NewWasmPlugin(ctx context.Context, wasmCfg WasmConfig) (*WasmPlugin, error) {
	return nil, nil
}

func (p *WasmPlugin) Call(ctx context.Context, fn *CELFunction, md []byte, args ...ref.Val) ref.Val {
	return nil
}
