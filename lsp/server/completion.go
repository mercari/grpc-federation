package server

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/source"
)

func (h *Handler) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	file, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	pos := source.Position{Line: int(params.Position.Line) + 1, Col: int(params.Position.Character) + 1}
	nodeInfo, candidates, err := h.completer.Completion(ctx, file, h.importPaths, pos)
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
	h.logger.Info("processing candidates", slog.String("text", curText), slog.Any("candidates", candidates))
	items := make([]protocol.CompletionItem, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, curText) {
			endPos := nodeInfo.End()
			start := protocol.Position{
				Line:      uint32(endPos.Line) - 1, //nolint:gosec
				Character: uint32(endPos.Col) - 2,  //nolint:gosec
			}
			end := protocol.Position{
				Line:      uint32(endPos.Line) - 1, //nolint:gosec
				Character: uint32(endPos.Col) - 2,  //nolint:gosec
			}
			items = append(items, protocol.CompletionItem{
				Label: candidate,
				Kind:  protocol.CompletionItemKindText,
				Data:  len(items),
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
	logger *slog.Logger
}

func NewCompleter(logger *slog.Logger) *Completer {
	return &Completer{logger: logger}
}

func (c *Completer) Completion(ctx context.Context, file *File, importPaths []string, pos source.Position) (*ast.NodeInfo, []string, error) {
	r := resolver.New(file.getCompiledProtos(), resolver.ImportPathOption(importPaths...))
	_, _ = r.Resolve()
	loc := file.getSource().FindLocationByPos(pos)
	if loc == nil {
		return nil, nil, nil
	}
	return file.getSource().NodeInfoByLocation(loc), r.Candidates(loc), nil
}
