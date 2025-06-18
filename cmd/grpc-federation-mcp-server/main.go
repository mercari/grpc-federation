package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
)

//go:embed assets/*
var assets embed.FS

func main() {
	docs, err := assets.ReadFile("assets/docs.md")
	if err != nil {
		log.Fatalf("failed to read document: %v", err)
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"gRPC Federation MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
		server.WithInstructions(string(docs)),
	)

	s.AddTool(
		mcp.NewTool(
			"get_import_proto_list",
			mcp.WithDescription("Returns a list of import proto files used in the specified proto file"),
			mcp.WithString(
				"path",
				mcp.Description("The absolute path to the proto file to be analyzed"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := r.Params.Arguments.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("unexpected argument format: %T", r.Params.Arguments)
			}
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("failed to find path parameter from arguments: %v", args)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			f, err := source.NewFile(path, content)
			if err != nil {
				return nil, err
			}
			list, err := json.Marshal(append(f.Imports(), f.ImportsByImportRule()...))
			if err != nil {
				return nil, err
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: string(list),
					},
				},
			}, nil
		},
	)

	s.AddTool(
		mcp.NewTool(
			"compile_proto",
			mcp.WithDescription("Compile the proto file using the gRPC Federation option"),
			mcp.WithString(
				"path",
				mcp.Description("The absolute path to the proto file to be analyzed"),
				mcp.Required(),
			),
			mcp.WithArray(
				"import_paths",
				mcp.Description("Specify the list of import paths required to locate dependency files during compilation. It is recommended to obtain this list using the get_import_proto_list tool"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, ok := r.Params.Arguments.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("unexpected argument format: %T", r.Params.Arguments)
			}
			path, ok := args["path"].(string)
			if !ok {
				return nil, fmt.Errorf("failed to find path parameter from arguments: %v", args)
			}
			importPathsArg, ok := args["import_paths"].([]any)
			if !ok {
				return nil, fmt.Errorf("failed to find import_paths parameter from arguments: %v", args)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			file, err := source.NewFile(path, content)
			if err != nil {
				return nil, err
			}
			importPaths := make([]string, 0, len(importPathsArg))
			for _, p := range importPathsArg {
				importPath, ok := p.(string)
				if !ok {
					return nil, fmt.Errorf("failed to get import_paths element. required type is string but got %T", p)
				}
				importPaths = append(importPaths, importPath)
			}
			v := validator.New()
			outs := v.Validate(context.Background(), file, validator.ImportPathOption(importPaths...))
			if validator.ExistsError(outs) {
				return nil, fmt.Errorf("failed to compile:\n%s", validator.Format(outs))
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "build successful",
					},
				},
			}, nil
		},
	)

	s.AddResource(
		mcp.NewResource(
			"grpc-federation",
			"grpc-federation",
			mcp.WithResourceDescription("gRPC Federation Document"),
			mcp.WithMIMEType("text/markdown"),
		),
		func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      "https://github.com/mercari/grpc-federation",
					MIMEType: "text/markdown",
					Text:     string(docs),
				},
			}, nil
		},
	)

	s.AddPrompt(
		mcp.NewPrompt(
			"grpc_federation",
			mcp.WithPromptDescription("How to use gRPC Federation"),
		),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult("How to use gRPC Federation",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(
						mcp.RoleAssistant,
						mcp.NewTextContent(string(docs)),
					),
				},
			), nil
		},
	)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
