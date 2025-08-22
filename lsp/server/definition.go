//nolint:gosec
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

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/types"
)

func (h *Handler) definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	locs, err := h.definitionWithLink(ctx, params)
	if err != nil {
		return nil, err
	}
	ret := make([]protocol.Location, 0, len(locs))
	for _, loc := range locs {
		ret = append(ret, protocol.Location{
			URI:   loc.TargetURI,
			Range: loc.TargetRange,
		})
	}
	return ret, nil
}

func (h *Handler) definitionWithLink(_ context.Context, params *protocol.DefinitionParams) ([]protocol.LocationLink, error) {
	file, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	pos := source.Position{Line: int(params.Position.Line) + 1, Col: int(params.Position.Character) + 1}
	h.logger.Debug("definition", slog.Any("pos", pos))
	loc := file.getSource().FindLocationByPos(pos)
	if loc == nil {
		return nil, nil
	}
	h.logger.Debug("definition", slog.Any("loc", loc))
	nodeInfo := file.getSource().NodeInfoByLocation(loc)
	if nodeInfo == nil {
		return nil, nil
	}
	h.logger.Debug("definition", slog.String("node_text", nodeInfo.RawText()))
	switch {
	case isImportNameDefinition(loc):
		importFilePath, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Debug("import", slog.String("path", importFilePath))
		locs := h.findImportFileDefinition(file.getPath(), file.getCompiledProtos(), importFilePath)
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case loc.IsDefinedTypeName():
		typeName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Debug("type", slog.String("name", typeName))
		locs, err := h.findTypeDefinition(file.getPath(), file.getCompiledProtos(), typeName)
		if err != nil {
			return nil, err
		}
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case loc.IsDefinedFieldType():
		var prefix []string
		if loc.Message != nil {
			prefix = loc.Message.MessageNames()
		}
		typeName := nodeInfo.RawText()
		if types.ToKind(typeName) != types.Unknown {
			// builtin types
			return nil, nil
		}
		firstPriorTypeName := strings.Join(append(prefix, typeName), ".")
		for _, name := range []string{firstPriorTypeName, typeName} {
			h.logger.Debug("type", slog.String("name", name))
			locs, err := h.findTypeDefinition(file.getPath(), file.getCompiledProtos(), name)
			if err != nil {
				continue
			}
			if len(locs) == 0 {
				continue
			}
			return h.toLocationLinks(nodeInfo, locs, false), nil
		}
	case loc.IsDefinedTypeAlias():
		typeName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Debug("type", slog.String("name", typeName))
		locs, err := h.findTypeDefinition(file.getPath(), file.getCompiledProtos(), typeName)
		if err != nil {
			return nil, err
		}
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case isMethodNameDefinition(loc):
		mtdName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Debug("method", slog.String("name", mtdName))
		locs, err := h.findMethodDefinition(file.getPath(), file.getCompiledProtos(), mtdName)
		if err != nil {
			return nil, err
		}
		return h.toLocationLinks(nodeInfo, locs, true), nil
	}
	return nil, nil
}

func (h *Handler) findImportFileDefinition(path string, protoFiles []*descriptorpb.FileDescriptorProto, fileName string) []protocol.Location {
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
			continue
		}
		if strings.HasSuffix(filePath, fileName) {
			return []protocol.Location{
				{
					URI: protocol.DocumentURI(fmt.Sprintf("file://%s", filePath)),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 0},
						End:   protocol.Position{Line: 0, Character: 0},
					},
				},
			}
		}
	}
	return nil
}

func (h *Handler) findTypeDefinition(path string, protoFiles []*descriptorpb.FileDescriptorProto, defTypeName string) ([]protocol.Location, error) {
	typeName, filePath, err := h.typeAndFilePath(path, protoFiles, defTypeName)
	if filePath == "" {
		return nil, err
	}
	f, err := h.getFileByPath(filePath)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]

	var foundNode ast.Node
	_ = ast.Walk(f.getSource().AST(), &ast.SimpleVisitor{
		DoVisitMessageNode: func(n *ast.MessageNode) error {
			if string(n.Name.AsIdentifier()) == lastPart {
				foundNode = n.Name
			}
			return nil
		},
		DoVisitEnumNode: func(n *ast.EnumNode) error {
			if string(n.Name.AsIdentifier()) == lastPart {
				foundNode = n.Name
			}
			return nil
		},
	})

	if foundNode == nil {
		return nil, fmt.Errorf("failed to find %s type from ast.Node in %s file", typeName, filePath)
	}

	return h.toLocation(f.getSource(), foundNode), nil
}

