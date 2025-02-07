package source

type Locationer interface {
	Location() *Location
}

func NewLocationBuilder(fileName string) *LocationBuilder {
	return &LocationBuilder{
		location: &Location{FileName: fileName},
	}
}

func NewServiceBuilder(fileName, svcName string) *ServiceBuilder {
	return NewLocationBuilder(fileName).WithService(svcName)
}

func NewMessageBuilder(fileName, msgName string) *MessageBuilder {
	return NewLocationBuilder(fileName).WithMessage(msgName)
}

func NewEnumBuilder(fileName, msgName, enumName string) *EnumBuilder {
	builder := NewLocationBuilder(fileName)
	if msgName == "" {
		return builder.WithEnum(enumName)
	}
	return builder.WithMessage(msgName).WithEnum(enumName)
}

type LocationBuilder struct {
	location *Location
}

func (b *LocationBuilder) WithGoPackage() *LocationBuilder {
	loc := b.location.Clone()
	loc.GoPackage = true
	return &LocationBuilder{
		location: loc,
	}
}

func (b *LocationBuilder) WithImportName(name string) *LocationBuilder {
	loc := b.location.Clone()
	loc.ImportName = name
	return &LocationBuilder{
		location: loc,
	}
}

func (b *LocationBuilder) WithExport(name string) *ExportBuilder {
	root := b.location.Clone()
	root.Export = &Export{Name: name}
	return &ExportBuilder{
		root: root,
		export: func(loc *Location) *Export {
			return loc.Export
		},
	}
}

func (b *LocationBuilder) WithService(name string) *ServiceBuilder {
	root := b.location.Clone()
	root.Service = &Service{Name: name}
	return &ServiceBuilder{
		root: root,
		service: func(loc *Location) *Service {
			return loc.Service
		},
	}
}

func (b *LocationBuilder) WithEnum(name string) *EnumBuilder {
	root := b.location.Clone()
	root.Enum = &Enum{Name: name}
	return &EnumBuilder{
		root: root,
		enum: func(loc *Location) *Enum {
			return loc.Enum
		},
	}
}

func (b *LocationBuilder) WithMessage(name string) *MessageBuilder {
	root := b.location.Clone()
	root.Message = &Message{Name: name}
	return &MessageBuilder{
		root: root,
		message: func(loc *Location) *Message {
			return loc.Message
		},
	}
}

func (b *LocationBuilder) Location() *Location {
	return b.location
}

type ExportBuilder struct {
	root   *Location
	export func(*Location) *Export
}

func (b *ExportBuilder) WithWasm() *WasmBuilder {
	root := b.root.Clone()
	export := b.export(root)
	export.Wasm = &Wasm{}
	return &WasmBuilder{
		root: root,
		wasm: func(loc *Location) *Wasm {
			return b.export(loc).Wasm
		},
	}
}

func (b *ExportBuilder) WithTypes(idx int) *PluginTypeBuilder {
	root := b.root.Clone()
	export := b.export(root)
	export.Types = &PluginType{Idx: idx}
	return &PluginTypeBuilder{
		root: root,
		typ: func(loc *Location) *PluginType {
			return b.export(loc).Types
		},
	}
}

func (b *ExportBuilder) WithFunctions(idx int) *PluginFunctionBuilder {
	root := b.root.Clone()
	export := b.export(root)
	export.Functions = &PluginFunction{Idx: idx}
	return &PluginFunctionBuilder{
		root: root,
		fn: func(loc *Location) *PluginFunction {
			return b.export(loc).Functions
		},
	}
}

func (b *ExportBuilder) Location() *Location {
	return b.root
}

type WasmBuilder struct {
	root *Location
	wasm func(*Location) *Wasm
}

func (b *WasmBuilder) WithURL() *WasmBuilder {
	root := b.root.Clone()
	wasm := b.wasm(root)
	wasm.URL = true
	return &WasmBuilder{
		root: root,
		wasm: b.wasm,
	}
}

func (b *WasmBuilder) WithSha256() *WasmBuilder {
	root := b.root.Clone()
	wasm := b.wasm(root)
	wasm.Sha256 = true
	return &WasmBuilder{
		root: root,
		wasm: b.wasm,
	}
}

func (b *WasmBuilder) Location() *Location {
	return b.root
}

