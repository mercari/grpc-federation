package generator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func evalCodeGeneratorPlugin(ctx context.Context, pluginFile []byte, req io.Reader) (*pluginpb.CodeGeneratorResponse, error) {
	runtimeCfg := wazero.NewRuntimeConfigInterpreter()
	if cache := getCompilationCache(); cache != nil {
		runtimeCfg = runtimeCfg.WithCompilationCache(cache)
	}

	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	buf := bytes.NewBuffer([]byte{})

	modCfg := wazero.NewModuleConfig().
		WithFSConfig(wazero.NewFSConfig().WithDirMount(".", "/")).
		WithStdin(req).
		WithStdout(buf).
		WithStderr(os.Stderr).
		WithArgs("wasi")
	if _, err := r.InstantiateWithConfig(ctx, pluginFile, modCfg); err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to instantiate code-generator plugin: %w", err)
	}
	var res pluginpb.CodeGeneratorResponse
	resBytes := buf.Bytes()
	if len(resBytes) != 0 {
		if err := proto.Unmarshal(resBytes, &res); err != nil {
			return nil, err
		}
	}
	return &res, nil
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