func (h *Handler) findMethodDefinition(path string, protoFiles []*descriptorpb.FileDescriptorProto, defMethodName string) ([]protocol.Location, error) {
	methodName, filePath, err := h.methodAndFilePath(path, protoFiles, defMethodName)
	if filePath == "" {
		return nil, err
	}
	f, err := h.getFileByPath(filePath)
	if err != nil {
		return nil, err
	}
	var foundNode ast.Node
	_ = ast.Walk(f.getSource().AST(), &ast.SimpleVisitor{
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

	return h.toLocation(f.getSource(), foundNode), nil
}

func (h *Handler) toLocationLinks(nodeInfo *ast.NodeInfo, locs []protocol.Location, isQuoted bool) []protocol.LocationLink {
	ret := make([]protocol.LocationLink, 0, len(locs))
	startPos := nodeInfo.Start()
	endPos := nodeInfo.End()
	for _, loc := range locs {
		startCol := uint32(startPos.Col - 1)
		endCol := uint32(endPos.Col - 1)
		if isQuoted {
			startCol += 1
			endCol -= 1
		}
		ret = append(ret, protocol.LocationLink{
			OriginSelectionRange: &protocol.Range{
				Start: protocol.Position{
					Line:      uint32(startPos.Line) - 1,
					Character: startCol,
				},
				End: protocol.Position{
					Line:      uint32(endPos.Line) - 1,
					Character: endCol,
				},
			},
			TargetURI:            loc.URI,
			TargetRange:          loc.Range,
			TargetSelectionRange: loc.Range,
		})
	}
	return ret
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

func (h *Handler) typeAndFilePath(path string, protoFiles []*descriptorpb.FileDescriptorProto, defTypeName string) (string, string, error) {
	currentPkgName, err := h.currentPackageName(path, protoFiles)
	if err != nil {
		return "", "", err
	}
	h.logger.Debug("current package", slog.String("name", currentPkgName))
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
			continue
		}
		pkg := protoFile.GetPackage()
		for _, msg := range protoFile.GetMessageType() {
			for _, name := range getDeclaredMessageNames(msg) {
				var typeName string
				if pkg == currentPkgName {
					typeName = name
				} else {
					typeName = fmt.Sprintf("%s.%s", pkg, name)
				}
				if defTypeName == typeName {
					return typeName, filePath, nil
				}
			}
		}
		for _, enum := range protoFile.GetEnumType() {
			name := enum.GetName()
			var typeName string
			if pkg == currentPkgName {
				typeName = name
			} else {
				typeName = fmt.Sprintf("%s.%s", pkg, name)
			}
			if defTypeName == typeName {
				return typeName, filePath, nil
			}
		}
	}
	return "", "", fmt.Errorf("failed to find %s type from all proto files", defTypeName)
}

func getDeclaredMessageNames(msg *descriptorpb.DescriptorProto) []string {
	name := msg.GetName()
	ret := []string{name}
	for _, msg := range msg.GetNestedType() {
		for _, nested := range getDeclaredMessageNames(msg) {
			ret = append(ret, name+"."+nested)
		}
	}
	for _, enum := range msg.GetEnumType() {
		ret = append(ret, name+"."+enum.GetName())
	}
	return ret
}

func (h *Handler) methodAndFilePath(path string, protoFiles []*descriptorpb.FileDescriptorProto, defMethodName string) (string, string, error) {
	currentPkgName, err := h.currentPackageName(path, protoFiles)
	if err != nil {
		return "", "", err
	}
	h.logger.Debug("current package", slog.String("name", currentPkgName))
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
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

func isImportNameDefinition(loc *source.Location) bool {
	return loc.ImportName != ""
}

func isMethodNameDefinition(loc *source.Location) bool {
	if loc.Message == nil {
		return false
	}
	if loc.Message.Option == nil {
		return false
	}
	if loc.Message.Option.Def == nil {
		return false
	}
	if loc.Message.Option.Def.Call == nil {
		return false
	}
	return loc.Message.Option.Def.Call.Method
}
