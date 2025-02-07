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
	Ids               []string
	ItemTypes         []Item_ItemType
	Items             []*Posts_PostItem
	PostIds           []string
	Posts             []*post.Post
	Res               *post.GetPostsResponse
	SelectedItemTypes []Item_ItemType
	SourceItemTypes   []user.Item_ItemType
	Users             []*User
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
	// User_UserServiceClient create a gRPC Client to be used to call methods in user.UserService.
	User_UserServiceClient(FederationServiceClientConfig) (user.UserServiceClient, error)
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
	User_UserServiceClient user.UserServiceClient
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
	FederationService_DependentMethod_Post_PostService_GetPosts = "/post.PostService/GetPosts"
	FederationService_DependentMethod_User_UserService_GetUser  = "/user.UserService/GetUser"
)

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
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
	Post_PostServiceClient, err := cfg.Client.Post_PostServiceClient(FederationServiceClientConfig{
		Service: "post.PostService",
	})
	if err != nil {
		return nil, err
	}
	User_UserServiceClient, err := cfg.Client.User_UserServiceClient(FederationServiceClientConfig{
		Service: "user.UserService",
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
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.Item.ItemType", Item_ItemType_value, Item_ItemType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("user.Item.ItemType", user.Item_ItemType_value, user.Item_ItemType_name)...)
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
			User_UserServiceClient: User_UserServiceClient,
		},
	}
	return svc, nil
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
			Posts *Posts
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
				value.vars.Posts = v
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
	req.Posts = value.vars.Posts

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
			Ids               []string
			ItemTypes         []Item_ItemType
			Items             []*Posts_PostItem
			Posts             []*post.Post
			Res               *post.GetPostsResponse
			SelectedItemTypes []Item_ItemType
			SourceItemTypes   []user.Item_ItemType
			Users             []*User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.PostsArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "post.PostService/GetPosts"
		    request { field: "ids", by: "$.post_ids" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostsResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("post.GetPostsResponse"),
			Setter: func(value *localValueType, v *post.GetPostsResponse) error {
				value.vars.Res = v
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
				grpcfed.Logger(ctx).DebugContext(ctx, "call post.PostService/GetPosts", slog.Any("post.GetPostsRequest", s.logvalue_Post_GetPostsRequest(args)))
				ret, err := s.client.Post_PostServiceClient.GetPosts(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPosts, err); err != nil {
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
			Type: grpcfed.CELListType(grpcfed.CELObjectType("post.Post")),
			Setter: func(value *localValueType, v []*post.Post) error {
				value.vars.Posts = v
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
				value.vars.Ids = v
				return nil
			},
			IteratorName:   `post`,
			IteratorType:   grpcfed.CELObjectType("post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.Posts },
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
				value.vars.Items = v
				return nil
			},
			IteratorName:   `iter`,
			IteratorType:   grpcfed.CELObjectType("post.Post"),
			IteratorSource: func(value *localValueType) []*post.Post { return value.vars.Posts },
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
		  name: "source_item_types"
		  by: "[user.Item.ItemType.value('ITEM_TYPE_1'), user.Item.ItemType.value('ITEM_TYPE_2')]"
		}
	*/
	def_source_item_types := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[[]user.Item_ItemType, *localValueType]{
			Name: `source_item_types`,
			Type: grpcfed.CELListType(grpcfed.CELIntType),
			Setter: func(value *localValueType, v []user.Item_ItemType) error {
				value.vars.SourceItemTypes = v
				return nil
			},
			By:           `[user.Item.ItemType.value('ITEM_TYPE_1'), user.Item.ItemType.value('ITEM_TYPE_2')]`,
			ByCacheIndex: 8,
		})
	}

	/*
		def {
		  name: "item_types"
		  map {
		    iterator {
		      name: "typ"
		      src: "source_item_types"
		    }
		  }
		}
	*/
	def_item_types := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]Item_ItemType, user.Item_ItemType, *localValueType]{
			Name: `item_types`,
			Type: grpcfed.CELListType(grpcfed.CELIntType),
			Setter: func(value *localValueType, v []Item_ItemType) error {
				value.vars.ItemTypes = v
				return nil
			},
			IteratorName:   `typ`,
			IteratorType:   grpcfed.CELIntType,
			IteratorSource: func(value *localValueType) []user.Item_ItemType { return value.vars.SourceItemTypes },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				src, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
					Value:      value,
					Expr:       `typ`,
					OutType:    reflect.TypeOf(user.Item_ItemType(0)),
					CacheIndex: 9,
				})
				if err != nil {
					return 0, err
				}
				v := src.(user.Item_ItemType)
				return s.cast_User_Item_ItemType__to__Org_Federation_Item_ItemType(v)
			},
		})
	}

	/*
		def {
		  name: "selected_item_types"
		  map {
		    iterator {
		      name: "typ"
		      src: "source_item_types"
		    }
		  }
		}
	*/
	def_selected_item_types := func(ctx context.Context) error {
		return grpcfed.EvalDefMap(ctx, value, grpcfed.DefMap[[]Item_ItemType, user.Item_ItemType, *localValueType]{
			Name: `selected_item_types`,
			Type: grpcfed.CELListType(grpcfed.CELIntType),
			Setter: func(value *localValueType, v []Item_ItemType) error {
				value.vars.SelectedItemTypes = v
				return nil
			},
			IteratorName:   `typ`,
			IteratorType:   grpcfed.CELIntType,
			IteratorSource: func(value *localValueType) []user.Item_ItemType { return value.vars.SourceItemTypes },
			Iterator: func(ctx context.Context, value *grpcfed.MapIteratorValue) (any, error) {
				src, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
					Value:      value,
					Expr:       `grpc.federation.enum.select(true, typ, user.Item.ItemType.value('ITEM_TYPE_UNSPECIFIED'))`,
					OutType:    reflect.TypeOf((*grpcfedcel.EnumSelector)(nil)),
					CacheIndex: 10,
				})
				if err != nil {
					return 0, err
				}
				v := src.(*grpcfedcel.EnumSelector)
				var dst Item_ItemType
				if err := func() error {
					if v.GetCond() {
						casted, err := s.cast_User_Item_ItemType__to__Org_Federation_Item_ItemType(user.Item_ItemType(v.GetTrueValue()))
						if err != nil {
							return err
						}
						dst = casted
					} else {
						casted, err := s.cast_User_Item_ItemType__to__Org_Federation_Item_ItemType(user.Item_ItemType(v.GetFalseValue()))
						if err != nil {
							return err
						}
						dst = casted
					}
					return nil
				}(); err != nil {
					return 0, err
				}
				return dst, nil
			},
		})
	}

	// A tree view of message dependencies is shown below.
	/*
	   res ─┐
	                    posts ─┐
	                                           ids ─┐
	        source_item_types ─┐                    │
	                                    item_types ─┤
	   res ─┐                                       │
	                    posts ─┐                    │
	                                         items ─┤
	        source_item_types ─┐                    │
	                           selected_item_types ─┤
	   res ─┐                                       │
	                    posts ─┐                    │
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
		if err := def_source_item_types(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_item_types(ctx1); err != nil {
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
		if err := def_source_item_types(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_selected_item_types(ctx1); err != nil {
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
	req.Ids = value.vars.Ids
	req.ItemTypes = value.vars.ItemTypes
	req.Items = value.vars.Items
	req.Posts = value.vars.Posts
	req.Res = value.vars.Res
	req.SelectedItemTypes = value.vars.SelectedItemTypes
	req.SourceItemTypes = value.vars.SourceItemTypes
	req.Users = value.vars.Users

	// create a message value to be returned.
	ret := &Posts{}

	// field binding section.
	// (grpc.federation.field).by = "ids"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]string]{
		Value:      value,
		Expr:       `ids`,
		CacheIndex: 11,
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
		CacheIndex: 12,
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
		CacheIndex: 13,
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
		CacheIndex: 14,
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
		CacheIndex: 15,
		Setter: func(v []*Posts_PostItem) error {
			ret.Items = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "item_types"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]Item_ItemType]{
		Value:      value,
		Expr:       `item_types`,
		CacheIndex: 16,
		Setter: func(v []Item_ItemType) error {
			ret.ItemTypes = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "selected_item_types"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[[]Item_ItemType]{
		Value:      value,
		Expr:       `selected_item_types`,
		CacheIndex: 17,
		Setter: func(v []Item_ItemType) error {
			ret.SelectedItemTypes = v
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
		CacheIndex: 18,
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
			Res  *user.GetUserResponse
			User *user.User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.UserArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "user.UserService/GetUser"
		    request { field: "id", by: "$.user_id" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("user.GetUserResponse"),
			Setter: func(value *localValueType, v *user.GetUserResponse) error {
				value.vars.Res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &user.GetUserRequest{}
				// { field: "id", by: "$.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.user_id`,
					CacheIndex: 19,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call user.UserService/GetUser", slog.Any("user.GetUserRequest", s.logvalue_User_GetUserRequest(args)))
				ret, err := s.client.User_UserServiceClient.GetUser(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_User_UserService_GetUser, err); err != nil {
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
			Type: grpcfed.CELObjectType("user.User"),
			Setter: func(value *localValueType, v *user.User) error {
				value.vars.User = v
				return nil
			},
			By:           `res.user`,
			ByCacheIndex: 20,
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
	req.Res = value.vars.Res
	req.User = value.vars.User

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	ret.Id = value.vars.User.GetId()     // { name: "user", autobind: true }
	ret.Name = value.vars.User.GetName() // { name: "user", autobind: true }

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// cast_User_Item_ItemType__to__Org_Federation_Item_ItemType cast from "user.Item.ItemType" to "org.federation.Item.ItemType".
func (s *FederationService) cast_User_Item_ItemType__to__Org_Federation_Item_ItemType(from user.Item_ItemType) (Item_ItemType, error) {
	switch from {
	case user.Item_ITEM_TYPE_UNSPECIFIED:
		return Item_ITEM_TYPE_UNSPECIFIED, nil
	case user.Item_ITEM_TYPE_1:
		return Item_ITEM_TYPE_1, nil
	case user.Item_ITEM_TYPE_2:
		return Item_ITEM_TYPE_2, nil
	case user.Item_ITEM_TYPE_3:
		return Item_ITEM_TYPE_3, nil
	default:
		return 0, nil
	}
}

// cast_repeated_User_Item_ItemType__to__repeated_Org_Federation_Item_ItemType cast from "repeated user.Item.ItemType" to "repeated org.federation.Item.ItemType".
func (s *FederationService) cast_repeated_User_Item_ItemType__to__repeated_Org_Federation_Item_ItemType(from []user.Item_ItemType) ([]Item_ItemType, error) {
	ret := make([]Item_ItemType, 0, len(from))
	for _, v := range from {
		casted, err := s.cast_User_Item_ItemType__to__Org_Federation_Item_ItemType(v)
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

func (s *FederationService) logvalue_Org_Federation_Item_ItemType(v Item_ItemType) slog.Value {
	switch v {
	case Item_ITEM_TYPE_UNSPECIFIED:
		return slog.StringValue("ITEM_TYPE_UNSPECIFIED")
	case Item_ITEM_TYPE_1:
		return slog.StringValue("ITEM_TYPE_1")
	case Item_ITEM_TYPE_2:
		return slog.StringValue("ITEM_TYPE_2")
	case Item_ITEM_TYPE_3:
		return slog.StringValue("ITEM_TYPE_3")
	}
	return slog.StringValue("")
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
		slog.Any("item_types", s.logvalue_repeated_Org_Federation_Item_ItemType(v.GetItemTypes())),
		slog.Any("selected_item_types", s.logvalue_repeated_Org_Federation_Item_ItemType(v.GetSelectedItemTypes())),
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

func (s *FederationService) logvalue_User_GetUserRequest(v *user.GetUserRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}

func (s *FederationService) logvalue_User_GetUsersRequest(v *user.GetUsersRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("ids", v.GetIds()),
	)
}

func (s *FederationService) logvalue_repeated_Org_Federation_Item_ItemType(v []Item_ItemType) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Org_Federation_Item_ItemType(vv),
		})
	}
	return slog.GroupValue(attrs...)
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
