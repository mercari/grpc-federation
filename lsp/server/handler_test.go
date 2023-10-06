package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
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

func TestHandler_Completion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc        string
		params      *protocol.CompletionParams
		expected    *protocol.CompletionList
		expectedErr string
	}{
		{
			desc: "resolver method",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      36,
						Character: 15,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "post.PostService/GetPost",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      36,
									Character: 39,
								},
								End: protocol.Position{
									Line:      36,
									Character: 39,
								},
							},
						},
					},
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "post.PostService/GetPosts",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      36,
									Character: 39,
								},
								End: protocol.Position{
									Line:      36,
									Character: 39,
								},
							},
							NewText: "s",
						},
					},
				},
			},
		},
		{
			desc: "resolver request field",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      38,
						Character: 17,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "id",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      38,
									Character: 20,
								},
								End: protocol.Position{
									Line:      38,
									Character: 20,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "resolver request by",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      38,
						Character: 28,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "$.id",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      38,
									Character: 32,
								},
								End: protocol.Position{
									Line:      38,
									Character: 32,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "resolver response field",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      40,
						Character: 41,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "post",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      40,
									Character: 46,
								},
								End: protocol.Position{
									Line:      40,
									Character: 46,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "messages message",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      43,
						Character: 32,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{Data: 0, Kind: protocol.CompletionItemKindText, Label: "GetPostRequest"},
					{Data: 1, Kind: protocol.CompletionItemKindText, Label: "GetPostResponse"},
					{Data: 2, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.DescriptorProto"},
					{Data: 3, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.DescriptorProto.ExtensionRange"},
					{Data: 4, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.DescriptorProto.ReservedRange"},
					{Data: 5, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.EnumDescriptorProto"},
					{Data: 6, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.EnumDescriptorProto.EnumReservedRange"},
					{Data: 7, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.EnumOptions"},
					{Data: 8, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.EnumValueDescriptorProto"},
					{Data: 9, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.EnumValueOptions"},
					{Data: 10, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.ExtensionRangeOptions"},
					{Data: 11, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.ExtensionRangeOptions.Declaration"},
					{Data: 12, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.FieldDescriptorProto"},
					{Data: 13, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.FieldOptions"},
					{Data: 14, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.FileDescriptorProto"},
					{Data: 15, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.FileDescriptorSet"},
					{Data: 16, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.FileOptions"},
					{Data: 17, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.GeneratedCodeInfo"},
					{Data: 18, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.GeneratedCodeInfo.Annotation"},
					{Data: 19, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.MessageOptions"},
					{Data: 20, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.MethodDescriptorProto"},
					{Data: 21, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.MethodOptions"},
					{Data: 22, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.OneofDescriptorProto"},
					{Data: 23, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.OneofOptions"},
					{Data: 24, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.ServiceDescriptorProto"},
					{Data: 25, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.ServiceOptions"},
					{Data: 26, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.SourceCodeInfo"},
					{Data: 27, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.SourceCodeInfo.Location"},
					{Data: 28, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.UninterpretedOption"},
					{Data: 29, Kind: protocol.CompletionItemKindText, Label: "google.protobuf.UninterpretedOption.NamePart"},
					{Data: 30, Kind: protocol.CompletionItemKindText, Label: "post.GetPostRequest"},
					{Data: 31, Kind: protocol.CompletionItemKindText, Label: "post.GetPostResponse"},
					{Data: 32, Kind: protocol.CompletionItemKindText, Label: "post.GetPostsRequest"},
					{Data: 33, Kind: protocol.CompletionItemKindText, Label: "post.GetPostsResponse"},
					{Data: 34, Kind: protocol.CompletionItemKindText, Label: "post.Post"},
					{Data: 35, Kind: protocol.CompletionItemKindText, Label: "user.GetUserRequest"},
					{Data: 36, Kind: protocol.CompletionItemKindText, Label: "user.GetUserResponse"},
					{Data: 37, Kind: protocol.CompletionItemKindText, Label: "user.GetUsersRequest"},
					{Data: 38, Kind: protocol.CompletionItemKindText, Label: "user.GetUsersResponse"},
					{Data: 39, Kind: protocol.CompletionItemKindText, Label: "user.User"},
				},
			},
		},
		{
			desc: "invalid uri schema",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: protocol.URI("https://mercari.com"),
					},
				},
			},
			expectedErr: "text document uri must specify the file scheme, got https",
		},
		{
			desc: "field by",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      49,
						Character: 50,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "user",
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      49,
									Character: 54,
								},
								End: protocol.Position{
									Line:      49,
									Character: 54,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "file does not exist",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: protocol.URI("file://unknown.proto"),
					},
				},
			},
			expectedErr: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			conn := jsonrpc2.NewConn(
				jsonrpc2.NewStream(
					&struct {
						io.ReadCloser
						io.Writer
					}{
						io.NopCloser(&bytes.Buffer{}),
						&bytes.Buffer{},
					},
				),
			)
			client := protocol.ClientDispatcher(conn, zap.NewNop())

			handler := server.NewHandler(client, &bytes.Buffer{}, nil)

			got, err := handler.Completion(context.Background(), tc.params)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("failed to call Completion: %v", err)
				}

				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("received an unpecpected error: %q is supposed to contain %q", err.Error(), tc.expectedErr)
				}
				return
			}

			if tc.expectedErr != "" {
				t.Error("expected to receive an error but got nil")
			}

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
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
						Line:      43,
						Character: 32,
					},
				},
			},
			expected: []protocol.Location{
				{
					URI: mustTestdataAbs(t, "testdata/service.proto"),
					Range: protocol.Range{
						Start: protocol.Position{Line: 52, Character: 8},
						End:   protocol.Position{Line: 52, Character: 12},
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
						Line:      36,
						Character: 15,
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
			expected: nil,
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
