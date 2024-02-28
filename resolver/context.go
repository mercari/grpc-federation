//go:build !tinygo.wasm

package resolver

type context struct {
	errorBuilder         *errorBuilder
	allWarnings          *allWarnings
	fileRef              *File
	msg                  *Message
	enum                 *Enum
	plugin               *CELPlugin
	defIdx               int
	errDetailIdx         int
	ignoreNameValidation bool
	variableMap          map[string]*VariableDefinition
}

type allWarnings struct {
	warnings []*Warning
}

func newContext() *context {
	return &context{
		errorBuilder: &errorBuilder{},
		allWarnings:  &allWarnings{},
		variableMap:  make(map[string]*VariableDefinition),
	}
}

func (c *context) clone() *context {
	return &context{
		errorBuilder:         c.errorBuilder,
		allWarnings:          c.allWarnings,
		fileRef:              c.fileRef,
		msg:                  c.msg,
		enum:                 c.enum,
		plugin:               c.plugin,
		defIdx:               c.defIdx,
		ignoreNameValidation: c.ignoreNameValidation,

		errDetailIdx: c.errDetailIdx,
		variableMap:  c.variableMap,
	}
}

func (c *context) withFile(file *File) *context {
	ctx := c.clone()
	ctx.fileRef = file
	return ctx
}

func (c *context) withMessage(msg *Message) *context {
	ctx := c.clone()
	ctx.msg = msg
	return ctx
}

func (c *context) withEnum(enum *Enum) *context {
	ctx := c.clone()
	ctx.enum = enum
	return ctx
}

func (c *context) withDefIndex(idx int) *context {
	ctx := c.clone()
	ctx.defIdx = idx
	return ctx
}

func (c *context) withPlugin(plugin *CELPlugin) *context {
	ctx := c.clone()
	ctx.plugin = plugin
	return ctx
}

func (c *context) withErrDetailIndex(idx int) *context {
	ctx := c.clone()
	ctx.errDetailIdx = idx
	return ctx
}

func (c *context) withIgnoreNameValidation() *context {
	ctx := c.clone()
	ctx.ignoreNameValidation = true
	return ctx
}

func (c *context) clearVariableDefinitions() {
	c.variableMap = make(map[string]*VariableDefinition)
}

func (c *context) addVariableDefinition(def *VariableDefinition) *context {
	c.variableMap[def.Name] = def
	return c
}

func (c *context) variableDef(name string) *VariableDefinition {
	return c.variableMap[name]
}

func (c *context) file() *File {
	return c.fileRef
}

func (c *context) fileName() string {
	if c.fileRef == nil {
		return ""
	}
	return c.fileRef.Name
}

func (c *context) messageName() string {
	if c.msg == nil {
		return ""
	}
	return c.msg.Name
}

func (c *context) enumName() string {
	if c.enum == nil {
		return ""
	}
	return c.enum.Name
}

func (c *context) defIndex() int {
	return c.defIdx
}

func (c *context) errDetailIndex() int {
	return c.errDetailIdx
}

func (c *context) addError(e *LocationError) {
	c.errorBuilder.add(e)
}

func (c *context) error() error {
	return c.errorBuilder.build()
}

func (c *context) addWarning(w *Warning) {
	c.allWarnings.warnings = append(c.allWarnings.warnings, w)
}

func (c *context) warnings() []*Warning {
	return c.allWarnings.warnings
}
