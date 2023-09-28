package server

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/k0kubun/pp/v3"
	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

func (h *Handler) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	path, content, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	pos := source.Position{Line: int(params.Position.Line) + 1, Col: int(params.Position.Character) + 1}
	nodeInfo, candidates, err := h.completer.Completion(ctx, h.importPaths, path, content, pos)
	if err != nil {
		return nil, nil
	}
	if nodeInfo == nil {
		return nil, nil
	}
	curText := nodeInfo.RawText()
	unquoted, err := strconv.Unquote(curText)
	if err == nil {
		curText = unquoted
	}
	h.logger.Println("text = ", curText, "candidates = ", candidates)
	items := make([]protocol.CompletionItem, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, curText) {
			endPos := nodeInfo.End()
			start := protocol.Position{
				Line:      uint32(endPos.Line) - 1,
				Character: uint32(endPos.Col) - 2,
			}
			end := protocol.Position{
				Line:      uint32(endPos.Line) - 1,
				Character: uint32(endPos.Col) - 2,
			}
			items = append(items, protocol.CompletionItem{
				Label: candidate,
				Kind:  protocol.CompletionItemKindText,
				TextEdit: &protocol.TextEdit{
					Range:   protocol.Range{Start: start, End: end},
					NewText: candidate[len(curText):],
				},
			})
		}
	}
	if len(items) == 0 {
		for idx, candidate := range candidates {
			items = append(items, protocol.CompletionItem{
				Label: candidate,
				Kind:  protocol.CompletionItemKindText,
				Data:  idx,
			})
		}
	}
	return &protocol.CompletionList{
		Items: items,
	}, nil
}

type Completer struct {
	compiler *compiler.Compiler
	logger   *log.Logger
	pp       *pp.PrettyPrinter
}

func NewCompleter(c *compiler.Compiler, logger *log.Logger, printer *pp.PrettyPrinter) *Completer {
	return &Completer{compiler: c, logger: logger, pp: printer}
}

func (c *Completer) Completion(ctx context.Context, importPaths []string, path string, content []byte, pos source.Position) (*ast.NodeInfo, []string, error) {
	file, err := source.NewFile(path, content)
	if err != nil {
		return nil, nil, err
	}
	protos, err := c.compiler.Compile(ctx, file, compiler.ImportPathOption(importPaths...))
	if err != nil {
		return nil, nil, err
	}
	r := resolver.New(protos)
	_, _ = r.Resolve()
	loc := file.FindLocationByPos(pos)
	if loc == nil {
		return nil, nil, nil
	}
	return file.NodeInfoByLocation(loc), r.Candidates(loc), nil
}
