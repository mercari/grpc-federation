package source

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"go.lsp.dev/protocol"
)

const (
	serviceOptionName   = "grpc.federation.service"
	methodOptionName    = "grpc.federation.method"
	msgOptionName       = "grpc.federation.message"
	fieldOptionName     = "grpc.federation.field"
	enumOptionName      = "grpc.federation.enum"
	enumValueOptionName = "grpc.federation.enum_value"
)

type File struct {
	path     string
	content  []byte
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
		fileNode: fileNode,
	}, nil
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
		importNode, ok := decl.(*ast.ImportNode)
		if !ok {
			continue
		}
		imports = append(imports, importNode.Name.AsString())
	}
	return imports
}

type findContext struct {
	fileName         string
	message          *Message
	service          *Service
	method           *Method
	field            *Field
	fieldOption      *FieldOption
	serviceDepOption *ServiceDependencyOption
	messageOption    *MessageOption
	def              *VariableDefinitionOption
}

func (f *File) buildLocation(ctx findContext) *Location {
	loc := &Location{FileName: ctx.fileName}
	if ctx.service != nil {
		loc.Service = ctx.service
		if ctx.serviceDepOption != nil {
			loc.Service.Option = &ServiceOption{
				Dependencies: ctx.serviceDepOption,
			}
		}
		if ctx.method != nil {
			loc.Service.Method = ctx.method
		}
	}
	if ctx.message != nil {
		loc.Message = ctx.message
		if ctx.field != nil {
			loc.Message.Field = ctx.field
			if ctx.fieldOption != nil {
				loc.Message.Field.Option = ctx.fieldOption
			}
		}
		if ctx.messageOption != nil {
			loc.Message.Option = &MessageOption{}
		}
		if ctx.def != nil {
			loc.Message.Option.VariableDefinitions = ctx.def
		}
	}
	return loc
}

