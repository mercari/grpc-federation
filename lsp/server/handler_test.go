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
			desc: "def.call.method",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      43,
						Character: 18,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "post.PostService/GetPost",
						Data:  0,
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      43,
									Character: 43,
								},
								End: protocol.Position{
									Line:      43,
									Character: 43,
								},
							},
						},
					},
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "post.PostService/GetPosts",
						Data:  1,
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      43,
									Character: 43,
								},
								End: protocol.Position{
									Line:      43,
									Character: 43,
								},
							},
							NewText: "s",
						},
					},
				},
			},
		},
		{
			desc: "def.call.request.field",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      44,
						Character: 27,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "id",
						Data:  0,
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      44,
									Character: 30,
								},
								End: protocol.Position{
									Line:      44,
									Character: 30,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "def.call.request.by",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/service.proto"),
					},
					Position: protocol.Position{
						Line:      44,
						Character: 37,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "$.id",
						Data:  0,
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      44,
									Character: 42,
								},
								End: protocol.Position{
									Line:      44,
									Character: 42,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "filter response",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/completion.proto"),
					},
					Position: protocol.Position{
						Line:      47,
						Character: 27,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "res",
						Data:  0,
						TextEdit: &protocol.TextEdit{
							NewText: "res",
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      47,
									Character: 27,
								},
								End: protocol.Position{
									Line:      47,
									Character: 27,
								},
							},
						},
					},
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "$.id",
						Data:  1,
						TextEdit: &protocol.TextEdit{
							NewText: "$.id",
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      47,
									Character: 27,
								},
								End: protocol.Position{
									Line:      47,
									Character: 27,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "message",
			params: &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: mustTestdataAbs(t, "testdata/completion.proto"),
					},
					Position: protocol.Position{
						Line:      51,
						Character: 17,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					messageCompletionItem(0, "GetPostRequest"),
					messageCompletionItem(1, "GetPostResponse"),
					messageCompletionItem(2, "User"),
					messageCompletionItem(3, "google.protobuf.DescriptorProto"),
					messageCompletionItem(4, "google.protobuf.DescriptorProto.ExtensionRange"),
					messageCompletionItem(5, "google.protobuf.DescriptorProto.ReservedRange"),
					messageCompletionItem(6, "google.protobuf.Duration"),
					messageCompletionItem(7, "google.protobuf.EnumDescriptorProto"),
					messageCompletionItem(8, "google.protobuf.EnumDescriptorProto.EnumReservedRange"),
					messageCompletionItem(9, "google.protobuf.EnumOptions"),
					messageCompletionItem(10, "google.protobuf.EnumValueDescriptorProto"),
					messageCompletionItem(11, "google.protobuf.EnumValueOptions"),
					messageCompletionItem(12, "google.protobuf.ExtensionRangeOptions"),
					messageCompletionItem(13, "google.protobuf.ExtensionRangeOptions.Declaration"),
					messageCompletionItem(14, "google.protobuf.FieldDescriptorProto"),
					messageCompletionItem(15, "google.protobuf.FieldOptions"),
					messageCompletionItem(16, "google.protobuf.FileDescriptorProto"),
					messageCompletionItem(17, "google.protobuf.FileDescriptorSet"),
					messageCompletionItem(18, "google.protobuf.FileOptions"),
					messageCompletionItem(19, "google.protobuf.GeneratedCodeInfo"),
					messageCompletionItem(20, "google.protobuf.GeneratedCodeInfo.Annotation"),
					messageCompletionItem(21, "google.protobuf.MessageOptions"),
					messageCompletionItem(22, "google.protobuf.MethodDescriptorProto"),
					messageCompletionItem(23, "google.protobuf.MethodOptions"),
					messageCompletionItem(24, "google.protobuf.OneofDescriptorProto"),
					messageCompletionItem(25, "google.protobuf.OneofOptions"),
					messageCompletionItem(26, "google.protobuf.ServiceDescriptorProto"),
					messageCompletionItem(27, "google.protobuf.ServiceOptions"),
					messageCompletionItem(28, "google.protobuf.SourceCodeInfo"),
					messageCompletionItem(29, "google.protobuf.SourceCodeInfo.Location"),
					messageCompletionItem(30, "google.protobuf.UninterpretedOption"),
					messageCompletionItem(31, "google.protobuf.UninterpretedOption.NamePart"),
					messageCompletionItem(32, "google.rpc.BadRequest"),
					messageCompletionItem(33, "google.rpc.BadRequest.FieldViolation"),
					messageCompletionItem(34, "google.rpc.DebugInfo"),
					messageCompletionItem(35, "google.rpc.ErrorInfo"),
					messageCompletionItem(36, "google.rpc.ErrorInfo.MetadataEntry"),
					messageCompletionItem(37, "google.rpc.Help"),
					messageCompletionItem(38, "google.rpc.Help.Link"),
					messageCompletionItem(39, "google.rpc.LocalizedMessage"),
					messageCompletionItem(40, "google.rpc.PreconditionFailure"),
					messageCompletionItem(41, "google.rpc.PreconditionFailure.Violation"),
					messageCompletionItem(42, "google.rpc.QuotaFailure"),
					messageCompletionItem(43, "google.rpc.QuotaFailure.Violation"),
					messageCompletionItem(44, "google.rpc.RequestInfo"),
					messageCompletionItem(45, "google.rpc.ResourceInfo"),
					messageCompletionItem(46, "google.rpc.RetryInfo"),
					messageCompletionItem(47, "post.GetPostRequest"),
					messageCompletionItem(48, "post.GetPostResponse"),
					messageCompletionItem(49, "post.GetPostsRequest"),
					messageCompletionItem(50, "post.GetPostsResponse"),
					messageCompletionItem(51, "post.Post"),
					messageCompletionItem(52, "user.GetUserRequest"),
					messageCompletionItem(53, "user.GetUserResponse"),
					messageCompletionItem(54, "user.GetUsersRequest"),
					messageCompletionItem(55, "user.GetUsersResponse"),
					messageCompletionItem(56, "user.User"),
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
						Line:      60,
						Character: 47,
					},
				},
			},
			expected: &protocol.CompletionList{
				Items: []protocol.CompletionItem{
					{
						Kind:  protocol.CompletionItemKindText,
						Label: "user",
						Data:  0,
						TextEdit: &protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      60,
									Character: 51,
								},
								End: protocol.Position{
									Line:      60,
									Character: 51,
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

func messageCompletionItem(idx int, candidate string) protocol.CompletionItem {
	return protocol.CompletionItem{
		Data:  idx,
		Kind:  protocol.CompletionItemKindText,
		Label: candidate,
		TextEdit: &protocol.TextEdit{
			NewText: candidate,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      51,
					Character: 17,
				},
				End: protocol.Position{
					Line:      51,
					Character: 17,
				},
			},
		},
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
