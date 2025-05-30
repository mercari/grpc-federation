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
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PostV2DevType int32

const (
	PostV2DevType_POST_V2_DEV_TYPE PostV2DevType = 0
)

// Enum value maps for PostV2DevType.
var (
	PostV2DevType_name = map[int32]string{
		0: "POST_V2_DEV_TYPE",
	}
	PostV2DevType_value = map[string]int32{
		"POST_V2_DEV_TYPE": 0,
	}
)

func (x PostV2DevType) Enum() *PostV2DevType {
	p := new(PostV2DevType)
	*p = x
	return p
}

func (x PostV2DevType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PostV2DevType) Descriptor() protoreflect.EnumDescriptor {
	return file_federation_federation_proto_enumTypes[0].Descriptor()
}

func (PostV2DevType) Type() protoreflect.EnumType {
	return &file_federation_federation_proto_enumTypes[0]
}

func (x PostV2DevType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PostV2DevType.Descriptor instead.
func (PostV2DevType) EnumDescriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0}
}

type GetPostV2DevRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetPostV2DevRequest) Reset() {
	*x = GetPostV2DevRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostV2DevRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostV2DevRequest) ProtoMessage() {}

func (x *GetPostV2DevRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use GetPostV2DevRequest.ProtoReflect.Descriptor instead.
func (*GetPostV2DevRequest) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0}
}

func (x *GetPostV2DevRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetPostV2DevResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Post      *PostV2Dev           `protobuf:"bytes,1,opt,name=post,proto3" json:"post,omitempty"`
	Type      PostV2DevType        `protobuf:"varint,2,opt,name=type,proto3,enum=federation.v2dev.PostV2DevType" json:"type,omitempty"`
	EnvA      string               `protobuf:"bytes,3,opt,name=env_a,json=envA,proto3" json:"env_a,omitempty"`
	EnvB      int64                `protobuf:"varint,4,opt,name=env_b,json=envB,proto3" json:"env_b,omitempty"`
	EnvCValue *durationpb.Duration `protobuf:"bytes,5,opt,name=env_c_value,json=envCValue,proto3" json:"env_c_value,omitempty"`
	Ref       *Ref                 `protobuf:"bytes,6,opt,name=ref,proto3" json:"ref,omitempty"`
}

func (x *GetPostV2DevResponse) Reset() {
	*x = GetPostV2DevResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostV2DevResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostV2DevResponse) ProtoMessage() {}

func (x *GetPostV2DevResponse) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use GetPostV2DevResponse.ProtoReflect.Descriptor instead.
func (*GetPostV2DevResponse) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{1}
}

func (x *GetPostV2DevResponse) GetPost() *PostV2Dev {
	if x != nil {
		return x.Post
	}
	return nil
}

func (x *GetPostV2DevResponse) GetType() PostV2DevType {
	if x != nil {
		return x.Type
	}
	return PostV2DevType_POST_V2_DEV_TYPE
}

func (x *GetPostV2DevResponse) GetEnvA() string {
	if x != nil {
		return x.EnvA
	}
	return ""
}

func (x *GetPostV2DevResponse) GetEnvB() int64 {
	if x != nil {
		return x.EnvB
	}
	return 0
}

func (x *GetPostV2DevResponse) GetEnvCValue() *durationpb.Duration {
	if x != nil {
		return x.EnvCValue
	}
	return nil
}

func (x *GetPostV2DevResponse) GetRef() *Ref {
	if x != nil {
		return x.Ref
	}
	return nil
}

type PostV2Dev struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title     string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Content   string `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
	User      *User  `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
	NullCheck bool   `protobuf:"varint,5,opt,name=null_check,json=nullCheck,proto3" json:"null_check,omitempty"`
}

func (x *PostV2Dev) Reset() {
	*x = PostV2Dev{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostV2Dev) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostV2Dev) ProtoMessage() {}

