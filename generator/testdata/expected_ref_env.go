// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
//
// source: ref_env.proto
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

// Org_Federation_ConstantArgument is argument for "org.federation.Constant" message.
type RefEnvService_Org_Federation_ConstantArgument struct {
}

// RefEnvServiceConfig configuration required to initialize the service that use GRPC Federation.
type RefEnvServiceConfig struct {
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler grpcfed.ErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// RefEnvServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type RefEnvServiceClientFactory interface {
}

// RefEnvServiceClientConfig helper to create gRPC client.
// Hints for creating a gRPC Client.
type RefEnvServiceClientConfig struct {
	// Service FQDN ( `<package-name>.<service-name>` ) of the service on Protocol Buffers.
	Service string
}

// RefEnvServiceDependentClientSet has a gRPC client for all services on which the federation service depends.
// This is provided as an argument when implementing the custom resolver.
type RefEnvServiceDependentClientSet struct {
}

// RefEnvServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type RefEnvServiceResolver interface {
}

// RefEnvServiceCELPluginWasmConfig type alias for grpcfedcel.WasmConfig.
type RefEnvServiceCELPluginWasmConfig = grpcfedcel.WasmConfig

// RefEnvServiceCELPluginConfig hints for loading a WebAssembly based plugin.
type RefEnvServiceCELPluginConfig struct {
}

// RefEnvServiceEnv keeps the values read from environment variables.
type RefEnvServiceEnv struct {
	Aaa string                      `envconfig:"AAA" default:"xxx"`
	Bbb []int64                     `envconfig:"yyy"`
	Ccc map[string]grpcfed.Duration `envconfig:"c" required:"true"`
	Ddd float64                     `envconfig:"DDD" ignored:"true"`
}

type keyRefEnvServiceEnv struct{}

// GetRefEnvServiceEnv gets environment variables.
func GetRefEnvServiceEnv(ctx context.Context) *RefEnvServiceEnv {
	value := ctx.Value(keyRefEnvServiceEnv{})
	if value == nil {
		return nil
	}
	return value.(*RefEnvServiceEnv)
}

func withRefEnvServiceEnv(ctx context.Context, env *RefEnvServiceEnv) context.Context {
	return context.WithValue(ctx, keyRefEnvServiceEnv{}, env)
}

// RefEnvServiceVariable keeps the initial values.
type RefEnvServiceVariable struct {
	Constant *Constant
}

type keyRefEnvServiceVariable struct{}

// GetRefEnvServiceVariable gets initial variables.
func GetRefEnvServiceVariable(ctx context.Context) *RefEnvServiceVariable {
	value := ctx.Value(keyRefEnvServiceVariable{})
	if value == nil {
		return nil
	}
	return value.(*RefEnvServiceVariable)
}

func withRefEnvServiceVariable(ctx context.Context, svcVar *RefEnvServiceVariable) context.Context {
	return context.WithValue(ctx, keyRefEnvServiceVariable{}, svcVar)
}

// RefEnvServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type RefEnvServiceUnimplementedResolver struct{}

// RefEnvService represents Federation Service.
type RefEnvService struct {
	UnimplementedRefEnvServiceServer
	cfg                RefEnvServiceConfig
	logger             *slog.Logger
	errorHandler       grpcfed.ErrorHandler
	celCacheMap        *grpcfed.CELCacheMap
	tracer             trace.Tracer
	env                *RefEnvServiceEnv
	svcVar             *RefEnvServiceVariable
	celTypeHelper      *grpcfed.CELTypeHelper
	celEnvOpts         []grpcfed.CELEnvOption
	celPluginInstances []*grpcfedcel.CELPluginInstance
	client             *RefEnvServiceDependentClientSet
}

