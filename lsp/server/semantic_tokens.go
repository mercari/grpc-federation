package server

import (
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/source"
)

func (h *Handler) semanticTokensFull(params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	path, content, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		return nil, nil
	}
	var (
		line = 1
		col  = 1
		data []uint32
	)
	tokenTypeMap := file.SemanticTokenTypeMap()
	_ = ast.Walk(file.AST(), &ast.SimpleVisitor{
		DoVisitTerminalNode: func(n ast.TerminalNode) error {
			curToken := n.Token()
			info := file.AST().TokenInfo(curToken)

			comments := info.LeadingComments()
			for i := 0; i < comments.Len(); i++ {
				comment := comments.Index(i)
				pos := comment.Start()
				for _, text := range strings.Split(comment.RawText(), "\n") {
					diffLine, diffCol := h.calcLineAndCol(pos, line, col)
					line = pos.Line
					col = pos.Col
					tokenTypeNumber := h.tokenTypeMap[string(protocol.SemanticTokenComment)]
					data = append(data, uint32(diffLine), uint32(diffCol), uint32(len(text)), tokenTypeNumber, 0)
					pos = ast.SourcePos{
						Line: pos.Line + 1,
						Col:  pos.Col,
					}
				}
			}

			startPos := info.Start()
			diffLine, diffCol := h.calcLineAndCol(startPos, line, col)
			line = startPos.Line
			col = startPos.Col
			tokenType, exists := tokenTypeMap[n.Token()]
			if !exists {
				switch n.(type) {
				case *ast.IdentNode:
					tokenType = protocol.SemanticTokenVariable
				case *ast.KeywordNode:
					tokenType = protocol.SemanticTokenKeyword
				case *ast.UintLiteralNode:
					tokenType = protocol.SemanticTokenNumber
				case *ast.StringLiteralNode:
					tokenType = protocol.SemanticTokenString
				case *ast.RuneNode:
					tokenType = protocol.SemanticTokenOperator
				}
			}
			tokenTypeNumber := h.tokenTypeMap[string(tokenType)]
			data = append(data, uint32(diffLine), uint32(diffCol), uint32(len(info.RawText())), tokenTypeNumber, 0)
			return nil
		},
	})
	return &protocol.SemanticTokens{Data: data}, nil
}

func (h *Handler) calcLineAndCol(pos ast.SourcePos, line, col int) (int, int) {
	if pos.Line == line {
		return 0, pos.Col - col
	}
	return pos.Line - line, pos.Col - 1
}
