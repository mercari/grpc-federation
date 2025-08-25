//nolint:gosec
package server

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/google/cel-go/common"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/parser"
	"go.lsp.dev/protocol"
)

type SemanticTokenProvider struct {
	logger           *slog.Logger
	tokenMap         map[ast.Token][]*SemanticToken
	file             *File
	tree             *ast.FileNode
	tokenTypeMap     map[string]uint32
	tokenModifierMap map[string]uint32
}

type SemanticToken struct {
	Line      uint32
	Col       uint32
	Text      string
	Type      string
	Modifiers []string
}

func NewSemanticTokenProvider(logger *slog.Logger, file *File, tokenTypeMap map[string]uint32, tokenModifierMap map[string]uint32) *SemanticTokenProvider {
	return &SemanticTokenProvider{
		logger:           logger,
		tokenMap:         make(map[ast.Token][]*SemanticToken),
		file:             file,
		tree:             file.getSource().AST(),
		tokenTypeMap:     tokenTypeMap,
		tokenModifierMap: tokenModifierMap,
	}
}

func (p *SemanticTokenProvider) createSemanticToken(tk ast.Token, tokenType string, modifiers []string) *SemanticToken {
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	return &SemanticToken{
		Line:      uint32(pos.Line),
		Col:       uint32(pos.Col),
		Text:      text,
		Type:      tokenType,
		Modifiers: modifiers,
	}
}

