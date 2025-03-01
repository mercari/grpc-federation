// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: custom_resolver.proto
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

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type FederationService_Org_Federation_GetPostResponseArgument struct {
	Id   string
	Post *Post
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
	Id   string
	Post *post.Post
	Res  *post.GetPostResponse
	User *User
}

// Org_Federation_Post_UserArgument is custom resolver's argument for "user" field of "org.federation.Post" message.
type FederationService_Org_Federation_Post_UserArgument struct {
	*FederationService_Org_Federation_PostArgument
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type FederationService_Org_Federation_UserArgument struct {
	Content string
	Id      string
	Res     *user.GetUserResponse
	Title   string
	U       *user.User
	UserId  string
}

// Org_Federation_User_NameArgument is custom resolver's argument for "name" field of "org.federation.User" message.
type FederationService_Org_Federation_User_NameArgument struct {
	*FederationService_Org_Federation_UserArgument
	Org_Federation_User *User
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
	// Resolve_Org_Federation_Post_User implements resolver for "org.federation.Post.user".
	Resolve_Org_Federation_Post_User(context.Context, *FederationService_Org_Federation_Post_UserArgument) (*User, error)
	// Resolve_Org_Federation_User implements resolver for "org.federation.User".
	Resolve_Org_Federation_User(context.Context, *FederationService_Org_Federation_UserArgument) (*User, error)
	// Resolve_Org_Federation_User_Name implements resolver for "org.federation.User.name".
	Resolve_Org_Federation_User_Name(context.Context, *FederationService_Org_Federation_User_NameArgument) (string, error)
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

// Resolve_Org_Federation_Post_User resolve "org.federation.Post.user".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_Post_User(context.Context, *FederationService_Org_Federation_Post_UserArgument) (ret *User, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_Post_User not implemented")
	return
}

// Resolve_Org_Federation_User resolve "org.federation.User".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User(context.Context, *FederationService_Org_Federation_UserArgument) (ret *User, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_User not implemented")
	return
}

// Resolve_Org_Federation_User_Name resolve "org.federation.User.name".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User_Name(context.Context, *FederationService_Org_Federation_User_NameArgument) (ret string, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_User_Name not implemented")
	return
}

const (
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
	FederationService_DependentMethod_Org_User_UserService_GetUser = "/org.user.UserService/GetUser"
)

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
	cfg                FederationServiceConfig
	logger             *slog.Logger
	errorHandler       grpcfed.ErrorHandler
	celCacheMap        *grpcfed.CELCacheMap
	tracer             trace.Tracer
	resolver           FederationServiceResolver
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
		"grpc.federation.private.org.federation.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.org.federation.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.org.federation.UserArgument": {
			"id":      grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.user.Item.ItemType", user.Item_ItemType_value, user.Item_ItemType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.user.UserType", user.UserType_value, user.UserType_name)...)
	svc := &FederationService{
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
			Post *post.Post
			Res  *post.GetPostResponse
			User *User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.PostArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "org.post.PostService/GetPost"
		    request { field: "id", by: "$.id" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("org.post.GetPostResponse"),
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
					CacheIndex: 3,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/GetPost", slog.Any("org.post.GetPostRequest", s.logvalue_Org_Post_GetPostRequest(args)))
				ret, err := s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
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
		  autobind: true
		  by: "res.post"
		}
	*/
	def_post := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.Post, *localValueType]{
			Name: `post`,
			Type: grpcfed.CELObjectType("org.post.Post"),
			Setter: func(value *localValueType, v *post.Post) error {
				value.vars.Post = v
				return nil
			},
			By:           `res.post`,
			ByCacheIndex: 4,
		})
	}

	/*
		def {
		  name: "user"
		  message {
		    name: "User"
		    args { inline: "post" }
		  }
		}
	*/
	def_user := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name: `user`,
			Type: grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.User = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_UserArgument{}
				// { inline: "post" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*post.Post]{
					Value:      value,
					Expr:       `post`,
					CacheIndex: 5,
					Setter: func(v *post.Post) error {
						args.Id = v.GetId()
						args.Title = v.GetTitle()
						args.Content = v.GetContent()
						args.UserId = v.GetUserId()
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

	if err := def_res(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_post(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_user(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.Post
	req.Res = value.vars.Res
	req.User = value.vars.User

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.Post.GetId()           // { name: "post", autobind: true }
	ret.Title = value.vars.Post.GetTitle()     // { name: "post", autobind: true }
	ret.Content = value.vars.Post.GetContent() // { name: "post", autobind: true }
	{
		// (grpc.federation.field).custom_resolver = true
		ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
		var err error
		ret.User, err = s.resolver.Resolve_Org_Federation_Post_User(ctx, &FederationService_Org_Federation_Post_UserArgument{
			FederationService_Org_Federation_PostArgument: req,
		})
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
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
			Res *user.GetUserResponse
			U   *user.User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.UserArgument", req)}
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
				value.vars.Res = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &user.GetUserRequest{}
				// { field: "id", by: "$.user_id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.user_id`,
					CacheIndex: 6,
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
		  name: "u"
		  by: "res.user"
		}
	*/
	def_u := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.User, *localValueType]{
			Name: `u`,
			Type: grpcfed.CELObjectType("org.user.User"),
			Setter: func(value *localValueType, v *user.User) error {
				value.vars.U = v
				return nil
			},
			By:           `res.user`,
			ByCacheIndex: 7,
		})
	}

	if err := def_res(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	if err := def_u(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = value.vars.Res
	req.U = value.vars.U

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
	ret, err := s.resolver.Resolve_Org_Federation_User(ctx, req)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// field binding section.
	{
		// (grpc.federation.field).custom_resolver = true
		ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
		var err error
		ret.Name, err = s.resolver.Resolve_Org_Federation_User_Name(ctx, &FederationService_Org_Federation_User_NameArgument{
			FederationService_Org_Federation_UserArgument: req,
			Org_Federation_User:                           ret,
		})
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
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
		slog.String("content", v.GetContent()),
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
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *FederationService_Org_Federation_UserArgument) slog.Value {
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
