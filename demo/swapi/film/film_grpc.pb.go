// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: film/film.proto

package filmpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	FilmService_GetFilm_FullMethodName   = "/swapi.film.FilmService/GetFilm"
	FilmService_ListFilms_FullMethodName = "/swapi.film.FilmService/ListFilms"
)

// FilmServiceClient is the client API for FilmService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FilmServiceClient interface {
	GetFilm(ctx context.Context, in *GetFilmRequest, opts ...grpc.CallOption) (*GetFilmReply, error)
	ListFilms(ctx context.Context, in *ListFilmsRequest, opts ...grpc.CallOption) (*ListFilmsReply, error)
}

type filmServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFilmServiceClient(cc grpc.ClientConnInterface) FilmServiceClient {
	return &filmServiceClient{cc}
}

func (c *filmServiceClient) GetFilm(ctx context.Context, in *GetFilmRequest, opts ...grpc.CallOption) (*GetFilmReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetFilmReply)
	err := c.cc.Invoke(ctx, FilmService_GetFilm_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *filmServiceClient) ListFilms(ctx context.Context, in *ListFilmsRequest, opts ...grpc.CallOption) (*ListFilmsReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListFilmsReply)
	err := c.cc.Invoke(ctx, FilmService_ListFilms_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FilmServiceServer is the server API for FilmService service.
// All implementations must embed UnimplementedFilmServiceServer
// for forward compatibility
type FilmServiceServer interface {
	GetFilm(context.Context, *GetFilmRequest) (*GetFilmReply, error)
	ListFilms(context.Context, *ListFilmsRequest) (*ListFilmsReply, error)
	mustEmbedUnimplementedFilmServiceServer()
}

// UnimplementedFilmServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFilmServiceServer struct {
}

func (UnimplementedFilmServiceServer) GetFilm(context.Context, *GetFilmRequest) (*GetFilmReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFilm not implemented")
}
func (UnimplementedFilmServiceServer) ListFilms(context.Context, *ListFilmsRequest) (*ListFilmsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFilms not implemented")
}
func (UnimplementedFilmServiceServer) mustEmbedUnimplementedFilmServiceServer() {}

// UnsafeFilmServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FilmServiceServer will
// result in compilation errors.
type UnsafeFilmServiceServer interface {
	mustEmbedUnimplementedFilmServiceServer()
}

func RegisterFilmServiceServer(s grpc.ServiceRegistrar, srv FilmServiceServer) {
	s.RegisterService(&FilmService_ServiceDesc, srv)
}

func _FilmService_GetFilm_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFilmRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FilmServiceServer).GetFilm(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FilmService_GetFilm_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FilmServiceServer).GetFilm(ctx, req.(*GetFilmRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FilmService_ListFilms_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFilmsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FilmServiceServer).ListFilms(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FilmService_ListFilms_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FilmServiceServer).ListFilms(ctx, req.(*ListFilmsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FilmService_ServiceDesc is the grpc.ServiceDesc for FilmService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FilmService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "swapi.film.FilmService",
	HandlerType: (*FilmServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetFilm",
			Handler:    _FilmService_GetFilm_Handler,
		},
		{
			MethodName: "ListFilms",
			Handler:    _FilmService_ListFilms_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "film/film.proto",
}