func (p *SemanticTokenProvider) SemanticTokens() *protocol.SemanticTokens {
	_ = ast.Walk(p.tree, &ast.SimpleVisitor{
		DoVisitMessageNode: func(msg *ast.MessageNode) error {
			tk := msg.Name.Token()
			p.tokenMap[tk] = append(
				p.tokenMap[tk],
				p.createSemanticToken(tk, string(protocol.SemanticTokenType), nil),
			)
			return nil
		},
		DoVisitFieldNode: func(field *ast.FieldNode) error {
			switch typ := field.FldType.(type) {
			case *ast.IdentNode:
				tk := typ.Token()
				p.tokenMap[tk] = append(
					p.tokenMap[tk],
					p.createSemanticToken(tk, string(protocol.SemanticTokenType), nil),
				)
			case *ast.CompoundIdentNode:
				for _, component := range typ.Components {
					tk := component.Token()
					p.tokenMap[tk] = append(
						p.tokenMap[tk],
						p.createSemanticToken(tk, string(protocol.SemanticTokenType), nil),
					)
				}
			}
			nameTk := field.Name.Token()
			p.tokenMap[nameTk] = append(
				p.tokenMap[nameTk],
				p.createSemanticToken(nameTk, string(protocol.SemanticTokenProperty), nil),
			)
			return nil
		},
		DoVisitEnumNode: func(enum *ast.EnumNode) error {
			tk := enum.Name.Token()
			p.tokenMap[tk] = append(
				p.tokenMap[tk],
				p.createSemanticToken(tk, string(protocol.SemanticTokenType), nil),
			)
			return nil
		},
		DoVisitEnumValueNode: func(value *ast.EnumValueNode) error {
			tk := value.Name.Token()
			p.tokenMap[tk] = append(
				p.tokenMap[tk],
				p.createSemanticToken(tk, string(protocol.SemanticTokenEnumMember), nil),
			)
			return nil
		},
		DoVisitServiceNode: func(svc *ast.ServiceNode) error {
			tk := svc.Name.Token()
			p.tokenMap[tk] = append(
				p.tokenMap[tk],
				p.createSemanticToken(tk, string(protocol.SemanticTokenType), nil),
			)
			return nil
		},
		DoVisitRPCNode: func(rpc *ast.RPCNode) error {
			nameTk := rpc.Name.Token()
			inputTk := rpc.Input.MessageType.Start()
			outputTk := rpc.Output.MessageType.Start()

			p.tokenMap[nameTk] = append(
				p.tokenMap[nameTk],
				p.createSemanticToken(nameTk, string(protocol.SemanticTokenMethod), nil),
			)
			p.tokenMap[inputTk] = append(
				p.tokenMap[inputTk],
				p.createSemanticToken(inputTk, string(protocol.SemanticTokenType), nil),
			)
			p.tokenMap[outputTk] = append(
				p.tokenMap[outputTk],
				p.createSemanticToken(outputTk, string(protocol.SemanticTokenType), nil),
			)
			return nil
		},
		DoVisitOptionNode: func(opt *ast.OptionNode) error {
			parts := make([]string, 0, len(opt.Name.Parts))
			for _, part := range opt.Name.Parts {
				parts = append(parts, string(part.Name.AsIdentifier()))
				switch n := part.Name.(type) {
				case *ast.CompoundIdentNode:
					for _, comp := range n.Components {
						tk := comp.Token()
						p.tokenMap[tk] = append(
							p.tokenMap[tk],
							p.createSemanticToken(tk, string(protocol.SemanticTokenNamespace), nil),
						)
					}
				}
			}
			optName := strings.TrimPrefix(strings.Join(parts, "."), ".")
			if strings.HasPrefix(optName, "grpc.federation.") {
				var optType FederationOptionType
				switch {
				case strings.Contains(optName, ".service"):
					optType = ServiceOption
				case strings.Contains(optName, ".method"):
					optType = MethodOption
				case strings.Contains(optName, ".message"):
					optType = MessageOption
				case strings.Contains(optName, ".field"):
					optType = FieldOption
				case strings.Contains(optName, ".enum_value"):
					optType = EnumValueOption
				case strings.Contains(optName, ".enum"):
					optType = EnumOption
				}
				ctx := newSemanticTokenProviderContext(optType)
				switch n := opt.Val.(type) {
				case *ast.StringLiteralNode:
					lastPart := opt.Name.Parts[len(opt.Name.Parts)-1]
					tk := lastPart.Name.Start()
					p.tokenMap[tk] = append(
						p.tokenMap[tk],
						p.createSemanticToken(tk, string(protocol.SemanticTokenProperty), nil),
					)
					switch {
					case strings.HasSuffix(optName, "by"):
						if err := p.setCELSemanticTokens(n); err != nil {
							p.logger.Error(err.Error())
						}
					case strings.HasSuffix(optName, "alias"):
						p.setAliasToken(ctx, n)
					}
				case *ast.CompoundStringLiteralNode:
					switch {
					case strings.HasSuffix(optName, "by"):
						if err := p.setCELSemanticTokens(n); err != nil {
							p.logger.Error(err.Error())
						}
					}
				default:
					p.setFederationOptionTokens(ctx, opt.Val)
				}
			}
			return nil
		},
	})

	var (
		line = uint32(1)
		col  = uint32(1)
		data []uint32
	)
	_ = ast.Walk(p.tree, &ast.SimpleVisitor{
		DoVisitTerminalNode: func(n ast.TerminalNode) error {
			curToken := n.Token()
			info := p.tree.TokenInfo(curToken)

			comments := info.LeadingComments()
			for i := 0; i < comments.Len(); i++ {
				comment := comments.Index(i)
				pos := comment.Start()
				for _, text := range strings.Split(comment.RawText(), "\n") {
					diffLine, diffCol := p.calcLineAndCol(pos, line, col)
					line = uint32(pos.Line)
					col = uint32(pos.Col)
					tokenNum := p.tokenTypeMap[string(protocol.SemanticTokenComment)]
					data = append(data, diffLine, diffCol, uint32(len(text)), tokenNum, 0)
					pos = ast.SourcePos{
						Line: pos.Line + 1,
						Col:  pos.Col,
					}
				}
			}

			tokens, exists := p.tokenMap[n.Token()]
			if !exists {
				switch nn := n.(type) {
				case *ast.IdentNode:
					ident := nn.AsIdentifier()
					var tokenType string
					if ident == "true" || ident == "false" {
						tokenType = string(protocol.SemanticTokenKeyword)
					} else {
						tokenType = string(protocol.SemanticTokenVariable)
					}
					tokens = append(tokens, p.createSemanticToken(n.Token(), tokenType, nil))
				case *ast.KeywordNode:
					tokens = append(tokens, p.createSemanticToken(n.Token(), string(protocol.SemanticTokenKeyword), nil))
				case *ast.UintLiteralNode:
					tokens = append(tokens, p.createSemanticToken(n.Token(), string(protocol.SemanticTokenNumber), nil))
				case *ast.StringLiteralNode:
					tokens = append(tokens, p.createSemanticToken(n.Token(), string(protocol.SemanticTokenString), nil))
				default:
					tokens = append(tokens, p.createSemanticToken(n.Token(), string(protocol.SemanticTokenOperator), nil))
				}
			}

			for _, token := range tokens {
				diffLine, diffCol := p.calcLineAndCol(ast.SourcePos{
					Line: int(token.Line),
					Col:  int(token.Col),
				}, line, col)

				line = token.Line
				col = token.Col

				tokenNum := p.tokenTypeMap[token.Type]
				var mod uint32
				for _, modifier := range token.Modifiers {
					mod |= p.tokenModifierMap[modifier]
				}
				data = append(data, diffLine, diffCol, uint32(len([]rune(token.Text))), tokenNum, mod)
			}

			{
				comments := info.TrailingComments()
				for i := 0; i < comments.Len(); i++ {
					comment := comments.Index(i)
					pos := comment.Start()
					for _, text := range strings.Split(comment.RawText(), "\n") {
						diffLine, diffCol := p.calcLineAndCol(pos, line, col)
						line = uint32(pos.Line)
						col = uint32(pos.Col)
						tokenNum := p.tokenTypeMap[string(protocol.SemanticTokenComment)]
						data = append(data, diffLine, diffCol, uint32(len(text)), tokenNum, 0)
						pos = ast.SourcePos{
							Line: pos.Line + 1,
							Col:  pos.Col,
						}
					}
				}
			}
			return nil
		},
	})

	return &protocol.SemanticTokens{Data: data}
}

