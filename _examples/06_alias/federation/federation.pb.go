// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: federation/federation.proto

package federation

import (
	_ "example/post"
	_ "example/post/v2"
	_ "github.com/mercari/grpc-federation/grpc/federation"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PostType int32

const (
	PostType_POST_TYPE_UNKNOWN PostType = 0
	PostType_POST_TYPE_FOO     PostType = 1
	PostType_POST_TYPE_BAR     PostType = 2
)

// Enum value maps for PostType.
var (
	PostType_name = map[int32]string{
		0: "POST_TYPE_UNKNOWN",
		1: "POST_TYPE_FOO",
		2: "POST_TYPE_BAR",
	}
	PostType_value = map[string]int32{
		"POST_TYPE_UNKNOWN": 0,
		"POST_TYPE_FOO":     1,
		"POST_TYPE_BAR":     2,
	}
)

func (x PostType) Enum() *PostType {
	p := new(PostType)
	*p = x
	return p
}

func (x PostType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PostType) Descriptor() protoreflect.EnumDescriptor {
	return file_federation_federation_proto_enumTypes[0].Descriptor()
}

func (PostType) Type() protoreflect.EnumType {
	return &file_federation_federation_proto_enumTypes[0]
}

func (x PostType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PostType.Descriptor instead.
func (PostType) EnumDescriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0}
}

type PostContent_Category int32

const (
	PostContent_CATEGORY_A PostContent_Category = 0
	PostContent_CATEGORY_B PostContent_Category = 1
)

// Enum value maps for PostContent_Category.
var (
	PostContent_Category_name = map[int32]string{
		0: "CATEGORY_A",
		1: "CATEGORY_B",
	}
	PostContent_Category_value = map[string]int32{
		"CATEGORY_A": 0,
		"CATEGORY_B": 1,
	}
)

func (x PostContent_Category) Enum() *PostContent_Category {
	p := new(PostContent_Category)
	*p = x
	return p
}

func (x PostContent_Category) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PostContent_Category) Descriptor() protoreflect.EnumDescriptor {
	return file_federation_federation_proto_enumTypes[1].Descriptor()
}

func (PostContent_Category) Type() protoreflect.EnumType {
	return &file_federation_federation_proto_enumTypes[1]
}

func (x PostContent_Category) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PostContent_Category.Descriptor instead.
func (PostContent_Category) EnumDescriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{4, 0}
}

type GetPostRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Types that are assignable to Condition:
	//
	//	*GetPostRequest_A
	//	*GetPostRequest_ConditionB_
	Condition isGetPostRequest_Condition `protobuf_oneof:"condition"`
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

func (m *GetPostRequest) GetCondition() isGetPostRequest_Condition {
	if m != nil {
		return m.Condition
	}
	return nil
}

func (x *GetPostRequest) GetA() *GetPostRequest_ConditionA {
	if x, ok := x.GetCondition().(*GetPostRequest_A); ok {
		return x.A
	}
	return nil
}

func (x *GetPostRequest) GetConditionB() *GetPostRequest_ConditionB {
	if x, ok := x.GetCondition().(*GetPostRequest_ConditionB_); ok {
		return x.ConditionB
	}
	return nil
}

type isGetPostRequest_Condition interface {
	isGetPostRequest_Condition()
}

type GetPostRequest_A struct {
	A *GetPostRequest_ConditionA `protobuf:"bytes,2,opt,name=a,proto3,oneof"`
}

type GetPostRequest_ConditionB_ struct {
	ConditionB *GetPostRequest_ConditionB `protobuf:"bytes,3,opt,name=condition_b,json=conditionB,proto3,oneof"`
}

func (*GetPostRequest_A) isGetPostRequest_Condition() {}

