package resolver

import (
	"fmt"
)

func (v *Value) Fields(msgArgType *Type) ([]*Field, error) {
	if v.Literal != nil {
		return []*Field{{Type: v.Literal.Type}}, nil
	}
	if v.PathType == MessageArgumentPathType {
		fieldType, err := v.Path.Type(msgArgType)
		if err != nil {
			return nil, err
		}
		if v.Inline {
			if fieldType.Ref == nil {
				sels := v.Path.Selectors()
				if len(sels) != 0 {
					return nil, fmt.Errorf(`inline keyword must refer to a message type. but "%s" is not a message type`, sels[0])
				}
				return nil, fmt.Errorf(`inline keyword must refer to a message type. but path selector is empty`)
			}
			return fieldType.Ref.Fields, nil
		}
		return []*Field{{Type: fieldType}}, nil
	}
	if v.Filtered == nil {
		return nil, fmt.Errorf("undefined name reference")
	}
	if v.Inline {
		nameReference := v.Filtered.Ref
		if nameReference == nil {
			return nil, fmt.Errorf("undefined name reference")
		}
		return nameReference.Fields, nil
	}
	return []*Field{{Type: v.Filtered}}, nil
}

func NewDoubleValue(v float64) *Value {
	return &Value{Literal: &Literal{Type: DoubleType, Value: v}}
}

func NewFloatValue(v float32) *Value {
	return &Value{Literal: &Literal{Type: FloatType, Value: v}}
}

func NewInt32Value(v int32) *Value {
	return &Value{Literal: &Literal{Type: Int32Type, Value: v}}
}

func NewInt64Value(v int64) *Value {
	return &Value{Literal: &Literal{Type: Int64Type, Value: v}}
}

func NewUint32Value(v uint32) *Value {
	return &Value{Literal: &Literal{Type: Uint32Type, Value: v}}
}

func NewUint64Value(v uint64) *Value {
	return &Value{Literal: &Literal{Type: Uint64Type, Value: v}}
}

func NewSint32Value(v int32) *Value {
	return &Value{Literal: &Literal{Type: Sint32Type, Value: v}}
}

func NewSint64Value(v int64) *Value {
	return &Value{Literal: &Literal{Type: Sint64Type, Value: v}}
}

func NewFixed32Value(v uint32) *Value {
	return &Value{Literal: &Literal{Type: Fixed32Type, Value: v}}
}

func NewFixed64Value(v uint64) *Value {
	return &Value{Literal: &Literal{Type: Fixed64Type, Value: v}}
}

func NewSfixed32Value(v int32) *Value {
	return &Value{Literal: &Literal{Type: Sfixed32Type, Value: v}}
}

func NewSfixed64Value(v int64) *Value {
	return &Value{Literal: &Literal{Type: Sfixed64Type, Value: v}}
}

func NewBoolValue(v bool) *Value {
	return &Value{Literal: &Literal{Type: BoolType, Value: v}}
}

func NewStringValue(v string) *Value {
	return &Value{Literal: &Literal{Type: StringType, Value: v}}
}

func NewBytesValue(v []byte) *Value {
	return &Value{Literal: &Literal{Type: BytesType, Value: v}}
}

func NewMessageValue(typ *Type, v map[string]*Value) *Value {
	return &Value{Literal: &Literal{Type: typ, Value: v}}
}

func NewDoubleListValue(v ...float64) *Value {
	return &Value{Literal: &Literal{Type: DoubleRepeatedType, Value: v}}
}

func NewFloatListValue(v ...float32) *Value {
	return &Value{Literal: &Literal{Type: FloatRepeatedType, Value: v}}
}

func NewInt32ListValue(v ...int32) *Value {
	return &Value{Literal: &Literal{Type: Int32RepeatedType, Value: v}}
}

func NewInt64ListValue(v ...int64) *Value {
	return &Value{Literal: &Literal{Type: Int64RepeatedType, Value: v}}
}

func NewUint32ListValue(v ...uint32) *Value {
	return &Value{Literal: &Literal{Type: Uint32RepeatedType, Value: v}}
}

func NewUint64ListValue(v ...uint64) *Value {
	return &Value{Literal: &Literal{Type: Uint64RepeatedType, Value: v}}
}

func NewSint32ListValue(v ...int32) *Value {
	return &Value{Literal: &Literal{Type: Sint32RepeatedType, Value: v}}
}

func NewSint64ListValue(v ...int64) *Value {
	return &Value{Literal: &Literal{Type: Sint64RepeatedType, Value: v}}
}

func NewFixed32ListValue(v ...uint32) *Value {
	return &Value{Literal: &Literal{Type: Fixed32RepeatedType, Value: v}}
}

func NewFixed64ListValue(v ...uint64) *Value {
	return &Value{Literal: &Literal{Type: Fixed64RepeatedType, Value: v}}
}

func NewSfixed32ListValue(v ...int32) *Value {
	return &Value{Literal: &Literal{Type: Sfixed32RepeatedType, Value: v}}
}

func NewSfixed64ListValue(v ...int64) *Value {
	return &Value{Literal: &Literal{Type: Sfixed64RepeatedType, Value: v}}
}

func NewBoolListValue(v ...bool) *Value {
	return &Value{Literal: &Literal{Type: BoolRepeatedType, Value: v}}
}

func NewStringListValue(v ...string) *Value {
	return &Value{Literal: &Literal{Type: StringRepeatedType, Value: v}}
}

func NewBytesListValue(v ...[]byte) *Value {
	return &Value{Literal: &Literal{Type: BytesRepeatedType, Value: v}}
}

func NewMessageListValue(typ *Type, v ...map[string]*Value) *Value {
	return &Value{Literal: &Literal{Type: typ, Value: v}}
}
