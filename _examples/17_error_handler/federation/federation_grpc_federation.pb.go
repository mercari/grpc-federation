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
	code "google.golang.org/genproto/googleapis/rpc/code"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CustomMessageArgument is argument for "org.federation.CustomMessage" message.
type FederationService_Org_Federation_CustomMessageArgument struct {
	ErrorInfo *grpcfedcel.Error
}

// Org_Federation_GetPost2ResponseArgument is argument for "org.federation.GetPost2Response" message.
type FederationService_Org_Federation_GetPost2ResponseArgument struct {
	Code code.Code
	Id   string
}

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type FederationService_Org_Federation_GetPostResponseArgument struct {
	Id   string
	Post *Post
}

// Org_Federation_LocalizedMessageArgument is argument for "org.federation.LocalizedMessage" message.
type FederationService_Org_Federation_LocalizedMessageArgument struct {
	Value string
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
	Id                  string
	LocalizedMsg        *LocalizedMessage
	Post                *post.Post
	Res                 *post.GetPostResponse
	XDef0ErrDetail0Msg0 *CustomMessage
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
)

// FederationService represents Federation Service.
type FederationService struct {
	UnimplementedFederationServiceServer
	cfg                FederationServiceConfig
	logger             *slog.Logger
	errorHandler       grpcfed.ErrorHandler
	celCacheMap        *grpcfed.CELCacheMap
	tracer             trace.Tracer
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
	Post_PostServiceClient, err := cfg.Client.Post_PostServiceClient(FederationServiceClientConfig{
		Service: "post.PostService",
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
		"grpc.federation.private.CustomMessageArgument": {
			"error_info": grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("grpc.federation.private.Error"), "ErrorInfo"),
		},
		"grpc.federation.private.GetPost2ResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.LocalizedMessageArgument": {
			"value": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Value"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
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

// GetPost2 implements "org.federation.FederationService/GetPost2" method.
func (s *FederationService) GetPost2(ctx context.Context, req *GetPost2Request) (res *GetPost2Response, e error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.FederationService/GetPost2")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, s.logger)
	ctx = grpcfed.WithCELCacheMap(ctx, s.celCacheMap)
	defer func() {
		if r := recover(); r != nil {
			e = grpcfed.RecoverError(r, grpcfed.StackTrace())
			grpcfed.OutputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetPost2Response(ctx, &FederationService_Org_Federation_GetPost2ResponseArgument{
		Id: req.GetId(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_CustomMessage resolve "org.federation.CustomMessage" message.
func (s *FederationService) resolve_Org_Federation_CustomMessage(ctx context.Context, req *FederationService_Org_Federation_CustomMessageArgument) (*CustomMessage, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CustomMessage")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.CustomMessage", slog.Any("message_args", s.logvalue_Org_Federation_CustomMessageArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.CustomMessageArgument", req)}

	// create a message value to be returned.
	ret := &CustomMessage{}

	// field binding section.
	// (grpc.federation.field).by = "'custom error message:' + $.error_info.message"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'custom error message:' + $.error_info.message`,
		CacheIndex: 1,
		Setter: func(v string) error {
			ret.Msg = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.CustomMessage", slog.Any("org.federation.CustomMessage", s.logvalue_Org_Federation_CustomMessage(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetPost2Response resolve "org.federation.GetPost2Response" message.
func (s *FederationService) resolve_Org_Federation_GetPost2Response(ctx context.Context, req *FederationService_Org_Federation_GetPost2ResponseArgument) (*GetPost2Response, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPost2Response")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPost2Response", slog.Any("message_args", s.logvalue_Org_Federation_GetPost2ResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			Code  code.Code
			XDef0 *post.GetPostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.GetPost2ResponseArgument", req)}
	/*
		def {
		  name: "_def0"
		  call {
		    method: "post.PostService/GetPost"
		    request { field: "id", by: "$.id" }
		  }
		}
	*/
	def__def0 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
			Name: `_def0`,
			Type: grpcfed.CELObjectType("post.GetPostResponse"),
			Setter: func(value *localValueType, v *post.GetPostResponse) error {
				value.vars.XDef0 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &post.GetPostRequest{}
				// { field: "id", by: "$.id" }
				if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
					Value:      value,
					Expr:       `$.id`,
					CacheIndex: 2,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call post.PostService/GetPost", slog.Any("post.GetPostRequest", s.logvalue_Post_GetPostRequest(args)))
				ret, err := s.client.Post_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					grpcfed.SetGRPCError(ctx, value, err)
					var (
						defaultMsg     string
						defaultCode    grpcfed.Code
						defaultDetails []grpcfed.ProtoMessage
					)
					if stat, exists := grpcfed.GRPCStatusFromError(err); exists {
						defaultMsg = stat.Message()
						defaultCode = stat.Code()
						details := stat.Details()
						defaultDetails = make([]grpcfed.ProtoMessage, 0, len(details))
						for _, detail := range details {
							msg, ok := detail.(grpcfed.ProtoMessage)
							if ok {
								defaultDetails = append(defaultDetails, msg)
							}
						}
						_ = defaultMsg
						_ = defaultCode
						_ = defaultDetails
					}

					type localStatusType struct {
						status   *grpcfed.Status
						logLevel slog.Level
					}
					stat, handleErr := func() (*localStatusType, error) {
						var stat *grpcfed.Status
						/*
							def {
							  name: "code"
							  by: "google.rpc.Code.from(error.code)"
							}
						*/
						def_code := func(ctx context.Context) error {
							return grpcfed.EvalDef(ctx, value, grpcfed.Def[code.Code, *localValueType]{
								Name: `code`,
								Type: grpcfed.CELIntType,
								Setter: func(value *localValueType, v code.Code) error {
									value.vars.Code = v
									return nil
								},
								By:           `google.rpc.Code.from(error.code)`,
								ByCacheIndex: 3,
							})
						}

						if err := def_code(ctx); err != nil {
							grpcfed.RecordErrorToSpan(ctx, err)
							return nil, err
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `code == google.rpc.Code.FAILED_PRECONDITION`,
							CacheIndex: 4,
							Body: func(value *localValueType) error {
								var errorMessage string
								if defaultMsg != "" {
									errorMessage = defaultMsg
								} else {
									errorMessage = "error"
								}

								var code grpcfed.Code
								code = defaultCode
								status := grpcfed.NewGRPCStatus(code, errorMessage)
								statusWithDetails, err := status.WithDetails(defaultDetails...)
								if err != nil {
									grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
									stat = status
								} else {
									stat = statusWithDetails
								}
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelWarn}, nil
						}
						return nil, nil
					}()
					if handleErr != nil {
						grpcfed.Logger(ctx).ErrorContext(ctx, "failed to handle error", slog.String("error", handleErr.Error()))
						// If it fails during error handling, return the original error.
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
						}
					} else if stat != nil {
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, stat.status.Err()); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, stat.logLevel, grpcfed.LogAttrs(ctx))
						}
					} else {
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
						}
					}
				}
				return ret, nil
			},
		})
	}

	if err := def__def0(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Code = value.vars.Code

	// create a message value to be returned.
	ret := &GetPost2Response{}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.GetPost2Response", slog.Any("org.federation.GetPost2Response", s.logvalue_Org_Federation_GetPost2Response(ret)))
	return ret, nil
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
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.GetPostResponseArgument", req)}
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
					CacheIndex: 5,
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
		CacheIndex: 6,
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

// resolve_Org_Federation_LocalizedMessage resolve "org.federation.LocalizedMessage" message.
func (s *FederationService) resolve_Org_Federation_LocalizedMessage(ctx context.Context, req *FederationService_Org_Federation_LocalizedMessageArgument) (*LocalizedMessage, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.LocalizedMessage")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.LocalizedMessage", slog.Any("message_args", s.logvalue_Org_Federation_LocalizedMessageArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.LocalizedMessageArgument", req)}

	// create a message value to be returned.
	ret := &LocalizedMessage{}

	// field binding section.
	// (grpc.federation.field).by = "'localized value:' + $.value"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'localized value:' + $.value`,
		CacheIndex: 7,
		Setter: func(v string) error {
			ret.Value = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.LocalizedMessage", slog.Any("org.federation.LocalizedMessage", s.logvalue_Org_Federation_LocalizedMessage(ret)))
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
			Id                  string
			LocalizedMsg        *LocalizedMessage
			Post                *post.Post
			Res                 *post.GetPostResponse
			XDef0ErrDetail0Msg0 *CustomMessage
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.PostArgument", req)}
	/*
		def {
		  name: "res"
		  call {
		    method: "post.PostService/GetPost"
		    request { field: "id", by: "$.id" }
		  }
		}
	*/
	def_res := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
			Name: `res`,
			Type: grpcfed.CELObjectType("post.GetPostResponse"),
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
					CacheIndex: 8,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call post.PostService/GetPost", slog.Any("post.GetPostRequest", s.logvalue_Post_GetPostRequest(args)))
				ret, err := s.client.Post_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					grpcfed.SetGRPCError(ctx, value, err)
					var (
						defaultMsg     string
						defaultCode    grpcfed.Code
						defaultDetails []grpcfed.ProtoMessage
					)
					if stat, exists := grpcfed.GRPCStatusFromError(err); exists {
						defaultMsg = stat.Message()
						defaultCode = stat.Code()
						details := stat.Details()
						defaultDetails = make([]grpcfed.ProtoMessage, 0, len(details))
						for _, detail := range details {
							msg, ok := detail.(grpcfed.ProtoMessage)
							if ok {
								defaultDetails = append(defaultDetails, msg)
							}
						}
						_ = defaultMsg
						_ = defaultCode
						_ = defaultDetails
					}

					type localStatusType struct {
						status   *grpcfed.Status
						logLevel slog.Level
					}
					stat, handleErr := func() (*localStatusType, error) {
						var stat *grpcfed.Status
						/*
							def {
							  name: "id"
							  by: "$.id"
							}
						*/
						def_id := func(ctx context.Context) error {
							return grpcfed.EvalDef(ctx, value, grpcfed.Def[string, *localValueType]{
								Name: `id`,
								Type: grpcfed.CELStringType,
								Setter: func(value *localValueType, v string) error {
									value.vars.Id = v
									return nil
								},
								By:           `$.id`,
								ByCacheIndex: 9,
							})
						}

						if err := def_id(ctx); err != nil {
							grpcfed.RecordErrorToSpan(ctx, err)
							return nil, err
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `error.precondition_failures[?0].violations[?0].subject == optional.of('bar') && error.localized_messages[?0].message == optional.of('hello') && error.custom_messages[?0].id == optional.of('xxx')`,
							CacheIndex: 10,
							Body: func(value *localValueType) error {
								errmsg, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
									Value:      value,
									Expr:       `'this is custom error message'`,
									OutType:    reflect.TypeOf(""),
									CacheIndex: 11,
								})
								if err != nil {
									return err
								}
								errorMessage := errmsg.(string)
								var details []grpcfed.ProtoMessage
								if _, err := func() (any, error) {
									/*
										def {
										  name: "localized_msg"
										  message {
										    name: "LocalizedMessage"
										    args { name: "value", by: "id" }
										  }
										}
									*/
									def_localized_msg := func(ctx context.Context) error {
										return grpcfed.EvalDef(ctx, value, grpcfed.Def[*LocalizedMessage, *localValueType]{
											Name: `localized_msg`,
											Type: grpcfed.CELObjectType("org.federation.LocalizedMessage"),
											Setter: func(value *localValueType, v *LocalizedMessage) error {
												value.vars.LocalizedMsg = v
												return nil
											},
											Message: func(ctx context.Context, value *localValueType) (any, error) {
												args := &FederationService_Org_Federation_LocalizedMessageArgument{}
												// { name: "value", by: "id" }
												if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
													Value:      value,
													Expr:       `id`,
													CacheIndex: 12,
													Setter: func(v string) error {
														args.Value = v
														return nil
													},
												}); err != nil {
													return nil, err
												}
												ret, err := s.resolve_Org_Federation_LocalizedMessage(ctx, args)
												if err != nil {
													return nil, err
												}
												return ret, nil
											},
										})
									}

									if err := def_localized_msg(ctx); err != nil {
										grpcfed.RecordErrorToSpan(ctx, err)
										return nil, err
									}
									return nil, nil
								}(); err != nil {
									return err
								}
								if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
									Value:      value,
									Expr:       `true`,
									CacheIndex: 13,
									Body: func(value *localValueType) error {
										if _, err := func() (any, error) {
											/*
												def {
												  name: "_def0_err_detail0_msg0"
												  message {
												    name: "CustomMessage"
												    args { name: "error_info", by: "error" }
												  }
												}
											*/
											def__def0_err_detail0_msg0 := func(ctx context.Context) error {
												return grpcfed.EvalDef(ctx, value, grpcfed.Def[*CustomMessage, *localValueType]{
													Name: `_def0_err_detail0_msg0`,
													Type: grpcfed.CELObjectType("org.federation.CustomMessage"),
													Setter: func(value *localValueType, v *CustomMessage) error {
														value.vars.XDef0ErrDetail0Msg0 = v
														return nil
													},
													Message: func(ctx context.Context, value *localValueType) (any, error) {
														args := &FederationService_Org_Federation_CustomMessageArgument{}
														// { name: "error_info", by: "error" }
														if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*grpcfedcel.Error]{
															Value:      value,
															Expr:       `error`,
															CacheIndex: 14,
															Setter: func(v *grpcfedcel.Error) error {
																args.ErrorInfo = v
																return nil
															},
														}); err != nil {
															return nil, err
														}
														ret, err := s.resolve_Org_Federation_CustomMessage(ctx, args)
														if err != nil {
															return nil, err
														}
														return ret, nil
													},
												})
											}

											if err := def__def0_err_detail0_msg0(ctx); err != nil {
												grpcfed.RecordErrorToSpan(ctx, err)
												return nil, err
											}
											return nil, nil
										}(); err != nil {
											return err
										}
										if detail := grpcfed.CustomMessage(ctx, &grpcfed.CustomMessageParam{
											Value:            value,
											MessageValueName: "_def0_err_detail0_msg0",
											CacheIndex:       15,
											MessageIndex:     0,
										}); detail != nil {
											details = append(details, detail)
										}
										{
											detail, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
												Value:      value,
												Expr:       `post.Post{id: 'foo'}`,
												OutType:    reflect.TypeOf((*post.Post)(nil)),
												CacheIndex: 16,
											})
											if err != nil {
												grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
											}
											if detail != nil {
												details = append(details, detail.(grpcfed.ProtoMessage))
											}
										}
										if detail := grpcfed.PreconditionFailure(ctx, value, []*grpcfed.PreconditionFailureViolation{
											{
												Type:              `'some-type'`,
												Subject:           `'some-subject'`,
												Desc:              `'some-description'`,
												TypeCacheIndex:    17,
												SubjectCacheIndex: 18,
												DescCacheIndex:    19,
											},
										}); detail != nil {
											details = append(details, detail)
										}
										if detail := grpcfed.LocalizedMessage(ctx, &grpcfed.LocalizedMessageParam{
											Value:      value,
											Locale:     "en-US",
											Message:    `localized_msg.value`,
											CacheIndex: 20,
										}); detail != nil {
											details = append(details, detail)
										}
										return nil
									},
								}); err != nil {
									return err
								}

								var code grpcfed.Code
								code = grpcfed.FailedPreconditionCode
								status := grpcfed.NewGRPCStatus(code, errorMessage)
								statusWithDetails, err := status.WithDetails(details...)
								if err != nil {
									grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
									stat = status
								} else {
									stat = statusWithDetails
								}
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelError}, nil
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `error.code == google.rpc.Code.INVALID_ARGUMENT`,
							CacheIndex: 21,
							Body: func(value *localValueType) error {
								errmsg, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
									Value:      value,
									Expr:       `'this is custom log level'`,
									OutType:    reflect.TypeOf(""),
									CacheIndex: 22,
								})
								if err != nil {
									return err
								}
								errorMessage := errmsg.(string)

								var code grpcfed.Code
								code = grpcfed.InvalidArgumentCode
								status := grpcfed.NewGRPCStatus(code, errorMessage)
								statusWithDetails, err := status.WithDetails(defaultDetails...)
								if err != nil {
									grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
									stat = status
								} else {
									stat = statusWithDetails
								}
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelWarn}, nil
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `error.code == google.rpc.Code.UNIMPLEMENTED`,
							CacheIndex: 23,
							Body: func(value *localValueType) error {
								stat = grpcfed.NewGRPCStatus(grpcfed.OKCode, "ignore error")
								if err := grpcfed.IgnoreAndResponse(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
									Name: "res",
									Type: grpcfed.CELObjectType("post.GetPostResponse"),
									Setter: func(value *localValueType, v *post.GetPostResponse) error {
										ret = v // assign customized response to the result value.
										return nil
									},
									By:           `post.GetPostResponse{post: post.Post{id: 'anonymous'}}`,
									ByCacheIndex: 24,
								}); err != nil {
									grpcfed.Logger(ctx).ErrorContext(ctx, "failed to set response when ignored", slog.String("error", err.Error()))
									return nil
								}
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelError}, nil
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `true`,
							CacheIndex: 25,
							Body: func(value *localValueType) error {
								stat = grpcfed.NewGRPCStatus(grpcfed.OKCode, "ignore error")
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelError}, nil
						}
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `true`,
							CacheIndex: 26,
							Body: func(value *localValueType) error {
								var errorMessage string
								if defaultMsg != "" {
									errorMessage = defaultMsg
								} else {
									errorMessage = "error"
								}

								var code grpcfed.Code
								code = grpcfed.CancelledCode
								status := grpcfed.NewGRPCStatus(code, errorMessage)
								statusWithDetails, err := status.WithDetails(defaultDetails...)
								if err != nil {
									grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
									stat = status
								} else {
									stat = statusWithDetails
								}
								return nil
							},
						}); err != nil {
							return nil, err
						}
						if stat != nil {
							return &localStatusType{status: stat, logLevel: slog.LevelError}, nil
						}
						return nil, nil
					}()
					if handleErr != nil {
						grpcfed.Logger(ctx).ErrorContext(ctx, "failed to handle error", slog.String("error", handleErr.Error()))
						// If it fails during error handling, return the original error.
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
						}
					} else if stat != nil {
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, stat.status.Err()); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, stat.logLevel, grpcfed.LogAttrs(ctx))
						}
					} else {
						if err := s.errorHandler(ctx, FederationService_DependentMethod_Post_PostService_GetPost, err); err != nil {
							return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
						}
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
			Type: grpcfed.CELObjectType("post.Post"),
			Setter: func(value *localValueType, v *post.Post) error {
				value.vars.Post = v
				return nil
			},
			By:           `res.post`,
			ByCacheIndex: 27,
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

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Id = value.vars.Id
	req.LocalizedMsg = value.vars.LocalizedMsg
	req.Post = value.vars.Post
	req.Res = value.vars.Res
	req.XDef0ErrDetail0Msg0 = value.vars.XDef0ErrDetail0Msg0

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.Post.GetId()       // { name: "post", autobind: true }
	ret.Title = value.vars.Post.GetTitle() // { name: "post", autobind: true }

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Grpc_Federation_Private_Error(v *grpcfedcel.Error) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_CustomMessage(v *CustomMessage) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("msg", v.GetMsg()),
	)
}

func (s *FederationService) logvalue_Org_Federation_CustomMessageArgument(v *FederationService_Org_Federation_CustomMessageArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.Any("error_info", s.logvalue_Grpc_Federation_Private_Error(v.ErrorInfo)),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPost2Response(v *GetPost2Response) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_GetPost2ResponseArgument(v *FederationService_Org_Federation_GetPost2ResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
	)
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

func (s *FederationService) logvalue_Org_Federation_LocalizedMessage(v *LocalizedMessage) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_LocalizedMessageArgument(v *FederationService_Org_Federation_LocalizedMessageArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.Value),
	)
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
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

func (s *FederationService) logvalue_Post_GetPostRequest(v *post.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}
