package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/types"
)

var DefaultProtoFormatOption = &ProtoFormatOption{IndentSpaceNum: 2}

type ProtoFormatOption struct {
	IndentLevel    int
	Prefix         string
	IndentSpaceNum int
}

func (o *ProtoFormatOption) indentFormat() string {
	opt := o
	if opt == nil {
		opt = DefaultProtoFormatOption
	}
	return opt.Prefix + strings.Repeat(" ", opt.IndentSpaceNum*opt.IndentLevel)
}

func (o *ProtoFormatOption) toNextIndentLevel() *ProtoFormatOption {
	opt := o
	if opt == nil {
		opt = DefaultProtoFormatOption
	}
	return &ProtoFormatOption{
		IndentLevel:    opt.IndentLevel + 1,
		Prefix:         opt.Prefix,
		IndentSpaceNum: opt.IndentSpaceNum,
	}
}

func (r *FieldRule) ProtoFormat(opt *ProtoFormatOption) string {
	if r == nil {
		return ""
	}
	indent := opt.indentFormat()
	switch {
	case r.CustomResolver:
		return indent + "(grpc.federation.field).custom_resolver = true"
	case r.Alias != nil:
		return indent + fmt.Sprintf("(grpc.federation.field).alias = %q", r.Alias.Name)
	case r.Value != nil:
		value := r.Value.ProtoFormat(opt)
		value = strings.Replace(value, ":", " =", 1)
		return indent + fmt.Sprintf("(grpc.federation.field).%s", value)
	}
	return ""
}

func (r *MessageRule) ProtoFormat(opt *ProtoFormatOption) string {
	if r == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()

	var elems []string
	varDefs := r.VariableDefinitions.ProtoFormat(nextOpt)
	if varDefs != "" {
		elems = append(elems, varDefs)
	}
	if r.CustomResolver {
		elems = append(elems, nextOpt.indentFormat()+"custom_resolver: true")
	}
	if r.Alias != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("alias: %q", r.Alias.FQDN()))
	}
	if len(elems) == 0 {
		return indent + "option (grpc.federation.message) = {}"
	}
	return indent + fmt.Sprintf("option (grpc.federation.message) = {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (defs VariableDefinitions) ProtoFormat(opt *ProtoFormatOption) string {
	if len(defs) == 0 {
		return ""
	}
	formattedDefs := make([]string, 0, len(defs))
	for _, def := range defs {
		format := def.ProtoFormat(opt)
		if format == "" {
			continue
		}
		formattedDefs = append(formattedDefs, format)
	}
	return strings.Join(formattedDefs, "\n")
}

func (def *VariableDefinition) ProtoFormat(opt *ProtoFormatOption) string {
	if def == nil {
		return ""
	}
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if def.Name != "" {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("name: %q", def.Name))
	}
	if def.AutoBind {
		elems = append(elems, nextOpt.indentFormat()+`autobind: true`)
	}
	elems = append(elems, def.Expr.ProtoFormat(nextOpt))
	indent := opt.indentFormat()
	return indent + fmt.Sprintf("def {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (e *VariableExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	switch {
	case e.By != nil:
		return opt.indentFormat() + fmt.Sprintf("by: %q", e.By.Expr)
	case e.Call != nil:
		return e.Call.ProtoFormat(opt)
	case e.Message != nil:
		return e.Message.ProtoFormat(opt)
	case e.Map != nil:
		return e.Map.ProtoFormat(opt)
	case e.Validation != nil:
		return e.Validation.ProtoFormat(opt)
	}
	return ""
}

