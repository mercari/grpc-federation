// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
//
// source: create_post.proto
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
	"google.golang.org/protobuf/types/known/emptypb"

	post "example/post"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CreatePostVariable represents variable definitions in "org.federation.CreatePost".
type FederationService_Org_Federation_CreatePostVariable struct {
}

// Org_Federation_CreatePostArgument is argument for "org.federation.CreatePost" message.
type FederationService_Org_Federation_CreatePostArgument struct {
	Content string
	Title   string
	Type    PostType
	UserId  string
	FederationService_Org_Federation_CreatePostVariable
}

// Org_Federation_CreatePostResponseVariable represents variable definitions in "org.federation.CreatePostResponse".
type FederationService_Org_Federation_CreatePostResponseVariable struct {
	Cp  *CreatePost
	P   *post.Post
	Res *post.CreatePostResponse
}

// Org_Federation_CreatePostResponseArgument is argument for "org.federation.CreatePostResponse" message.
type FederationService_Org_Federation_CreatePostResponseArgument struct {
	Content string
	Title   string
	Type    PostType
	UserId  string
	FederationService_Org_Federation_CreatePostResponseVariable
}

// Org_Federation_UpdatePostResponseVariable represents variable definitions in "org.federation.UpdatePostResponse".
type FederationService_Org_Federation_UpdatePostResponseVariable struct {
}