// FindLocationByPos returns the corresponding location information from the position in the source code.
func (f *File) FindLocationByPos(pos Position) *Location {
	ctx := findContext{fileName: f.fileNode.Name()}
	for _, decl := range f.fileNode.Decls {
		switch n := decl.(type) {
		case *ast.MessageNode:
			if found := f.findMessageByPos(ctx, pos, n); found != nil {
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

func (f *File) findMessageByPos(ctx findContext, pos Position, node *ast.MessageNode) *Location {
	ctx.message = &Message{Name: string(node.Name.AsIdentifier())}
	for _, decl := range node.MessageBody.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), msgOptionName) {
				continue
			}
			if found := f.findMessageOptionByPos(ctx, pos, n); found != nil {
				return found
			}
		case *ast.FieldNode:
			if found := f.findFieldByPos(ctx, pos, n); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findFieldByPos(ctx findContext, pos Position, node *ast.FieldNode) *Location {
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
	return nil
}

func (f *File) findFieldOptionByPos(ctx findContext, pos Position, node *ast.OptionNode) *Location {
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

func (f *File) findMessageOptionByPos(ctx findContext, pos Position, node *ast.OptionNode) *Location {
	literal, ok := node.Val.(*ast.MessageLiteralNode)
	if !ok {
		return nil
	}
	ctx.messageOption = &MessageOption{}
	for _, elem := range literal.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "def":
			if found := f.findDefByPos(ctx, pos, f.getMessageListFromNode(elem.Val)); found != nil {
				return found
			}
		}
	}
	if f.containsPos(literal, pos) {
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

func (f *File) findDefByPos(ctx findContext, pos Position, list []*ast.MessageLiteralNode) *Location {
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

func (f *File) findMessageExprByPos(ctx findContext, defIdx int, pos Position, node *ast.MessageLiteralNode) *Location {
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

func (f *File) findMessageArgumentByPos(ctx findContext, defIdx int, pos Position, list []*ast.MessageLiteralNode) *Location {
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

func (f *File) findCallExprByPos(ctx findContext, defIdx int, pos Position, node *ast.MessageLiteralNode) *Location {
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

func (f *File) findMethodRequestByPos(ctx findContext, defIdx int, pos Position, list []*ast.MessageLiteralNode) *Location {
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

func (f *File) findServiceByPos(ctx findContext, pos Position, node *ast.ServiceNode) *Location {
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

func (f *File) findServiceOptionByPos(ctx findContext, pos Position, node *ast.OptionNode) *Location {
	literal, ok := node.Val.(*ast.MessageLiteralNode)
	if !ok {
		return nil
	}
	for _, elem := range literal.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "dependencies":
			if found := f.findServiceDependencyByPos(ctx, pos, f.getMessageListFromNode(elem.Val)); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findServiceDependencyByPos(ctx findContext, pos Position, list []*ast.MessageLiteralNode) *Location {
	for idx, literal := range list {
		for _, dep := range literal.Elements {
			fieldName := dep.Name.Name.AsIdentifier()
			switch fieldName {
			case "name":
				value, ok := dep.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.serviceDepOption = &ServiceDependencyOption{
						Idx:  idx,
						Name: true,
					}
					return f.buildLocation(ctx)
				}
			case "service":
				value, ok := dep.Val.(*ast.StringLiteralNode)
				if !ok {
					return nil
				}
				if f.containsPos(value, pos) {
					ctx.serviceDepOption = &ServiceDependencyOption{
						Idx:     idx,
						Service: true,
					}
					return f.buildLocation(ctx)
				}
			}
		}
		if f.containsPos(literal, pos) {
			ctx.serviceDepOption = &ServiceDependencyOption{}
			return f.buildLocation(ctx)
		}
	}
	return nil
}

func (f *File) findMethodByPos(ctx findContext, pos Position, node *ast.RPCNode) *Location {
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

func (f *File) findMethodOptionByPos(ctx findContext, pos Position, node *ast.OptionNode) *Location {
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
		case *ast.OptionNode:
			if f.matchOption(f.optionName(n), "go_package") && loc.GoPackage {
				return f.nodeInfo(n.Val)
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
		case *ast.EnumNode:
			if msg.Enum != nil {
				if string(n.Name.AsIdentifier()) == msg.Enum.Name {
					return f.nodeInfoByEnum(n, msg.Enum)
				}
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
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Default && optName == "default":
				return f.nodeInfo(elem.Val)
			case opt.Alias && optName == "alias":
				return f.nodeInfo(elem.Val)
			}
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
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.VariableDefinitions != nil && optName == "def":
				return f.nodeInfoByDef(f.getMessageListFromNode(elem.Val), opt.VariableDefinitions)
			case opt.Alias && optName == "alias":
				return f.nodeInfo(elem.Val)
			}
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

func (f *File) nodeInfoByMapExpr(node *ast.MessageLiteralNode, _ *MapExprOption) *ast.NodeInfo {
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByCallExpr(node *ast.MessageLiteralNode, call *CallExprOption) *ast.NodeInfo {
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
			return f.nodeInfoByMethodRequest(f.getMessageListFromNode(elem.Val), call.Request)
		case call.Timeout && fieldName == "timeout":
			return f.nodeInfo(elem.Val)
		case call.Retry != nil && fieldName == "retry":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByRetry(value, call.Retry)
		}
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
		}
	}
	return f.nodeInfo(literal)
}

func (f *File) nodeInfoByRetry(node *ast.MessageLiteralNode, retry *RetryOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
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
			return f.nodeInfoByArgument(f.getMessageListFromNode(elem.Val), msg.Args)
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
			return f.nodeInfoByValidationError(value, opt)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByValidationError(node *ast.MessageLiteralNode, opt *ValidationExprOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Detail != nil && fieldName == "details":
			return f.nodeInfoByValidationErrorDetail(f.getMessageListFromNode(elem.Val), opt.Detail)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByValidationErrorDetail(list []*ast.MessageLiteralNode, detail *ValidationDetailOption) *ast.NodeInfo {
	if detail.Idx >= len(list) {
		return nil
	}
	node := list[detail.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case detail.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case detail.Message != nil && fieldName == "message":
			return f.nodeInfoByDetailMessage(f.getMessageListFromNode(elem.Val), detail.Message)
		case detail.PreconditionFailure != nil && fieldName == "precondition_failure":
			return f.nodeInfoByPreconditionFailure(f.getMessageListFromNode(elem.Val), detail.PreconditionFailure)
		case detail.BadRequest != nil && fieldName == "bad_request":
			return f.nodeInfoByBadRequest(f.getMessageListFromNode(elem.Val), detail.BadRequest)
		case detail.LocalizedMessage != nil && fieldName == "localized_message":
			return f.nodeInfoByLocalizedMessage(f.getMessageListFromNode(elem.Val), detail.LocalizedMessage)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByDetailMessage(list []*ast.MessageLiteralNode, message *ValidationDetailMessageOption) *ast.NodeInfo {
	if message.Idx >= len(list) {
		return nil
	}
	return f.nodeInfoByMessageExpr(list[message.Idx], message.Message)
}

func (f *File) nodeInfoByPreconditionFailure(list []*ast.MessageLiteralNode, failure *ValidationDetailPreconditionFailureOption) *ast.NodeInfo {
	if failure.Idx >= len(list) {
		return nil
	}
	node := list[failure.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if fieldName == "violations" {
			return f.nodeInfoByPreconditionFailureViolations(f.getMessageListFromNode(elem.Val), failure.Violation)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByPreconditionFailureViolations(list []*ast.MessageLiteralNode, violation ValidationDetailPreconditionFailureViolationOption) *ast.NodeInfo {
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

func (f *File) nodeInfoByBadRequest(list []*ast.MessageLiteralNode, req *ValidationDetailBadRequestOption) *ast.NodeInfo {
	if req.Idx >= len(list) {
		return nil
	}
	node := list[req.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		if fieldName == "field_violations" {
			return f.nodeInfoByBadRequestFieldViolations(f.getMessageListFromNode(elem.Val), req.FieldViolation)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByLocalizedMessage(list []*ast.MessageLiteralNode, msg *ValidationDetailLocalizedMessageOption) *ast.NodeInfo {
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

func (f *File) nodeInfoByBadRequestFieldViolations(list []*ast.MessageLiteralNode, violation ValidationDetailBadRequestFieldViolationOption) *ast.NodeInfo {
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

func (f *File) nodeInfoByField(node *ast.FieldNode, field *Field) *ast.NodeInfo {
	opts := node.GetOptions()
	if field.Option != nil && opts != nil {
		for _, opt := range opts.Options {
			if !f.matchOption(f.optionName(opt), fieldOptionName) {
				continue
			}
			return f.nodeInfoByFieldOption(opt, field.Option)
		}
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
		case opt.VariableDefinitions != nil && fieldName == "def":
			return f.nodeInfoByDef(f.getMessageListFromNode(elem.Val), opt.VariableDefinitions)
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
			if svc.Method != nil {
				return f.nodeInfoByMethod(n, svc.Method)
			}
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByServiceOption(node *ast.OptionNode, opt *ServiceOption) *ast.NodeInfo {
	literal, ok := node.Val.(*ast.MessageLiteralNode)
	if !ok {
		return nil
	}
	for _, elem := range literal.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.Dependencies != nil && optName == "dependencies":
			return f.nodeInfoByServiceDependency(f.getMessageListFromNode(elem.Val), opt.Dependencies)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByServiceDependency(list []*ast.MessageLiteralNode, dep *ServiceDependencyOption) *ast.NodeInfo {
	if dep.Idx >= len(list) {
		return nil
	}
	literal := list[dep.Idx]
	for _, elem := range literal.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case dep.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case dep.Service && fieldName == "service":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		}
	}
	return f.nodeInfo(literal)
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

func (f *File) SemanticTokenTypeMap() map[ast.Token]protocol.SemanticTokenTypes {
	tokenToSemanticTokenTypeMap := map[ast.Token]protocol.SemanticTokenTypes{}
	_ = ast.Walk(f.fileNode, &ast.SimpleVisitor{
		DoVisitMessageNode: func(msg *ast.MessageNode) error {
			tokenToSemanticTokenTypeMap[msg.Name.Token()] = protocol.SemanticTokenType
			return nil
		},
		DoVisitFieldNode: func(field *ast.FieldNode) error {
			nameToken := field.Name.Token()
			tokenToSemanticTokenTypeMap[nameToken-1] = protocol.SemanticTokenType
			tokenToSemanticTokenTypeMap[nameToken] = protocol.SemanticTokenProperty
			return nil
		},
		DoVisitServiceNode: func(svc *ast.ServiceNode) error {
			tokenToSemanticTokenTypeMap[svc.Name.Token()] = protocol.SemanticTokenType
			return nil
		},
		DoVisitRPCNode: func(rpc *ast.RPCNode) error {
			tokenToSemanticTokenTypeMap[rpc.Name.Token()] = protocol.SemanticTokenMethod
			tokenToSemanticTokenTypeMap[rpc.Input.OpenParen.Token()+1] = protocol.SemanticTokenType
			tokenToSemanticTokenTypeMap[rpc.Output.OpenParen.Token()+1] = protocol.SemanticTokenType
			return nil
		},
		DoVisitOptionNode: func(opt *ast.OptionNode) error {
			for _, part := range opt.Name.Parts {
				switch n := part.Name.(type) {
				case *ast.CompoundIdentNode:
					for _, comp := range n.Components {
						tokenToSemanticTokenTypeMap[comp.Token()] = protocol.SemanticTokenNamespace
					}
				}
			}
			return nil
		},
	})
	return tokenToSemanticTokenTypeMap
}
