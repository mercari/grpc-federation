package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
)

type PluginCommand struct {
	Cache PluginCacheCommand `description:"manage cache for plugin" command:"cache"`
}

type PluginCacheCommand struct {
	Create PluginCacheCreateCommand `description:"create cache for plugin" command:"create"`
}

type PluginCacheCreateCommand struct {
	Name string `description:"plugin name" long:"name" required:"true"`
	Out  string `description:"specify the output directory. If not specified, it will be created in the current directory" short:"o" long:"out"`
}

func (c *PluginCacheCreateCommand) Execute(args []string) error {
	if len(args) != 1 {
		return errors.New("you need to specify single wasm file")
	}
	ctx := context.Background()
	outDir := c.Out
	if outDir == "" {
		outDir = "."
	}
	cache, err := wazero.NewCompilationCacheWithDir(filepath.Join(outDir, "grpc-federation", c.Name))
	if err != nil {
		return fmt.Errorf("failed to setup cache directory %s: %w", outDir, err)
	}
	path := args[0]
	f, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	if _, err := wazero.NewRuntimeWithConfig(
		ctx,
		wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache),
	).CompileModule(ctx, f); err != nil {
		return fmt.Errorf("failed to compile wasm file: %w", err)
	}
	return nil
}