func (e *CallExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()

	var elems []string
	method := e.Method.ProtoFormat(nextOpt)
	if method != "" {
		elems = append(elems, method)
	}
	request := e.Request.ProtoFormat(nextOpt)
	if request != "" {
		elems = append(elems, request)
	}
	if len(elems) == 0 {
		return ""
	}
	return indent + fmt.Sprintf("call {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (e *MapExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if e.Iterator != nil {
		elems = append(elems, e.Iterator.ProtoFormat(nextOpt))
	}
	if e.Expr != nil {
		elems = append(elems, e.Expr.ProtoFormat(nextOpt)...)
	}
	if len(elems) == 0 {
		return ""
	}
	return indent + fmt.Sprintf("map {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (iter *Iterator) ProtoFormat(opt *ProtoFormatOption) string {
	if iter == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if iter.Name != "" {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("name: %q", iter.Name))
	}
	if iter.Source != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("src: %q", iter.Source.Name))
	}
	if len(elems) == 0 {
		return ""
	}
	return indent + fmt.Sprintf("iterator {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (e *MapIteratorExpr) ProtoFormat(opt *ProtoFormatOption) []string {
	if e == nil {
		return nil
	}
	var elems []string
	if e.By != nil {
		elems = append(elems, opt.indentFormat()+fmt.Sprintf("by: %q", e.By.Expr))
	}
	if e.Message != nil {
		elems = append(elems, e.Message.ProtoFormat(opt))
	}
	return elems
}

func (e *MessageExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if e.Message != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("name: %q", e.Message.Name))
	}
	args := e.protoFormatMessageArgs(nextOpt)
	if args != "" {
		elems = append(elems, args)
	}
	if len(elems) == 0 {
		return ""
	}
	return indent + fmt.Sprintf("message {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (e *MessageExpr) protoFormatMessageArgs(opt *ProtoFormatOption) string {
	var args []string
	for _, arg := range e.Args {
		value := arg.ProtoFormat(opt, false)
		if value == "" {
			continue
		}
		args = append(args, value)
	}

	indent := opt.indentFormat()
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return indent + fmt.Sprintf(`args %s`, args[0])
	}

	nextOpt := opt.toNextIndentLevel()
	formattedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		formattedArgs = append(formattedArgs, nextOpt.indentFormat()+arg)
	}
	return indent + fmt.Sprintf("args: [\n%s\n%s]", strings.Join(formattedArgs, ",\n"), indent)
}

func (e *ValidationExpr) ProtoFormat(opt *ProtoFormatOption) string {
	var elems []string
	elems = append(elems, e.Error.ProtoFormat(opt.toNextIndentLevel()))
	indent := opt.indentFormat()
	return indent + fmt.Sprintf("validation {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (m *Method) ProtoFormat(opt *ProtoFormatOption) string {
	if m == nil {
		return ""
	}
	return opt.indentFormat() + fmt.Sprintf("method: %q", m.FQDN())
}

func (r *Request) ProtoFormat(opt *ProtoFormatOption) string {
	if r == nil {
		return ""
	}
	indent := opt.indentFormat()
	fields := make([]string, 0, len(r.Args))
	for _, arg := range r.Args {
		field := arg.ProtoFormat(opt, true)
		if field == "" {
			continue
		}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return ""
	}
	if len(fields) == 1 {
		return indent + fmt.Sprintf("request %s", fields[0])
	}

	nextOpt := opt.toNextIndentLevel()
	formattedFields := make([]string, 0, len(fields))
	for _, field := range fields {
		formattedFields = append(formattedFields, nextOpt.indentFormat()+field)
	}
	return indent + fmt.Sprintf("request: [\n%s\n%s]", strings.Join(formattedFields, ",\n"), indent)
}

func (f *AutoBindField) ProtoFormat(opt *ProtoFormatOption) string {
	if f.VariableDefinition != nil {
		return f.VariableDefinition.ProtoFormat(opt)
	}
	return ""
}

func (vs MessageValidations) ProtoFormat(opt *ProtoFormatOption) string {
	if len(vs) == 0 {
		return ""
	}
	indent := opt.indentFormat()
	if len(vs) == 1 {
		return indent + fmt.Sprintf("validations %s", strings.TrimLeft(vs[0].ProtoFormat(opt), " "))
	}
	validations := make([]string, 0, len(vs))
	for _, validation := range vs {
		if format := validation.ProtoFormat(opt.toNextIndentLevel()); format != "" {
			validations = append(validations, format)
		}
	}
	return indent + fmt.Sprintf("validations: [\n%s\n%s]", strings.Join(validations, ",\n"), indent)
}

func (v *ValidationRule) ProtoFormat(opt *ProtoFormatOption) string {
	if v == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	elems = append(
		elems,
		nextOpt.indentFormat()+fmt.Sprintf("name: %q", v.Name),
		v.Error.ProtoFormat(nextOpt),
	)
	return indent + fmt.Sprintf("{\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (v *ValidationError) ProtoFormat(opt *ProtoFormatOption) string {
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	elems := []string{
		nextOpt.indentFormat() + fmt.Sprintf("code: %s", v.Code),
	}
	if r := v.Rule; r != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("rule: %q", v.Rule.Expr))
	}
	if len(v.Details) != 0 {
		elems = append(elems, v.Details.ProtoFormat(nextOpt))
	}
	return indent + fmt.Sprintf("error {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (v ValidationErrorDetails) ProtoFormat(opt *ProtoFormatOption) string {
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	if len(v) == 1 {
		return indent + fmt.Sprintf("details {\n%s\n%s}", v[0].ProtoFormat(opt), indent)
	}
	var elems []string
	for _, detail := range v {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("{\n%s\n%s}", detail.ProtoFormat(nextOpt), nextOpt.indentFormat()))
	}
	return indent + fmt.Sprintf("details: [\n%s\n%s]", strings.Join(elems, ",\n"), indent)
}

