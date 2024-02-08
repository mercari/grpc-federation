package generator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func evalCodeGeneratorPlugin(ctx context.Context, pluginFile []byte, req io.Reader) error {
	runtimeCfg := wazero.NewRuntimeConfig()
	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg)

	if _, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(wasmDebugLog).Export("grpc_federation_log").
		Instantiate(ctx); err != nil {
		return fmt.Errorf("grpc-federation: failed to create wasm host module: %w", err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	buf := bytes.NewBuffer([]byte{})
	modCfg := wazero.NewModuleConfig().
		WithFSConfig(wazero.NewFSConfig().WithDirMount(".", "/")).
		WithStdin(req).
		WithStdout(buf).
		WithStderr(buf).
		WithArgs("wasi")
	defer func() {
		out := buf.String()
		if out != "" {
			fmt.Fprintln(os.Stderr, out)
		}
	}()
	if _, err := r.InstantiateWithConfig(ctx, pluginFile, modCfg); err != nil {
		return fmt.Errorf("grpc-federation: failed to instantiate code-generator plugin: %w", err)
	}
	return nil
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
