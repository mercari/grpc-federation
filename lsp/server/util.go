package server

import (
	"fmt"
	"net/url"
	"os"

	"go.lsp.dev/uri"
)

func (h *Handler) getFile(docURI uri.URI) (string, []byte, error) {
	parsedURI, err := url.ParseRequestURI(string(docURI))
	if err != nil {
		return "", nil, err
	}
	if parsedURI.Scheme != uri.FileScheme {
		return "", nil, fmt.Errorf("text document uri must specify the %s scheme, got %v", uri.FileScheme, parsedURI.Scheme)
	}
	path := parsedURI.Path

	h.fileContentMu.Lock()
	content, exists := h.fileToContentMap[path]
	h.fileContentMu.Unlock()
	if exists {
		return path, content, nil
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	return path, file, nil
}