func (p *SemanticTokenProvider) calcLineAndCol(pos ast.SourcePos, line, col uint32) (uint32, uint32) {
	if uint32(pos.Line) == line {
		return 0, uint32(pos.Col) - col
	}
	return uint32(pos.Line) - line, uint32(pos.Col) - 1
}

type FederationOptionType int

const (
	ServiceOption FederationOptionType = iota
	MethodOption
	MessageOption
	FieldOption
	EnumOption
	EnumValueOption
)

type SemanticTokenProviderContext struct {
	optionType FederationOptionType
	message    bool
	enum       bool
}

func newSemanticTokenProviderContext(optType FederationOptionType) *SemanticTokenProviderContext {
	return &SemanticTokenProviderContext{
		optionType: optType,
	}
}

func (c *SemanticTokenProviderContext) clone() *SemanticTokenProviderContext {
	ctx := new(SemanticTokenProviderContext)
	ctx.optionType = c.optionType
	ctx.message = c.message
	ctx.enum = c.enum
	return ctx
}

func (c *SemanticTokenProviderContext) withMessage() *SemanticTokenProviderContext {
	ctx := c.clone()
	ctx.message = true
	return ctx
}

func (c *SemanticTokenProviderContext) withEnum() *SemanticTokenProviderContext {
	ctx := c.clone()
	ctx.enum = true
	return ctx
}

