package source

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

const (
	serviceOptionName   = "grpc.federation.service"
	methodOptionName    = "grpc.federation.method"
	msgOptionName       = "grpc.federation.message"
	fieldOptionName     = "grpc.federation.field"
	enumOptionName      = "grpc.federation.enum"
	enumValueOptionName = "grpc.federation.enum_value"
)

var _ io.ReadCloser = new(File)

type File struct {
	path     string
	content  []byte
	buf      *bytes.Buffer
	fileNode *ast.FileNode
}

type ignoreErrorReporter struct{}

func (*ignoreErrorReporter) Error(pos reporter.ErrorWithPos) error { return nil }
func (*ignoreErrorReporter) Warning(pos reporter.ErrorWithPos)     {}

func NewFile(path string, content []byte) (*File, error) {
	fileName := filepath.Base(path)

	fileNode, err := func() (f *ast.FileNode, e error) {
		defer func() {
			if err := recover(); err != nil {
				e = fmt.Errorf("failed to parse %s: %v", path, err)
			}
		}()
		return parser.Parse(
			fileName,
			bytes.NewBuffer(content),
			reporter.NewHandler(&ignoreErrorReporter{}),
		)
	}()
	// If fileNode is nil, an error is returned.
	// otherwise, no error is returned even if a syntax error occurs.
	if fileNode == nil {
		return nil, err
	}
	return &File{
		path:     path,
		content:  content,
		buf:      bytes.NewBuffer(content),
		fileNode: fileNode,
	}, nil
}

func (f *File) Read(b []byte) (int, error) {
	return f.buf.Read(b)
}

func (f *File) Close() error {
	return nil
}

func (f *File) AST() *ast.FileNode {
	return f.fileNode
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Content() []byte {
	return f.content
}

func (f *File) Imports() []string {
	var imports []string
	for _, decl := range f.fileNode.Decls {
		switch declNode := decl.(type) {
		case *ast.ImportNode:
			imports = append(imports, declNode.Name.AsString())
		}
	}
	return imports
}

// ImportsByImportRule returns import path defined in grpc.federation.file.import rule.
func (f *File) ImportsByImportRule() []string {
	var imports []string
	for _, decl := range f.fileNode.Decls {
		switch declNode := decl.(type) {
		case *ast.OptionNode:
			vals := f.optionValuesByImportRule(declNode)
			for _, val := range vals {
				if s, ok := val.Value().(string); ok {
					imports = append(imports, s)
				}
			}
		}
	}
	return imports
}

type findContext struct {
	fileName        string
	message         *Message
	nestedMessage   *Message
	enum            *Enum
	enumValue       *EnumValue
	service         *Service
	method          *Method
	field           *Field
	fieldOption     *FieldOption
	messageOption   *MessageOption
	enumOption      *EnumOption
	enumValueOption *EnumValueOption
	def             *VariableDefinitionOption
}

func (c *findContext) child() *findContext {
	msg := c.message.Clone()
	var nestedMsg *Message
	if c.nestedMessage != nil {
		nestedMsg = msg.LastNestedMessage()
	}
	return &findContext{
		fileName:        c.fileName,
		message:         msg,
		nestedMessage:   nestedMsg,
		enum:            c.enum.Clone(),
		enumValue:       c.enumValue.Clone(),
		service:         c.service.Clone(),
		method:          c.method.Clone(),
		field:           c.field.Clone(),
		fieldOption:     c.fieldOption.Clone(),
		messageOption:   c.messageOption.Clone(),
		enumOption:      c.enumOption.Clone(),
		enumValueOption: c.enumValueOption.Clone(),
		def:             c.def.Clone(),
	}
}

func (f *File) buildLocation(ctx *findContext) *Location {
	loc := &Location{FileName: ctx.fileName}
	if ctx.service != nil {
		loc.Service = ctx.service
		if ctx.method != nil {
			loc.Service.Method = ctx.method
		}
	}
	var msg *Message
	if ctx.message != nil {
		loc.Message = ctx.message
		msg = ctx.message
		if ctx.nestedMessage != nil {
			msg = ctx.nestedMessage
		}
		if ctx.field != nil {
			msg.Field = ctx.field
			if ctx.fieldOption != nil {
				msg.Field.Option = ctx.fieldOption
			}
		}
		if ctx.messageOption != nil {
			msg.Option = ctx.messageOption
		}
		if ctx.def != nil {
			msg.Option.Def = ctx.def
		}
	}
	if ctx.enum != nil {
		enum := ctx.enum
		if msg != nil {
			msg.Enum = enum
		} else {
			loc.Enum = enum
		}
		if ctx.enumOption != nil {
			enum.Option = ctx.enumOption
		}
		if ctx.enumValue != nil {
			enum.Value = ctx.enumValue
		}
		if ctx.enumValueOption != nil {
			enum.Value.Option = ctx.enumValueOption
		}
	}
	return loc
}

// FindLocationByPos returns the corresponding location information from the position in the source code.
func (f *File) FindLocationByPos(pos Position) *Location {
	ctx := &findContext{fileName: f.Path()}
	for _, decl := range f.fileNode.Decls {
		ctx := ctx.child()
		switch n := decl.(type) {
		case *ast.ImportNode:
			if found := f.findImportByPos(ctx, pos, n); found != nil {
				return found
			}
		case *ast.MessageNode:
			if found := f.findMessageByPos(ctx, pos, n, false); found != nil {
				return found
			}
		case *ast.ServiceNode:
			if found := f.findServiceByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findImportByPos(ctx *findContext, pos Position, node *ast.ImportNode) *Location {
	if f.containsPos(node.Name, pos) {
		return &Location{
			FileName:   ctx.fileName,
			ImportName: node.Name.AsString(),
		}
	}
	return nil
}

func (f *File) findMessageByPos(ctx *findContext, pos Position, node *ast.MessageNode, isNested bool) (loc *Location) {
	msg := &Message{Name: string(node.Name.AsIdentifier())}
	if isNested {
		if ctx.nestedMessage == nil {
			ctx.message.NestedMessage = msg
		} else {
			ctx.nestedMessage.NestedMessage = msg
		}
		ctx.nestedMessage = msg
	} else {
		ctx.message = msg
	}
	for _, decl := range node.MessageBody.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), msgOptionName) {
				continue
			}
			if found := f.findMessageOptionByPos(ctx.child(), pos, n); found != nil {
				return found
			}
		case *ast.FieldNode:
			if found := f.findFieldByPos(ctx, pos, n); found != nil {
				return found
			}
		case *ast.MessageNode:
			if found := f.findMessageByPos(ctx.child(), pos, n, true); found != nil {
				return found
			}
		case *ast.EnumNode:
			if found := f.findEnumByPos(ctx.child(), pos, n); found != nil {
				return found
			}
		case *ast.OneofNode:
			if found := f.findOneofByPos(ctx, pos, n); found != nil {
				return found
			}
		case *ast.MapFieldNode:
		}
	}
	return nil
}

