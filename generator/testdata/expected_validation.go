// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
//
// source: validation.proto
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
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CustomMessageArgument is argument for "org.federation.CustomMessage" message.
type FederationService_Org_Federation_CustomMessageArgument struct {
	Message string
}

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type FederationService_Org_Federation_GetPostResponseArgument struct {
	Id                  string
	Post                *Post
	XDef2ErrDetail0Msg0 *CustomMessage
	XDef2ErrDetail0Msg1 *CustomMessage
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
}

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler grpcfed.ErrorHandler
	// Logger sets the logger used to output Debug/Info/Error information.
	Logger *slog.Logger
}

// FederationServiceClientFactory provides a factory that creates the gRPC Client needed to invoke methods of the gRPC Service on which the Federation Service depends.
type FederationServiceClientFactory interface {
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
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	celTypeHelperFieldMap := grpcfed.CELTypeHelperFieldMap{
		"grpc.federation.private.org.federation.CustomMessageArgument": {
			"message": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Message"),
		},
		"grpc.federation.private.org.federation.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.org.federation.PostArgument": {},
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
		client:        &FederationServiceDependentClientSet{},
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
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.CustomMessageArgument", req)}

	// create a message value to be returned.
	ret := &CustomMessage{}

	// field binding section.
	// (grpc.federation.field).by = "$.message"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `$.message`,
		CacheIndex: 1,
		Setter: func(v string) error {
			ret.Message = v
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
			Post                *Post
			XDef1               bool
			XDef2               bool
			XDef2ErrDetail0Msg0 *CustomMessage
			XDef2ErrDetail0Msg1 *CustomMessage
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.GetPostResponseArgument", req)}
	/*
		def {
		  name: "post"
		  message {
		    name: "Post"
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
				ret, err := s.resolve_Org_Federation_Post(ctx, args)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		})
	}

	/*
		def {
		  name: "_def1"
		  validation {
		    error {
		      code: FAILED_PRECONDITION
		      if: "post.id != 'some-id'"
		      message: "'validation message 1'"
		    }
		  }
		}
	*/
	def__def1 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[bool, *localValueType]{
			Name: `_def1`,
			Type: grpcfed.CELBoolType,
			Setter: func(value *localValueType, v bool) error {
				value.vars.XDef1 = v
				return nil
			},
			Validation: func(ctx context.Context, value *localValueType) error {
				var stat *grpcfed.Status
				if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
					Value:      value,
					Expr:       `post.id != 'some-id'`,
					CacheIndex: 2,
					Body: func(value *localValueType) error {
						errmsg, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
							Value:      value,
							Expr:       `'validation message 1'`,
							OutType:    reflect.TypeOf(""),
							CacheIndex: 3,
						})
						if err != nil {
							return err
						}
						errorMessage := errmsg.(string)
						stat = grpcfed.NewGRPCStatus(grpcfed.FailedPreconditionCode, errorMessage)
						return nil
					},
				}); err != nil {
					return err
				}
				return grpcfed.NewErrorWithLogAttrs(stat.Err(), slog.LevelError, grpcfed.LogAttrs(ctx))
			},
		})
	}

	/*
		def {
		  name: "_def2"
		  validation {
		    error {
		      code: FAILED_PRECONDITION
		      if: "true"
		      message: "'validation message 2'"
		      details {
		        if: "post.title != 'some-title'"
		        message: [
		          {...},
		          {...}
		        ]
		        precondition_failure {...}
		        bad_request {...}
		        localized_message {...}
		      }
		    }
		  }
		}
	*/
	def__def2 := func(ctx context.Context) error {
		return grpcfed.EvalDef(ctx, value, grpcfed.Def[bool, *localValueType]{
			Name: `_def2`,
			Type: grpcfed.CELBoolType,
			Setter: func(value *localValueType, v bool) error {
				value.vars.XDef2 = v
				return nil
			},
			Validation: func(ctx context.Context, value *localValueType) error {
				var stat *grpcfed.Status
				if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
					Value:      value,
					Expr:       `true`,
					CacheIndex: 4,
					Body: func(value *localValueType) error {
						errmsg, err := grpcfed.EvalCEL(ctx, &grpcfed.EvalCELRequest{
							Value:      value,
							Expr:       `'validation message 2'`,
							OutType:    reflect.TypeOf(""),
							CacheIndex: 5,
						})
						if err != nil {
							return err
						}
						errorMessage := errmsg.(string)
						var details []grpcfed.ProtoMessage
						if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
							Value:      value,
							Expr:       `post.title != 'some-title'`,
							CacheIndex: 6,
							Body: func(value *localValueType) error {
								if _, err := func() (any, error) {
									/*
										def {
										  name: "_def2_err_detail0_msg0"
										  message {
										    name: "CustomMessage"
										    args { name: "message", by: "'message1'" }
										  }
										}
									*/
									def__def2_err_detail0_msg0 := func(ctx context.Context) error {
										return grpcfed.EvalDef(ctx, value, grpcfed.Def[*CustomMessage, *localValueType]{
											Name: `_def2_err_detail0_msg0`,
											Type: grpcfed.CELObjectType("org.federation.CustomMessage"),
											Setter: func(value *localValueType, v *CustomMessage) error {
												value.vars.XDef2ErrDetail0Msg0 = v
												return nil
											},
											Message: func(ctx context.Context, value *localValueType) (any, error) {
												args := &FederationService_Org_Federation_CustomMessageArgument{}
												// { name: "message", by: "'message1'" }
												if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
													Value:      value,
													Expr:       `'message1'`,
													CacheIndex: 7,
													Setter: func(v string) error {
														args.Message = v
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

									/*
										def {
										  name: "_def2_err_detail0_msg1"
										  message {
										    name: "CustomMessage"
										    args { name: "message", by: "'message2'" }
										  }
										}
									*/
									def__def2_err_detail0_msg1 := func(ctx context.Context) error {
										return grpcfed.EvalDef(ctx, value, grpcfed.Def[*CustomMessage, *localValueType]{
											Name: `_def2_err_detail0_msg1`,
											Type: grpcfed.CELObjectType("org.federation.CustomMessage"),
											Setter: func(value *localValueType, v *CustomMessage) error {
												value.vars.XDef2ErrDetail0Msg1 = v
												return nil
											},
											Message: func(ctx context.Context, value *localValueType) (any, error) {
												args := &FederationService_Org_Federation_CustomMessageArgument{}
												// { name: "message", by: "'message2'" }
												if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
													Value:      value,
													Expr:       `'message2'`,
													CacheIndex: 8,
													Setter: func(v string) error {
														args.Message = v
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

									// A tree view of message dependencies is shown below.
									/*
									   _def2_err_detail0_msg0 ─┐
									   _def2_err_detail0_msg1 ─┤
									*/
									eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

									grpcfed.GoWithRecover(eg, func() (any, error) {
										if err := def__def2_err_detail0_msg0(ctx1); err != nil {
											grpcfed.RecordErrorToSpan(ctx1, err)
											return nil, err
										}
										return nil, nil
									})

									grpcfed.GoWithRecover(eg, func() (any, error) {
										if err := def__def2_err_detail0_msg1(ctx1); err != nil {
											grpcfed.RecordErrorToSpan(ctx1, err)
											return nil, err
										}
										return nil, nil
									})

									if err := eg.Wait(); err != nil {
										return nil, err
									}
									return nil, nil
								}(); err != nil {
									return err
								}
								if detail := grpcfed.CustomMessage(ctx, &grpcfed.CustomMessageParam{
									Value:            value,
									MessageValueName: "_def2_err_detail0_msg0",
									CacheIndex:       9,
									MessageIndex:     0,
								}); detail != nil {
									details = append(details, detail)
								}
								if detail := grpcfed.CustomMessage(ctx, &grpcfed.CustomMessageParam{
									Value:            value,
									MessageValueName: "_def2_err_detail0_msg1",
									CacheIndex:       10,
									MessageIndex:     1,
								}); detail != nil {
									details = append(details, detail)
								}
								if detail := grpcfed.PreconditionFailure(ctx, value, []*grpcfed.PreconditionFailureViolation{
									{
										Type:              `'some-type'`,
										Subject:           `'some-subject'`,
										Desc:              `'some-description'`,
										TypeCacheIndex:    11,
										SubjectCacheIndex: 12,
										DescCacheIndex:    13,
									},
								}); detail != nil {
									details = append(details, detail)
								}
								if detail := grpcfed.BadRequest(ctx, value, []*grpcfed.BadRequestFieldViolation{
									{
										Field:           `'some-field'`,
										Desc:            `'some-description'`,
										FieldCacheIndex: 14,
										DescCacheIndex:  15,
									},
								}); detail != nil {
									details = append(details, detail)
								}
								if detail := grpcfed.LocalizedMessage(ctx, &grpcfed.LocalizedMessageParam{
									Value:      value,
									Locale:     "en-US",
									Message:    `'some-message'`,
									CacheIndex: 16,
								}); detail != nil {
									details = append(details, detail)
								}
								return nil
							},
						}); err != nil {
							return err
						}
						status := grpcfed.NewGRPCStatus(grpcfed.FailedPreconditionCode, errorMessage)
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
					return err
				}
				return grpcfed.NewErrorWithLogAttrs(stat.Err(), slog.LevelWarn, grpcfed.LogAttrs(ctx))
			},
		})
	}

	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)
	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_post(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		if err := def__def1(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})
	grpcfed.GoWithRecover(eg, func() (any, error) {
		if err := def_post(ctx1); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}
		return nil, nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if err := def__def2(ctx); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Post = value.vars.Post
	req.XDef2ErrDetail0Msg0 = value.vars.XDef2ErrDetail0Msg0
	req.XDef2ErrDetail0Msg1 = value.vars.XDef2ErrDetail0Msg1

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*Post]{
		Value:      value,
		Expr:       `post`,
		CacheIndex: 17,
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
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.org.federation.PostArgument", req)}

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	// (grpc.federation.field).by = "'some-id'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'some-id'`,
		CacheIndex: 18,
		Setter: func(v string) error {
			ret.Id = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "'some-title'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'some-title'`,
		CacheIndex: 19,
		Setter: func(v string) error {
			ret.Title = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "'some-content'"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[string]{
		Value:      value,
		Expr:       `'some-content'`,
		CacheIndex: 20,
		Setter: func(v string) error {
			ret.Content = v
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_CustomMessage(v *CustomMessage) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("message", v.GetMessage()),
	)
}

func (s *FederationService) logvalue_Org_Federation_CustomMessageArgument(v *FederationService_Org_Federation_CustomMessageArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("message", v.Message),
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

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.String("title", v.GetTitle()),
		slog.String("content", v.GetContent()),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *FederationService_Org_Federation_PostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}
