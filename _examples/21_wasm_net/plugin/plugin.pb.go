// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: plugin/plugin.proto

package pluginpb

import (
	_ "github.com/mercari/grpc-federation/grpc/federation"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_plugin_plugin_proto protoreflect.FileDescriptor

var file_plugin_plugin_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x6e,
	0x65, 0x74, 0x1a, 0x20, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x42, 0xf6, 0x01, 0x9a, 0x4a, 0x6f, 0x0a, 0x6d, 0x0a, 0x6b, 0x0a, 0x03,
	0x6e, 0x65, 0x74, 0x22, 0x24, 0x0a, 0x07, 0x68, 0x74, 0x74, 0x70, 0x47, 0x65, 0x74, 0x1a, 0x15,
	0x0a, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x0a, 0x75, 0x72, 0x6c, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x1a, 0x02, 0x08, 0x01, 0x22, 0x02, 0x08, 0x01, 0x22, 0x0f, 0x0a, 0x09, 0x67, 0x65, 0x74,
	0x46, 0x6f, 0x6f, 0x45, 0x6e, 0x76, 0x22, 0x02, 0x08, 0x01, 0x22, 0x20, 0x0a, 0x0e, 0x67, 0x65,
	0x74, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x1a, 0x0a, 0x0a, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x1a, 0x02, 0x08, 0x01, 0x22, 0x02, 0x08, 0x01, 0x32, 0x0b, 0x0a, 0x02,
	0x08, 0x01, 0x12, 0x03, 0x0a, 0x01, 0x2f, 0x1a, 0x00, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x6e, 0x65, 0x74, 0x42, 0x0b, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x17, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x3b, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x70, 0x62, 0xa2, 0x02, 0x03, 0x45, 0x4e, 0x58, 0xaa, 0x02, 0x0b, 0x45, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2e, 0x4e, 0x65, 0x74, 0xca, 0x02, 0x0b, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x5c, 0x4e, 0x65, 0x74, 0xe2, 0x02, 0x17, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5c, 0x4e,
	0x65, 0x74, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x0c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x3a, 0x3a, 0x4e, 0x65, 0x74, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_plugin_plugin_proto_goTypes = []interface{}{}
var file_plugin_plugin_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_plugin_plugin_proto_init() }
func file_plugin_plugin_proto_init() {
	if File_plugin_plugin_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_plugin_plugin_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_plugin_plugin_proto_goTypes,
		DependencyIndexes: file_plugin_plugin_proto_depIdxs,
	}.Build()
	File_plugin_plugin_proto = out.File
	file_plugin_plugin_proto_rawDesc = nil
	file_plugin_plugin_proto_goTypes = nil
	file_plugin_plugin_proto_depIdxs = nil
}
