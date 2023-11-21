package resolver

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func getExtensionRule[T proto.Message](opts proto.Message, extType protoreflect.ExtensionType) (T, error) {
	var ret T

	typ := reflect.TypeOf(ret)
	if typ.Kind() != reflect.Ptr {
		return ret, fmt.Errorf("proto.Message value must be pointer type")
	}
	v := reflect.New(typ.Elem()).Interface().(proto.Message)

	if opts == nil {
		return ret, nil
	}
	if !proto.HasExtension(opts, extType) {
		return ret, nil
	}

	extFullName := extType.TypeDescriptor().Descriptor().FullName()

	if setRuleFromDynamicMessage(opts, extFullName, v) {
		return v.(T), nil
	}

	ext := proto.GetExtension(opts, extType)
	if ext == nil {
		return ret, fmt.Errorf("%s extension does not exist", extFullName)
	}
	rule, ok := ext.(T)
	if !ok {
		return ret, fmt.Errorf("%s extension cannot not be converted from %T", extFullName, ext)
	}
	return rule, nil
}

// setRuleFromDynamicMessage if each options are represented dynamicpb.Message type, convert and set it to rule instance.
// NOTE: compile proto files by compiler package, extension is replaced by dynamicpb.Message.
func setRuleFromDynamicMessage(opts proto.Message, extFullName protoreflect.FullName, rule proto.Message) bool {
	isSet := false
	opts.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}
		if fd.FullName() != extFullName {
			return true
		}
		ext := proto.GetExtension(opts, dynamicpb.NewExtensionType(fd))
		if ext == nil {
			return true
		}
		msg, ok := ext.(*dynamicpb.Message)
		if !ok {
			return true
		}
		bytes, err := proto.Marshal(msg)
		if err != nil {
			return true
		}
		if err := proto.Unmarshal(bytes, rule); err != nil {
			return true
		}

		isSet = true

		return true
	})
	return isSet
}
