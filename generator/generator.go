package generator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

type Option func(*Generator) error

type PostProcessHandler func(context.Context, string, []*resolver.Service) error

func WatchMode() func(*Generator) error {
	return func(g *Generator) error {
		w, err := NewWatcher()
		if err != nil {
			return err
		}
		if err := g.setWatcher(w); err != nil {
			return err
		}
		return nil
	}
}

type Generator struct {
	cfg                *Config
	watcher            *Watcher
	compiler           *compiler.Compiler
	codeGenerator      *CodeGenerator
	importPaths        []string
	postProcessHandler PostProcessHandler
}

func New(cfg Config) *Generator {
	return &Generator{
		cfg:           &cfg,
		compiler:      compiler.New(),
		codeGenerator: NewCodeGenerator(),
		importPaths:   cfg.Imports,
	}
}

func (g *Generator) SetPostProcessHandler(postProcessHandler func(ctx context.Context, path string, svcs []*resolver.Service) error) {
	g.postProcessHandler = postProcessHandler
}

func (g *Generator) Generate(ctx context.Context, protoPath string, opts ...Option) error {
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return err
		}
	}
	if g.watcher != nil {
		defer g.watcher.Close()
		for _, src := range g.cfg.Src {
			log.Printf("watch %v directory's proto files", src)
		}
		return g.watcher.Run(ctx)
	}
	if _, err := g.generate(ctx, protoPath); err != nil {
		return err
	}
	return nil
}

func (g *Generator) setWatcher(w *Watcher) error {
	if err := w.SetWatchPath(g.cfg.Src...); err != nil {
		return err
	}
	w.SetHandler(func(ctx context.Context, path string) {
		svcs, err := g.generate(ctx, path)
		if err != nil {
			log.Printf("%+v", err)
			return
		}
		if len(svcs) == 0 {
			return
		}
		if g.postProcessHandler != nil {
			if err := g.postProcessHandler(ctx, path, svcs); err != nil {
				log.Printf("%+v", err)
			}
		}
	})
	g.watcher = w
	return nil
}

func (g *Generator) generate(ctx context.Context, protoPath string) ([]*resolver.Service, error) {
	result, err := g.compileProto(ctx, protoPath)
	if err != nil {
		return nil, err
	}
	if len(result.Services) == 0 {
		return nil, nil
	}

	outputFilePathResolver := resolver.NewOutputFilePathResolver(g.getOutputFilePathConfig(protoPath))
	for _, svc := range result.Services {
		log.Printf("generate for %s service...", svc.FQDN())
		federationOut, err := g.codeGenerator.Generate(svc)
		if err != nil {
			return nil, err
		}
		path, err := outputFilePathResolver.OutputPath(svc)
		if err != nil {
			return nil, err
		}
		outputPath := filepath.Join(g.cfg.Out, path)
		if err := os.WriteFile(outputPath, federationOut, 0o600); err != nil {
			return nil, err
		}
		log.Printf("generated to %s", outputPath)
	}
	return result.Services, nil
}

func (g *Generator) compileProto(ctx context.Context, protoPath string) (*resolver.Result, error) {
	if protoPath == "" {
		return nil, fmt.Errorf("grpc-federation: proto file path is empty")
	}
	if filepath.Ext(protoPath) != ".proto" {
		return nil, fmt.Errorf("grpc-federation: %s is not proto file", protoPath)
	}
	log.Printf("compile %s", protoPath)
	content, err := os.ReadFile(protoPath)
	if err != nil {
		return nil, err
	}
	file, err := source.NewFile(protoPath, content)
	if err != nil {
		return nil, fmt.Errorf("failed to create source file: %w", err)
	}
	protos, err := g.compiler.Compile(ctx, file, compiler.ImportPathOption(g.importPaths...))
	if err != nil {
		return nil, err
	}
	result, err := resolver.New(protos).Resolve()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *Generator) getOutputFilePathConfig(filePath string) resolver.OutputFilePathConfig {
	if g.cfg.Option == nil {
		return resolver.OutputFilePathConfig{
			ImportPaths: g.importPaths,
			FilePath:    filePath,
		}
	}
	if g.cfg.Option.Module != "" {
		return resolver.OutputFilePathConfig{
			Mode:        resolver.ModulePrefixMode,
			Prefix:      g.cfg.Option.Module,
			ImportPaths: g.importPaths,
			FilePath:    filePath,
		}
	}
	if g.cfg.Option.Paths == "source_relative" {
		return resolver.OutputFilePathConfig{
			Mode:        resolver.SourceRelativeMode,
			ImportPaths: g.importPaths,
			FilePath:    filePath,
		}
	}
	return resolver.OutputFilePathConfig{
		ImportPaths: g.importPaths,
		FilePath:    filePath,
	}
}
