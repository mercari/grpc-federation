package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func (h *Handler) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	parsedURI, err := url.ParseRequestURI(string(params.TextDocument.URI))
	if err != nil {
		return err
	}
	if parsedURI.Scheme != uri.FileScheme {
		return fmt.Errorf("text document uri must specify the %s scheme, got %v", uri.FileScheme, parsedURI.Scheme)
	}
	if len(params.ContentChanges) == 0 {
		return nil
	}
	path := parsedURI.Path
	content := params.ContentChanges[0].Text
	h.logger.Debug("sync text", slog.String("path", path))
	if _, err := h.setFileCache(path, []byte(content)); err != nil {
		return err
	}
	h.validateText(ctx, params.TextDocument.URI)
	return nil
}