func (f *File) findEnumByPos(ctx *findContext, pos Position, node *ast.EnumNode) *Location {
	ctx.enum = &Enum{Name: string(node.Name.AsIdentifier())}
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), enumOptionName) {
				continue
			}
			if found := f.findEnumOptionByPos(ctx.child(), pos, n); found != nil {
				return found
			}
		case *ast.EnumValueNode:
			if found := f.findEnumValueByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findEnumOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.enumOption = &EnumOption{
						Alias: true,
					}
					return f.buildLocation(ctx)
				}
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				ctx.enumOption = &EnumOption{Alias: true}
				return f.buildLocation(ctx)
			}
		}
	}
	return nil
}

func (f *File) findEnumValueByPos(ctx *findContext, pos Position, node *ast.EnumValueNode) *Location {
	ctx.enumValue = &EnumValue{Value: string(node.Name.AsIdentifier())}
	if node.Options != nil {
		for _, opt := range node.Options.Options {
			if !f.matchOption(f.optionName(opt), enumValueOptionName) {
				continue
			}
			if found := f.findEnumValueOptionByPos(ctx.child(), pos, opt); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findEnumValueOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				switch value := elem.Val.(type) {
				case *ast.StringLiteralNode:
					if f.containsPos(value, pos) {
						ctx.enumValueOption = &EnumValueOption{
							Alias: true,
						}
						return f.buildLocation(ctx)
					}
				case *ast.ArrayLiteralNode:
					for _, elem := range value.Elements {
						str, ok := elem.(*ast.StringLiteralNode)
						if !ok {
							continue
						}
						if f.containsPos(str, pos) {
							ctx.enumValueOption = &EnumValueOption{
								Alias: true,
							}
							return f.buildLocation(ctx)
						}
					}
				}
			case "default":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.enumValueOption = &EnumValueOption{
						Default: true,
					}
					return f.buildLocation(ctx)
				}
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				ctx.enumValueOption = &EnumValueOption{Alias: true}
				return f.buildLocation(ctx)
			}
		}
		if strings.HasSuffix(f.optionName(node), "default") {
			if f.containsPos(n, pos) {
				ctx.enumValueOption = &EnumValueOption{Default: true}
				return f.buildLocation(ctx)
			}
		}
	}
	return nil
}

