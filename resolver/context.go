package resolver

type context struct {
	errorBuilder      *errorBuilder
	allWarnings       *allWarnings
	fileRef           *File
	svc               *Service
	mtd               *Method
	msg               *Message
	enum              *Enum
	owner             *VariableDefinitionOwner
	plugin            *CELPlugin
	defIdx            int
	depIdx            int
	argIdx            int
	errDetailIdx      int
	pluginTypeIdx     int
	pluginFunctionIdx int
	isPluginMethod    bool
	variableMap       map[string]*VariableDefinition
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
		errorBuilder:      c.errorBuilder,
		allWarnings:       c.allWarnings,
		fileRef:           c.fileRef,
		svc:               c.svc,
		mtd:               c.mtd,
		msg:               c.msg,
		enum:              c.enum,
		owner:             c.owner,
		plugin:            c.plugin,
		defIdx:            c.defIdx,
		depIdx:            c.depIdx,
		argIdx:            c.argIdx,
		pluginTypeIdx:     c.pluginTypeIdx,
		pluginFunctionIdx: c.pluginFunctionIdx,
		isPluginMethod:    c.isPluginMethod,

		errDetailIdx: c.errDetailIdx,
		variableMap:  c.variableMap,
	}
}

func (c *context) withFile(file *File) *context {
	ctx := c.clone()
	ctx.fileRef = file
	return ctx
}

func (c *context) withService(svc *Service) *context {
	ctx := c.clone()
	ctx.svc = svc
	return ctx
}

func (c *context) withMethod(mtd *Method) *context {
	ctx := c.clone()
	ctx.mtd = mtd
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

func (c *context) withDepIndex(idx int) *context {
	ctx := c.clone()
	ctx.depIdx = idx
	return ctx
}

func (c *context) withArgIndex(idx int) *context {
	ctx := c.clone()
	ctx.argIdx = idx
	return ctx
}

func (c *context) withDefOwner(owner *VariableDefinitionOwner) *context {
	ctx := c.clone()
	ctx.owner = owner
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

func (c *context) withPluginTypeIndex(idx int) *context {
	ctx := c.clone()
	ctx.pluginTypeIdx = idx
	return ctx
}

func (c *context) withPluginFunctionIndex(idx int) *context {
	ctx := c.clone()
	ctx.pluginFunctionIdx = idx
	return ctx
}

func (c *context) withPluginIsMethod(v bool) *context {
	ctx := c.clone()
	ctx.isPluginMethod = v
	return ctx
}

func (c *context) withErrDetailIndex(idx int) *context {
	ctx := c.clone()
	ctx.errDetailIdx = idx
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

func (c *context) pluginName() string {
	if c.plugin == nil {
		return ""
	}
	return c.plugin.Name
}

func (c *context) fileName() string {
	if c.fileRef == nil {
		return ""
	}
	return c.fileRef.Name
}

func (c *context) serviceName() string {
	if c.svc == nil {
		return ""
	}
	return c.svc.Name
}

func (c *context) methodName() string {
	if c.mtd == nil {
		return ""
	}
	return c.mtd.Name
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

func (c *context) defOwner() *VariableDefinitionOwner {
	return c.owner
}

func (c *context) defIndex() int {
	return c.defIdx
}

func (c *context) depIndex() int {
	return c.depIdx
}

func (c *context) argIndex() int {
	return c.argIdx
}

func (c *context) pluginTypeIndex() int {
	return c.pluginTypeIdx
}

func (c *context) pluginFunctionIndex() int {
	return c.pluginFunctionIdx
}

func (c *context) pluginIsMethod() bool {
	return c.isPluginMethod
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