func (x *PostV2Dev) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PostV2Dev.ProtoReflect.Descriptor instead.
func (*PostV2Dev) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{2}
}

func (x *PostV2Dev) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *PostV2Dev) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *PostV2Dev) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *PostV2Dev) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *PostV2Dev) GetNullCheck() bool {
	if x != nil {
		return x.NullCheck
	}
	return false
}

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{3}
}

func (x *User) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *User) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type Unused struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Foo string `protobuf:"bytes,1,opt,name=foo,proto3" json:"foo,omitempty"`
}

func (x *Unused) Reset() {
	*x = Unused{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Unused) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Unused) ProtoMessage() {}

func (x *Unused) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Unused.ProtoReflect.Descriptor instead.
func (*Unused) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{4}
}

func (x *Unused) GetFoo() string {
	if x != nil {
		return x.Foo
	}
	return ""
}

type ForNameless struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bar string `protobuf:"bytes,1,opt,name=bar,proto3" json:"bar,omitempty"`
}

func (x *ForNameless) Reset() {
	*x = ForNameless{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForNameless) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForNameless) ProtoMessage() {}

func (x *ForNameless) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForNameless.ProtoReflect.Descriptor instead.
func (*ForNameless) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{5}
}

func (x *ForNameless) GetBar() string {
	if x != nil {
		return x.Bar
	}
	return ""
}

type TypedNil struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TypedNil) Reset() {
	*x = TypedNil{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TypedNil) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TypedNil) ProtoMessage() {}

func (x *TypedNil) ProtoReflect() protoreflect.Message {
	mi := &file_federation_federation_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TypedNil.ProtoReflect.Descriptor instead.
func (*TypedNil) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{6}
}

var File_federation_federation_proto protoreflect.FileDescriptor

