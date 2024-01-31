package source

func NewLocationBuilder(fileName string) *LocationBuilder {
	return &LocationBuilder{
		location: &Location{FileName: fileName},
	}
}

type LocationBuilder struct {
	location *Location
}

func (b *LocationBuilder) WithGoPackage() *LocationBuilder {
	b.location.GoPackage = true
	return b
}

func (b *LocationBuilder) WithExport(name string) *ExportBuilder {
	export := &Export{Name: name}
	b.location.Export = export
	return &ExportBuilder{root: b.location, export: export}
}

func (b *LocationBuilder) WithService(name string) *ServiceBuilder {
	service := &Service{Name: name}
	b.location.Service = service
	return &ServiceBuilder{root: b.location, service: service}
}

func (b *LocationBuilder) WithEnum(name string) *EnumBuilder {
	enum := &Enum{Name: name}
	b.location.Enum = enum
	return &EnumBuilder{root: b.location, enum: enum}
}

func (b *LocationBuilder) WithMessage(name string) *MessageBuilder {
	message := &Message{Name: name}
	b.location.Message = message
	return &MessageBuilder{root: b.location, message: message}
}

func (b *LocationBuilder) Location() *Location {
	return b.location
}

type ExportBuilder struct {
	root   *Location
	export *Export
}

func (b *ExportBuilder) WithWasm() *WasmBuilder {
	w := &Wasm{}
	b.export.Wasm = w
	return &WasmBuilder{root: b.root, wasm: w}
}

func (b *ExportBuilder) WithTypes(idx int) *PluginTypeBuilder {
	p := &PluginType{Idx: idx}
	b.export.Types = p
	return &PluginTypeBuilder{root: b.root, typ: p}
}

func (b *ExportBuilder) WithFunctions(idx int) *PluginFunctionBuilder {
	fn := &PluginFunction{Idx: idx}
	b.export.Functions = fn
	return &PluginFunctionBuilder{root: b.root, fn: fn}
}

func (b *ExportBuilder) Location() *Location {
	return b.root
}

type WasmBuilder struct {
	root *Location
	wasm *Wasm
}

func (b *WasmBuilder) WithURL() *WasmBuilder {
	b.wasm.URL = true
	return b
}

func (b *WasmBuilder) WithSha256() *WasmBuilder {
	b.wasm.Sha256 = true
	return b
}

func (b *WasmBuilder) Location() *Location {
	return b.root
}

type PluginTypeBuilder struct {
	root *Location
	typ  *PluginType
}

func (b *PluginTypeBuilder) WithName() *PluginTypeBuilder {
	b.typ.Name = true
	return b
}

func (b *PluginTypeBuilder) WithMethods(idx int) *PluginFunctionBuilder {
	fn := &PluginFunction{Idx: idx}
	b.typ.Methods = fn
	return &PluginFunctionBuilder{root: b.root, fn: fn}
}

func (b *PluginTypeBuilder) Location() *Location {
	return b.root
}

type PluginFunctionBuilder struct {
	root *Location
	fn   *PluginFunction
}

func (b *PluginFunctionBuilder) WithName() *PluginFunctionBuilder {
	b.fn.Name = true
	return b
}

func (b *PluginFunctionBuilder) WithReturnType() *PluginFunctionBuilder {
	b.fn.ReturnType = true
	return b
}

func (b *PluginFunctionBuilder) WithArgs(idx int) *PluginFunctionArgumentBuilder {
	arg := &PluginFunctionArgument{Idx: idx}
	b.fn.Args = arg
	return &PluginFunctionArgumentBuilder{root: b.root, arg: arg}
}

func (b *PluginFunctionBuilder) Location() *Location {
	return b.root
}

type PluginFunctionArgumentBuilder struct {
	root *Location
	arg  *PluginFunctionArgument
}

func (b *PluginFunctionArgumentBuilder) WithType() *PluginFunctionArgumentBuilder {
	b.arg.Type = true
	return b
}

func (b *PluginFunctionArgumentBuilder) Location() *Location {
	return b.root
}