func (f *File) findOneofByPos(ctx *findContext, pos Position, node *ast.OneofNode) *Location {
	if node == nil {
		return nil
	}
	for _, elem := range node.Decls {
		switch n := elem.(type) {
		case *ast.FieldNode:
			if found := f.findFieldByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findFieldByPos(ctx *findContext, pos Position, node *ast.FieldNode) *Location {
	opts := node.GetOptions()
	fieldName := string(node.Name.AsIdentifier())
	ctx.field = &Field{Name: fieldName}
	if opts != nil {
		for _, opt := range opts.Options {
			if !f.matchOption(f.optionName(opt), fieldOptionName) {
				continue
			}
			if found := f.findFieldOptionByPos(ctx, pos, opt); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node.FldType, pos) {
		ctx.field.Type = true
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findFieldOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "by") {
			if f.containsPos(n, pos) {
				ctx.fieldOption = &FieldOption{By: true}
				return f.buildLocation(ctx)
			}
		}
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "by":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.fieldOption = &FieldOption{By: true}
					return f.buildLocation(ctx)
				}
			}
		}
	}
	return nil
}

func (f *File) findMessageOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		ctx.messageOption = &MessageOption{}
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				if f.containsPos(elem.Val, pos) {
					ctx.messageOption.Alias = true
					return f.buildLocation(ctx)
				}
			case "def":
				if found := f.findDefByPos(ctx, pos, f.getMessageListFromNode(elem.Val)); found != nil {
					return found
				}
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				ctx.messageOption = &MessageOption{Alias: true}
				return f.buildLocation(ctx)
			}
		}
	}
	if f.containsPos(node.Val, pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) getMessageListFromNode(node ast.Node) []*ast.MessageLiteralNode {
	switch value := node.(type) {
	case *ast.MessageLiteralNode:
		return []*ast.MessageLiteralNode{value}
	case *ast.ArrayLiteralNode:
		values := make([]*ast.MessageLiteralNode, 0, len(value.Elements))
		for _, elem := range value.Elements {
			literal, ok := elem.(*ast.MessageLiteralNode)
			if !ok {
				continue
			}
			values = append(values, literal)
		}
		return values
	}
	return nil
}

