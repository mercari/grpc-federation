// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: federation/federation.proto
package federation

import (
	"context"
	"io"
	"log/slog"
	"reflect"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type FederationService_Org_Federation_GetResponseArgument struct {
	Id int64
}

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// Resolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
	// If this interface is not provided, an error is returned during initialization.
	Resolver FederationServiceResolver // required
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler grpcfed.ErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// FederationServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type FederationServiceClientFactory interface {
}

// FederationServiceClientConfig helper to create gRPC client.
// Hints for creating a gRPC Client.
type FederationServiceClientConfig struct {
	// Service FQDN ( `<package-name>.<service-name>` ) of the service on Protocol Buffers.
	Service string
}

// FederationServiceDependentClientSet has a gRPC client for all services on which the federation service depends.
// This is provided as an argument when implementing the custom resolver.
type FederationServiceDependentClientSet struct {
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
	// Resolve_Org_Federation_GetResponse implements resolver for "org.federation.GetResponse".
	Resolve_Org_Federation_GetResponse(context.Context, *FederationService_Org_Federation_GetResponseArgument) (*GetResponse, error)
}

// FederationServiceCELPluginWasmConfig type alias for grpcfedcel.WasmConfig.
type FederationServiceCELPluginWasmConfig = grpcfedcel.WasmConfig

// FederationServiceCELPluginConfig hints for loading a WebAssembly based plugin.
type FederationServiceCELPluginConfig struct {
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

// Resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_GetResponse(context.Context, *FederationService_Org_Federation_GetResponseArgument) (ret *GetResponse, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_GetResponse not implemented")
	return
}

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
	cfg                FederationServiceConfig
	logger             *slog.Logger
	errorHandler       grpcfed.ErrorHandler
	celCacheMap        *grpcfed.CELCacheMap
	tracer             trace.Tracer
	resolver           FederationServiceResolver
	celTypeHelper      *grpcfed.CELTypeHelper
	celEnvOpts         []grpcfed.CELEnvOption
	celPluginInstances []*grpcfedcel.CELPluginInstance
	client             *FederationServiceDependentClientSet
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if cfg.Resolver == nil {
		return nil, grpcfed.ErrResolverConfig
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	celTypeHelperFieldMap := grpcfed.CELTypeHelperFieldMap{
		"grpc.federation.private.GetResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELIntType, "Id"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	svc := &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		resolver:      cfg.Resolver,
		client:        &FederationServiceDependentClientSet{},
	}
	return svc, nil
}

// CleanupFederationService cleanup all resources to prevent goroutine leaks.
func CleanupFederationService(ctx context.Context, svc *FederationService) {
	svc.cleanup(ctx)
}

func (s *FederationService) cleanup(ctx context.Context) {
	for _, instance := range s.celPluginInstances {
		instance.Close(ctx)
	}
}

// Get implements "org.federation.FederationService/Get" method.
func (s *FederationService) Get(ctx context.Context, req *GetRequest) (res *GetResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/Get")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, grpcfed.StackTrace())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := grpcfed.WithTimeout[GetResponse](ctx, "org.federation.FederationService/Get", 10000000000 /* 10s */, func(ctx context.Context) (*GetResponse, error) {
		return s.resolve_Org_Federation_GetResponse(ctx, &FederationService_Org_Federation_GetResponseArgument{
			Id: req.GetId(),
		})
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *FederationService_Org_Federation_GetResponseArgument) (*GetResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
	ret, err := s.resolver.Resolve_Org_Federation_GetResponse(ctx, req)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetResponse(v *GetResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Int64("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *FederationService_Org_Federation_GetResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Int64("id", v.Id),
	)
}
