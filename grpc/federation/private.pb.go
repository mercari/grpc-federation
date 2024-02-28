// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: grpc/federation/private.proto

package federation

import (
	errdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Error type information of the error variable used when evaluating CEL.
type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code                 int32                             `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string                            `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Details              []*anypb.Any                      `protobuf:"bytes,3,rep,name=details,proto3" json:"details,omitempty"`
	CustomMessages       []*anypb.Any                      `protobuf:"bytes,4,rep,name=custom_messages,json=customMessages,proto3" json:"custom_messages,omitempty"`
	ErrorInfo            []*errdetails.ErrorInfo           `protobuf:"bytes,5,rep,name=error_info,json=errorInfo,proto3" json:"error_info,omitempty"`
	RetryInfo            []*errdetails.RetryInfo           `protobuf:"bytes,6,rep,name=retry_info,json=retryInfo,proto3" json:"retry_info,omitempty"`
	DebugInfo            []*errdetails.DebugInfo           `protobuf:"bytes,7,rep,name=debug_info,json=debugInfo,proto3" json:"debug_info,omitempty"`
	QuotaFailures        []*errdetails.QuotaFailure        `protobuf:"bytes,8,rep,name=quota_failures,json=quotaFailures,proto3" json:"quota_failures,omitempty"`
	PreconditionFailures []*errdetails.PreconditionFailure `protobuf:"bytes,9,rep,name=precondition_failures,json=preconditionFailures,proto3" json:"precondition_failures,omitempty"`
	BadRequests          []*errdetails.BadRequest          `protobuf:"bytes,10,rep,name=bad_requests,json=badRequests,proto3" json:"bad_requests,omitempty"`
	RequestInfo          []*errdetails.RequestInfo         `protobuf:"bytes,11,rep,name=request_info,json=requestInfo,proto3" json:"request_info,omitempty"`
	ResourceInfo         []*errdetails.ResourceInfo        `protobuf:"bytes,12,rep,name=resource_info,json=resourceInfo,proto3" json:"resource_info,omitempty"`
	Helps                []*errdetails.Help                `protobuf:"bytes,13,rep,name=helps,proto3" json:"helps,omitempty"`
	LocalizedMessages    []*errdetails.LocalizedMessage    `protobuf:"bytes,14,rep,name=localized_messages,json=localizedMessages,proto3" json:"localized_messages,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_federation_private_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_federation_private_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_grpc_federation_private_proto_rawDescGZIP(), []int{0}
}

func (x *Error) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Error) GetDetails() []*anypb.Any {
	if x != nil {
		return x.Details
	}
	return nil
}

func (x *Error) GetCustomMessages() []*anypb.Any {
	if x != nil {
		return x.CustomMessages
	}
	return nil
}

func (x *Error) GetErrorInfo() []*errdetails.ErrorInfo {
	if x != nil {
		return x.ErrorInfo
	}
	return nil
}

func (x *Error) GetRetryInfo() []*errdetails.RetryInfo {
	if x != nil {
		return x.RetryInfo
	}
	return nil
}

func (x *Error) GetDebugInfo() []*errdetails.DebugInfo {
	if x != nil {
		return x.DebugInfo
	}
	return nil
}

func (x *Error) GetQuotaFailures() []*errdetails.QuotaFailure {
	if x != nil {
		return x.QuotaFailures
	}
	return nil
}

func (x *Error) GetPreconditionFailures() []*errdetails.PreconditionFailure {
	if x != nil {
		return x.PreconditionFailures
	}
	return nil
}

func (x *Error) GetBadRequests() []*errdetails.BadRequest {
	if x != nil {
		return x.BadRequests
	}
	return nil
}

func (x *Error) GetRequestInfo() []*errdetails.RequestInfo {
	if x != nil {
		return x.RequestInfo
	}
	return nil
}

func (x *Error) GetResourceInfo() []*errdetails.ResourceInfo {
	if x != nil {
		return x.ResourceInfo
	}
	return nil
}

func (x *Error) GetHelps() []*errdetails.Help {
	if x != nil {
		return x.Helps
	}
	return nil
}

func (x *Error) GetLocalizedMessages() []*errdetails.LocalizedMessage {
	if x != nil {
		return x.LocalizedMessages
	}
	return nil
}

var File_grpc_federation_private_proto protoreflect.FileDescriptor

var file_grpc_federation_private_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x17, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x72, 0x70, 0x63, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x88, 0x06, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x64,
	0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41,
	0x6e, 0x79, 0x52, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x3d, 0x0a, 0x0f, 0x63,
	0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0e, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x34, 0x0a, 0x0a, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x34, 0x0a, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x06,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x72, 0x65, 0x74,
	0x72, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x34, 0x0a, 0x0a, 0x64, 0x65, 0x62, 0x75, 0x67, 0x5f,
	0x69, 0x6e, 0x66, 0x6f, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x62, 0x75, 0x67, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x09, 0x64, 0x65, 0x62, 0x75, 0x67, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x3f, 0x0a, 0x0e,
	0x71, 0x75, 0x6f, 0x74, 0x61, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x73, 0x18, 0x08,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x51, 0x75, 0x6f, 0x74, 0x61, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x0d,
	0x71, 0x75, 0x6f, 0x74, 0x61, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x73, 0x12, 0x54, 0x0a,
	0x15, 0x70, 0x72, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x66, 0x61,
	0x69, 0x6c, 0x75, 0x72, 0x65, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x72, 0x65, 0x63, 0x6f, 0x6e,
	0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x14, 0x70,
	0x72, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61, 0x69, 0x6c, 0x75,
	0x72, 0x65, 0x73, 0x12, 0x39, 0x0a, 0x0c, 0x62, 0x61, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x42, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x52, 0x0b, 0x62, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x12, 0x3a,
	0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x0b,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x3d, 0x0a, 0x0d, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x0c, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x18, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0c, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x26, 0x0a, 0x05, 0x68, 0x65, 0x6c,
	0x70, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x48, 0x65, 0x6c, 0x70, 0x52, 0x05, 0x68, 0x65, 0x6c, 0x70,
	0x73, 0x12, 0x4b, 0x0a, 0x12, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x6c,
	0x69, 0x7a, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x11, 0x6c, 0x6f, 0x63,
	0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x42, 0x3f,
	0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x72,
	0x63, 0x61, 0x72, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2d, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x3b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpc_federation_private_proto_rawDescOnce sync.Once
	file_grpc_federation_private_proto_rawDescData = file_grpc_federation_private_proto_rawDesc
)

func file_grpc_federation_private_proto_rawDescGZIP() []byte {
	file_grpc_federation_private_proto_rawDescOnce.Do(func() {
		file_grpc_federation_private_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_federation_private_proto_rawDescData)
	})
	return file_grpc_federation_private_proto_rawDescData
}

var file_grpc_federation_private_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_grpc_federation_private_proto_goTypes = []interface{}{
	(*Error)(nil),                          // 0: grpc.federation.private.Error
	(*anypb.Any)(nil),                      // 1: google.protobuf.Any
	(*errdetails.ErrorInfo)(nil),           // 2: google.rpc.ErrorInfo
	(*errdetails.RetryInfo)(nil),           // 3: google.rpc.RetryInfo
	(*errdetails.DebugInfo)(nil),           // 4: google.rpc.DebugInfo
	(*errdetails.QuotaFailure)(nil),        // 5: google.rpc.QuotaFailure
	(*errdetails.PreconditionFailure)(nil), // 6: google.rpc.PreconditionFailure
	(*errdetails.BadRequest)(nil),          // 7: google.rpc.BadRequest
	(*errdetails.RequestInfo)(nil),         // 8: google.rpc.RequestInfo
	(*errdetails.ResourceInfo)(nil),        // 9: google.rpc.ResourceInfo
	(*errdetails.Help)(nil),                // 10: google.rpc.Help
	(*errdetails.LocalizedMessage)(nil),    // 11: google.rpc.LocalizedMessage
}
var file_grpc_federation_private_proto_depIdxs = []int32{
	1,  // 0: grpc.federation.private.Error.details:type_name -> google.protobuf.Any
	1,  // 1: grpc.federation.private.Error.custom_messages:type_name -> google.protobuf.Any
	2,  // 2: grpc.federation.private.Error.error_info:type_name -> google.rpc.ErrorInfo
	3,  // 3: grpc.federation.private.Error.retry_info:type_name -> google.rpc.RetryInfo
	4,  // 4: grpc.federation.private.Error.debug_info:type_name -> google.rpc.DebugInfo
	5,  // 5: grpc.federation.private.Error.quota_failures:type_name -> google.rpc.QuotaFailure
	6,  // 6: grpc.federation.private.Error.precondition_failures:type_name -> google.rpc.PreconditionFailure
	7,  // 7: grpc.federation.private.Error.bad_requests:type_name -> google.rpc.BadRequest
	8,  // 8: grpc.federation.private.Error.request_info:type_name -> google.rpc.RequestInfo
	9,  // 9: grpc.federation.private.Error.resource_info:type_name -> google.rpc.ResourceInfo
	10, // 10: grpc.federation.private.Error.helps:type_name -> google.rpc.Help
	11, // 11: grpc.federation.private.Error.localized_messages:type_name -> google.rpc.LocalizedMessage
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_grpc_federation_private_proto_init() }
func file_grpc_federation_private_proto_init() {
	if File_grpc_federation_private_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_federation_private_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_grpc_federation_private_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_grpc_federation_private_proto_goTypes,
		DependencyIndexes: file_grpc_federation_private_proto_depIdxs,
		MessageInfos:      file_grpc_federation_private_proto_msgTypes,
	}.Build()
	File_grpc_federation_private_proto = out.File
	file_grpc_federation_private_proto_rawDesc = nil
	file_grpc_federation_private_proto_goTypes = nil
	file_grpc_federation_private_proto_depIdxs = nil
}
