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

	user "example/user"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type Org_Federation_GetResponseArgument[T any] struct {
	Sel    *UserSelection
	Client T
}

// Org_Federation_MArgument is argument for "org.federation.M" message.
type Org_Federation_MArgument[T any] struct {
	Client T
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument[T any] struct {
	UserId string
	Client T
}

// Org_Federation_UserSelectionArgument is argument for "org.federation.UserSelection" message.
type Org_Federation_UserSelectionArgument[T any] struct {
	Ua     *User
	Ub     *User
	Uc     *User
	Value  string
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
	User_UserServiceClient, err := cfg.Client.User_UserServiceClient(FederationServiceClientConfig{
		Service: "user.UserService",
		Name:    "",
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
		"grpc.federation.private.GetResponseArgument": {},
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
	})
	envOpts := grpcfed.NewDefaultEnvOptions(celHelper)
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
			User_UserServiceClient: User_UserServiceClient,
		},
	}, nil
}

// Get implements "org.federation.FederationService/Get" method.
func (s *FederationService) Get(ctx context.Context, req *GetRequest) (res *GetResponse, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/Get")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetResponse(ctx, &Org_Federation_GetResponseArgument[*FederationServiceDependentClientSet]{
		Client: s.client,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, s.logger, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *Org_Federation_GetResponseArgument[*FederationServiceDependentClientSet]) (*GetResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetResponse")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			sel *UserSelection
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.GetResponseArgument", req)}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "sel"
	     message {
	       name: "UserSelection"
	       args { name: "value", string: "foo" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*UserSelection, *localValueType]{
		Name:   "sel",
		Type:   grpcfed.CELObjectType("org.federation.UserSelection"),
		Setter: func(value *localValueType, v *UserSelection) { value.vars.sel = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &Org_Federation_UserSelectionArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
				Value:  "foo", // { name: "value", string: "foo" }
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
	if err := grpcfed.SetCELValue(ctx, value, "sel.user", func(v *User) { ret.User = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument[*FederationServiceDependentClientSet]) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.User")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			_def0 *user.GetUserResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.UserArgument", req)}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "_def0"
	     call {
	       method: "user.UserService/GetUser"
	       request { field: "id", by: "$.user_id" }
	     }
	   }
	*/
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*user.GetUserResponse, *localValueType]{
		Name:   "_def0",
		Type:   grpcfed.CELObjectType("user.GetUserResponse"),
		Setter: func(value *localValueType, v *user.GetUserResponse) { value.vars._def0 = v },
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &user.GetUserRequest{}
			// { field: "id", by: "$.user_id" }
			if err := grpcfed.SetCELValue(ctx, value, "$.user_id", func(v string) {
				args.Id = v
			}); err != nil {
				return nil, err
			}
			return s.client.User_UserServiceClient.GetUser(ctx, args)
		},
	}); err != nil {
		if err := s.errorHandler(ctx, FederationService_DependentMethod_User_UserService_GetUser, err); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	// (grpc.federation.field).by = "$.user_id"
	if err := grpcfed.SetCELValue(ctx, value, "$.user_id", func(v string) { ret.Id = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// resolve_Org_Federation_UserSelection resolve "org.federation.UserSelection" message.
func (s *FederationService) resolve_Org_Federation_UserSelection(ctx context.Context, req *Org_Federation_UserSelectionArgument[*FederationServiceDependentClientSet]) (*UserSelection, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.UserSelection")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.UserSelection", slog.Any("message_args", s.logvalue_Org_Federation_UserSelectionArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			ua *User
			ub *User
			uc *User
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.UserSelectionArgument", req)}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Ua = value.vars.ua
	req.Ub = value.vars.ub
	req.Uc = value.vars.uc

	// create a message value to be returned.
	ret := &UserSelection{}

	// field binding section.

	oneof_UserA, err := grpcfed.EvalCEL(ctx, value, "false", reflect.TypeOf(true))
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	oneof_UserB, err := grpcfed.EvalCEL(ctx, value, "true", reflect.TypeOf(true))
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
		       args { name: "user_id", string: "a" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name:   "ua",
			Type:   grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) { value.vars.ua = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
					UserId: "a", // { name: "user_id", string: "a" }
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, value, "ua", func(v *User) {
			ret.User = &UserSelection_UserA{UserA: v}
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
		       args { name: "user_id", string: "b" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*User, *localValueType]{
			Name:   "ub",
			Type:   grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) { value.vars.ub = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
					UserId: "b", // { name: "user_id", string: "b" }
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, value, "ub", func(v *User) {
			ret.User = &UserSelection_UserB{UserB: v}
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
			Name:   "uc",
			Type:   grpcfed.CELObjectType("org.federation.User"),
			Setter: func(value *localValueType, v *User) { value.vars.uc = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_UserArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
				}
				// { name: "user_id", by: "$.value" }
				if err := grpcfed.SetCELValue(ctx, value, "$.value", func(v string) {
					args.UserId = v
				}); err != nil {
					return nil, err
				}
				return s.resolve_Org_Federation_User(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		if err := grpcfed.SetCELValue(ctx, value, "uc", func(v *User) {
			ret.User = &UserSelection_UserC{UserC: v}
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
	}

	s.logger.DebugContext(ctx, "resolved org.federation.UserSelection", slog.Any("org.federation.UserSelection", s.logvalue_Org_Federation_UserSelection(ret)))
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

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *Org_Federation_GetResponseArgument[*FederationServiceDependentClientSet]) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *Org_Federation_UserArgument[*FederationServiceDependentClientSet]) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_UserSelectionArgument(v *Org_Federation_UserSelectionArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.Value),
	)
}
