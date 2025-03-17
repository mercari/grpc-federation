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

	post "example/post"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type FederationService_Org_Federation_GetPostResponseArgument struct {
	Id   string
	Post *Post
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
	Id    string
	Post  *post.Post
	Posts []*post.Post
	Res   *post.GetPostResponse
	User  *User
	Users []*User
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type FederationService_Org_Federation_UserArgument struct {
	UserId string
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
	// Post_PostServiceClient create a gRPC Client to be used to call methods in post.PostService.
	Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
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
	Post_PostServiceClient post.PostServiceClient
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
	FederationService_DependentMethod_Post_PostService_GetPost = "/post.PostService/GetPost"
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
	Post_PostServiceClient, err := cfg.Client.Post_PostServiceClient(FederationServiceClientConfig{
		Service: "post.PostService",
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
		"grpc.federation.private.org.federation.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.org.federation.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.org.federation.UserArgument": {
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
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
		client: &FederationServiceDependentClientSet{
			Post_PostServiceClient: Post_PostServiceClient,
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

// GetPost implements "org.federation.FederationService/GetPost" method.
func (s *FederationService) GetPost(ctx context.Context, req *GetPostRequest) (res *GetPostResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost")
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
	res, err := s.resolve_Org_Federation_GetPostResponse(ctx, &FederationService_Org_Federation_GetPostResponseArgument{
		Id: req.GetId(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *FederationService_Org_Federation_GetPostResponseArgument) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			Post *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.GetPostResponseArgument", req)}
	/*
		def {
		  name: "post"
		  message {
		    name: "Post"
		    args { name: "id", by: "$.id" }
		  }
		}
	*/
	def_post := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*Post, *localValueType]{
			Name: `post`,
			Type: grpcfed.CELObjectType("org.federation.Post"),
			Setter: func(value *localValueType, v *Post) error {
				value.vars.Post = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_PostArgument{}
				// { name: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.id`,
					CacheIndex: 1,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				ret, err := s.resolve_Org_Federation_Post(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	if err := def_post(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.Post

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:      value,
		Expr:       `post`,
		CacheIndex: 2,
		Setter: func(v *Post) error {
			ret.Post = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
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
			Post  *post.Post
			Posts []*post.Post
			Res   *post.GetPostResponse
			User  *User
			Users []*User
			XDef5 bool
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.PostArgument", req)}
	/*
		def {
		  name: "res"
		  if: "$.id != ''"
		  call {
		    method: "post.PostService/GetPost"
		    request { field: "id", by: "$.id" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
			If:           `$.id != ''`,
			IfCacheIndex: 3,
			Name:         `res`,
			Type:         grpcfed.CELObjectType("post.GetPostResponse"),
			Setter: func(value *localValueType, v *post.GetPostResponse) error {
				value.vars.Res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.GetPostRequest{}
				// { field: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.id`,
					CacheIndex: 4,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call post.PostService/GetPost", slog.Any("post.GetPostRequest", s.logvalue_Post_GetPostRequest(args)))
				ret, err := s.client.Post_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "post"
		  if: "res != null"
		  by: "res.post"
		}
	*/
	def_post := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.Post, *localValueType]{
			If:           `res != null`,
			IfCacheIndex: 5,
			Name:         `post`,
			Type:         grpcfed.CELObjectType("post.Post"),
			Setter: func(value *localValueType, v *post.Post) error {
				value.vars.Post = v
				return nil
			},
			By:           `res.post`,
			ByCacheIndex: 6,
		})
	}

	/*
		def {
		  name: "user"
		  if: "post != null"
		  message {
		    name: "User"
		    args { name: "user_id", by: "post.user_id" }
		  }
		}
	*/
	def_user := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			If:           `post != null`,
			IfCacheIndex: 7,
			Name:         `user`,
			Type:         grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.User = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_UserArgument{}
				// { name: "user_id", by: "post.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `post.user_id`,
					CacheIndex: 8,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				ret, err := s.resolve_Org_Federation_User(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "posts"
		  by: "[post]"
		}
	*/
	def_posts := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[[]*post.Post, *localValueType]{
			Name: `posts`,
			Type: grpcfed.CELListType(grpcfed.CELObjectType("post.Post")),
			Setter: func(value *localValueType, v []*post.Post) error {
				value.vars.Posts = v
				return nil
			},
			By:           `[post]`,
			ByCacheIndex: 9,
		})
	}

	/*
		def {
		  name: "users"
		  if: "user != null"
		  map {
		    iterator {
		      name: "iter"
		      src: "posts"
		    }
		    message {
		      name: "User"
		      args { name: "user_id", by: "iter.user_id" }
		    }
		  }
		}
	*/
	def_users := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]*User, *post.Post, *localValueType]{
			If:           `user != null`,
			IfCacheIndex: 10,
			Name:         `users`,
			Type:         grpcfed.CELListType(grpcfed.CELObjectType("org.federation.User")),
			Setter: func(value *localValueType, v []*User) error {
				value.vars.Users = v
				return nil
			},
			IteratorName:   `iter`,
			IteratorType:   grpcfed.CELObjectType("post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.Posts },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				args := &FederationService_Org_Federation_UserArgument{}
				// { name: "user_id", by: "iter.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `iter.user_id`,
					CacheIndex: 11,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		})
	}

	/*
		def {
		  name: "_def5"
		  if: "users.size() > 0"
		  validation {
		    error {
		      code: INVALID_ARGUMENT
		      if: "users[0].id == ''"
		    }
		  }
		}
	*/
	def__def5 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[bool, *localValueType]{
			If:           `users.size() > 0`,
			IfCacheIndex: 12,
			Name:         `_def5`,
			Type:         grpcfed.CELBoolType,
			Setter: func(value *localValueType, v bool) error {
				value.vars.XDef5 = v
				return nil
			},
			Validation: func(ctx context.Context, value *localValueType) error {
				var stat *grpcfed.Status
				if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
					Value:      value,
					Expr:       `users[0].id == ''`,
					CacheIndex: 13,
					Body: func(value *localValueType) error {
						errorMessage := "error"
						stat = grpcfed.NewGRPCStatus(grpcfed.InvalidArgumentCode, errorMessage)
						return nil
					},
				}); err != nil {
					return err
				}
				return grpcfed.NewErrorWithLogAttrs(stat.Err(), slog.LevelError, grpcfed.LogAttrs(ctx))
			},
		})
	}

	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)
	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_res(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_post(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_posts(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})
	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_res(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_post(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_user(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if err := def_users(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def__def5(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.Post
	req.Posts = value.vars.Posts
	req.Res = value.vars.Res
	req.User = value.vars.User
	req.Users = value.vars.Users

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	// (grpc.federation.field).by = "post.id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `post.id`,
		CacheIndex: 14,
		Setter: func(v string) error {
			ret.Id = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "post.title"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `post.title`,
		CacheIndex: 15,
		Setter: func(v string) error {
			ret.Title = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "users[0]"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
		Value:      value,
		Expr:       `users[0]`,
		CacheIndex: 16,
		Setter: func(v *User) error {
			ret.User = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *FederationService_Org_Federation_UserArgument) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.User")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.UserArgument", req)}

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	// (grpc.federation.field).by = "$.user_id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `$.user_id`,
		CacheIndex: 17,
		Setter: func(v string) error {
			ret.Id = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponseArgument(v *FederationService_Org_Federation_GetPostResponseArgument) slog.Value {
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
		slog.String("title", v.GetTitle()),
		slog.Any("user", s.logvalue_Org_Federation_User(v.GetUser())),
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

func (s *FederationService) logvalue_Org_Federation_User(v *User) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *FederationService_Org_Federation_UserArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("user_id", v.UserId),
	)
}

func (s *FederationService) logvalue_Post_GetPostRequest(v *post.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Post_GetPostsRequest(v *post.GetPostsRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.GetIds()),
	)
}
