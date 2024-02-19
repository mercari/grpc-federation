package server

import (
	"go.lsp.dev/protocol"
)

func (h *Handler) initialize(params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	var (
		tokenTypes     []protocol.SemanticTokenTypes
		tokenModifiers []protocol.SemanticTokenModifiers
	)
	semanticTokens := params.Capabilities.TextDocument.SemanticTokens
	for idx, tokenType := range semanticTokens.TokenTypes {
		tokenTypes = append(tokenTypes, protocol.SemanticTokenTypes(tokenType))
		h.tokenTypeMap[tokenType] = uint32(idx)
	}
	for idx, modifier := range semanticTokens.TokenModifiers {
		tokenModifiers = append(tokenModifiers, protocol.SemanticTokenModifiers(modifier))
		h.tokenModifierMap[modifier] = 1 << uint32(idx)
	}
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync:   protocol.TextDocumentSyncKindFull,
			DefinitionProvider: true,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"$", ".", " "},
			},
			SemanticTokensProvider: map[string]interface{}{
				"legend": protocol.SemanticTokensLegend{
					TokenTypes:     tokenTypes,
					TokenModifiers: tokenModifiers,
				},
				"full": true,
			},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "grpc-federation",
			Version: "v0.0.1",
		},
	}, nil
}