func (p *SemanticTokenProvider) setFederationOptionTokens(ctx *SemanticTokenProviderContext, node ast.Node) {
	switch n := node.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			tk := elem.Name.Name.Start()
			p.tokenMap[tk] = append(
				p.tokenMap[tk],
				p.createSemanticToken(
					tk,
					string(protocol.SemanticTokenProperty),
					[]string{
						string(protocol.SemanticTokenModifierDeclaration),
						string(protocol.SemanticTokenModifierDefaultLibrary),
					},
				),
			)

			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "enum":
				p.setFederationOptionTokens(ctx.withEnum(), elem.Val)
			case "message":
				if _, ok := elem.Val.(*ast.MessageLiteralNode); ok {
					p.setFederationOptionTokens(ctx.withMessage(), elem.Val)
				} else {
					if err := p.setCELSemanticTokens(elem.Val); err != nil {
						p.logger.Error(err.Error())
					}
				}
			case "name":
				p.setNameToken(ctx, elem.Val)
			case "if", "by", "inline", "src", "type", "subject", "description":
				if err := p.setCELSemanticTokens(elem.Val); err != nil {
					p.logger.Error(err.Error())
				}
			case "alias":
				p.setAliasToken(ctx, elem.Val)
			case "method":
				p.setMethodToken(ctx, elem.Val)
			case "field":
				p.setFieldToken(ctx, elem.Val)
			case "locale":
				p.setLocaleToken(ctx, elem.Val)
			default:
				p.setFederationOptionTokens(ctx, elem.Val)
			}
		}
	case *ast.ArrayLiteralNode:
		for _, elem := range n.Elements {
			p.setFederationOptionTokens(ctx, elem)
		}
	}
}

func (p *SemanticTokenProvider) setNameToken(ctx *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	tokenType := string(protocol.SemanticTokenVariable)
	if ctx.message || ctx.enum {
		tokenType = string(protocol.SemanticTokenType)
	}
	p.tokenMap[tk] = append(p.tokenMap[tk], []*SemanticToken{
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col) + 1,
			Text: name,
			Type: tokenType,
			Modifiers: []string{
				string(protocol.SemanticTokenModifierDefinition),
				string(protocol.SemanticTokenModifierReadonly),
			},
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + len(text) - 1),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
	}...)
	ctx.message = false
	ctx.enum = false
}

func (p *SemanticTokenProvider) setLocaleToken(_ *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	tokenType := string(protocol.SemanticTokenVariable)
	p.tokenMap[tk] = append(p.tokenMap[tk], []*SemanticToken{
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col) + 1,
			Text: name,
			Type: tokenType,
			Modifiers: []string{
				string(protocol.SemanticTokenModifierDefinition),
				string(protocol.SemanticTokenModifierReadonly),
			},
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + len(text) - 1),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
	}...)
}

func (p *SemanticTokenProvider) setMethodToken(_ *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], []*SemanticToken{
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + 1),
			Text: name,
			Type: string(protocol.SemanticTokenMethod),
			Modifiers: []string{
				string(protocol.SemanticTokenModifierReadonly),
			},
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + len(text) - 1),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
	}...)
}

func (p *SemanticTokenProvider) setFieldToken(_ *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], []*SemanticToken{
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + 1),
			Text: name,
			Type: string(protocol.SemanticTokenVariable),
		},
		{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col + len(text) - 1),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		},
	}...)
}

func (p *SemanticTokenProvider) setAliasToken(ctx *SemanticTokenProviderContext, val ast.ValueNode) {
	var tokenType string
	switch ctx.optionType {
	case MessageOption:
		tokenType = string(protocol.SemanticTokenType)
	case FieldOption:
		tokenType = string(protocol.SemanticTokenVariable)
	case EnumOption:
		tokenType = string(protocol.SemanticTokenType)
	case EnumValueOption:
		tokenType = string(protocol.SemanticTokenEnumMember)
	}

	addToken := func(val *ast.StringLiteralNode) {
		tk := val.Start()
		info := p.tree.TokenInfo(tk)
		pos := info.Start()
		text := info.RawText()
		name := strings.Trim(text, `"`)
		p.tokenMap[tk] = append(p.tokenMap[tk], []*SemanticToken{
			{
				Line: uint32(pos.Line),
				Col:  uint32(pos.Col),
				Text: `"`,
				Type: string(protocol.SemanticTokenString),
			},
			{
				Line: uint32(pos.Line),
				Col:  uint32(pos.Col + 1),
				Text: name,
				Type: tokenType,
				Modifiers: []string{
					string(protocol.SemanticTokenModifierReadonly),
				},
			},
			{
				Line: uint32(pos.Line),
				Col:  uint32(pos.Col + len(text) - 1),
				Text: `"`,
				Type: string(protocol.SemanticTokenString),
			},
		}...)
	}

	switch n := val.(type) {
	case *ast.StringLiteralNode:
		addToken(n)
	case *ast.ArrayLiteralNode:
		for _, elem := range n.Elements {
			val, ok := elem.(*ast.StringLiteralNode)
			if !ok {
				continue
			}
			addToken(val)
		}
	}
}

