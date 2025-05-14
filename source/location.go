package source

// Location represents semantic location information for grpc federation option.
type Location struct {
	FileName   string
	ImportName string
	Export     *Export
	GoPackage  bool
	Service    *Service
	Message    *Message
	Enum       *Enum
}

type Export struct {
	Name      string
	Wasm      *Wasm
	Types     *PluginType
	Functions *PluginFunction
}

type Wasm struct {
	URL    bool
	Sha256 bool
}

type PluginType struct {
	Idx     int
	Name    bool
	Methods *PluginFunction
}

type PluginFunction struct {
	Idx        int
	Name       bool
	Args       *PluginFunctionArgument
	ReturnType bool
}

type PluginFunctionArgument struct {
	Idx  int
	Type bool
}

// Service represents service location.
type Service struct {
	Name   string
	Method *Method
	Option *ServiceOption
}

// Method represents service's method location.
type Method struct {
	Name     string
	Request  bool
	Response bool
	Option   *MethodOption
}

// ServiceOption represents grpc.federation.service option location.
type ServiceOption struct {
	Env *Env
	Var *ServiceVariable
}

// Env represents grpc.federation.service.env location.
type Env struct {
	Message bool
	Var     *EnvVar
}

// EnvVar represents grpc.federation.service.env.var location.
type EnvVar struct {
	Idx    int
	Name   bool
	Type   bool
	Option *EnvVarOption
}

// EnvVarOption represents grpc.federation.service.env.var.option location.
type EnvVarOption struct {
	Alternate bool
	Default   bool
	Required  bool
	Ignored   bool
}

// ServiceVariable represents grpc.federation.service.var option location.
type ServiceVariable struct {
	Idx        int
	Name       bool
	If         bool
	By         bool
	Map        *MapExprOption
	Message    *MessageExprOption
	Enum       *EnumExprOption
	Validation *ServiceVariableValidationExpr
}

// ServiceVariableValidationExpr represents grpc.federation.service.var.validation option location.
type ServiceVariableValidationExpr struct {
	If      bool
	Message bool
}

// MethodOption represents grpc.federation.method option location.
type MethodOption struct {
	Timeout  bool
	Response bool
}

// Enum represents enum location.
type Enum struct {
	Name   string
	Option *EnumOption
	Value  *EnumValue
}

// EnumOption represents grpc.federation.enum option location.
type EnumOption struct {
	Alias bool
}

// EnumValue represents enum value location.
type EnumValue struct {
	Value  string
	Option *EnumValueOption
}

// EnumValueOption represents grpc.federation.enum_value option location.
type EnumValueOption struct {
	Alias   bool
	Default bool
	Attr    *EnumValueAttribute
}

// EnumValueAttribute represents grpc.federation.enum_value.attr location.
type EnumValueAttribute struct {
	Idx   int
	Name  bool
	Value bool
}

// Message represents message location.
type Message struct {
	Name          string
	Option        *MessageOption
	Field         *Field
	Enum          *Enum
	Oneof         *Oneof
	NestedMessage *Message
}

func (m *Message) MessageNames() []string {
	ret := []string{m.Name}
	if m.NestedMessage != nil {
		ret = append(ret, m.NestedMessage.MessageNames()...)
	}
	return ret
}

func (m *Message) LastNestedMessage() *Message {
	if m.NestedMessage != nil {
		return m.NestedMessage.LastNestedMessage()
	}
	return m
}

// Field represents message field location.
type Field struct {
	Type   bool
	Name   string
	Option *FieldOption
}

// FieldOption represents grpc.federation.field option location.
type FieldOption struct {
	By    bool
	Alias bool
	Oneof *FieldOneof
}

// FieldOneof represents grpc.federation.field.oneof location.
type FieldOneof struct {
	If      bool
	Default bool
	Def     *VariableDefinitionOption
	By      bool
}

type Oneof struct {
	Name   string
	Option *OneofOption
}

type OneofOption struct {
}

// MessageOption represents grpc.federation.message option location.
type MessageOption struct {
	Def   *VariableDefinitionOption
	Alias bool
}

