package types

import "google.golang.org/protobuf/types/descriptorpb"

type Kind = descriptorpb.FieldDescriptorProto_Type

var (
	Double   = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
	Float    = descriptorpb.FieldDescriptorProto_TYPE_FLOAT
	Int64    = descriptorpb.FieldDescriptorProto_TYPE_INT64
	Uint64   = descriptorpb.FieldDescriptorProto_TYPE_UINT64
	Int32    = descriptorpb.FieldDescriptorProto_TYPE_INT32
	Fixed64  = descriptorpb.FieldDescriptorProto_TYPE_FIXED64
	Fixed32  = descriptorpb.FieldDescriptorProto_TYPE_FIXED32
	Bool     = descriptorpb.FieldDescriptorProto_TYPE_BOOL
	String   = descriptorpb.FieldDescriptorProto_TYPE_STRING
	Group    = descriptorpb.FieldDescriptorProto_TYPE_GROUP
	Message  = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	Bytes    = descriptorpb.FieldDescriptorProto_TYPE_BYTES
	Uint32   = descriptorpb.FieldDescriptorProto_TYPE_UINT32
	Enum     = descriptorpb.FieldDescriptorProto_TYPE_ENUM
	Sfixed32 = descriptorpb.FieldDescriptorProto_TYPE_SFIXED32
	Sfixed64 = descriptorpb.FieldDescriptorProto_TYPE_SFIXED64
	Sint32   = descriptorpb.FieldDescriptorProto_TYPE_SINT32
	Sint64   = descriptorpb.FieldDescriptorProto_TYPE_SINT64
)

func ToString(kind Kind) string {
	switch kind {
	case Double:
		return "double"
	case Float:
		return "float"
	case Int64:
		return "int64"
	case Uint64:
		return "uint64"
	case Int32:
		return "int32"
	case Fixed64:
		return "fixed64"
	case Fixed32:
		return "fixed32"
	case Bool:
		return "bool"
	case String:
		return "string"
	case Group:
		return "group"
	case Message:
		return "message"
	case Bytes:
		return "bytes"
	case Uint32:
		return "uint32"
	case Enum:
		return "enum"
	case Sfixed32:
		return "sfixed32"
	case Sfixed64:
		return "sfixed64"
	case Sint32:
		return "sint32"
	case Sint64:
		return "sint64"
	}
	return ""
}