type ServiceBuilder struct {
	root    *Location
	service *Service
}

func (b *ServiceBuilder) WithMethod(name string) *MethodBuilder {
	m := &Method{Name: name}
	b.service.Method = m
	return &MethodBuilder{root: b.root, method: m}
}

func (b *ServiceBuilder) WithOption() *ServiceOptionBuilder {
	o := &ServiceOption{}
	b.service.Option = o
	return &ServiceOptionBuilder{root: b.root, option: o}
}

func (b *ServiceBuilder) Location() *Location {
	return b.root
}

type MethodBuilder struct {
	root   *Location
	method *Method
}

func (b *MethodBuilder) WithRequest() *MethodBuilder {
	b.method.Request = true
	return b
}

func (b *MethodBuilder) WithResponse() *MethodBuilder {
	b.method.Response = true
	return b
}

func (b *MethodBuilder) WithOption() *MethodOptionBuilder {
	o := &MethodOption{}
	b.method.Option = o
	return &MethodOptionBuilder{root: b.root, option: o}
}

func (b *MethodBuilder) Location() *Location {
	return b.root
}

type MethodOptionBuilder struct {
	root   *Location
	option *MethodOption
}

func (b *MethodOptionBuilder) WithTimeout() *MethodOptionBuilder {
	b.option.Timeout = true
	return b
}

func (b *MethodOptionBuilder) Location() *Location {
	return b.root
}

type ServiceOptionBuilder struct {
	root   *Location
	option *ServiceOption
}

func (b *ServiceOptionBuilder) WithDependencies(idx int) *ServiceDependencyOptionBuilder {
	o := &ServiceDependencyOption{Idx: idx}
	b.option.Dependencies = o
	return &ServiceDependencyOptionBuilder{root: b.root, option: o}
}

type ServiceDependencyOptionBuilder struct {
	root   *Location
	option *ServiceDependencyOption
}

func (b *ServiceDependencyOptionBuilder) WithName() *ServiceDependencyOptionBuilder {
	b.option.Name = true
	return b
}

func (b *ServiceDependencyOptionBuilder) WithService() *ServiceDependencyOptionBuilder {
	b.option.Service = true
	return b
}

func (b *ServiceDependencyOptionBuilder) Location() *Location {
	return b.root
}

func NewEnumBuilder(fileName, msgName, enumName string) *EnumBuilder {
	builder := NewLocationBuilder(fileName)
	if msgName != "" {
		return builder.WithMessage(msgName).WithEnum(enumName)
	}
	return builder.WithEnum(enumName)
}

type EnumBuilder struct {
	root *Location
	enum *Enum
}

func (b *EnumBuilder) WithOption() *EnumBuilder {
	b.enum.Option = &EnumOption{Alias: true}
	return b
}

func (b *EnumBuilder) WithValue(value string) *EnumValueBuilder {
	v := &EnumValue{Value: value}
	b.enum.Value = v
	return &EnumValueBuilder{root: b.root, value: v}
}

func (b *EnumBuilder) Location() *Location {
	return b.root
}

type EnumValueBuilder struct {
	root  *Location
	value *EnumValue
}

func (b *EnumValueBuilder) WithOption() *EnumValueOptionBuilder {
	option := &EnumValueOption{}
	b.value.Option = option
	return &EnumValueOptionBuilder{root: b.root, option: option}
}

func (b *EnumValueBuilder) Location() *Location {
	return b.root
}

type EnumValueOptionBuilder struct {
	root   *Location
	option *EnumValueOption
}

func (b *EnumValueOptionBuilder) WithAlias() *EnumValueOptionBuilder {
	b.option.Alias = true
	return b
}

func (b *EnumValueOptionBuilder) WithDefault() *EnumValueOptionBuilder {
	b.option.Default = true
	return b
}

func (b *EnumValueOptionBuilder) Location() *Location {
	return b.root
}

type MessageBuilder struct {
	root    *Location
	message *Message
}

func (b *MessageBuilder) WithEnum(name string) *EnumBuilder {
	enum := &Enum{Name: name}
	b.message.Enum = enum
	return &EnumBuilder{root: b.root, enum: enum}
}

