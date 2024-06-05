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
	"runtime/debug"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	pluginpb "example/plugin"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_IsMatchResponseArgument is argument for "org.federation.IsMatchResponse" message.
type Org_Federation_IsMatchResponseArgument struct {
	Expr    string
	Matched bool
	Re      *pluginpb.Regexp
	Target  string
}

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// CELPlugin If you use the plugin feature to extend the CEL API,
	// you must write a plugin and output WebAssembly.
	// In this field, configure to load wasm with the path to the WebAssembly file and the sha256 value.
	CELPlugin *FederationServiceCELPluginConfig
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
}

// FederationServiceCELPluginWasmConfig type alias for grpcfedcel.WasmConfig.
type FederationServiceCELPluginWasmConfig = grpcfedcel.WasmConfig

// FederationServiceCELPluginConfig hints for loading a WebAssembly based plugin.
type FederationServiceCELPluginConfig struct {
	Regexp FederationServiceCELPluginWasmConfig
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

// FederationService represents Federation Service.
type FederationService struct {
	*UnimplementedFederationServiceServer
	cfg           FederationServiceConfig
	logger        *slog.Logger
	errorHandler  grpcfed.ErrorHandler
	celCacheMap   *grpcfed.CELCacheMap
	tracer        trace.Tracer
	celTypeHelper *grpcfed.CELTypeHelper
	envOpts       []grpcfed.CELEnvOption
	celPlugins    []*grpcfedcel.CELPlugin
	client        *FederationServiceDependentClientSet
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if cfg.CELPlugin == nil {
		return nil, grpcfed.ErrCELPluginConfig
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
		"grpc.federation.private.IsMatchResponseArgument": {
			"expr":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Expr"),
			"target": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Target"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper(celTypeHelperFieldMap)
	var envOpts []grpcfed.CELEnvOption
	envOpts = append(envOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	var celPlugins []*grpcfedcel.CELPlugin
	{
		plugin, err := grpcfedcel.NewCELPlugin(context.Background(), grpcfedcel.CELPluginConfig{
			Name: "regexp",
			Wasm: cfg.CELPlugin.Regexp,
			Functions: []*grpcfedcel.CELFunction{
				{
					Name: "example.regexp.compile",
					ID:   "example_regexp_compile_string_example_regexp_Regexp",
					Args: []*grpcfed.CELTypeDeclare{
						grpcfed.CELStringType,
					},
					Return:   grpcfed.NewCELObjectType("example.regexp.Regexp"),
					IsMethod: false,
				},
				{
					Name: "matchString",
					ID:   "example_regexp_Regexp_matchString_example_regexp_Regexp_string_bool",
					Args: []*grpcfed.CELTypeDeclare{
						grpcfed.NewCELObjectType("example.regexp.Regexp"),
						grpcfed.CELStringType,
					},
					Return:   grpcfed.CELBoolType,
					IsMethod: true,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		if err := func() error {
			ctx := context.Background()
			instance := plugin.CreateInstance(ctx, celTypeHelper.CELRegistry())
			defer instance.Close(ctx)
			return instance.ValidatePlugin(ctx)
		}(); err != nil {
			return nil, err
		}
		celPlugins = append(celPlugins, plugin)
	}
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		envOpts:       envOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		celPlugins:    celPlugins,
		client:        &FederationServiceDependentClientSet{},
	}, nil
}

// IsMatch implements "org.federation.FederationService/IsMatch" method.
func (s *FederationService) IsMatch(ctx context.Context, req *IsMatchRequest) (res *IsMatchResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/IsMatch")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_IsMatchResponse(ctx, &Org_Federation_IsMatchResponseArgument{
		Expr:   req.Expr,
		Target: req.Target,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_IsMatchResponse resolve "org.federation.IsMatchResponse" message.
func (s *FederationService) resolve_Org_Federation_IsMatchResponse(ctx context.Context, req *Org_Federation_IsMatchResponseArgument) (*IsMatchResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.IsMatchResponse")
	defer span.End()

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.IsMatchResponse", slog.Any("message_args", s.logvalue_Org_Federation_IsMatchResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			matched bool
			re      *pluginpb.Regexp
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.envOpts, s.celPlugins, "grpc.federation.private.IsMatchResponseArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "re"
	     by: "example.regexp.compile($.expr)"
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*pluginpb.Regexp, *localValueType]{
		Name: `re`,
		Type: grpcfed.CELObjectType("example.regexp.Regexp"),
		Setter: func(value *localValueType, v *pluginpb.Regexp) error {
			value.vars.re = v
			return nil
		},
		By:                  `example.regexp.compile($.expr)`,
		ByUseContextLibrary: true,
		ByCacheIndex:        1,
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "matched"
	     by: "re.matchString($.target)"
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[bool, *localValueType]{
		Name: `matched`,
		Type: grpcfed.CELBoolType,
		Setter: func(value *localValueType, v bool) error {
			value.vars.matched = v
			return nil
		},
		By:                  `re.matchString($.target)`,
		ByUseContextLibrary: true,
		ByCacheIndex:        2,
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Matched = value.vars.matched
	req.Re = value.vars.re

	// create a message value to be returned.
	ret := &IsMatchResponse{}

	// field binding section.
	// (grpc.federation.field).by = "matched"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[bool]{
		Value:             value,
		Expr:              `matched`,
		UseContextLibrary: false,
		CacheIndex:        3,
		Setter: func(v bool) error {
			ret.Result = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.IsMatchResponse", slog.Any("org.federation.IsMatchResponse", s.logvalue_Org_Federation_IsMatchResponse(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_IsMatchResponse(v *IsMatchResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Bool("result", v.GetResult()),
	)
}

func (s *FederationService) logvalue_Org_Federation_IsMatchResponseArgument(v *Org_Federation_IsMatchResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("expr", v.Expr),
		slog.String("target", v.Target),
	)
}
