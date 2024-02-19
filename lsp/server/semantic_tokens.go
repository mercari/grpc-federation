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

	"github.com/mercari/grpc-federation/source"
)

type SemanticTokenProvider struct {
	logger           *slog.Logger
	tokenMap         map[ast.Token][]*SemanticToken
	file             *source.File
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

func NewSemanticTokenProvider(logger *slog.Logger, file *source.File, tokenTypeMap map[string]uint32, tokenModifierMap map[string]uint32) *SemanticTokenProvider {
	return &SemanticTokenProvider{
		logger:           logger,
		tokenMap:         make(map[ast.Token][]*SemanticToken),
		file:             file,
		tree:             file.AST(),
		tokenTypeMap:     tokenTypeMap,
		tokenModifierMap: tokenModifierMap,
	}
}

func (p *SemanticTokenProvider) createSemanticToken(tk ast.Token, tokenType string, modifiers []string) *SemanticToken {
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	return &SemanticToken{
		Line:      uint32(pos.Line),
		Col:       uint32(pos.Col),
		Text:      info.RawText(),
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
				if str, ok := opt.Val.(*ast.StringLiteralNode); ok {
					lastPart := opt.Name.Parts[len(opt.Name.Parts)-1]
					tk := lastPart.Name.Start()
					p.tokenMap[tk] = append(
						p.tokenMap[tk],
						p.createSemanticToken(tk, string(protocol.SemanticTokenProperty), nil),
					)
					switch {
					case strings.HasSuffix(optName, "by"):
						if err := p.setCELSemanticTokens(str); err != nil {
							p.logger.Error(err.Error())
						}
					case strings.HasSuffix(optName, "alias"):
						p.setAliasToken(ctx, str)
					}
				} else {
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
				data = append(data, diffLine, diffCol, uint32(len(token.Text)), tokenNum, mod)
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
	return ctx
}

func (c *SemanticTokenProviderContext) withMessage() *SemanticTokenProviderContext {
	ctx := c.clone()
	ctx.message = true
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
			case "message":
				p.setFederationOptionTokens(ctx.withMessage(), elem.Val)
			case "name":
				p.setNameToken(ctx, elem.Val)
			case "by", "inline", "src":
				if err := p.setCELSemanticTokens(elem.Val); err != nil {
					p.logger.Error(err.Error())
				}
			case "alias":
				p.setAliasToken(ctx, elem.Val)
			case "method":
				p.setMethodToken(ctx, elem.Val)
			case "enum":
				p.setEnumToken(ctx, elem.Val)
			case "field":
				p.setFieldToken(ctx, elem.Val)
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
	if ctx.message {
		tokenType = string(protocol.SemanticTokenType)
	}
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col) + 1,
		Text: name,
		Type: tokenType,
		Modifiers: []string{
			string(protocol.SemanticTokenModifierDefinition),
			string(protocol.SemanticTokenModifierReadonly),
		},
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + len(text) - 1),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	ctx.message = false
}

func (p *SemanticTokenProvider) setMethodToken(ctx *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + 1),
		Text: name,
		Type: string(protocol.SemanticTokenMethod),
		Modifiers: []string{
			string(protocol.SemanticTokenModifierReadonly),
		},
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + len(text) - 1),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
}

func (p *SemanticTokenProvider) setEnumToken(ctx *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + 1),
		Text: name,
		Type: string(protocol.SemanticTokenEnumMember),
		Modifiers: []string{
			string(protocol.SemanticTokenModifierReadonly),
		},
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + len(text) - 1),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
}

