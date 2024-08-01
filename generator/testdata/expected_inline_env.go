// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: inline_env.proto
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

// InlineEnvServiceConfig configuration required to initialize the service that use GRPC Federation.
type InlineEnvServiceConfig struct {
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler grpcfed.ErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// InlineEnvServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type InlineEnvServiceClientFactory interface {
}

// InlineEnvServiceClientConfig helper to create gRPC client.
// Hints for creating a gRPC Client.
type InlineEnvServiceClientConfig struct {
	// Service FQDN ( `<package-name>.<service-name>` ) of the service on Protocol Buffers.
	Service string
}

// InlineEnvServiceDependentClientSet has a gRPC client for all services on which the federation service depends.
// This is provided as an argument when implementing the custom resolver.
type InlineEnvServiceDependentClientSet struct {
}

// InlineEnvServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type InlineEnvServiceResolver interface {
}

// InlineEnvServiceCELPluginWasmConfig type alias for grpcfedcel.WasmConfig.
type InlineEnvServiceCELPluginWasmConfig = grpcfedcel.WasmConfig

// InlineEnvServiceCELPluginConfig hints for loading a WebAssembly based plugin.
type InlineEnvServiceCELPluginConfig struct {
}

// InlineEnvServiceEnv keeps the values read from environment variables.
type InlineEnvServiceEnv struct {
	Aaa string                      `default:"xxx"`
	Bbb []int64                     `envconfig:"yyy"`
	Ccc map[string]grpcfed.Duration `envconfig:"c" required:"true"`
	Ddd float64                     `ignored:"true"`
}

type keyInlineEnvServiceEnv struct{}

// GetInlineEnvServiceEnv gets environment variables.
func GetInlineEnvServiceEnv(ctx context.Context) *InlineEnvServiceEnv {
	value := ctx.Value(keyInlineEnvServiceEnv{})
	if value == nil {
		return nil
	}
	return value.(*InlineEnvServiceEnv)
}

func withInlineEnvServiceEnv(ctx context.Context, env *InlineEnvServiceEnv) context.Context {
	return context.WithValue(ctx, keyInlineEnvServiceEnv{}, env)
}

// InlineEnvServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type InlineEnvServiceUnimplementedResolver struct{}

// InlineEnvService represents Federation Service.
type InlineEnvService struct {
	*UnimplementedInlineEnvServiceServer
	cfg           InlineEnvServiceConfig
	logger        *slog.Logger
	errorHandler  grpcfed.ErrorHandler
	celCacheMap   *grpcfed.CELCacheMap
	tracer        trace.Tracer
	env           *InlineEnvServiceEnv
	celTypeHelper *grpcfed.CELTypeHelper
	celEnvOpts    []grpcfed.CELEnvOption
	celPlugins    []*grpcfedcel.CELPlugin
	client        *InlineEnvServiceDependentClientSet
}

// NewInlineEnvService creates InlineEnvService instance by InlineEnvServiceConfig.
func NewInlineEnvService(cfg InlineEnvServiceConfig) (*InlineEnvService, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	celTypeHelperFieldMap := grpcfed.CELTypeHelperFieldMap{
		"grpc.federation.private.Env": {
			"aaa": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Aaa"),
			"bbb": grpcfed.NewCELFieldType(grpcfed.NewCELListType(grpcfed.CELIntType), "Bbb"),
			"ccc": grpcfed.NewCELFieldType(grpcfed.NewCELMapType(grpcfed.CELStringType, grpcfed.CELDurationType), "Ccc"),
			"ddd": grpcfed.NewCELFieldType(grpcfed.CELDoubleType, "Ddd"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.NewCELVariable("grpc.federation.env", grpcfed.CELObjectType("grpc.federation.private.Env")))
	var env InlineEnvServiceEnv
	if err := grpcfed.LoadEnv("", &env); err != nil {
		return nil, err
	}
	return &InlineEnvService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.InlineEnvService"),
		env:           &env,
		client:        &InlineEnvServiceDependentClientSet{},
	}, nil
}