type PluginTypeBuilder struct {
	root *Location
	typ  func(*Location) *PluginType
}

func (b *PluginTypeBuilder) WithName() *PluginTypeBuilder {
	root := b.root.Clone()
	typ := b.typ(root)
	typ.Name = true
	return &PluginTypeBuilder{
		root: root,
		typ:  b.typ,
	}
}

func (b *PluginTypeBuilder) WithMethods(idx int) *PluginFunctionBuilder {
	root := b.root.Clone()
	typ := b.typ(root)
	typ.Methods = &PluginFunction{Idx: idx}
	return &PluginFunctionBuilder{
		root: root,
		fn: func(loc *Location) *PluginFunction {
			return b.typ(loc).Methods
		},
	}
}

func (b *PluginTypeBuilder) Location() *Location {
	return b.root
}

type PluginFunctionBuilder struct {
	root *Location
	fn   func(*Location) *PluginFunction
}

func (b *PluginFunctionBuilder) WithName() *PluginFunctionBuilder {
	root := b.root.Clone()
	fn := b.fn(root)
	fn.Name = true
	return &PluginFunctionBuilder{
		root: root,
		fn:   b.fn,
	}
}

func (b *PluginFunctionBuilder) WithReturnType() *PluginFunctionBuilder {
	root := b.root.Clone()
	fn := b.fn(root)
	fn.ReturnType = true
	return &PluginFunctionBuilder{
		root: root,
		fn:   b.fn,
	}
}

func (b *PluginFunctionBuilder) WithArgs(idx int) *PluginFunctionArgumentBuilder {
	root := b.root.Clone()
	fn := b.fn(root)
	fn.Args = &PluginFunctionArgument{Idx: idx}
	return &PluginFunctionArgumentBuilder{
		root: root,
		arg: func(loc *Location) *PluginFunctionArgument {
			return b.fn(loc).Args
		},
	}
}

func (b *PluginFunctionBuilder) Location() *Location {
	return b.root
}

type PluginFunctionArgumentBuilder struct {
	root *Location
	arg  func(*Location) *PluginFunctionArgument
}

func (b *PluginFunctionArgumentBuilder) WithType() *PluginFunctionArgumentBuilder {
	root := b.root.Clone()
	arg := b.arg(root)
	arg.Type = true
	return &PluginFunctionArgumentBuilder{
		root: root,
		arg:  b.arg,
	}
}

func (b *PluginFunctionArgumentBuilder) Location() *Location {
	return b.root
}

type ServiceBuilder struct {
	root    *Location
	service func(*Location) *Service
}

func (b *ServiceBuilder) WithMethod(name string) *MethodBuilder {
	root := b.root.Clone()
	svc := b.service(root)
	svc.Method = &Method{Name: name}
	return &MethodBuilder{
		root: root,
		method: func(loc *Location) *Method {
			return b.service(loc).Method
		},
	}
}

func (b *ServiceBuilder) WithOption() *ServiceOptionBuilder {
	root := b.root.Clone()
	svc := b.service(root)
	svc.Option = &ServiceOption{}
	return &ServiceOptionBuilder{
		root: root,
		option: func(loc *Location) *ServiceOption {
			return b.service(loc).Option
		},
	}
}

func (b *ServiceBuilder) Location() *Location {
	return b.root
}

type MethodBuilder struct {
	root   *Location
	method func(*Location) *Method
}

func (b *MethodBuilder) WithRequest() *MethodBuilder {
	root := b.root.Clone()
	method := b.method(root)
	method.Request = true
	return &MethodBuilder{
		root:   root,
		method: b.method,
	}
}

func (b *MethodBuilder) WithResponse() *MethodBuilder {
	root := b.root.Clone()
	method := b.method(root)
	method.Response = true
	return &MethodBuilder{
		root:   root,
		method: b.method,
	}
}

func (b *MethodBuilder) WithOption() *MethodOptionBuilder {
	root := b.root.Clone()
	method := b.method(root)
	method.Option = &MethodOption{}
	return &MethodOptionBuilder{
		root: root,
		option: func(loc *Location) *MethodOption {
			return b.method(loc).Option
		},
	}
}

func (b *MethodBuilder) Location() *Location {
	return b.root
}

