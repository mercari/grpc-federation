package cel

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

type CELPlugin struct {
	Name      string
	Functions []*CELFunction
	wasm      *WasmPlugin
}

type CELFunction struct {
	Name     string
	ID       string
	Args     []*cel.Type
	Return   *cel.Type
	IsMethod bool
}

type CELPluginConfig struct {
	Name      string
	Wasm      WasmConfig
	Functions []*CELFunction
}

type WasmConfig struct {
	Path   string
	Sha256 string
}

func NewCELPlugin(ctx context.Context, cfg CELPluginConfig) (*CELPlugin, error) {
	wasm, err := NewWasmPlugin(ctx, cfg.Wasm)
	if err != nil {
		return nil, err
	}
	return &CELPlugin{
		Name:      cfg.Name,
		Functions: cfg.Functions,
		wasm:      wasm,
	}, nil
}

func (p *CELPlugin) LibraryName() string {
	return p.Name
}

func (p *CELPlugin) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, fn := range p.Functions {
		fn := fn
		bindFunc := cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			return p.wasm.Call(fn, args...)
		})
		var overload cel.FunctionOpt
		if fn.IsMethod {
			overload = cel.MemberOverload(fn.ID, fn.Args, fn.Return, bindFunc)
		} else {
			overload = cel.Overload(fn.ID, fn.Args, fn.Return, bindFunc)
		}
		opts = append(opts, cel.Function(fn.Name, overload))
	}
	return opts
}

func (p *CELPlugin) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
