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
	"google.golang.org/protobuf/types/known/timestamppb"

	post "example/post"
	user "example/user"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Federation_GetPostResponseArgument is argument for "federation.GetPostResponse" message.
type Federation_GetPostResponseArgument[T any] struct {
	Date       *timestamppb.Timestamp
	FixedRand  *grpcfedcel.Rand
	Id         string
	Loc        *grpcfedcel.Location
	Post       *Post
	RandSource *grpcfedcel.Source
	Uuid       *grpcfedcel.UUID
	Value1     string
	Client     T
}

// Federation_ItemArgument is argument for "federation.Item" message.
type Federation_ItemArgument[T any] struct {
	Client T
}

// Federation_PostArgument is argument for "federation.Post" message.
type Federation_PostArgument[T any] struct {
	Id     string
	Post   *post.Post
	Res    *post.GetPostResponse
	User   *User
	Client T
}

// Federation_UserArgument is argument for "federation.User" message.
type Federation_UserArgument[T any] struct {
	Content string
	Id      string
	Res     *user.GetUserResponse
	Title   string
	User    *user.User
	UserId  string
	Client  T
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
	FederationService_DependentMethod_Post_PostService_GetPost = "/post.PostService/GetPost"
	FederationService_DependentMethod_User_UserService_GetUser = "/user.UserService/GetUser"
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
	Post_PostServiceClient, err := cfg.Client.Post_PostServiceClient(FederationServiceClientConfig{
		Service: "post.PostService",
		Name:    "post_service",
	})
	if err != nil {
		return nil, err
	}
	User_UserServiceClient, err := cfg.Client.User_UserServiceClient(FederationServiceClientConfig{
		Service: "user.UserService",
		Name:    "user_service",
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
		"grpc.federation.private.UserArgument": {
			"id":      grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
			"title":   grpcfed.NewCELFieldType(grpcfed.CELStringType, "Title"),
			"content": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Content"),
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
	})
	envOpts := grpcfed.NewDefaultEnvOptions(celHelper)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("federation.Item.ItemType", Item_ItemType_value, Item_ItemType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("federation.Item.Location.LocationType", Item_Location_LocationType_value, Item_Location_LocationType_name)...)
	envOpts = append(envOpts, grpcfed.EnumAccessorOptions("user.Item.ItemType", user.Item_ItemType_value, user.Item_ItemType_name)...)
	env, err := grpcfed.NewCELEnv(envOpts...)
	if err != nil {
		return nil, err
	}
	return &FederationService{
		cfg:          cfg,
		logger:       logger,
		errorHandler: errorHandler,
		env:          env,
		tracer:       otel.Tracer("federation.FederationService"),
		client: &FederationServiceDependentClientSet{
			Post_PostServiceClient: Post_PostServiceClient,
			User_UserServiceClient: User_UserServiceClient,
		},
	}, nil
}

// GetPost implements "federation.FederationService/GetPost" method.
func (s *FederationService) GetPost(ctx context.Context, req *GetPostRequest) (res *GetPostResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "federation.FederationService/GetPost")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := s.resolve_Federation_GetPostResponse(ctx, &Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]{
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

// resolve_Federation_GetPostResponse resolve "federation.GetPostResponse" message.
func (s *FederationService) resolve_Federation_GetPostResponse(ctx context.Context, req *Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "federation.GetPostResponse")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve federation.GetPostResponse", slog.Any("message_args", s.logvalue_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			date        *timestamppb.Timestamp
			fixed_rand  *grpcfedcel.Rand
			loc         *grpcfedcel.Location
			post        *Post
			rand_source *grpcfedcel.Source
			uuid        *grpcfedcel.UUID
			value1      string
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.GetPostResponseArgument", req)}
	// A tree view of message dependencies is shown below.
	/*
	                                     loc ─┐
	                                    post ─┤
	   date ─┐                                │
	         rand_source ─┐                   │
	                      fixed_rand ─┐       │
	                                    uuid ─┤
	                                  value1 ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "loc"
		     by: "grpc.federation.time.loadLocation('Asia/Tokyo')"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.Location, *localValueType]{
			Name:   "loc",
			Type:   grpcfed.CELObjectType("grpc.federation.time.Location"),
			Setter: func(value *localValueType, v *grpcfedcel.Location) { value.vars.loc = v },
			By:     "grpc.federation.time.loadLocation('Asia/Tokyo')",
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
		     name: "post"
		     message {
		       name: "Post"
		       args { name: "id", by: "$.id" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*Post, *localValueType]{
			Name:   "post",
			Type:   grpcfed.CELObjectType("federation.Post"),
			Setter: func(value *localValueType, v *Post) { value.vars.post = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Federation_PostArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
				}
				// { name: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, value, "$.id", func(v string) {
					args.Id = v
				}); err != nil {
					return nil, err
				}
				return s.resolve_Federation_Post(ctx, args)
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
		     name: "date"
		     by: "grpc.federation.time.date(2023, 12, 25, 12, 10, 5, 0, grpc.federation.time.UTC())"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*timestamppb.Timestamp, *localValueType]{
			Name:   "date",
			Type:   grpcfed.CELObjectType("google.protobuf.Timestamp"),
			Setter: func(value *localValueType, v *timestamppb.Timestamp) { value.vars.date = v },
			By:     "grpc.federation.time.date(2023, 12, 25, 12, 10, 5, 0, grpc.federation.time.UTC())",
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "rand_source"
		     by: "grpc.federation.rand.newSource(date.unix())"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.Source, *localValueType]{
			Name:   "rand_source",
			Type:   grpcfed.CELObjectType("grpc.federation.rand.Source"),
			Setter: func(value *localValueType, v *grpcfedcel.Source) { value.vars.rand_source = v },
			By:     "grpc.federation.rand.newSource(date.unix())",
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "fixed_rand"
		     by: "grpc.federation.rand.new(rand_source)"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.Rand, *localValueType]{
			Name:   "fixed_rand",
			Type:   grpcfed.CELObjectType("grpc.federation.rand.Rand"),
			Setter: func(value *localValueType, v *grpcfedcel.Rand) { value.vars.fixed_rand = v },
			By:     "grpc.federation.rand.new(rand_source)",
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "uuid"
		     by: ".grpc.federation.uuid.newRandomFromRand(fixed_rand)"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.UUID, *localValueType]{
			Name:   "uuid",
			Type:   grpcfed.CELObjectType("grpc.federation.uuid.UUID"),
			Setter: func(value *localValueType, v *grpcfedcel.UUID) { value.vars.uuid = v },
			By:     ".grpc.federation.uuid.newRandomFromRand(fixed_rand)",
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
		     name: "value1"
		     by: "grpc.federation.metadata.incoming()['key1'][0]"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[string, *localValueType]{
			Name:   "value1",
			Type:   grpcfed.CELStringType,
			Setter: func(value *localValueType, v string) { value.vars.value1 = v },
			By:     "grpc.federation.metadata.incoming()['key1'][0]",
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
	req.Date = value.vars.date
	req.FixedRand = value.vars.fixed_rand
	req.Loc = value.vars.loc
	req.Post = value.vars.post
	req.RandSource = value.vars.rand_source
	req.Uuid = value.vars.uuid
	req.Value1 = value.vars.value1

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, value, "post", func(v *Post) { ret.Post = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	ret.Str = "hello" // (grpc.federation.field).string = "hello"
	// (grpc.federation.field).by = "uuid.string()"
	if err := grpcfed.SetCELValue(ctx, value, "uuid.string()", func(v string) { ret.Uuid = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "loc.string()"
	if err := grpcfed.SetCELValue(ctx, value, "loc.string()", func(v string) { ret.Loc = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "value1"
	if err := grpcfed.SetCELValue(ctx, value, "value1", func(v string) { ret.Value1 = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "federation.Item.ItemType.name(federation.Item.ItemType.ITEM_TYPE_1)"
	if err := grpcfed.SetCELValue(ctx, value, "federation.Item.ItemType.name(federation.Item.ItemType.ITEM_TYPE_1)", func(v string) { ret.ItemTypeName = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "federation.Item.Location.LocationType.name(federation.Item.Location.LocationType.LOCATION_TYPE_1)"
	if err := grpcfed.SetCELValue(ctx, value, "federation.Item.Location.LocationType.name(federation.Item.Location.LocationType.LOCATION_TYPE_1)", func(v string) { ret.LocationTypeName = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "user.Item.ItemType.name(user.Item.ItemType.ITEM_TYPE_2)"
	if err := grpcfed.SetCELValue(ctx, value, "user.Item.ItemType.name(user.Item.ItemType.ITEM_TYPE_2)", func(v string) { ret.UserItemTypeName = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "federation.Item.ItemType.value('ITEM_TYPE_1')"
	if err := grpcfed.SetCELValue(ctx, value, "federation.Item.ItemType.value('ITEM_TYPE_1')", func(v int32) { ret.ItemTypeValue = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "federation.Item.Location.LocationType.value('LOCATION_TYPE_1')"
	if err := grpcfed.SetCELValue(ctx, value, "federation.Item.Location.LocationType.value('LOCATION_TYPE_1')", func(v int32) { ret.LocationTypeValue = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "user.Item.ItemType.value('ITEM_TYPE_2')"
	if err := grpcfed.SetCELValue(ctx, value, "user.Item.ItemType.value('ITEM_TYPE_2')", func(v int32) { ret.UserItemTypeValue = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved federation.GetPostResponse", slog.Any("federation.GetPostResponse", s.logvalue_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Federation_Post resolve "federation.Post" message.
func (s *FederationService) resolve_Federation_Post(ctx context.Context, req *Federation_PostArgument[*FederationServiceDependentClientSet]) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve federation.Post", slog.Any("message_args", s.logvalue_Federation_PostArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *post.Post
			res  *post.GetPostResponse
			user *User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.PostArgument", req)}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "res"
	     call {
	       method: "post.PostService/GetPost"
	       request { field: "id", by: "$.id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
		Name:   "res",
		Type:   grpcfed.CELObjectType("post.GetPostResponse"),
		Setter: func(value *localValueType, v *post.GetPostResponse) { value.vars.res = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &post.GetPostRequest{}
			// { field: "id", by: "$.id" }
			if err := grpcfed.SetCELValue(ctx, value, "$.id", func(v string) {
				args.Id = v
			}); err != nil {
				return nil, err
			}
			return grpcfed.WithTimeout[post.GetPostResponse](ctx, "post.PostService/GetPost", 10000000000 /* 10s */, func(ctx context.Context) (*post.GetPostResponse, error) {
				b := grpcfed.NewConstantBackOff(2000000000) /* 2s */
				b = grpcfed.BackOffWithMaxRetries(b, 3)
				b = grpcfed.BackOffWithContext(b, ctx)
				return grpcfed.WithRetry[post.GetPostResponse](b, func() (*post.GetPostResponse, error) {
					return s.client.Post_PostServiceClient.GetPost(ctx, args)
				})
			})
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
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
		Type:   grpcfed.CELObjectType("post.Post"),
		Setter: func(value *localValueType, v *post.Post) { value.vars.post = v },
		By:     "res.post",
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
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
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
		Name:   "user",
		Type:   grpcfed.CELObjectType("federation.User"),
		Setter: func(value *localValueType, v *User) { value.vars.user = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Federation_UserArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			// { inline: "post" }
			if err := grpcfed.SetCELValue(ctx, value, "post", func(v *post.Post) {
				args.Id = v.GetId()
				args.Title = v.GetTitle()
				args.Content = v.GetContent()
				args.UserId = v.GetUserId()
			}); err != nil {
				return nil, err
			}
			return s.resolve_Federation_User(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
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
	if err := grpcfed.SetCELValue(ctx, value, "user", func(v *User) { ret.User = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved federation.Post", slog.Any("federation.Post", s.logvalue_Federation_Post(ret)))
	return ret, nil
}

// resolve_Federation_User resolve "federation.User" message.
func (s *FederationService) resolve_Federation_User(ctx context.Context, req *Federation_UserArgument[*FederationServiceDependentClientSet]) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "federation.User")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve federation.User", slog.Any("message_args", s.logvalue_Federation_UserArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			res  *user.GetUserResponse
			user *user.User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.UserArgument", req)}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "res"
	     call {
	       method: "user.UserService/GetUser"
	       request { field: "id", by: "$.user_id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
		Name:   "res",
		Type:   grpcfed.CELObjectType("user.GetUserResponse"),
		Setter: func(value *localValueType, v *user.GetUserResponse) { value.vars.res = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &user.GetUserRequest{}
			// { field: "id", by: "$.user_id" }
			if err := grpcfed.SetCELValue(ctx, value, "$.user_id", func(v string) {
				args.Id = v
			}); err != nil {
				return nil, err
			}
			return grpcfed.WithTimeout[user.GetUserResponse](ctx, "user.UserService/GetUser", 20000000000 /* 20s */, func(ctx context.Context) (*user.GetUserResponse, error) {
				b := grpcfed.NewExponentialBackOff(&grpcfed.ExponentialBackOffConfig{
					InitialInterval:     1000000000, /* 1s */
					RandomizationFactor: 0.7,
					Multiplier:          1.7,
					MaxInterval:         30000000000, /* 30s */
					MaxElapsedTime:      20000000000, /* 20s */
				})
				b = grpcfed.BackOffWithMaxRetries(b, 3)
				b = grpcfed.BackOffWithContext(b, ctx)
				return grpcfed.WithRetry[user.GetUserResponse](b, func() (*user.GetUserResponse, error) {
					return s.client.User_UserServiceClient.GetUser(ctx, args)
				})
			})
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_User_UserService_GetUser, err); err != nil {
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
		Name:   "user",
		Type:   grpcfed.CELObjectType("user.User"),
		Setter: func(value *localValueType, v *user.User) { value.vars.user = v },
		By:     "res.user",
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
	ret.Id = value.vars.user.GetId()                                                                // { name: "user", autobind: true }
	ret.Name = value.vars.user.GetName()                                                            // { name: "user", autobind: true }
	ret.Items = s.cast_repeated_User_Item__to__repeated_Federation_Item(value.vars.user.GetItems()) // { name: "user", autobind: true }
	ret.Profile = value.vars.user.GetProfile()                                                      // { name: "user", autobind: true }

	switch {
	case s.cast_User_User_AttrA___to__Federation_User_AttrA_(value.vars.user.GetAttrA()) != nil:

		ret.Attr = s.cast_User_User_AttrA___to__Federation_User_AttrA_(value.vars.user.GetAttrA())
	case s.cast_User_User_B__to__Federation_User_B(value.vars.user.GetB()) != nil:

		ret.Attr = s.cast_User_User_B__to__Federation_User_B(value.vars.user.GetB())
	}

	s.logger.DebugContext(ctx, "resolved federation.User", slog.Any("federation.User", s.logvalue_Federation_User(ret)))
	return ret, nil
}

// cast_User_Item_ItemType__to__Federation_Item_ItemType cast from "user.Item.ItemType" to "federation.Item.ItemType".
func (s *FederationService) cast_User_Item_ItemType__to__Federation_Item_ItemType(from user.Item_ItemType) Item_ItemType {
	switch from {
	case user.Item_ITEM_TYPE_UNSPECIFIED:
		return Item_ITEM_TYPE_UNSPECIFIED
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

// cast_User_Item_Location_AddrA___to__Federation_Item_Location_AddrA_ cast from "user.Item.Location.addr_a" to "federation.Item.Location.addr_a".
func (s *FederationService) cast_User_Item_Location_AddrA___to__Federation_Item_Location_AddrA_(from *user.Item_Location_AddrA) *Item_Location_AddrA_ {
	if from == nil {
		return nil
	}
	return &Item_Location_AddrA_{
		AddrA: s.cast_User_Item_Location_AddrA__to__Federation_Item_Location_AddrA(from),
	}
}

// cast_User_Item_Location_AddrA__to__Federation_Item_Location_AddrA cast from "user.Item.Location.AddrA" to "federation.Item.Location.AddrA".
func (s *FederationService) cast_User_Item_Location_AddrA__to__Federation_Item_Location_AddrA(from *user.Item_Location_AddrA) *Item_Location_AddrA {
	if from == nil {
		return nil
	}

	return &Item_Location_AddrA{
		Foo: from.GetFoo(),
	}
}

// cast_User_Item_Location_AddrB__to__Federation_Item_Location_AddrB cast from "user.Item.Location.AddrB" to "federation.Item.Location.AddrB".
func (s *FederationService) cast_User_Item_Location_AddrB__to__Federation_Item_Location_AddrB(from *user.Item_Location_AddrB) *Item_Location_AddrB {
	if from == nil {
		return nil
	}

	return &Item_Location_AddrB{
		Bar: from.GetBar(),
	}
}

// cast_User_Item_Location_B__to__Federation_Item_Location_B cast from "user.Item.Location.b" to "federation.Item.Location.b".
func (s *FederationService) cast_User_Item_Location_B__to__Federation_Item_Location_B(from *user.Item_Location_AddrB) *Item_Location_B {
	if from == nil {
		return nil
	}
	return &Item_Location_B{
		B: s.cast_User_Item_Location_AddrB__to__Federation_Item_Location_AddrB(from),
	}
}

// cast_User_Item_Location__to__Federation_Item_Location cast from "user.Item.Location" to "federation.Item.Location".
func (s *FederationService) cast_User_Item_Location__to__Federation_Item_Location(from *user.Item_Location) *Item_Location {
	if from == nil {
		return nil
	}

	ret := &Item_Location{
		Addr1: from.GetAddr1(),
		Addr2: from.GetAddr2(),
	}
	switch {

	case from.GetAddrA() != nil:
		ret.Addr3 = s.cast_User_Item_Location_AddrA___to__Federation_Item_Location_AddrA_(from.GetAddrA())
	case from.GetB() != nil:
		ret.Addr3 = s.cast_User_Item_Location_B__to__Federation_Item_Location_B(from.GetB())
	}
	return ret
}

// cast_User_Item__to__Federation_Item cast from "user.Item" to "federation.Item".
func (s *FederationService) cast_User_Item__to__Federation_Item(from *user.Item) *Item {
	if from == nil {
		return nil
	}

	return &Item{
		Name:     from.GetName(),
		Type:     s.cast_User_Item_ItemType__to__Federation_Item_ItemType(from.GetType()),
		Value:    from.GetValue(),
		Location: s.cast_User_Item_Location__to__Federation_Item_Location(from.GetLocation()),
	}
}

// cast_User_User_AttrA___to__Federation_User_AttrA_ cast from "user.User.attr_a" to "federation.User.attr_a".
func (s *FederationService) cast_User_User_AttrA___to__Federation_User_AttrA_(from *user.User_AttrA) *User_AttrA_ {
	if from == nil {
		return nil
	}
	return &User_AttrA_{
		AttrA: s.cast_User_User_AttrA__to__Federation_User_AttrA(from),
	}
}

// cast_User_User_AttrA__to__Federation_User_AttrA cast from "user.User.AttrA" to "federation.User.AttrA".
func (s *FederationService) cast_User_User_AttrA__to__Federation_User_AttrA(from *user.User_AttrA) *User_AttrA {
	if from == nil {
		return nil
	}

	return &User_AttrA{
		Foo: from.GetFoo(),
	}
}

// cast_User_User_AttrB__to__Federation_User_AttrB cast from "user.User.AttrB" to "federation.User.AttrB".
func (s *FederationService) cast_User_User_AttrB__to__Federation_User_AttrB(from *user.User_AttrB) *User_AttrB {
	if from == nil {
		return nil
	}

	return &User_AttrB{
		Bar: from.GetBar(),
	}
}

// cast_User_User_B__to__Federation_User_B cast from "user.User.b" to "federation.User.b".
func (s *FederationService) cast_User_User_B__to__Federation_User_B(from *user.User_AttrB) *User_B {
	if from == nil {
		return nil
	}
	return &User_B{
		B: s.cast_User_User_AttrB__to__Federation_User_AttrB(from),
	}
}

// cast_repeated_User_Item__to__repeated_Federation_Item cast from "repeated user.Item" to "repeated federation.Item".
func (s *FederationService) cast_repeated_User_Item__to__repeated_Federation_Item(from []*user.Item) []*Item {
	ret := make([]*Item, 0, len(from))
	for _, v := range from {
		ret = append(ret, s.cast_User_Item__to__Federation_Item(v))
	}
	return ret
}

func (s *FederationService) logvalue_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Federation_Post(v.GetPost())),
		slog.String("str", v.GetStr()),
		slog.String("uuid", v.GetUuid()),
		slog.String("loc", v.GetLoc()),
		slog.String("value1", v.GetValue1()),
		slog.String("item_type_name", v.GetItemTypeName()),
		slog.String("location_type_name", v.GetLocationTypeName()),
		slog.String("user_item_type_name", v.GetUserItemTypeName()),
		slog.Int64("item_type_value", int64(v.GetItemTypeValue())),
		slog.Int64("location_type_value", int64(v.GetLocationTypeValue())),
		slog.Int64("user_item_type_value", int64(v.GetUserItemTypeValue())),
	)
}

func (s *FederationService) logvalue_Federation_GetPostResponseArgument(v *Federation_GetPostResponseArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Federation_Item(v *Item) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
		slog.String("type", s.logvalue_Federation_Item_ItemType(v.GetType()).String()),
		slog.Int64("value", v.GetValue()),
		slog.Any("location", s.logvalue_Federation_Item_Location(v.GetLocation())),
	)
}

func (s *FederationService) logvalue_Federation_Item_ItemType(v Item_ItemType) slog.Value {
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

func (s *FederationService) logvalue_Federation_Item_Location(v *Item_Location) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("addr1", v.GetAddr1()),
		slog.String("addr2", v.GetAddr2()),
		slog.Any("addr_a", s.logvalue_Federation_Item_Location_AddrA(v.GetAddrA())),
		slog.Any("b", s.logvalue_Federation_Item_Location_AddrB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Federation_Item_Location_AddrA(v *Item_Location_AddrA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Federation_Item_Location_AddrB(v *Item_Location_AddrB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Int64("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Federation_Item_Location_LocationType(v Item_Location_LocationType) slog.Value {
	switch v {
	case Item_Location_LOCATION_TYPE_0:
		return slog.StringValue("LOCATION_TYPE_0")
	case Item_Location_LOCATION_TYPE_1:
		return slog.StringValue("LOCATION_TYPE_1")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.Any("user", s.logvalue_Federation_User(v.GetUser())),
	)
}

func (s *FederationService) logvalue_Federation_PostArgument(v *Federation_PostArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Federation_User(v *User) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("name", v.GetName()),
		slog.Any("items", s.logvalue_repeated_Federation_Item(v.GetItems())),
		slog.Any("profile", s.logvalue_Federation_User_ProfileEntry(v.GetProfile())),
		slog.Any("attr_a", s.logvalue_Federation_User_AttrA(v.GetAttrA())),
		slog.Any("b", s.logvalue_Federation_User_AttrB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Federation_UserArgument(v *Federation_UserArgument[*FederationServiceDependentClientSet]) slog.Value {
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

func (s *FederationService) logvalue_Federation_User_AttrA(v *User_AttrA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("foo", v.GetFoo()),
	)
}

func (s *FederationService) logvalue_Federation_User_AttrB(v *User_AttrB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Bool("bar", v.GetBar()),
	)
}

func (s *FederationService) logvalue_Federation_User_ProfileEntry(v map[string]*anypb.Any) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for key, value := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(key),
			Value: s.logvalue_Google_Protobuf_Any(value),
		})
	}
	return slog.GroupValue(attrs...)
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

func (s *FederationService) logvalue_repeated_Federation_Item(v []*Item) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for idx, vv := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(idx),
			Value: s.logvalue_Federation_Item(vv),
		})
	}
	return slog.GroupValue(attrs...)
}