func (v *ValidationErrorDetail) ProtoFormat(opt *ProtoFormatOption) string {
	nextOpt := opt.toNextIndentLevel()
	elems := []string{
		nextOpt.indentFormat() + fmt.Sprintf("rule: %q", v.Rule.Expr),
	}
	if s := len(v.Messages); s != 0 {
		elems = append(elems, v.protoFormatDetails(nextOpt, "message", s))
	}
	if s := len(v.PreconditionFailures); s != 0 {
		elems = append(elems, v.protoFormatDetails(nextOpt, "precondition_failure", s))
	}
	if s := len(v.BadRequests); s != 0 {
		elems = append(elems, v.protoFormatDetails(nextOpt, "bad_request", s))
	}
	if s := len(v.LocalizedMessages); s != 0 {
		elems = append(elems, v.protoFormatDetails(nextOpt, "localized_message", s))
	}

	return strings.Join(elems, "\n")
}

func (v *ValidationErrorDetail) protoFormatDetails(opt *ProtoFormatOption, name string, size int) string {
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	if size == 1 {
		return indent + fmt.Sprintf("%s {...}", name)
	}
	var elems []string
	for i := 0; i < size; i++ {
		elems = append(elems, nextOpt.indentFormat()+"{...}")
	}
	return indent + fmt.Sprintf("%s: [\n%s\n%s]", name, strings.Join(elems, ",\n"), indent)
}

func (a *Argument) ProtoFormat(opt *ProtoFormatOption, isRequestArg bool) string {
	var elems []string
	if a.Name != "" {
		if isRequestArg {
			elems = append(elems, fmt.Sprintf("field: %q", a.Name))
		} else {
			elems = append(elems, fmt.Sprintf("name: %q", a.Name))
		}
	}
	if a.Value != nil {
		elems = append(elems, a.Value.ProtoFormat(opt))
	}
	if len(elems) == 0 {
		return ""
	}
	return fmt.Sprintf("{ %s }", strings.Join(elems, ", "))
}

func (v *Value) ProtoFormat(opt *ProtoFormatOption) string {
	if v == nil {
		return ""
	}
	if v.CEL != nil {
		if v.Inline {
			return fmt.Sprintf("inline: %q", v.CEL.Expr)
		}
		return fmt.Sprintf("by: %q", v.CEL.Expr)
	}
	if v.Const != nil {
		return v.Const.ProtoFormat(opt)
	}
	return ""
}

