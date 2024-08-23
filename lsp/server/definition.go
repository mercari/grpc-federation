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

func (h *Handler) definitionWithLink(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.LocationLink, error) {
	path, content, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		return nil, err
	}
	pos := source.Position{Line: int(params.Position.Line) + 1, Col: int(params.Position.Character) + 1}
	loc := file.FindLocationByPos(pos)
	if loc == nil {
		return nil, nil
	}
	nodeInfo := file.NodeInfoByLocation(loc)
	if nodeInfo == nil {
		return nil, nil
	}
	h.logger.Info("node", slog.String("text", nodeInfo.RawText()))
	protoFiles, err := h.compiler.Compile(ctx, file, compiler.ImportPathOption(h.importPaths...), compiler.ImportRuleOption())
	if err != nil {
		return nil, err
	}
	switch {
	case isImportNameDefinition(loc):
		foundImportFile, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found import", slog.String("file", foundImportFile))
		locs := h.findImportFileDefinition(path, protoFiles, foundImportFile)
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case isMessageNameDefinition(loc):
		foundMsgName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found message", slog.String("name", foundMsgName))
		locs, err := h.findTypeDefinition(path, protoFiles, foundMsgName)
		if err != nil {
			return nil, err
		}
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case isFieldType(loc.Message):
		var prefix []string
		if loc.Message != nil {
			prefix = loc.Message.MessageNames()
		}
		foundTypeName := nodeInfo.RawText()
		if types.ToKind(foundTypeName) != types.Unknown {
			// buildtin types
			return nil, nil
		}
		firstPriorTypeName := strings.Join(append(prefix, foundTypeName), ".")
		for _, typeName := range []string{firstPriorTypeName, foundTypeName} {
			h.logger.Info("found message", slog.String("name", typeName))
			locs, err := h.findTypeDefinition(path, protoFiles, typeName)
			if err != nil {
				continue
			}
			if len(locs) == 0 {
				continue
			}
			return h.toLocationLinks(nodeInfo, locs, false), nil
		}
	case isTypeAlias(loc):
		typeName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found type", slog.String("name", typeName))
		locs, err := h.findTypeDefinition(path, protoFiles, typeName)
		if err != nil {
			return nil, err
		}
		return h.toLocationLinks(nodeInfo, locs, true), nil
	case isMethodNameDefinition(loc):
		foundMethodName, err := strconv.Unquote(nodeInfo.RawText())
		if err != nil {
			return nil, err
		}
		h.logger.Info("found method", slog.String("name", foundMethodName))
		locs, err := h.findMethodDefinition(path, protoFiles, foundMethodName)
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
			h.logger.Warn("failed to find a path from a proto", slog.String("error", err.Error()))
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
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	defFile, err := source.NewFile(filePath, content)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(typeName, ".")
	lastPart := parts[len(parts)-1]

	var foundNode ast.Node
	_ = ast.Walk(defFile.AST(), &ast.SimpleVisitor{
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

	return h.toLocation(defFile, foundNode), nil
}

func (h *Handler) findMethodDefinition(path string, protoFiles []*descriptorpb.FileDescriptorProto, defMethodName string) ([]protocol.Location, error) {
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
	h.logger.Info("current package", slog.String("name", currentPkgName))
	for _, protoFile := range protoFiles {
		filePath, err := h.filePathFromFileDescriptorProto(path, protoFile)
		if err != nil {
			h.logger.Warn("failed to find a path from a proto", slog.String("error", err.Error()))
			continue
		}
		pkg := protoFile.GetPackage()
		for _, msg := range protoFile.GetMessageType() {
			for _, name := range getDeclaredTypeNames(msg) {
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
	}
	return "", "", fmt.Errorf("failed to find %s type from all proto files", defTypeName)
}

func getDeclaredTypeNames(msg *descriptorpb.DescriptorProto) []string {
	name := msg.GetName()
	ret := []string{name}
	for _, msg := range msg.GetNestedType() {
		for _, nested := range getDeclaredTypeNames(msg) {
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

func isImportNameDefinition(loc *source.Location) bool {
	return loc.ImportName != ""
}

func isMessageNameDefinition(loc *source.Location) bool {
	if loc.Message == nil {
		return false
	}
	if loc.Message.Option == nil {
		return false
	}
	if loc.Message.Option.Def == nil {
		return false
	}
	if loc.Message.Option.Def.Message == nil {
		return false
	}
	return loc.Message.Option.Def.Message.Name
}

func isTypeAlias(loc *source.Location) bool {
	if isTypeAliasByMessage(loc.Message) {
		return true
	}
	if isTypeAliasByEnum(loc.Enum) {
		return true
	}
	return false
}

func isTypeAliasByMessage(msg *source.Message) bool {
	if msg == nil {
		return false
	}
	if msg.Option != nil && msg.Option.Alias {
		return true
	}
	if msg.Field != nil {
		if msg.Field.Option != nil && msg.Field.Option.Alias {
			return true
		}
	}
	if msg.NestedMessage != nil {
		return isTypeAliasByMessage(msg.NestedMessage)
	}
	if msg.Enum != nil {
		return isTypeAliasByEnum(msg.Enum)
	}
	return false
}

func isTypeAliasByEnum(enum *source.Enum) bool {
	if enum == nil {
		return false
	}
	if enum.Option != nil && enum.Option.Alias {
		return true
	}
	if enum.Value != nil {
		if enum.Value.Option != nil && enum.Value.Option.Alias {
			return true
		}
	}
	return false
}

func isFieldType(msg *source.Message) bool {
	if msg == nil {
		return false
	}
	if msg.NestedMessage != nil {
		if isFieldType(msg.NestedMessage) {
			return true
		}
	}
	if msg.Field == nil {
		return false
	}
	return msg.Field.Type
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
