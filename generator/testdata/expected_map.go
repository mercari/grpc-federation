// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: map.proto
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
	user "example/user"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostsResponseArgument is argument for "org.federation.GetPostsResponse" message.
type FederationService_Org_Federation_GetPostsResponseArgument struct {
	Ids   []string
	Posts *Posts
}

// Org_Federation_PostsArgument is argument for "org.federation.Posts" message.
type FederationService_Org_Federation_PostsArgument struct {
	Ids             []string
	Items           []*Posts_PostItem
	PostIds         []string
	Posts           []*post.Post
	Res             *post.GetPostsResponse
	SourceUserTypes []user.UserType
	UserTypes       []UserType
	Users           []*User
}

// Org_Federation_Posts_PostItemArgument is argument for "org.federation.PostItem" message.
type FederationService_Org_Federation_Posts_PostItemArgument struct {
	Id string
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type FederationService_Org_Federation_UserArgument struct {
	Res    *user.GetUserResponse
	User   *user.User
	UserId string
}

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// Client provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
	// If this interface is not provided, an error is returned during initialization.
	Client FederationServiceClientFactory // required
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
	// Org_Post_PostServiceClient create a gRPC Client to be used to call methods in org.post.PostService.
	Org_Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
	// Org_User_UserServiceClient create a gRPC Client to be used to call methods in org.user.UserService.
	Org_User_UserServiceClient(FederationServiceClientConfig) (user.UserServiceClient, error)
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
	Org_User_UserServiceClient user.UserServiceClient
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
	// Resolve_Org_Federation_User implements resolver for "org.federation.User".
	Resolve_Org_Federation_User(context.Context, *FederationService_Org_Federation_UserArgument) (*User, error)
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

// Resolve_Org_Federation_User resolve "org.federation.User".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User(context.Context, *FederationService_Org_Federation_UserArgument) (ret *User, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_User not implemented")
	return
}

