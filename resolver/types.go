package resolver

import (
	"time"

	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/types"
)

type Package struct {
	Name  string
	Files Files
}

type CELPlugin struct {
	Name      string
	Desc      string
	Functions []*CELFunction
}

type CELFunction struct {
	Name     string
	ID       string
	Args     []*Type
	Return   *Type
	Receiver *Message
}

type File struct {
	Package     *Package
	GoPackage   *GoPackage
	Name        string
	Desc        *descriptorpb.FileDescriptorProto
	ImportFiles []*File
	Services    []*Service
	Messages    []*Message
	Enums       []*Enum
	CELPlugins  []*CELPlugin
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
	CELPlugins  []*CELPlugin
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
	MessageArgument *Message
	CustomResolver  bool
	Alias           *Message
	DefSet          *VariableDefinitionSet
}

type VariableDefinitionSet struct {
	Defs   VariableDefinitions
	Groups []VariableDefinitionGroup
	Graph  *MessageDependencyGraph
}

type VariableDefinition struct {
	Idx      int
	Name     string
	If       *CELValue
	AutoBind bool
	Used     bool
	Expr     *VariableExpr
	builder  *source.VariableDefinitionOptionBuilder
}

type VariableDefinitions []*VariableDefinition

type VariableDefinitionGroupType string

const (
	SequentialVariableDefinitionGroupType VariableDefinitionGroupType = "sequential"
	ConcurrentVariableDefinitionGroupType VariableDefinitionGroupType = "concurrent"
)

type VariableDefinitionGroup interface {
	Type() VariableDefinitionGroupType
	VariableDefinitions() VariableDefinitions
	treeFormat(*variableDefinitionGroupTreeFormatContext) string
	setTextMaxLength(*variableDefinitionGroupTreeFormatContext)
}

type variableDefinitionGroupTreeFormatContext struct {
	depth            int
	depthToMaxLength map[int]int
	depthToIndent    map[int]int
	lineDepth        map[int]struct{}
}

type SequentialVariableDefinitionGroup struct {
	Start VariableDefinitionGroup
	End   *VariableDefinition
}

func (g *SequentialVariableDefinitionGroup) Type() VariableDefinitionGroupType {
	return SequentialVariableDefinitionGroupType
}

type ConcurrentVariableDefinitionGroup struct {
	Starts []VariableDefinitionGroup
	End    *VariableDefinition
}

func (g *ConcurrentVariableDefinitionGroup) Type() VariableDefinitionGroupType {
	return ConcurrentVariableDefinitionGroupType
}

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
	Errors  []*GRPCError
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
	Name  string
	Error *GRPCError
}

type GRPCError struct {
	DefSet  *VariableDefinitionSet
	If      *CELValue
	Code    code.Code
	Message string
	Details GRPCErrorDetails
	Ignore  bool
}

type GRPCErrorDetails []*GRPCErrorDetail

type GRPCErrorDetail struct {
	If                   *CELValue
	DefSet               *VariableDefinitionSet
	Messages             *VariableDefinitionSet
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
	Name    string
	Type    *Type
	Oneof   *Oneof
	Rule    *FieldRule
	Message *Message
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
	If      *CELValue
	Default bool
	By      *CELValue
	DefSet  *VariableDefinitionSet
}

type AllMessageDependencyGraph struct {
	Roots []*AllMessageDependencyGraphNode
}

type AllMessageDependencyGraphNode struct {
	Parent   []*AllMessageDependencyGraphNode
	Children []*AllMessageDependencyGraphNode
	Message  *Message
}

type MessageDependencyGraph struct {
	Roots []*MessageDependencyGraphNode
}

type MessageDependencyGraphNode struct {
	Parent             []*MessageDependencyGraphNode
	Children           []*MessageDependencyGraphNode
	ParentMap          map[*MessageDependencyGraphNode]struct{}
	ChildrenMap        map[*MessageDependencyGraphNode]struct{}
	BaseMessage        *Message
	VariableDefinition *VariableDefinition
}

type Type struct {
	Kind       types.Kind
	Repeated   bool
	Message    *Message
	Enum       *Enum
	OneofField *OneofField
}

func (t *Type) Clone() *Type {
	return &Type{
		Kind:       t.Kind,
		Repeated:   t.Repeated,
		Message:    t.Message,
		Enum:       t.Enum,
		OneofField: t.OneofField,
	}
}

func (t *Type) IsNumber() bool {
	switch t.Kind {
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

type GRPCErrorIndexes struct {
	DefIdx       int
	ErrDetailIdx int
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
	DoubleType           = &Type{Kind: types.Double}
	FloatType            = &Type{Kind: types.Float}
	Int32Type            = &Type{Kind: types.Int32}
	Int64Type            = &Type{Kind: types.Int64}
	Uint32Type           = &Type{Kind: types.Uint32}
	Uint64Type           = &Type{Kind: types.Uint64}
	Sint32Type           = &Type{Kind: types.Sint32}
	Sint64Type           = &Type{Kind: types.Sint64}
	Fixed32Type          = &Type{Kind: types.Fixed32}
	Fixed64Type          = &Type{Kind: types.Fixed64}
	Sfixed32Type         = &Type{Kind: types.Sfixed32}
	Sfixed64Type         = &Type{Kind: types.Sfixed64}
	BoolType             = &Type{Kind: types.Bool}
	StringType           = &Type{Kind: types.String}
	BytesType            = &Type{Kind: types.Bytes}
	EnumType             = &Type{Kind: types.Enum}
	EnvType              = &Type{Kind: types.String}
	DoubleRepeatedType   = &Type{Kind: types.Double, Repeated: true}
	FloatRepeatedType    = &Type{Kind: types.Float, Repeated: true}
	Int32RepeatedType    = &Type{Kind: types.Int32, Repeated: true}
	Int64RepeatedType    = &Type{Kind: types.Int64, Repeated: true}
	Uint32RepeatedType   = &Type{Kind: types.Uint32, Repeated: true}
	Uint64RepeatedType   = &Type{Kind: types.Uint64, Repeated: true}
	Sint32RepeatedType   = &Type{Kind: types.Sint32, Repeated: true}
	Sint64RepeatedType   = &Type{Kind: types.Sint64, Repeated: true}
	Fixed32RepeatedType  = &Type{Kind: types.Fixed32, Repeated: true}
	Fixed64RepeatedType  = &Type{Kind: types.Fixed64, Repeated: true}
	Sfixed32RepeatedType = &Type{Kind: types.Sfixed32, Repeated: true}
	Sfixed64RepeatedType = &Type{Kind: types.Sfixed64, Repeated: true}
	BoolRepeatedType     = &Type{Kind: types.Bool, Repeated: true}
	StringRepeatedType   = &Type{Kind: types.String, Repeated: true}
	BytesRepeatedType    = &Type{Kind: types.Bytes, Repeated: true}
	EnumRepeatedType     = &Type{Kind: types.Enum, Repeated: true}
	EnvRepeatedType      = &Type{Kind: types.String, Repeated: true}
)
