// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
//
// source: multi_user.proto
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

	user "example/user"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type FederationService_Org_Federation_GetResponseArgument struct {
	Uid   *UserID
	User  *User
	User2 *User
}

// Org_Federation_SubArgument is argument for "org.federation.Sub" message.
type FederationService_Org_Federation_SubArgument struct {
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type FederationService_Org_Federation_UserArgument struct {
	Res    *user.GetUserResponse
	User   *user.User
	UserId string
	XDef2  *Sub
}

// Org_Federation_UserIDArgument is argument for "org.federation.UserID" message.
type FederationService_Org_Federation_UserIDArgument struct {
}

// Org_Federation_User_NameArgument is custom resolver's argument for "name" field of "org.federation.User" message.
type FederationService_Org_Federation_User_NameArgument struct {
	*FederationService_Org_Federation_UserArgument
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
	Org_User_UserServiceClient user.UserServiceClient
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
	// Resolve_Org_Federation_Sub implements resolver for "org.federation.Sub".
	Resolve_Org_Federation_Sub(context.Context, *FederationService_Org_Federation_SubArgument) (*Sub, error)
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

// Resolve_Org_Federation_Sub resolve "org.federation.Sub".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_Sub(context.Context, *FederationService_Org_Federation_SubArgument) (ret *Sub, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_Sub not implemented")
	return
}

// Resolve_Org_Federation_User_Name resolve "org.federation.User.name".
// This method always returns Unimplemented error.
func (FederationServiceUnimplementedResolver) Resolve_Org_Federation_User_Name(context.Context, *FederationService_Org_Federation_User_NameArgument) (ret string, e error) {
	e = grpcfed.GRPCErrorf(grpcfed.UnimplementedCode, "method Resolve_Org_Federation_User_Name not implemented")
	return
}

const (
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
		"grpc.federation.private.org.federation.GetResponseArgument": {},
		"grpc.federation.private.org.federation.SubArgument":         {},
		"grpc.federation.private.org.federation.UserArgument": {
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
		"grpc.federation.private.org.federation.UserIDArgument": {},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
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

// Get implements "org.federation.FederationService/Get" method.
func (s *FederationService) Get(ctx context.Context, req *GetRequest) (res *GetResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/Get")
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
	res, err := s.resolve_Org_Federation_GetResponse(ctx, &FederationService_Org_Federation_GetResponseArgument{})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *FederationService_Org_Federation_GetResponseArgument) (*GetResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			Uid   *UserID
			User  *User
			User2 *User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.GetResponseArgument", req)}
	/*
		def {
		  name: "uid"
		  message {
		    name: "UserID"
		  }
		}
	*/
	def_uid := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*UserID, *localValueType]{
			Name: `uid`,
			Type: grpcfed.CELObjectType("org.federation.UserID"),
			Setter: func(value *localValueType, v *UserID) error {
				value.vars.Uid = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_UserIDArgument{}
				ret, err := s.resolve_Org_Federation_UserID(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "user"
		  message {
		    name: "User"
		    args { name: "user_id", by: "uid.value" }
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
				// { name: "user_id", by: "uid.value" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `uid.value`,
					CacheIndex: 1,
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
		  name: "user2"
		  message {
		    name: "User"
		    args { name: "user_id", by: "uid.value" }
		  }
		}
	*/
	def_user2 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name: `user2`,
			Type: grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.User2 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_UserArgument{}
				// { name: "user_id", by: "uid.value" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `uid.value`,
					CacheIndex: 2,
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

	// A tree view of message dependencies is shown below.
	/*
	   uid ─┐
	         user ─┐
	   uid ─┐      │
	        user2 ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_uid(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_user(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_uid(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def_user2(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Uid = value.vars.Uid
	req.User = value.vars.User
	req.User2 = value.vars.User2

	// create a message value to be returned.
	ret := &GetResponse{}

	// field binding section.
	// (grpc.federation.field).by = "user"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
		Value:      value,
		Expr:       `user`,
		CacheIndex: 3,
		Setter: func(v *User) error {
			ret.User = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "user2"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
		Value:      value,
		Expr:       `user2`,
		CacheIndex: 4,
		Setter: func(v *User) error {
			ret.User2 = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_Sub resolve "org.federation.Sub" message.
func (s *FederationService) resolve_Org_Federation_Sub(ctx context.Context, req *FederationService_Org_Federation_SubArgument) (*Sub, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Sub")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.Sub", slog.Any("message_args", s.logvalue_Org_Federation_SubArgument(req)))

	// create a message value to be returned.
	// `custom_resolver = true` in "grpc.federation.message" option.
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
	ret, err := s.resolver.Resolve_Org_Federation_Sub(ctx, req)
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Sub", slog.Any("org.federation.Sub", s.logvalue_Org_Federation_Sub(ret)))
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
			Res   *user.GetUserResponse
			User  *user.User
			XDef2 *Sub
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
					CacheIndex: 5,
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
				value.vars.User = v
				return nil
			},
			By:           `res.user`,
			ByCacheIndex: 6,
		})
	}

	/*
		def {
		  name: "_def2"
		  message {
		    name: "Sub"
		  }
		}
	*/
	def__def2 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*Sub, *localValueType]{
			Name: `_def2`,
			Type: grpcfed.CELObjectType("org.federation.Sub"),
			Setter: func(value *localValueType, v *Sub) error {
				value.vars.XDef2 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_SubArgument{}
				ret, err := s.resolve_Org_Federation_Sub(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	// A tree view of message dependencies is shown below.
	/*
	        _def2 ─┐
	   res ─┐      │
	         user ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def__def2(ctx1); err != nil {
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
		if err := def_user(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = value.vars.Res
	req.User = value.vars.User
	req.XDef2 = value.vars.XDef2

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	ret.Id = value.vars.User.GetId() // { name: "user", autobind: true }
	{
		// (grpc.federation.field).custom_resolver = true
		ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx)) // create a new reference to logger.
		var err error
		ret.Name, err = s.resolver.Resolve_Org_Federation_User_Name(ctx, &FederationService_Org_Federation_User_NameArgument{
			FederationService_Org_Federation_UserArgument: req,
		})
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// resolve_Org_Federation_UserID resolve "org.federation.UserID" message.
func (s *FederationService) resolve_Org_Federation_UserID(ctx context.Context, req *FederationService_Org_Federation_UserIDArgument) (*UserID, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.UserID")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.UserID", slog.Any("message_args", s.logvalue_Org_Federation_UserIDArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			XDef0 *Sub
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.UserIDArgument", req)}
	/*
		def {
		  name: "_def0"
		  message {
		    name: "Sub"
		  }
		}
	*/
	def__def0 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*Sub, *localValueType]{
			Name: `_def0`,
			Type: grpcfed.CELObjectType("org.federation.Sub"),
			Setter: func(value *localValueType, v *Sub) error {
				value.vars.XDef0 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &FederationService_Org_Federation_SubArgument{}
				ret, err := s.resolve_Org_Federation_Sub(ctx, args)
				if err != nil {
					return nil, err
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
	ret := &UserID{}

	// field binding section.
	// (grpc.federation.field).by = "'xxx'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'xxx'`,
		CacheIndex: 7,
		Setter: func(v string) error {
			ret.Value = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.UserID", slog.Any("org.federation.UserID", s.logvalue_Org_Federation_UserID(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetResponse(v *GetResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("user", s.logvalue_Org_Federation_User(v.GetUser())),
		slog.Any("user2", s.logvalue_Org_Federation_User(v.GetUser2())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *FederationService_Org_Federation_GetResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_Sub(v *Sub) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_SubArgument(v *FederationService_Org_Federation_SubArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
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

func (s *FederationService) logvalue_Org_Federation_UserID(v *UserID) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserIDArgument(v *FederationService_Org_Federation_UserIDArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
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
