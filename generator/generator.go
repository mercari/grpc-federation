package generator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
)

const (
	protocGenGRPCFederation = "protoc-gen-grpc-federation"
	protocGenGo             = "protoc-gen-go"
	protocGenGoGRPC         = "protoc-gen-go-grpc"
)

type Generator struct {
	cfg                   *Config
	watcher               *Watcher
	compiler              *compiler.Compiler
	validator             *validator.Validator
	importPaths           []string
	postProcessHandler    PostProcessHandler
	buildCacheMap         BuildCacheMap
	absPathToRelativePath map[string]string
}

type Option func(*Generator) error

type PostProcessHandler func(context.Context, string, Result) error

type BuildCache struct {
	Responses       []*pluginpb.CodeGeneratorResponse
	FederationFiles []*resolver.File
}

type BuildCacheMap map[string]*BuildCache

type Result []*ProtoFileResult

type ProtoFileResult struct {
	ProtoPath       string
	Type            ActionType
	Files           []*pluginpb.CodeGeneratorResponse_File
	FederationFiles []*resolver.File
	Out             string
}

type ActionType string

const (
	KeepAction   ActionType = "keep"
	CreateAction ActionType = "create"
	DeleteAction ActionType = "delete"
	UpdateAction ActionType = "update"
)

type PluginRequest struct {
	req       *pluginpb.CodeGeneratorRequest
	content   *bytes.Buffer
	genplugin *protogen.Plugin
	protoPath string
}

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