func (p *SemanticTokenProvider) setCELSemanticTokens(value ast.ValueNode) error {
	var strs []*ast.StringLiteralNode
	switch v := value.(type) {
	case *ast.StringLiteralNode:
		strs = append(strs, v)
	case *ast.CompoundStringLiteralNode:
		for _, child := range v.Children() {
			str, ok := child.(*ast.StringLiteralNode)
			if !ok {
				continue
			}
			strs = append(strs, str)
		}
	default:
		return fmt.Errorf("unexpected value node type for CEL expression. type is %T", value)
	}
	var (
		allTexts    []string
		lineOffsets []LineOffset
		offset      int
	)
	for _, str := range strs {
		tk := str.Token()
		info := p.tree.TokenInfo(tk)
		pos := info.Start()
		text := info.RawText()
		celText := strings.Trim(strings.Replace(text, "$", "a", -1), `"`)
		allTexts = append(allTexts, celText)

		celTextLen := len([]rune(celText))

		lineOffsets = append(lineOffsets, LineOffset{
			OriginalLine: uint32(pos.Line),
			StartOffset:  offset,
			EndOffset:    offset + celTextLen,
		})
		offset += celTextLen
	}
	src := strings.Join(allTexts, "")
	celParser, err := parser.NewParser(parser.EnableOptionalSyntax(true))
	if err != nil {
		return fmt.Errorf("failed to create CEL parser: %w", err)
	}
	celAST, errs := celParser.Parse(common.NewTextSource(src))
	if len(errs.GetErrors()) != 0 {
		return fmt.Errorf("failed to parse %s: %s", src, errs.ToDisplayString())
	}
	provider := newCELSemanticTokenProvider(celAST, lineOffsets, src)
	lineNumToTokens := provider.SemanticTokens()

	for _, value := range strs {
		tk := value.Token()
		info := p.tree.TokenInfo(tk)
		pos := info.Start()
		text := info.RawText()
		originalLine := uint32(pos.Line)
		tokens := lineNumToTokens[originalLine]

		p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		})
		lineEndCol := uint32(pos.Col + len([]rune(text)) - 1)
		if len(tokens) == 0 {
			p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
				Line: uint32(pos.Line),
				Col:  uint32(pos.Col + 1),
				Text: strings.Trim(text, `"`),
				Type: string(protocol.SemanticTokenString),
			})
			continue
		}

		for _, token := range tokens {
			token.Col += uint32(pos.Col)
			p.tokenMap[tk] = append(p.tokenMap[tk], token)
		}
		p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
			Line: uint32(pos.Line),
			Col:  lineEndCol,
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		})
	}
	return nil
}

type CELSemanticTokenProvider struct {
	lineNumToTokens map[uint32][]*SemanticToken
	tree            *celast.AST
	info            *celast.SourceInfo
	lineOffsets     []LineOffset
	fullText        string // Complete concatenated string
}

type LineOffset struct {
	OriginalLine uint32
	StartOffset  int // Character offset (cumulative position)
	EndOffset    int // Character offset (cumulative position)
}