var file_federation_federation_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x66, 0x65, 0x64,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x1a,
	0x20, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x16, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6f, 0x74,
	0x68, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x25, 0x0a, 0x13, 0x47, 0x65, 0x74,
	0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x22, 0xd9, 0x03, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65,
	0x76, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3a, 0x0a, 0x04, 0x70, 0x6f, 0x73,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x56,
	0x32, 0x64, 0x65, 0x76, 0x42, 0x09, 0x9a, 0x4a, 0x06, 0x12, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x52,
	0x04, 0x70, 0x6f, 0x73, 0x74, 0x12, 0x61, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76,
	0x54, 0x79, 0x70, 0x65, 0x42, 0x2c, 0x9a, 0x4a, 0x29, 0x12, 0x27, 0x50, 0x6f, 0x73, 0x74, 0x56,
	0x32, 0x64, 0x65, 0x76, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x28, 0x27,
	0x50, 0x4f, 0x53, 0x54, 0x5f, 0x56, 0x32, 0x5f, 0x44, 0x45, 0x56, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x27, 0x29, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x2f, 0x0a, 0x05, 0x65, 0x6e, 0x76, 0x5f,
	0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0x9a, 0x4a, 0x17, 0x12, 0x15, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x65, 0x6e,
	0x76, 0x2e, 0x61, 0x52, 0x04, 0x65, 0x6e, 0x76, 0x41, 0x12, 0x32, 0x0a, 0x05, 0x65, 0x6e, 0x76,
	0x5f, 0x62, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x42, 0x1d, 0x9a, 0x4a, 0x1a, 0x12, 0x18, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x65,
	0x6e, 0x76, 0x2e, 0x62, 0x5b, 0x31, 0x5d, 0x52, 0x04, 0x65, 0x6e, 0x76, 0x42, 0x12, 0x5a, 0x0a,
	0x0b, 0x65, 0x6e, 0x76, 0x5f, 0x63, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x1f, 0x9a,
	0x4a, 0x1c, 0x12, 0x1a, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x65, 0x6e, 0x76, 0x2e, 0x63, 0x5b, 0x27, 0x7a, 0x27, 0x5d, 0x52, 0x09,
	0x65, 0x6e, 0x76, 0x43, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2f, 0x0a, 0x03, 0x72, 0x65, 0x66,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x2e, 0x52, 0x65, 0x66, 0x42, 0x06, 0x9a,
	0x4a, 0x03, 0x12, 0x01, 0x72, 0x52, 0x03, 0x72, 0x65, 0x66, 0x3a, 0x30, 0x9a, 0x4a, 0x2d, 0x0a,
	0x1f, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x6a, 0x17, 0x0a, 0x09, 0x50, 0x6f, 0x73, 0x74, 0x56,
	0x32, 0x64, 0x65, 0x76, 0x12, 0x0a, 0x0a, 0x02, 0x69, 0x64, 0x12, 0x04, 0x24, 0x2e, 0x69, 0x64,
	0x0a, 0x0a, 0x0a, 0x01, 0x72, 0x6a, 0x05, 0x0a, 0x03, 0x52, 0x65, 0x66, 0x22, 0xf3, 0x03, 0x0a,
	0x09, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x31, 0x0a, 0x04, 0x75, 0x73,
	0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x2e, 0x55, 0x73, 0x65, 0x72,
	0x42, 0x05, 0x9a, 0x4a, 0x02, 0x08, 0x01, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x2e, 0x0a,
	0x0a, 0x6e, 0x75, 0x6c, 0x6c, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x08, 0x42, 0x0f, 0x9a, 0x4a, 0x0c, 0x12, 0x0a, 0x6e, 0x75, 0x6c, 0x6c, 0x5f, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x52, 0x09, 0x6e, 0x75, 0x6c, 0x6c, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x3a, 0xc2, 0x02,
	0x9a, 0x4a, 0xbe, 0x02, 0x0a, 0x2d, 0x0a, 0x03, 0x72, 0x65, 0x73, 0x72, 0x26, 0x0a, 0x18, 0x70,
	0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f,
	0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0a, 0x0a, 0x02, 0x69, 0x64, 0x12, 0x04, 0x24,
	0x2e, 0x69, 0x64, 0x0a, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x5a, 0x08, 0x72,
	0x65, 0x73, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x0a, 0x16, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x6a,
	0x0e, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x06, 0x1a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x0a,
	0x20, 0x0a, 0x06, 0x75, 0x6e, 0x75, 0x73, 0x65, 0x64, 0x6a, 0x16, 0x0a, 0x06, 0x55, 0x6e, 0x75,
	0x73, 0x65, 0x64, 0x12, 0x0c, 0x0a, 0x03, 0x66, 0x6f, 0x6f, 0x12, 0x05, 0x27, 0x66, 0x6f, 0x6f,
	0x27, 0x0a, 0x1d, 0x6a, 0x1b, 0x0a, 0x0b, 0x46, 0x6f, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x6c, 0x65,
	0x73, 0x73, 0x12, 0x0c, 0x0a, 0x03, 0x62, 0x61, 0x72, 0x12, 0x05, 0x27, 0x62, 0x61, 0x72, 0x27,
	0x0a, 0x17, 0x0a, 0x09, 0x74, 0x79, 0x70, 0x65, 0x64, 0x5f, 0x6e, 0x69, 0x6c, 0x6a, 0x0a, 0x0a,
	0x08, 0x54, 0x79, 0x70, 0x65, 0x64, 0x4e, 0x69, 0x6c, 0x0a, 0x25, 0x0a, 0x0a, 0x6e, 0x75, 0x6c,
	0x6c, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x11, 0x74, 0x79, 0x70, 0x65, 0x64, 0x5f, 0x6e,
	0x69, 0x6c, 0x20, 0x3d, 0x3d, 0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x5a, 0x04, 0x74, 0x72, 0x75, 0x65,
	0x0a, 0x60, 0x12, 0x11, 0x74, 0x79, 0x70, 0x65, 0x64, 0x5f, 0x6e, 0x69, 0x6c, 0x20, 0x3d, 0x3d,
	0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x5a, 0x4b, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x66, 0x65, 0x64, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x6c, 0x6f, 0x67, 0x2e, 0x69, 0x6e, 0x66, 0x6f, 0x28,
	0x27, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x20, 0x74, 0x79, 0x70, 0x65, 0x64, 0x5f, 0x6e, 0x69,
	0x6c, 0x27, 0x2c, 0x20, 0x7b, 0x27, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x27, 0x3a, 0x20, 0x74,
	0x79, 0x70, 0x65, 0x64, 0x5f, 0x6e, 0x69, 0x6c, 0x20, 0x3d, 0x3d, 0x20, 0x6e, 0x75, 0x6c, 0x6c,
	0x7d, 0x29, 0x22, 0x7b, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x05, 0x9a, 0x4a, 0x02, 0x08, 0x01, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x3a, 0x48, 0x9a, 0x4a, 0x45, 0x0a, 0x32, 0x0a, 0x03, 0x72, 0x65,
	0x73, 0x72, 0x2b, 0x0a, 0x18, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x12, 0x0f, 0x0a,
	0x02, 0x69, 0x64, 0x12, 0x09, 0x24, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x0a, 0x0d,
	0x0a, 0x01, 0x75, 0x5a, 0x08, 0x72, 0x65, 0x73, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x10, 0x01, 0x22,
	0x21, 0x0a, 0x06, 0x55, 0x6e, 0x75, 0x73, 0x65, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x6f, 0x6f,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x66, 0x6f, 0x6f, 0x3a, 0x05, 0x9a, 0x4a, 0x02,
	0x10, 0x01, 0x22, 0x26, 0x0a, 0x0b, 0x46, 0x6f, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x73,
	0x73, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x61, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x62, 0x61, 0x72, 0x3a, 0x05, 0x9a, 0x4a, 0x02, 0x10, 0x01, 0x22, 0x11, 0x0a, 0x08, 0x54, 0x79,
	0x70, 0x65, 0x64, 0x4e, 0x69, 0x6c, 0x3a, 0x05, 0x9a, 0x4a, 0x02, 0x10, 0x01, 0x2a, 0x25, 0x0a,
	0x0d, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14,
	0x0a, 0x10, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x56, 0x32, 0x5f, 0x44, 0x45, 0x56, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x10, 0x00, 0x32, 0xeb, 0x01, 0x0a, 0x16, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x56, 0x32, 0x64, 0x65, 0x76, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x5f, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x12,
	0x25, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64,
	0x65, 0x76, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73,
	0x74, 0x56, 0x32, 0x64, 0x65, 0x76, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x1a, 0x70, 0x9a, 0x4a, 0x6d, 0x0a, 0x49, 0x0a, 0x0e, 0x0a, 0x01, 0x61, 0x12, 0x02, 0x08, 0x01,
	0x1a, 0x05, 0x12, 0x03, 0x78, 0x78, 0x78, 0x0a, 0x10, 0x0a, 0x01, 0x62, 0x12, 0x04, 0x12, 0x02,
	0x08, 0x03, 0x1a, 0x05, 0x0a, 0x03, 0x79, 0x79, 0x79, 0x0a, 0x18, 0x0a, 0x01, 0x63, 0x12, 0x0a,
	0x1a, 0x08, 0x0a, 0x02, 0x08, 0x01, 0x12, 0x02, 0x08, 0x06, 0x1a, 0x07, 0x0a, 0x03, 0x7a, 0x7a,
	0x7a, 0x18, 0x01, 0x0a, 0x0b, 0x0a, 0x01, 0x64, 0x12, 0x02, 0x08, 0x05, 0x1a, 0x02, 0x20, 0x01,
	0x12, 0x20, 0x0a, 0x1b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x5a,
	0x01, 0x31, 0x42, 0xcc, 0x01, 0x9a, 0x4a, 0x22, 0x12, 0x0f, 0x70, 0x6f, 0x73, 0x74, 0x2f, 0x70,
	0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x75, 0x73, 0x65, 0x72, 0x2f,
	0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x0a, 0x14, 0x63, 0x6f, 0x6d, 0x2e,
	0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x64, 0x65, 0x76,
	0x42, 0x0f, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x1d, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x66, 0x65, 0x64,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0xa2, 0x02, 0x03, 0x46, 0x56, 0x58, 0xaa, 0x02, 0x10, 0x46, 0x65, 0x64, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x56, 0x32, 0x64, 0x65, 0x76, 0xca, 0x02, 0x10, 0x46, 0x65,
	0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5c, 0x56, 0x32, 0x64, 0x65, 0x76, 0xe2, 0x02,
	0x1c, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5c, 0x56, 0x32, 0x64, 0x65,
	0x76, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x11,
	0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x3a, 0x56, 0x32, 0x64, 0x65,
	0x76, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_federation_federation_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_federation_federation_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_federation_federation_proto_goTypes = []interface{}{
	(PostV2DevType)(0),           // 0: federation.v2dev.PostV2devType
	(*GetPostV2DevRequest)(nil),  // 1: federation.v2dev.GetPostV2devRequest
	(*GetPostV2DevResponse)(nil), // 2: federation.v2dev.GetPostV2devResponse
	(*PostV2Dev)(nil),            // 3: federation.v2dev.PostV2dev
	(*User)(nil),                 // 4: federation.v2dev.User
	(*Unused)(nil),               // 5: federation.v2dev.Unused
	(*ForNameless)(nil),          // 6: federation.v2dev.ForNameless
	(*TypedNil)(nil),             // 7: federation.v2dev.TypedNil
	(*durationpb.Duration)(nil),  // 8: google.protobuf.Duration
	(*Ref)(nil),                  // 9: federation.v2dev.Ref
}
var file_federation_federation_proto_depIdxs = []int32{
	3, // 0: federation.v2dev.GetPostV2devResponse.post:type_name -> federation.v2dev.PostV2dev
	0, // 1: federation.v2dev.GetPostV2devResponse.type:type_name -> federation.v2dev.PostV2devType
	8, // 2: federation.v2dev.GetPostV2devResponse.env_c_value:type_name -> google.protobuf.Duration
	9, // 3: federation.v2dev.GetPostV2devResponse.ref:type_name -> federation.v2dev.Ref
	4, // 4: federation.v2dev.PostV2dev.user:type_name -> federation.v2dev.User
	1, // 5: federation.v2dev.FederationV2devService.GetPostV2dev:input_type -> federation.v2dev.GetPostV2devRequest
	2, // 6: federation.v2dev.FederationV2devService.GetPostV2dev:output_type -> federation.v2dev.GetPostV2devResponse
	6, // [6:7] is the sub-list for method output_type
	5, // [5:6] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_federation_federation_proto_init() }
func file_federation_federation_proto_init() {
	if File_federation_federation_proto != nil {
		return
	}
	file_federation_other_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_federation_federation_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPostV2DevRequest); i {
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
			switch v := v.(*GetPostV2DevResponse); i {
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
			switch v := v.(*PostV2Dev); i {
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
			switch v := v.(*User); i {
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
			switch v := v.(*Unused); i {
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
		file_federation_federation_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForNameless); i {
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
		file_federation_federation_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TypedNil); i {
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
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_federation_federation_proto_goTypes,
		DependencyIndexes: file_federation_federation_proto_depIdxs,
		EnumInfos:         file_federation_federation_proto_enumTypes,
		MessageInfos:      file_federation_federation_proto_msgTypes,
	}.Build()
	File_federation_federation_proto = out.File
	file_federation_federation_proto_rawDesc = nil
	file_federation_federation_proto_goTypes = nil
	file_federation_federation_proto_depIdxs = nil
}