func (r *ProtoFileResult) WriteFiles(ctx context.Context) error {
	switch r.Type {
	case DeleteAction:
		for _, file := range r.Files {
			path := filepath.Join(r.Out, file.GetName())
			if !existsPath(path) {
				continue
			}
			log.Printf("remove %s file", path)
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	default:
		for _, file := range r.Files {
			path := filepath.Join(r.Out, file.GetName())
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			log.Printf("write %s file", path)
			if err := os.WriteFile(path, []byte(file.GetContent()), 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}

func New(cfg Config) *Generator {
	return &Generator{
		cfg:                   &cfg,
		compiler:              compiler.New(),
		validator:             validator.New(),
		importPaths:           cfg.Imports,
		absPathToRelativePath: make(map[string]string),
	}
}

func (g *Generator) SetPostProcessHandler(postProcessHandler func(ctx context.Context, path string, result Result) error) {
	g.postProcessHandler = postProcessHandler
}

func (g *Generator) Generate(ctx context.Context, protoPath string, opts ...Option) error {
	path, err := filepath.Abs(protoPath)
	if err != nil {
		return err
	}
	g.absPathToRelativePath[path] = protoPath
	for _, opt := range opts {
		if err := opt(g); err != nil {
			return err
		}
	}
	if g.buildCacheMap == nil {
		buildCacheMap, err := g.GenerateAll(ctx)
		if err != nil {
			return err
		}
		g.buildCacheMap = buildCacheMap
	}
	if g.watcher != nil {
		defer g.watcher.Close()
		for _, src := range g.cfg.Src {
			log.Printf("watch %v directory's proto files", src)
		}
		return g.watcher.Run(ctx)
	}

	results := g.otherResults(path)
	if _, exists := g.buildCacheMap[path]; exists {
		result, err := g.updateProtoFile(ctx, path)
		if err != nil {
			return err
		}
		results = append(results, result)
	} else {
		result, err := g.createProtoFile(ctx, path)
		if err != nil {
			return err
		}
		results = append(results, result)
	}
	if g.postProcessHandler != nil {
		if err := g.postProcessHandler(ctx, path, results); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) GenerateAll(ctx context.Context) (BuildCacheMap, error) {
	protoPathMap, err := g.createProtoPathMap()
	if err != nil {
		return nil, err
	}
	buildCacheMap := BuildCacheMap{}
	for protoPath := range protoPathMap {
		req, err := g.compileProto(ctx, protoPath)
		if err != nil {
			return nil, err
		}
		for _, pluginCfg := range g.cfg.Plugins {
			pluginReq, err := newPluginRequest(protoPath, req, pluginCfg.Opt)
			if err != nil {
				return nil, err
			}
			res, err := g.generateByPlugin(ctx, pluginReq, pluginCfg)
			if err != nil {
				return nil, err
			}
			if res == nil {
				continue
			}
			buildCache := buildCacheMap[protoPath]
			if buildCache == nil {
				buildCache = &BuildCache{}
				buildCacheMap[protoPath] = buildCache
			}
			if pluginCfg.Plugin == "protoc-gen-grpc-federation" {
				files, err := g.createGRPCFederationFiles(pluginReq)
				if err != nil {
					return nil, err
				}
				buildCache.FederationFiles = append(buildCache.FederationFiles, files...)
			}
			buildCache.Responses = append(buildCache.Responses, res)
		}
	}
	return buildCacheMap, nil
}

func newPluginRequest(protoPath string, org *pluginpb.CodeGeneratorRequest, opt string) (*PluginRequest, error) {
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate:  org.FileToGenerate,
		Parameter:       &opt,
		ProtoFile:       org.ProtoFile,
		CompilerVersion: org.CompilerVersion,
	}
	content, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	genplugin, err := protogen.Options{}.New(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create protogen.Plugin: %w", err)
	}
	return &PluginRequest{
		protoPath: protoPath,
		req:       req,
		content:   bytes.NewBuffer(content),
		genplugin: genplugin,
	}, nil
}

func (g *Generator) generateByPlugin(ctx context.Context, req *PluginRequest, cfg *PluginConfig) (*pluginpb.CodeGeneratorResponse, error) {
	if cfg.installedPath == "" {
		switch cfg.Plugin {
		case protocGenGo:
			return g.generateByProtogenGo(req)
		case protocGenGoGRPC:
			return g.generateByProtogenGoGRPC(req)
		case protocGenGRPCFederation:
			return g.generateByGRPCFederation(req)
		}
		return nil, fmt.Errorf("failed to find installed path for %s", cfg.Plugin)
	}
	var stdout bytes.Buffer
	//nolint:gosec // only valid values are set to cfg.installedPath
	cmd := exec.CommandContext(ctx, cfg.installedPath)
	cmd.Stdin = req.content
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var res pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(stdout.Bytes(), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (g *Generator) createProtoPathMap() (map[string]struct{}, error) {
	protoPathMap := map[string]struct{}{}
	for _, src := range g.cfg.Src {
		if err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) != ".proto" {
				return nil
			}
			abspath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			g.absPathToRelativePath[abspath] = path
			protoPathMap[abspath] = struct{}{}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return protoPathMap, nil
}

func (g *Generator) setWatcher(w *Watcher) error {
	if err := w.SetWatchPath(g.cfg.Src...); err != nil {
		return err
	}
	w.SetHandler(func(ctx context.Context, event fsnotify.Event) {
		path, err := filepath.Abs(event.Name)
		if err != nil {
			log.Printf("failed to create absolute path from %s: %+v", event.Name, err)
			return
		}
		g.absPathToRelativePath[path] = event.Name

		var results []*ProtoFileResult
		switch {
		case event.Has(fsnotify.Create):
			result, err := g.createProtoFile(ctx, path)
			if err != nil {
				log.Printf("failed to generate from created proto %s: %+v", path, err)
				return
			}
			results = append(results, result)
		case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
			results = append(results, g.deleteProtoFile(path))
		case event.Has(fsnotify.Write):
			result, err := g.updateProtoFile(ctx, path)
			if err != nil {
				log.Printf("failed to generate from updated proto %s: %+v", path, err)
				return
			}
			results = append(results, result)
		}
		results = append(results, g.otherResults(path)...)
		if g.postProcessHandler != nil {
			if err := g.postProcessHandler(ctx, path, results); err != nil {
				log.Printf("%+v", err)
			}
		}
	})
	g.watcher = w
	return nil
}

func existsPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (g *Generator) otherResults(path string) []*ProtoFileResult {
	results := make([]*ProtoFileResult, 0, len(g.buildCacheMap))
	for p, buildCache := range g.buildCacheMap {
		if path == p {
			continue
		}
		if !existsPath(p) {
			// Sometimes fsnotify cannot detect a file remove event, so it may contain a path that does not exist.
			// If the path does not exist, create result for delete.
			results = append(results, g.deleteProtoFile(p))
			continue
		}

		result := &ProtoFileResult{
			ProtoPath:       p,
			Out:             g.cfg.Out,
			Type:            KeepAction,
			FederationFiles: buildCache.FederationFiles,
		}
		for _, r := range buildCache.Responses {
			result.Files = append(result.Files, r.GetFile()...)
		}
		results = append(results, result)
	}
	return results
}

func (g *Generator) createProtoFile(ctx context.Context, path string) (*ProtoFileResult, error) {
	result, err := g.createGeneratorResult(ctx, path)
	if err != nil {
		return nil, err
	}
	result.Type = CreateAction
	return result, nil
}

func (g *Generator) updateProtoFile(ctx context.Context, path string) (*ProtoFileResult, error) {
	delete(g.buildCacheMap, path)

	result, err := g.createGeneratorResult(ctx, path)
	if err != nil {
		return nil, err
	}
	result.Type = UpdateAction
	return result, nil
}

func (g *Generator) deleteProtoFile(path string) *ProtoFileResult {
	result := &ProtoFileResult{
		ProtoPath: path,
		Type:      DeleteAction,
		Out:       g.cfg.Out,
	}
	buildCache, exists := g.buildCacheMap[path]
	if exists {
		result.FederationFiles = buildCache.FederationFiles
		for _, resp := range buildCache.Responses {
			result.Files = append(result.Files, resp.GetFile()...)
		}
	}
	delete(g.buildCacheMap, path)
	return result
}

func (g *Generator) createGeneratorResult(ctx context.Context, path string) (*ProtoFileResult, error) {
	req, err := g.compileProto(ctx, path)
	if err != nil {
		return nil, err
	}
	result := &ProtoFileResult{
		ProtoPath: path,
		Out:       g.cfg.Out,
	}
	for _, pluginCfg := range g.cfg.Plugins {
		pluginReq, err := newPluginRequest(path, req, pluginCfg.Opt)
		if err != nil {
			return nil, err
		}
		resp, err := g.generateByPlugin(ctx, pluginReq, pluginCfg)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			continue
		}
		buildCache := g.buildCacheMap[path]
		if buildCache == nil {
			buildCache = &BuildCache{}
			g.buildCacheMap[path] = buildCache
		}
		if pluginCfg.Plugin == "protoc-gen-grpc-federation" {
			files, err := g.createGRPCFederationFiles(pluginReq)
			if err != nil {
				return nil, err
			}
			result.FederationFiles = files
			buildCache.FederationFiles = append(buildCache.FederationFiles, files...)
		}
		result.Files = append(result.Files, resp.GetFile()...)
		buildCache.Responses = append(buildCache.Responses, resp)
	}
	return result, nil
}

func (g *Generator) compileProto(ctx context.Context, protoPath string) (*pluginpb.CodeGeneratorRequest, error) {
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
	if outs := g.validator.Validate(ctx, file, validator.ImportPathOption(g.importPaths...)); len(outs) != 0 {
		out := validator.Format(outs)
		if validator.ExistsError(outs) {
			return nil, errors.New(out)
		}
		fmt.Fprint(os.Stdout, out)
	}
	protos, err := g.compiler.Compile(ctx, file, compiler.ImportPathOption(g.importPaths...))
	if err != nil {
		return nil, err
	}
	relPath, err := compiler.RelativePathFromImportPaths(protoPath, g.importPaths)
	if err != nil {
		return nil, err
	}
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{relPath},
		ProtoFile:      protos,
	}, nil
}

func (g *Generator) generateByProtogenGo(r *PluginRequest) (*pluginpb.CodeGeneratorResponse, error) {
	cfg, err := parseOpt(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	cfg.ImportPaths = g.cfg.Imports
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(cfg)
	var res pluginpb.CodeGeneratorResponse
	for _, f := range r.genplugin.Files {
		if !f.Generate {
			continue
		}
		gopkg, err := resolver.ResolveGoPackage(f.Proto)
		if err != nil {
			return nil, err
		}
		dir, err := pathResolver.OutputDir(relativePath, gopkg)
		if err != nil {
			return nil, err
		}
		generatedFile := gengo.GenerateFile(r.genplugin, f)
		content, err := generatedFile.Content()
		if err != nil {
			return nil, err
		}
		c := string(content)
		fileName := filepath.Base(f.Proto.GetName())
		path := filepath.Join(dir, g.fileNameWithoutExt(fileName)+".pb.go")
		res.File = append(res.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    &path,
			Content: &c,
		})
	}
	return &res, nil
}

func (g *Generator) generateByProtogenGoGRPC(r *PluginRequest) (*pluginpb.CodeGeneratorResponse, error) {
	cfg, err := parseOpt(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	cfg.ImportPaths = g.cfg.Imports
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(cfg)

	var res pluginpb.CodeGeneratorResponse
	for _, f := range r.genplugin.Files {
		if !f.Generate {
			continue
		}
		generatedFile := runProtogenGoGRPC(r.genplugin, f, true)
		content, err := generatedFile.Content()
		if err != nil {
			return nil, err
		}
		gopkg, err := resolver.ResolveGoPackage(f.Proto)
		if err != nil {
			return nil, err
		}
		dir, err := pathResolver.OutputDir(relativePath, gopkg)
		if err != nil {
			return nil, err
		}
		c := string(content)
		fileName := filepath.Base(f.Proto.GetName())
		path := filepath.Join(dir, g.fileNameWithoutExt(fileName)+"_grpc.pb.go")
		res.File = append(res.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    &path,
			Content: &c,
		})
	}
	return &res, nil
}

func (g *Generator) generateByGRPCFederation(r *PluginRequest) (*pluginpb.CodeGeneratorResponse, error) {
	cfg, err := parseOpt(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	cfg.ImportPaths = g.cfg.Imports
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(cfg)

	result, err := resolver.New(r.req.GetProtoFile()).Resolve()
	if err != nil {
		return nil, err
	}

	var resp pluginpb.CodeGeneratorResponse
	for _, file := range result.Files {
		out, err := NewCodeGenerator().Generate(file)
		if err != nil {
			return nil, err
		}
		dir, err := pathResolver.OutputDir(relativePath, file.GoPackage)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(dir, pathResolver.FileName(file))
		resp.File = append(resp.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(path),
			Content: proto.String(string(out)),
		})
	}
	return &resp, nil
}

func (g *Generator) createGRPCFederationFiles(r *PluginRequest) ([]*resolver.File, error) {
	result, err := resolver.New(r.req.GetProtoFile()).Resolve()
	if err != nil {
		return nil, err
	}
	return result.Files, nil
}

func (g *Generator) fileNameWithoutExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func CreateCodeGeneratorResponse(ctx context.Context, req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	cfg, err := parseOpt(req.GetParameter())
	if err != nil {
		return nil, err
	}
	outputPathResolver := resolver.NewOutputFilePathResolver(cfg)
	result, err := resolver.New(req.GetProtoFile()).Resolve()
	if err != nil {
		return nil, err
	}

	var resp pluginpb.CodeGeneratorResponse
	for _, file := range result.Files {
		out, err := NewCodeGenerator().Generate(file)
		if err != nil {
			return nil, err
		}
		outputFilePath, err := outputPathResolver.OutputPath(file)
		if err != nil {
			return nil, err
		}
		resp.File = append(resp.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(outputFilePath),
			Content: proto.String(string(out)),
		})
	}
	return &resp, nil
}

var (
	modulePrefixMatcher = regexp.MustCompile(`module=(.+)`)
)

func parseOpt(opt string) (resolver.OutputFilePathConfig, error) {
	var cfg resolver.OutputFilePathConfig
	switch {
	case strings.Contains(opt, "module="):
		cfg.Mode = resolver.ModulePrefixMode
		matched := modulePrefixMatcher.FindAllStringSubmatch(opt, 1)
		if len(matched) != 1 {
			return cfg, fmt.Errorf(`grpc-federation: failed to find prefix name from module option`)
		}
		if len(matched[0]) != 2 {
			return cfg, fmt.Errorf(`grpc-federation: failed to find prefix name from module option`)
		}
		cfg.Prefix = matched[0][1]
	case strings.Contains(opt, "paths=source_relative"):
		cfg.Mode = resolver.SourceRelativeMode
	case strings.Contains(opt, "paths=import"):
		cfg.Mode = resolver.ImportMode
	default:
		if opt == "" {
			cfg.Mode = resolver.ImportMode // default output mode
		} else {
			return cfg, fmt.Errorf(`grpc-federation: unexpected options found "%s"`, opt)
		}
	}
	return cfg, nil
}
