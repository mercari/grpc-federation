package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"go.lsp.dev/protocol"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/source"
)

func (h *Handler) definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	path, content, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		return nil, err
	}
	pos := source.Position{Line: int(params.Position.Line) + 1, Col: int(params.Position.Character) + 1}
	h.logger.Info("created file", slog.String("path", file.Path()), slog.Any("ast", file.AST()))
	loc := file.FindLocationByPos(pos)
	if loc == nil {
		return nil, nil
	}
	nodeInfo := file.NodeInfoByLocation(loc)
	if nodeInfo == nil {
		return nil, nil
	}
	switch {
	case isMessageNameDefinition(loc):
		foundMsgName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found message", slog.String("name", foundMsgName))
		locs, err := h.findMessageDefinition(ctx, path, file, foundMsgName)
		if err != nil {
			return nil, err
		}
		return locs, nil
	case isMethodNameDefinition(loc):
		foundMethodName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found method", slog.String("name", foundMethodName))
		locs, err := h.findMethodDefinition(ctx, path, file, foundMethodName)
		if err != nil {
			return nil, err
		}
		return locs, nil
	}
	return nil, nil
}

func (h *Handler) findMessageDefinition(ctx context.Context, path string, file *source.File, defMsgName string) ([]protocol.Location, error) {
	protoFiles, err := h.compiler.Compile(ctx, file, compiler.ImportPathOption(h.importPaths...))
	if err != nil {
		return nil, err
	}
	msgName, filePath, err := h.messageAndFilePath(path, protoFiles, defMsgName)
	if filePath == "" {
		return nil, err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	defFile, err := source.NewFile(filePath, content)
	if err != nil {
		return nil, err
	}

	var foundNode ast.Node
	_ = ast.Walk(defFile.AST(), &ast.SimpleVisitor{
		DoVisitMessageNode: func(n *ast.MessageNode) error {
			if string(n.Name.AsIdentifier()) == msgName {
				foundNode = n.Name
			}
			return nil
		},
	})

	if foundNode == nil {
		return nil, fmt.Errorf("failed to find %s message from ast.Node in %s file", msgName, filePath)
	}

	return h.toLocation(defFile, foundNode), nil
}

func (h *Handler) findMethodDefinition(ctx context.Context, path string, file *source.File, defMethodName string) ([]protocol.Location, error) {
	protoFiles, err := h.compiler.Compile(ctx, file, compiler.ImportPathOption(h.importPaths...))
	if err != nil {
		return nil, err
	}
	methodName, filePath, err := h.methodAndFilePath(path, protoFiles, defMethodName)
	if filePath == "" {
		return nil, err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	defFile, err := source.NewFile(filePath, content)
	if err != nil {
		return nil, err
	}

	var foundNode ast.Node
	_ = ast.Walk(defFile.AST(), &ast.SimpleVisitor{
		DoVisitRPCNode: func(n *ast.RPCNode) error {
			if string(n.Name.AsIdentifier()) == methodName {
				foundNode = n.Name
			}
			return nil
		},
	})

	if foundNode == nil {
		return nil, fmt.Errorf("failed to find %s method from ast.Node in %s file", methodName, filePath)
	}

	return h.toLocation(defFile, foundNode), nil
}

func (h *Handler) toLocation(file *source.File, node ast.Node) []protocol.Location {
	info := file.AST().NodeInfo(node)
	startPos := info.Start()
	endPos := info.End()
	locRange := protocol.Range{
		Start: protocol.Position{
			Line:      uint32(startPos.Line) - 1,
			Character: uint32(startPos.Col) - 1,
		},
		End: protocol.Position{
			Line:      uint32(endPos.Line) - 1,
			Character: uint32(endPos.Col) - 1,
		},
	}
	return []protocol.Location{
		{
			URI:   protocol.DocumentURI(fmt.Sprintf("file://%s", file.Path())),
			Range: locRange,
		},
	}
}

func (h *Handler) messageAndFilePath(path string, protoFiles []*descriptorpb.FileDescriptorProto, defMsgName string) (string, string, error) {
	currentPkgName, err := h.currentPackageName(path, protoFiles)
	if err != nil {
		return "", "", err
	}
	h.logger.Info("current package", slog.String("name", currentPkgName))
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
			h.logger.Warn("failed to find a path from a proto", slog.String("error", err.Error()))
			continue
		}
		pkg := protoFile.GetPackage()
		for _, msg := range protoFile.GetMessageType() {
			var msgName string
			if pkg == currentPkgName {
				msgName = msg.GetName()
			} else {
				msgName = fmt.Sprintf("%s.%s", pkg, msg.GetName())
			}
			if defMsgName == msgName {
				return msg.GetName(), filePath, nil
			}
		}
	}
	return "", "", fmt.Errorf("failed to find %s message from all proto files", defMsgName)
}

func (h *Handler) methodAndFilePath(path string, protoFiles []*descriptorpb.FileDescriptorProto, defMethodName string) (string, string, error) {
	currentPkgName, err := h.currentPackageName(path, protoFiles)
	if err != nil {
		return "", "", err
	}
	h.logger.Info("current package", slog.String("name", currentPkgName))
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
			h.logger.Warn("failed to find a path from a proto", slog.String("error", err.Error()))
			continue
		}
		pkg := protoFile.GetPackage()
		for _, svc := range protoFile.GetService() {
			svcName := svc.GetName()
			for _, method := range svc.GetMethod() {
				var methodName string
				if pkg == currentPkgName {
					methodName = fmt.Sprintf("%s/%s", svcName, method.GetName())
				} else {
					methodName = fmt.Sprintf("%s.%s/%s", pkg, svcName, method.GetName())
				}
				if defMethodName == methodName {
					return method.GetName(), filePath, nil
				}
			}
		}
	}
	return "", "", fmt.Errorf("failed to find %s method from all proto files", defMethodName)
}

func (h *Handler) currentPackageName(path string, protoFiles []*descriptorpb.FileDescriptorProto) (string, error) {
	for _, protoFile := range protoFiles {
		if strings.HasSuffix(path, protoFile.GetName()) {
			return protoFile.GetPackage(), nil
		}
	}
	return "", fmt.Errorf("failed to find current package")
}

func (h *Handler) filePathFromFileDescriptorProto(path string, protoFile *descriptorpb.FileDescriptorProto) (string, error) {
	fileName := protoFile.GetName()
	if filepath.Base(path) == fileName {
		return path, nil
	}
	for _, importPath := range h.importPaths {
		fullpath := filepath.Join(importPath, fileName)
		if !filepath.IsAbs(fullpath) {
			p, err := filepath.Abs(fullpath)
			if err != nil {
				return "", err
			}
			fullpath = p
		}
		if _, err := os.Stat(fullpath); err == nil {
			return fullpath, nil
		}
	}
	return "", fmt.Errorf("failed to find absolute path from %s", fileName)
}

func isMessageNameDefinition(loc *source.Location) bool {
	if loc.Message == nil {
		return false
	}
	if loc.Message.Option == nil {
		return false
	}
	if loc.Message.Option.VariableDefinitions == nil {
		return false
	}
	if loc.Message.Option.VariableDefinitions.Message == nil {
		return false
	}
	return loc.Message.Option.VariableDefinitions.Message.Name
}

func isMethodNameDefinition(loc *source.Location) bool {
	if loc.Message == nil {
		return false
	}
	if loc.Message.Option == nil {
		return false
	}
	if loc.Message.Option.VariableDefinitions == nil {
		return false
	}
	if loc.Message.Option.VariableDefinitions.Call == nil {
		return false
	}
	return loc.Message.Option.VariableDefinitions.Call.Method
}
