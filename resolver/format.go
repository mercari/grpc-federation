package resolver

import (
	"fmt"
	"sort"
	"strings"
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
	case r.Value != nil:
		value := r.Value.ProtoFormat(opt)
		value = strings.Replace(value, ":", " =", 1)
		return indent + fmt.Sprintf("(grpc.federation.field).%s", value)
	case len(r.Aliases) != 0:
		// In cases where the output of an alias is needed,
		// we only need to output the first specified alias directly, so it's fine to ignore the other elements.
		return indent + fmt.Sprintf("(grpc.federation.field).alias = %q", r.Aliases[0].Name)
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
	varDefs := r.DefSet.Definitions().ProtoFormat(nextOpt)
	if varDefs != "" {
		elems = append(elems, varDefs)
	}
	if r.CustomResolver {
		elems = append(elems, nextOpt.indentFormat()+"custom_resolver: true")
	}
	var aliases []string
	for _, alias := range r.Aliases {
		aliases = append(aliases, fmt.Sprintf("%q", alias.FQDN()))
	}
	if len(aliases) == 1 {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("alias: %s", aliases[0]))
	} else if len(aliases) > 1 {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("alias: [%s]", strings.Join(aliases, ", ")))
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
	if def.If != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("if: %q", def.If.Expr))
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
	case e.Enum != nil:
		return e.Enum.ProtoFormat(opt)
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

func (e *EnumExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if e.Enum != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("name: %q", e.Enum.FQDN()))
	}
	if e.By != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("by: %q", e.By.Expr))
	}
	if len(elems) == 0 {
		return ""
	}
	return indent + fmt.Sprintf("enum {\n%s\n%s}", strings.Join(elems, "\n"), indent)
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

func (e *ValidationExpr) ProtoFormat(opt *ProtoFormatOption) string {
	if e == nil {
		return ""
	}
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	var elems []string
	if e.Name != "" {
		elems = append(
			elems,
			nextOpt.indentFormat()+fmt.Sprintf("name: %q", e.Name),
		)
	}
	if e.Error != nil {
		elems = append(elems, e.Error.ProtoFormat(nextOpt))
	}
	return indent + fmt.Sprintf("validation {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (e *GRPCError) ProtoFormat(opt *ProtoFormatOption) string {
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	elems := []string{
		nextOpt.indentFormat() + fmt.Sprintf("code: %s", e.Code),
	}
	if val := e.If; val != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("if: %q", val.Expr))
	}
	if m := e.Message; m != nil {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("message: %q", m.Expr))
	}
	if len(e.Details) != 0 {
		elems = append(elems, e.Details.ProtoFormat(nextOpt))
	}
	return indent + fmt.Sprintf("error {\n%s\n%s}", strings.Join(elems, "\n"), indent)
}

func (d GRPCErrorDetails) ProtoFormat(opt *ProtoFormatOption) string {
	indent := opt.indentFormat()
	nextOpt := opt.toNextIndentLevel()
	if len(d) == 1 {
		return indent + fmt.Sprintf("details {\n%s\n%s}", d[0].ProtoFormat(opt), indent)
	}
	var elems []string
	for _, detail := range d {
		elems = append(elems, nextOpt.indentFormat()+fmt.Sprintf("{\n%s\n%s}", detail.ProtoFormat(nextOpt), nextOpt.indentFormat()))
	}
	return indent + fmt.Sprintf("details: [\n%s\n%s]", strings.Join(elems, ",\n"), indent)
}

func (detail *GRPCErrorDetail) ProtoFormat(opt *ProtoFormatOption) string {
	nextOpt := opt.toNextIndentLevel()
	elems := []string{
		nextOpt.indentFormat() + fmt.Sprintf("if: %q", detail.If.Expr),
	}
	if s := len(detail.Messages.Definitions()); s != 0 {
		elems = append(elems, detail.protoFormatDetails(nextOpt, "message", s))
	}
	if s := len(detail.PreconditionFailures); s != 0 {
		elems = append(elems, detail.protoFormatDetails(nextOpt, "precondition_failure", s))
	}
	if s := len(detail.BadRequests); s != 0 {
		elems = append(elems, detail.protoFormatDetails(nextOpt, "bad_request", s))
	}
	if s := len(detail.LocalizedMessages); s != 0 {
		elems = append(elems, detail.protoFormatDetails(nextOpt, "localized_message", s))
	}

	return strings.Join(elems, "\n")
}

