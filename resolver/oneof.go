package resolver

import (
	"fmt"

	"github.com/mercari/grpc-federation/types"
	"github.com/mercari/grpc-federation/util"
)

func (oneof *Oneof) IsSameType() bool {
	if len(oneof.Fields) == 0 {
		return false
	}
	var prevType *Type
	for _, field := range oneof.Fields {
		fieldType := field.Type
		if prevType == nil {
			prevType = fieldType
		}
		if prevType.Kind != fieldType.Kind {
			return false
		}
		switch fieldType.Kind {
		case types.Message:
			if prevType.Message != fieldType.Message {
				return false
			}
		case types.Enum:
			if prevType.Enum != fieldType.Enum {
				return false
			}
		}
		prevType = fieldType
	}
	return true
}

func (f *OneofField) IsConflict() bool {
	msg := f.Oneof.Message
	fieldName := util.ToPublicGoVariable(f.Name)
	for _, m := range msg.NestedMessages {
		if util.ToPublicGoVariable(m.Name) == fieldName {
			return true
		}
	}
	for _, e := range msg.Enums {
		if util.ToPublicGoVariable(e.Name) == fieldName {
			return true
		}
	}
	return false
}

func (f *OneofField) FQDN() string {
	return fmt.Sprintf("%s.%s", f.Oneof.Message.FQDN(), f.Name)
}
