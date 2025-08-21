package server_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"

	"github.com/mercari/grpc-federation/lsp/server"
)

func TestHandler_Initialize(t *testing.T) {
	t.Parallel()

	handler := server.NewHandler(nil, &bytes.Buffer{}, nil)

	params := &protocol.InitializeParams{
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				SemanticTokens: &protocol.SemanticTokensClientCapabilities{
					TokenTypes: []string{string(protocol.SemanticTokenNamespace), string(protocol.SemanticTokenType)},
				},
			},
		},
	}

	expected := &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync:   protocol.TextDocumentSyncKindFull,
			DefinitionProvider: true,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"$", ".", " "},
			},
			SemanticTokensProvider: map[string]interface{}{
				"legend": protocol.SemanticTokensLegend{
					TokenTypes: []protocol.SemanticTokenTypes{
						protocol.SemanticTokenNamespace,
						protocol.SemanticTokenType,
					},
				},
				"full": true,
			},
			SelectionRangeProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "grpc-federation",
			Version: "v0.0.1",
		},
	}

	got, err := handler.Initialize(context.Background(), params)
	if err != nil {
		t.Errorf("failed to call Initialize: %v", err)
	}

	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("(-want +got)\n%s", diff)
	}
}

func TestHandler_Definition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc     string
		params   *protocol.DefinitionParams
		expected []protocol.Location
	}{
		{
			desc: "message name definition",
			params: &protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      25,
						Character: 15,
					},
				},
			},
			expected: []protocol.Location{
				{
					URI: mustTestdataAbs(t, "testdata/service.proto"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 33, Character: 8},
						End:   protocol.Position{Line: 33, Character: 12},
					},
				},
			},
		},
		{
			desc: "method name definition",
			params: &protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      39,
						Character: 19,
					},
				},
			},
			expected: []protocol.Location{
				{
					URI: mustTestdataAbs(t, "testdata/post.proto"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 7, Character: 6},
						End:   protocol.Position{Line: 7, Character: 13},
					},
				},
			},
		},
		{
			desc: "no definition",
			params: &protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      20,
						Character: 0,
					},
				},
			},
			expected: []protocol.Location{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			client := protocol.ClientDispatcher(nil, zap.NewNop())
			handler := server.NewHandler(client, &bytes.Buffer{}, []string{"testdata"})

			got, err := handler.Definition(context.Background(), tc.params)
			if err != nil {
				t.Fatalf("failed to call Definition: %v", err)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}

func mustTestdataAbs(t *testing.T, path string) uri.URI {
	t.Helper()

	p, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to get an absolute file path: %v", err)
	}
	return uri.URI("file://" + p)
}