type MethodOptionBuilder struct {
	root   *Location
	option func(*Location) *MethodOption
}

func (b *MethodOptionBuilder) WithTimeout() *MethodOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Timeout = true
	return &MethodOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MethodOptionBuilder) WithResponse() *MethodOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Response = true
	return &MethodOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MethodOptionBuilder) Location() *Location {
	return b.root
}

type ServiceOptionBuilder struct {
	root   *Location
	option func(*Location) *ServiceOption
}

func (b *ServiceOptionBuilder) WithEnv() *EnvBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Env = &Env{}
	return &EnvBuilder{
		root: root,
		env: func(loc *Location) *Env {
			return b.option(loc).Env
		},
	}
}

func (b *ServiceOptionBuilder) WithVar(idx int) *ServiceVariableBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Var = &ServiceVariable{Idx: idx}
	return &ServiceVariableBuilder{
		root: root,
		svcvar: func(loc *Location) *ServiceVariable {
			return b.option(loc).Var
		},
	}
}

func (b *ServiceOptionBuilder) Location() *Location {
	return b.root
}

type EnvBuilder struct {
	root *Location
	env  func(*Location) *Env
}

func (b *EnvBuilder) WithMessage() *EnvBuilder {
	root := b.root.Clone()
	env := b.env(root)
	env.Message = true
	return &EnvBuilder{
		root: root,
		env:  b.env,
	}
}

func (b *EnvBuilder) WithVar(idx int) *EnvVarBuilder {
	root := b.root.Clone()
	env := b.env(root)
	env.Var = &EnvVar{Idx: idx}
	return &EnvVarBuilder{
		root: root,
		envVar: func(loc *Location) *EnvVar {
			return b.env(root).Var
		},
	}
}

func (b *EnvBuilder) Location() *Location {
	return b.root
}

type EnvVarBuilder struct {
	root   *Location
	envVar func(*Location) *EnvVar
}

func (b *EnvVarBuilder) WithName() *EnvVarBuilder {
	root := b.root.Clone()
	envVar := b.envVar(root)
	envVar.Name = true
	return &EnvVarBuilder{
		root:   root,
		envVar: b.envVar,
	}
}

func (b *EnvVarBuilder) WithType() *EnvVarBuilder {
	root := b.root.Clone()
	envVar := b.envVar(root)
	envVar.Type = true
	return &EnvVarBuilder{
		root:   root,
		envVar: b.envVar,
	}
}

func (b *EnvVarBuilder) WithOption() *EnvVarOptionBuilder {
	root := b.root.Clone()
	envVar := b.envVar(root)
	envVar.Option = &EnvVarOption{}
	return &EnvVarOptionBuilder{
		root: root,
		option: func(loc *Location) *EnvVarOption {
			return b.envVar(loc).Option
		},
	}
}

func (b *EnvVarBuilder) Location() *Location {
	return b.root
}

type EnvVarOptionBuilder struct {
	root   *Location
	option func(*Location) *EnvVarOption
}

func (b *EnvVarOptionBuilder) WithAlternate() *EnvVarOptionBuilder {
	root := b.root.Clone()
	opt := b.option(root)
	opt.Alternate = true
	return &EnvVarOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnvVarOptionBuilder) WithDefault() *EnvVarOptionBuilder {
	root := b.root.Clone()
	opt := b.option(root)
	opt.Default = true
	return &EnvVarOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnvVarOptionBuilder) WithRequired() *EnvVarOptionBuilder {
	root := b.root.Clone()
	opt := b.option(root)
	opt.Required = true
	return &EnvVarOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnvVarOptionBuilder) WithIgnored() *EnvVarOptionBuilder {
	root := b.root.Clone()
	opt := b.option(root)
	opt.Ignored = true
	return &EnvVarOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnvVarOptionBuilder) Location() *Location {
	return b.root
}

type ServiceVariableBuilder struct {
	root   *Location
	svcvar func(*Location) *ServiceVariable
}

func (b *ServiceVariableBuilder) WithName() *ServiceVariableBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.Name = true
	return &ServiceVariableBuilder{
		root:   root,
		svcvar: b.svcvar,
	}
}

func (b *ServiceVariableBuilder) WithIf() *ServiceVariableBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.If = true
	return &ServiceVariableBuilder{
		root:   root,
		svcvar: b.svcvar,
	}
}

