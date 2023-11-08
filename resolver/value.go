package resolver

import (
	"github.com/mercari/grpc-federation/grpc/federation"
	"github.com/mercari/grpc-federation/types"
)

func (v *Value) Type() *Type {
	if v == nil {
		return nil
	}
	if v.CEL != nil {
		return v.CEL.Out
	}
	if v.Const != nil {
		return v.Const.Type
	}
	return nil
}

func (v *Value) ReferenceNames() []string {
	if v == nil {
		return nil
	}
	return v.CEL.ReferenceNames()
}

func (v *CELValue) ReferenceNames() []string {
	if v == nil {
		return nil
	}
	if v.CheckedExpr == nil {
		return nil
	}

	var refNames []string
	for _, ref := range v.CheckedExpr.ReferenceMap {
		if ref.Name == federation.MessageArgumentVariableName {
			continue
		}
		// Name can be empty sting if the reference points to a function
		if ref.Name == "" {
			continue
		}
		refNames = append(refNames, ref.Name)
	}
	return refNames
}

type commonValueDef struct {
	CustomResolver *bool
	By             *string
	Inline         *string
	Alias          *string
	Double         *float64
	Doubles        []float64
	Float          *float32
	Floats         []float32
	Int32          *int32
	Int32S         []int32
	Int64          *int64
	Int64S         []int64
	Uint32         *uint32
	Uint32S        []uint32
	Uint64         *uint64
	Uint64S        []uint64
	Sint32         *int32
	Sint32S        []int32
	Sint64         *int64
	Sint64S        []int64
	Fixed32        *uint32
	Fixed32S       []uint32
	Fixed64        *uint64
	Fixed64S       []uint64
	Sfixed32       *int32
	Sfixed32S      []int32
	Sfixed64       *int64
	Sfixed64S      []int64
	Bool           *bool
	Bools          []bool
	String         *string
	Strings        []string
	ByteString     []byte
	ByteStrings    [][]byte
	Message        *federation.MessageValue
	Messages       []*federation.MessageValue
	Enum           *string
	Enums          []string
	Env            *string
	Envs           []string
}

func (def *commonValueDef) GetBy() string {
	if def != nil && def.By != nil {
		return *def.By
	}
	return ""
}

func (def *commonValueDef) GetInline() string {
	if def != nil && def.Inline != nil {
		return *def.Inline
	}
	return ""
}

func (def *commonValueDef) GetAlias() string {
	if def != nil && def.Alias != nil {
		return *def.Alias
	}
	return ""
}

func (def *commonValueDef) GetDouble() float64 {
	if def != nil && def.Double != nil {
		return *def.Double
	}
	return 0
}

func (def *commonValueDef) GetDoubles() []float64 {
	if def != nil {
		return def.Doubles
	}
	return nil
}

func (def *commonValueDef) GetFloat() float32 {
	if def != nil && def.Float != nil {
		return *def.Float
	}
	return 0
}

func (def *commonValueDef) GetFloats() []float32 {
	if def != nil {
		return def.Floats
	}
	return nil
}

func (def *commonValueDef) GetInt32() int32 {
	if def != nil && def.Int32 != nil {
		return *def.Int32
	}
	return 0
}

func (def *commonValueDef) GetInt32S() []int32 {
	if def != nil {
		return def.Int32S
	}
	return nil
}

func (def *commonValueDef) GetInt64() int64 {
	if def != nil && def.Int64 != nil {
		return *def.Int64
	}
	return 0
}

func (def *commonValueDef) GetInt64S() []int64 {
	if def != nil {
		return def.Int64S
	}
	return nil
}

func (def *commonValueDef) GetUint32() uint32 {
	if def != nil && def.Uint32 != nil {
		return *def.Uint32
	}
	return 0
}

func (def *commonValueDef) GetUint32S() []uint32 {
	if def != nil {
		return def.Uint32S
	}
	return nil
}

func (def *commonValueDef) GetUint64() uint64 {
	if def != nil && def.Uint64 != nil {
		return *def.Uint64
	}
	return 0
}

func (def *commonValueDef) GetUint64S() []uint64 {
	if def != nil {
		return def.Uint64S
	}
	return nil
}

func (def *commonValueDef) GetSint32() int32 {
	if def != nil && def.Sint32 != nil {
		return *def.Sint32
	}
	return 0
}

func (def *commonValueDef) GetSint32S() []int32 {
	if def != nil {
		return def.Sint32S
	}
	return nil
}

func (def *commonValueDef) GetSint64() int64 {
	if def != nil && def.Sint64 != nil {
		return *def.Sint64
	}
	return 0
}

func (def *commonValueDef) GetSint64S() []int64 {
	if def != nil {
		return def.Sint64S
	}
	return nil
}