const (
	FederationService_DependentMethod_Org_Post_PostService_GetPosts = "/org.post.PostService/GetPosts"
	FederationService_DependentMethod_Org_User_UserService_GetUser  = "/org.user.UserService/GetUser"
)

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
	cfg           FederationServiceConfig
	logger        *slog.Logger
	errorHandler  grpcfed.ErrorHandler
	celCacheMap   *grpcfed.CELCacheMap
	tracer        trace.Tracer
	resolver      FederationServiceResolver
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
	if cfg.Resolver == nil {
		return nil, grpcfed.ErrResolverConfig
	}
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
	})
	if err != nil {
		return nil, err
	}
	Org_User_UserServiceClient, err := cfg.Client.Org_User_UserServiceClient(FederationServiceClientConfig{
		Service: "org.user.UserService",
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
		"grpc.federation.private.GetPostsResponseArgument": {
			"ids": grpcfed.NewCELFieldType(grpcfed.NewCELListType(grpcfed.CELStringType), "Ids"),
		},
		"grpc.federation.private.PostsArgument": {
			"post_ids": grpcfed.NewCELFieldType(grpcfed.NewCELListType(grpcfed.CELStringType), "PostIds"),
		},
		"grpc.federation.private.Posts_PostItemArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.UserArgument": {
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.UserType", UserType_value, UserType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.user.Item.ItemType", user.Item_ItemType_value, user.Item_ItemType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.user.UserType", user.UserType_value, user.UserType_name)...)
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		resolver:      cfg.Resolver,
		client: &FederationServiceDependentClientSet{
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
			Org_User_UserServiceClient: Org_User_UserServiceClient,
		},
	}, nil
}

// GetPosts implements "org.federation.FederationService/GetPosts" method.
func (s *FederationService) GetPosts(ctx context.Context, req *GetPostsRequest) (res *GetPostsResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPosts")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, grpcfed.StackTrace())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostsResponse(ctx, &FederationService_Org_Federation_GetPostsResponseArgument{
		Ids: req.GetIds(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostsResponse resolve "org.federation.GetPostsResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostsResponse(ctx context.Context, req *FederationService_Org_Federation_GetPostsResponseArgument) (*GetPostsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostsResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPostsResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostsResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			posts *Posts
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.GetPostsResponseArgument", req)}
	/*
		def {
		  name: "posts"
		  message {
		    name: "Posts"
		    args { name: "post_ids", by: "$.ids" }
		  }
		}
	*/
	def_posts := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*Posts, *localValueType]{
			Name: `posts`,
			Type: grpcfed.CELObjectType("org.federation.Posts"),
			Setter: func(value *localValueType, v *Posts) error {
				value.vars.posts = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_PostsArgument{}
				// { name: "post_ids", by: "$.ids" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
					Value:      value,
					Expr:       `$.ids`,
					CacheIndex: 1,
					Setter: func(v []string) error {
						args.PostIds = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				ret, err := s.resolve_Org_Federation_Posts(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	if err := def_posts(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Posts = value.vars.posts

	// create a message value to be returned.
	ret := &GetPostsResponse{}

	// field binding section.
	// (grpc.federation.field).by = "posts"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Posts]{
		Value:      value,
		Expr:       `posts`,
		CacheIndex: 2,
		Setter: func(v *Posts) error {
			ret.Posts = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetPostsResponse", slog.Any("org.federation.GetPostsResponse", s.logvalue_Org_Federation_GetPostsResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_Posts resolve "org.federation.Posts" message.
func (s *FederationService) resolve_Org_Federation_Posts(ctx context.Context, req *FederationService_Org_Federation_PostsArgument) (*Posts, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Posts")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.Posts", slog.Any("message_args", s.logvalue_Org_Federation_PostsArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			ids               []string
			items             []*Posts_PostItem
			posts             []*post.Post
			res               *post.GetPostsResponse
			source_user_types []user.UserType
			user_types        []UserType
			users             []*User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.PostsArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "org.post.PostService/GetPosts"
		    request { field: "ids", by: "$.post_ids" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostsResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("org.post.GetPostsResponse"),
			Setter: func(value *localValueType, v *post.GetPostsResponse) error {
				value.vars.res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.GetPostsRequest{}
				// { field: "ids", by: "$.post_ids" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
					Value:      value,
					Expr:       `$.post_ids`,
					CacheIndex: 3,
					Setter: func(v []string) error {
						args.Ids = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/GetPosts", slog.Any("org.post.GetPostsRequest", s.logvalue_Org_Post_GetPostsRequest(args)))
				ret, err := s.client.Org_Post_PostServiceClient.GetPosts(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPosts, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "posts"
		  by: "res.posts"
		}
	*/
	def_posts := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[[]*post.Post, *localValueType]{
			Name: `posts`,
			Type: grpcfed.CELListType(grpcfed.CELObjectType("org.post.Post")),
			Setter: func(value *localValueType, v []*post.Post) error {
				value.vars.posts = v
				return nil
			},
			By:           `res.posts`,
			ByCacheIndex: 4,
		})
	}

	/*
		def {
		  name: "ids"
		  map {
		    iterator {
		      name: "post"
		      src: "posts"
		    }
		    by: "post.id"
		  }
		}
	*/
	def_ids := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]string, *post.Post, *localValueType]{
			Name: `ids`,
			Type: grpcfed.CELListType(grpcfed.CELStringType),
			Setter: func(value *localValueType, v []string) error {
				value.vars.ids = v
				return nil
			},
			IteratorName:   `post`,
			IteratorType:   grpcfed.CELObjectType("org.post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.posts },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				return grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
					Value:      value,
					Expr:       `post.id`,
					OutType:    reflect.TypeOf(""),
					CacheIndex: 5,
				})
			},
		})
	}

	/*
		def {
		  name: "users"
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
			Name: `users`,
			Type: grpcfed.CELListType(grpcfed.CELObjectType("org.federation.User")),
			Setter: func(value *localValueType, v []*User) error {
				value.vars.users = v
				return nil
			},
			IteratorName:   `iter`,
			IteratorType:   grpcfed.CELObjectType("org.post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.posts },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				args := &FederationService_Org_Federation_UserArgument{}
				// { name: "user_id", by: "iter.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `iter.user_id`,
					CacheIndex: 6,
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
		  name: "items"
		  map {
		    iterator {
		      name: "iter"
		      src: "posts"
		    }
		    message {
		      name: "PostItem"
		      args { name: "id", by: "iter.id" }
		    }
		  }
		}
	*/
	def_items := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]*Posts_PostItem, *post.Post, *localValueType]{
			Name: `items`,
			Type: grpcfed.CELListType(grpcfed.CELObjectType("org.federation.Posts.PostItem")),
			Setter: func(value *localValueType, v []*Posts_PostItem) error {
				value.vars.items = v
				return nil
			},
			IteratorName:   `iter`,
			IteratorType:   grpcfed.CELObjectType("org.post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.posts },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				args := &FederationService_Org_Federation_Posts_PostItemArgument{}
				// { name: "id", by: "iter.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `iter.id`,
					CacheIndex: 7,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_Posts_PostItem(ctx, args)
			},
		})
	}

	/*
		def {
		  name: "source_user_types"
		  by: "[org.user.UserType.value('USER_TYPE_1'), org.user.UserType.value('USER_TYPE_2')]"
		}
	*/
	def_source_user_types := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[[]user.UserType, *localValueType]{
			Name: `source_user_types`,
			Type: grpcfed.CELListType(grpcfed.CELIntType),
			Setter: func(value *localValueType, v []user.UserType) error {
				value.vars.source_user_types = v
				return nil
			},
			By:           `[org.user.UserType.value('USER_TYPE_1'), org.user.UserType.value('USER_TYPE_2')]`,
			ByCacheIndex: 8,
		})
	}

	/*
		def {
		  name: "user_types"
		  map {
		    iterator {
		      name: "typ"
		      src: "source_user_types"
		    }
		  }
		}
	*/
	def_user_types := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]UserType, user.UserType, *localValueType]{
			Name: `user_types`,
			Type: grpcfed.CELListType(grpcfed.CELIntType),
			Setter: func(value *localValueType, v []UserType) error {
				value.vars.user_types = v
				return nil
			},
			IteratorName:   `typ`,
			IteratorType:   grpcfed.CELIntType,
			IteratorSource: func(value *localValueType) []user.UserType { return value.vars.source_user_types },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				src, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
					Value:      value,
					Expr:       `typ`,
					OutType:    reflect.TypeOf(user.UserType(0)),
					CacheIndex: 9,
				})
				if err != nil {
					return 0, err
				}
				v := src.(user.UserType)
				return s.cast_Org_User_UserType__to__Org_Federation_UserType(v)
			},
		})
	}

	// A tree view of message dependencies is shown below.
	/*
	   res ─┐
	                    posts ─┐
	                                  ids ─┐
	   res ─┐                              │
	                    posts ─┐           │
	                                items ─┤
	        source_user_types ─┐           │
	                           user_types ─┤
	   res ─┐                              │
	                    posts ─┐           │
	                                users ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_res(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_posts(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_ids(ctx1); err != nil {
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
		if err := def_posts(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_items(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_source_user_types(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_user_types(ctx1); err != nil {
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
		if err := def_posts(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_users(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Ids = value.vars.ids
	req.Items = value.vars.items
	req.Posts = value.vars.posts
	req.Res = value.vars.res
	req.SourceUserTypes = value.vars.source_user_types
	req.UserTypes = value.vars.user_types
	req.Users = value.vars.users

	// create a message value to be returned.
	ret := &Posts{}

	// field binding section.
	// (grpc.federation.field).by = "ids"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
		Value:      value,
		Expr:       `ids`,
		CacheIndex: 10,
		Setter: func(v []string) error {
			ret.Ids = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "posts.map(post, post.title)"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
		Value:      value,
		Expr:       `posts.map(post, post.title)`,
		CacheIndex: 11,
		Setter: func(v []string) error {
			ret.Titles = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "posts.map(post, post.content)"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
		Value:      value,
		Expr:       `posts.map(post, post.content)`,
		CacheIndex: 12,
		Setter: func(v []string) error {
			ret.Contents = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "users"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]*User]{
		Value:      value,
		Expr:       `users`,
		CacheIndex: 13,
		Setter: func(v []*User) error {
			ret.Users = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "items"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]*Posts_PostItem]{
		Value:      value,
		Expr:       `items`,
		CacheIndex: 14,
		Setter: func(v []*Posts_PostItem) error {
			ret.Items = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "user_types"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]UserType]{
		Value:      value,
		Expr:       `user_types`,
		CacheIndex: 15,
		Setter: func(v []UserType) error {
			ret.UserTypes = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Posts", slog.Any("org.federation.Posts", s.logvalue_Org_Federation_Posts(ret)))
	return ret, nil
}

// resolve_Org_Federation_Posts_PostItem resolve "org.federation.Posts.PostItem" message.
func (s *FederationService) resolve_Org_Federation_Posts_PostItem(ctx context.Context, req *FederationService_Org_Federation_Posts_PostItemArgument) (*Posts_PostItem, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Posts.PostItem")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.Posts.PostItem", slog.Any("message_args", s.logvalue_Org_Federation_Posts_PostItemArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.Posts_PostItemArgument", req)}

	// create a message value to be returned.
	ret := &Posts_PostItem{}

	// field binding section.
	// (grpc.federation.field).by = "'item_' + $.id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'item_' + $.id`,
		CacheIndex: 16,
		Setter: func(v string) error {
			ret.Name = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Posts.PostItem", slog.Any("org.federation.Posts.PostItem", s.logvalue_Org_Federation_Posts_PostItem(ret)))
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
			res  *user.GetUserResponse
			user *user.User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.UserArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "org.user.UserService/GetUser"
		    request { field: "id", by: "$.user_id" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("org.user.GetUserResponse"),
			Setter: func(value *localValueType, v *user.GetUserResponse) error {
				value.vars.res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &user.GetUserRequest{}
				// { field: "id", by: "$.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.user_id`,
					CacheIndex: 17,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.user.UserService/GetUser", slog.Any("org.user.GetUserRequest", s.logvalue_Org_User_GetUserRequest(args)))
				ret, err := s.client.Org_User_UserServiceClient.GetUser(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_User_UserService_GetUser, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "user"
		  autobind: true
		  by: "res.user"
		}
	*/
	def_user := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.User, *localValueType]{
			Name: `user`,
			Type: grpcfed.CELObjectType("org.user.User"),
			Setter: func(value *localValueType, v *user.User) error {
				value.vars.user = v
				return nil
			},
			By:           `res.user`,
			ByCacheIndex: 18,
		})
	}

	if err := def_res(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_user(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = value.vars.res
	req.User = value.vars.user

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
	ret, err := s.resolver.Resolve_Org_Federation_User(ctx, req)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// cast_Org_User_UserType__to__Org_Federation_UserType cast from "org.user.UserType" to "org.federation.UserType".
func (s *FederationService) cast_Org_User_UserType__to__Org_Federation_UserType(from user.UserType) (UserType, error) {
	switch from {
	case user.UserType_USER_TYPE_1:
		return UserType_USER_TYPE_1, nil
	case user.UserType_USER_TYPE_2:
		return UserType_USER_TYPE_2, nil
	default:
		return 0, nil
	}
}

// cast_repeated_Org_User_UserType__to__repeated_Org_Federation_UserType cast from "repeated org.user.UserType" to "repeated org.federation.UserType".
func (s *FederationService) cast_repeated_Org_User_UserType__to__repeated_Org_Federation_UserType(from []user.UserType) ([]UserType, error) {
	ret := make([]UserType, 0, len(from))
	for _, v := range from {
		casted, err := s.cast_Org_User_UserType__to__Org_Federation_UserType(v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, casted)
	}
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostsResponse(v *GetPostsResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("posts", s.logvalue_Org_Federation_Posts(v.GetPosts())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostsResponseArgument(v *FederationService_Org_Federation_GetPostsResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.Ids),
	)
}

func (s *FederationService) logvalue_Org_Federation_Posts(v *Posts) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.GetIds()),
		slog.Any("titles", v.GetTitles()),
		slog.Any("contents", v.GetContents()),
		slog.Any("users", s.logvalue_repeated_Org_Federation_User(v.GetUsers())),
		slog.Any("items", s.logvalue_repeated_Org_Federation_Posts_PostItem(v.GetItems())),
		slog.Any("user_types", s.logvalue_repeated_Org_Federation_UserType(v.GetUserTypes())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostsArgument(v *FederationService_Org_Federation_PostsArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post_ids", v.PostIds),
	)
}

func (s *FederationService) logvalue_Org_Federation_Posts_PostItem(v *Posts_PostItem) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_Posts_PostItemArgument(v *FederationService_Org_Federation_Posts_PostItemArgument) slog.Value {
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
		slog.String("name", v.GetName()),
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

func (s *FederationService) logvalue_Org_Federation_UserType(v UserType) slog.Value {
	switch v {
	case UserType_USER_TYPE_1:
		return slog.StringValue("USER_TYPE_1")
	case UserType_USER_TYPE_2:
		return slog.StringValue("USER_TYPE_2")
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

func (s *FederationService) logvalue_Org_Post_UpdatePostRequest(v *post.UpdatePostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_Org_User_GetUserRequest(v *user.GetUserRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.Int64("foo", v.GetFoo()),
		slog.String("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_User_GetUsersRequest(v *user.GetUsersRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.GetIds()),
	)
}

func (s *FederationService) logvalue_repeated_Org_Federation_Posts_PostItem(v []*Posts_PostItem) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Org_Federation_Posts_PostItem(vv),
		})
	}
	return slog.GroupValue(attrs...)
}

func (s *FederationService) logvalue_repeated_Org_Federation_User(v []*User) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Org_Federation_User(vv),
		})
	}
	return slog.GroupValue(attrs...)
}

func (s *FederationService) logvalue_repeated_Org_Federation_UserType(v []UserType) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Org_Federation_UserType(vv),
		})
	}
	return slog.GroupValue(attrs...)
}
