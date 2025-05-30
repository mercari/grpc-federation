// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
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

// Org_Federation_GetResponseVariable represents variable definitions in "org.federation.GetResponse".
type FederationService_Org_Federation_GetResponseVariable struct {
	Body string
	File string
	Foo  string
}

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type FederationService_Org_Federation_GetResponseArgument struct {
	Path string
	Url  string
	FederationService_Org_Federation_GetResponseVariable
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
	Net      FederationServiceCELPluginWasmConfig
	CacheDir string
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
	cfg                FederationServiceConfig
	logger             *slog.Logger
	errorHandler       grpcfed.ErrorHandler
	celCacheMap        *grpcfed.CELCacheMap
	tracer             trace.Tracer
	celTypeHelper      *grpcfed.CELTypeHelper
	celEnvOpts         []grpcfed.CELEnvOption
	celPluginInstances []*grpcfedcel.CELPluginInstance
	client             *FederationServiceDependentClientSet
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
		"grpc.federation.private.org.federation.GetResponseArgument": {
			"url":  grpcfed.NewCELFieldType(grpcfed.CELStringType, "Url"),
			"path": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Path"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	var celPluginInstances []*grpcfedcel.CELPluginInstance
	{
		plugin, err := grpcfedcel.NewCELPlugin(context.Background(), grpcfedcel.CELPluginConfig{
			Name:     "net",
			Wasm:     cfg.CELPlugin.Net,
			CacheDir: cfg.CELPlugin.CacheDir,
			Functions: []*grpcfedcel.CELFunction{
				{
					Name: "example.net.httpGet",
					ID:   "example_net_httpGet_string_string",
					Args: []*grpcfed.CELTypeDeclare{
						grpcfed.CELStringType,
					},
					Return:   grpcfed.CELStringType,
					IsMethod: false,
				},
				{
					Name:     "example.net.getFooEnv",
					ID:       "example_net_getFooEnv_string",
					Args:     []*grpcfed.CELTypeDeclare{},
					Return:   grpcfed.CELStringType,
					IsMethod: false,
				},
				{
					Name: "example.net.getFileContent",
					ID:   "example_net_getFileContent_string_string",
					Args: []*grpcfed.CELTypeDeclare{
						grpcfed.CELStringType,
					},
					Return:   grpcfed.CELStringType,
					IsMethod: false,
				},
			},
			Capability: &grpcfedcel.CELPluginCapability{
				Env: &grpcfedcel.CELPluginEnvCapability{
					All:   true,
					Names: []string{},
				},
				FileSystem: &grpcfedcel.CELPluginFileSystemCapability{
					MountPath: "/",
				},
				Network: &grpcfedcel.CELPluginNetworkCapability{},
			},
		})
		if err != nil {
			return nil, err
		}
		ctx := context.Background()
		instance, err := plugin.CreateInstance(ctx, celTypeHelper.CELRegistry())
		if err != nil {
			return nil, err
		}
		if err := instance.ValidatePlugin(ctx); err != nil {
			return nil, err
		}
		celPluginInstances = append(celPluginInstances, instance)
		celEnvOpts = append(celEnvOpts, grpcfed.CELLib(instance))
	}
	svc := &FederationService{
		cfg:                cfg,
		logger:             logger,
		errorHandler:       errorHandler,
		celEnvOpts:         celEnvOpts,
		celTypeHelper:      celTypeHelper,
		celCacheMap:        grpcfed.NewCELCacheMap(),
		tracer:             otel.Tracer("org.federation.FederationService"),
		celPluginInstances: celPluginInstances,
		client:             &FederationServiceDependentClientSet{},
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

	defer func() {
		// cleanup plugin instance memory.
		for _, instance := range s.celPluginInstances {
			instance.GC()
		}
	}()
	res, err := s.resolve_Org_Federation_GetResponse(ctx, &FederationService_Org_Federation_GetResponseArgument{
		Url:  req.GetUrl(),
		Path: req.GetPath(),
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
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			Body string
			File string
			Foo  string
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.GetResponseArgument", req)}
	/*
		def {
		  name: "body"
		  by: "example.net.httpGet($.url)"
		}
	*/
	def_body := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[string, *localValueType]{
			Name: `body`,
			Type: grpcfed.CELStringType,
			Setter: func(value *localValueType, v string) error {
				value.vars.Body = v
				return nil
			},
			By:           `example.net.httpGet($.url)`,
			ByCacheIndex: 1,
		})
	}

	/*
		def {
		  name: "foo"
		  by: "example.net.getFooEnv()"
		}
	*/
	def_foo := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[string, *localValueType]{
			Name: `foo`,
			Type: grpcfed.CELStringType,
			Setter: func(value *localValueType, v string) error {
				value.vars.Foo = v
				return nil
			},
			By:           `example.net.getFooEnv()`,
			ByCacheIndex: 2,
		})
	}

	/*
		def {
		  name: "file"
		  by: "example.net.getFileContent($.path)"
		}
	*/
	def_file := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[string, *localValueType]{
			Name: `file`,
			Type: grpcfed.CELStringType,
			Setter: func(value *localValueType, v string) error {
				value.vars.File = v
				return nil
			},
			By:           `example.net.getFileContent($.path)`,
			ByCacheIndex: 3,
		})
	}

	// A tree view of message dependencies is shown below.
	/*
	   body ─┐
	   file ─┤
	    foo ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_body(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_file(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_foo(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.FederationService_Org_Federation_GetResponseVariable.Body = value.vars.Body
	req.FederationService_Org_Federation_GetResponseVariable.File = value.vars.File
	req.FederationService_Org_Federation_GetResponseVariable.Foo = value.vars.Foo

	// create a message value to be returned.
	ret := &GetResponse{}

	// field binding section.
	// (grpc.federation.field).by = "body"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `body`,
		CacheIndex: 4,
		Setter: func(v string) error {
			ret.Body = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "foo"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `foo`,
		CacheIndex: 5,
		Setter: func(v string) error {
			ret.Foo = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "file"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `file`,
		CacheIndex: 6,
		Setter: func(v string) error {
			ret.File = v
			return nil
		},
	}); err != nil {
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
		slog.String("body", v.GetBody()),
		slog.String("foo", v.GetFoo()),
		slog.String("file", v.GetFile()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *FederationService_Org_Federation_GetResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("url", v.Url),
		slog.String("path", v.Path),
	)
}
