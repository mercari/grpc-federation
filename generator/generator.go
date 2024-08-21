package generator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/grpc/federation/generator"
	"github.com/mercari/grpc-federation/grpc/federation/generator/plugin"
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
	cfg                       *Config
	federationGeneratorOption *CodeGeneratorOption
	watcher                   *Watcher
	compiler                  *compiler.Compiler
	validator                 *validator.Validator
	importPaths               []string
	postProcessHandler        PostProcessHandler
	buildCacheMap             BuildCacheMap
	absPathToRelativePath     map[string]string
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
	ProtocAction ActionType = "protoc"
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
	if err := evalAllCodeGenerationPlugin(ctx, results, g.federationGeneratorOption); err != nil {
		return err
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

func newPluginRequest(protoPath string, org *pluginpb.CodeGeneratorRequest, pluginOpt *PluginOption) (*PluginRequest, error) {
	opt := pluginOpt.String()
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
	var (
		stdout, stderr bytes.Buffer
	)
	//nolint:gosec // only valid values are set to cfg.installedPath
	cmd := exec.CommandContext(ctx, cfg.installedPath)
	cmd.Stdin = req.content
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"grpc-federation: %s: %s: %w",
			stdout.String(),
			stderr.String(),
			err,
		)
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
		if err := evalAllCodeGenerationPlugin(ctx, results, g.federationGeneratorOption); err != nil {
			log.Printf("failed to run code generator plugin: %+v", err)
		}
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
	opt, err := parseOptString(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	opt.Path.ImportPaths = g.cfg.Imports
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(opt.Path)
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
	opt, err := parseOptString(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	opt.Path.ImportPaths = g.cfg.Imports
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(opt.Path)

	var res pluginpb.CodeGeneratorResponse
	for _, f := range r.genplugin.Files {
		if !f.Generate {
			continue
		}
		generatedFile := runProtogenGoGRPC(r.genplugin, f, true)
		if generatedFile == nil {
			continue
		}
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
	opt, err := parseOptString(r.req.GetParameter())
	if err != nil {
		return nil, err
	}
	opt.Path.ImportPaths = g.cfg.Imports
	g.federationGeneratorOption = opt
	relativePath := g.absPathToRelativePath[r.protoPath]
	pathResolver := resolver.NewOutputFilePathResolver(opt.Path)

	result, err := resolver.New(r.req.GetProtoFile(), resolver.ImportPathOption(opt.Path.ImportPaths...)).Resolve()
	if err != nil {
		return nil, err
	}

	var resp pluginpb.CodeGeneratorResponse
	for _, file := range result.Files {
		out, err := NewCodeGenerator().Generate(file, result.Enums)
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
	result, err := resolver.New(r.req.GetProtoFile(), resolver.ImportPathOption(g.cfg.Imports...)).Resolve()
	if err != nil {
		return nil, err
	}
	return result.Files, nil
}

func (g *Generator) fileNameWithoutExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func CreateCodeGeneratorResponse(ctx context.Context, req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	opt, err := parseOptString(req.GetParameter())
	if err != nil {
		return nil, err
	}
	outputPathResolver := resolver.NewOutputFilePathResolver(opt.Path)
	result, err := resolver.New(req.GetProtoFile(), resolver.ImportPathOption(opt.Path.ImportPaths...)).Resolve()
	if outs := validator.New().ToValidationOutputByResolverResult(result, err, validator.ImportPathOption(opt.Path.ImportPaths...)); len(outs) > 0 {
		if validator.ExistsError(outs) {
			return nil, errors.New(validator.Format(outs))
		}
		fmt.Fprint(os.Stderr, validator.Format(outs))
	}

	var outDir string
	if len(result.Files) != 0 {
		outPath, err := outputPathResolver.OutputPath(result.Files[0])
		if err != nil {
			return nil, err
		}
		outDir = filepath.Dir(outPath)
	}
	if err := evalAllCodeGenerationPlugin(ctx, []*ProtoFileResult{
		{
			Type:            ProtocAction,
			ProtoPath:       "",
			Out:             outDir,
			FederationFiles: result.Files,
		},
	}, opt); err != nil {
		return nil, err
	}

	var resp pluginpb.CodeGeneratorResponse
	for _, file := range result.Files {
		out, err := NewCodeGenerator().Generate(file, result.Enums)
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

func evalAllCodeGenerationPlugin(ctx context.Context, results []*ProtoFileResult, opt *CodeGeneratorOption) error {
	if len(results) == 0 {
		return nil
	}
	if opt == nil || len(opt.Plugins) == 0 {
		return nil
	}
	for _, result := range results {
		pluginFiles := make([]*plugin.ProtoCodeGeneratorResponse_File, 0, len(result.Files))
		for _, file := range result.Files {
			fileBytes, err := proto.Marshal(file)
			if err != nil {
				return err
			}
			var pluginFile plugin.ProtoCodeGeneratorResponse_File
			if err := proto.Unmarshal(fileBytes, &pluginFile); err != nil {
				return err
			}
			pluginFiles = append(pluginFiles, &pluginFile)
		}
		genReq := generator.CreateCodeGeneratorRequest(&generator.CodeGeneratorRequestConfig{
			Type:                generator.ActionType(result.Type),
			ProtoPath:           result.ProtoPath,
			OutDir:              result.Out,
			Files:               pluginFiles,
			GRPCFederationFiles: result.FederationFiles,
		})
		encodedGenReq, err := proto.Marshal(genReq)
		if err != nil {
			return err
		}
		genReqReader := bytes.NewBuffer(encodedGenReq)
		for _, plugin := range opt.Plugins {
			wasmFile, err := os.ReadFile(plugin.Path)
			if err != nil {
				return fmt.Errorf("grpc-federation: failed to read plugin file: %s: %w", plugin.Path, err)
			}
			hash := sha256.Sum256(wasmFile)
			gotHash := hex.EncodeToString(hash[:])
			if plugin.Sha256 != gotHash {
				return fmt.Errorf(
					`grpc-federation: expected plugin sha256 value is [%s] but got [%s]`,
					plugin.Sha256,
					gotHash,
				)
			}
			if err := evalCodeGeneratorPlugin(ctx, wasmFile, genReqReader); err != nil {
				return err
			}
			genReqReader.Reset()
		}
	}
	return nil
}

type CodeGeneratorOption struct {
	Path    resolver.OutputFilePathConfig
	Plugins []*WasmPluginOption
}

type WasmPluginOption struct {
	Path   string
	Sha256 string
}

func parseOptString(opt string) (*CodeGeneratorOption, error) {
	var ret CodeGeneratorOption
	for _, part := range splitOpt(opt) {
		if err := parseOpt(&ret, part); err != nil {
			return nil, err
		}
	}
	return &ret, nil
}

func parseOpt(opt *CodeGeneratorOption, pat string) error {
	if pat == "" {
		// nothing option.
		return nil
	}
	partOpt, err := splitOptPattern(pat)
	if err != nil {
		return err
	}
	switch partOpt.kind {
	case "module":
		if err := parseModuleOption(opt, partOpt.value); err != nil {
			return err
		}
	case "paths":
		if err := parsePathsOption(opt, partOpt.value); err != nil {
			return err
		}
	case "plugins":
		if err := parsePluginsOption(opt, partOpt.value); err != nil {
			return err
		}
	case "import_paths":
		if err := parseImportPathsOption(opt, partOpt.value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("grpc-federation: unexpected option: %s", partOpt.kind)
	}
	return nil
}

func parseModuleOption(opt *CodeGeneratorOption, value string) error {
	if value == "" {
		return fmt.Errorf(`grpc-federation: failed to find prefix name for module option`)
	}
	opt.Path.Mode = resolver.ModulePrefixMode
	opt.Path.Prefix = value
	return nil
}

func parsePathsOption(opt *CodeGeneratorOption, value string) error {
	switch value {
	case "source_relative":
		opt.Path.Mode = resolver.SourceRelativeMode
	case "import":
		opt.Path.Mode = resolver.ImportMode
	default:
		return fmt.Errorf("grpc-federation: unexpected paths option: %s", value)
	}
	return nil
}

var (
	schemeToOptionParser = map[string]func(*CodeGeneratorOption, string) error{
		"file://":  parseFileSchemeOption,
		"http://":  parseHTTPSchemeOption,
		"https://": parseHTTPSSchemeOption,
	}
)

func parsePluginsOption(opt *CodeGeneratorOption, value string) error {
	for scheme, parser := range schemeToOptionParser {
		if strings.HasPrefix(value, scheme) {
			if err := parser(opt, strings.TrimPrefix(value, scheme)); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf(
		`grpc-federation: location of the plugin file must be specified with the "file://" or "http(s)://" schemes but specified %s`,
		value,
	)
}

func parseFileSchemeOption(opt *CodeGeneratorOption, value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf(`grpc-federation: plugin option must be specified with "file://path/to/file.wasm:sha256hash" format like "file://plugin.wasm:abcdefg"`)
	}
	path := parts[0]
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("grpc-federation: failed to get absolute path by %s: %w", path, err)
		}
		path = abs
	}
	opt.Plugins = append(opt.Plugins, &WasmPluginOption{
		Path:   path,
		Sha256: parts[1],
	})
	return nil
}

func parseHTTPSchemeOption(opt *CodeGeneratorOption, value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf(`grpc-federation: plugin option must be specified with "file://path/to/file.wasm:sha256hash" format like "file://plugin.wasm:abcdefg"`)
	}
	file, err := downloadFile("http://" + parts[0])
	if err != nil {
		return err
	}
	opt.Plugins = append(opt.Plugins, &WasmPluginOption{
		Path:   file.Name(),
		Sha256: parts[1],
	})
	return nil
}

func parseHTTPSSchemeOption(opt *CodeGeneratorOption, value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf(`grpc-federation: plugin option must be specified with "file://path/to/file.wasm:sha256hash" format like "file://plugin.wasm:abcdefg"`)
	}
	file, err := downloadFile("https://" + parts[0])
	if err != nil {
		return err
	}
	opt.Plugins = append(opt.Plugins, &WasmPluginOption{
		Path:   file.Name(),
		Sha256: parts[1],
	})
	return nil
}

func downloadFile(url string) (*os.File, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp("", "grpc-federation-code-generation-plugin")
	if err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return nil, fmt.Errorf("grpc-federation: failed to copy downloaded content to temp file: %w", err)
	}
	return f, nil
}

func splitOpt(opt string) []string {
	return strings.Split(opt, ",")
}

type partOption struct {
	kind  string
	value string
}

func splitOptPattern(opt string) (*partOption, error) {
	parts := strings.Split(opt, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("grpc-federation: unexpected option format: %s", opt)
	}
	return &partOption{
		kind:  parts[0],
		value: parts[1],
	}, nil
}

func parseImportPathsOption(opt *CodeGeneratorOption, value string) error {
	opt.Path.ImportPaths = append(opt.Path.ImportPaths, value)
	return nil
}