func (f *File) findDefByPos(ctx *findContext, pos Position, list []*ast.MessageLiteralNode) *Location {
	for idx, node := range list {
		for _, elem := range node.Elements {
			fieldName := elem.Name.Name.AsIdentifier()
			switch fieldName {
			case "name":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx:  idx,
						Name: true,
					}
					return f.buildLocation(ctx)
				}
			case "if":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: idx,
						If:  true,
					}
					return f.buildLocation(ctx)
				}
			case "by":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: idx,
						By:  true,
					}
					return f.buildLocation(ctx)
				}
			case "call":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				if found := f.findCallExprByPos(ctx, idx, pos, value); found != nil {
					return found
				}
			case "message":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				if found := f.findMessageExprByPos(ctx, idx, pos, value); found != nil {
					return found
				}
			}
		}
	}
	if len(list) == 0 {
		return nil
	}
	if f.containsPos(list[0], pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findMessageExprByPos(ctx *findContext, defIdx int, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			if f.containsPos(value, pos) {
				ctx.def = &VariableDefinitionOption{
					Idx:     defIdx,
					Message: &MessageExprOption{Name: true},
				}
				return f.buildLocation(ctx)
			}
		case "args":
			if found := f.findMessageArgumentByPos(ctx, defIdx, pos, f.getMessageListFromNode(elem.Val)); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findMessageArgumentByPos(ctx *findContext, defIdx int, pos Position, list []*ast.MessageLiteralNode) *Location {
	for argIdx, literal := range list {
		for _, arg := range literal.Elements {
			fieldName := arg.Name.Name.AsIdentifier()
			switch fieldName {
			case "name":
				value, ok := arg.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Message: &MessageExprOption{
							Args: &ArgumentOption{
								Idx:  argIdx,
								Name: true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			case "by":
				value, ok := arg.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Message: &MessageExprOption{
							Args: &ArgumentOption{
								Idx: argIdx,
								By:  true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			case "inline":
				value, ok := arg.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Message: &MessageExprOption{
							Args: &ArgumentOption{
								Idx:    argIdx,
								Inline: true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			}
		}
	}
	return nil
}

func (f *File) findCallExprByPos(ctx *findContext, defIdx int, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "method":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			if f.containsPos(value, pos) {
				ctx.def = &VariableDefinitionOption{
					Idx:  defIdx,
					Call: &CallExprOption{Method: true},
				}
				return f.buildLocation(ctx)
			}
		case "request":
			if found := f.findMethodRequestByPos(ctx, defIdx, pos, f.getMessageListFromNode(elem.Val)); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		ctx.def = &VariableDefinitionOption{
			Idx:  defIdx,
			Call: &CallExprOption{},
		}
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findMethodRequestByPos(ctx *findContext, defIdx int, pos Position, list []*ast.MessageLiteralNode) *Location {
	for idx, literal := range list {
		for _, field := range literal.Elements {
			fieldName := field.Name.Name.AsIdentifier()
			switch fieldName {
			case "field":
				value, ok := field.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Call: &CallExprOption{
							Request: &RequestOption{
								Idx:   idx,
								Field: true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			case "by":
				value, ok := field.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Call: &CallExprOption{
							Request: &RequestOption{
								Idx: idx,
								By:  true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			case "if":
				value, ok := field.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.def = &VariableDefinitionOption{
						Idx: defIdx,
						Call: &CallExprOption{
							Request: &RequestOption{
								Idx: idx,
								If:  true,
							},
						},
					}
					return f.buildLocation(ctx)
				}
			}
		}
		if f.containsPos(literal, pos) {
			ctx.def = &VariableDefinitionOption{
				Idx: defIdx,
				Call: &CallExprOption{
					Request: &RequestOption{
						Idx: idx,
					},
				},
			}
			return f.buildLocation(ctx)
		}
	}
	return nil
}

func (f *File) findServiceByPos(ctx *findContext, pos Position, node *ast.ServiceNode) *Location {
	ctx.service = &Service{Name: string(node.Name.AsIdentifier())}
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), serviceOptionName) {
				continue
			}
			if found := f.findServiceOptionByPos(ctx, pos, n); found != nil {
				return found
			}
		case *ast.RPCNode:
			if found := f.findMethodByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findServiceOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "env":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				if found := f.findEnvByPos(ctx, pos, value); found != nil {
					return found
				}
			}
		}
	}
	if f.containsPos(node.Val, pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findEnvByPos(ctx *findContext, pos Position, node *ast.MessageLiteralNode) *Location {
	for idx, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "message":
			if f.containsPos(elem.Val, pos) {
				return f.buildLocation(ctx)
			}
		case "var":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findEnvVarByPos(ctx, idx, pos, value); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findEnvVarByPos(ctx *findContext, _ int, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return f.buildLocation(ctx)
			}
		}
	}
	if f.containsPos(node, pos) {
		return f.buildLocation(ctx)
	}
	return nil
}

func (f *File) findMethodByPos(ctx *findContext, pos Position, node *ast.RPCNode) *Location {
	ctx.method = &Method{Name: string(node.Name.AsIdentifier())}
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), methodOptionName) {
				continue
			}
			if found := f.findMethodOptionByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findMethodOptionByPos(ctx *findContext, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "timeout") {
			if f.containsPos(n, pos) {
				ctx.method.Option = &MethodOption{Timeout: true}
				return f.buildLocation(ctx)
			}
		}
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "timeout":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.method.Option = &MethodOption{Timeout: true}
					return f.buildLocation(ctx)
				}
			}
		}
	}
	return nil
}

func (f *File) containsPos(node ast.Node, pos Position) bool {
	info := f.fileNode.NodeInfo(node)
	startPos := info.Start()
	endPos := info.End()
	if startPos.Line > pos.Line {
		return false
	}
	if startPos.Line == pos.Line {
		if startPos.Col > pos.Col {
			return false
		}
	}
	if endPos.Line < pos.Line {
		return false
	}
	if endPos.Line == pos.Line {
		if endPos.Col < pos.Col {
			return false
		}
	}
	return true
}

func (f *File) optionName(node *ast.OptionNode) string {
	parts := make([]string, 0, len(node.Name.Parts))
	for _, part := range node.Name.Parts {
		parts = append(parts, string(part.Name.AsIdentifier()))
	}
	return strings.Join(parts, ".")
}

func (f *File) optionValuesByImportRule(node *ast.OptionNode) []ast.ValueNode {
	optName := f.optionName(node)
	if !f.matchOption(optName, "grpc.federation.file") {
		return nil
	}

	// option (grpc.federation.file).import = "example.proto";
	if strings.HasSuffix(optName, ".import") {
		return []ast.ValueNode{node.GetValue()}
	}

	// option (grpc.federation.file)= {
	//   import: ["example.proto"]
	// };
	if fieldNodes, ok := node.GetValue().Value().([]*ast.MessageFieldNode); ok {
		var fieldNode *ast.MessageFieldNode
		for _, n := range fieldNodes {
			if n.Name.Name.AsIdentifier() == "import" {
				fieldNode = n
				break
			}
		}
		if fieldNode == nil {
			return nil
		}
		if valueNodes, ok := fieldNode.Val.Value().([]ast.ValueNode); ok {
			return valueNodes
		}
	}
	return nil
}

// NodeInfoByLocation returns information about the node at the position specified by location in the AST of the Protocol Buffers.
func (f *File) NodeInfoByLocation(loc *Location) *ast.NodeInfo {
	if loc.FileName == "" {
		return nil
	}
	if f.fileNode.Name() != filepath.Base(loc.FileName) {
		return nil
	}
	for _, decl := range f.fileNode.Decls {
		switch n := decl.(type) {
		case *ast.ImportNode:
			if loc.ImportName == n.Name.AsString() {
				return f.nodeInfo(n.Name)
			}
		case *ast.OptionNode:
			if f.matchOption(f.optionName(n), "go_package") && loc.GoPackage {
				return f.nodeInfo(n.Val)
			}
			if f.matchOption(f.optionName(n), "grpc.federation.file") {
				vals := f.optionValuesByImportRule(n)
				for _, val := range vals {
					if s, ok := val.Value().(string); ok && s == loc.ImportName {
						return f.nodeInfo(val)
					}
				}
			}
		case *ast.MessageNode:
			if loc.Message != nil {
				if string(n.Name.AsIdentifier()) == loc.Message.Name {
					return f.nodeInfoByMessage(n, loc.Message)
				}
			}
		case *ast.EnumNode:
			if loc.Enum != nil {
				if string(n.Name.AsIdentifier()) == loc.Enum.Name {
					return f.nodeInfoByEnum(n, loc.Enum)
				}
			}
		case *ast.ServiceNode:
			if loc.Service != nil {
				if string(n.Name.AsIdentifier()) == loc.Service.Name {
					return f.nodeInfoByService(n, loc.Service)
				}
			}
		}
	}
	return nil
}

func (f *File) nodeInfoByMessage(node *ast.MessageNode, msg *Message) *ast.NodeInfo {
	if string(node.Name.AsIdentifier()) != msg.Name {
		return nil
	}
	for _, decl := range node.MessageBody.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), msgOptionName) {
				continue
			}
			if msg.Option != nil {
				return f.nodeInfoByMessageOption(n, msg.Option)
			}
		case *ast.FieldNode:
			if msg.Field != nil {
				if string(n.Name.AsIdentifier()) != msg.Field.Name {
					continue
				}
				return f.nodeInfoByField(n, msg.Field)
			}
		case *ast.OneofNode:
			if info := f.nodeInfoByOneof(n, msg); info != nil {
				return info
			}
		case *ast.MessageNode:
			if msg.NestedMessage != nil {
				if info := f.nodeInfoByMessage(n, msg.NestedMessage); info != nil {
					return info
				}
			}
		case *ast.EnumNode:
			if msg.Enum != nil {
				if string(n.Name.AsIdentifier()) == msg.Enum.Name {
					return f.nodeInfoByEnum(n, msg.Enum)
				}
			}
		case *ast.MapFieldNode:
			if msg.Field != nil {
				if string(n.Name.AsIdentifier()) != msg.Field.Name {
					continue
				}
				return f.nodeInfoByField(n, msg.Field)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByOneof(node *ast.OneofNode, msg *Message) *ast.NodeInfo {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), msgOptionName) {
				continue
			}
			if msg.Option != nil {
				return f.nodeInfoByMessageOption(n, msg.Option)
			}
		case *ast.FieldNode:
			if msg.Field != nil {
				if string(n.Name.AsIdentifier()) != msg.Field.Name {
					continue
				}
				return f.nodeInfoByField(n, msg.Field)
			}
		}
	}
	return nil
}