func (def *commonValueDef) GetFixed32() uint32 {
	if def != nil && def.Fixed32 != nil {
		return *def.Fixed32
	}
	return 0
}

func (def *commonValueDef) GetFixed32S() []uint32 {
	if def != nil {
		return def.Fixed32S
	}
	return nil
}

func (def *commonValueDef) GetFixed64() uint64 {
	if def != nil && def.Fixed64 != nil {
		return *def.Fixed64
	}
	return 0
}

func (def *commonValueDef) GetFixed64S() []uint64 {
	if def != nil {
		return def.Fixed64S
	}
	return nil
}

func (def *commonValueDef) GetSfixed32() int32 {
	if def != nil && def.Sfixed32 != nil {
		return *def.Sfixed32
	}
	return 0
}

func (def *commonValueDef) GetSfixed32S() []int32 {
	if def != nil {
		return def.Sfixed32S
	}
	return nil
}

func (def *commonValueDef) GetSfixed64() int64 {
	if def != nil && def.Sfixed64 != nil {
		return *def.Sfixed64
	}
	return 0
}

func (def *commonValueDef) GetSfixed64S() []int64 {
	if def != nil {
		return def.Sfixed64S
	}
	return nil
}

func (def *commonValueDef) GetBool() bool {
	if def != nil && def.Bool != nil {
		return *def.Bool
	}
	return false
}

func (def *commonValueDef) GetBools() []bool {
	if def != nil {
		return def.Bools
	}
	return nil
}

func (def *commonValueDef) GetString() string {
	if def != nil && def.String != nil {
		return *def.String
	}
	return ""
}

func (def *commonValueDef) GetStrings() []string {
	if def != nil {
		return def.Strings
	}
	return nil
}

func (def *commonValueDef) GetByteString() []byte {
	if def != nil {
		return def.ByteString
	}
	return nil
}

func (def *commonValueDef) GetByteStrings() [][]byte {
	if def != nil {
		return def.ByteStrings
	}
	return nil
}

func (def *commonValueDef) GetMessage() *federation.MessageValue {
	if def != nil {
		return def.Message
	}
	return nil
}

func (def *commonValueDef) GetMessages() []*federation.MessageValue {
	if def != nil {
		return def.Messages
	}
	return nil
}

func (def *commonValueDef) GetEnum() string {
	if def != nil && def.Enum != nil {
		return *def.Enum
	}
	return ""
}

func (def *commonValueDef) GetEnums() []string {
	if def != nil {
		return def.Enums
	}
	return nil
}

func (def *commonValueDef) GetEnv() string {
	if def != nil && def.Env != nil {
		return *def.Env
	}
	return ""
}

func (def *commonValueDef) GetEnvs() []string {
	if def != nil {
		return def.Envs
	}
	return nil
}

func fieldRuleToCommonValueDef(def *federation.FieldRule) *commonValueDef {
	return &commonValueDef{
		CustomResolver: def.CustomResolver,
		By:             def.By,
		Alias:          def.Alias,
		Double:         def.Double,
		Doubles:        def.Doubles,
		Float:          def.Float,
		Floats:         def.Floats,
		Int32:          def.Int32,
		Int32S:         def.Int32S,
		Int64:          def.Int64,
		Int64S:         def.Int64S,
		Uint32:         def.Uint32,
		Uint32S:        def.Uint32S,
		Uint64:         def.Uint64,
		Uint64S:        def.Uint64S,
		Sint32:         def.Sint32,
		Sint32S:        def.Sint32S,
		Sint64:         def.Sint64,
		Sint64S:        def.Sint64S,
		Fixed32:        def.Fixed32,
		Fixed32S:       def.Fixed32S,
		Fixed64:        def.Fixed64,
		Fixed64S:       def.Fixed64S,
		Sfixed32:       def.Sfixed32,
		Sfixed32S:      def.Sfixed32S,
		Sfixed64:       def.Sfixed64,
		Sfixed64S:      def.Sfixed64S,
		Bool:           def.Bool,
		Bools:          def.Bools,
		String:         def.String_,
		Strings:        def.Strings,
		ByteString:     def.ByteString,
		ByteStrings:    def.ByteStrings,
		Message:        def.Message,
		Messages:       def.Messages,
		Enum:           def.Enum,
		Enums:          def.Enums,
		Env:            def.Env,
		Envs:           def.Envs,
	}
}