func (b *ServiceVariableBuilder) WithBy() *ServiceVariableBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.By = true
	return &ServiceVariableBuilder{
		root:   root,
		svcvar: b.svcvar,
	}
}

func (b *ServiceVariableBuilder) WithMap() *MapExprOptionBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.Map = &MapExprOption{}
	return &MapExprOptionBuilder{
		root: root,
		option: func(loc *Location) *MapExprOption {
			return b.svcvar(loc).Map
		},
	}
}

func (b *ServiceVariableBuilder) WithMessage() *MessageExprOptionBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.Message = &MessageExprOption{}
	return &MessageExprOptionBuilder{
		root: root,
		option: func(loc *Location) *MessageExprOption {
			return b.svcvar(loc).Message
		},
	}
}

func (b *ServiceVariableBuilder) WithEnum() *EnumExprOptionBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.Enum = &EnumExprOption{}
	return &EnumExprOptionBuilder{
		root: root,
		option: func(loc *Location) *EnumExprOption {
			return b.svcvar(loc).Enum
		},
	}
}

func (b *ServiceVariableBuilder) WithValidation() *ServiceVariableValidationExprBuilder {
	root := b.root.Clone()
	svcvar := b.svcvar(root)
	svcvar.Validation = &ServiceVariableValidationExpr{}
	return &ServiceVariableValidationExprBuilder{
		root: root,
		option: func(loc *Location) *ServiceVariableValidationExpr {
			return b.svcvar(loc).Validation
		},
	}
}

func (b *ServiceVariableBuilder) Location() *Location {
	return b.root
}

type ServiceVariableValidationExprBuilder struct {
	root   *Location
	option func(*Location) *ServiceVariableValidationExpr
}

func (b *ServiceVariableValidationExprBuilder) WithIf() *ServiceVariableValidationExprBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &ServiceVariableValidationExprBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ServiceVariableValidationExprBuilder) WithMessage() *ServiceVariableValidationExprBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Message = true
	return &ServiceVariableValidationExprBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ServiceVariableValidationExprBuilder) Location() *Location {
	return b.root
}

type EnumBuilder struct {
	root *Location
	enum func(*Location) *Enum
}

func (b *EnumBuilder) WithOption() *EnumBuilder {
	root := b.root.Clone()
	enum := b.enum(root)
	enum.Option = &EnumOption{Alias: true}
	return &EnumBuilder{
		root: root,
		enum: b.enum,
	}
}

func (b *EnumBuilder) WithValue(value string) *EnumValueBuilder {
	root := b.root.Clone()
	enum := b.enum(root)
	enum.Value = &EnumValue{Value: value}
	return &EnumValueBuilder{
		root: root,
		value: func(loc *Location) *EnumValue {
			return b.enum(loc).Value
		},
	}
}

func (b *EnumBuilder) Location() *Location {
	return b.root
}

type EnumValueBuilder struct {
	root  *Location
	value func(*Location) *EnumValue
}

func (b *EnumValueBuilder) WithOption() *EnumValueOptionBuilder {
	root := b.root.Clone()
	value := b.value(root)
	value.Option = &EnumValueOption{}
	return &EnumValueOptionBuilder{
		root: root,
		option: func(loc *Location) *EnumValueOption {
			return b.value(loc).Option
		},
	}
}

func (b *EnumValueBuilder) Location() *Location {
	return b.root
}

type EnumValueOptionBuilder struct {
	root   *Location
	option func(*Location) *EnumValueOption
}

func (b *EnumValueOptionBuilder) WithAlias() *EnumValueOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Alias = true
	return &EnumValueOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnumValueOptionBuilder) WithDefault() *EnumValueOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Default = true
	return &EnumValueOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnumValueOptionBuilder) WithAttr(idx int) *EnumValueAttributeBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Attr = &EnumValueAttribute{Idx: idx}
	return &EnumValueAttributeBuilder{
		root: root,
		attr: func(loc *Location) *EnumValueAttribute {
			return b.option(loc).Attr
		},
	}
}

func (b *EnumValueOptionBuilder) Location() *Location {
	return b.root
}

type EnumValueAttributeBuilder struct {
	root *Location
	attr func(*Location) *EnumValueAttribute
}