// Org_Federation_UpdatePostResponseArgument is argument for "org.federation.UpdatePostResponse" message.
type FederationService_Org_Federation_UpdatePostResponseArgument struct {
	Id string
	FederationService_Org_Federation_UpdatePostResponseVariable
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
	CacheDir string
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

const (
	FederationService_DependentMethod_Org_Post_PostService_CreatePost = "/org.post.PostService/CreatePost"
	FederationService_DependentMethod_Org_Post_PostService_UpdatePost = "/org.post.PostService/UpdatePost"
)

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
		"grpc.federation.private.org.federation.CreatePostArgument": {
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
			"type":    grpcfed.NewCELFieldType(grpcfed.CELIntType, "Type"),
		},
		"grpc.federation.private.org.federation.CreatePostResponseArgument": {
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
			"type":    grpcfed.NewCELFieldType(grpcfed.CELIntType, "Type"),
		},
		"grpc.federation.private.org.federation.UpdatePostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.PostType", PostType_value, PostType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
	svc := &FederationService{
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

// CreatePost implements "org.federation.FederationService/CreatePost" method.
func (s *FederationService) CreatePost(ctx context.Context, req *CreatePostRequest) (res *CreatePostResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/CreatePost")
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
	res, err := s.resolve_Org_Federation_CreatePostResponse(ctx, &FederationService_Org_Federation_CreatePostResponseArgument{
		Title:   req.GetTitle(),
		Content: req.GetContent(),
		UserId:  req.GetUserId(),
		Type:    req.GetType(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// UpdatePost implements "org.federation.FederationService/UpdatePost" method.
func (s *FederationService) UpdatePost(ctx context.Context, req *UpdatePostRequest) (res *emptypb.Empty, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/UpdatePost")
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
	customRes, err := s.resolve_Org_Federation_UpdatePostResponse(ctx, &FederationService_Org_Federation_UpdatePostResponseArgument{
		Id: req.GetId(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	ret, err := s.cast_Org_Federation_UpdatePostResponse__to__Google_Protobuf_Empty(customRes)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return ret, nil
}

// resolve_Org_Federation_CreatePost resolve "org.federation.CreatePost" message.
func (s *FederationService) resolve_Org_Federation_CreatePost(ctx context.Context, req *FederationService_Org_Federation_CreatePostArgument) (*CreatePost, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CreatePost")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.CreatePost", slog.Any("message_args", s.logvalue_Org_Federation_CreatePostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.CreatePostArgument", req)}

	// create a message value to be returned.
	ret := &CreatePost{}

	// field binding section.
	// (grpc.federation.field).by = "$.title"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `$.title`,
		CacheIndex: 1,
		Setter: func(v string) error {
			ret.Title = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "$.content"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `$.content`,
		CacheIndex: 2,
		Setter: func(v string) error {
			ret.Content = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "$.user_id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `$.user_id`,
		CacheIndex: 3,
		Setter: func(v string) error {
			ret.UserId = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "PostType.from($.type)"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[PostType]{
		Value:      value,
		Expr:       `PostType.from($.type)`,
		CacheIndex: 4,
		Setter: func(v PostType) error {
			ret.Type = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "PostType.TYPE_A"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[int32]{
		Value:      value,
		Expr:       `PostType.TYPE_A`,
		CacheIndex: 5,
		Setter: func(v int32) error {
			ret.PostType = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.CreatePost", slog.Any("org.federation.CreatePost", s.logvalue_Org_Federation_CreatePost(ret)))
	return ret, nil
}

// resolve_Org_Federation_CreatePostResponse resolve "org.federation.CreatePostResponse" message.
func (s *FederationService) resolve_Org_Federation_CreatePostResponse(ctx context.Context, req *FederationService_Org_Federation_CreatePostResponseArgument) (*CreatePostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CreatePostResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.CreatePostResponse", slog.Any("message_args", s.logvalue_Org_Federation_CreatePostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			Cp  *CreatePost
			P   *post.Post
			Res *post.CreatePostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.CreatePostResponseArgument", req)}
	/*
		def {
		  name: "cp"
		  message {
		    name: "CreatePost"
		    args: [
		      { name: "title", by: "$.title" },
		      { name: "content", by: "$.content" },
		      { name: "user_id", by: "$.user_id" },
		      { name: "type", by: "$.type" }
		    ]
		  }
		}
	*/
	def_cp := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*CreatePost, *localValueType]{
			Name: `cp`,
			Type: grpcfed.CELObjectType("org.federation.CreatePost"),
			Setter: func(value *localValueType, v *CreatePost) error {
				value.vars.Cp = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_CreatePostArgument{}
				// { name: "title", by: "$.title" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.title`,
					CacheIndex: 6,
					Setter: func(v string) error {
						args.Title = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				// { name: "content", by: "$.content" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.content`,
					CacheIndex: 7,
					Setter: func(v string) error {
						args.Content = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				// { name: "user_id", by: "$.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.user_id`,
					CacheIndex: 8,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				// { name: "type", by: "$.type" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[PostType]{
					Value:      value,
					Expr:       `$.type`,
					CacheIndex: 9,
					Setter: func(v PostType) error {
						args.Type = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				ret, err := s.resolve_Org_Federation_CreatePost(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "res"
		  call {
		    method: "org.post.PostService/CreatePost"
		    request { field: "post", by: "cp" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.CreatePostResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("org.post.CreatePostResponse"),
			Setter: func(value *localValueType, v *post.CreatePostResponse) error {
				value.vars.Res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.CreatePostRequest{}
				// { field: "post", by: "cp" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*CreatePost]{
					Value:      value,
					Expr:       `cp`,
					CacheIndex: 10,
					Setter: func(v *CreatePost) error {
						postValue, err := s.cast_Org_Federation_CreatePost__to__Org_Post_CreatePost(v)
						if err != nil {
							return err
						}
						args.Post = postValue
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/CreatePost", slog.Any("org.post.CreatePostRequest", s.logvalue_Org_Post_CreatePostRequest(args)))
				ret, err := s.client.Org_Post_PostServiceClient.CreatePost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_CreatePost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "p"
		  by: "res.post"
		}
	*/
	def_p := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.Post, *localValueType]{
			Name: `p`,
			Type: grpcfed.CELObjectType("org.post.Post"),
			Setter: func(value *localValueType, v *post.Post) error {
				value.vars.P = v
				return nil
			},
			By:           `res.post`,
			ByCacheIndex: 11,
		})
	}

	if err := def_cp(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_res(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_p(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.FederationService_Org_Federation_CreatePostResponseVariable.Cp = value.vars.Cp
	req.FederationService_Org_Federation_CreatePostResponseVariable.P = value.vars.P
	req.FederationService_Org_Federation_CreatePostResponseVariable.Res = value.vars.Res

	// create a message value to be returned.
	ret := &CreatePostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "p"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*post.Post]{
		Value:      value,
		Expr:       `p`,
		CacheIndex: 12,
		Setter: func(v *post.Post) error {
			postValue, err := s.cast_Org_Post_Post__to__Org_Federation_Post(v)
			if err != nil {
				return err
			}
			ret.Post = postValue
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.CreatePostResponse", slog.Any("org.federation.CreatePostResponse", s.logvalue_Org_Federation_CreatePostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_UpdatePostResponse resolve "org.federation.UpdatePostResponse" message.
func (s *FederationService) resolve_Org_Federation_UpdatePostResponse(ctx context.Context, req *FederationService_Org_Federation_UpdatePostResponseArgument) (*UpdatePostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.UpdatePostResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.UpdatePostResponse", slog.Any("message_args", s.logvalue_Org_Federation_UpdatePostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			XDef0 *post.UpdatePostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.UpdatePostResponseArgument", req)}
	/*
		def {
		  name: "_def0"
		  call {
		    method: "org.post.PostService/UpdatePost"
		    request { field: "id", by: "$.id" }
		  }
		}
	*/
	def__def0 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.UpdatePostResponse, *localValueType]{
			Name: `_def0`,
			Type: grpcfed.CELObjectType("org.post.UpdatePostResponse"),
			Setter: func(value *localValueType, v *post.UpdatePostResponse) error {
				value.vars.XDef0 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.UpdatePostRequest{}
				// { field: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.id`,
					CacheIndex: 13,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/UpdatePost", slog.Any("org.post.UpdatePostRequest", s.logvalue_Org_Post_UpdatePostRequest(args)))
				ret, err := s.client.Org_Post_PostServiceClient.UpdatePost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_UpdatePost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		})
	}

	if err := def__def0(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// create a message value to be returned.
	ret := &UpdatePostResponse{}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.UpdatePostResponse", slog.Any("org.federation.UpdatePostResponse", s.logvalue_Org_Federation_UpdatePostResponse(ret)))
	return ret, nil
}

// cast_Org_Federation_CreatePost__to__Org_Post_CreatePost cast from "org.federation.CreatePost" to "org.post.CreatePost".
func (s *FederationService) cast_Org_Federation_CreatePost__to__Org_Post_CreatePost(from *CreatePost) (*post.CreatePost, error) {
	if from == nil {
		return nil, nil
	}

	titleValue := from.GetTitle()
	contentValue := from.GetContent()
	userIdValue := from.GetUserId()
	typeValue, err := s.cast_Org_Federation_PostType__to__Org_Post_PostType(from.GetType())
	if err != nil {
		return nil, err
	}
	postTypeValue := from.GetPostType()

	ret := &post.CreatePost{
		Title:    titleValue,
		Content:  contentValue,
		UserId:   userIdValue,
		Type:     typeValue,
		PostType: postTypeValue,
	}
	return ret, nil
}

// cast_Org_Federation_PostType__to__Org_Post_PostType cast from "org.federation.PostType" to "org.post.PostType".
func (s *FederationService) cast_Org_Federation_PostType__to__Org_Post_PostType(from PostType) (post.PostType, error) {
	var ret post.PostType
	switch from {
	case PostType_TYPE_UNKNOWN:
		ret = post.PostType_POST_TYPE_UNKNOWN
	case PostType_TYPE_A:
		ret = post.PostType_POST_TYPE_A
	case PostType_TYPE_B:
		ret = post.PostType_POST_TYPE_B
	default:
		ret = 0
	}
	return ret, nil
}

// cast_Org_Federation_UpdatePostResponse__to__Google_Protobuf_Empty cast from "org.federation.UpdatePostResponse" to "google.protobuf.Empty".
func (s *FederationService) cast_Org_Federation_UpdatePostResponse__to__Google_Protobuf_Empty(from *UpdatePostResponse) (*emptypb.Empty, error) {
	if from == nil {
		return nil, nil
	}

	ret := &emptypb.Empty{}
	return ret, nil
}

// cast_Org_Post_Post__to__Org_Federation_Post cast from "org.post.Post" to "org.federation.Post".
func (s *FederationService) cast_Org_Post_Post__to__Org_Federation_Post(from *post.Post) (*Post, error) {
	if from == nil {
		return nil, nil
	}

	idValue := from.GetId()
	titleValue := from.GetTitle()
	contentValue := from.GetContent()
	userIdValue := from.GetUserId()

	ret := &Post{
		Id:      idValue,
		Title:   titleValue,
		Content: contentValue,
		UserId:  userIdValue,
	}
	return ret, nil
}

// cast_int64__to__int32 cast from "int64" to "int32".
func (s *FederationService) cast_int64__to__int32(from int64) (int32, error) {
	ret, err := grpcfed.Int64ToInt32(from)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_CreatePost(v *CreatePost) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.String("user_id", v.GetUserId()),
		slog.String("type", s.logvalue_Org_Federation_PostType(v.GetType()).String()),
		slog.Int64("post_type", int64(v.GetPostType())),
	)
}

func (s *FederationService) logvalue_Org_Federation_CreatePostArgument(v *FederationService_Org_Federation_CreatePostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("title", v.Title),
		slog.String("content", v.Content),
		slog.String("user_id", v.UserId),
		slog.String("type", s.logvalue_Org_Federation_PostType(v.Type).String()),
	)
}

func (s *FederationService) logvalue_Org_Federation_CreatePostResponse(v *CreatePostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_CreatePostResponseArgument(v *FederationService_Org_Federation_CreatePostResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("title", v.Title),
		slog.String("content", v.Content),
		slog.String("user_id", v.UserId),
		slog.String("type", s.logvalue_Org_Federation_PostType(v.Type).String()),
	)
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.String("user_id", v.GetUserId()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostType(v PostType) slog.Value {
	switch v {
	case PostType_TYPE_UNKNOWN:
		return slog.StringValue("TYPE_UNKNOWN")
	case PostType_TYPE_A:
		return slog.StringValue("TYPE_A")
	case PostType_TYPE_B:
		return slog.StringValue("TYPE_B")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_UpdatePostResponse(v *UpdatePostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_UpdatePostResponseArgument(v *FederationService_Org_Federation_UpdatePostResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Post_CreatePost(v *post.CreatePost) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.String("user_id", v.GetUserId()),
		slog.String("type", s.logvalue_Org_Post_PostType(v.GetType()).String()),
		slog.Int64("post_type", int64(v.GetPostType())),
	)
}

func (s *FederationService) logvalue_Org_Post_CreatePostRequest(v *post.CreatePostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Post_CreatePost(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Post_GetPostRequest(v *post.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Org_Post_GetPostsRequest(v *post.GetPostsRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.GetIds()),
	)
}

func (s *FederationService) logvalue_Org_Post_PostType(v post.PostType) slog.Value {
	switch v {
	case post.PostType_POST_TYPE_UNKNOWN:
		return slog.StringValue("POST_TYPE_UNKNOWN")
	case post.PostType_POST_TYPE_A:
		return slog.StringValue("POST_TYPE_A")
	case post.PostType_POST_TYPE_B:
		return slog.StringValue("POST_TYPE_B")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Post_UpdatePostRequest(v *post.UpdatePostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}