func methodRequestToCommonValueDef(def *federation.MethodRequest) *commonValueDef {
	return &commonValueDef{
		By:          def.By,
		Double:      def.Double,
		Doubles:     def.Doubles,
		Float:       def.Float,
		Floats:      def.Floats,
		Int32:       def.Int32,
		Int32S:      def.Int32S,
		Int64:       def.Int64,
		Int64S:      def.Int64S,
		Uint32:      def.Uint32,
		Uint32S:     def.Uint32S,
		Uint64:      def.Uint64,
		Uint64S:     def.Uint64S,
		Sint32:      def.Sint32,
		Sint32S:     def.Sint32S,
		Sint64:      def.Sint64,
		Sint64S:     def.Sint64S,
		Fixed32:     def.Fixed32,
		Fixed32S:    def.Fixed32S,
		Fixed64:     def.Fixed64,
		Fixed64S:    def.Fixed64S,
		Sfixed32:    def.Sfixed32,
		Sfixed32S:   def.Sfixed32S,
		Sfixed64:    def.Sfixed64,
		Sfixed64S:   def.Sfixed64S,
		Bool:        def.Bool,
		Bools:       def.Bools,
		String:      def.String_,
		Strings:     def.Strings,
		ByteString:  def.ByteString,
		ByteStrings: def.ByteStrings,
		Message:     def.Message,
		Messages:    def.Messages,
		Enum:        def.Enum,
		Enums:       def.Enums,
		Env:         def.Env,
		Envs:        def.Envs,
	}
}

func argumentToCommonValueDef(def *federation.Argument) *commonValueDef {
	return &commonValueDef{
		By:          def.By,
		Inline:      def.Inline,
		Double:      def.Double,
		Doubles:     def.Doubles,
		Float:       def.Float,
		Floats:      def.Floats,
		Int32:       def.Int32,
		Int32S:      def.Int32S,
		Int64:       def.Int64,
		Int64S:      def.Int64S,
		Uint32:      def.Uint32,
		Uint32S:     def.Uint32S,
		Uint64:      def.Uint64,
		Uint64S:     def.Uint64S,
		Sint32:      def.Sint32,
		Sint32S:     def.Sint32S,
		Sint64:      def.Sint64,
		Sint64S:     def.Sint64S,
		Fixed32:     def.Fixed32,
		Fixed32S:    def.Fixed32S,
		Fixed64:     def.Fixed64,
		Fixed64S:    def.Fixed64S,
		Sfixed32:    def.Sfixed32,
		Sfixed32S:   def.Sfixed32S,
		Sfixed64:    def.Sfixed64,
		Sfixed64S:   def.Sfixed64S,
		Bool:        def.Bool,
		Bools:       def.Bools,
		String:      def.String_,
		Strings:     def.Strings,
		ByteString:  def.ByteString,
		ByteStrings: def.ByteStrings,
		Message:     def.Message,
		Messages:    def.Messages,
		Enum:        def.Enum,
		Enums:       def.Enums,
		Env:         def.Env,
		Envs:        def.Envs,
	}
}

func messageFieldValueToCommonValueDef(def *federation.MessageFieldValue) *commonValueDef {
	return &commonValueDef{
		Double:      def.Double,
		Doubles:     def.Doubles,
		Float:       def.Float,
		Floats:      def.Floats,
		Int32:       def.Int32,
		Int32S:      def.Int32S,
		Int64:       def.Int64,
		Int64S:      def.Int64S,
		Uint32:      def.Uint32,
		Uint32S:     def.Uint32S,
		Uint64:      def.Uint64,
		Uint64S:     def.Uint64S,
		Sint32:      def.Sint32,
		Sint32S:     def.Sint32S,
		Sint64:      def.Sint64,
		Sint64S:     def.Sint64S,
		Fixed32:     def.Fixed32,
		Fixed32S:    def.Fixed32S,
		Fixed64:     def.Fixed64,
		Fixed64S:    def.Fixed64S,
		Sfixed32:    def.Sfixed32,
		Sfixed32S:   def.Sfixed32S,
		Sfixed64:    def.Sfixed64,
		Sfixed64S:   def.Sfixed64S,
		Bool:        def.Bool,
		Bools:       def.Bools,
		String:      def.String_,
		Strings:     def.Strings,
		ByteString:  def.ByteString,
		ByteStrings: def.ByteStrings,
		Message:     def.Message,
		Messages:    def.Messages,
		Enum:        def.Enum,
		Enums:       def.Enums,
		Env:         def.Env,
		Envs:        def.Envs,
	}
}

func NewDoubleValue(v float64) *Value {
	return &Value{Const: &ConstValue{Type: DoubleType, Value: v}}
}

func NewFloatValue(v float32) *Value {
	return &Value{Const: &ConstValue{Type: FloatType, Value: v}}
}

func NewInt32Value(v int32) *Value {
	return &Value{Const: &ConstValue{Type: Int32Type, Value: v}}
}

func NewInt64Value(v int64) *Value {
	return &Value{Const: &ConstValue{Type: Int64Type, Value: v}}
}

func NewUint32Value(v uint32) *Value {
	return &Value{Const: &ConstValue{Type: Uint32Type, Value: v}}
}

func NewUint64Value(v uint64) *Value {
	return &Value{Const: &ConstValue{Type: Uint64Type, Value: v}}
}