func (b *EnumValueAttributeBuilder) WithName() *EnumValueAttributeBuilder {
	root := b.root.Clone()
	attr := b.attr(root)
	attr.Name = true
	return &EnumValueAttributeBuilder{
		root: root,
		attr: b.attr,
	}
}

func (b *EnumValueAttributeBuilder) WithValue() *EnumValueAttributeBuilder {
	root := b.root.Clone()
	attr := b.attr(root)
	attr.Value = true
	return &EnumValueAttributeBuilder{
		root: root,
		attr: b.attr,
	}
}

func (b *EnumValueAttributeBuilder) Location() *Location {
	return b.root
}

type MessageBuilder struct {
	root    *Location
	message func(*Location) *Message
}

func (b *MessageBuilder) WithMessage(name string) *MessageBuilder {
	root := b.root.Clone()
	message := b.message(root)
	message.NestedMessage = &Message{Name: name}
	return &MessageBuilder{
		root: root,
		message: func(loc *Location) *Message {
			return b.message(loc).NestedMessage
		},
	}
}

func (b *MessageBuilder) WithEnum(name string) *EnumBuilder {
	root := b.root.Clone()
	message := b.message(root)
	message.Enum = &Enum{Name: name}
	return &EnumBuilder{
		root: root,
		enum: func(loc *Location) *Enum {
			return b.message(loc).Enum
		},
	}
}

func (b *MessageBuilder) WithField(name string) *FieldBuilder {
	root := b.root.Clone()
	message := b.message(root)
	message.Field = &Field{Name: name}
	return &FieldBuilder{
		root: root,
		field: func(loc *Location) *Field {
			return b.message(loc).Field
		},
	}
}

func (b *MessageBuilder) WithOption() *MessageOptionBuilder {
	root := b.root.Clone()
	message := b.message(root)
	message.Option = &MessageOption{}
	return &MessageOptionBuilder{
		root: root,
		option: func(loc *Location) *MessageOption {
			return b.message(loc).Option
		},
	}
}

func (b *MessageBuilder) WithOneof(name string) *OneofBuilder {
	root := b.root.Clone()
	message := b.message(root)
	message.Oneof = &Oneof{Name: name}
	return &OneofBuilder{
		root: root,
		oneof: func(loc *Location) *Oneof {
			return b.message(loc).Oneof
		},
	}
}

func (b *MessageBuilder) Location() *Location {
	return b.root
}

// ToLazyMessageBuilder sets new Message to the root lazily to return the original Location
// unless MessageBuilder's methods are called afterword.
func ToLazyMessageBuilder(l Locationer, name string) *MessageBuilder {
	root := l.Location().Clone()
	return &MessageBuilder{
		root: root,
		message: func(loc *Location) *Message {
			loc.Message = &Message{Name: name}
			return loc.Message
		},
	}
}

// ToLazyEnumBuilder sets new Enum to the root lazily to return the original Location
// unless EnumBuilder's methods are called afterword.
func ToLazyEnumBuilder(l Locationer, name string) *EnumBuilder {
	root := l.Location().Clone()
	return &EnumBuilder{
		root: root,
		enum: func(loc *Location) *Enum {
			loc.Enum = &Enum{Name: name}
			return loc.Enum
		},
	}
}

type FieldBuilder struct {
	root  *Location
	field func(*Location) *Field
}

func (b *FieldBuilder) Location() *Location {
	return b.root
}

func (b *FieldBuilder) WithOption() *FieldOptionBuilder {
	root := b.root.Clone()
	field := b.field(root)
	field.Option = &FieldOption{}
	return &FieldOptionBuilder{
		root: root,
		option: func(loc *Location) *FieldOption {
			return b.field(loc).Option
		},
	}
}

type FieldOptionBuilder struct {
	root   *Location
	option func(*Location) *FieldOption
}

func (b *FieldOptionBuilder) WithBy() *FieldOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &FieldOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *FieldOptionBuilder) WithAlias() *FieldOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Alias = true
	return &FieldOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *FieldOptionBuilder) WithOneOf() *FieldOneofBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Oneof = &FieldOneof{}
	return &FieldOneofBuilder{
		root: root,
		oneof: func(loc *Location) *FieldOneof {
			return b.option(loc).Oneof
		},
	}
}

