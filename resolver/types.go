package resolver

import (
	"time"

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
	Rule  *EnumValueRule
}

type EnumValueRule struct {
	Default bool
	Aliases []*EnumValue
}

type MessageRule struct {
	MethodCall          *MethodCall
	MessageDependencies MessageDependencies
	MessageArgument     *Message
	DependencyGraph     *MessageRuleDependencyGraph
	Resolvers           []MessageResolverGroup
	CustomResolver      bool
	Alias               *Message
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
	Name              string
	MethodCall        *MethodCall
	MessageDependency *MessageDependency
}

type TypeConversionDecl struct {
	From *Type
	To   *Type
}

type Field struct {
	Name string
	Type *Type
	Rule *FieldRule
}

type AutoBindField struct {
	ResponseField     *ResponseField
	MessageDependency *MessageDependency
	Field             *Field
}

type FieldRule struct {
	// Value valid if `by` or `literal` is specified in grpc.federation.field option.
	Value *Value
	// CustomResolver whether `custom_resolver = true` is set in grpc.federation.field option.
	CustomResolver bool
	// MessageCustomResolver whether `custom_resolver = true` is set in grpc.federation.message option.
	MessageCustomResolver bool
	// Alias valid if `alias` is specified in grpc.federation.field option.
	Alias *Field
	// AutoBindField valid if `autobind = true` is specified in resolver.response of grpc.federation.message option.
	AutoBindField *AutoBindField
}

type Type struct {
	Type     types.Type
	Repeated bool
	Ref      *Message
	Enum     *Enum
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
		types.Sint64:
		return true
	}
	return false
}

func (t *Type) FQDN() string {
	var repeated string
	if t.Repeated {
		repeated = "repeated "
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
	Method   *Method
	Request  *Request
	Response *Response
	Timeout  *time.Duration
	Retry    *RetryPolicy
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

type Response struct {
	Fields []*ResponseField
	Type   *Message
}

type ResponseField struct {
	Name      string
	FieldName string
	Type      *Type
	AutoBind  bool
	Used      bool
}

type MessageDependency struct {
	Name     string
	Message  *Message
	Args     []*Argument
	AutoBind bool
	Used     bool
}

type MessageDependencies []*MessageDependency

type Argument struct {
	Name  string
	Type  *Type
	Value *Value
}

type Value struct {
	Path     Path
	PathType PathType
	// Whether fields should be expanded when referencing types defined with Filtered.
	Inline bool
	// Literal value
	Literal *Literal
	// A reference to the value specified in Path.
	// If the value is defined in response, it will be the value before filtering by response.field.
	Ref *Type
	// Filtered the value of the field after filtering
	// using Path from the message defined by Ref.
	Filtered *Type
}

type Literal struct {
	Type  *Type
	Value interface{}
}

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
)