func (detail *GRPCErrorDetail) protoFormatDetails(opt *ProtoFormatOption, name string, size int) string {
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
	if a.If != nil {
		elems = append(elems, fmt.Sprintf("if: %q", a.If.Expr))
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
	return ""
}

func DependencyGraphTreeFormat(groups []VariableDefinitionGroup) string {
	ctx := newVariableDefinitionGroupTreeFormatContext()
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

func (g *SequentialVariableDefinitionGroup) treeFormat(ctx *variableDefinitionGroupTreeFormatContext) string {
	var (
		ret string
	)
	if g.Start != nil {
		ret += treeFormatByVariableDefinitionGroup(ctx, g.Start, true)
	}
	if g.End != nil {
		ret += treeFormatByVariableDefinition(ctx, g.End)
	}
	return ret
}

func (g *ConcurrentVariableDefinitionGroup) treeFormat(ctx *variableDefinitionGroupTreeFormatContext) string {
	var (
		ret string
	)
	for i := 0; i < len(g.Starts); i++ {
		ret += treeFormatByVariableDefinitionGroup(ctx, g.Starts[i], i == 0)
	}
	if g.End != nil {
		ret += treeFormatByVariableDefinition(ctx, g.End)
	}
	return ret
}

func treeFormatByVariableDefinitionGroup(ctx *variableDefinitionGroupTreeFormatContext, g VariableDefinitionGroup, isFirst bool) string {
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

func treeFormatByVariableDefinition(ctx *variableDefinitionGroupTreeFormatContext, def *VariableDefinition) string {
	format := fmt.Sprintf("%%%ds", ctx.currentMaxLength())
	prefix := strings.Repeat(" ", ctx.currentIndent())
	return prefix + fmt.Sprintf(format, def.Name)
}

func (g *SequentialVariableDefinitionGroup) setTextMaxLength(ctx *variableDefinitionGroupTreeFormatContext) {
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

func (g *ConcurrentVariableDefinitionGroup) setTextMaxLength(ctx *variableDefinitionGroupTreeFormatContext) {
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

func newVariableDefinitionGroupTreeFormatContext() *variableDefinitionGroupTreeFormatContext {
	return &variableDefinitionGroupTreeFormatContext{
		depthToMaxLength: map[int]int{0: 0},
		depthToIndent:    make(map[int]int),
		lineDepth:        make(map[int]struct{}),
	}
}

const lineSpace = 2

func (c *variableDefinitionGroupTreeFormatContext) setupIndent() {
	maxDepth := c.maxDepth()
	for depth := range c.depthToMaxLength {
		diff := maxDepth - depth
		c.depthToIndent[depth] = c.getTotalMaxLength(depth+1) + diff*lineSpace
	}
}

func (c *variableDefinitionGroupTreeFormatContext) getTotalMaxLength(depth int) int {
	length, exists := c.depthToMaxLength[depth]
	if !exists {
		return 0
	}
	return length + c.getTotalMaxLength(depth+1)
}

// withLineDepth clone context after adding '│' character's depth position.
func (c *variableDefinitionGroupTreeFormatContext) withLineDepth() *variableDefinitionGroupTreeFormatContext {
	lineDepth := make(map[int]struct{})
	for depth := range c.lineDepth {
		lineDepth[depth] = struct{}{}
	}
	lineDepth[c.depth] = struct{}{}
	return &variableDefinitionGroupTreeFormatContext{
		depth:            c.depth,
		depthToMaxLength: c.depthToMaxLength,
		depthToIndent:    c.depthToIndent,
		lineDepth:        lineDepth,
	}
}

// withNextDepth clone context after incrementing the depth.
func (c *variableDefinitionGroupTreeFormatContext) withNextDepth() *variableDefinitionGroupTreeFormatContext {
	return &variableDefinitionGroupTreeFormatContext{
		depth:            c.depth + 1,
		depthToMaxLength: c.depthToMaxLength,
		depthToIndent:    c.depthToIndent,
		lineDepth:        c.lineDepth,
	}
}

// maxDepth return max depth number for current tree.
func (c *variableDefinitionGroupTreeFormatContext) maxDepth() int {
	var max int
	for depth := range c.depthToMaxLength {
		if max < depth {
			max = depth
		}
	}
	return max
}

// lineIndents returns all '│' character's indent position.
func (c *variableDefinitionGroupTreeFormatContext) lineIndents() []int {
	indents := []int{}
	for depth := range c.lineDepth {
		indents = append(indents, c.depthToIndent[depth])
	}
	sort.Ints(indents)
	return indents
}

// currentIndent return indent at current depth.
func (c *variableDefinitionGroupTreeFormatContext) currentIndent() int {
	return c.depthToIndent[c.depth]
}

// currentMaxLength return max name length at current depth.
func (c *variableDefinitionGroupTreeFormatContext) currentMaxLength() int {
	return c.depthToMaxLength[c.depth]
}
