package source

// Location represents semantic location information for grpc federation option.
type Location struct {
	FileName  string
	Export    *Export
	GoPackage bool
	Service   *Service
	Message   *Message
	Enum      *Enum
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
	Dependencies *ServiceDependencyOption
}

// ServiceDependencyOption represents dependencies option of service option.
type ServiceDependencyOption struct {
	Idx     int
	Name    bool
	Service bool
}

// MethodOption represents grpc.federation.method option location.
type MethodOption struct {
	Timeout bool
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
}

// Message represents message location.
type Message struct {
	Name   string
	Option *MessageOption
	Field  *Field
	Enum   *Enum
	Oneof  *Oneof
}

// Field represents message field location.
type Field struct {
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
	If                  bool
	Default             bool
	VariableDefinitions *VariableDefinitionOption
	By                  bool
}

type Oneof struct {
	Name   string
	Option *OneofOption
}

type OneofOption struct {
}

// MessageOption represents grpc.federation.message option location.
type MessageOption struct {
	VariableDefinitions *VariableDefinitionOption
	Alias               bool
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
	Validation *ValidationExprOption
}

// MapExprOption represents def.map location of grpc.federation.message option.
type MapExprOption struct {
	Iterator *IteratorOption
	By       bool
	Message  *MessageExprOption
}

// IteratorOption represents def.map.iterator location of grpc.federation.message option.
type IteratorOption struct {
	Name   bool
	Source bool
}

// CallExprOption represents def.call location of grpc.federation.message option.
type CallExprOption struct {
	Method  bool
	Request *RequestOption
	Timeout bool
	Retry   *RetryOption
}

// MessageExprOption represents def.message location of grpc.federation.message option.
type MessageExprOption struct {
	Name bool
	Args *ArgumentOption
}

// RequestOption represents resolver.request location of grpc.federation.message option.
type RequestOption struct {
	Idx   int
	Field bool
	By    bool
}

// RetryOption represents resolver.retry location of grpc.federation.message option.
type RetryOption struct {
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
	Name   bool
	If     bool
	Detail *ValidationDetailOption
}

type ValidationDetailOption struct {
	Idx                 int
	If                  bool
	Message             *ValidationDetailMessageOption
	PreconditionFailure *ValidationDetailPreconditionFailureOption
	BadRequest          *ValidationDetailBadRequestOption
	LocalizedMessage    *ValidationDetailLocalizedMessageOption
}

type ValidationDetailMessageOption struct {
	Idx     int
	Message *MessageExprOption
}

type ValidationDetailPreconditionFailureOption struct {
	Idx       int
	Violation ValidationDetailPreconditionFailureViolationOption
}

type ValidationDetailPreconditionFailureViolationOption struct {
	Idx       int
	FieldName string
}

type ValidationDetailBadRequestOption struct {
	Idx            int
	FieldViolation ValidationDetailBadRequestFieldViolationOption
}

type ValidationDetailBadRequestFieldViolationOption struct {
	Idx       int
	FieldName string
}

type ValidationDetailLocalizedMessageOption struct {
	Idx       int
	FieldName string
}

// Position represents source position in proto file.
type Position struct {
	Line int
	Col  int
}
