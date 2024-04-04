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
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_CustomMessageArgument is argument for "org.federation.CustomMessage" message.
type Org_Federation_CustomMessageArgument struct {
	Message string
}

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type Org_Federation_GetPostResponseArgument struct {
	Id                  string
	Post                *Post
	XDef2ErrDetail0Msg0 *CustomMessage
	XDef2ErrDetail0Msg1 *CustomMessage
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type Org_Federation_PostArgument struct {
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
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	errorHandler := cfg.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, methodName string, err error) error { return err }
	}
	celHelper := grpcfed.NewCELTypeHelper(map[string]map[string]*grpcfed.CELFieldType{
		"grpc.federation.private.CustomMessageArgument": {
			"message": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Message"),
		},
		"grpc.federation.private.GetPostResponseArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
		},
		"grpc.federation.private.PostArgument": {},
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
		client:       &FederationServiceDependentClientSet{},
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
	res, err := s.resolve_Org_Federation_GetPostResponse(ctx, &Org_Federation_GetPostResponseArgument{
		Id: req.Id,
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, s.logger, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_CustomMessage resolve "org.federation.CustomMessage" message.
func (s *FederationService) resolve_Org_Federation_CustomMessage(ctx context.Context, req *Org_Federation_CustomMessageArgument) (*CustomMessage, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.CustomMessage")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.CustomMessage", slog.Any("message_args", s.logvalue_Org_Federation_CustomMessageArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.CustomMessageArgument", req)}

	// create a message value to be returned.
	ret := &CustomMessage{}

	// field binding section.
	// (grpc.federation.field).by = "$.message"
	if err := grpcfed.SetCELValue(ctx, value, "$.message", func(v string) { ret.Message = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.CustomMessage", slog.Any("org.federation.CustomMessage", s.logvalue_Org_Federation_CustomMessage(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetPostResponse resolve "org.federation.GetPostResponse" message.
func (s *FederationService) resolve_Org_Federation_GetPostResponse(ctx context.Context, req *Org_Federation_GetPostResponseArgument) (*GetPostResponse, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.GetPostResponse")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.GetPostResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetPostResponseArgument(req)))
	type localValueType struct {
		*grpcfed.LocalValue
		vars struct {
			_def1                  bool
			_def2                  bool
			_def2_err_detail0_msg0 *CustomMessage
			_def2_err_detail0_msg1 *CustomMessage
			post                   *Post
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(s.env, "grpc.federation.private.GetPostResponseArgument", req)}
	// A tree view of message dependencies is shown below.
	/*
	   post ─┐
	         _def1 ─┐
	   post ─┐      │
	         _def2 ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "post"
		     message {
		       name: "Post"
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*Post, *localValueType]{
			Name:   "post",
			Type:   grpcfed.CELObjectType("org.federation.Post"),
			Setter: func(value *localValueType, v *Post) { value.vars.post = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_PostArgument{}
				return s.resolve_Org_Federation_Post(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[bool, *localValueType]{
			Name:   "_def1",
			Type:   grpcfed.CELBoolType,
			Setter: func(value *localValueType, v bool) { value.vars._def1 = v },
			Validation: func(ctx context.Context, value *localValueType) error {
				var stat *grpcfed.Status
				if err := grpcfed.If(ctx1, value, "post.id != 'some-id'", func(value *localValueType) error {
					errorMessage, err := grpcfed.EvalCEL(ctx, value, "'validation message 1'", reflect.TypeOf(""))
					if err != nil {
						return err
					}
					stat = grpcfed.NewGRPCStatus(grpcfed.FailedPreconditionCode, errorMessage.(string))
					return nil
				}); err != nil {
					return err
				}
				return stat.Err()
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
		     name: "post"
		     message {
		       name: "Post"
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*Post, *localValueType]{
			Name:   "post",
			Type:   grpcfed.CELObjectType("org.federation.Post"),
			Setter: func(value *localValueType, v *Post) { value.vars.post = v },
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &Org_Federation_PostArgument{}
				return s.resolve_Org_Federation_Post(ctx, args)
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[bool, *localValueType]{
			Name:   "_def2",
			Type:   grpcfed.CELBoolType,
			Setter: func(value *localValueType, v bool) { value.vars._def2 = v },
			Validation: func(ctx context.Context, value *localValueType) error {
				var stat *grpcfed.Status
				if err := grpcfed.If(ctx1, value, "true", func(value *localValueType) error {
					errorMessage, err := grpcfed.EvalCEL(ctx, value, "'validation message 2'", reflect.TypeOf(""))
					if err != nil {
						return err
					}
					var details []grpcfed.ProtoMessage
					if err := grpcfed.If(ctx1, value, "post.title != 'some-title'", func(value *localValueType) error {
						if _, err := func() (any, error) {
							// A tree view of message dependencies is shown below.
							/*
							   _def2_err_detail0_msg0 ─┐
							   _def2_err_detail0_msg1 ─┤
							*/
							eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

							grpcfed.GoWithRecover(eg, func() (any, error) {

								// This section's codes are generated by the following proto definition.
								/*
								   def {
								     name: "_def2_err_detail0_msg0"
								     message {
								       name: "CustomMessage"
								       args { name: "message", string: "message1" }
								     }
								   }
								*/
								if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*CustomMessage, *localValueType]{
									Name:   "_def2_err_detail0_msg0",
									Type:   grpcfed.CELObjectType("org.federation.CustomMessage"),
									Setter: func(value *localValueType, v *CustomMessage) { value.vars._def2_err_detail0_msg0 = v },
									Message: func(ctx context.Context, value *localValueType) (any, error) {
										args := &Org_Federation_CustomMessageArgument{
											Message: "message1", // { name: "message", string: "message1" }
										}
										return s.resolve_Org_Federation_CustomMessage(ctx, args)
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
								     name: "_def2_err_detail0_msg1"
								     message {
								       name: "CustomMessage"
								       args { name: "message", string: "message2" }
								     }
								   }
								*/
								if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*CustomMessage, *localValueType]{
									Name:   "_def2_err_detail0_msg1",
									Type:   grpcfed.CELObjectType("org.federation.CustomMessage"),
									Setter: func(value *localValueType, v *CustomMessage) { value.vars._def2_err_detail0_msg1 = v },
									Message: func(ctx context.Context, value *localValueType) (any, error) {
										args := &Org_Federation_CustomMessageArgument{
											Message: "message2", // { name: "message", string: "message2" }
										}
										return s.resolve_Org_Federation_CustomMessage(ctx, args)
									},
								}); err != nil {
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
						if detail := grpcfed.CustomMessage(ctx1, value, "_def2_err_detail0_msg0", 0); detail != nil {
							details = append(details, detail)
						}
						if detail := grpcfed.CustomMessage(ctx1, value, "_def2_err_detail0_msg1", 1); detail != nil {
							details = append(details, detail)
						}
						if detail := grpcfed.PreconditionFailure(ctx, value, []*grpcfed.PreconditionFailureViolation{
							{
								Type:    "'some-type'",
								Subject: "'some-subject'",
								Desc:    "'some-description'",
							},
						}); detail != nil {
							details = append(details, detail)
						}
						if detail := grpcfed.BadRequest(ctx, value, []*grpcfed.BadRequestFieldViolation{
							{
								Field: "'some-field'",
								Desc:  "'some-description'",
							},
						}); detail != nil {
							details = append(details, detail)
						}
						if detail := grpcfed.LocalizedMessage(ctx, value, "en-US", "'some-message'"); detail != nil {
							details = append(details, detail)
						}
						return nil
					}); err != nil {
						return err
					}
					status := grpcfed.NewGRPCStatus(grpcfed.FailedPreconditionCode, errorMessage.(string))
					statusWithDetails, err := status.WithDetails(details...)
					if err != nil {
						s.logger.ErrorContext(ctx1, "failed setting error details", slog.String("error", err.Error()))
						stat = status
					} else {
						stat = statusWithDetails
					}
					return nil
				}); err != nil {
					return err
				}
				return stat.Err()
			},
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
	req.Post = value.vars.post
	req.XDef2ErrDetail0Msg0 = value.vars._def2_err_detail0_msg0
	req.XDef2ErrDetail0Msg1 = value.vars._def2_err_detail0_msg1

	// create a message value to be returned.
	ret := &GetPostResponse{}

	// field binding section.
	// (grpc.federation.field).by = "post"
	if err := grpcfed.SetCELValue(ctx, value, "post", func(v *Post) { ret.Post = v }); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetPostResponse", slog.Any("org.federation.GetPostResponse", s.logvalue_Org_Federation_GetPostResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_Post resolve "org.federation.Post" message.
func (s *FederationService) resolve_Org_Federation_Post(ctx context.Context, req *Org_Federation_PostArgument) (*Post, error) {
	ctx, span := s.tracer.Start(ctx, "org.federation.Post")
	defer span.End()

	s.logger.DebugContext(ctx, "resolve org.federation.Post", slog.Any("message_args", s.logvalue_Org_Federation_PostArgument(req)))

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = "some-id"           // (grpc.federation.field).string = "some-id"
	ret.Title = "some-title"     // (grpc.federation.field).string = "some-title"
	ret.Content = "some-content" // (grpc.federation.field).string = "some-content"

	s.logger.DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
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

func (s *FederationService) logvalue_Org_Federation_CustomMessageArgument(v *Org_Federation_CustomMessageArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_GetPostResponseArgument(v *Org_Federation_GetPostResponseArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *Org_Federation_PostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}
