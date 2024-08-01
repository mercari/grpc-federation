// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: film/film.proto

package filmpb

import (
	date "google.golang.org/genproto/googleapis/type/date"
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

// GetFilmRequest.
type GetFilmRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The film id.
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetFilmRequest) Reset() {
	*x = GetFilmRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_film_film_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFilmRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFilmRequest) ProtoMessage() {}

func (x *GetFilmRequest) ProtoReflect() protoreflect.Message {
	mi := &file_film_film_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFilmRequest.ProtoReflect.Descriptor instead.
func (*GetFilmRequest) Descriptor() ([]byte, []int) {
	return file_film_film_proto_rawDescGZIP(), []int{0}
}

func (x *GetFilmRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

// GetFilmReply.
type GetFilmReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Film *Film `protobuf:"bytes,1,opt,name=film,proto3" json:"film,omitempty"`
}

func (x *GetFilmReply) Reset() {
	*x = GetFilmReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_film_film_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFilmReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFilmReply) ProtoMessage() {}

func (x *GetFilmReply) ProtoReflect() protoreflect.Message {
	mi := &file_film_film_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFilmReply.ProtoReflect.Descriptor instead.
func (*GetFilmReply) Descriptor() ([]byte, []int) {
	return file_film_film_proto_rawDescGZIP(), []int{1}
}

func (x *GetFilmReply) GetFilm() *Film {
	if x != nil {
		return x.Film
	}
	return nil
}

// ListFilmsRequest.
type ListFilmsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []int64 `protobuf:"varint,1,rep,packed,name=ids,proto3" json:"ids,omitempty"`
}

func (x *ListFilmsRequest) Reset() {
	*x = ListFilmsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_film_film_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFilmsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFilmsRequest) ProtoMessage() {}

func (x *ListFilmsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_film_film_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFilmsRequest.ProtoReflect.Descriptor instead.
func (*ListFilmsRequest) Descriptor() ([]byte, []int) {
	return file_film_film_proto_rawDescGZIP(), []int{2}
}

func (x *ListFilmsRequest) GetIds() []int64 {
	if x != nil {
		return x.Ids
	}
	return nil
}

// ListFilmsReply.
type ListFilmsReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Films []*Film `protobuf:"bytes,1,rep,name=films,proto3" json:"films,omitempty"`
}

func (x *ListFilmsReply) Reset() {
	*x = ListFilmsReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_film_film_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFilmsReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFilmsReply) ProtoMessage() {}

func (x *ListFilmsReply) ProtoReflect() protoreflect.Message {
	mi := &file_film_film_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFilmsReply.ProtoReflect.Descriptor instead.
func (*ListFilmsReply) Descriptor() ([]byte, []int) {
	return file_film_film_proto_rawDescGZIP(), []int{3}
}

func (x *ListFilmsReply) GetFilms() []*Film {
	if x != nil {
		return x.Films
	}
	return nil
}

// Film is a single film.
type Film struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// film id.
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// The title of this film.
	Title string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	// The episode number of this film.
	EpisodeId int32 `protobuf:"varint,3,opt,name=episode_id,json=episodeId,proto3" json:"episode_id,omitempty"`
	// The opening paragraphs at the beginning of this film.
	OpeningCrawl string `protobuf:"bytes,4,opt,name=opening_crawl,json=openingCrawl,proto3" json:"opening_crawl,omitempty"`
	// The name of the director of this film.
	Director string `protobuf:"bytes,5,opt,name=director,proto3" json:"director,omitempty"`
	// The name(s) of the producer(s) of this film. Comma separated.
	Producer string `protobuf:"bytes,6,opt,name=producer,proto3" json:"producer,omitempty"`
	// The ISO 8601 date format of film release at original creator country.
	ReleaseDate *date.Date `protobuf:"bytes,7,opt,name=release_date,json=releaseDate,proto3" json:"release_date,omitempty"`
	// the hypermedia URL of this resource.
	Url string `protobuf:"bytes,8,opt,name=url,proto3" json:"url,omitempty"`
	// the ISO 8601 date format of the time that this resource was created.
	Created string `protobuf:"bytes,9,opt,name=created,proto3" json:"created,omitempty"`
	// the ISO 8601 date format of the time that this resource was edited.
	Edited string `protobuf:"bytes,10,opt,name=edited,proto3" json:"edited,omitempty"`
	// species ids.
	SpeciesIds []int64 `protobuf:"varint,11,rep,packed,name=species_ids,json=speciesIds,proto3" json:"species_ids,omitempty"`
	// starship ids.
	StarshipIds []int64 `protobuf:"varint,12,rep,packed,name=starship_ids,json=starshipIds,proto3" json:"starship_ids,omitempty"`
	// vehicle ids.
	VehicleIds []int64 `protobuf:"varint,13,rep,packed,name=vehicle_ids,json=vehicleIds,proto3" json:"vehicle_ids,omitempty"`
	// character ids.
	CharacterIds []int64 `protobuf:"varint,14,rep,packed,name=character_ids,json=characterIds,proto3" json:"character_ids,omitempty"`
	// planet ids.
	PlanetIds []int64 `protobuf:"varint,15,rep,packed,name=planet_ids,json=planetIds,proto3" json:"planet_ids,omitempty"`
}

func (x *Film) Reset() {
	*x = Film{}
	if protoimpl.UnsafeEnabled {
		mi := &file_film_film_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Film) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Film) ProtoMessage() {}

func (x *Film) ProtoReflect() protoreflect.Message {
	mi := &file_film_film_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Film.ProtoReflect.Descriptor instead.
func (*Film) Descriptor() ([]byte, []int) {
	return file_film_film_proto_rawDescGZIP(), []int{4}
}

func (x *Film) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Film) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Film) GetEpisodeId() int32 {
	if x != nil {
		return x.EpisodeId
	}
	return 0
}

