package server

import (
	"context"
	"fmt"
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
	path := parsedURI.Path
	h.fileContentMu.Lock()
	for _, change := range params.ContentChanges {
		h.fileToContentMap[path] = []byte(change.Text)
	}
	h.fileContentMu.Unlock()
	h.validateText(ctx, params.TextDocument.URI)
	return nil
}
