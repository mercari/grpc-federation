package types

import "google.golang.org/protobuf/types/descriptorpb"

type Kind int

var (
	Unknown  Kind = -1
	Double        = Kind(descriptorpb.FieldDescriptorProto_TYPE_DOUBLE)
	Float         = Kind(descriptorpb.FieldDescriptorProto_TYPE_FLOAT)
	Int64         = Kind(descriptorpb.FieldDescriptorProto_TYPE_INT64)
	Uint64        = Kind(descriptorpb.FieldDescriptorProto_TYPE_UINT64)
	Int32         = Kind(descriptorpb.FieldDescriptorProto_TYPE_INT32)
	Fixed64       = Kind(descriptorpb.FieldDescriptorProto_TYPE_FIXED64)
	Fixed32       = Kind(descriptorpb.FieldDescriptorProto_TYPE_FIXED32)
	Bool          = Kind(descriptorpb.FieldDescriptorProto_TYPE_BOOL)
	String        = Kind(descriptorpb.FieldDescriptorProto_TYPE_STRING)
	Group         = Kind(descriptorpb.FieldDescriptorProto_TYPE_GROUP)
	Message       = Kind(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE)
	Bytes         = Kind(descriptorpb.FieldDescriptorProto_TYPE_BYTES)
	Uint32        = Kind(descriptorpb.FieldDescriptorProto_TYPE_UINT32)
	Enum          = Kind(descriptorpb.FieldDescriptorProto_TYPE_ENUM)
	Sfixed32      = Kind(descriptorpb.FieldDescriptorProto_TYPE_SFIXED32)
	Sfixed64      = Kind(descriptorpb.FieldDescriptorProto_TYPE_SFIXED64)
	Sint32        = Kind(descriptorpb.FieldDescriptorProto_TYPE_SINT32)
	Sint64        = Kind(descriptorpb.FieldDescriptorProto_TYPE_SINT64)
)

func (k Kind) ToString() string {
	switch k {
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

func ToKind(s string) Kind {
	switch s {
	case "double":
		return Double
	case "float":
		return Float
	case "int64":
		return Int64
	case "uint64":
		return Uint64
	case "int32":
		return Int32
	case "fixed64":
		return Fixed64
	case "fixed32":
		return Fixed32
	case "bool":
		return Bool
	case "string":
		return String
	case "group":
		return Group
	case "bytes":
		return Bytes
	case "uint32":
		return Uint32
	case "sfixed32":
		return Sfixed32
	case "sfixed64":
		return Sfixed64
	case "sint32":
		return Sint32
	case "sint64":
		return Sint64
	}
	return Unknown
}