func (b *MessageBuilder) WithField(name string) *FieldBuilder {
	field := &Field{Name: name}
	b.message.Field = field
	return &FieldBuilder{root: b.root, field: field}
}

func (b *MessageBuilder) WithOption() *MessageOptionBuilder {
	option := &MessageOption{}
	b.message.Option = option
	return &MessageOptionBuilder{root: b.root, option: option}
}

func (b *MessageBuilder) WithOneof(name string) *OneofBuilder {
	oneof := &Oneof{Name: name}
	b.message.Oneof = oneof
	return &OneofBuilder{root: b.root, oneof: oneof}
}

func (b *MessageBuilder) Location() *Location {
	return b.root
}

type FieldBuilder struct {
	root  *Location
	field *Field
}

func (b *FieldBuilder) Location() *Location {
	return b.root
}

func (b *FieldBuilder) WithOption() *FieldOptionBuilder {
	option := &FieldOption{}
	b.field.Option = option
	return &FieldOptionBuilder{root: b.root, option: option}
}

type FieldOptionBuilder struct {
	root   *Location
	option *FieldOption
}

func (b *FieldOptionBuilder) WithBy() *FieldOptionBuilder {
	b.option.By = true
	return b
}

func (b *FieldOptionBuilder) WithAlias() *FieldOptionBuilder {
	b.option.Alias = true
	return b
}

func (b *FieldOptionBuilder) WithOneOf() *FieldOneofBuilder {
	oneof := &FieldOneof{}
	b.option.Oneof = oneof
	return &FieldOneofBuilder{root: b.root, oneof: oneof}
}

func (b *FieldOptionBuilder) Location() *Location {
	return b.root
}

type FieldOneofBuilder struct {
	root  *Location
	oneof *FieldOneof
}

func (b *FieldOneofBuilder) WithIf() *FieldOneofBuilder {
	b.oneof.If = true
	return b
}

func (b *FieldOneofBuilder) WithDefault() *FieldOneofBuilder {
	b.oneof.Default = true
	return b
}

func (b *FieldOneofBuilder) WithBy() *FieldOneofBuilder {
	b.oneof.By = true
	return b
}

func (b *FieldOneofBuilder) WithVariableDefinitions(idx int) *VariableDefinitionOptionBuilder {
	option := &VariableDefinitionOption{Idx: idx}
	b.oneof.VariableDefinitions = option
	return &VariableDefinitionOptionBuilder{root: b.root, option: option}
}

func (b *FieldOneofBuilder) Location() *Location {
	return b.root
}

func NewMsgVarDefOptionBuilder(fileName, msgName string, idx int) *VariableDefinitionOptionBuilder {
	return NewLocationBuilder(fileName).WithMessage(msgName).WithOption().WithVariableDefinitions(idx)
}

type VariableDefinitionOptionBuilder struct {
	root   *Location
	option *VariableDefinitionOption
}

func (b *VariableDefinitionOptionBuilder) WithName() *VariableDefinitionOptionBuilder {
	b.option.Name = true
	return b
}

func (b *VariableDefinitionOptionBuilder) WithIf() *VariableDefinitionOptionBuilder {
	b.option.If = true
	return b
}

func (b *VariableDefinitionOptionBuilder) WithBy() *VariableDefinitionOptionBuilder {
	b.option.By = true
	return b
}

func (b *VariableDefinitionOptionBuilder) WithMessage() *MessageExprOptionBuilder {
	option := &MessageExprOption{}
	b.option.Message = option
	return &MessageExprOptionBuilder{root: b.root, option: option}
}

func (b *VariableDefinitionOptionBuilder) WithCall() *CallExprOptionBuilder {
	option := &CallExprOption{}
	b.option.Call = option
	return &CallExprOptionBuilder{root: b.root, option: option}
}

func (b *VariableDefinitionOptionBuilder) WithValidation() *ValidationExprOptionBuilder {
	option := &ValidationExprOption{}
	b.option.Validation = option
	return &ValidationExprOptionBuilder{root: b.root, option: option}
}

