// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: post/post.proto

package post

import (
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

type PostDataType int32

const (
	PostDataType_POST_TYPE_A PostDataType = 0
	PostDataType_POST_TYPE_B PostDataType = 1
	PostDataType_POST_TYPE_C PostDataType = 2
)

// Enum value maps for PostDataType.
var (
	PostDataType_name = map[int32]string{
		0: "POST_TYPE_A",
		1: "POST_TYPE_B",
		2: "POST_TYPE_C",
	}
	PostDataType_value = map[string]int32{
		"POST_TYPE_A": 0,
		"POST_TYPE_B": 1,
		"POST_TYPE_C": 2,
	}
)

func (x PostDataType) Enum() *PostDataType {
	p := new(PostDataType)
	*p = x
	return p
}

func (x PostDataType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PostDataType) Descriptor() protoreflect.EnumDescriptor {
	return file_post_post_proto_enumTypes[0].Descriptor()
}

func (PostDataType) Type() protoreflect.EnumType {
	return &file_post_post_proto_enumTypes[0]
}

func (x PostDataType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PostDataType.Descriptor instead.
func (PostDataType) EnumDescriptor() ([]byte, []int) {
	return file_post_post_proto_rawDescGZIP(), []int{0}
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
	return file_post_post_proto_enumTypes[1].Descriptor()
}

func (PostContent_Category) Type() protoreflect.EnumType {
	return &file_post_post_proto_enumTypes[1]
}

func (x PostContent_Category) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PostContent_Category.Descriptor instead.
func (PostContent_Category) EnumDescriptor() ([]byte, []int) {
	return file_post_post_proto_rawDescGZIP(), []int{4, 0}
}

type GetPostRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetPostRequest) Reset() {
	*x = GetPostRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_post_post_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostRequest) ProtoMessage() {}

func (x *GetPostRequest) ProtoReflect() protoreflect.Message {
	mi := &file_post_post_proto_msgTypes[0]
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
	return file_post_post_proto_rawDescGZIP(), []int{0}
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
		mi := &file_post_post_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPostResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPostResponse) ProtoMessage() {}

func (x *GetPostResponse) ProtoReflect() protoreflect.Message {
	mi := &file_post_post_proto_msgTypes[1]
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
	return file_post_post_proto_rawDescGZIP(), []int{1}
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

	Id   string    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data *PostData `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Post) Reset() {
	*x = Post{}
	if protoimpl.UnsafeEnabled {
		mi := &file_post_post_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Post) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Post) ProtoMessage() {}

func (x *Post) ProtoReflect() protoreflect.Message {
	mi := &file_post_post_proto_msgTypes[2]
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
	return file_post_post_proto_rawDescGZIP(), []int{2}
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

type PostData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type    PostDataType `protobuf:"varint,1,opt,name=type,proto3,enum=org.post.PostDataType" json:"type,omitempty"`
	Title   string       `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Content *PostContent `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *PostData) Reset() {
	*x = PostData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_post_post_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostData) ProtoMessage() {}

func (x *PostData) ProtoReflect() protoreflect.Message {
	mi := &file_post_post_proto_msgTypes[3]
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
	return file_post_post_proto_rawDescGZIP(), []int{3}
}

func (x *PostData) GetType() PostDataType {
	if x != nil {
		return x.Type
	}
	return PostDataType_POST_TYPE_A
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

	Category PostContent_Category `protobuf:"varint,1,opt,name=category,proto3,enum=org.post.PostContent_Category" json:"category,omitempty"`
	Head     string               `protobuf:"bytes,2,opt,name=head,proto3" json:"head,omitempty"`
	Body     string               `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
}

func (x *PostContent) Reset() {
	*x = PostContent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_post_post_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PostContent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostContent) ProtoMessage() {}

func (x *PostContent) ProtoReflect() protoreflect.Message {
	mi := &file_post_post_proto_msgTypes[4]
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
	return file_post_post_proto_rawDescGZIP(), []int{4}
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

var File_post_post_proto protoreflect.FileDescriptor

var file_post_post_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x08, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x22, 0x20, 0x0a, 0x0e, 0x47,
	0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x35, 0x0a,
	0x0f, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x22, 0x0a, 0x04, 0x70, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x04,
	0x70, 0x6f, 0x73, 0x74, 0x22, 0x3e, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6f, 0x72, 0x67,
	0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x7d, 0x0a, 0x08, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61,
	0x12, 0x2a, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16,
	0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x12, 0x2f, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x50,
	0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x22, 0x9d, 0x01, 0x0a, 0x0b, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x12, 0x3a, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74,
	0x2e, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x61, 0x74,
	0x65, 0x67, 0x6f, 0x72, 0x79, 0x52, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68,
	0x65, 0x61, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x22, 0x2a, 0x0a, 0x08, 0x43, 0x61, 0x74, 0x65, 0x67,
	0x6f, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f,
	0x41, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f,
	0x42, 0x10, 0x01, 0x2a, 0x41, 0x0a, 0x0c, 0x50, 0x6f, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x41, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x42, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x50, 0x4f, 0x53, 0x54, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x43, 0x10, 0x02, 0x32, 0x4f, 0x0a, 0x0b, 0x50, 0x6f, 0x73, 0x74, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x40, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74,
	0x12, 0x18, 0x2e, 0x6f, 0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x6f, 0x72, 0x67,
	0x2e, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x6d, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x6f,
	0x72, 0x67, 0x2e, 0x70, 0x6f, 0x73, 0x74, 0x42, 0x09, 0x50, 0x6f, 0x73, 0x74, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x11, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x70, 0x6f,
	0x73, 0x74, 0x3b, 0x70, 0x6f, 0x73, 0x74, 0xa2, 0x02, 0x03, 0x4f, 0x50, 0x58, 0xaa, 0x02, 0x08,
	0x4f, 0x72, 0x67, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0xca, 0x02, 0x08, 0x4f, 0x72, 0x67, 0x5c, 0x50,
	0x6f, 0x73, 0x74, 0xe2, 0x02, 0x14, 0x4f, 0x72, 0x67, 0x5c, 0x50, 0x6f, 0x73, 0x74, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x4f, 0x72, 0x67,
	0x3a, 0x3a, 0x50, 0x6f, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_post_post_proto_rawDescOnce sync.Once
	file_post_post_proto_rawDescData = file_post_post_proto_rawDesc
)

func file_post_post_proto_rawDescGZIP() []byte {
	file_post_post_proto_rawDescOnce.Do(func() {
		file_post_post_proto_rawDescData = protoimpl.X.CompressGZIP(file_post_post_proto_rawDescData)
	})
	return file_post_post_proto_rawDescData
}

var file_post_post_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_post_post_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_post_post_proto_goTypes = []interface{}{
	(PostDataType)(0),         // 0: org.post.PostDataType
	(PostContent_Category)(0), // 1: org.post.PostContent.Category
	(*GetPostRequest)(nil),    // 2: org.post.GetPostRequest
	(*GetPostResponse)(nil),   // 3: org.post.GetPostResponse
	(*Post)(nil),              // 4: org.post.Post
	(*PostData)(nil),          // 5: org.post.PostData
	(*PostContent)(nil),       // 6: org.post.PostContent
}
var file_post_post_proto_depIdxs = []int32{
	4, // 0: org.post.GetPostResponse.post:type_name -> org.post.Post
	5, // 1: org.post.Post.data:type_name -> org.post.PostData
	0, // 2: org.post.PostData.type:type_name -> org.post.PostDataType
	6, // 3: org.post.PostData.content:type_name -> org.post.PostContent
	1, // 4: org.post.PostContent.category:type_name -> org.post.PostContent.Category
	2, // 5: org.post.PostService.GetPost:input_type -> org.post.GetPostRequest
	3, // 6: org.post.PostService.GetPost:output_type -> org.post.GetPostResponse
	6, // [6:7] is the sub-list for method output_type
	5, // [5:6] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_post_post_proto_init() }
func file_post_post_proto_init() {
	if File_post_post_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_post_post_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_post_post_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_post_post_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_post_post_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_post_post_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_post_post_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_post_post_proto_goTypes,
		DependencyIndexes: file_post_post_proto_depIdxs,
		EnumInfos:         file_post_post_proto_enumTypes,
		MessageInfos:      file_post_post_proto_msgTypes,
	}.Build()
	File_post_post_proto = out.File
	file_post_post_proto_rawDesc = nil
	file_post_post_proto_goTypes = nil
	file_post_post_proto_depIdxs = nil
}
