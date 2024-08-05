// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
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

	post "example/post"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CreatePostArgument is argument for "org.federation.CreatePost" message.
type Org_Federation_CreatePostArgument struct {
	Content string
	Title   string
	Type    PostType
	UserId  string
}

// Org_Federation_CreatePostResponseArgument is argument for "org.federation.CreatePostResponse" message.
type Org_Federation_CreatePostResponseArgument struct {
	Content string
	Cp      *CreatePost
	P       *post.Post
	Res     *post.CreatePostResponse
	Title   string
	Type    PostType
	UserId  string
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
	FederationService_DependentMethod_Org_Post_PostService_CreatePost = "/org.post.PostService/CreatePost"
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
		"grpc.federation.private.CreatePostArgument": {
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
			"type":    grpcfed.NewCELFieldType(grpcfed.CELIntType, "Type"),
		},
		"grpc.federation.private.CreatePostResponseArgument": {
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
			"type":    grpcfed.NewCELFieldType(grpcfed.CELIntType, "Type"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.PostType", PostType_value, PostType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
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
	res, err := s.resolve_Org_Federation_CreatePostResponse(ctx, &Org_Federation_CreatePostResponseArgument{
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

// resolve_Org_Federation_CreatePost resolve "org.federation.CreatePost" message.
func (s *FederationService) resolve_Org_Federation_CreatePost(ctx context.Context, req *Org_Federation_CreatePostArgument) (*CreatePost, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CreatePost")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.CreatePost", slog.Any("message_args", s.logvalue_Org_Federation_CreatePostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, false, "grpc.federation.private.CreatePostArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// create a message value to be returned.
	ret := &CreatePost{}

	// field binding section.
	// (grpc.federation.field).by = "$.title"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              `$.title`,
		UseContextLibrary: false,
		CacheIndex:        1,
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
		Value:             value,
		Expr:              `$.content`,
		UseContextLibrary: false,
		CacheIndex:        2,
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
		Value:             value,
		Expr:              `$.user_id`,
		UseContextLibrary: false,
		CacheIndex:        3,
		Setter: func(v string) error {
			ret.UserId = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "$.type"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[PostType]{
		Value:             value,
		Expr:              `$.type`,
		UseContextLibrary: false,
		CacheIndex:        4,
		Setter: func(v PostType) error {
			ret.Type = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "org.federation.PostType.TYPE_A"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[PostType]{
		Value:             value,
		Expr:              `org.federation.PostType.TYPE_A`,
		UseContextLibrary: false,
		CacheIndex:        5,
		Setter: func(v PostType) error {
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
func (s *FederationService) resolve_Org_Federation_CreatePostResponse(ctx context.Context, req *Org_Federation_CreatePostResponseArgument) (*CreatePostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CreatePostResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.CreatePostResponse", slog.Any("message_args", s.logvalue_Org_Federation_CreatePostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			cp  *CreatePost
			p   *post.Post
			res *post.CreatePostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, false, "grpc.federation.private.CreatePostResponseArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
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
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*CreatePost, *localValueType]{
		Name: `cp`,
		Type: grpcfed.CELObjectType("org.federation.CreatePost"),
		Setter: func(value *localValueType, v *CreatePost) error {
			value.vars.cp = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Org_Federation_CreatePostArgument{}
			// { name: "title", by: "$.title" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.title`,
				UseContextLibrary: false,
				CacheIndex:        6,
				Setter: func(v string) error {
					args.Title = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { name: "content", by: "$.content" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.content`,
				UseContextLibrary: false,
				CacheIndex:        7,
				Setter: func(v string) error {
					args.Content = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { name: "user_id", by: "$.user_id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.user_id`,
				UseContextLibrary: false,
				CacheIndex:        8,
				Setter: func(v string) error {
					args.UserId = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { name: "type", by: "$.type" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[PostType]{
				Value:             value,
				Expr:              `$.type`,
				UseContextLibrary: false,
				CacheIndex:        9,
				Setter: func(v PostType) error {
					args.Type = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			return s.resolve_Org_Federation_CreatePost(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "res"
	     call {
	       method: "org.post.PostService/CreatePost"
	       request { field: "post", by: "cp" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.CreatePostResponse, *localValueType]{
		Name: `res`,
		Type: grpcfed.CELObjectType("org.post.CreatePostResponse"),
		Setter: func(value *localValueType, v *post.CreatePostResponse) error {
			value.vars.res = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &post.CreatePostRequest{}
			// { field: "post", by: "cp" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*CreatePost]{
				Value:             value,
				Expr:              `cp`,
				UseContextLibrary: false,
				CacheIndex:        10,
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
			return s.client.Org_Post_PostServiceClient.CreatePost(ctx, args)
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_CreatePost, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, grpcfed.NewErrorWithLogAttrs(err, grpcfed.LogAttrs(ctx))
		}
	}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "p"
	     by: "res.post"
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.Post, *localValueType]{
		Name: `p`,
		Type: grpcfed.CELObjectType("org.post.Post"),
		Setter: func(value *localValueType, v *post.Post) error {
			value.vars.p = v
			return nil
		},
		By:                  `res.post`,
		ByUseContextLibrary: false,
		ByCacheIndex:        11,
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Cp = value.vars.cp
	req.P = value.vars.p
	req.Res = value.vars.res

	// create a message value to be returned.
	ret := &CreatePostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "p"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*post.Post]{
		Value:             value,
		Expr:              `p`,
		UseContextLibrary: false,
		CacheIndex:        12,
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
	postTypeValue, err := s.cast_Org_Federation_PostType__to__int32(from.GetPostType())
	if err != nil {
		return nil, err
	}

	return &post.CreatePost{
		Title:    titleValue,
		Content:  contentValue,
		UserId:   userIdValue,
		Type:     typeValue,
		PostType: postTypeValue,
	}, nil
}

// cast_Org_Federation_PostType__to__Org_Post_PostType cast from "org.federation.PostType" to "org.post.PostType".
func (s *FederationService) cast_Org_Federation_PostType__to__Org_Post_PostType(from PostType) (post.PostType, error) {
	switch from {
	case PostType_TYPE_UNKNOWN:
		return post.PostType_POST_TYPE_UNKNOWN, nil
	case PostType_TYPE_A:
		return post.PostType_POST_TYPE_A, nil
	case PostType_TYPE_B:
		return post.PostType_POST_TYPE_B, nil
	default:
		return 0, nil
	}
}

// cast_Org_Federation_PostType__to__int32 cast from "org.federation.PostType" to "int32".
func (s *FederationService) cast_Org_Federation_PostType__to__int32(from PostType) (int32, error) {
	return int32(from), nil
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

	return &Post{
		Id:      idValue,
		Title:   titleValue,
		Content: contentValue,
		UserId:  userIdValue,
	}, nil
}

// cast_int64__to__Org_Federation_PostType cast from "int64" to "org.federation.PostType".
func (s *FederationService) cast_int64__to__Org_Federation_PostType(from int64) (PostType, error) {
	return PostType(from), nil
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
		slog.String("post_type", s.logvalue_Org_Federation_PostType(v.GetPostType()).String()),
	)
}

func (s *FederationService) logvalue_Org_Federation_CreatePostArgument(v *Org_Federation_CreatePostArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_CreatePostResponseArgument(v *Org_Federation_CreatePostResponseArgument) slog.Value {
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