func (b *FieldOptionBuilder) Location() *Location {
	return b.root
}

type FieldOneofBuilder struct {
	root  *Location
	oneof func(*Location) *FieldOneof
}

func (b *FieldOneofBuilder) WithIf() *FieldOneofBuilder {
	root := b.root.Clone()
	oneof := b.oneof(root)
	oneof.If = true
	return &FieldOneofBuilder{
		root:  root,
		oneof: b.oneof,
	}
}

func (b *FieldOneofBuilder) WithDefault() *FieldOneofBuilder {
	root := b.root.Clone()
	oneof := b.oneof(root)
	oneof.Default = true
	return &FieldOneofBuilder{
		root:  root,
		oneof: b.oneof,
	}
}

func (b *FieldOneofBuilder) WithBy() *FieldOneofBuilder {
	root := b.root.Clone()
	oneof := b.oneof(root)
	oneof.By = true
	return &FieldOneofBuilder{
		root:  root,
		oneof: b.oneof,
	}
}

func (b *FieldOneofBuilder) WithDef(idx int) *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	oneof := b.oneof(root)
	oneof.Def = &VariableDefinitionOption{Idx: idx}
	return &VariableDefinitionOptionBuilder{
		root: root,
		option: func(loc *Location) *VariableDefinitionOption {
			return b.oneof(loc).Def
		},
	}
}

func (b *FieldOneofBuilder) Location() *Location {
	return b.root
}

type VariableDefinitionOptionBuilder struct {
	root   *Location
	option func(*Location) *VariableDefinitionOption
}

func (b *VariableDefinitionOptionBuilder) WithName() *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Name = true
	return &VariableDefinitionOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *VariableDefinitionOptionBuilder) WithIf() *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &VariableDefinitionOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *VariableDefinitionOptionBuilder) WithBy() *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &VariableDefinitionOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *VariableDefinitionOptionBuilder) WithMessage() *MessageExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Message = &MessageExprOption{}
	return &MessageExprOptionBuilder{
		root: root,
		option: func(loc *Location) *MessageExprOption {
			return b.option(loc).Message
		},
	}
}

func (b *VariableDefinitionOptionBuilder) WithEnum() *EnumExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Enum = &EnumExprOption{}
	return &EnumExprOptionBuilder{
		root: root,
		option: func(loc *Location) *EnumExprOption {
			return b.option(loc).Enum
		},
	}
}

func (b *VariableDefinitionOptionBuilder) WithCall() *CallExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Call = &CallExprOption{}
	return &CallExprOptionBuilder{
		root: root,
		option: func(loc *Location) *CallExprOption {
			return b.option(loc).Call
		},
	}
}

func (b *VariableDefinitionOptionBuilder) WithValidation() *ValidationExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Validation = &ValidationExprOption{}
	return &ValidationExprOptionBuilder{
		root: root,
		option: func(loc *Location) *ValidationExprOption {
			return b.option(loc).Validation
		},
	}
}

func (b *VariableDefinitionOptionBuilder) WithMap() *MapExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Map = &MapExprOption{}
	return &MapExprOptionBuilder{
		root: root,
		option: func(loc *Location) *MapExprOption {
			return b.option(loc).Map
		},
	}
}

func (b *VariableDefinitionOptionBuilder) Location() *Location {
	return b.root
}

type MessageExprOptionBuilder struct {
	root   *Location
	option func(*Location) *MessageExprOption
}

func (b *MessageExprOptionBuilder) WithName() *MessageExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Name = true
	return &MessageExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MessageExprOptionBuilder) WithArgs(idx int) *ArgumentOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Args = &ArgumentOption{Idx: idx}
	return &ArgumentOptionBuilder{
		root: root,
		option: func(loc *Location) *ArgumentOption {
			return b.option(loc).Args
		},
	}
}

func (b *MessageExprOptionBuilder) Location() *Location {
	return b.root
}

type EnumExprOptionBuilder struct {
	root   *Location
	option func(*Location) *EnumExprOption
}

func (b *EnumExprOptionBuilder) WithName() *EnumExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Name = true
	return &EnumExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnumExprOptionBuilder) WithBy() *EnumExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &EnumExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *EnumExprOptionBuilder) Location() *Location {
	return b.root
}