func (*GetPostRequest_ConditionB_) isGetPostRequest_Condition() {}

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

	Id    string    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data  *PostData `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Data2 *PostData `protobuf:"bytes,3,opt,name=data2,proto3" json:"data2,omitempty"`
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

func (x *Post) GetData() *PostData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Post) GetData2() *PostData {
	if x != nil {
		return x.Data2
	}
	return nil
}

type PostData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type    PostType     `protobuf:"varint,1,opt,name=type,proto3,enum=org.federation.PostType" json:"type,omitempty"`
	Title   string       `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Content *PostContent `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *PostData) Reset() {
	*x = PostData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostData) ProtoMessage() {}

func (x *PostData) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PostData.ProtoReflect.Descriptor instead.
func (*PostData) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{3}
}

func (x *PostData) GetType() PostType {
	if x != nil {
		return x.Type
	}
	return PostType_POST_TYPE_UNKNOWN
}

func (x *PostData) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *PostData) GetContent() *PostContent {
	if x != nil {
		return x.Content
	}
	return nil
}

type PostContent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Category PostContent_Category `protobuf:"varint,1,opt,name=category,proto3,enum=org.federation.PostContent_Category" json:"category,omitempty"`
	Head     string               `protobuf:"bytes,2,opt,name=head,proto3" json:"head,omitempty"`
	Body     string               `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
	DupBody  string               `protobuf:"bytes,4,opt,name=dup_body,json=dupBody,proto3" json:"dup_body,omitempty"`
	Counts   map[int32]int32      `protobuf:"bytes,5,rep,name=counts,proto3" json:"counts,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *PostContent) Reset() {
	*x = PostContent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostContent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostContent) ProtoMessage() {}

func (x *PostContent) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PostContent.ProtoReflect.Descriptor instead.
func (*PostContent) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{4}
}

func (x *PostContent) GetCategory() PostContent_Category {
	if x != nil {
		return x.Category
	}
	return PostContent_CATEGORY_A
}

func (x *PostContent) GetHead() string {
	if x != nil {
		return x.Head
	}
	return ""
}

func (x *PostContent) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

func (x *PostContent) GetDupBody() string {
	if x != nil {
		return x.DupBody
	}
	return ""
}

func (x *PostContent) GetCounts() map[int32]int32 {
	if x != nil {
		return x.Counts
	}
	return nil
}

type GetPostRequest_ConditionA struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prop string `protobuf:"bytes,1,opt,name=prop,proto3" json:"prop,omitempty"`
}

func (x *GetPostRequest_ConditionA) Reset() {
	*x = GetPostRequest_ConditionA{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostRequest_ConditionA) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostRequest_ConditionA) ProtoMessage() {}

func (x *GetPostRequest_ConditionA) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use GetPostRequest_ConditionA.ProtoReflect.Descriptor instead.
func (*GetPostRequest_ConditionA) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0, 0}
}

func (x *GetPostRequest_ConditionA) GetProp() string {
	if x != nil {
		return x.Prop
	}
	return ""
}

type GetPostRequest_ConditionB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetPostRequest_ConditionB) Reset() {
	*x = GetPostRequest_ConditionB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_federation_federation_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostRequest_ConditionB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostRequest_ConditionB) ProtoMessage() {}

func (x *GetPostRequest_ConditionB) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use GetPostRequest_ConditionB.ProtoReflect.Descriptor instead.
func (*GetPostRequest_ConditionB) Descriptor() ([]byte, []int) {
	return file_federation_federation_proto_rawDescGZIP(), []int{0, 1}
}

var File_federation_federation_proto protoreflect.FileDescriptor

var file_federation_federation_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x66, 0x65, 0x64,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6f,
	0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x20, 0x67,
	0x72, 0x70, 0x63, 0x2f, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0f, 0x70, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x12, 0x70, 0x6f, 0x73, 0x74, 0x2f, 0x76, 0x32, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa2, 0x02, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x01, 0x61, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x48, 0x00, 0x52,
	0x01, 0x61, 0x12, 0x4c, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x62, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65,
	0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x48, 0x00, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x42,
	0x1a, 0x3e, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x72, 0x6f, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x72,
	0x6f, 0x70, 0x3a, 0x1c, 0x9a, 0x4a, 0x19, 0x1a, 0x17, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73,
	0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x41,
	0x1a, 0x2a, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x3a, 0x1c,
	0x9a, 0x4a, 0x19, 0x1a, 0x17, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f,
	0x73, 0x74, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x42, 0x0b, 0x0a, 0x09,
	0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x85, 0x01, 0x0a, 0x0f, 0x47, 0x65,
	0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a,
	0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6f, 0x72,
	0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73,
	0x74, 0x42, 0x09, 0x9a, 0x4a, 0x06, 0x12, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x52, 0x04, 0x70, 0x6f,
	0x73, 0x74, 0x3a, 0x3d, 0x9a, 0x4a, 0x3a, 0x0a, 0x38, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x6a,
	0x30, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0a, 0x0a, 0x02, 0x69, 0x64, 0x12, 0x04, 0x24,
	0x2e, 0x69, 0x64, 0x12, 0x08, 0x0a, 0x01, 0x61, 0x12, 0x03, 0x24, 0x2e, 0x61, 0x12, 0x12, 0x0a,
	0x01, 0x62, 0x12, 0x0d, 0x24, 0x2e, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x62, 0x22, 0xce, 0x02, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x2c, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x3a, 0x0a, 0x05, 0x64, 0x61, 0x74, 0x61,
	0x32, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65,
	0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74,
	0x61, 0x42, 0x0a, 0x9a, 0x4a, 0x07, 0x12, 0x05, 0x64, 0x61, 0x74, 0x61, 0x32, 0x52, 0x05, 0x64,
	0x61, 0x74, 0x61, 0x32, 0x3a, 0xcb, 0x01, 0x9a, 0x4a, 0xc7, 0x01, 0x0a, 0x61, 0x0a, 0x03, 0x72,
	0x65, 0x73, 0x72, 0x5a, 0x0a, 0x1c, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50,
	0x6f, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x47, 0x65, 0x74, 0x50, 0x6f,
	0x73, 0x74, 0x12, 0x0a, 0x0a, 0x02, 0x69, 0x64, 0x12, 0x04, 0x24, 0x2e, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x01, 0x61, 0x12, 0x03, 0x24, 0x2e, 0x61, 0xba, 0x02, 0x0b, 0x24, 0x2e, 0x61, 0x20, 0x21,
	0x3d, 0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x12, 0x16, 0x0a, 0x01, 0x62, 0x12, 0x03, 0x24, 0x2e, 0x62,
	0xba, 0x02, 0x0b, 0x24, 0x2e, 0x62, 0x20, 0x21, 0x3d, 0x20, 0x6e, 0x75, 0x6c, 0x6c, 0x0a, 0x12,
	0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x5a, 0x08, 0x72, 0x65, 0x73, 0x2e, 0x70, 0x6f,
	0x73, 0x74, 0x0a, 0x35, 0x0a, 0x04, 0x72, 0x65, 0x73, 0x32, 0x72, 0x2d, 0x0a, 0x1f, 0x6f, 0x72,
	0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0a, 0x0a,
	0x02, 0x69, 0x64, 0x12, 0x04, 0x24, 0x2e, 0x69, 0x64, 0x0a, 0x17, 0x0a, 0x05, 0x64, 0x61, 0x74,
	0x61, 0x32, 0x5a, 0x0e, 0x72, 0x65, 0x73, 0x32, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x64, 0x61,
	0x74, 0x61, 0x22, 0xb3, 0x01, 0x0a, 0x08, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12,
	0x2c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e,
	0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50,
	0x6f, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x12, 0x35, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x3a, 0x2c, 0x9a, 0x4a, 0x29, 0x1a,
	0x11, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x1a, 0x14, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76, 0x32, 0x2e,
	0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x22, 0xbf, 0x03, 0x0a, 0x0b, 0x50, 0x6f, 0x73,
	0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x40, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65,
	0x67, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x6f, 0x72, 0x67,
	0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73, 0x74,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x52, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x65,
	0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x6f,
	0x64, 0x79, 0x12, 0x24, 0x0a, 0x08, 0x64, 0x75, 0x70, 0x5f, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0x9a, 0x4a, 0x06, 0x1a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x52,
	0x07, 0x64, 0x75, 0x70, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x3f, 0x0a, 0x06, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x06, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x70, 0x0a, 0x08, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x41, 0x10, 0x00,
	0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x42, 0x10, 0x01,
	0x1a, 0x44, 0x9a, 0x4a, 0x41, 0x0a, 0x1d, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e,
	0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x61, 0x74, 0x65,
	0x67, 0x6f, 0x72, 0x79, 0x0a, 0x20, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76,
	0x32, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x61,
	0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x3a, 0x32, 0x9a, 0x4a, 0x2f, 0x1a, 0x14, 0x6f, 0x72, 0x67,
	0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x1a, 0x17, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x50,
	0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2a, 0xb6, 0x02, 0x0a, 0x08, 0x50,
	0x6f, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x11, 0x50, 0x4f, 0x53, 0x54, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x1a, 0x05,
	0x9a, 0x4a, 0x02, 0x08, 0x01, 0x12, 0x23, 0x0a, 0x0d, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x46, 0x4f, 0x4f, 0x10, 0x01, 0x1a, 0x10, 0x9a, 0x4a, 0x0d, 0x12, 0x0b, 0x50,
	0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x41, 0x12, 0xb0, 0x01, 0x0a, 0x0d, 0x50,
	0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x42, 0x41, 0x52, 0x10, 0x02, 0x1a, 0x9c,
	0x01, 0x9a, 0x4a, 0x98, 0x01, 0x12, 0x21, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e,
	0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x50, 0x4f, 0x53,
	0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x42, 0x12, 0x21, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f,
	0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x2e,
	0x50, 0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x12, 0x27, 0x6f, 0x72, 0x67,
	0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74,
	0x61, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x56, 0x32, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x42, 0x12, 0x27, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x76,
	0x32, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x50,
	0x4f, 0x53, 0x54, 0x5f, 0x56, 0x32, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x1a, 0x34, 0x9a,
	0x4a, 0x31, 0x0a, 0x15, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73,
	0x74, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x0a, 0x18, 0x6f, 0x72, 0x67, 0x2e, 0x70,
	0x6f, 0x73, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x54,
	0x79, 0x70, 0x65, 0x32, 0x66, 0x0a, 0x11, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4c, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x50,
	0x6f, 0x73, 0x74, 0x12, 0x1e, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x1a, 0x03, 0x9a, 0x4a, 0x00, 0x42, 0x9d, 0x01, 0x0a, 0x12,
	0x63, 0x6f, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x42, 0x0f, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x1d, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x66,
	0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3b, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0xa2, 0x02, 0x03, 0x4f, 0x46, 0x58, 0xaa, 0x02, 0x0e, 0x4f, 0x72, 0x67,
	0x2e, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xca, 0x02, 0x0e, 0x4f, 0x72,
	0x67, 0x5c, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xe2, 0x02, 0x1a, 0x4f,
	0x72, 0x67, 0x5c, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f, 0x4f, 0x72, 0x67, 0x3a,
	0x3a, 0x46, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
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

var file_federation_federation_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_federation_federation_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_federation_federation_proto_goTypes = []interface{}{
	(PostType)(0),                     // 0: org.federation.PostType
	(PostContent_Category)(0),         // 1: org.federation.PostContent.Category
	(*GetPostRequest)(nil),            // 2: org.federation.GetPostRequest
	(*GetPostResponse)(nil),           // 3: org.federation.GetPostResponse
	(*Post)(nil),                      // 4: org.federation.Post
	(*PostData)(nil),                  // 5: org.federation.PostData
	(*PostContent)(nil),               // 6: org.federation.PostContent
	(*GetPostRequest_ConditionA)(nil), // 7: org.federation.GetPostRequest.ConditionA
	(*GetPostRequest_ConditionB)(nil), // 8: org.federation.GetPostRequest.ConditionB
	nil,                               // 9: org.federation.PostContent.CountsEntry
}
var file_federation_federation_proto_depIdxs = []int32{
	7,  // 0: org.federation.GetPostRequest.a:type_name -> org.federation.GetPostRequest.ConditionA
	8,  // 1: org.federation.GetPostRequest.condition_b:type_name -> org.federation.GetPostRequest.ConditionB
	4,  // 2: org.federation.GetPostResponse.post:type_name -> org.federation.Post
	5,  // 3: org.federation.Post.data:type_name -> org.federation.PostData
	5,  // 4: org.federation.Post.data2:type_name -> org.federation.PostData
	0,  // 5: org.federation.PostData.type:type_name -> org.federation.PostType
	6,  // 6: org.federation.PostData.content:type_name -> org.federation.PostContent
	1,  // 7: org.federation.PostContent.category:type_name -> org.federation.PostContent.Category
	9,  // 8: org.federation.PostContent.counts:type_name -> org.federation.PostContent.CountsEntry
	2,  // 9: org.federation.FederationService.GetPost:input_type -> org.federation.GetPostRequest
	3,  // 10: org.federation.FederationService.GetPost:output_type -> org.federation.GetPostResponse
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
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
			switch v := v.(*PostData); i {
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
			switch v := v.(*PostContent); i {
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
			switch v := v.(*GetPostRequest_ConditionA); i {
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
			switch v := v.(*GetPostRequest_ConditionB); i {
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
	file_federation_federation_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*GetPostRequest_A)(nil),
		(*GetPostRequest_ConditionB_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_federation_federation_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
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