func (c *ConstValue) ProtoFormat(opt *ProtoFormatOption) string {
	if c == nil {
		return ""
	}
	switch c.Type {
	case DoubleType:
		return fmt.Sprintf(`double: %v`, c.Value)
	case FloatType:
		return fmt.Sprintf(`float: %v`, c.Value)
	case Int32Type:
		return fmt.Sprintf(`int32: %v`, c.Value)
	case Int64Type:
		return fmt.Sprintf(`int64: %v`, c.Value)
	case Uint32Type:
		return fmt.Sprintf(`uint32: %v`, c.Value)
	case Uint64Type:
		return fmt.Sprintf(`uint64: %v`, c.Value)
	case Sint32Type:
		return fmt.Sprintf(`sint32: %v`, c.Value)
	case Sint64Type:
		return fmt.Sprintf(`sint64: %v`, c.Value)
	case Fixed32Type:
		return fmt.Sprintf(`fixed32: %v`, c.Value)
	case Fixed64Type:
		return fmt.Sprintf(`fixed64: %v`, c.Value)
	case Sfixed32Type:
		return fmt.Sprintf(`sfixed32: %v`, c.Value)
	case Sfixed64Type:
		return fmt.Sprintf("sfixed64: %v", c.Value)
	case BoolType:
		return fmt.Sprintf(`bool: %v`, c.Value)
	case StringType:
		return fmt.Sprintf("string: %q", c.Value)
	case BytesType:
		return fmt.Sprintf("byte_string: %q", c.Value)
	case EnvType:
		return fmt.Sprintf("env: %q", c.Value)
	case DoubleRepeatedType:
		var elems []string
		for _, v := range c.Value.([]float64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`doubles: [%s]`, strings.Join(elems, ", "))
	case FloatRepeatedType:
		var elems []string
		for _, v := range c.Value.([]float32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`floats: [%s]`, strings.Join(elems, ", "))
	case Int32RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`int32s: [%s]`, strings.Join(elems, ", "))
	case Int64RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`int64s: [%s]`, strings.Join(elems, ", "))
	case Uint32RepeatedType:
		var elems []string
		for _, v := range c.Value.([]uint32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`uint32s: [%s]`, strings.Join(elems, ", "))
	case Uint64RepeatedType:
		var elems []string
		for _, v := range c.Value.([]uint64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`uint64s: [%s]`, strings.Join(elems, ", "))
	case Sint32RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`sint32s: [%s]`, strings.Join(elems, ", "))
	case Sint64RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`sint64s: [%s]`, strings.Join(elems, ", "))
	case Fixed32RepeatedType:
		var elems []string
		for _, v := range c.Value.([]uint32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`fixed32s: [%s]`, strings.Join(elems, ", "))
	case Fixed64RepeatedType:
		var elems []string
		for _, v := range c.Value.([]uint64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`fixed64s: [%s]`, strings.Join(elems, ", "))
	case Sfixed32RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int32) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`sfixed32s: [%s]`, strings.Join(elems, ", "))
	case Sfixed64RepeatedType:
		var elems []string
		for _, v := range c.Value.([]int64) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`sfixed64s: [%s]`, strings.Join(elems, ", "))
	case BoolRepeatedType:
		var elems []string
		for _, v := range c.Value.([]bool) {
			elems = append(elems, fmt.Sprint(v))
		}
		return fmt.Sprintf(`bools: [%s]`, strings.Join(elems, ", "))
	case StringRepeatedType:
		var elems []string
		for _, v := range c.Value.([]string) {
			elems = append(elems, fmt.Sprintf(`%q`, v))
		}
		return fmt.Sprintf(`strings: [%s]`, strings.Join(elems, ", "))
	case BytesRepeatedType:
		var elems []string
		for _, v := range c.Value.([][]byte) {
			elems = append(elems, fmt.Sprintf(`%q`, string(v)))
		}
		return fmt.Sprintf(`byte_strings: [%s]`, strings.Join(elems, ", "))
	case EnvRepeatedType:
		var elems []string
		for _, v := range c.Value.([]EnvKey) {
			elems = append(elems, fmt.Sprintf(`%q`, v))
		}
		return fmt.Sprintf(`envs: [%s]`, strings.Join(elems, ", "))
	}

	switch c.Type.Type {
	case types.Enum:
		if c.Type.Repeated {
			var elems []string
			for _, v := range c.Value.([]*EnumValue) {
				elems = append(elems, fmt.Sprintf("%q", v.FQDN()))
			}
			return fmt.Sprintf(`enums: [%s]`, strings.Join(elems, ", "))
		}
		return fmt.Sprintf("enum: %q", c.Value.(*EnumValue).FQDN())
	case types.Message:
		msg := c.Type.Ref
		if c.Type.Repeated {
			var elems []string
			for _, v := range c.Value.([]map[string]*Value) {
				var fields []string
				for fieldName, fieldValue := range v {
					fields = append(fields, fmt.Sprintf("{ field: %q, %s }", fieldName, fieldValue.ProtoFormat(opt)))
				}
				sort.Strings(fields)
				elems = append(elems, fmt.Sprintf("{ name: %q, fields: [%s] }", msg.FQDN(), strings.Join(fields, ", ")))
			}
			return fmt.Sprintf("messages: [%s]", strings.Join(elems, ", "))
		}
		var fields []string
		for fieldName, fieldValue := range c.Value.(map[string]*Value) {
			fields = append(fields, fmt.Sprintf("{ field: %q, %s }", fieldName, fieldValue.ProtoFormat(opt)))
		}
		sort.Strings(fields)
		return fmt.Sprintf("message: { name: %q, fields: [%s] }", msg.FQDN(), strings.Join(fields, ", "))
	}
	return ""
}

