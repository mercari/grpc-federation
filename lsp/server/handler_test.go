package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.lsp.dev/jsonrpc2"
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

func TestHandler_DidChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc     string
		params   *protocol.DidChangeTextDocumentParams
		expected protocol.PublishDiagnosticsParams
	}{
		{
			desc: "no validation errors",
			params: &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
				},
			},
			expected: protocol.PublishDiagnosticsParams{
				URI:         mustTestdataAbs(t, "testdata/service.proto"),
				Diagnostics: []protocol.Diagnostic{},
			},
		},
		{
			desc: "duplicated service dependency name",
			params: &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "../../validator/testdata/duplicate_service_dependency_name.proto"),
					},
				},
			},
			expected: protocol.PublishDiagnosticsParams{
				URI: mustTestdataAbs(t, "../../validator/testdata/duplicate_service_dependency_name.proto"),
				Diagnostics: []protocol.Diagnostic{
					{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      14,
								Character: 16,
							},
							End: protocol.Position{
								Line:      14,
								Character: 26,
							},
						},
						Severity: protocol.DiagnosticSeverityError,
						Message:  `"dup_name" name duplicated`,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()

			conn := jsonrpc2.NewConn(
				jsonrpc2.NewStream(
					&struct {
						io.ReadCloser
						io.Writer
					}{
						io.NopCloser(&bytes.Buffer{}),
						writer,
					},
				),
			)
			client := protocol.ClientDispatcher(conn, zap.NewNop())

			handler := server.NewHandler(client, &bytes.Buffer{}, nil)

			err := handler.DidChange(context.Background(), tc.params)
			if err != nil {
				t.Errorf("failed to call DidChange: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			errCh := make(chan error)
			resCh := make(chan []byte)
			go func() {
				// Ignore a header
				if _, err := reader.Read(make([]byte, 1024)); err != nil {
					errCh <- err
				}

				res := make([]byte, 1024)
				n, err := reader.Read(res)
				if err != nil {
					errCh <- err
				}

				resCh <- res[:n]
			}()

			select {
			case err := <-errCh:
				t.Errorf("failed to read a response: %v", err)
			case <-ctx.Done():
				t.Error("timeout occurs")
			case res := <-resCh:
				var notification jsonrpc2.Notification
				if err := json.Unmarshal(res, &notification); err != nil {
					t.Error(err)
				}
				var params protocol.PublishDiagnosticsParams
				if err := json.Unmarshal(notification.Params(), &params); err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(params, tc.expected); diff != "" {
					t.Errorf("(-got, +want)\n%s", diff)
				}
			}
		})
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
						Line:      29,
						Character: 15,
					},
				},
			},
			expected: []protocol.Location{
				{
					URI: mustTestdataAbs(t, "testdata/service.proto"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 37, Character: 8},
						End:   protocol.Position{Line: 37, Character: 12},
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
						Line:      43,
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
						Line:      19,
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