func (x *Film) GetOpeningCrawl() string {
	if x != nil {
		return x.OpeningCrawl
	}
	return ""
}

func (x *Film) GetDirector() string {
	if x != nil {
		return x.Director
	}
	return ""
}

func (x *Film) GetProducer() string {
	if x != nil {
		return x.Producer
	}
	return ""
}

func (x *Film) GetReleaseDate() *date.Date {
	if x != nil {
		return x.ReleaseDate
	}
	return nil
}

func (x *Film) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Film) GetCreated() string {
	if x != nil {
		return x.Created
	}
	return ""
}

func (x *Film) GetEdited() string {
	if x != nil {
		return x.Edited
	}
	return ""
}

func (x *Film) GetSpeciesIds() []int64 {
	if x != nil {
		return x.SpeciesIds
	}
	return nil
}

func (x *Film) GetStarshipIds() []int64 {
	if x != nil {
		return x.StarshipIds
	}
	return nil
}

func (x *Film) GetVehicleIds() []int64 {
	if x != nil {
		return x.VehicleIds
	}
	return nil
}

func (x *Film) GetCharacterIds() []int64 {
	if x != nil {
		return x.CharacterIds
	}
	return nil
}

func (x *Film) GetPlanetIds() []int64 {
	if x != nil {
		return x.PlanetIds
	}
	return nil
}

var File_film_film_proto protoreflect.FileDescriptor