func DependencyGraphTreeFormat(groups []MessageResolverGroup) string {
	ctx := newMessageResolverGroupTreeFormatContext()
	for _, group := range groups {
		group.setTextMaxLength(ctx.withNextDepth())
	}
	ctx.setupIndent()
	if len(groups) == 1 {
		return groups[0].treeFormat(ctx)
	}
	var ret string
	for i := 0; i < len(groups); i++ {
		if i != 0 {
			ctx = ctx.withLineDepth()
		}
		ret += groups[i].treeFormat(ctx.withNextDepth())
		if i == 0 {
			ret += " ─┐"
		} else {
			ret += " ─┤"
		}
		ret += "\n"
	}
	return ret
}

func (g *SequentialMessageResolverGroup) treeFormat(ctx *messageResolverGroupTreeFormatContext) string {
	var (
		ret string
	)
	if g.Start != nil {
		ret += treeFormatByMessageResolverGroup(ctx, g.Start, true)
	}
	if g.End != nil {
		ret += treeFormatByMessageResolver(ctx, g.End)
	}
	return ret
}

func (g *ConcurrentMessageResolverGroup) treeFormat(ctx *messageResolverGroupTreeFormatContext) string {
	var (
		ret string
	)
	for i := 0; i < len(g.Starts); i++ {
		ret += treeFormatByMessageResolverGroup(ctx, g.Starts[i], i == 0)
	}
	if g.End != nil {
		ret += treeFormatByMessageResolver(ctx, g.End)
	}
	return ret
}

func treeFormatByMessageResolverGroup(ctx *messageResolverGroupTreeFormatContext, g MessageResolverGroup, isFirst bool) string {
	if !isFirst {
		ctx = ctx.withLineDepth()
	}
	text := g.treeFormat(ctx.withNextDepth())
	if isFirst {
		text += " ─┐"
	} else {
		text += " ─┤"
	}
	prevIndent := ctx.currentIndent()
	for _, indent := range ctx.lineIndents() {
		diff := indent - prevIndent - 1
		if diff < 0 {
			prevIndent = indent
			continue
		}
		text += strings.Repeat(" ", diff)
		text += "│"
		prevIndent = indent
	}
	text += "\n"
	return text
}