// NewRefEnvService creates RefEnvService instance by RefEnvServiceConfig.
func NewRefEnvService(cfg RefEnvServiceConfig) (*RefEnvService, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	celTypeHelperFieldMap := grpcfed.CELTypeHelperFieldMap{
		"grpc.federation.private.org.federation.ConstantArgument": {},
		"grpc.federation.private.Env": {
			"aaa": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Aaa"),
			"bbb": grpcfed.NewCELFieldType(grpcfed.NewCELListType(grpcfed.CELIntType), "Bbb"),
			"ccc": grpcfed.NewCELFieldType(grpcfed.NewCELMapType(grpcfed.CELStringType, grpcfed.NewCELObjectType("google.protobuf.Duration")), "Ccc"),
			"ddd": grpcfed.NewCELFieldType(grpcfed.CELDoubleType, "Ddd"),
		},
		"grpc.federation.private.ServiceVariable": {
			"constant": grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("org.federation.Constant"), "Constant"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.NewCELVariable("grpc.federation.env", grpcfed.CELObjectType("grpc.federation.private.Env")))
	celEnvOpts = append(celEnvOpts, grpcfed.NewCELVariable("grpc.federation.var", grpcfed.CELObjectType("grpc.federation.private.ServiceVariable")))
	var env RefEnvServiceEnv
	if err := grpcfed.LoadEnv("", &env); err != nil {
		return nil, err
	}
	svc := &RefEnvService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.RefEnvService"),
		env:           &env,
		svcVar:        new(RefEnvServiceVariable),
		client:        &RefEnvServiceDependentClientSet{},
	}
	if err := svc.initServiceVariables(); err != nil {
		return nil, err
	}
	return svc, nil
}

// CleanupRefEnvService cleanup all resources to prevent goroutine leaks.
func CleanupRefEnvService(ctx context.Context, svc *RefEnvService) {
	svc.cleanup(ctx)
}

func (s *RefEnvService) cleanup(ctx context.Context) {
	for _, instance := range s.celPluginInstances {
		instance.Close(ctx)
	}
}
func (s *RefEnvService) initServiceVariables() error {
	ctx := grpcfed.WithCELCacheMap(grpcfed.WithLogger(context.Background(), s.logger), s.celCacheMap)
	type localValueType struct {
		*grpcfed.LocalValue
		vars *RefEnvServiceVariable
	}
	value := &localValueType{
		LocalValue: grpcfed.NewServiceVariableLocalValue(s.celEnvOpts),
		vars:       s.svcVar,
	}
	value.AddEnv(s.env)
	value.AddServiceVariable(s.svcVar)

	/*
		def {
		  name: "constant"
		  message {
		    name: "Constant"
		  }
		}
	*/
	def_constant := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*Constant, *localValueType]{
			Name: `constant`,
			Type: grpcfed.CELObjectType("org.federation.Constant"),
			Setter: func(value *localValueType, v *Constant) error {
				value.vars.Constant = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &RefEnvService_Org_Federation_ConstantArgument{}
				ret, err := s.resolve_Org_Federation_Constant(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}
	if err := def_constant(ctx); err != nil {
		return err
	}

	return nil
}

// resolve_Org_Federation_Constant resolve "org.federation.Constant" message.
func (s *RefEnvService) resolve_Org_Federation_Constant(ctx context.Context, req *RefEnvService_Org_Federation_ConstantArgument) (*Constant, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Constant")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.Constant", slog.Any("message_args", s.logvalue_Org_Federation_ConstantArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.ConstantArgument", req)}
	value.AddEnv(s.env)
	value.AddServiceVariable(s.svcVar)

	// create a message value to be returned.
	ret := &Constant{}

	// field binding section.
	// (grpc.federation.field).by = "grpc.federation.env.aaa + 'xxx'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `grpc.federation.env.aaa + 'xxx'`,
		CacheIndex: 1,
		Setter: func(v string) error {
			ret.X = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Constant", slog.Any("org.federation.Constant", s.logvalue_Org_Federation_Constant(ret)))
	return ret, nil
}

func (s *RefEnvService) logvalue_Org_Federation_Constant(v *Constant) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("x", v.GetX()),
	)
}

func (s *RefEnvService) logvalue_Org_Federation_ConstantArgument(v *RefEnvService_Org_Federation_ConstantArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}