var file_film_film_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x69, 0x6c, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0a, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66, 0x69, 0x6c, 0x6d, 0x1a, 0x16, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x20, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x6d,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22, 0x34, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x46, 0x69,
	0x6c, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x24, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66, 0x69,
	0x6c, 0x6d, 0x2e, 0x46, 0x69, 0x6c, 0x6d, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x6d, 0x22, 0x24, 0x0a,
	0x10, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x03,
	0x69, 0x64, 0x73, 0x22, 0x38, 0x0a, 0x0e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x73,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x26, 0x0a, 0x05, 0x66, 0x69, 0x6c, 0x6d, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66, 0x69, 0x6c,
	0x6d, 0x2e, 0x46, 0x69, 0x6c, 0x6d, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x6d, 0x73, 0x22, 0xcb, 0x03,
	0x0a, 0x04, 0x46, 0x69, 0x6c, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x65, 0x70, 0x69, 0x73, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x09, 0x65, 0x70, 0x69, 0x73, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x6f,
	0x70, 0x65, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x6f, 0x70, 0x65, 0x6e, 0x69, 0x6e, 0x67, 0x43, 0x72, 0x61, 0x77, 0x6c,
	0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x1a, 0x0a, 0x08,
	0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x12, 0x34, 0x0a, 0x0c, 0x72, 0x65, 0x6c, 0x65,
	0x61, 0x73, 0x65, 0x5f, 0x64, 0x61, 0x74, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x44, 0x61, 0x74,
	0x65, 0x52, 0x0b, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x44, 0x61, 0x74, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x64,
	0x69, 0x74, 0x65, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x64, 0x69, 0x74,
	0x65, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x70, 0x65, 0x63, 0x69, 0x65, 0x73, 0x5f, 0x69, 0x64,
	0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0a, 0x73, 0x70, 0x65, 0x63, 0x69, 0x65, 0x73,
	0x49, 0x64, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x74, 0x61, 0x72, 0x73, 0x68, 0x69, 0x70, 0x5f,
	0x69, 0x64, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0b, 0x73, 0x74, 0x61, 0x72, 0x73,
	0x68, 0x69, 0x70, 0x49, 0x64, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c,
	0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0a, 0x76, 0x65, 0x68,
	0x69, 0x63, 0x6c, 0x65, 0x49, 0x64, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x68, 0x61, 0x72, 0x61,
	0x63, 0x74, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0c,
	0x63, 0x68, 0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x49, 0x64, 0x73, 0x12, 0x1d, 0x0a, 0x0a,
	0x70, 0x6c, 0x61, 0x6e, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x03,
	0x52, 0x09, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x74, 0x49, 0x64, 0x73, 0x32, 0x99, 0x01, 0x0a, 0x0b,
	0x46, 0x69, 0x6c, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x07, 0x47,
	0x65, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x12, 0x1a, 0x2e, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66,
	0x69, 0x6c, 0x6d, 0x2e, 0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x18, 0x2e, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66, 0x69, 0x6c, 0x6d, 0x2e,
	0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x47,
	0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x73, 0x12, 0x1c, 0x2e, 0x73, 0x77,
	0x61, 0x70, 0x69, 0x2e, 0x66, 0x69, 0x6c, 0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c,
	0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x73, 0x77, 0x61, 0x70,
	0x69, 0x2e, 0x66, 0x69, 0x6c, 0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x6d, 0x73,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x9f, 0x01, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x2e,
	0x73, 0x77, 0x61, 0x70, 0x69, 0x2e, 0x66, 0x69, 0x6c, 0x6d, 0x42, 0x09, 0x46, 0x69, 0x6c, 0x6d,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x65, 0x72, 0x63, 0x61, 0x72, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x2d, 0x66, 0x65, 0x64, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x64, 0x65, 0x6d, 0x6f,
	0x2f, 0x73, 0x77, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x6c, 0x6d, 0x3b, 0x66, 0x69, 0x6c, 0x6d,
	0x70, 0x62, 0xa2, 0x02, 0x03, 0x53, 0x46, 0x58, 0xaa, 0x02, 0x0a, 0x53, 0x77, 0x61, 0x70, 0x69,
	0x2e, 0x46, 0x69, 0x6c, 0x6d, 0xca, 0x02, 0x0a, 0x53, 0x77, 0x61, 0x70, 0x69, 0x5c, 0x46, 0x69,
	0x6c, 0x6d, 0xe2, 0x02, 0x16, 0x53, 0x77, 0x61, 0x70, 0x69, 0x5c, 0x46, 0x69, 0x6c, 0x6d, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0b, 0x53, 0x77,
	0x61, 0x70, 0x69, 0x3a, 0x3a, 0x46, 0x69, 0x6c, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_film_film_proto_rawDescOnce sync.Once
	file_film_film_proto_rawDescData = file_film_film_proto_rawDesc
)

func file_film_film_proto_rawDescGZIP() []byte {
	file_film_film_proto_rawDescOnce.Do(func() {
		file_film_film_proto_rawDescData = protoimpl.X.CompressGZIP(file_film_film_proto_rawDescData)
	})
	return file_film_film_proto_rawDescData
}

var file_film_film_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_film_film_proto_goTypes = []interface{}{
	(*GetFilmRequest)(nil),   // 0: swapi.film.GetFilmRequest
	(*GetFilmReply)(nil),     // 1: swapi.film.GetFilmReply
	(*ListFilmsRequest)(nil), // 2: swapi.film.ListFilmsRequest
	(*ListFilmsReply)(nil),   // 3: swapi.film.ListFilmsReply
	(*Film)(nil),             // 4: swapi.film.Film
	(*date.Date)(nil),        // 5: google.type.Date
}
var file_film_film_proto_depIdxs = []int32{
	4, // 0: swapi.film.GetFilmReply.film:type_name -> swapi.film.Film
	4, // 1: swapi.film.ListFilmsReply.films:type_name -> swapi.film.Film
	5, // 2: swapi.film.Film.release_date:type_name -> google.type.Date
	0, // 3: swapi.film.FilmService.GetFilm:input_type -> swapi.film.GetFilmRequest
	2, // 4: swapi.film.FilmService.ListFilms:input_type -> swapi.film.ListFilmsRequest
	1, // 5: swapi.film.FilmService.GetFilm:output_type -> swapi.film.GetFilmReply
	3, // 6: swapi.film.FilmService.ListFilms:output_type -> swapi.film.ListFilmsReply
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_film_film_proto_init() }
func file_film_film_proto_init() {
	if File_film_film_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_film_film_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetFilmRequest); i {
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
		file_film_film_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetFilmReply); i {
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
		file_film_film_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFilmsRequest); i {
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
		file_film_film_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFilmsReply); i {
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
		file_film_film_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Film); i {
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
			RawDescriptor: file_film_film_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_film_film_proto_goTypes,
		DependencyIndexes: file_film_film_proto_depIdxs,
		MessageInfos:      file_film_film_proto_msgTypes,
	}.Build()
	File_film_film_proto = out.File
	file_film_film_proto_rawDesc = nil
	file_film_film_proto_goTypes = nil
	file_film_film_proto_depIdxs = nil
}