// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: federation/federation.proto

package federation

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	FederationService_GetPost_FullMethodName  = "/org.federation.FederationService/GetPost"
	FederationService_GetPost2_FullMethodName = "/org.federation.FederationService/GetPost2"
)

// FederationServiceClient is the client API for FederationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FederationServiceClient interface {
	GetPost(ctx context.Context, in *GetPostRequest, opts ...grpc.CallOption) (*GetPostResponse, error)
	GetPost2(ctx context.Context, in *GetPost2Request, opts ...grpc.CallOption) (*GetPost2Response, error)
}

type federationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFederationServiceClient(cc grpc.ClientConnInterface) FederationServiceClient {
	return &federationServiceClient{cc}
}

func (c *federationServiceClient) GetPost(ctx context.Context, in *GetPostRequest, opts ...grpc.CallOption) (*GetPostResponse, error) {
	out := new(GetPostResponse)
	err := c.cc.Invoke(ctx, FederationService_GetPost_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *federationServiceClient) GetPost2(ctx context.Context, in *GetPost2Request, opts ...grpc.CallOption) (*GetPost2Response, error) {
	out := new(GetPost2Response)
	err := c.cc.Invoke(ctx, FederationService_GetPost2_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FederationServiceServer is the server API for FederationService service.
// All implementations must embed UnimplementedFederationServiceServer
// for forward compatibility
type FederationServiceServer interface {
	GetPost(context.Context, *GetPostRequest) (*GetPostResponse, error)
	GetPost2(context.Context, *GetPost2Request) (*GetPost2Response, error)
	mustEmbedUnimplementedFederationServiceServer()
}

// UnimplementedFederationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFederationServiceServer struct {
}

func (UnimplementedFederationServiceServer) GetPost(context.Context, *GetPostRequest) (*GetPostResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPost not implemented")
}
func (UnimplementedFederationServiceServer) GetPost2(context.Context, *GetPost2Request) (*GetPost2Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPost2 not implemented")
}
func (UnimplementedFederationServiceServer) mustEmbedUnimplementedFederationServiceServer() {}

// UnsafeFederationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FederationServiceServer will
// result in compilation errors.
type UnsafeFederationServiceServer interface {
	mustEmbedUnimplementedFederationServiceServer()
}

func RegisterFederationServiceServer(s grpc.ServiceRegistrar, srv FederationServiceServer) {
	s.RegisterService(&FederationService_ServiceDesc, srv)
}

func _FederationService_GetPost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FederationServiceServer).GetPost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FederationService_GetPost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FederationServiceServer).GetPost(ctx, req.(*GetPostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FederationService_GetPost2_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPost2Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FederationServiceServer).GetPost2(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FederationService_GetPost2_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FederationServiceServer).GetPost2(ctx, req.(*GetPost2Request))
	}
	return interceptor(ctx, in, info, handler)
}

// FederationService_ServiceDesc is the grpc.ServiceDesc for FederationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FederationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "org.federation.FederationService",
	HandlerType: (*FederationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPost",
			Handler:    _FederationService_GetPost_Handler,
		},
		{
			MethodName: "GetPost2",
			Handler:    _FederationService_GetPost2_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "federation/federation.proto",
}
