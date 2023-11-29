// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
package federation

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	post "example/post"
)

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type Org_Federation_GetPostResponseArgument[T any] struct {
	Id     string
	XDef0  *Post
	Client T
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument[T any] struct {
	Id     string
	Res    *post.GetPostResponse
	XDef1  *post.Post
	XDef2  *User
	Client T
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument[T any] struct {
	UserId string
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
	// Post_PostServiceClient create a gRPC Client to be used to call methods in post.PostService.
	Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
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
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
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
	*UnimplementedFederationServiceServer
	cfg          FederationServiceConfig
	logger       *slog.Logger
	errorHandler grpcfed.ErrorHandler
	env          *cel.Env
	tracer       trace.Tracer
	client       *FederationServiceDependentClientSet
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if cfg.Client == nil {
		return nil, fmt.Errorf("Client field in FederationServiceConfig is not set. this field must be set")
	}
	Post_PostServiceClient, err := cfg.Client.Post_PostServiceClient(FederationServiceClientConfig{
		Service: "post.PostService",
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
	celHelper := grpcfed.NewCELTypeHelper(map[string]map[string]*celtypes.FieldType{
		"grpc.federation.private.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(celtypes.StringType, "Id"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(celtypes.StringType, "Id"),
		},
		"grpc.federation.private.UserArgument": {
			"user_id": grpcfed.NewCELFieldType(celtypes.StringType, "UserId"),
		},
	})
	env, err := cel.NewCustomEnv(
		cel.StdLib(),
		cel.CustomTypeAdapter(celHelper.TypeAdapter()),
		cel.CustomTypeProvider(celHelper.TypeProvider()),
	)
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
			Post_PostServiceClient: Post_PostServiceClient,
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
	var (
		sg         singleflight.Group
		valueMu    sync.RWMutex
		value_Def0 *Post
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.GetPostResponseArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

	// This section's codes are generated by the following proto definition.
	/*
	   def {
	     name: "_def0"
	     autobind: true
	     message {
	       name: "Post"
	       args { name: "id", by: "$.id" }
	     }
	   }
	*/
	{
		valueIface, err, _ := sg.Do("_def0", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_PostArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			// { name: "id", by: "$.id" }
			{
				value, err := grpcfed.EvalCEL(s.env, "$.id", envOpts, evalValues, reflect.TypeOf(""))
				if err != nil {
					grpcfed.RecordErrorToSpan(ctx, err)
					return nil, err
				}
				args.Id = value.(string)
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_Post(ctx, args)
		})
		if err != nil {
			return nil, err
		}
		value := valueIface.(*Post)
		valueMu.Lock()
		value_Def0 = value
		envOpts = append(envOpts, cel.Variable("_def0", cel.ObjectType("org.federation.Post")))
		evalValues["_def0"] = value_Def0
		valueMu.Unlock()
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.XDef0 = value_Def0

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	ret.Id = value_Def0.GetId()           // { name: "_def0", autobind: true }
	ret.Title = value_Def0.GetTitle()     // { name: "_def0", autobind: true }
	ret.Content = value_Def0.GetContent() // { name: "_def0", autobind: true }
	ret.Uid = value_Def0.GetUid()         // { name: "_def0", autobind: true }

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument[*FederationServiceDependentClientSet]) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	var (
		sg         singleflight.Group
		valueMu    sync.RWMutex
		valueRes   *post.GetPostResponse
		value_Def1 *post.Post
		value_Def2 *User
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.PostArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}
	// A tree view of message dependencies is shown below.
	/*
	   res ─┐
	        _def1 ─┐
	        _def2 ─┤
	*/
	eg, ctx1 := errgroup.WithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

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
		{
			valueIface, err, _ := sg.Do("res", func() (any, error) {
				valueMu.RLock()
				args := &post.GetPostRequest{}
				// { field: "id", by: "$.id" }
				{
					value, err := grpcfed.EvalCEL(s.env, "$.id", envOpts, evalValues, reflect.TypeOf(""))
					if err != nil {
						grpcfed.RecordErrorToSpan(ctx, err)
						return nil, err
					}
					args.Id = value.(string)
				}
				valueMu.RUnlock()
				return s.client.Post_PostServiceClient.GetPost(ctx1, args)
			})
			if err != nil {
				if err := s.errorHandler(ctx1, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
					grpcfed.RecordErrorToSpan(ctx, err)
					return nil, err
				}
			}
			value := valueIface.(*post.GetPostResponse)
			valueMu.Lock()
			valueRes = value
			envOpts = append(envOpts, cel.Variable("res", cel.ObjectType("post.GetPostResponse")))
			evalValues["res"] = valueRes
			valueMu.Unlock()
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "_def1"
		     autobind: true
		     by: "res.post"
		   }
		*/
		{
			valueIface, err, _ := sg.Do("_def1", func() (any, error) {
				valueMu.RLock()
				valueMu.RUnlock()
				return grpcfed.EvalCEL(s.env, "res.post", envOpts, evalValues, reflect.TypeOf((*post.Post)(nil)))
			})
			if err != nil {
				return nil, err
			}
			value := valueIface.(*post.Post)
			valueMu.Lock()
			value_Def1 = value
			envOpts = append(envOpts, cel.Variable("_def1", cel.ObjectType("post.Post")))
			evalValues["_def1"] = value_Def1
			valueMu.Unlock()
		}
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "_def2"
		     autobind: true
		     message {
		       name: "User"
		       args { name: "user_id", string: "foo" }
		     }
		   }
		*/
		{
			valueIface, err, _ := sg.Do("_def2", func() (any, error) {
				valueMu.RLock()
				args := &Org_Federation_UserArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
					UserId: "foo", // { name: "user_id", string: "foo" }
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_User(ctx1, args)
			})
			if err != nil {
				return nil, err
			}
			value := valueIface.(*User)
			valueMu.Lock()
			value_Def2 = value
			envOpts = append(envOpts, cel.Variable("_def2", cel.ObjectType("org.federation.User")))
			evalValues["_def2"] = value_Def2
			valueMu.Unlock()
		}
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = valueRes
	req.XDef1 = value_Def1
	req.XDef2 = value_Def2

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value_Def1.GetId()           // { name: "_def1", autobind: true }
	ret.Title = value_Def1.GetTitle()     // { name: "_def1", autobind: true }
	ret.Content = value_Def1.GetContent() // { name: "_def1", autobind: true }
	ret.Uid = value_Def2.GetUid()         // { name: "_def2", autobind: true }

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument[*FederationServiceDependentClientSet]) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.User")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.UserArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	// (grpc.federation.field).by = "$.user_id"
	{
		value, err := grpcfed.EvalCEL(s.env, "$.user_id", envOpts, evalValues, reflect.TypeOf(""))
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		ret.Uid = value.(string)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse(v *GetPostResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.String("uid", v.GetUid()),
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
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
		slog.String("uid", v.GetUid()),
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

func (s *FederationService) logvalue_Org_Federation_User(v *User) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("uid", v.GetUid()),
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