func (b *VariableDefinitionOptionBuilder) WithMap() *MapExprOptionBuilder {
	option := &MapExprOption{}
	b.option.Map = option
	return &MapExprOptionBuilder{root: b.root, option: option}
}

func (b *VariableDefinitionOptionBuilder) Location() *Location {
	return b.root
}

type MessageExprOptionBuilder struct {
	root   *Location
	option *MessageExprOption
}

func (b *MessageExprOptionBuilder) WithName() *MessageExprOptionBuilder {
	b.option.Name = true
	return b
}

func (b *MessageExprOptionBuilder) WithArgs(idx int) *ArgumentOptionBuilder {
	option := &ArgumentOption{Idx: idx}
	b.option.Args = option
	return &ArgumentOptionBuilder{root: b.root, option: option}
}

func (b *MessageExprOptionBuilder) Location() *Location {
	return b.root
}

type ArgumentOptionBuilder struct {
	root   *Location
	option *ArgumentOption
}

func (b *ArgumentOptionBuilder) WithName() *ArgumentOptionBuilder {
	b.option.Name = true
	return b
}

func (b *ArgumentOptionBuilder) WithBy() *ArgumentOptionBuilder {
	b.option.By = true
	return b
}

func (b *ArgumentOptionBuilder) WithInline() *ArgumentOptionBuilder {
	b.option.Inline = true
	return b
}

func (b *ArgumentOptionBuilder) Location() *Location {
	return b.root
}

type MessageOptionBuilder struct {
	root   *Location
	option *MessageOption
}

func (b *MessageOptionBuilder) WithAlias() *MessageOptionBuilder {
	b.option.Alias = true
	return b
}

func (b *MessageOptionBuilder) WithVariableDefinitions(idx int) *VariableDefinitionOptionBuilder {
	option := &VariableDefinitionOption{Idx: idx}
	b.option.VariableDefinitions = option
	return &VariableDefinitionOptionBuilder{root: b.root, option: option}
}

func (b *MessageOptionBuilder) Location() *Location {
	return b.root
}

type OneofBuilder struct {
	root  *Location
	oneof *Oneof
}

func (b *OneofBuilder) Location() *Location {
	return b.root
}

type CallExprOptionBuilder struct {
	root   *Location
	option *CallExprOption
}

func (b *CallExprOptionBuilder) WithMethod() *CallExprOptionBuilder {
	b.option.Method = true
	return b
}

func (b *CallExprOptionBuilder) WithTimeout() *CallExprOptionBuilder {
	b.option.Timeout = true
	return b
}

func (b *CallExprOptionBuilder) WithRetry() *RetryOptionBuilder {
	option := &RetryOption{}
	b.option.Retry = option
	return &RetryOptionBuilder{root: b.root, option: option}
}

func (b *CallExprOptionBuilder) WithRequest(idx int) *RequestOptionBuilder {
	option := &RequestOption{Idx: idx}
	b.option.Request = option
	return &RequestOptionBuilder{root: b.root, option: option}
}

func (b *CallExprOptionBuilder) Location() *Location {
	return b.root
}

type RetryOptionBuilder struct {
	root   *Location
	option *RetryOption
}

func (b *RetryOptionBuilder) WithConstantInterval() *RetryOptionBuilder {
	b.option.Constant = &RetryConstantOption{Interval: true}
	return b
}

func (b *RetryOptionBuilder) WithExponentialInitialInterval() *RetryOptionBuilder {
	b.option.Exponential = &RetryExponentialOption{InitialInterval: true}
	return b
}

func (b *RetryOptionBuilder) WithExponentialMaxInterval() *RetryOptionBuilder {
	b.option.Exponential = &RetryExponentialOption{MaxInterval: true}
	return b
}

func (b *RetryOptionBuilder) Location() *Location {
	return b.root
}

type RequestOptionBuilder struct {
	root   *Location
	option *RequestOption
}

func (b *RequestOptionBuilder) WithField() *RequestOptionBuilder {
	b.option.Field = true
	return b
}

