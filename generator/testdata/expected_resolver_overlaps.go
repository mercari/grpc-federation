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
	"golang.org/x/sync/singleflight"

	post "example/post"
)

// Org_Federation_GetPostResponse1Argument is argument for "org.federation.GetPostResponse1" message.
type Org_Federation_GetPostResponse1Argument[T any] struct {
	Id     string
	Post   *Post
	Client T
}

// Org_Federation_GetPostResponse2Argument is argument for "org.federation.GetPostResponse2" message.
type Org_Federation_GetPostResponse2Argument[T any] struct {
	Id     string
	Post   *Post
	Client T
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument[T any] struct {
	Id     string
	Res    *post.GetPostResponse
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
	// Org_Post_PostServiceClient create a gRPC Client to be used to call methods in org.post.PostService.
	Org_Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
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
	Org_Post_PostServiceClient post.PostServiceClient
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
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
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
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
		Name:    "post_service",
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
		"grpc.federation.private.GetPostResponse1Argument": {
			"id": grpcfed.NewCELFieldType(celtypes.StringType, "Id"),
		},
		"grpc.federation.private.GetPostResponse2Argument": {
			"id": grpcfed.NewCELFieldType(celtypes.StringType, "Id"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(celtypes.StringType, "Id"),
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
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
		},
	}, nil
}

// GetPost1 implements "org.federation.FederationService/GetPost1" method.
func (s *FederationService) GetPost1(ctx context.Context, req *GetPostRequest) (res *GetPostResponse1, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost1")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostResponse1(ctx, &Org_Federation_GetPostResponse1Argument[*FederationServiceDependentClientSet]{
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

// GetPost2 implements "org.federation.FederationService/GetPost2" method.
func (s *FederationService) GetPost2(ctx context.Context, req *GetPostRequest) (res *GetPostResponse2, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost2")
	defer span.End()

	ctx = grpcfed.WithLogger(ctx, s.logger)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, debug.Stack())
			grpcfed.OutputErrorLog(ctx, s.logger, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPostResponse2(ctx, &Org_Federation_GetPostResponse2Argument[*FederationServiceDependentClientSet]{
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

// resolve_Org_Federation_GetPostResponse1 resolve "org.federation.GetPostResponse1" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse1(ctx context.Context, req *Org_Federation_GetPostResponse1Argument[*FederationServiceDependentClientSet]) (*GetPostResponse1, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse1")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetPostResponse1", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponse1Argument(req)))
	var (
		sg        singleflight.Group
		valueMu   sync.RWMutex
		valuePost *Post
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.GetPostResponse1Argument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

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
	{
		valueIface, err, _ := sg.Do("post", func() (any, error) {
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
		valuePost = value
		envOpts = append(envOpts, cel.Variable("post", cel.ObjectType("org.federation.Post")))
		evalValues["post"] = valuePost
		valueMu.Unlock()
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = valuePost

	// create a message value to be returned.
	ret := &GetPostResponse1{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	{
		value, err := grpcfed.EvalCEL(s.env, "post", envOpts, evalValues, reflect.TypeOf((*Post)(nil)))
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		ret.Post = value.(*Post)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse1", slog.Any("org.federation.GetPostResponse1", s.logvalue_Org_Federation_GetPostResponse1(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetPostResponse2 resolve "org.federation.GetPostResponse2" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse2(ctx context.Context, req *Org_Federation_GetPostResponse2Argument[*FederationServiceDependentClientSet]) (*GetPostResponse2, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse2")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetPostResponse2", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponse2Argument(req)))
	var (
		sg        singleflight.Group
		valueMu   sync.RWMutex
		valuePost *Post
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.GetPostResponse2Argument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

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
	{
		valueIface, err, _ := sg.Do("post", func() (any, error) {
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
		valuePost = value
		envOpts = append(envOpts, cel.Variable("post", cel.ObjectType("org.federation.Post")))
		evalValues["post"] = valuePost
		valueMu.Unlock()
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = valuePost

	// create a message value to be returned.
	ret := &GetPostResponse2{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	{
		value, err := grpcfed.EvalCEL(s.env, "post", envOpts, evalValues, reflect.TypeOf((*Post)(nil)))
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		ret.Post = value.(*Post)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse2", slog.Any("org.federation.GetPostResponse2", s.logvalue_Org_Federation_GetPostResponse2(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument[*FederationServiceDependentClientSet]) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))
	var (
		sg       singleflight.Group
		valueMu  sync.RWMutex
		valueRes *post.GetPostResponse
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.PostArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

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
			return s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
		})
		if err != nil {
			if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
				grpcfed.RecordErrorToSpan(ctx, err)
				return nil, err
			}
		}
		value := valueIface.(*post.GetPostResponse)
		valueMu.Lock()
		valueRes = value
		envOpts = append(envOpts, cel.Variable("res", cel.ObjectType("org.post.GetPostResponse")))
		evalValues["res"] = valueRes
		valueMu.Unlock()
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Res = valueRes

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	// (grpc.federation.field).by = "res.post.id"
	{
		value, err := grpcfed.EvalCEL(s.env, "res.post.id", envOpts, evalValues, reflect.TypeOf(""))
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		ret.Id = value.(string)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse1(v *GetPostResponse1) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse1Argument(v *Org_Federation_GetPostResponse1Argument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse2(v *GetPostResponse2) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("post", s.logvalue_Org_Federation_Post(v.GetPost())),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostResponse2Argument(v *Org_Federation_GetPostResponse2Argument[*FederationServiceDependentClientSet]) slog.Value {
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
