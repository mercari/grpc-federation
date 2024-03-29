// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
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

	post "example/post"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type Org_Federation_GetPostResponseArgument[T any] struct {
	Id     string
	Post   *Post
	Client T
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument[T any] struct {
	Id     string
	Post   *post.Post
	Res    *post.GetPostResponse
	Client T
}

// Org_Federation_PostContentArgument is argument for "org.federation.PostContent" message.
type Org_Federation_PostContentArgument[T any] struct {
	Client T
}

// Org_Federation_PostDataArgument is argument for "org.federation.PostData" message.
type Org_Federation_PostDataArgument[T any] struct {
	Client T
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

// FederationServiceClientConfig information set in `dependencies` of the `grpc.federation.service` option.
// Hints for creating a gRPC Client.
type FederationServiceClientConfig struct {
	// Service returns the name of the service on Protocol Buffers.
	Service string
	// Name is the value set for `name` in `dependencies` of the `grpc.federation.service` option.
	// It must be unique among the services on which the Federation Service depends.
	Name string
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
	cfg          FederationServiceConfig
	logger       *slog.Logger
	errorHandler grpcfed.ErrorHandler
	env          *grpcfed.CELEnv
	tracer       trace.Tracer
	client       *FederationServiceDependentClientSet
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if cfg.Client == nil {
		return nil, grpcfed.ErrClientConfig
	}
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
		Name:    "post_service",
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
	celHelper := grpcfed.NewCELTypeHelper(map[string]map[string]*grpcfed.CELFieldType{
		"grpc.federation.private.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
	})
	envOpts := grpcfed.NewDefaultEnvOptions(celHelper)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.federation.PostContent.Category", PostContent_Category_value, PostContent_Category_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.federation.PostType", PostType_value, PostType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.post.PostContent.Category", post.PostContent_Category_value, post.PostContent_Category_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.post.PostDataType", post.PostDataType_value, post.PostDataType_name)...)
	env, err := grpcfed.NewCELEnv(envOpts...)
	if err != nil {
		return nil, err
	}
	return &FederationService{
		cfg:          cfg,
		logger:       logger,
		errorHandler: errorHandler,
		env:          env,
		tracer:       otel.Tracer("org.federation.FederationService"),
		client: &FederationServiceDependentClientSet{
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
		},
	}, nil
}

// GetPost implements "org.federation.FederationService/GetPost" method.
func (s *FederationService) GetPost(ctx context.Context, req *GetPostRequest) (res *GetPostResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostResponse(ctx, &Org_Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]{
		Client: s.client,
		Id:     req.Id,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, s.logger, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *Org_Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.GetPostResponseArgument", req)}

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
		Name:   "post",
		Type:   grpcfed.CELObjectType("org.federation.Post"),
		Setter: func(value *localValueType, v *Post) { value.vars.post = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Org_Federation_PostArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			// { name: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, value, "$.id", func(v string) {
				args.Id = v
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
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, value, "post", func(v *Post) { ret.Post = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument[*FederationServiceDependentClientSet]) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *post.Post
			res  *post.GetPostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.PostArgument", req)}

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
		Name:   "res",
		Type:   grpcfed.CELObjectType("org.post.GetPostResponse"),
		Setter: func(value *localValueType, v *post.GetPostResponse) { value.vars.res = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &post.GetPostRequest{}
			// { field: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, value, "$.id", func(v string) {
				args.Id = v
			}); err != nil {
				return nil, err
			}
			return s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "post"
	     autobind: true
	     by: "res.post"
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.Post, *localValueType]{
		Name:   "post",
		Type:   grpcfed.CELObjectType("org.post.Post"),
		Setter: func(value *localValueType, v *post.Post) { value.vars.post = v },
		By:     "res.post",
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.post
	req.Res = value.vars.res

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.post.GetId()                                                            // { name: "post", autobind: true }
	ret.Data = s.cast_Org_Post_PostData__to__Org_Federation_PostData(value.vars.post.GetData()) // { name: "post", autobind: true }

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category cast from "org.post.PostContent.Category" to "org.federation.PostContent.Category".
func (s *FederationService) cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category(from post.PostContent_Category) PostContent_Category {
	switch from {
	case post.PostContent_CATEGORY_A:
		return PostContent_CATEGORY_A
	case post.PostContent_CATEGORY_B:
		return PostContent_CATEGORY_B
	default:
		return 0
	}
}

// cast_Org_Post_PostContent__to__Org_Federation_PostContent cast from "org.post.PostContent" to "org.federation.PostContent".
func (s *FederationService) cast_Org_Post_PostContent__to__Org_Federation_PostContent(from *post.PostContent) *PostContent {
	if from == nil {
		return nil
	}

	return &PostContent{
		Category: s.cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category(from.GetCategory()),
		Head:     from.GetHead(),
		Body:     from.GetBody(),
		DupBody:  from.GetBody(),
	}
}

// cast_Org_Post_PostDataType__to__Org_Federation_PostType cast from "org.post.PostDataType" to "org.federation.PostType".
func (s *FederationService) cast_Org_Post_PostDataType__to__Org_Federation_PostType(from post.PostDataType) PostType {
	switch from {
	case post.PostDataType_POST_TYPE_A:
		return PostType_POST_TYPE_FOO
	case post.PostDataType_POST_TYPE_B:
		return PostType_POST_TYPE_BAR
	case post.PostDataType_POST_TYPE_C:
		return PostType_POST_TYPE_BAR
	default:
		return PostType_POST_TYPE_UNKNOWN
	}
}

// cast_Org_Post_PostData__to__Org_Federation_PostData cast from "org.post.PostData" to "org.federation.PostData".
func (s *FederationService) cast_Org_Post_PostData__to__Org_Federation_PostData(from *post.PostData) *PostData {
	if from == nil {
		return nil
	}

	return &PostData{
		Type:    s.cast_Org_Post_PostDataType__to__Org_Federation_PostType(from.GetType()),
		Title:   from.GetTitle(),
		Content: s.cast_Org_Post_PostContent__to__Org_Federation_PostContent(from.GetContent()),
	}
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponseArgument(v *Org_Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]) slog.Value {
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
		slog.Any("data", s.logvalue_Org_Federation_PostData(v.GetData())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *Org_Federation_PostArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostContent(v *PostContent) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("category", s.logvalue_Org_Federation_PostContent_Category(v.GetCategory()).String()),
		slog.String("head", v.GetHead()),
		slog.String("body", v.GetBody()),
		slog.String("dup_body", v.GetDupBody()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostContent_Category(v PostContent_Category) slog.Value {
	switch v {
	case PostContent_CATEGORY_A:
		return slog.StringValue("CATEGORY_A")
	case PostContent_CATEGORY_B:
		return slog.StringValue("CATEGORY_B")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_PostData(v *PostData) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("type", s.logvalue_Org_Federation_PostType(v.GetType()).String()),
		slog.String("title", v.GetTitle()),
		slog.Any("content", s.logvalue_Org_Federation_PostContent(v.GetContent())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostType(v PostType) slog.Value {
	switch v {
	case PostType_POST_TYPE_UNKNOWN:
		return slog.StringValue("POST_TYPE_UNKNOWN")
	case PostType_POST_TYPE_FOO:
		return slog.StringValue("POST_TYPE_FOO")
	case PostType_POST_TYPE_BAR:
		return slog.StringValue("POST_TYPE_BAR")
	}
	return slog.StringValue("")
}