// VariableDefinitionOption represents def location of grpc.federation.message option.
type VariableDefinitionOption struct {
	Idx        int
	Name       bool
	If         bool
	By         bool
	Map        *MapExprOption
	Call       *CallExprOption
	Message    *MessageExprOption
	Enum       *EnumExprOption
	Validation *ValidationExprOption
}

// MapExprOption represents def.map location of grpc.federation.message option.
type MapExprOption struct {
	Iterator *IteratorOption
	By       bool
	Message  *MessageExprOption
	Enum     *EnumExprOption
}

// IteratorOption represents def.map.iterator location of grpc.federation.message option.
type IteratorOption struct {
	Name   bool
	Source bool
}

// CallExprOption represents def.call location of grpc.federation.message option.
type CallExprOption struct {
	Method   bool
	Request  *RequestOption
	Timeout  bool
	Retry    *RetryOption
	Error    *GRPCErrorOption
	Option   *GRPCCallOption
	Metadata bool
}

// GRPCCallOption represents def.call.option of grpc.federation.message option.
type GRPCCallOption struct {
	ContentSubtype     bool
	Header             bool
	Trailer            bool
	MaxCallRecvMsgSize bool
	MaxCallSendMsgSize bool
	StaticMethod       bool
	WaitForReady       bool
}

// MessageExprOption represents def.message location of grpc.federation.message option.
type MessageExprOption struct {
	Name bool
	Args *ArgumentOption
}

// EnumExprOption represents def.enum location of grpc.federation.message option.
type EnumExprOption struct {
	Name bool
	By   bool
}

// RequestOption represents resolver.request location of grpc.federation.message option.
type RequestOption struct {
	Idx   int
	Field bool
	By    bool
	If    bool
}

// RetryOption represents resolver.retry location of grpc.federation.message option.
type RetryOption struct {
	If          bool
	Constant    *RetryConstantOption
	Exponential *RetryExponentialOption
}

// RetryConstantOption represents resolver.retry.constant location of grpc.federation.message option.
type RetryConstantOption struct {
	Interval   bool
	MaxRetries bool
}

// RetryExponentialOption represents resolver.retry.exponential location of grpc.federation.message option.
type RetryExponentialOption struct {
	InitialInterval     bool
	RandomizationFactor bool
	Multiplier          bool
	MaxInterval         bool
	MaxRetries          bool
}

// ArgumentOption represents message argument location of grpc.federation.message option.
type ArgumentOption struct {
	Idx    int
	Name   bool
	By     bool
	Inline bool
}

type ValidationExprOption struct {
	Name  bool
	Error *GRPCErrorOption
}

type GRPCErrorOption struct {
	Idx               int
	Def               *VariableDefinitionOption
	If                bool
	Code              bool
	Message           bool
	Ignore            bool
	IgnoreAndResponse bool
	Detail            *GRPCErrorDetailOption
}

type GRPCErrorDetailOption struct {
	Idx                 int
	Def                 *VariableDefinitionOption
	If                  bool
	By                  bool
	Message             *VariableDefinitionOption
	PreconditionFailure *GRPCErrorDetailPreconditionFailureOption
	BadRequest          *GRPCErrorDetailBadRequestOption
	LocalizedMessage    *GRPCErrorDetailLocalizedMessageOption
}

type GRPCErrorDetailPreconditionFailureOption struct {
	Idx       int
	Violation GRPCErrorDetailPreconditionFailureViolationOption
}

type GRPCErrorDetailPreconditionFailureViolationOption struct {
	Idx       int
	FieldName string
}

type GRPCErrorDetailBadRequestOption struct {
	Idx            int
	FieldViolation GRPCErrorDetailBadRequestFieldViolationOption
}

type GRPCErrorDetailBadRequestFieldViolationOption struct {
	Idx       int
	FieldName string
}

type GRPCErrorDetailLocalizedMessageOption struct {
	Idx       int
	FieldName string
}

// Position represents source position in proto file.
type Position struct {
	Line int
	Col  int
}