func (p *SemanticTokenProvider) setFieldToken(ctx *SemanticTokenProviderContext, val ast.ValueNode) {
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + 1),
		Text: name,
		Type: string(protocol.SemanticTokenVariable),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + len(text) - 1),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
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
	tk := val.Start()
	info := p.tree.TokenInfo(tk)
	pos := info.Start()
	text := info.RawText()
	name := strings.Trim(text, `"`)
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + 1),
		Text: name,
		Type: tokenType,
		Modifiers: []string{
			string(protocol.SemanticTokenModifierReadonly),
		},
	})
	p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
		Line: uint32(pos.Line),
		Col:  uint32(pos.Col + len(text) - 1),
		Text: `"`,
		Type: string(protocol.SemanticTokenString),
	})
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
	var allTexts []string
	for _, str := range strs {
		tk := str.Token()
		info := p.tree.TokenInfo(tk)
		text := info.RawText()
		allTexts = append(allTexts, strings.Trim(strings.Replace(text, "$", "a", -1), `"`))
	}
	src := strings.Join(allTexts, "\n")
	celParser, err := parser.NewParser(parser.EnableOptionalSyntax(true))
	if err != nil {
		return fmt.Errorf("failed to create CEL parser: %w", err)
	}
	celAST, errs := celParser.Parse(common.NewTextSource(src))
	if len(errs.GetErrors()) != 0 {
		return fmt.Errorf("failed to parse %s: %s", src, errs.ToDisplayString())
	}
	provider := newCELSemanticTokenProvider(celAST)
	lineNumToTokens := provider.SemanticTokens()

	var remainTokens []*SemanticToken
	for idx, value := range strs {
		tk := value.Token()
		info := p.tree.TokenInfo(tk)
		pos := info.Start()
		text := info.RawText()
		tokens := lineNumToTokens[uint32(idx+1)] // line number is 1-based.
		p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
			Line: uint32(pos.Line),
			Col:  uint32(pos.Col),
			Text: `"`,
			Type: string(protocol.SemanticTokenString),
		})
		for _, remainToken := range remainTokens {
			remainToken.Line = uint32(pos.Line)
			// skip start quotation character.
			// TODO: column calculation is not supported when there are multiple remainTokens.
			remainToken.Col = uint32(pos.Col) + 1
			p.tokenMap[tk] = append(p.tokenMap[tk], remainToken)
		}
		remainTokens = nil
		lineEndCol := uint32(pos.Col + len(text) - 1)
		if len(tokens) == 0 {
			p.tokenMap[tk] = append(p.tokenMap[tk], &SemanticToken{
				Line: uint32(pos.Line),
				Col:  uint32(pos.Col + 1),
				Text: strings.Trim(text, `"`),
				Type: string(protocol.SemanticTokenString),
			})
		} else {
			for _, token := range tokens {
				token.Line = uint32(pos.Line)
				token.Col += uint32(pos.Col)
				if token.Col > lineEndCol {
					// This token should be the first token in the next line.
					remainTokens = append(remainTokens, token)
				} else {
					p.tokenMap[tk] = append(p.tokenMap[tk], token)
				}
			}
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
}

func newCELSemanticTokenProvider(tree *celast.AST) *CELSemanticTokenProvider {
	return &CELSemanticTokenProvider{
		lineNumToTokens: make(map[uint32][]*SemanticToken),
		tree:            tree,
		info:            tree.SourceInfo(),
	}
}

func (p *CELSemanticTokenProvider) SemanticTokens() map[uint32][]*SemanticToken {
	celast.PostOrderVisit(p.tree.Expr(), p)
	for _, tokens := range p.lineNumToTokens {
		sort.Slice(tokens, func(i, j int) bool {
			return tokens[i].Col < tokens[j].Col
		})
	}
	return p.lineNumToTokens
}

func (p *CELSemanticTokenProvider) VisitExpr(expr celast.Expr) {
	id := expr.ID()
	start := p.info.GetStartLocation(id)
	line := uint32(start.Line())
	col := uint32(start.Column() + 1)
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
		} else {
			col -= uint32(len(fnName))
		}
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
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
		}
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
			Text: text,
			Type: string(protocol.SemanticTokenKeyword),
		})
	case celast.MapKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
		})
	case celast.SelectKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line:      line,
			Col:       col + 1,
			Text:      expr.AsSelect().FieldName(),
			Type:      string(protocol.SemanticTokenVariable),
			Modifiers: []string{string(protocol.SemanticTokenModifierReadonly)},
		})
	case celast.StructKind:
		p.lineNumToTokens[line] = append(p.lineNumToTokens[line], &SemanticToken{
			Line: line,
			Col:  col,
			Text: expr.AsStruct().TypeName(),
			Type: string(protocol.SemanticTokenStruct),
		})
	default:
	}
}

func (p *CELSemanticTokenProvider) VisitEntryExpr(expr celast.EntryExpr) {
	// EntryExpr is unsupported.
}

func (h *Handler) semanticTokensFull(params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	path, content, err := h.getFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		return nil, nil
	}
	return NewSemanticTokenProvider(h.logger, file, h.tokenTypeMap, h.tokenModifierMap).SemanticTokens(), nil
}
