package resolver

import (
	"time"

	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/types"
)

type Package struct {
	Name  string
	Files Files
}

type File struct {
	Package   *Package
	GoPackage *GoPackage
	Name      string
	Desc      *descriptorpb.FileDescriptorProto
	Services  []*Service
	Messages  []*Message
	Enums     []*Enum
}

type Files []*File

type GoPackage struct {
	Name       string
	ImportPath string
	AliasName  string
}

type Service struct {
	File        *File
	Name        string
	Methods     []*Method
	Rule        *ServiceRule
	Messages    []*Message
	MessageArgs []*Message
}

type Method struct {
	Service  *Service
	Name     string
	Request  *Message
	Response *Message
	Rule     *MethodRule
}

type ServiceRule struct {
	Dependencies []*ServiceDependency
}

type ServiceDependency struct {
	Name    string
	Service *Service
}

type MethodRule struct {
	Timeout *time.Duration
}

type Message struct {
	File           *File
	Name           string
	IsMapEntry     bool
	ParentMessage  *Message
	NestedMessages []*Message
	Enums          []*Enum
	Fields         []*Field
	Oneofs         []*Oneof
	Rule           *MessageRule
}

type Enum struct {
	File    *File
	Name    string
	Values  []*EnumValue
	Message *Message
	Rule    *EnumRule
}

type EnumRule struct {
	Alias *Enum
}

type EnumValue struct {
	Value string
	Enum  *Enum
	Rule  *EnumValueRule
}

type EnumValueRule struct {
	Default bool
	Aliases []*EnumValue
}

type MessageRule struct {
	MessageArgument     *Message
	DependencyGraph     *MessageDependencyGraph
	Resolvers           []MessageResolverGroup
	CustomResolver      bool
	Alias               *Message
	VariableDefinitions VariableDefinitions
}

type VariableDefinition struct {
	Idx      int
	Owner    *VariableDefinitionOwner
	Name     string
	If       *CELValue
	AutoBind bool
	Used     bool
	Expr     *VariableExpr
}

type VariableDefinitions []*VariableDefinition

type VariableExpr struct {
	Type       *Type
	By         *CELValue
	Map        *MapExpr
	Call       *CallExpr
	Message    *MessageExpr
	Validation *ValidationExpr
}

type CallExpr struct {
	Method  *Method
	Request *Request
	Timeout *time.Duration
	Retry   *RetryPolicy
}

type MapExpr struct {
	Iterator *Iterator
	Expr     *MapIteratorExpr
}

type Iterator struct {
	Name   string
	Source *VariableDefinition
}

type MapIteratorExpr struct {
	Type    *Type
	By      *CELValue
	Message *MessageExpr
}

type MessageExpr struct {
	Message *Message
	Args    []*Argument
}

type ValidationExpr struct {
	Error *ValidationError
}

type MessageResolverGroupType string

const (
	SequentialMessageResolverGroupType MessageResolverGroupType = "sequential"
	ConcurrentMessageResolverGroupType MessageResolverGroupType = "concurrent"
)

type MessageResolverGroup interface {
	Type() MessageResolverGroupType
	Resolvers() []*MessageResolver
	treeFormat(*messageResolverGroupTreeFormatContext) string
	setTextMaxLength(*messageResolverGroupTreeFormatContext)
}

type messageResolverGroupTreeFormatContext struct {
	depth            int
	depthToMaxLength map[int]int
	depthToIndent    map[int]int
	lineDepth        map[int]struct{}
}

type SequentialMessageResolverGroup struct {
	Start MessageResolverGroup
	End   *MessageResolver
}

func (g *SequentialMessageResolverGroup) Type() MessageResolverGroupType {
	return SequentialMessageResolverGroupType
}

type ConcurrentMessageResolverGroup struct {
	Starts []MessageResolverGroup
	End    *MessageResolver
}

func (g *ConcurrentMessageResolverGroup) Type() MessageResolverGroupType {
	return ConcurrentMessageResolverGroupType
}

type MessageResolver struct {
	Name               string
	VariableDefinition *VariableDefinition
}

type MessageValidations []*ValidationRule

type ValidationRule struct {
	Name  string
	Error *ValidationError
}

type ValidationError struct {
	Code    code.Code
	Rule    *CELValue
	Details ValidationErrorDetails
}

type ValidationErrorDetails []*ValidationErrorDetail

type ValidationErrorDetail struct {
	Rule                 *CELValue
	PreconditionFailures []*PreconditionFailure
	BadRequests          []*BadRequest
	LocalizedMessages    []*LocalizedMessage
}

type PreconditionFailure struct {
	Violations []*PreconditionFailureViolation
}

type PreconditionFailureViolation struct {
	Type        *CELValue
	Subject     *CELValue
	Description *CELValue
}

type BadRequest struct {
	FieldViolations []*BadRequestFieldViolation
}

type BadRequestFieldViolation struct {
	Field       *CELValue
	Description *CELValue
}

type LocalizedMessage struct {
	Locale  string
	Message *CELValue
}

type TypeConversionDecl struct {
	From *Type
	To   *Type
}

type Oneof struct {
	Name    string
	Message *Message
	Fields  []*Field
}

type Field struct {
	Name  string
	Type  *Type
	Oneof *Oneof
	Rule  *FieldRule
}

type OneofField struct {
	*Field
}

type AutoBindField struct {
	VariableDefinition *VariableDefinition
	Field              *Field
}

type FieldRule struct {
	// Value value to bind to field.
	Value *Value
	// CustomResolver whether `custom_resolver = true` is set in grpc.federation.field option.
	CustomResolver bool
	// MessageCustomResolver whether `custom_resolver = true` is set in grpc.federation.message option.
	MessageCustomResolver bool
	// Alias valid if `alias` is specified in grpc.federation.field option.
	Alias *Field
	// AutoBindField valid if `autobind = true` is specified in resolver.response of grpc.federation.message option.
	AutoBindField *AutoBindField
	// Oneof represents oneof for field option.
	Oneof *FieldOneofRule
}

