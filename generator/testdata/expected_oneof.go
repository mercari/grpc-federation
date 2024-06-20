// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: oneof.proto
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
type Org_Federation_GetResponseArgument struct {
	Sel *UserSelection
}

// Org_Federation_MArgument is argument for "org.federation.M" message.
type Org_Federation_MArgument struct {
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument struct {
	UserId string
}

// Org_Federation_UserSelectionArgument is argument for "org.federation.UserSelection" message.
type Org_Federation_UserSelectionArgument struct {
	M     *M
	Ua    *User
	Ub    *User
	Uc    *User
	Value string
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
		"grpc.federation.private.GetResponseArgument": {},
		"grpc.federation.private.MArgument":           {},
		"grpc.federation.private.UserArgument": {
			"user_id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "UserId"),
		},
		"grpc.federation.private.UserSelectionArgument": {
			"value": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Value"),
		},
		"org.federation.UserSelection": {
			"user": grpcfed.NewOneofSelectorFieldType(
				grpcfed.NewCELObjectType("org.federation.User"), "User",
				[]reflect.Type{reflect.TypeOf((*UserSelection_UserA)(nil)), reflect.TypeOf((*UserSelection_UserB)(nil)), reflect.TypeOf((*UserSelection_UserC)(nil))},
				[]string{"GetUserA", "GetUserB", "GetUserC"},
				reflect.Zero(reflect.TypeOf((*User)(nil))),
			),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper(celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
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
		client: &FederationServiceDependentClientSet{
			Org_User_UserServiceClient: Org_User_UserServiceClient,
		},
	}, nil
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
	res, err := s.resolve_Org_Federation_GetResponse(ctx, &Org_Federation_GetResponseArgument{})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *Org_Federation_GetResponseArgument) (*GetResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetResponse")
	defer span.End()

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			sel *UserSelection
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, "grpc.federation.private.GetResponseArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "sel"
	     message {
	       name: "UserSelection"
	       args { name: "value", by: "'foo'" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*UserSelection, *localValueType]{
		Name: `sel`,
		Type: grpcfed.CELObjectType("org.federation.UserSelection"),
		Setter: func(value *localValueType, v *UserSelection) error {
			value.vars.sel = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Org_Federation_UserSelectionArgument{}
			// { name: "value", by: "'foo'" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `'foo'`,
				UseContextLibrary: false,
				CacheIndex:        1,
				Setter: func(v string) error {
					args.Value = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			return s.resolve_Org_Federation_UserSelection(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Sel = value.vars.sel

	// create a message value to be returned.
	ret := &GetResponse{}

	// field binding section.
	// (grpc.federation.field).by = "sel.user"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
		Value:             value,
		Expr:              `sel.user`,
		UseContextLibrary: false,
		CacheIndex:        2,
		Setter: func(v *User) error {
			ret.User = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_M resolve "org.federation.M" message.
func (s *FederationService) resolve_Org_Federation_M(ctx context.Context, req *Org_Federation_MArgument) (*M, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.M")
	defer span.End()

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.M", slog.Any("message_args", s.logvalue_Org_Federation_MArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, "grpc.federation.private.MArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// create a message value to be returned.
	ret := &M{}

	// field binding section.
	// (grpc.federation.field).by = "'foo'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              `'foo'`,
		UseContextLibrary: false,
		CacheIndex:        3,
		Setter: func(v string) error {
			ret.Value = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.M", slog.Any("org.federation.M", s.logvalue_Org_Federation_M(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.User")
	defer span.End()

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			_def0 *user.GetUserResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, "grpc.federation.private.UserArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "_def0"
	     call {
	       method: "org.user.UserService/GetUser"
	       request: [
	         { field: "id", by: "$.user_id" },
	         { field: "foo", by: "1", if: "false" },
	         { field: "bar", by: "'hello'", if: "true" }
	       ]
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
		Name: `_def0`,
		Type: grpcfed.CELObjectType("org.user.GetUserResponse"),
		Setter: func(value *localValueType, v *user.GetUserResponse) error {
			value.vars._def0 = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &user.GetUserRequest{}
			// { field: "id", by: "$.user_id" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
				Value:             value,
				Expr:              `$.user_id`,
				UseContextLibrary: false,
				CacheIndex:        4,
				Setter: func(v string) error {
					args.Id = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { field: "foo", by: "1", if: "false" }
			if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
				Value:             value,
				Expr:              `false`,
				UseContextLibrary: false,
				CacheIndex:        5,
				Body: func(value *localValueType) error {
					return grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[int64]{
						Value:             value,
						Expr:              `1`,
						UseContextLibrary: false,
						CacheIndex:        6,
						Setter: func(v int64) error {
							args.Foobar = &user.GetUserRequest_Foo{
								Foo: v,
							}
							return nil
						},
					})
				},
			}); err != nil {
				return nil, err
			}
			// { field: "bar", by: "'hello'", if: "true" }
			if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
				Value:             value,
				Expr:              `true`,
				UseContextLibrary: false,
				CacheIndex:        7,
				Body: func(value *localValueType) error {
					return grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
						Value:             value,
						Expr:              `'hello'`,
						UseContextLibrary: false,
						CacheIndex:        8,
						Setter: func(v string) error {
							args.Foobar = &user.GetUserRequest_Bar{
								Bar: v,
							}
							return nil
						},
					})
				},
			}); err != nil {
				return nil, err
			}
			grpcfed.Logger(ctx).DebugContext(ctx, "call org.user.UserService/GetUser", slog.Any("org.user.GetUserRequest", s.logvalue_Org_User_GetUserRequest(args)))
			return s.client.Org_User_UserServiceClient.GetUser(ctx, args)
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_User_UserService_GetUser, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	// (grpc.federation.field).by = "$.user_id"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:             value,
		Expr:              `$.user_id`,
		UseContextLibrary: false,
		CacheIndex:        9,
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

// resolve_Org_Federation_UserSelection resolve "org.federation.UserSelection" message.
func (s *FederationService) resolve_Org_Federation_UserSelection(ctx context.Context, req *Org_Federation_UserSelectionArgument) (*UserSelection, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.UserSelection")
	defer span.End()

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.UserSelection", slog.Any("message_args", s.logvalue_Org_Federation_UserSelectionArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			m  *M
			ua *User
			ub *User
			uc *User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celTypeHelper, s.celEnvOpts, s.celPlugins, "grpc.federation.private.UserSelectionArgument", req)}
	defer func() {
		if err := value.Close(ctx); err != nil {
			grpcfed.Logger(ctx).ErrorContext(ctx, err.Error())
		}
	}()

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "m"
	     message {
	       name: "M"
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*M, *localValueType]{
		Name: `m`,
		Type: grpcfed.CELObjectType("org.federation.M"),
		Setter: func(value *localValueType, v *M) error {
			value.vars.m = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Org_Federation_MArgument{}
			return s.resolve_Org_Federation_M(ctx, args)
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.M = value.vars.m
	req.Ua = value.vars.ua
	req.Ub = value.vars.ub
	req.Uc = value.vars.uc

	// create a message value to be returned.
	ret := &UserSelection{}

	// field binding section.

	oneof_UserA, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
		Value:             value,
		Expr:              `m.value == $.value`,
		UseContextLibrary: false,
		OutType:           reflect.TypeOf(true),
		CacheIndex:        10,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	oneof_UserB, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
		Value:             value,
		Expr:              `m.value != $.value`,
		UseContextLibrary: false,
		OutType:           reflect.TypeOf(true),
		CacheIndex:        11,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	switch {
	case oneof_UserA.(bool):

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "ua"
		     message {
		       name: "User"
		       args { name: "user_id", by: "'a'" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name: `ua`,
			Type: grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.ua = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument{}
				// { name: "user_id", by: "'a'" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:             value,
					Expr:              `'a'`,
					UseContextLibrary: false,
					CacheIndex:        12,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
			Value:             value,
			Expr:              `ua`,
			UseContextLibrary: false,
			CacheIndex:        13,
			Setter: func(v *User) error {
				ret.User = &UserSelection_UserA{UserA: v}
				return nil
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	case oneof_UserB.(bool):

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "ub"
		     message {
		       name: "User"
		       args { name: "user_id", by: "'b'" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name: `ub`,
			Type: grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.ub = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument{}
				// { name: "user_id", by: "'b'" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:             value,
					Expr:              `'b'`,
					UseContextLibrary: false,
					CacheIndex:        14,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
			Value:             value,
			Expr:              `ub`,
			UseContextLibrary: false,
			CacheIndex:        15,
			Setter: func(v *User) error {
				ret.User = &UserSelection_UserB{UserB: v}
				return nil
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	default:

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "uc"
		     message {
		       name: "User"
		       args { name: "user_id", by: "$.value" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name: `uc`,
			Type: grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) error {
				value.vars.uc = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument{}
				// { name: "user_id", by: "$.value" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:             value,
					Expr:              `$.value`,
					UseContextLibrary: false,
					CacheIndex:        16,
					Setter: func(v string) error {
						args.UserId = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*User]{
			Value:             value,
			Expr:              `uc`,
			UseContextLibrary: false,
			CacheIndex:        17,
			Setter: func(v *User) error {
				ret.User = &UserSelection_UserC{UserC: v}
				return nil
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.UserSelection", slog.Any("org.federation.UserSelection", s.logvalue_Org_Federation_UserSelection(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetResponse(v *GetResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("user", s.logvalue_Org_Federation_User(v.GetUser())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *Org_Federation_GetResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_M(v *M) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_MArgument(v *Org_Federation_MArgument) slog.Value {
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
	)
}

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *Org_Federation_UserArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("user_id", v.UserId),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserSelection(v *UserSelection) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("user_a", s.logvalue_Org_Federation_User(v.GetUserA())),
		slog.Any("user_b", s.logvalue_Org_Federation_User(v.GetUserB())),
		slog.Any("user_c", s.logvalue_Org_Federation_User(v.GetUserC())),
	)
}

func (s *FederationService) logvalue_Org_Federation_UserSelectionArgument(v *Org_Federation_UserSelectionArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.Value),
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
