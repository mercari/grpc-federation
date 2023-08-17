package resolver

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/mercari/grpc-federation/grpc/federation"
)

func getServiceRule(service *descriptorpb.ServiceDescriptorProto) (*federation.ServiceRule, error) {
	opts := service.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_Service) {
		return nil, nil
	}
	var serviceRule federation.ServiceRule
	if setRuleFromDynamicMessage(opts, &serviceRule) {
		return &serviceRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_Service)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.service extension does not exist")
	}
	rule, ok := ext.(*federation.ServiceRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.service extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

func getMethodRule(service *descriptorpb.MethodDescriptorProto) (*federation.MethodRule, error) {
	opts := service.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_Method) {
		return nil, nil
	}
	var methodRule federation.MethodRule
	if setRuleFromDynamicMessage(opts, &methodRule) {
		return &methodRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_Method)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.method extension does not exist")
	}
	rule, ok := ext.(*federation.MethodRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.method extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

func getMessageRule(msg *descriptorpb.DescriptorProto) (*federation.MessageRule, error) {
	opts := msg.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_Message) {
		return nil, nil
	}
	var msgRule federation.MessageRule
	if setRuleFromDynamicMessage(opts, &msgRule) {
		return &msgRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_Message)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.message extension does not exist")
	}
	rule, ok := ext.(*federation.MessageRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.message extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

func getEnumRule(enum *descriptorpb.EnumDescriptorProto) (*federation.EnumRule, error) {
	opts := enum.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_Enum) {
		return nil, nil
	}
	var enumRule federation.EnumRule
	if setRuleFromDynamicMessage(opts, &enumRule) {
		return &enumRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_Enum)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.enum extension does not exist")
	}
	rule, ok := ext.(*federation.EnumRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.enum extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

func getEnumValueRule(value *descriptorpb.EnumValueDescriptorProto) (*federation.EnumValueRule, error) {
	opts := value.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_EnumValue) {
		return nil, nil
	}
	var valueRule federation.EnumValueRule
	if setRuleFromDynamicMessage(opts, &valueRule) {
		return &valueRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_EnumValue)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.enum_value extension does not exist")
	}
	rule, ok := ext.(*federation.EnumValueRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.enum_value extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

func getFieldRule(field *descriptorpb.FieldDescriptorProto) (*federation.FieldRule, error) {
	opts := field.GetOptions()
	if opts == nil {
		return nil, nil
	}
	if !proto.HasExtension(opts, federation.E_Field) {
		return nil, nil
	}
	var fieldRule federation.FieldRule
	if setRuleFromDynamicMessage(opts, &fieldRule) {
		return &fieldRule, nil
	}

	ext := proto.GetExtension(opts, federation.E_Field)
	if ext == nil {
		return nil, fmt.Errorf("grpc.federation.field extension does not exist")
	}
	rule, ok := ext.(*federation.FieldRule)
	if !ok {
		return nil, fmt.Errorf("grpc.federation.field extension cannot not be converted from %T", ext)
	}
	return rule, nil
}

// setRuleFromDynamicMessage if each options are represented dynamicpb.Message type, convert and set it to rule instance.
// NOTE: compile proto files by compiler package, extension is replaced by dynamicpb.Message.
func setRuleFromDynamicMessage(opts proto.Message, rule proto.Message) bool {
	isSet := false

	opts.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
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