func (f *File) nodeInfoByEnum(node *ast.EnumNode, enum *Enum) *ast.NodeInfo {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), enumOptionName) {
				continue
			}
			if enum.Option != nil {
				return f.nodeInfoByEnumOption(n, enum.Option)
			}
		case *ast.EnumValueNode:
			if enum.Value != nil {
				if string(n.Name.AsIdentifier()) != enum.Value.Value {
					continue
				}
				return f.nodeInfoByEnumValue(n, enum.Value)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnumOption(node *ast.OptionNode, opt *EnumOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if opt.Alias && strings.HasSuffix(f.optionName(node), "alias") {
			return f.nodeInfo(n)
		}
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Alias && optName == "alias":
				return f.nodeInfo(elem.Val)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnumValue(node *ast.EnumValueNode, value *EnumValue) *ast.NodeInfo {
	opts := node.Options
	if value.Option != nil && opts != nil {
		for _, opt := range opts.Options {
			if !f.matchOption(f.optionName(opt), enumValueOptionName) {
				continue
			}
			return f.nodeInfoByEnumValueOption(opt, value.Option)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnumValueOption(node *ast.OptionNode, opt *EnumValueOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if opt.Default && strings.HasSuffix(f.optionName(node), "default") {
			return f.nodeInfo(n)
		}
		if opt.Alias && strings.HasSuffix(f.optionName(node), "alias") {
			return f.nodeInfo(n)
		}
	case *ast.MessageLiteralNode:
		var attrs []*ast.MessageLiteralNode
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Default && optName == "default":
				return f.nodeInfo(elem.Val)
			case opt.Alias && optName == "alias":
				return f.nodeInfo(elem.Val)
			case opt.Attr != nil && optName == "attr":
				attrs = append(attrs, f.getMessageListFromNode(elem.Val)...)
			}
		}
		if len(attrs) != 0 {
			return f.nodeInfoByEnumValueAttribute(attrs, opt.Attr)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnumValueAttribute(list []*ast.MessageLiteralNode, attr *EnumValueAttribute) *ast.NodeInfo {
	if attr.Idx >= len(list) {
		return nil
	}
	node := list[attr.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case attr.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case attr.Value && fieldName == "value":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByMessageOption(node *ast.OptionNode, opt *MessageOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if opt.Alias && strings.HasSuffix(f.optionName(node), "alias") {
			return f.nodeInfo(n)
		}
	case *ast.MessageLiteralNode:
		var defs []*ast.MessageLiteralNode
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Def != nil && optName == "def":
				defs = append(defs, f.getMessageListFromNode(elem.Val)...)
			case opt.Alias && optName == "alias":
				return f.nodeInfo(elem.Val)
			}
		}
		if len(defs) != 0 {
			return f.nodeInfoByDef(defs, opt.Def)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByDef(list []*ast.MessageLiteralNode, def *VariableDefinitionOption) *ast.NodeInfo {
	if def.Idx >= len(list) {
		return nil
	}
	node := list[def.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case def.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case def.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case def.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case def.Map != nil && fieldName == "map":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByMapExpr(value, def.Map)
		case def.Call != nil && fieldName == "call":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByCallExpr(value, def.Call)
		case def.Message != nil && fieldName == "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByMessageExpr(value, def.Message)
		case def.Enum != nil && fieldName == "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByEnumExpr(value, def.Enum)
		case def.Validation != nil && fieldName == "validation":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByValidationExpr(value, def.Validation)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByMapExpr(node *ast.MessageLiteralNode, opt *MapExprOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.Iterator != nil && fieldName == "iterator":
			n, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			for _, elem := range n.Elements {
				optName := elem.Name.Name.AsIdentifier()
				switch {
				case opt.Iterator.Name && optName == "name":
					return f.nodeInfo(elem.Val)
				case opt.Iterator.Source && optName == "src":
					return f.nodeInfo(elem.Val)
				}
			}
		case opt.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Message != nil && fieldName == "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByMessageExpr(value, opt.Message)
		case opt.Enum != nil && fieldName == "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByEnumExpr(value, opt.Enum)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByCallExpr(node *ast.MessageLiteralNode, call *CallExprOption) *ast.NodeInfo {
	var (
		requests []*ast.MessageLiteralNode
		grpcErrs []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case call.Method && fieldName == "method":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case call.Request != nil && fieldName == "request":
			requests = append(requests, f.getMessageListFromNode(elem.Val)...)
		case call.Timeout && fieldName == "timeout":
			return f.nodeInfo(elem.Val)
		case call.Retry != nil && fieldName == "retry":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByRetry(value, call.Retry)
		case call.Error != nil && fieldName == "error":
			grpcErrs = append(grpcErrs, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(requests) != 0 {
		return f.nodeInfoByMethodRequest(requests, call.Request)
	}
	if len(grpcErrs) != 0 {
		return f.nodeInfoByGRPCError(grpcErrs, call.Error)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByMethodRequest(list []*ast.MessageLiteralNode, req *RequestOption) *ast.NodeInfo {
	if req.Idx >= len(list) {
		return nil
	}
	literal := list[req.Idx]
	for _, elem := range literal.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case req.Field && fieldName == "field":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case req.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case req.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(literal)
}

func (f *File) nodeInfoByRetry(node *ast.MessageLiteralNode, retry *RetryOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case retry.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case retry.Constant != nil && fieldName == "constant":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByRetryConstant(value, retry.Constant)
		case retry.Exponential != nil && fieldName == "exponential":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByRetryExponential(value, retry.Exponential)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByRetryConstant(node *ast.MessageLiteralNode, constant *RetryConstantOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case constant.Interval && fieldName == "interval":
			return f.nodeInfo(elem.Val)
		case constant.MaxRetries && fieldName == "max_retries":
			return f.nodeInfo(elem.Val)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByRetryExponential(node *ast.MessageLiteralNode, exp *RetryExponentialOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case exp.InitialInterval && fieldName == "initial_interval":
			return f.nodeInfo(elem.Val)
		case exp.RandomizationFactor && fieldName == "randomization_factor":
			return f.nodeInfo(elem.Val)
		case exp.Multiplier && fieldName == "multiplier":
			return f.nodeInfo(elem.Val)
		case exp.MaxInterval && fieldName == "max_interval":
			return f.nodeInfo(elem.Val)
		case exp.MaxRetries && fieldName == "max_retries":
			return f.nodeInfo(elem.Val)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByMessageExpr(node *ast.MessageLiteralNode, msg *MessageExprOption) *ast.NodeInfo {
	var args []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case msg.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case msg.Args != nil && fieldName == "args":
			args = append(args, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(args) != 0 {
		return f.nodeInfoByArgument(args, msg.Args)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnumExpr(node *ast.MessageLiteralNode, expr *EnumExprOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case expr.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case expr.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByValidationExpr(node *ast.MessageLiteralNode, opt *ValidationExprOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case fieldName == "error":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByGRPCError([]*ast.MessageLiteralNode{value}, opt.Error)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByGRPCError(list []*ast.MessageLiteralNode, opt *GRPCErrorOption) *ast.NodeInfo {
	if opt.Idx >= len(list) {
		return nil
	}
	node := list[opt.Idx]
	var (
		defs       []*ast.MessageLiteralNode
		errDetails []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Ignore && fieldName == "ignore":
			value, ok := elem.Val.(*ast.IdentNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.IgnoreAndResponse && fieldName == "ignore_and_response":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Def != nil && fieldName == "def":
			defs = append(defs, f.getMessageListFromNode(elem.Val)...)
		case opt.Detail != nil && fieldName == "details":
			errDetails = append(errDetails, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(defs) != 0 {
		return f.nodeInfoByDef(defs, opt.Def)
	}
	if len(errDetails) != 0 {
		return f.nodeInfoByGRPCErrorDetail(errDetails, opt.Detail)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByGRPCErrorDetail(list []*ast.MessageLiteralNode, detail *GRPCErrorDetailOption) *ast.NodeInfo {
	if detail.Idx >= len(list) {
		return nil
	}
	node := list[detail.Idx]
	var (
		defs                 []*ast.MessageLiteralNode
		messages             []*ast.MessageLiteralNode
		preconditionFailures []*ast.MessageLiteralNode
		badRequests          []*ast.MessageLiteralNode
		localizedMessages    []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case detail.Def != nil && fieldName == "def":
			defs = append(defs, f.getMessageListFromNode(elem.Val)...)
		case detail.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case detail.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case detail.Message != nil && fieldName == "message":
			messages = append(messages, f.getMessageListFromNode(elem.Val)...)
		case detail.PreconditionFailure != nil && fieldName == "precondition_failure":
			preconditionFailures = append(preconditionFailures, f.getMessageListFromNode(elem.Val)...)
		case detail.BadRequest != nil && fieldName == "bad_request":
			badRequests = append(badRequests, f.getMessageListFromNode(elem.Val)...)
		case detail.LocalizedMessage != nil && fieldName == "localized_message":
			localizedMessages = append(localizedMessages, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(defs) != 0 {
		return f.nodeInfoByDef(defs, detail.Def)
	}
	if len(messages) != 0 {
		return f.nodeInfoByDef(messages, detail.Message)
	}
	if len(preconditionFailures) != 0 {
		return f.nodeInfoByPreconditionFailure(preconditionFailures, detail.PreconditionFailure)
	}
	if len(badRequests) != 0 {
		return f.nodeInfoByBadRequest(badRequests, detail.BadRequest)
	}
	if len(localizedMessages) != 0 {
		return f.nodeInfoByLocalizedMessage(localizedMessages, detail.LocalizedMessage)
	}

	return f.nodeInfo(node)
}

func (f *File) nodeInfoByPreconditionFailure(list []*ast.MessageLiteralNode, failure *GRPCErrorDetailPreconditionFailureOption) *ast.NodeInfo {
	if failure.Idx >= len(list) {
		return nil
	}
	node := list[failure.Idx]
	var violations []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if fieldName == "violations" {
			violations = append(violations, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(violations) != 0 {
		return f.nodeInfoByPreconditionFailureViolations(violations, failure.Violation)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByPreconditionFailureViolations(list []*ast.MessageLiteralNode, violation GRPCErrorDetailPreconditionFailureViolationOption) *ast.NodeInfo {
	if violation.Idx >= len(list) {
		return nil
	}
	node := list[violation.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if string(fieldName) == violation.FieldName {
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByBadRequest(list []*ast.MessageLiteralNode, req *GRPCErrorDetailBadRequestOption) *ast.NodeInfo {
	if req.Idx >= len(list) {
		return nil
	}
	node := list[req.Idx]
	var violations []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if fieldName == "field_violations" {
			violations = append(violations, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(violations) != 0 {
		return f.nodeInfoByBadRequestFieldViolations(violations, req.FieldViolation)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByLocalizedMessage(list []*ast.MessageLiteralNode, msg *GRPCErrorDetailLocalizedMessageOption) *ast.NodeInfo {
	if msg.Idx >= len(list) {
		return nil
	}
	node := list[msg.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if string(fieldName) == msg.FieldName {
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByBadRequestFieldViolations(list []*ast.MessageLiteralNode, violation GRPCErrorDetailBadRequestFieldViolationOption) *ast.NodeInfo {
	if violation.Idx >= len(list) {
		return nil
	}
	node := list[violation.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if string(fieldName) == violation.FieldName {
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByArgument(list []*ast.MessageLiteralNode, arg *ArgumentOption) *ast.NodeInfo {
	if arg.Idx >= len(list) {
		return nil
	}
	literal := list[arg.Idx]
	for _, elem := range literal.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case arg.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case arg.Inline && fieldName == "inline":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(literal)
}

func (f *File) nodeInfoByField(node ast.FieldDeclNode, field *Field) *ast.NodeInfo {
	opts := node.GetOptions()
	if field.Option != nil && opts != nil {
		for _, opt := range opts.Options {
			if !f.matchOption(f.optionName(opt), fieldOptionName) {
				continue
			}
			return f.nodeInfoByFieldOption(opt, field.Option)
		}
	}
	if field.Type && node.FieldType() != nil {
		return f.nodeInfo(node.FieldType())
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByFieldOption(node *ast.OptionNode, opt *FieldOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if opt.By && strings.HasSuffix(f.optionName(node), "by") {
			return f.nodeInfo(n)
		}
	case *ast.MessageLiteralNode:
		if opt.Oneof != nil && strings.HasSuffix(f.optionName(node), "oneof") {
			return f.nodeInfoByFieldOneof(n, opt.Oneof)
		}
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.By && optName == "by":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfo(value)
			case opt.Oneof != nil && optName == "oneof":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfoByFieldOneof(value, opt.Oneof)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByFieldOneof(node *ast.MessageLiteralNode, opt *FieldOneof) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Default && fieldName == "default":
			return f.nodeInfo(elem.Val)
		case opt.Def != nil && fieldName == "def":
			return f.nodeInfoByDef(f.getMessageListFromNode(elem.Val), opt.Def)
		case opt.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByService(node *ast.ServiceNode, svc *Service) *ast.NodeInfo {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), serviceOptionName) {
				continue
			}
			if svc.Option != nil {
				return f.nodeInfoByServiceOption(n, svc.Option)
			}
		case *ast.RPCNode:
			if svc.Method == nil {
				continue
			}
			if n.Name.Val != svc.Method.Name {
				continue
			}
			return f.nodeInfoByMethod(n, svc.Method)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByServiceOption(node *ast.OptionNode, opt *ServiceOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			fieldName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Env != nil && fieldName == "env":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfoByEnv(value, opt.Env)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnv(node *ast.MessageLiteralNode, env *Env) *ast.NodeInfo {
	var vars []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case env.Message && fieldName == "message":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case env.Var != nil && fieldName == "var":
			vars = append(vars, f.getMessageListFromNode(elem.Val)...)
		}
	}
	if len(vars) != 0 {
		return f.nodeInfoByEnvVar(vars, env.Var)
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByEnvVar(list []*ast.MessageLiteralNode, v *EnvVar) *ast.NodeInfo {
	if v.Idx >= len(list) {
		return nil
	}
	node := list[v.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case v.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByMethod(node *ast.RPCNode, mtd *Method) *ast.NodeInfo {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), methodOptionName) {
				continue
			}
			if mtd.Option != nil {
				return f.nodeInfoByMethodOption(n, mtd.Option)
			}
		}
	}
	return nil
}

func (f *File) nodeInfoByMethodOption(node *ast.OptionNode, opt *MethodOption) *ast.NodeInfo {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if opt.Timeout && strings.HasSuffix(f.optionName(node), "timeout") {
			return f.nodeInfo(n)
		}
		if opt.Response && strings.HasSuffix(f.optionName(node), "response") {
			return f.nodeInfo(n)
		}
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Timeout && optName == "timeout":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfo(value)
			case opt.Response && optName == "response":
				value, ok := elem.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfo(value)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfo(node ast.Node) *ast.NodeInfo {
	if node == nil {
		return nil
	}
	n := f.fileNode.NodeInfo(node)
	return &n
}

func (f *File) matchOption(target, opt string) bool {
	return strings.HasPrefix(strings.TrimPrefix(target, "."), opt)
}
