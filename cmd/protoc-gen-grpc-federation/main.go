package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/generator"
	"github.com/mercari/grpc-federation/resolver"
)

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	req, err := parseRequest(os.Stdin)
	if err != nil {
		return err
	}
	resp, err := run(req)
	if err != nil {
		return err
	}
	if err := outputResponse(resp); err != nil {
		return err
	}
	return nil
}

func parseRequest(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req pluginpb.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func run(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	cfg, err := parseOpt(req.GetParameter())
	if err != nil {
		return nil, err
	}
	result, err := resolver.New(req.GetProtoFile()).Resolve()
	if err != nil {
		return nil, err
	}
	outputPathResolver := resolver.NewOutputFilePathResolver(cfg)
	var resp pluginpb.CodeGeneratorResponse
	for _, service := range result.Services {
		out, err := generator.NewCodeGenerator().Generate(service)
		if err != nil {
			return nil, err
		}
		outputFilePath, err := outputPathResolver.OutputPath(service)
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

func outputResponse(resp *pluginpb.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err = os.Stdout.Write(buf); err != nil {
		return err
	}
	return nil
}