// FieldOneofRule represents grpc.federation.field.oneof.
type FieldOneofRule struct {
	If                  *CELValue
	Default             bool
	VariableDefinitions VariableDefinitions
	By                  *CELValue
	DependencyGraph     *MessageDependencyGraph
	Resolvers           []MessageResolverGroup
}

type Type struct {
	Type       types.Type
	Repeated   bool
	Ref        *Message
	Enum       *Enum
	OneofField *OneofField
}

func (t *Type) Clone() *Type {
	return &Type{
		Type:     t.Type,
		Repeated: t.Repeated,
		Ref:      t.Ref,
		Enum:     t.Enum,
	}
}

func (t *Type) IsNumber() bool {
	switch t.Type {
	case types.Double,
		types.Float,
		types.Int64,
		types.Uint64,
		types.Int32,
		types.Fixed64,
		types.Fixed32,
		types.Uint32,
		types.Sfixed32,
		types.Sfixed64,
		types.Sint32,
		types.Sint64,
		types.Enum:
		return true
	}
	return false
}

func (t *Type) FQDN() string {
	var repeated string
	if t.Repeated {
		repeated = "repeated "
	}
	if t.OneofField != nil {
		return repeated + t.OneofField.FQDN()
	}
	if t.Ref != nil {
		return repeated + t.Ref.FQDN()
	}
	if t.Enum != nil {
		return repeated + t.Enum.FQDN()
	}
	return repeated + types.ToString(t.Type)
}

type MethodCall struct {
	Method  *Method
	Request *Request
	Timeout *time.Duration
	Retry   *RetryPolicy
}

type RetryPolicy struct {
	Constant    *RetryPolicyConstant
	Exponential *RetryPolicyExponential
}

func (p *RetryPolicy) MaxRetries() uint64 {
	if p.Constant != nil {
		return p.Constant.MaxRetries
	}
	if p.Exponential != nil {
		return p.Exponential.MaxRetries
	}
	return 0
}

type RetryPolicyConstant struct {
	Interval   time.Duration
	MaxRetries uint64
}

type RetryPolicyExponential struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	MaxRetries          uint64
	MaxElapsedTime      time.Duration
}

type Request struct {
	Args []*Argument
	Type *Message
}

type VariableDefinitionOwnerType int

const (
	VariableDefinitionOwnerUnknown    VariableDefinitionOwnerType = 0
	VariableDefinitionOwnerMessage    VariableDefinitionOwnerType = 1
	VariableDefinitionOwnerOneofField VariableDefinitionOwnerType = 2
)

type VariableDefinitionOwner struct {
	Type    VariableDefinitionOwnerType
	Message *Message
	Field   *Field
}

type Argument struct {
	Name  string
	Type  *Type
	Value *Value
}

type Value struct {
	Inline bool
	CEL    *CELValue
	Const  *ConstValue
}

type CELValue struct {
	Expr        string
	Out         *Type
	CheckedExpr *exprv1.CheckedExpr
}

type ConstValue struct {
	Type  *Type
	Value interface{}
}

type EnvKey string

var (
	DoubleType           = &Type{Type: types.Double}
	FloatType            = &Type{Type: types.Float}
	Int32Type            = &Type{Type: types.Int32}
	Int64Type            = &Type{Type: types.Int64}
	Uint32Type           = &Type{Type: types.Uint32}
	Uint64Type           = &Type{Type: types.Uint64}
	Sint32Type           = &Type{Type: types.Sint32}
	Sint64Type           = &Type{Type: types.Sint64}
	Fixed32Type          = &Type{Type: types.Fixed32}
	Fixed64Type          = &Type{Type: types.Fixed64}
	Sfixed32Type         = &Type{Type: types.Sfixed32}
	Sfixed64Type         = &Type{Type: types.Sfixed64}
	BoolType             = &Type{Type: types.Bool}
	StringType           = &Type{Type: types.String}
	BytesType            = &Type{Type: types.Bytes}
	EnumType             = &Type{Type: types.Enum}
	EnvType              = &Type{Type: types.String}
	DoubleRepeatedType   = &Type{Type: types.Double, Repeated: true}
	FloatRepeatedType    = &Type{Type: types.Float, Repeated: true}
	Int32RepeatedType    = &Type{Type: types.Int32, Repeated: true}
	Int64RepeatedType    = &Type{Type: types.Int64, Repeated: true}
	Uint32RepeatedType   = &Type{Type: types.Uint32, Repeated: true}
	Uint64RepeatedType   = &Type{Type: types.Uint64, Repeated: true}
	Sint32RepeatedType   = &Type{Type: types.Sint32, Repeated: true}
	Sint64RepeatedType   = &Type{Type: types.Sint64, Repeated: true}
	Fixed32RepeatedType  = &Type{Type: types.Fixed32, Repeated: true}
	Fixed64RepeatedType  = &Type{Type: types.Fixed64, Repeated: true}
	Sfixed32RepeatedType = &Type{Type: types.Sfixed32, Repeated: true}
	Sfixed64RepeatedType = &Type{Type: types.Sfixed64, Repeated: true}
	BoolRepeatedType     = &Type{Type: types.Bool, Repeated: true}
	StringRepeatedType   = &Type{Type: types.String, Repeated: true}
	BytesRepeatedType    = &Type{Type: types.Bytes, Repeated: true}
	EnumRepeatedType     = &Type{Type: types.Enum, Repeated: true}
	EnvRepeatedType      = &Type{Type: types.String, Repeated: true}
)