type ArgumentOptionBuilder struct {
	root   *Location
	option func(*Location) *ArgumentOption
}

func (b *ArgumentOptionBuilder) WithName() *ArgumentOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Name = true
	return &ArgumentOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ArgumentOptionBuilder) WithBy() *ArgumentOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &ArgumentOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ArgumentOptionBuilder) WithInline() *ArgumentOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Inline = true
	return &ArgumentOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ArgumentOptionBuilder) Location() *Location {
	return b.root
}

type MessageOptionBuilder struct {
	root   *Location
	option func(*Location) *MessageOption
}

func (b *MessageOptionBuilder) WithAlias() *MessageOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Alias = true
	return &MessageOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MessageOptionBuilder) WithDef(idx int) *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Def = &VariableDefinitionOption{Idx: idx}
	return &VariableDefinitionOptionBuilder{
		root: root,
		option: func(loc *Location) *VariableDefinitionOption {
			return b.option(loc).Def
		},
	}
}

func (b *MessageOptionBuilder) Location() *Location {
	return b.root
}

type OneofBuilder struct {
	root  *Location
	oneof func(*Location) *Oneof
}

func (b *OneofBuilder) Location() *Location {
	return b.root
}

type CallExprOptionBuilder struct {
	root   *Location
	option func(*Location) *CallExprOption
}

func (b *CallExprOptionBuilder) WithMethod() *CallExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Method = true
	return &CallExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *CallExprOptionBuilder) WithTimeout() *CallExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Timeout = true
	return &CallExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *CallExprOptionBuilder) WithRetry() *RetryOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Retry = &RetryOption{}
	return &RetryOptionBuilder{
		root: root,
		option: func(loc *Location) *RetryOption {
			return b.option(loc).Retry
		},
	}
}

func (b *CallExprOptionBuilder) WithRequest(idx int) *RequestOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Request = &RequestOption{Idx: idx}
	return &RequestOptionBuilder{
		root: root,
		option: func(loc *Location) *RequestOption {
			return b.option(loc).Request
		},
	}
}

func (b *CallExprOptionBuilder) WithError(idx int) *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Error = &GRPCErrorOption{Idx: idx}
	return &GRPCErrorOptionBuilder{
		root: root,
		option: func(loc *Location) *GRPCErrorOption {
			return b.option(loc).Error
		},
	}
}

func (b *CallExprOptionBuilder) Location() *Location {
	return b.root
}

type RetryOptionBuilder struct {
	root   *Location
	option func(*Location) *RetryOption
}

func (b *RetryOptionBuilder) WithIf() *RetryOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &RetryOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RetryOptionBuilder) WithConstantInterval() *RetryOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Constant = &RetryConstantOption{Interval: true}
	return &RetryOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RetryOptionBuilder) WithExponentialInitialInterval() *RetryOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Exponential = &RetryExponentialOption{InitialInterval: true}
	return &RetryOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RetryOptionBuilder) WithExponentialMaxInterval() *RetryOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Exponential = &RetryExponentialOption{MaxInterval: true}
	return &RetryOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RetryOptionBuilder) Location() *Location {
	return b.root
}

type RequestOptionBuilder struct {
	root   *Location
	option func(*Location) *RequestOption
}

func (b *RequestOptionBuilder) WithField() *RequestOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Field = true
	return &RequestOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RequestOptionBuilder) WithBy() *RequestOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &RequestOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RequestOptionBuilder) WithIf() *RequestOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &RequestOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *RequestOptionBuilder) Location() *Location {
	return b.root
}

type ValidationExprOptionBuilder struct {
	root   *Location
	option func(*Location) *ValidationExprOption
}

func (b *ValidationExprOptionBuilder) WithName() *ValidationExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Name = true
	return &ValidationExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *ValidationExprOptionBuilder) WithError() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Error = &GRPCErrorOption{Idx: 0}
	return &GRPCErrorOptionBuilder{
		root: root,
		option: func(loc *Location) *GRPCErrorOption {
			return b.option(loc).Error
		},
	}
}

type GRPCErrorOptionBuilder struct {
	root   *Location
	option func(*Location) *GRPCErrorOption
}

