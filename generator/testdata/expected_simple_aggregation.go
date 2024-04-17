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
	"google.golang.org/protobuf/types/known/anypb"

	post "example/post"
	user "example/user"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type Org_Federation_GetPostResponseArgument struct {
	Id   string
	Post *Post
	Uuid *grpcfedcel.UUID
}

// Org_Federation_ItemArgument is argument for "org.federation.Item" message.
type Org_Federation_ItemArgument struct {
}

// Org_Federation_MArgument is argument for "org.federation.M" message.
type Org_Federation_MArgument struct {
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument struct {
	Id   string
	M    *M
	Post *post.Post
	Res  *post.GetPostResponse
	User *User
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument struct {
	Content string
	Id      string
	Res     *user.GetUserResponse
	Title   string
	User    *user.User
	UserId  string
}

// Org_Federation_User_AgeArgument is custom resolver's argument for "age" field of "org.federation.User" message.
type Org_Federation_User_AgeArgument struct {
	*Org_Federation_UserArgument
}

// Org_Federation_User_AttrAArgument is argument for "org.federation.AttrA" message.
type Org_Federation_User_AttrAArgument struct {
}

// Org_Federation_User_AttrBArgument is argument for "org.federation.AttrB" message.
type Org_Federation_User_AttrBArgument struct {
}

// Org_Federation_ZArgument is argument for "org.federation.Z" message.
type Org_Federation_ZArgument struct {
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
	// Resolve_Org_Federation_User_Age implements resolver for "org.federation.User.age".
	Resolve_Org_Federation_User_Age(context.Context, *Org_Federation_User_AgeArgument) (uint64, error)
	// Resolve_Org_Federation_Z implements resolver for "org.federation.Z".
	Resolve_Org_Federation_Z(context.Context, *Org_Federation_ZArgument) (*Z, error)
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

// Resolve_Org_Federation_User_Age resolve "org.federation.User.age".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User_Age(context.Context, *Org_Federation_User_AgeArgument) (ret uint64, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_User_Age not implemented")
	return
}

// Resolve_Org_Federation_Z resolve "org.federation.Z".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_Z(context.Context, *Org_Federation_ZArgument) (ret *Z, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_Z not implemented")
	return
}

const (
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
	FederationService_DependentMethod_Org_User_UserService_GetUser = "/org.user.UserService/GetUser"
)

// FederationService represents Federation Service.
type FederationService struct {
	*UnimplementedFederationServiceServer
	cfg           FederationServiceConfig
	logger        *slog.Logger
	errorHandler  grpcfed.ErrorHandler
	celCacheMap   *grpcfed.CELCacheMap
	tracer        trace.Tracer
	resolver      FederationServiceResolver
	celTypeHelper *grpcfed.CELTypeHelper
	envOpts       []grpcfed.CELEnvOption
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
		"grpc.federation.private.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.MArgument": {},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.UserArgument": {
			"id":      grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
		"grpc.federation.private.ZArgument": {},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper(celTypeHelperFieldMap)
	var envOpts []grpcfed.CELEnvOption
	envOpts = append(envOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.federation.Item.ItemType", Item_ItemType_value, Item_ItemType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.federation.UserType", UserType_value, UserType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.user.Item.ItemType", user.Item_ItemType_value, user.Item_ItemType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("org.user.UserType", user.UserType_value, user.UserType_name)...)
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		envOpts:       envOpts,
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

// GetPost implements "org.federation.FederationService/GetPost" method.
func (s *FederationService) GetPost(ctx context.Context, req *GetPostRequest) (res *GetPostResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := grpcfed.WithTimeout[GetPostResponse](ctx, "org.federation.FederationService/GetPost", 60000000000 /* 1m0s */, func(ctx context.Context) (*GetPostResponse, error) {
		return s.resolve_Org_Federation_GetPostResponse(ctx, &Org_Federation_GetPostResponseArgument{
			Id: req.Id,
		})
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, s.logger, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *Org_Federation_GetPostResponseArgument) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *Post
			uuid *grpcfedcel.UUID
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.envOpts, s.celPlugins, "grpc.federation.private.GetPostResponseArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			s.logger.ErrorContext(ctx, err.Error())
		}
	}()
	// A tree view of message dependencies is shown below.
	/*
	   post ─┐
	   uuid ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*Post, *localValueType]{
			Name:   "post",
			Type:   grpcfed.CELObjectType("org.federation.Post"),
			Setter: func(value *localValueType, v *Post) { value.vars.post = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_PostArgument{}
				// { name: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:             value,
					Expr:              "$.id",
					UseContextLibrary: false,
					CacheIndex:        1,
					Setter: func(v string) {
						args.Id = v
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_Post(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "uuid"
		     by: "grpc.federation.uuid.newRandom()"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.UUID, *localValueType]{
			Name:                "uuid",
			Type:                grpcfed.CELObjectType("grpc.federation.uuid.UUID"),
			Setter:              func(value *localValueType, v *grpcfedcel.UUID) { value.vars.uuid = v },
			By:                  "grpc.federation.uuid.newRandom()",
			ByUseContextLibrary: false,
			ByCacheIndex:        2,
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.post
	req.Uuid = value.vars.uuid

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:             value,
		Expr:              "post",
		UseContextLibrary: false,
		CacheIndex:        3,
		Setter:            func(v *Post) { ret.Post = v },
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	ret.Const = "foo" // (grpc.federation.field).string = "foo"
	// (grpc.federation.field).by = "uuid.string()"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              "uuid.string()",
		UseContextLibrary: false,
		CacheIndex:        4,
		Setter:            func(v string) { ret.Uuid = v },
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "org.federation.Item.ItemType.name(org.federation.Item.ItemType.ITEM_TYPE_1)"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              "org.federation.Item.ItemType.name(org.federation.Item.ItemType.ITEM_TYPE_1)",
		UseContextLibrary: false,
		CacheIndex:        5,
		Setter:            func(v string) { ret.EnumName = v },
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "org.federation.Item.ItemType.value('ITEM_TYPE_1')"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[Item_ItemType]{
		Value:             value,
		Expr:              "org.federation.Item.ItemType.value('ITEM_TYPE_1')",
		UseContextLibrary: false,
		CacheIndex:        6,
		Setter:            func(v Item_ItemType) { ret.EnumValue = s.cast_Org_Federation_Item_ItemType__to__int32(v) },
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_M resolve "org.federation.M" message.
func (s *FederationService) resolve_Org_Federation_M(ctx context.Context, req *Org_Federation_MArgument) (*M, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.M")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.M", slog.Any("message_args", s.logvalue_Org_Federation_MArgument(req)))

	// create a message value to be returned.
	ret := &M{}

	// field binding section.
	ret.Foo = "foo" // (grpc.federation.field).string = "foo"
	ret.Bar = 1     // (grpc.federation.field).int64 = 1

	s.logger.DebugContext(ctx, "resolved org.federation.M", slog.Any("org.federation.M", s.logvalue_Org_Federation_M(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			m    *M
			post *post.Post
			res  *post.GetPostResponse
			user *User
			z    *Z
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.envOpts, s.celPlugins, "grpc.federation.private.PostArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			s.logger.ErrorContext(ctx, err.Error())
		}
	}()
	// A tree view of message dependencies is shown below.
	/*
	                 m ─┐
	   res ─┐           │
	        post ─┐     │
	              user ─┤
	                 z ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "m"
		     autobind: true
		     message {
		       name: "M"
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*M, *localValueType]{
			Name:   "m",
			Type:   grpcfed.CELObjectType("org.federation.M"),
			Setter: func(value *localValueType, v *M) { value.vars.m = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_MArgument{}
				return s.resolve_Org_Federation_M(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
			Name:   "res",
			Type:   grpcfed.CELObjectType("org.post.GetPostResponse"),
			Setter: func(value *localValueType, v *post.GetPostResponse) { value.vars.res = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.GetPostRequest{}
				// { field: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:             value,
					Expr:              "$.id",
					UseContextLibrary: false,
					CacheIndex:        7,
					Setter: func(v string) {
						args.Id = v
					},
				}); err != nil {
					return nil, err
				}
				return grpcfed.WithTimeout[post.GetPostResponse](ctx, "org.post.PostService/GetPost", 10000000000 /* 10s */, func(ctx context.Context) (*post.GetPostResponse, error) {
					b := grpcfed.NewConstantBackOff(2000000000) /* 2s */
					b = grpcfed.BackOffWithMaxRetries(b, 3)
					b = grpcfed.BackOffWithContext(b, ctx)
					return grpcfed.WithRetry(ctx, &grpcfed.RetryParam[post.GetPostResponse]{
						Value:             value,
						If:                "true",
						UseContextLibrary: false,
						CacheIndex:        8,
						BackOff:           b,
						Body: func() (*post.GetPostResponse, error) {
							return s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
						},
					})
				})
			},
		}); err != nil {
			if err := s.errorHandler(ctx1, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
				grpcfed.RecordErrorToSpan(ctx1, err)
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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*post.Post, *localValueType]{
			Name:                "post",
			Type:                grpcfed.CELObjectType("org.post.Post"),
			Setter:              func(value *localValueType, v *post.Post) { value.vars.post = v },
			By:                  "res.post",
			ByUseContextLibrary: false,
			ByCacheIndex:        9,
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "user"
		     message {
		       name: "User"
		       args { inline: "post" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*User, *localValueType]{
			Name:   "user",
			Type:   grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) { value.vars.user = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument{}
				// { inline: "post" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*post.Post]{
					Value:             value,
					Expr:              "post",
					UseContextLibrary: false,
					CacheIndex:        10,
					Setter: func(v *post.Post) {
						args.Id = v.GetId()
						args.Title = v.GetTitle()
						args.Content = v.GetContent()
						args.UserId = v.GetUserId()
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "z"
		     message {
		       name: "Z"
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*Z, *localValueType]{
			Name:   "z",
			Type:   grpcfed.CELObjectType("org.federation.Z"),
			Setter: func(value *localValueType, v *Z) { value.vars.z = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_ZArgument{}
				return s.resolve_Org_Federation_Z(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.M = value.vars.m
	req.Post = value.vars.post
	req.Res = value.vars.res
	req.User = value.vars.user

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.post.GetId()           // { name: "post", autobind: true }
	ret.Title = value.vars.post.GetTitle()     // { name: "post", autobind: true }
	ret.Content = value.vars.post.GetContent() // { name: "post", autobind: true }
	// (grpc.federation.field).by = "user"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
		Value:             value,
		Expr:              "user",
		UseContextLibrary: false,
		CacheIndex:        11,
		Setter:            func(v *User) { ret.User = v },
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	ret.Foo = value.vars.m.GetFoo() // { name: "m", autobind: true }
	ret.Bar = value.vars.m.GetBar() // { name: "m", autobind: true }

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.User")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			res  *user.GetUserResponse
			user *user.User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.envOpts, s.celPlugins, "grpc.federation.private.UserArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			s.logger.ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "res"
	     call {
	       method: "org.user.UserService/GetUser"
	       request { field: "id", by: "$.user_id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
		Name:   "res",
		Type:   grpcfed.CELObjectType("org.user.GetUserResponse"),
		Setter: func(value *localValueType, v *user.GetUserResponse) { value.vars.res = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &user.GetUserRequest{}
			// { field: "id", by: "$.user_id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              "$.user_id",
				UseContextLibrary: false,
				CacheIndex:        12,
				Setter: func(v string) {
					args.Id = v
				},
			}); err != nil {
				return nil, err
			}
			return grpcfed.WithTimeout[user.GetUserResponse](ctx, "org.user.UserService/GetUser", 20000000000 /* 20s */, func(ctx context.Context) (*user.GetUserResponse, error) {
				b := grpcfed.NewExponentialBackOff(&grpcfed.ExponentialBackOffConfig{
					InitialInterval:     1000000000, /* 1s */
					RandomizationFactor: 0.7,
					Multiplier:          1.7,
					MaxInterval:         30000000000, /* 30s */
					MaxElapsedTime:      20000000000, /* 20s */
				})
				b = grpcfed.BackOffWithMaxRetries(b, 3)
				b = grpcfed.BackOffWithContext(b, ctx)
				return grpcfed.WithRetry(ctx, &grpcfed.RetryParam[user.GetUserResponse]{
					Value:             value,
					If:                "error.code != google.rpc.Code.UNIMPLEMENTED",
					UseContextLibrary: false,
					CacheIndex:        13,
					BackOff:           b,
					Body: func() (*user.GetUserResponse, error) {
						return s.client.Org_User_UserServiceClient.GetUser(ctx, args)
					},
				})
			})
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_User_UserService_GetUser, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "user"
	     autobind: true
	     by: "res.user"
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.User, *localValueType]{
		Name:                "user",
		Type:                grpcfed.CELObjectType("org.user.User"),
		Setter:              func(value *localValueType, v *user.User) { value.vars.user = v },
		By:                  "res.user",
		ByUseContextLibrary: false,
		ByCacheIndex:        14,
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = value.vars.res
	req.User = value.vars.user

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	ret.Id = value.vars.user.GetId()                                                            // { name: "user", autobind: true }
	ret.Type = s.cast_Org_User_UserType__to__Org_Federation_UserType(value.vars.user.GetType()) // { name: "user", autobind: true }
	ret.Name = value.vars.user.GetName()                                                        // { name: "user", autobind: true }
	{
		// (grpc.federation.field).custom_resolver = true
		var err error
		ret.Age, err = s.resolver.Resolve_Org_Federation_User_Age(ctx, &Org_Federation_User_AgeArgument{
			Org_Federation_UserArgument: req,
		})
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}
	ret.Desc = value.vars.user.GetDesc()                                                                    // { name: "user", autobind: true }
	ret.MainItem = s.cast_Org_User_Item__to__Org_Federation_Item(value.vars.user.GetMainItem())             // { name: "user", autobind: true }
	ret.Items = s.cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item(value.vars.user.GetItems()) // { name: "user", autobind: true }
	ret.Profile = value.vars.user.GetProfile()                                                              // { name: "user", autobind: true }

	switch {
	case s.cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(value.vars.user.GetAttrA()) != nil:

		ret.Attr = s.cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(value.vars.user.GetAttrA())
	case s.cast_Org_User_User_B__to__Org_Federation_User_B(value.vars.user.GetB()) != nil:

		ret.Attr = s.cast_Org_User_User_B__to__Org_Federation_User_B(value.vars.user.GetB())
	}

	s.logger.DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// resolve_Org_Federation_Z resolve "org.federation.Z" message.
func (s *FederationService) resolve_Org_Federation_Z(ctx context.Context, req *Org_Federation_ZArgument) (*Z, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Z")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Z", slog.Any("message_args", s.logvalue_Org_Federation_ZArgument(req)))

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ret, err := s.resolver.Resolve_Org_Federation_Z(ctx, req)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.Z", slog.Any("org.federation.Z", s.logvalue_Org_Federation_Z(ret)))
	return ret, nil
}

// cast_Org_Federation_Item_ItemType__to__int32 cast from "org.federation.Item.ItemType" to "int32".
func (s *FederationService) cast_Org_Federation_Item_ItemType__to__int32(from Item_ItemType) int32 {
	return int32(from)
}

// cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType cast from "org.user.Item.ItemType" to "org.federation.Item.ItemType".
func (s *FederationService) cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType(from user.Item_ItemType) Item_ItemType {
	switch from {
	case user.Item_ITEM_TYPE_1:
		return Item_ITEM_TYPE_1
	case user.Item_ITEM_TYPE_2:
		return Item_ITEM_TYPE_2
	case user.Item_ITEM_TYPE_3:
		return Item_ITEM_TYPE_3
	default:
		return 0
	}
}

// cast_Org_User_Item__to__Org_Federation_Item cast from "org.user.Item" to "org.federation.Item".
func (s *FederationService) cast_Org_User_Item__to__Org_Federation_Item(from *user.Item) *Item {
	if from == nil {
		return nil
	}

	return &Item{
		Name:  from.GetName(),
		Type:  s.cast_Org_User_Item_ItemType__to__Org_Federation_Item_ItemType(from.GetType()),
		Value: from.GetValue(),
	}
}

// cast_Org_User_UserType__to__Org_Federation_UserType cast from "org.user.UserType" to "org.federation.UserType".
func (s *FederationService) cast_Org_User_UserType__to__Org_Federation_UserType(from user.UserType) UserType {
	switch from {
	case user.UserType_USER_TYPE_1:
		return UserType_USER_TYPE_1
	case user.UserType_USER_TYPE_2:
		return UserType_USER_TYPE_2
	default:
		return 0
	}
}

// cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_ cast from "org.user.User.attr_a" to "org.federation.User.attr_a".
func (s *FederationService) cast_Org_User_User_AttrA___to__Org_Federation_User_AttrA_(from *user.User_AttrA) *User_AttrA_ {
	if from == nil {
		return nil
	}
	return &User_AttrA_{
		AttrA: s.cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA(from),
	}
}

// cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA cast from "org.user.User.AttrA" to "org.federation.User.AttrA".
func (s *FederationService) cast_Org_User_User_AttrA__to__Org_Federation_User_AttrA(from *user.User_AttrA) *User_AttrA {
	if from == nil {
		return nil
	}

	return &User_AttrA{
		Foo: from.GetFoo(),
	}
}

// cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB cast from "org.user.User.AttrB" to "org.federation.User.AttrB".
func (s *FederationService) cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB(from *user.User_AttrB) *User_AttrB {
	if from == nil {
		return nil
	}

	return &User_AttrB{
		Bar: from.GetBar(),
	}
}

// cast_Org_User_User_B__to__Org_Federation_User_B cast from "org.user.User.b" to "org.federation.User.b".
func (s *FederationService) cast_Org_User_User_B__to__Org_Federation_User_B(from *user.User_AttrB) *User_B {
	if from == nil {
		return nil
	}
	return &User_B{
		B: s.cast_Org_User_User_AttrB__to__Org_Federation_User_AttrB(from),
	}
}

// cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item cast from "repeated org.user.Item" to "repeated org.federation.Item".
func (s *FederationService) cast_repeated_Org_User_Item__to__repeated_Org_Federation_Item(from []*user.Item) []*Item {
	ret := make([]*Item, 0, len(from))
	for _, v := range from {
		ret = append(ret, s.cast_Org_User_Item__to__Org_Federation_Item(v))
	}
	return ret
}

func (s *FederationService) logvalue_Google_Protobuf_Any(v *anypb.Any) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("type_url", v.GetTypeUrl()),
		slog.String("value", string(v.GetValue())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
		slog.String("const", v.GetConst()),
		slog.String("uuid", v.GetUuid()),
		slog.String("enum_name", v.GetEnumName()),
		slog.Int64("enum_value", int64(v.GetEnumValue())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponseArgument(v *Org_Federation_GetPostResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_Item(v *Item) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
		slog.String("type", s.logvalue_Org_Federation_Item_ItemType(v.GetType()).String()),
		slog.Int64("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_Item_ItemType(v Item_ItemType) slog.Value {
	switch v {
	case Item_ITEM_TYPE_1:
		return slog.StringValue("ITEM_TYPE_1")
	case Item_ITEM_TYPE_2:
		return slog.StringValue("ITEM_TYPE_2")
	case Item_ITEM_TYPE_3:
		return slog.StringValue("ITEM_TYPE_3")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_M(v *M) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
		slog.Int64("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_MArgument(v *Org_Federation_MArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.Any("user", s.logvalue_Org_Federation_User(v.GetUser())),
		slog.String("foo", v.GetFoo()),
		slog.Int64("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *Org_Federation_PostArgument) slog.Value {
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
		slog.String("type", s.logvalue_Org_Federation_UserType(v.GetType()).String()),
		slog.String("name", v.GetName()),
		slog.Uint64("age", v.GetAge()),
		slog.Any("desc", v.GetDesc()),
		slog.Any("main_item", s.logvalue_Org_Federation_Item(v.GetMainItem())),
		slog.Any("items", s.logvalue_repeated_Org_Federation_Item(v.GetItems())),
		slog.Any("profile", s.logvalue_Org_Federation_User_ProfileEntry(v.GetProfile())),
		slog.Any("attr_a", s.logvalue_Org_Federation_User_AttrA(v.GetAttrA())),
		slog.Any("b", s.logvalue_Org_Federation_User_AttrB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *Org_Federation_UserArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
		slog.String("title", v.Title),
		slog.String("content", v.Content),
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

func (s *FederationService) logvalue_Org_Federation_User_AttrA(v *User_AttrA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Org_Federation_User_AttrB(v *User_AttrB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Bool("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Org_Federation_User_ProfileEntry(v map[string]*anypb.Any) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for key, value := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(key),
			Value: s.logvalue_Google_Protobuf_Any(value),
		})
	}
	return slog.GroupValue(attrs...)
}

func (s *FederationService) logvalue_Org_Federation_Z(v *Z) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Org_Federation_ZArgument(v *Org_Federation_ZArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_repeated_Org_Federation_Item(v []*Item) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Org_Federation_Item(vv),
		})
	}
	return slog.GroupValue(attrs...)
}
