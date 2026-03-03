package generator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// wasmPlugin holds a compiled WASM module and its runtime, allowing
// the module to be instantiated multiple times without recompilation.
// Execute is safe for concurrent use.
type wasmPlugin struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	counter  atomic.Int64
}

func newWasmPlugin(ctx context.Context, wasmBytes []byte) (*wasmPlugin, error) {
	runtimeCfg := wazero.NewRuntimeConfigInterpreter()
	if cache := getCompilationCache(); cache != nil {
		runtimeCfg = runtimeCfg.WithCompilationCache(cache)
	}
	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	compiled, err := r.CompileModule(ctx, wasmBytes)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("grpc-federation: failed to compile code-generator plugin: %w", err)
	}
	return &wasmPlugin{
		runtime:  r,
		compiled: compiled,
	}, nil
}

func (p *wasmPlugin) Execute(ctx context.Context, req io.Reader) (*pluginpb.CodeGeneratorResponse, error) {
	name := fmt.Sprintf("plugin_%d", p.counter.Add(1))

	buf := new(bytes.Buffer)
	modCfg := wazero.NewModuleConfig().
		WithFSConfig(wazero.NewFSConfig().WithDirMount(".", "/")).
		WithStdin(req).
		WithStdout(buf).
		WithStderr(os.Stderr).
		WithName(name).
		WithArgs("wasi")

	mod, err := p.runtime.InstantiateModule(ctx, p.compiled, modCfg)
	if err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to instantiate code-generator plugin: %w", err)
	}
	mod.Close(ctx)

	var res pluginpb.CodeGeneratorResponse
	resBytes := buf.Bytes()
	if len(resBytes) != 0 {
		if err := proto.Unmarshal(resBytes, &res); err != nil {
			return nil, err
		}
	}
	return &res, nil
}

func (p *wasmPlugin) Close(ctx context.Context) error {
	return p.runtime.Close(ctx)
}

// wasmPluginCache caches compiled WASM plugins so that the expensive
// compilation step is performed only once per plugin path.
type wasmPluginCache struct {
	plugins map[string]*wasmPlugin
}

func newWasmPluginCache() *wasmPluginCache {
	return &wasmPluginCache{plugins: make(map[string]*wasmPlugin)}
}

func (c *wasmPluginCache) getOrCreate(ctx context.Context, opt *WasmPluginOption) (*wasmPlugin, error) {
	if wp, ok := c.plugins[opt.Path]; ok {
		return wp, nil
	}
	wasmFile, err := os.ReadFile(opt.Path)
	if err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to read plugin file: %s: %w", opt.Path, err)
	}
	if opt.Sha256 != "" {
		hash := sha256.Sum256(wasmFile)
		gotHash := hex.EncodeToString(hash[:])
		if opt.Sha256 != gotHash {
			return nil, fmt.Errorf(
				`grpc-federation: expected plugin sha256 value is [%s] but got [%s]`,
				opt.Sha256,
				gotHash,
			)
		}
	}
	wp, err := newWasmPlugin(ctx, wasmFile)
	if err != nil {
		return nil, err
	}
	c.plugins[opt.Path] = wp
	return wp, nil
}

func (c *wasmPluginCache) Close(ctx context.Context) {
	for _, wp := range c.plugins {
		wp.Close(ctx)
	}
}

func getCompilationCache() wazero.CompilationCache {
	tmpDir := os.TempDir()
	if tmpDir == "" {
		return nil
	}
	cacheDir := filepath.Join(tmpDir, "grpc-federation")
	if _, err := os.Stat(cacheDir); err != nil {
		if err := os.Mkdir(cacheDir, 0o755); err != nil {
			return nil
		}
	}
	cache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		return nil
	}
	return cache
}