func (b *GRPCErrorOptionBuilder) WithDef(idx int) *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Def = &VariableDefinitionOption{Idx: idx}
	return &VariableDefinitionOptionBuilder{
		root: root,
		option: func(loc *Location) *VariableDefinitionOption {
			return b.option(loc).Def
		},
	}
}

func (b *GRPCErrorOptionBuilder) WithIf() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &GRPCErrorOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorOptionBuilder) WithCode() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Code = true
	return &GRPCErrorOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorOptionBuilder) WithMessage() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Message = true
	return &GRPCErrorOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorOptionBuilder) WithIgnore() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Ignore = true
	return &GRPCErrorOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorOptionBuilder) WithIgnoreAndResponse() *GRPCErrorOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.IgnoreAndResponse = true
	return &GRPCErrorOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorOptionBuilder) WithDetail(idx int) *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Detail = &GRPCErrorDetailOption{Idx: idx}
	return &GRPCErrorDetailOptionBuilder{
		root: root,
		option: func(loc *Location) *GRPCErrorDetailOption {
			return b.option(loc).Detail
		},
	}
}

func (b *GRPCErrorOptionBuilder) Location() *Location {
	return b.root
}

type GRPCErrorDetailOptionBuilder struct {
	root   *Location
	option func(*Location) *GRPCErrorDetailOption
}

func (b *GRPCErrorDetailOptionBuilder) WithDef(idx int) *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Def = &VariableDefinitionOption{Idx: idx}
	return &VariableDefinitionOptionBuilder{
		root: root,
		option: func(loc *Location) *VariableDefinitionOption {
			return b.option(loc).Def
		},
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithMessage(idx int) *VariableDefinitionOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Message = &VariableDefinitionOption{Idx: idx}
	return &VariableDefinitionOptionBuilder{
		root: root,
		option: func(loc *Location) *VariableDefinitionOption {
			return b.option(loc).Message
		},
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithIf() *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.If = true
	return &GRPCErrorDetailOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithBy() *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &GRPCErrorDetailOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithPreconditionFailure(i1, i2 int, fieldName string) *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.PreconditionFailure = &GRPCErrorDetailPreconditionFailureOption{
		Idx: i1,
		Violation: GRPCErrorDetailPreconditionFailureViolationOption{
			Idx:       i2,
			FieldName: fieldName,
		},
	}
	return &GRPCErrorDetailOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithBadRequest(i1, i2 int, fieldName string) *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.BadRequest = &GRPCErrorDetailBadRequestOption{
		Idx: i1,
		FieldViolation: GRPCErrorDetailBadRequestFieldViolationOption{
			Idx:       i2,
			FieldName: fieldName,
		},
	}
	return &GRPCErrorDetailOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorDetailOptionBuilder) WithLocalizedMessage(idx int, fieldName string) *GRPCErrorDetailOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.LocalizedMessage = &GRPCErrorDetailLocalizedMessageOption{
		Idx:       idx,
		FieldName: fieldName,
	}
	return &GRPCErrorDetailOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *GRPCErrorDetailOptionBuilder) Location() *Location {
	return b.root
}

type MapExprOptionBuilder struct {
	root   *Location
	option func(*Location) *MapExprOption
}

func (b *MapExprOptionBuilder) WithBy() *MapExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.By = true
	return &MapExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MapExprOptionBuilder) WithIteratorName() *MapExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Iterator = &IteratorOption{Name: true}
	return &MapExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MapExprOptionBuilder) WithIteratorSource() *MapExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Iterator = &IteratorOption{Source: true}
	return &MapExprOptionBuilder{
		root:   root,
		option: b.option,
	}
}

func (b *MapExprOptionBuilder) WithMessage() *MessageExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Message = &MessageExprOption{}
	return &MessageExprOptionBuilder{
		root: root,
		option: func(loc *Location) *MessageExprOption {
			return b.option(loc).Message
		},
	}
}

func (b *MapExprOptionBuilder) WithEnum() *EnumExprOptionBuilder {
	root := b.root.Clone()
	option := b.option(root)
	option.Enum = &EnumExprOption{}
	return &EnumExprOptionBuilder{
		root: root,
		option: func(loc *Location) *EnumExprOption {
			return b.option(loc).Enum
		},
	}
}

func (b *MapExprOptionBuilder) Location() *Location {
	return b.root
}