func newCELSemanticTokenProvider(tree *celast.AST, lineOffsets []LineOffset, fullText string) *CELSemanticTokenProvider {
	return &CELSemanticTokenProvider{
		lineNumToTokens: make(map[uint32][]*SemanticToken),
		tree:            tree,
		info:            tree.SourceInfo(),
		lineOffsets:     lineOffsets,
		fullText:        fullText,
	}
}

func (p *CELSemanticTokenProvider) SemanticTokens() map[uint32][]*SemanticToken {
	celast.PostOrderVisit(p.tree.Expr(), p)

	for _, tokens := range p.lineNumToTokens {
		sort.Slice(tokens, func(i, j int) bool {
			if tokens[i].Col == tokens[j].Col {
				// For tokens at the same column position, use stable ordering
				// Use text length as secondary sort key for consistency
				if len(tokens[i].Text) != len(tokens[j].Text) {
					return len(tokens[i].Text) < len(tokens[j].Text)
				}
				// If same length, use lexicographic ordering for consistency
				return tokens[i].Text < tokens[j].Text
			}
			return tokens[i].Col < tokens[j].Col
		})
	}
	return p.lineNumToTokens
}

func (p *CELSemanticTokenProvider) mapToOriginalPosition(pos common.Location) (uint32, uint32) {
	col := pos.Column()
	if len(p.lineOffsets) == 0 {
		// CEL parser reports correct field position (2) for SelectKind,
		// so adding +1 for 1-based conversion should give correct position (3)
		return uint32(pos.Line()), uint32(col + 1)
	}

	// Calculate which original line this rune position belongs to and the position within that line
	for _, lineOffset := range p.lineOffsets {
		if col >= lineOffset.StartOffset && col < lineOffset.EndOffset {
			// Calculate column position within this line
			colFromCurLine := col - lineOffset.StartOffset

			// Return 1-based column position
			return lineOffset.OriginalLine, uint32(colFromCurLine + 1)
		}
	}

	// Default for out-of-range cases
	return uint32(1), uint32(col + 1)
}

// addStringTokenWithSplitCheck adds a string token and automatically splits it across line boundaries.
func (p *CELSemanticTokenProvider) addStringTokenWithSplitCheck(line uint32, col uint32, text string, tokenType string) {
	if len(p.lineOffsets) == 0 {
		// Normal processing when LineOffsets are not available
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
			Text: text,
			Type: tokenType,
		})
		return
	}

	textRunes := []rune(text)
	textLen := len(textRunes)

	// Find LineOffset for current line
	var curLineOffset *LineOffset
	for i, lineOffset := range p.lineOffsets {
		if lineOffset.OriginalLine == line {
			curLineOffset = &p.lineOffsets[i]
			break
		}
	}

	if curLineOffset == nil {
		// Normal processing when LineOffset is not found
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
			Text: text,
			Type: tokenType,
		})
		return
	}

	// Calculate actual length of each line from LineOffset
	actualLineCount := curLineOffset.EndOffset - curLineOffset.StartOffset
	availableSpace := actualLineCount - int(col) + 1

	if availableSpace <= 0 || textLen <= availableSpace {
		// When string fits within the line
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
			Text: text,
			Type: tokenType,
		})
		return
	}

	// Split when string exceeds line
	firstPart := string(textRunes[:availableSpace])
	secondPart := string(textRunes[availableSpace:])

	// First part to current line
	p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
		Line: line,
		Col:  col,
		Text: firstPart,
		Type: tokenType,
	})

	// Second part to next line
	for _, nextLineOffset := range p.lineOffsets {
		if nextLineOffset.OriginalLine > line {
			p.lineNumToTokens[nextLineOffset.OriginalLine] = append(p.lineNumToTokens[nextLineOffset.OriginalLine], &SemanticToken{
				Line: nextLineOffset.OriginalLine,
				Col:  1,
				Text: secondPart,
				Type: tokenType,
			})
			return
		}
	}
}