func (b *RequestOptionBuilder) WithBy() *RequestOptionBuilder {
	b.option.By = true
	return b
}

func (b *RequestOptionBuilder) Location() *Location {
	return b.root
}

type ValidationExprOptionBuilder struct {
	root   *Location
	option *ValidationExprOption
}

func (b *ValidationExprOptionBuilder) WithName() *ValidationExprOptionBuilder {
	b.option.Name = true
	return b
}

func (b *ValidationExprOptionBuilder) WithIf() *ValidationExprOptionBuilder {
	b.option.If = true
	return b
}

func (b *ValidationExprOptionBuilder) WithDetail(idx int) *ValidationDetailOptionBuilder {
	option := &ValidationDetailOption{Idx: idx}
	b.option.Detail = option
	return &ValidationDetailOptionBuilder{root: b.root, option: option}
}

func (b *ValidationExprOptionBuilder) Location() *Location {
	return b.root
}

type ValidationDetailOptionBuilder struct {
	root   *Location
	option *ValidationDetailOption
}

func (b *ValidationDetailOptionBuilder) WithIf() *ValidationDetailOptionBuilder {
	b.option.If = true
	return b
}

func (b *ValidationDetailOptionBuilder) WithPreconditionFailure(i1, i2 int, fieldName string) *ValidationDetailOptionBuilder {
	b.option.PreconditionFailure = &ValidationDetailPreconditionFailureOption{
		Idx: i1,
		Violation: ValidationDetailPreconditionFailureViolationOption{
			Idx:       i2,
			FieldName: fieldName,
		},
	}
	return b
}

func (b *ValidationDetailOptionBuilder) WithBadRequest(i1, i2 int, fieldName string) *ValidationDetailOptionBuilder {
	b.option.BadRequest = &ValidationDetailBadRequestOption{
		Idx: i1,
		FieldViolation: ValidationDetailBadRequestFieldViolationOption{
			Idx:       i2,
			FieldName: fieldName,
		},
	}
	return b
}

func (b *ValidationDetailOptionBuilder) WithLocalizedMessage(idx int, fieldName string) *ValidationDetailOptionBuilder {
	b.option.LocalizedMessage = &ValidationDetailLocalizedMessageOption{
		Idx:       idx,
		FieldName: fieldName,
	}
	return b
}

func (b *ValidationDetailOptionBuilder) WithMessage(idx int) *ValidationDetailMessageOptionBuilder {
	option := &ValidationDetailMessageOption{Idx: idx}
	b.option.Message = option
	return &ValidationDetailMessageOptionBuilder{root: b.root, option: option}
}

func (b *ValidationDetailOptionBuilder) Location() *Location {
	return b.root
}

type ValidationDetailMessageOptionBuilder struct {
	root   *Location
	option *ValidationDetailMessageOption
}

func (b *ValidationDetailMessageOptionBuilder) WithMessage() *MessageExprOptionBuilder {
	option := &MessageExprOption{}
	b.option.Message = option
	return &MessageExprOptionBuilder{root: b.root, option: option}
}

func (b *ValidationDetailMessageOptionBuilder) Location() *Location {
	return b.root
}

type MapExprOptionBuilder struct {
	root   *Location
	option *MapExprOption
}

func (b *MapExprOptionBuilder) WithBy() *MapExprOptionBuilder {
	b.option.By = true
	return b
}

func (b *MapExprOptionBuilder) WithIteratorName() *MapExprOptionBuilder {
	b.option.Iterator = &IteratorOption{Name: true}
	return b
}

func (b *MapExprOptionBuilder) WithIteratorSource() *MapExprOptionBuilder {
	b.option.Iterator = &IteratorOption{Source: true}
	return b
}

func (b *MapExprOptionBuilder) WithMessage() *MessageExprOptionBuilder {
	option := &MessageExprOption{}
	b.option.Message = option
	return &MessageExprOptionBuilder{root: b.root, option: option}
}

func (b *MapExprOptionBuilder) Location() *Location {
	return b.root
}
