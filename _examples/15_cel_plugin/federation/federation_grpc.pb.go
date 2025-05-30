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
	FederationService_IsMatch_FullMethodName = "/org.federation.FederationService/IsMatch"
	FederationService_Example_FullMethodName = "/org.federation.FederationService/Example"
)

// FederationServiceClient is the client API for FederationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FederationServiceClient interface {
	IsMatch(ctx context.Context, in *IsMatchRequest, opts ...grpc.CallOption) (*IsMatchResponse, error)
	Example(ctx context.Context, in *ExampleRequest, opts ...grpc.CallOption) (*ExampleResponse, error)
}

type federationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFederationServiceClient(cc grpc.ClientConnInterface) FederationServiceClient {
	return &federationServiceClient{cc}
}

func (c *federationServiceClient) IsMatch(ctx context.Context, in *IsMatchRequest, opts ...grpc.CallOption) (*IsMatchResponse, error) {
	out := new(IsMatchResponse)
	err := c.cc.Invoke(ctx, FederationService_IsMatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *federationServiceClient) Example(ctx context.Context, in *ExampleRequest, opts ...grpc.CallOption) (*ExampleResponse, error) {
	out := new(ExampleResponse)
	err := c.cc.Invoke(ctx, FederationService_Example_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FederationServiceServer is the server API for FederationService service.
// All implementations must embed UnimplementedFederationServiceServer
// for forward compatibility
type FederationServiceServer interface {
	IsMatch(context.Context, *IsMatchRequest) (*IsMatchResponse, error)
	Example(context.Context, *ExampleRequest) (*ExampleResponse, error)
	mustEmbedUnimplementedFederationServiceServer()
}

// UnimplementedFederationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFederationServiceServer struct {
}

func (UnimplementedFederationServiceServer) IsMatch(context.Context, *IsMatchRequest) (*IsMatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsMatch not implemented")
}
func (UnimplementedFederationServiceServer) Example(context.Context, *ExampleRequest) (*ExampleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Example not implemented")
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

func _FederationService_IsMatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsMatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FederationServiceServer).IsMatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FederationService_IsMatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FederationServiceServer).IsMatch(ctx, req.(*IsMatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FederationService_Example_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExampleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FederationServiceServer).Example(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FederationService_Example_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FederationServiceServer).Example(ctx, req.(*ExampleRequest))
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
			MethodName: "IsMatch",
			Handler:    _FederationService_IsMatch_Handler,
		},
		{
			MethodName: "Example",
			Handler:    _FederationService_Example_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "federation/federation.proto",
}