func (p *CELSemanticTokenProvider) VisitExpr(expr celast.Expr) {
	id := expr.ID()
	start := p.info.GetStartLocation(id)
	originalLine, col := p.mapToOriginalPosition(start)
	line := originalLine

	switch expr.Kind() {
	case celast.CallKind:
		fnName := expr.AsCall().FunctionName()
		if strings.HasPrefix(fnName, "_") {
			opName := strings.Trim(fnName, "_")
			if strings.Contains(opName, "_") {
				fnName = "*" // operation type. this name is dummy text.
			} else {
				fnName = opName
			}
		}
		// Don't adjust column for function names to maintain correct ordering
		// The CEL parser position should be used as-is for consistency with other token types

		// Calculate actual position of function name
		adjustedCol := col
		originalFnName := expr.AsCall().FunctionName()

		// Adjust position only for regular function names (not operators or dummy tokens)
		if !strings.HasPrefix(originalFnName, "_") && fnName != "*" {
			// CEL parser reports position of '(' for CallKind,
			// so function name is before it
			fnNameRuneLen := len([]rune(fnName))
			if col >= uint32(fnNameRuneLen) {
				adjustedCol = col - uint32(fnNameRuneLen)
			}
		}

		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  adjustedCol,
			Text: fnName,
			Type: string(protocol.SemanticTokenMethod),
		})
	case celast.ComprehensionKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
		})
	case celast.IdentKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line:      line,
			Col:       col,
			Text:      expr.AsIdent(),
			Type:      string(protocol.SemanticTokenVariable),
			Modifiers: []string{string(protocol.SemanticTokenModifierReadonly)},
		})
	case celast.ListKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
		})
	case celast.LiteralKind:
		lit := expr.AsLiteral()
		text := fmt.Sprint(lit.Value())
		if lit.Type() == types.StringType {
			text = fmt.Sprintf(`'%s'`, text)
			// Check if string length may exceed line boundaries
			p.addStringTokenWithSplitCheck(line, col, text, string(protocol.SemanticTokenKeyword))
		} else {
			if rng, ok := p.info.GetOffsetRange(id); ok {
				text := []rune(p.fullText)[rng.Start:rng.Stop]
				p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
					Line: line,
					Col:  col,
					Text: string(text),
					Type: string(protocol.SemanticTokenKeyword),
				})
			} else {
				p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
					Line: line,
					Col:  col,
					Text: text,
					Type: string(protocol.SemanticTokenKeyword),
				})
			}
		}
	case celast.MapKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
		})
	case celast.SelectKind:
		// For SelectKind, CEL parser reports position of dot operator,
		// so actual position of field name needs +1
		fieldName := expr.AsSelect().FieldName()
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line:      line,
			Col:       col + 1, // Position after dot operator
			Text:      fieldName,
			Type:      string(protocol.SemanticTokenVariable),
			Modifiers: []string{string(protocol.SemanticTokenModifierReadonly)},
		})
	case celast.StructKind:
		st := expr.AsStruct()
		// start is the left brace character's position.
		// e.g.) typename{field: value}
		//               ^
		start := col - uint32(len([]rune(st.TypeName())))
		if start < 1 {
			start = 1
		}
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  start,
			Text: st.TypeName(),
			Type: string(protocol.SemanticTokenStruct),
		})
		for _, field := range st.Fields() {
			sf := field.AsStructField()
			id := field.ID()
			// start is the colon character's position
			// e.g.) typename{field: value}
			//                     ^
			start := p.info.GetStartLocation(id)
			line, col := p.mapToOriginalPosition(start)
			colPos := col - uint32(len([]rune(sf.Name())))
			p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
				Line: line,
				Col:  colPos,
				Text: sf.Name(),
				Type: string(protocol.SemanticTokenVariable),
			})
		}
	default:
	}
}

func (p *CELSemanticTokenProvider) VisitEntryExpr(expr celast.EntryExpr) {
	// EntryExpr is unsupported.
}

func (h *Handler) semanticTokensFull(params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	file, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	return NewSemanticTokenProvider(h.logger, file, h.tokenTypeMap, h.tokenModifierMap).SemanticTokens(), nil
}
