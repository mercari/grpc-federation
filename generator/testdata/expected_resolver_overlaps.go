// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: resolver_overlaps.proto
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

	post "example/post"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostResponse1Argument is argument for "org.federation.GetPostResponse1" message.
type FederationService_Org_Federation_GetPostResponse1Argument struct {
	Id   string
	Post *Post
}

// Org_Federation_GetPostResponse2Argument is argument for "org.federation.GetPostResponse2" message.
type FederationService_Org_Federation_GetPostResponse2Argument struct {
	Id   string
	Post *Post
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
	Id  string
	Res *post.GetPostResponse
}

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// Client provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
	// If this interface is not provided, an error is returned during initialization.
	Client FederationServiceClientFactory // required
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler grpcfed.ErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// FederationServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type FederationServiceClientFactory interface {
	// Org_Post_PostServiceClient create a gRPC Client to be used to call methods in org.post.PostService.
	Org_Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
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
	Org_Post_PostServiceClient post.PostServiceClient
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
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

const (
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
)

// FederationService represents Federation Service.
type FederationService struct {
	*UnimplementedFederationServiceServer
	cfg           FederationServiceConfig
	logger        *slog.Logger
	errorHandler  grpcfed.ErrorHandler
	celCacheMap   *grpcfed.CELCacheMap
	tracer        trace.Tracer
	celTypeHelper *grpcfed.CELTypeHelper
	celEnvOpts    []grpcfed.CELEnvOption
	celPlugins    []*grpcfedcel.CELPlugin
	client        *FederationServiceDependentClientSet
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if cfg.Client == nil {
		return nil, grpcfed.ErrClientConfig
	}
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
	})
	if err != nil {
		return nil, err
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
		"grpc.federation.private.GetPostResponse1Argument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.GetPostResponse2Argument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostContent.Category", post.PostContent_Category_value, post.PostContent_Category_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostDataType", post.PostDataType_value, post.PostDataType_name)...)
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		client: &FederationServiceDependentClientSet{
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
		},
	}, nil
}

// GetPost1 implements "org.federation.FederationService/GetPost1" method.
func (s *FederationService) GetPost1(ctx context.Context, req *GetPostRequest) (res *GetPostResponse1, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost1")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, grpcfed.StackTrace())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostResponse1(ctx, &FederationService_Org_Federation_GetPostResponse1Argument{
		Id: req.GetId(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// GetPost2 implements "org.federation.FederationService/GetPost2" method.
func (s *FederationService) GetPost2(ctx context.Context, req *GetPostRequest) (res *GetPostResponse2, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost2")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, grpcfed.StackTrace())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostResponse2(ctx, &FederationService_Org_Federation_GetPostResponse2Argument{
		Id: req.GetId(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostResponse1 resolve "org.federation.GetPostResponse1" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse1(ctx context.Context, req *FederationService_Org_Federation_GetPostResponse1Argument) (*GetPostResponse1, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse1")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPostResponse1", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponse1Argument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, false, "grpc.federation.private.GetPostResponse1Argument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "post"
	     message {
	       name: "Post"
	       args { name: "id", by: "$.id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*Post, *localValueType]{
		Name: `post`,
		Type: grpcfed.CELObjectType("org.federation.Post"),
		Setter: func(value *localValueType, v *Post) error {
			value.vars.post = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &FederationService_Org_Federation_PostArgument{}
			// { name: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.id`,
				UseContextLibrary: false,
				CacheIndex:        1,
				Setter: func(v string) error {
					args.Id = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			return s.resolve_Org_Federation_Post(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.post

	// create a message value to be returned.
	ret := &GetPostResponse1{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:             value,
		Expr:              `post`,
		UseContextLibrary: false,
		CacheIndex:        2,
		Setter: func(v *Post) error {
			ret.Post = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetPostResponse1", slog.Any("org.federation.GetPostResponse1", s.logvalue_Org_Federation_GetPostResponse1(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetPostResponse2 resolve "org.federation.GetPostResponse2" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse2(ctx context.Context, req *FederationService_Org_Federation_GetPostResponse2Argument) (*GetPostResponse2, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse2")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPostResponse2", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponse2Argument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, false, "grpc.federation.private.GetPostResponse2Argument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "post"
	     message {
	       name: "Post"
	       args { name: "id", by: "$.id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*Post, *localValueType]{
		Name: `post`,
		Type: grpcfed.CELObjectType("org.federation.Post"),
		Setter: func(value *localValueType, v *Post) error {
			value.vars.post = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &FederationService_Org_Federation_PostArgument{}
			// { name: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.id`,
				UseContextLibrary: false,
				CacheIndex:        3,
				Setter: func(v string) error {
					args.Id = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			return s.resolve_Org_Federation_Post(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.post

	// create a message value to be returned.
	ret := &GetPostResponse2{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:             value,
		Expr:              `post`,
		UseContextLibrary: false,
		CacheIndex:        4,
		Setter: func(v *Post) error {
			ret.Post = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetPostResponse2", slog.Any("org.federation.GetPostResponse2", s.logvalue_Org_Federation_GetPostResponse2(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *FederationService_Org_Federation_PostArgument) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			res *post.GetPostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, false, "grpc.federation.private.PostArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "res"
	     call {
	       method: "org.post.PostService/GetPost"
	       request { field: "id", by: "$.id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
		Name: `res`,
		Type: grpcfed.CELObjectType("org.post.GetPostResponse"),
		Setter: func(value *localValueType, v *post.GetPostResponse) error {
			value.vars.res = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &post.GetPostRequest{}
			// { field: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.id`,
				UseContextLibrary: false,
				CacheIndex:        5,
				Setter: func(v string) error {
					args.Id = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/GetPost", slog.Any("org.post.GetPostRequest", s.logvalue_Org_Post_GetPostRequest(args)))
			return s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, grpcfed.NewErrorWithLogAttrs(err, grpcfed.LogAttrs(ctx))
		}
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = value.vars.res

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	// (grpc.federation.field).by = "res.post.id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              `res.post.id`,
		UseContextLibrary: false,
		CacheIndex:        6,
		Setter: func(v string) error {
			ret.Id = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse1(v *GetPostResponse1) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse1Argument(v *FederationService_Org_Federation_GetPostResponse1Argument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse2(v *GetPostResponse2) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse2Argument(v *FederationService_Org_Federation_GetPostResponse2Argument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *FederationService_Org_Federation_PostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Post_GetPostRequest(v *post.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.Any("a", s.logvalue_Org_Post_PostConditionA(v.GetA())),
		slog.Any("b", s.logvalue_Org_Post_PostConditionB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Org_Post_PostConditionA(v *post.PostConditionA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("prop", v.GetProp()),
	)
}

func (s *FederationService) logvalue_Org_Post_PostConditionB(v *post.PostConditionB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}
