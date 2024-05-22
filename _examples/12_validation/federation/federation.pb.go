// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: federation/federation.proto

package federation

import (
	_ "github.com/mercari/grpc-federation/grpc/federation"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetPostRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetPostRequest) Reset() {
	*x = GetPostRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostRequest) ProtoMessage() {}

func (x *GetPostRequest) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPostRequest.ProtoReflect.Descriptor instead.
func (*GetPostRequest) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0}
}

func (x *GetPostRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetPostResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Post *Post `protobuf:"bytes,1,opt,name=post,proto3" json:"post,omitempty"`
}

func (x *GetPostResponse) Reset() {
	*x = GetPostResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostResponse) ProtoMessage() {}

func (x *GetPostResponse) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPostResponse.ProtoReflect.Descriptor instead.
func (*GetPostResponse) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{1}
}

func (x *GetPostResponse) GetPost() *Post {
	if x != nil {
		return x.Post
	}
	return nil
}

type Post struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title   string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Content string `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *Post) Reset() {
	*x = Post{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Post) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Post) ProtoMessage() {}

func (x *Post) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Post.ProtoReflect.Descriptor instead.
func (*Post) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{2}
}

func (x *Post) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Post) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Post) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type CustomMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *CustomMessage) Reset() {
	*x = CustomMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CustomMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CustomMessage) ProtoMessage() {}

func (x *CustomMessage) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CustomMessage.ProtoReflect.Descriptor instead.
func (*CustomMessage) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{3}
}

func (x *CustomMessage) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type CustomHandlerMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CustomHandlerMessage) Reset() {
	*x = CustomHandlerMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CustomHandlerMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CustomHandlerMessage) ProtoMessage() {}

func (x *CustomHandlerMessage) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CustomHandlerMessage.ProtoReflect.Descriptor instead.
func (*CustomHandlerMessage) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{4}
}

var File_federation_federation_proto protoreflect.FileDescriptor

var file_federation_federation_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x66, 0x65, 0x64,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6f,
	0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x19, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61,
	0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x20, 0x0a, 0x0e, 0x47, 0x65,
	0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0xb2, 0x04, 0x0a,
	0x0f, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x33, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x50, 0x6f, 0x73, 0x74, 0x42, 0x09, 0x9a, 0x4a, 0x06, 0x12, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x52,
	0x04, 0x70, 0x6f, 0x73, 0x74, 0x3a, 0xe9, 0x03, 0x9a, 0x4a, 0xe5, 0x03, 0x0a, 0x0e, 0x0a, 0x04,
	0x70, 0x6f, 0x73, 0x74, 0x6a, 0x06, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x0a, 0x39, 0x0a, 0x0d,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x6a, 0x28, 0x0a,
	0x14, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x72, 0x67, 0xf2, 0x01, 0x08, 0x73,
	0x6f, 0x6d, 0x65, 0x2d, 0x61, 0x72, 0x67, 0x0a, 0x33, 0x7a, 0x31, 0x12, 0x2f, 0x12, 0x14, 0x70,
	0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64, 0x20, 0x21, 0x3d, 0x20, 0x27, 0x73, 0x6f, 0x6d, 0x65, 0x2d,
	0x69, 0x64, 0x27, 0x18, 0x09, 0x22, 0x15, 0x27, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x31, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x21, 0x27, 0x0a, 0x33, 0x7a, 0x31,
	0x12, 0x2f, 0x12, 0x14, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64, 0x20, 0x21, 0x3d, 0x20, 0x27,
	0x73, 0x6f, 0x6d, 0x65, 0x2d, 0x69, 0x64, 0x27, 0x18, 0x09, 0x22, 0x15, 0x27, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x21,
	0x27, 0x0a, 0xe0, 0x01, 0x7a, 0xdd, 0x01, 0x12, 0xda, 0x01, 0x12, 0x14, 0x24, 0x2e, 0x69, 0x64,
	0x20, 0x21, 0x3d, 0x20, 0x27, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x63, 0x74, 0x2d, 0x69, 0x64, 0x27,
	0x18, 0x09, 0x22, 0x15, 0x27, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x33,
	0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x21, 0x27, 0x2a, 0xa8, 0x01, 0x1a, 0x25, 0x0a, 0x0d,
	0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0xf2, 0x01, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x31, 0x1a, 0x25, 0x0a, 0x0d, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0xf2,
	0x01, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x42, 0x24, 0x0a, 0x22, 0x0a, 0x07,
	0x27, 0x74, 0x79, 0x70, 0x65, 0x31, 0x27, 0x12, 0x07, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64,
	0x1a, 0x0e, 0x27, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x31, 0x27,
	0x4a, 0x1b, 0x0a, 0x19, 0x0a, 0x07, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64, 0x12, 0x0e, 0x27,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0x27, 0x6a, 0x15, 0x0a,
	0x05, 0x65, 0x6e, 0x2d, 0x55, 0x53, 0x12, 0x0c, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x0a, 0x4b, 0x7a, 0x49, 0x12, 0x47, 0x0a, 0x21, 0x0a, 0x09, 0x63, 0x6f,
	0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5a, 0x14, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64,
	0x20, 0x21, 0x3d, 0x20, 0x27, 0x73, 0x6f, 0x6d, 0x65, 0x2d, 0x69, 0x64, 0x27, 0x12, 0x09, 0x63,
	0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x09, 0x22, 0x15, 0x27, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x34, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x21,
	0x27, 0x22, 0x7b, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0d, 0x9a, 0x4a, 0x0a, 0xf2, 0x01, 0x07, 0x73, 0x6f, 0x6d,
	0x65, 0x2d, 0x69, 0x64, 0x52, 0x02, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x10, 0x9a, 0x4a, 0x0d, 0xf2, 0x01, 0x0a, 0x73,
	0x6f, 0x6d, 0x65, 0x2d, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65,
	0x12, 0x2c, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x12, 0x9a, 0x4a, 0x0f, 0xf2, 0x01, 0x0c, 0x73, 0x6f, 0x6d, 0x65, 0x2d, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x39,
	0x0a, 0x0d, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x28, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x0e, 0x9a, 0x4a, 0x0b, 0x12, 0x09, 0x24, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x37, 0x0a, 0x14, 0x43, 0x75, 0x73,
	0x74, 0x6f, 0x6d, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x3a, 0x1f, 0x9a, 0x4a, 0x1c, 0x0a, 0x18, 0x7a, 0x16, 0x12, 0x14, 0x12, 0x10, 0x24, 0x2e,
	0x61, 0x72, 0x67, 0x20, 0x3d, 0x3d, 0x20, 0x27, 0x77, 0x72, 0x6f, 0x6e, 0x67, 0x27, 0x18, 0x09,
	0x10, 0x01, 0x32, 0x66, 0x0a, 0x11, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4c, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x50, 0x6f,
	0x73, 0x74, 0x12, 0x1e, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x1a, 0x03, 0x9a, 0x4a, 0x00, 0x42, 0x9d, 0x01, 0x0a, 0x12, 0x63,
	0x6f, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x0f, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x1d, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x66, 0x65,
	0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0xa2, 0x02, 0x03, 0x4f, 0x46, 0x58, 0xaa, 0x02, 0x0e, 0x4f, 0x72, 0x67, 0x2e,
	0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xca, 0x02, 0x0e, 0x4f, 0x72, 0x67,
	0x5c, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xe2, 0x02, 0x1a, 0x4f, 0x72,
	0x67, 0x5c, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f, 0x4f, 0x72, 0x67, 0x3a, 0x3a,
	0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_federation_federation_proto_rawDescOnce sync.Once
	file_federation_federation_proto_rawDescData = file_federation_federation_proto_rawDesc
)

func file_federation_federation_proto_rawDescGZIP() []byte {
	file_federation_federation_proto_rawDescOnce.Do(func() {
		file_federation_federation_proto_rawDescData = protoimpl.X.CompressGZIP(file_federation_federation_proto_rawDescData)
	})
	return file_federation_federation_proto_rawDescData
}

var file_federation_federation_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_federation_federation_proto_goTypes = []interface{}{
	(*GetPostRequest)(nil),       // 0: org.federation.GetPostRequest
	(*GetPostResponse)(nil),      // 1: org.federation.GetPostResponse
	(*Post)(nil),                 // 2: org.federation.Post
	(*CustomMessage)(nil),        // 3: org.federation.CustomMessage
	(*CustomHandlerMessage)(nil), // 4: org.federation.CustomHandlerMessage
}
var file_federation_federation_proto_depIdxs = []int32{
	2, // 0: org.federation.GetPostResponse.post:type_name -> org.federation.Post
	0, // 1: org.federation.FederationService.GetPost:input_type -> org.federation.GetPostRequest
	1, // 2: org.federation.FederationService.GetPost:output_type -> org.federation.GetPostResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_federation_federation_proto_init() }
func file_federation_federation_proto_init() {
	if File_federation_federation_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_federation_federation_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPostRequest); i {
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
		file_federation_federation_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPostResponse); i {
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
		file_federation_federation_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Post); i {
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
		file_federation_federation_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CustomMessage); i {
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
		file_federation_federation_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CustomHandlerMessage); i {
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
			RawDescriptor: file_federation_federation_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_federation_federation_proto_goTypes,
		DependencyIndexes: file_federation_federation_proto_depIdxs,
		MessageInfos:      file_federation_federation_proto_msgTypes,
	}.Build()
	File_federation_federation_proto = out.File
	file_federation_federation_proto_rawDesc = nil
	file_federation_federation_proto_goTypes = nil
	file_federation_federation_proto_depIdxs = nil
}