func NewSint32Value(v int32) *Value {
	return &Value{Const: &ConstValue{Type: Sint32Type, Value: v}}
}

func NewSint64Value(v int64) *Value {
	return &Value{Const: &ConstValue{Type: Sint64Type, Value: v}}
}

func NewFixed32Value(v uint32) *Value {
	return &Value{Const: &ConstValue{Type: Fixed32Type, Value: v}}
}

func NewFixed64Value(v uint64) *Value {
	return &Value{Const: &ConstValue{Type: Fixed64Type, Value: v}}
}

func NewSfixed32Value(v int32) *Value {
	return &Value{Const: &ConstValue{Type: Sfixed32Type, Value: v}}
}

func NewSfixed64Value(v int64) *Value {
	return &Value{Const: &ConstValue{Type: Sfixed64Type, Value: v}}
}

func NewBoolValue(v bool) *Value {
	return &Value{Const: &ConstValue{Type: BoolType, Value: v}}
}

func NewStringValue(v string) *Value {
	return &Value{Const: &ConstValue{Type: StringType, Value: v}}
}

func NewByteStringValue(v []byte) *Value {
	return &Value{Const: &ConstValue{Type: BytesType, Value: v}}
}

func NewMessageValue(typ *Type, v map[string]*Value) *Value {
	return &Value{Const: &ConstValue{Type: typ, Value: v}}
}

func NewDoublesValue(v ...float64) *Value {
	return &Value{Const: &ConstValue{Type: DoubleRepeatedType, Value: v}}
}

func NewFloatsValue(v ...float32) *Value {
	return &Value{Const: &ConstValue{Type: FloatRepeatedType, Value: v}}
}

func NewInt32sValue(v ...int32) *Value {
	return &Value{Const: &ConstValue{Type: Int32RepeatedType, Value: v}}
}

func NewInt64sValue(v ...int64) *Value {
	return &Value{Const: &ConstValue{Type: Int64RepeatedType, Value: v}}
}

func NewUint32sValue(v ...uint32) *Value {
	return &Value{Const: &ConstValue{Type: Uint32RepeatedType, Value: v}}
}

func NewUint64sValue(v ...uint64) *Value {
	return &Value{Const: &ConstValue{Type: Uint64RepeatedType, Value: v}}
}

func NewSint32sValue(v ...int32) *Value {
	return &Value{Const: &ConstValue{Type: Sint32RepeatedType, Value: v}}
}

func NewSint64sValue(v ...int64) *Value {
	return &Value{Const: &ConstValue{Type: Sint64RepeatedType, Value: v}}
}

func NewFixed32sValue(v ...uint32) *Value {
	return &Value{Const: &ConstValue{Type: Fixed32RepeatedType, Value: v}}
}

func NewFixed64sValue(v ...uint64) *Value {
	return &Value{Const: &ConstValue{Type: Fixed64RepeatedType, Value: v}}
}

func NewSfixed32sValue(v ...int32) *Value {
	return &Value{Const: &ConstValue{Type: Sfixed32RepeatedType, Value: v}}
}

func NewSfixed64sValue(v ...int64) *Value {
	return &Value{Const: &ConstValue{Type: Sfixed64RepeatedType, Value: v}}
}

func NewBoolsValue(v ...bool) *Value {
	return &Value{Const: &ConstValue{Type: BoolRepeatedType, Value: v}}
}

func NewStringsValue(v ...string) *Value {
	return &Value{Const: &ConstValue{Type: StringRepeatedType, Value: v}}
}

func NewByteStringsValue(v ...[]byte) *Value {
	return &Value{Const: &ConstValue{Type: BytesRepeatedType, Value: v}}
}

func NewMessagesValue(typ *Type, v ...map[string]*Value) *Value {
	return &Value{Const: &ConstValue{Type: typ, Value: v}}
}

func NewEnumValue(v *EnumValue) *Value {
	var enum *Enum
	if v != nil {
		enum = v.Enum
	}
	return &Value{
		Const: &ConstValue{
			Type: &Type{
				Type: types.Enum,
				Enum: enum,
			},
			Value: v,
		},
	}
}

func NewEnumsValue(v ...*EnumValue) *Value {
	var enum *Enum
	if len(v) != 0 {
		enum = v[0].Enum
	}
	return &Value{
		Const: &ConstValue{
			Type: &Type{
				Type:     types.Enum,
				Enum:     enum,
				Repeated: true,
			},
			Value: v,
		},
	}
}

func NewEnvValue(v EnvKey) *Value {
	return &Value{Const: &ConstValue{Type: EnvType, Value: v}}
}

func NewEnvsValue(v ...EnvKey) *Value {
	return &Value{Const: &ConstValue{Type: EnvRepeatedType, Value: v}}
}
