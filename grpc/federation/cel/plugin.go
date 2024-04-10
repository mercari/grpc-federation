package cel

import (
	"context"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/grpc/metadata"
)

type CELPlugin struct {
	Name      string
	Functions []*CELFunction
	wasm      *WasmPlugin
	ctx       context.Context
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

func NewCELPlugin(ctx context.Context, cfg CELPluginConfig, celRegistry *types.Registry) (*CELPlugin, error) {
	wasm, err := NewWasmPlugin(ctx, cfg.Wasm, celRegistry)
	if err != nil {
		return nil, err
	}
	return &CELPlugin{
		Name:      cfg.Name,
		Functions: cfg.Functions,
		wasm:      wasm,
	}, nil
}

func (p *CELPlugin) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *CELPlugin) LibraryName() string {
	return p.Name
}

func (p *CELPlugin) CompileOptions() []cel.EnvOption {
	md, ok := metadata.FromIncomingContext(p.ctx)
	if !ok {
		md = make(metadata.MD)
	}
	var opts []cel.EnvOption
	for _, fn := range p.Functions {
		fn := fn
		bindFunc := cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			return p.wasm.Call(p.ctx, fn, md, args...)
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
