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
	fileOptionName      = "grpc.federation.file"
	serviceOptionName   = "grpc.federation.service"
	methodOptionName    = "grpc.federation.method"
	msgOptionName       = "grpc.federation.message"
	oneofOptionName     = "grpc.federation.oneof"
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
			for _, value := range f.importPathValuesByImportRule(declNode) {
				imports = append(imports, value.AsString())
			}
		}
	}
	return imports
}

// FindLocationByPos returns the corresponding location information from the position in the source code.
func (f *File) FindLocationByPos(pos Position) *Location {
	builder := NewLocationBuilder(f.Path())
	for _, decl := range f.fileNode.Decls {
		switch n := decl.(type) {
		case *ast.ImportNode:
			if found := f.findImportByPos(
				builder.WithImportName(n.Name.AsString()),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), fileOptionName) {
				continue
			}
			if found := f.findFileOptionByPos(
				builder,
				pos, n,
			); found != nil {
				return found
			}
		case *ast.MessageNode:
			if found := f.findMessageByPos(
				builder.WithMessage(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.EnumNode:
			if found := f.findEnumByPos(
				builder.WithEnum(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.ServiceNode:
			if found := f.findServiceByPos(
				builder.WithService(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findFileOptionByPos(builder *LocationBuilder, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "import":
				switch n := elem.Val.(type) {
				case *ast.StringLiteralNode:
					if f.containsPos(n, pos) {
						return builder.WithImportName(n.AsString()).Location()
					}
				case *ast.ArrayLiteralNode:
					for _, elem := range n.Elements {
						n, ok := elem.(*ast.StringLiteralNode)
						if !ok {
							continue
						}
						if f.containsPos(n, pos) {
							return builder.WithImportName(n.AsString()).Location()
						}
					}
				}
			}
		}
	}
	if f.containsPos(node.Name, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findImportByPos(builder *LocationBuilder, pos Position, node *ast.ImportNode) *Location {
	if f.containsPos(node.Name, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMessageByPos(builder *MessageBuilder, pos Position, node *ast.MessageNode) (loc *Location) {
	for _, decl := range node.MessageBody.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), msgOptionName) {
				continue
			}
			if found := f.findMessageOptionByPos(
				builder.WithOption(),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.FieldNode:
			if found := f.findFieldByPos(
				builder.WithField(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.MessageNode:
			if found := f.findMessageByPos(
				builder.WithMessage(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.EnumNode:
			if found := f.findEnumByPos(
				builder.WithEnum(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.OneofNode:
			if found := f.findOneofByPos(
				builder.WithOneof(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findEnumByPos(builder *EnumBuilder, pos Position, node *ast.EnumNode) *Location {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), enumOptionName) {
				continue
			}
			if found := f.findEnumOptionByPos(
				builder.WithOption(),
				pos, n,
			); found != nil {
				return found
			}
		case *ast.EnumValueNode:
			if found := f.findEnumValueByPos(
				builder.WithValue(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findEnumOptionByPos(builder *EnumBuilder, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				if f.containsPos(elem.Val, pos) {
					return builder.Location()
				}
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				return builder.Location()
			}
		}
	}
	return nil
}

func (f *File) findEnumValueByPos(builder *EnumValueBuilder, pos Position, node *ast.EnumValueNode) *Location {
	if node.Options != nil {
		for _, opt := range node.Options.Options {
			if !f.matchOption(f.optionName(opt), enumValueOptionName) {
				continue
			}
			if found := f.findEnumValueOptionByPos(
				builder.WithOption(),
				pos, opt,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findEnumValueOptionByPos(builder *EnumValueOptionBuilder, pos Position, node *ast.OptionNode) *Location {
	var attrs []*ast.MessageLiteralNode
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				switch value := elem.Val.(type) {
				case *ast.StringLiteralNode:
					if f.containsPos(value, pos) {
						return builder.WithAlias().Location()
					}
				case *ast.ArrayLiteralNode:
					for _, elem := range value.Elements {
						if f.containsPos(elem, pos) {
							return builder.WithAlias().Location()
						}
					}
				}
			case "noalias":
				if f.containsPos(elem.Val, pos) {
					return builder.WithNoAlias().Location()
				}
			case "default":
				if f.containsPos(elem.Val, pos) {
					return builder.WithDefault().Location()
				}
			case "attr":
				attrs = append(attrs, f.getMessageListFromNode(elem.Val)...)
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				return builder.WithAlias().Location()
			}
		}
	case *ast.IdentNode:
		if strings.HasSuffix(f.optionName(node), "noalias") {
			if f.containsPos(n, pos) {
				return builder.WithNoAlias().Location()
			}
		}
		if strings.HasSuffix(f.optionName(node), "default") {
			if f.containsPos(n, pos) {
				return builder.WithDefault().Location()
			}
		}
	}
	for idx, n := range attrs {
		if found := f.findAttrByPos(
			builder.WithAttr(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	return nil
}

func (f *File) findAttrByPos(builder *EnumValueAttributeBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "value":
			if f.containsPos(elem.Val, pos) {
				return builder.WithValue().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findOneofByPos(builder *OneofBuilder, pos Position, node *ast.OneofNode) *Location {
	if node == nil {
		return nil
	}
	for _, elem := range node.Decls {
		switch n := elem.(type) {
		case *ast.FieldNode:
			if found := f.findFieldByPos(
				builder.WithField(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findFieldByPos(builder *FieldBuilder, pos Position, node *ast.FieldNode) *Location {
	opts := node.GetOptions()
	if opts != nil {
		for _, opt := range opts.Options {
			if !f.matchOption(f.optionName(opt), fieldOptionName) {
				continue
			}
			if found := f.findFieldOptionByPos(
				builder.WithOption(),
				pos, opt,
			); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node.FldType, pos) {
		return builder.WithType().Location()
	}
	return nil
}

func (f *File) findFieldOptionByPos(builder *FieldOptionBuilder, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "by") {
			if f.containsPos(n, pos) {
				return builder.WithBy().Location()
			}
		}
	case *ast.MessageLiteralNode:
		switch {
		case strings.HasSuffix(f.optionName(node), "oneof"):
			if found := f.findFieldOneofByPos(
				builder.WithOneOf(),
				pos, n,
			); found != nil {
				return found
			}
		}
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "by":
				if f.containsPos(elem.Val, pos) {
					return builder.WithBy().Location()
				}
			case "oneof":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				if found := f.findFieldOneofByPos(
					builder.WithOneOf(),
					pos, value,
				); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

func (f *File) findFieldOneofByPos(builder *FieldOneofBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var defs []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "default":
			if f.containsPos(elem.Val, pos) {
				return builder.WithDefault().Location()
			}
		case "def":
			defs = append(defs, f.getMessageListFromNode(elem.Val)...)
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		}
	}
	for idx, n := range defs {
		if found := f.findDefByPos(
			builder.WithDef(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMessageOptionByPos(builder *MessageOptionBuilder, pos Position, node *ast.OptionNode) *Location {
	var defs []*ast.MessageLiteralNode
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "alias":
				if f.containsPos(elem.Val, pos) {
					return builder.WithAlias().Location()
				}
			case "def":
				defs = append(defs, f.getMessageListFromNode(elem.Val)...)
			}
		}
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "alias") {
			if f.containsPos(n, pos) {
				return builder.WithAlias().Location()
			}
		}
	}
	for idx, n := range defs {
		if found := f.findDefByPos(
			builder.WithDef(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node.Val, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findDefByPos(builder *VariableDefinitionOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		case "call":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findCallExprByPos(
				builder.WithCall(),
				pos, value,
			); found != nil {
				return found
			}
		case "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findMessageExprByPos(
				builder.WithMessage(),
				pos, value,
			); found != nil {
				return found
			}
		case "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findEnumExprByPos(
				builder.WithEnum(),
				pos, value,
			); found != nil {
				return found
			}
		case "validation":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findValidationExprByPos(
				builder.WithValidation(),
				pos, value,
			); found != nil {
				return found
			}
		case "map":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findMapExprByPos(
				builder.WithMap(),
				pos, value,
			); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMessageExprByPos(builder *MessageExprOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var args []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "args":
			args = append(args, f.getMessageListFromNode(elem.Val)...)
		}
	}
	for idx, n := range args {
		if found := f.findMessageArgumentByPos(
			builder.WithArgs(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findEnumExprByPos(builder *EnumExprOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMessageArgumentByPos(builder *ArgumentOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		case "inline":
			if f.containsPos(elem.Val, pos) {
				return builder.WithInline().Location()
			}
		}
	}
	return nil
}

func (f *File) findCallExprByPos(builder *CallExprOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var (
		requests []*ast.MessageLiteralNode
		grpcErrs []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "method":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMethod().Location()
			}
		case "timeout":
			if f.containsPos(elem.Val, pos) {
				return builder.WithTimeout().Location()
			}
		case "metadata":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMetadata().Location()
			}
		case "retry":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findRetryOptionByPos(
				builder.WithRetry(),
				pos, value,
			); found != nil {
				return found
			}
		case "option":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findGRPCCallOptionByPos(
				builder.WithOption(),
				pos, value,
			); found != nil {
				return found
			}
		case "request":
			requests = append(requests, f.getMessageListFromNode(elem.Val)...)
		case "error":
			grpcErrs = append(grpcErrs, f.getMessageListFromNode(elem.Val)...)
		}
	}
	for idx, n := range requests {
		if found := f.findMethodRequestByPos(
			builder.WithRequest(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	for idx, n := range grpcErrs {
		if found := f.findGRPCErrorByPos(
			builder.WithError(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findRetryOptionByPos(builder *RetryOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "constant":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			for _, subElem := range value.Elements {
				fieldName := subElem.Name.Name.AsIdentifier()
				switch fieldName {
				case "interval":
					if f.containsPos(subElem.Val, pos) {
						return builder.WithConstantInterval().Location()
					}
				}
			}
		case "exponential":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			for _, subElem := range value.Elements {
				fieldName := subElem.Name.Name.AsIdentifier()
				switch fieldName {
				case "initial_interval":
					if f.containsPos(subElem.Val, pos) {
						return builder.WithExponentialInitialInterval().Location()
					}
				case "max_interval":
					if f.containsPos(subElem.Val, pos) {
						return builder.WithExponentialMaxInterval().Location()
					}
				}
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findGRPCCallOptionByPos(builder *GRPCCallOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "content_subtype":
			if f.containsPos(elem.Val, pos) {
				return builder.WithContentSubtype().Location()
			}
		case "header":
			if f.containsPos(elem.Val, pos) {
				return builder.WithHeader().Location()
			}
		case "max_call_recv_msg_size":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMaxCallRecvMsgSize().Location()
			}
		case "max_call_send_msg_size":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMaxCallSendMsgSize().Location()
			}
		case "static_method":
			if f.containsPos(elem.Val, pos) {
				return builder.WithStaticMethod().Location()
			}
		case "trailer":
			if f.containsPos(elem.Val, pos) {
				return builder.WithTrailer().Location()
			}
		case "wait_for_ready":
			if f.containsPos(elem.Val, pos) {
				return builder.WithWaitForReady().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMethodRequestByPos(builder *RequestOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "field":
			if f.containsPos(elem.Val, pos) {
				return builder.WithField().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findGRPCErrorByPos(builder *GRPCErrorOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var (
		defs    []*ast.MessageLiteralNode
		details []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "def":
			defs = append(defs, f.getMessageListFromNode(elem.Val)...)
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "message":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMessage().Location()
			}
		case "ignore":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIgnore().Location()
			}
		case "ignore_and_Response":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIgnoreAndResponse().Location()
			}
		case "details":
			details = append(details, f.getMessageListFromNode(elem.Val)...)
		}
	}
	for idx, n := range defs {
		if found := f.findDefByPos(
			builder.WithDef(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	for idx, n := range details {
		if found := f.findGRPCErrorDetailByPos(
			builder.WithDetail(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	return nil
}

func (f *File) findGRPCErrorDetailByPos(builder *GRPCErrorDetailOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var (
		defs []*ast.MessageLiteralNode
		msgs []*ast.MessageLiteralNode
	)
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "def":
			defs = append(defs, f.getMessageListFromNode(elem.Val)...)
		case "message":
			msgs = append(msgs, f.getMessageListFromNode(elem.Val)...)
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		}
	}
	for idx, n := range defs {
		if found := f.findDefByPos(
			builder.WithDef(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	for idx, n := range msgs {
		if found := f.findDefByPos(
			builder.WithMessage(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	return nil
}

func (f *File) findValidationExprByPos(builder *ValidationExprOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var grpcErrs []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "error":
			grpcErrs = append(grpcErrs, f.getMessageListFromNode(elem.Val)...)
		}
	}
	for _, n := range grpcErrs {
		if found := f.findGRPCErrorByPos(
			builder.WithError(),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMapExprByPos(builder *MapExprOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch fieldName {
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		case "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findMessageExprByPos(
				builder.WithMessage(),
				pos, value,
			); found != nil {
				return found
			}
		case "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findEnumExprByPos(
				builder.WithEnum(),
				pos, value,
			); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findServiceByPos(builder *ServiceBuilder, pos Position, node *ast.ServiceNode) *Location {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), serviceOptionName) {
				continue
			}
			if found := f.findServiceOptionByPos(builder.WithOption(), pos, n); found != nil {
				return found
			}
		case *ast.RPCNode:
			if found := f.findMethodByPos(
				builder.WithMethod(string(n.Name.AsIdentifier())),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findServiceOptionByPos(builder *ServiceOptionBuilder, pos Position, node *ast.OptionNode) *Location {
	var vars []*ast.MessageLiteralNode
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
				if found := f.findEnvByPos(builder.WithEnv(), pos, value); found != nil {
					return found
				}
			case "var":
				vars = append(vars, f.getMessageListFromNode(elem.Val)...)
			}
		}
	}
	for idx, n := range vars {
		if found := f.findServiceVariableByPos(
			builder.WithVar(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node.Val, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findServiceVariableByPos(builder *ServiceVariableBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "by":
			if f.containsPos(elem.Val, pos) {
				return builder.WithBy().Location()
			}
		case "map":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findMapExprByPos(
				builder.WithMap(),
				pos, value,
			); found != nil {
				return found
			}
		case "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findMessageExprByPos(
				builder.WithMessage(),
				pos, value,
			); found != nil {
				return found
			}
		case "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findEnumExprByPos(
				builder.WithEnum(),
				pos, value,
			); found != nil {
				return found
			}
		case "validation":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findServiceVariableValidationExprByPos(
				builder.WithValidation(),
				pos, value,
			); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findServiceVariableValidationExprByPos(builder *ServiceVariableValidationExprBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "if":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIf().Location()
			}
		case "message":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMessage().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findEnvByPos(builder *EnvBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	var vars []*ast.MessageLiteralNode
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "message":
			if f.containsPos(elem.Val, pos) {
				return builder.WithMessage().Location()
			}
		case "var":
			vars = append(vars, f.getMessageListFromNode(elem.Val)...)
		}
	}
	for idx, n := range vars {
		if found := f.findEnvVarByPos(
			builder.WithVar(idx),
			pos, n,
		); found != nil {
			return found
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findEnvVarByPos(builder *EnvVarBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "name":
			if f.containsPos(elem.Val, pos) {
				return builder.WithName().Location()
			}
		case "type":
			if f.containsPos(elem.Val, pos) {
				return builder.WithType().Location()
			}
		case "option":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			if found := f.findEnvVarOptionByPos(
				builder.WithOption(),
				pos, value,
			); found != nil {
				return found
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findEnvVarOptionByPos(builder *EnvVarOptionBuilder, pos Position, node *ast.MessageLiteralNode) *Location {
	for _, elem := range node.Elements {
		optName := elem.Name.Name.AsIdentifier()
		switch optName {
		case "alternate":
			if f.containsPos(elem.Val, pos) {
				return builder.WithAlternate().Location()
			}
		case "default":
			if f.containsPos(elem.Val, pos) {
				return builder.WithDefault().Location()
			}
		case "required":
			if f.containsPos(elem.Val, pos) {
				return builder.WithRequired().Location()
			}
		case "ignored":
			if f.containsPos(elem.Val, pos) {
				return builder.WithIgnored().Location()
			}
		}
	}
	if f.containsPos(node, pos) {
		return builder.Location()
	}
	return nil
}

func (f *File) findMethodByPos(builder *MethodBuilder, pos Position, node *ast.RPCNode) *Location {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), methodOptionName) {
				continue
			}
			if found := f.findMethodOptionByPos(
				builder.WithOption(),
				pos, n,
			); found != nil {
				return found
			}
		}
	}
	return nil
}

func (f *File) findMethodOptionByPos(builder *MethodOptionBuilder, pos Position, node *ast.OptionNode) *Location {
	switch n := node.Val.(type) {
	case *ast.StringLiteralNode:
		if strings.HasSuffix(f.optionName(node), "timeout") {
			if f.containsPos(n, pos) {
				return builder.WithTimeout().Location()
			}
		}
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "timeout":
				if f.containsPos(elem.Val, pos) {
					return builder.WithTimeout().Location()
				}
			}
		}
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

func (f *File) importPathValuesByImportRule(node *ast.OptionNode) []*ast.StringLiteralNode {
	optName := f.optionName(node)
	if !f.matchOption(optName, "grpc.federation.file") {
		return nil
	}

	// option (grpc.federation.file).import = "example.proto";
	if strings.HasSuffix(optName, ".import") {
		if n, ok := node.GetValue().(*ast.StringLiteralNode); ok {
			return []*ast.StringLiteralNode{n}
		}
		return nil
	}

	var ret []*ast.StringLiteralNode
	switch n := node.Val.(type) {
	case *ast.MessageLiteralNode:
		for _, elem := range n.Elements {
			optName := elem.Name.Name.AsIdentifier()
			switch optName {
			case "import":
				switch n := elem.Val.(type) {
				case *ast.StringLiteralNode:
					// option (grpc.federation.file) = {
					//   import: "example.proto"
					// };
					ret = append(ret, n)
				case *ast.ArrayLiteralNode:
					// option (grpc.federation.file) = {
					//   import: ["example.proto"]
					// };
					for _, elem := range n.Elements {
						n, ok := elem.(*ast.StringLiteralNode)
						if !ok {
							continue
						}
						ret = append(ret, n)
					}
				}
			}
		}
	}
	return ret
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
				for _, v := range f.importPathValuesByImportRule(n) {
					if v.AsString() == loc.ImportName {
						return f.nodeInfo(v)
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
			if msg.Oneof != nil {
				if info := f.nodeInfoByOneof(n, msg.Oneof); info != nil {
					return info
				}
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

func (f *File) nodeInfoByOneof(node *ast.OneofNode, oneof *Oneof) *ast.NodeInfo {
	for _, decl := range node.Decls {
		switch n := decl.(type) {
		case *ast.OptionNode:
			if !f.matchOption(f.optionName(n), oneofOptionName) {
				continue
			}
			if oneof.Option != nil {
				return nil
			}
		case *ast.FieldNode:
			if oneof.Field != nil {
				if string(n.Name.AsIdentifier()) != oneof.Field.Name {
					continue
				}
				return f.nodeInfoByField(n, oneof.Field)
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
	case *ast.IdentNode:
		if opt.Default && strings.HasSuffix(f.optionName(node), "default") {
			return f.nodeInfo(n)
		}
	case *ast.StringLiteralNode:
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
			case opt.NoAlias && optName == "noalias":
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
		case call.Option != nil && fieldName == "option":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByGRPCCallOption(value, call.Option)
		case call.Metadata && fieldName == "metadata":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
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

func (f *File) nodeInfoByGRPCCallOption(node *ast.MessageLiteralNode, opt *GRPCCallOption) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case opt.ContentSubtype && fieldName == "content_subtype":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Header && fieldName == "header":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.Trailer && fieldName == "trailer":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.MaxCallRecvMsgSize && fieldName == "max_call_recv_msg_size":
			return f.nodeInfo(elem.Val)
		case opt.MaxCallSendMsgSize && fieldName == "max_call_send_msg_size":
			return f.nodeInfo(elem.Val)
		case opt.StaticMethod && fieldName == "static_method":
			value, ok := elem.Val.(*ast.IdentNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case opt.WaitForReady && fieldName == "wait_for_ready":
			value, ok := elem.Val.(*ast.IdentNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
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
		var vars []*ast.MessageLiteralNode
		for _, elem := range n.Elements {
			fieldName := elem.Name.Name.AsIdentifier()
			switch {
			case opt.Env != nil && fieldName == "env":
				value, ok := elem.Val.(*ast.MessageLiteralNode)
				if !ok {
					return nil
				}
				return f.nodeInfoByEnv(value, opt.Env)
			case opt.Var != nil && fieldName == "var":
				vars = append(vars, f.getMessageListFromNode(elem.Val)...)
			}
		}
		if len(vars) != 0 {
			return f.nodeInfoByServiceVariable(vars, opt.Var)
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

func (f *File) nodeInfoByServiceVariable(list []*ast.MessageLiteralNode, svcVar *ServiceVariable) *ast.NodeInfo {
	if svcVar.Idx >= len(list) {
		return nil
	}
	node := list[svcVar.Idx]
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case svcVar.Name && fieldName == "name":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case svcVar.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case svcVar.By && fieldName == "by":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case svcVar.Map != nil && fieldName == "map":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByMapExpr(value, svcVar.Map)
		case svcVar.Message != nil && fieldName == "message":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByMessageExpr(value, svcVar.Message)
		case svcVar.Enum != nil && fieldName == "enum":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByEnumExpr(value, svcVar.Enum)
		case svcVar.Validation != nil && fieldName == "validation":
			value, ok := elem.Val.(*ast.MessageLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfoByServiceVariableValidationExpr(value, svcVar.Validation)
		}
	}
	return f.nodeInfo(node)
}

func (f *File) nodeInfoByServiceVariableValidationExpr(node *ast.MessageLiteralNode, expr *ServiceVariableValidationExpr) *ast.NodeInfo {
	for _, elem := range node.Elements {
		fieldName := elem.Name.Name.AsIdentifier()
		switch {
		case expr.If && fieldName == "if":
			value, ok := elem.Val.(*ast.StringLiteralNode)
			if !ok {
				return nil
			}
			return f.nodeInfo(value)
		case expr.Message && fieldName == "message":
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
