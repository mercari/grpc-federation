package resolver

type context struct {
	allErrors   *Error
	allWarnings *allWarnings
	fileRef     *File
	svc         *Service
	mtd         *Method
	msg         *Message
	enum        *Enum
	depIdx      int
	argIdx      int
}

type allWarnings struct {
	warnings []*Warning
}

func newContext() *context {
	return &context{
		allErrors:   &Error{},
		allWarnings: &allWarnings{},
	}
}

func (c *context) clone() *context {
	return &context{
		allErrors:   c.allErrors,
		allWarnings: c.allWarnings,
		fileRef:     c.fileRef,
		svc:         c.svc,
		mtd:         c.mtd,
		msg:         c.msg,
		enum:        c.enum,
		depIdx:      c.depIdx,
		argIdx:      c.argIdx,
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

func (c *context) file() *File {
	return c.fileRef
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

func (c *context) depIndex() int {
	return c.depIdx
}

func (c *context) argIndex() int {
	return c.argIdx
}

func (c *context) addError(e error) {
	c.allErrors.add(e)
}

func (c *context) error() error {
	if len(c.allErrors.Errs) == 0 {
		return nil
	}

	return c.allErrors
}

func (c *context) addWarning(w *Warning) {
	c.allWarnings.warnings = append(c.allWarnings.warnings, w)
}

func (c *context) warnings() []*Warning {
	return c.allWarnings.warnings
}
