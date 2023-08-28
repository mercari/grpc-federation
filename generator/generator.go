package generator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
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
	validator          *validator.Validator
	codeGenerator      *CodeGenerator
	importPaths        []string
	postProcessHandler PostProcessHandler
}

func New(cfg Config) *Generator {
	return &Generator{
		cfg:           &cfg,
		compiler:      compiler.New(),
		validator:     validator.New(),
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
	svcs, err := g.generate(ctx, protoPath)
	if err != nil {
		return err
	}
	if len(svcs) != 0 && g.postProcessHandler != nil {
		if err := g.postProcessHandler(ctx, protoPath, svcs); err != nil {
			return err
		}
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

type CompileResult struct {
	GoFiles   []*ProtogenFile
	GRPCFiles []*ProtogenFile
	Services  []*resolver.Service
}

func (r *CompileResult) GoFile(name string) *ProtogenFile {
	for _, file := range r.GoFiles {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func (r *CompileResult) GRPCFile(name string) *ProtogenFile {
	for _, file := range r.GRPCFiles {
		if file.Name == name {
			return file
		}
	}
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
	writtenGoFileMap := make(map[string]struct{})
	writtenGRPCFileMap := make(map[string]struct{})
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
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return nil, err
		}
		if g.cfg.GetAutoProtocGenGo() {
			if file := result.GoFile(svc.File.Name); file != nil {
				if _, exists := writtenGoFileMap[file.Name]; !exists {
					fileName := g.fileNameWithoutExt(file.Name) + ".pb.go"
					if err := os.WriteFile(filepath.Join(filepath.Dir(outputPath), fileName), file.Content, 0o600); err != nil {
						return nil, err
					}
					writtenGoFileMap[file.Name] = struct{}{}
				}
			}
		}
		if g.cfg.GetAutoProtocGenGoGRPC() {
			if file := result.GRPCFile(svc.File.Name); file != nil {
				if _, exists := writtenGRPCFileMap[file.Name]; !exists {
					fileName := g.fileNameWithoutExt(file.Name) + "_grpc.pb.go"
					if err := os.WriteFile(filepath.Join(filepath.Dir(outputPath), fileName), file.Content, 0o600); err != nil {
						return nil, err
					}
					writtenGRPCFileMap[file.Name] = struct{}{}
				}
			}
		}
		if err := os.WriteFile(outputPath, federationOut, 0o600); err != nil {
			return nil, err
		}
		log.Printf("generated to %s", outputPath)
	}
	return result.Services, nil
}

func (g *Generator) fileNameWithoutExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func (g *Generator) compileProto(ctx context.Context, protoPath string) (*CompileResult, error) {
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
	if outs := g.validator.Validate(ctx, file, validator.ImportPathOption(g.importPaths...), validator.AutoImportOption()); len(outs) != 0 {
		return nil, errors.New(validator.Format(outs))
	}
	protos, err := g.compiler.Compile(ctx, file, compiler.ImportPathOption(g.importPaths...), compiler.AutoImportOption())
	if err != nil {
		return nil, err
	}
	plugin, err := protogen.Options{}.New(&pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{filepath.Base(protoPath)},
		ProtoFile:      protos,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}
	goFiles, err := g.runProtogenGo(plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to run protoc-gen-go: %w", err)
	}
	grpcFiles, err := g.runProtogenGoGRPC(plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to run protoc-gen-go-grpc: %w", err)
	}
	result, err := resolver.New(protos).Resolve()
	if err != nil {
		return nil, err
	}
	return &CompileResult{
		GoFiles:   goFiles,
		GRPCFiles: grpcFiles,
		Services:  result.Services,
	}, nil
}

type ProtogenFile struct {
	Name    string
	Content []byte
}

func (g *Generator) runProtogenGo(plugin *protogen.Plugin) ([]*ProtogenFile, error) {
	var files []*ProtogenFile
	for _, f := range plugin.Files {
		if f.Generate {
			generatedFile := gengo.GenerateFile(plugin, f)
			content, err := generatedFile.Content()
			if err != nil {
				return nil, err
			}
			files = append(files, &ProtogenFile{
				Name:    f.Proto.GetName(),
				Content: content,
			})
		}
	}
	return files, nil
}

func (g *Generator) runProtogenGoGRPC(plugin *protogen.Plugin) ([]*ProtogenFile, error) {
	var files []*ProtogenFile
	for _, f := range plugin.Files {
		if f.Generate {
			generatedFile := runProtogenGoGRPC(plugin, f, true)
			content, err := generatedFile.Content()
			if err != nil {
				return nil, err
			}
			files = append(files, &ProtogenFile{
				Name:    f.Proto.GetName(),
				Content: content,
			})
		}
	}
	return files, nil
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