func treeFormatByMessageResolver(ctx *messageResolverGroupTreeFormatContext, r *MessageResolver) string {
	format := fmt.Sprintf("%%%ds", ctx.currentMaxLength())
	prefix := strings.Repeat(" ", ctx.currentIndent())
	return prefix + fmt.Sprintf(format, r.Name)
}

func (g *SequentialMessageResolverGroup) setTextMaxLength(ctx *messageResolverGroupTreeFormatContext) {
	if g.Start != nil {
		g.Start.setTextMaxLength(ctx.withNextDepth())
	}
	if g.End != nil {
		max := ctx.depthToMaxLength[ctx.depth]
		length := len(g.End.Name)
		if max < length {
			ctx.depthToMaxLength[ctx.depth] = length
		}
	}
}

func (g *ConcurrentMessageResolverGroup) setTextMaxLength(ctx *messageResolverGroupTreeFormatContext) {
	for _, start := range g.Starts {
		start.setTextMaxLength(ctx.withNextDepth())
	}
	if g.End != nil {
		max := ctx.depthToMaxLength[ctx.depth]
		length := len(g.End.Name)
		if max < length {
			ctx.depthToMaxLength[ctx.depth] = length
		}
	}
}

func newMessageResolverGroupTreeFormatContext() *messageResolverGroupTreeFormatContext {
	return &messageResolverGroupTreeFormatContext{
		depthToMaxLength: map[int]int{0: 0},
		depthToIndent:    make(map[int]int),
		lineDepth:        make(map[int]struct{}),
	}
}

const lineSpace = 2

func (c *messageResolverGroupTreeFormatContext) setupIndent() {
	maxDepth := c.maxDepth()
	for depth := range c.depthToMaxLength {
		diff := maxDepth - depth
		c.depthToIndent[depth] = c.getTotalMaxLength(depth+1) + diff*lineSpace
	}
}

func (c *messageResolverGroupTreeFormatContext) getTotalMaxLength(depth int) int {
	length, exists := c.depthToMaxLength[depth]
	if !exists {
		return 0
	}
	return length + c.getTotalMaxLength(depth+1)
}

// withLineDepth clone context after adding '│' character's depth position.
func (c *messageResolverGroupTreeFormatContext) withLineDepth() *messageResolverGroupTreeFormatContext {
	lineDepth := make(map[int]struct{})
	for depth := range c.lineDepth {
		lineDepth[depth] = struct{}{}
	}
	lineDepth[c.depth] = struct{}{}
	return &messageResolverGroupTreeFormatContext{
		depth:            c.depth,
		depthToMaxLength: c.depthToMaxLength,
		depthToIndent:    c.depthToIndent,
		lineDepth:        lineDepth,
	}
}

// withNextDepth clone context after incrementing the depth.
func (c *messageResolverGroupTreeFormatContext) withNextDepth() *messageResolverGroupTreeFormatContext {
	return &messageResolverGroupTreeFormatContext{
		depth:            c.depth + 1,
		depthToMaxLength: c.depthToMaxLength,
		depthToIndent:    c.depthToIndent,
		lineDepth:        c.lineDepth,
	}
}

// maxDepth return max depth number for current tree.
func (c *messageResolverGroupTreeFormatContext) maxDepth() int {
	var max int
	for depth := range c.depthToMaxLength {
		if max < depth {
			max = depth
		}
	}
	return max
}

// lineIndents returns all '│' character's indent position.
func (c *messageResolverGroupTreeFormatContext) lineIndents() []int {
	indents := []int{}
	for depth := range c.lineDepth {
		indents = append(indents, c.depthToIndent[depth])
	}
	sort.Ints(indents)
	return indents
}

// currentIndent return indent at current depth.
func (c *messageResolverGroupTreeFormatContext) currentIndent() int {
	return c.depthToIndent[c.depth]
}

// currentMaxLength return max name length at current depth.
func (c *messageResolverGroupTreeFormatContext) currentMaxLength() int {
	return c.depthToMaxLength[c.depth]
}
