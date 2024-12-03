// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: error_handler.proto
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
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CustomMessageArgument is argument for "org.federation.CustomMessage" message.
type FederationService_Org_Federation_CustomMessageArgument struct {
	Msg string
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
	// Org_Post_PostServiceClient create a gRPC Client to be used to call methods in org.post.PostService.
	Org_Post_PostServiceClient(FederationServiceClientConfig) (post.PostServiceClient, error)
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
	FederationService_DependentMethod_Org_Post_PostService_GetPost = "/org.post.PostService/GetPost"
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
	Org_Post_PostServiceClient, err := cfg.Client.Org_Post_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.PostService",
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
			"msg": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Msg"),
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
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostType", post.PostType_value, post.PostType_name)...)
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		client: &FederationServiceDependentClientSet{
			Org_Post_PostServiceClient: Org_Post_PostServiceClient,
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
	// (grpc.federation.field).by = "'custom error message:' + $.msg"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'custom error message:' + $.msg`,
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

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *FederationService_Org_Federation_GetPostResponseArgument) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse")
	defer span.End()
	ctx = grpcfed.WithLogger(ctx, grpcfed.Logger(ctx), grpcfed.LogAttrs(ctx)...)

	grpcfed.Logger(ctx).DebugContext(ctx, "resolve org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			post *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.GetPostResponseArgument", req)}

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
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*Post, *localValueType]{
		Name: `post`,
		Type: grpcfed.CELObjectType("org.federation.Post"),
		Setter: func(value *localValueType, v *Post) error {
			value.vars.post = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &FederationService_Org_Federation_PostArgument{}
			// { name: "id", by: "$.id" }
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
			ret, err := s.resolve_Org_Federation_Post(ctx, args)
			if err != nil {
				return nil, err
			}
			return ret, nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.post

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:      value,
		Expr:       `post`,
		CacheIndex: 3,
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
		CacheIndex: 4,
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
			_def0_err_detail0_msg0 *CustomMessage
			id                     string
			localized_msg          *LocalizedMessage
			post                   *post.Post
			res                    *post.GetPostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.PostArgument", req)}

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
	if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
		Name: `res`,
		Type: grpcfed.CELObjectType("org.post.GetPostResponse"),
		Setter: func(value *localValueType, v *post.GetPostResponse) error {
			value.vars.res = v
			return nil
		},
		Message: func(ctx context.Context, value *localValueType) (any, error) {
			args := &post.GetPostRequest{}
			// { field: "id", by: "$.id" }
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
			grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/GetPost", slog.Any("org.post.GetPostRequest", s.logvalue_Org_Post_GetPostRequest(args)))
			ret, err := s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
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

					// This section's codes are generated by the following proto definition.
					/*
					   def {
					     name: "id"
					     by: "$.id"
					   }
					*/
					if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[string, *localValueType]{
						Name: `id`,
						Type: grpcfed.CELStringType,
						Setter: func(value *localValueType, v string) error {
							value.vars.id = v
							return nil
						},
						By:           `$.id`,
						ByCacheIndex: 6,
					}); err != nil {
						grpcfed.RecordErrorToSpan(ctx, err)
						return nil, err
					}
					if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
						Value:      value,
						Expr:       `error.precondition_failures.map(f, f.violations[0]).first(v, v.subject == '').?subject == optional.of('')`,
						CacheIndex: 7,
						Body: func(value *localValueType) error {
							errmsg, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
								Value:      value,
								Expr:       `'id must be not empty'`,
								OutType:    reflect.TypeOf(""),
								CacheIndex: 8,
							})
							if err != nil {
								return err
							}
							errorMessage := errmsg.(string)
							var details []grpcfed.ProtoMessage
							if _, err := func() (any, error) {

								// This section's codes are generated by the following proto definition.
								/*
								   def {
								     name: "localized_msg"
								     message {
								       name: "LocalizedMessage"
								       args { name: "value", by: "id" }
								     }
								   }
								*/
								if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*LocalizedMessage, *localValueType]{
									Name: `localized_msg`,
									Type: grpcfed.CELObjectType("org.federation.LocalizedMessage"),
									Setter: func(value *localValueType, v *LocalizedMessage) error {
										value.vars.localized_msg = v
										return nil
									},
									Message: func(ctx context.Context, value *localValueType) (any, error) {
										args := &FederationService_Org_Federation_LocalizedMessageArgument{}
										// { name: "value", by: "id" }
										if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
											Value:      value,
											Expr:       `id`,
											CacheIndex: 9,
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
								}); err != nil {
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
								CacheIndex: 10,
								Body: func(value *localValueType) error {
									if _, err := func() (any, error) {

										// This section's codes are generated by the following proto definition.
										/*
										   def {
										     name: "_def0_err_detail0_msg0"
										     message {
										       name: "CustomMessage"
										       args { name: "msg", by: "id" }
										     }
										   }
										*/
										if err := grpcfed.EvalDef(ctx, value, grpcfed.Def[*CustomMessage, *localValueType]{
											Name: `_def0_err_detail0_msg0`,
											Type: grpcfed.CELObjectType("org.federation.CustomMessage"),
											Setter: func(value *localValueType, v *CustomMessage) error {
												value.vars._def0_err_detail0_msg0 = v
												return nil
											},
											Message: func(ctx context.Context, value *localValueType) (any, error) {
												args := &FederationService_Org_Federation_CustomMessageArgument{}
												// { name: "msg", by: "id" }
												if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
													Value:      value,
													Expr:       `id`,
													CacheIndex: 11,
													Setter: func(v string) error {
														args.Msg = v
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
										}); err != nil {
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
										CacheIndex:       12,
										MessageIndex:     0,
									}); detail != nil {
										details = append(details, detail)
									}
									{
										detail, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
											Value:      value,
											Expr:       `org.post.Post{id: 'foo'}`,
											OutType:    reflect.TypeOf((*post.Post)(nil)),
											CacheIndex: 13,
										})
										if err != nil {
											grpcfed.Logger(ctx).ErrorContext(ctx, "failed setting error details", slog.String("error", err.Error()))
										}
										if detail != nil {
											details = append(details, detail.(grpcfed.ProtoMessage))
										}
									}
									{
										detail, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
											Value:      value,
											Expr:       `org.post.CreatePost{title: 'bar'}`,
											OutType:    reflect.TypeOf((*post.CreatePost)(nil)),
											CacheIndex: 14,
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
											TypeCacheIndex:    15,
											SubjectCacheIndex: 16,
											DescCacheIndex:    17,
										},
									}); detail != nil {
										details = append(details, detail)
									}
									if detail := grpcfed.LocalizedMessage(ctx, &grpcfed.LocalizedMessageParam{
										Value:      value,
										Locale:     "en-US",
										Message:    `localized_msg.value`,
										CacheIndex: 18,
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
						Expr:       `error.code == google.rpc.Code.UNIMPLEMENTED`,
						CacheIndex: 19,
						Body: func(value *localValueType) error {
							stat = grpcfed.NewGRPCStatus(grpcfed.OKCode, "ignore error")
							if err := grpcfed.IgnoreAndResponse(ctx, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
								Name: "res",
								Type: grpcfed.CELObjectType("org.post.GetPostResponse"),
								Setter: func(value *localValueType, v *post.GetPostResponse) error {
									ret = v // assign customized response to the result value.
									return nil
								},
								By:           `org.post.GetPostResponse{post: org.post.Post{id: 'anonymous', title: 'none'}}`,
								ByCacheIndex: 20,
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
						CacheIndex: 21,
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
					return nil, nil
				}()
				if handleErr != nil {
					grpcfed.Logger(ctx).ErrorContext(ctx, "failed to handle error", slog.String("error", handleErr.Error()))
					// If it fails during error handling, return the original error.
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				} else if stat != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, stat.status.Err()); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, stat.logLevel, grpcfed.LogAttrs(ctx))
					}
				} else {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
			}
			return ret, nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
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
		Name: `post`,
		Type: grpcfed.CELObjectType("org.post.Post"),
		Setter: func(value *localValueType, v *post.Post) error {
			value.vars.post = v
			return nil
		},
		By:           `res.post`,
		ByCacheIndex: 22,
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Id = value.vars.id
	req.LocalizedMsg = value.vars.localized_msg
	req.Post = value.vars.post
	req.Res = value.vars.res
	req.XDef0ErrDetail0Msg0 = value.vars._def0_err_detail0_msg0

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.post.GetId()       // { name: "post", autobind: true }
	ret.Title = value.vars.post.GetTitle() // { name: "post", autobind: true }

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
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
		slog.String("msg", v.Msg),
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
