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
	postv2 "example/post/v2"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

// Org_Federation_GetPostResponseArgument is argument for "org.federation.GetPostResponse" message.
type FederationService_Org_Federation_GetPostResponseArgument struct {
	A          *GetPostRequest_ConditionA
	ConditionB *GetPostRequest_ConditionB
	Id         string
	Post       *Post
}

// Org_Federation_PostArgument is argument for "org.federation.Post" message.
type FederationService_Org_Federation_PostArgument struct {
	A         *GetPostRequest_ConditionA
	B         *GetPostRequest_ConditionB
	Data2     *postv2.PostData
	DataType  *grpcfedcel.EnumSelector
	DataType2 *grpcfedcel.EnumSelector
	Id        string
	Post      *post.Post
	Res       *post.GetPostResponse
	Res2      *postv2.GetPostResponse
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
	// Org_Post_V2_PostServiceClient create a gRPC Client to be used to call methods in org.post.v2.PostService.
	Org_Post_V2_PostServiceClient(FederationServiceClientConfig) (postv2.PostServiceClient, error)
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
	Org_Post_PostServiceClient    post.PostServiceClient
	Org_Post_V2_PostServiceClient postv2.PostServiceClient
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
	FederationService_DependentMethod_Org_Post_PostService_GetPost    = "/org.post.PostService/GetPost"
	FederationService_DependentMethod_Org_Post_V2_PostService_GetPost = "/org.post.v2.PostService/GetPost"
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
	Org_Post_V2_PostServiceClient, err := cfg.Client.Org_Post_V2_PostServiceClient(FederationServiceClientConfig{
		Service: "org.post.v2.PostService",
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
		"grpc.federation.private.GetPostResponseArgument": {
			"id":          grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
			"a":           grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("org.federation.GetPostRequest.ConditionA"), "A"),
			"condition_b": grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("org.federation.GetPostRequest.ConditionB"), "ConditionB"),
		},
		"grpc.federation.private.PostArgument": {
			"id": grpcfed.NewCELFieldType(grpcfed.CELStringType, "Id"),
			"a":  grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("org.federation.GetPostRequest.ConditionA"), "A"),
			"b":  grpcfed.NewCELFieldType(grpcfed.NewCELObjectType("org.federation.GetPostRequest.ConditionB"), "B"),
		},
	}
	celTypeHelper := grpcfed.NewCELTypeHelper("org.federation", celTypeHelperFieldMap)
	var celEnvOpts []grpcfed.CELEnvOption
	celEnvOpts = append(celEnvOpts, grpcfed.NewDefaultEnvOptions(celTypeHelper)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.PostContent.Category", PostContent_Category_value, PostContent_Category_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.federation.PostType", PostType_value, PostType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostContent.Category", post.PostContent_Category_value, post.PostContent_Category_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.PostDataType", post.PostDataType_value, post.PostDataType_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.v2.PostContent.Category", postv2.PostContent_Category_value, postv2.PostContent_Category_name)...)
	celEnvOpts = append(celEnvOpts, grpcfed.EnumAccessorOptions("org.post.v2.PostDataType", postv2.PostDataType_value, postv2.PostDataType_name)...)
	return &FederationService{
		cfg:           cfg,
		logger:        logger,
		errorHandler:  errorHandler,
		celEnvOpts:    celEnvOpts,
		celTypeHelper: celTypeHelper,
		celCacheMap:   grpcfed.NewCELCacheMap(),
		tracer:        otel.Tracer("org.federation.FederationService"),
		client: &FederationServiceDependentClientSet{
			Org_Post_PostServiceClient:    Org_Post_PostServiceClient,
			Org_Post_V2_PostServiceClient: Org_Post_V2_PostServiceClient,
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
		Id:         req.GetId(),
		A:          req.GetA(),
		ConditionB: req.GetConditionB(),
	})
	if err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		grpcfed.OutputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
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
	       args: [
	         { name: "id", by: "$.id" },
	         { name: "a", by: "$.a" },
	         { name: "b", by: "$.condition_b" }
	       ]
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
				CacheIndex: 1,
				Setter: func(v string) error {
					args.Id = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { name: "a", by: "$.a" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*GetPostRequest_ConditionA]{
				Value:      value,
				Expr:       `$.a`,
				CacheIndex: 2,
				Setter: func(v *GetPostRequest_ConditionA) error {
					args.A = v
					return nil
				},
			}); err != nil {
				return nil, err
			}
			// { name: "b", by: "$.condition_b" }
			if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*GetPostRequest_ConditionB]{
				Value:      value,
				Expr:       `$.condition_b`,
				CacheIndex: 3,
				Setter: func(v *GetPostRequest_ConditionB) error {
					args.B = v
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
		CacheIndex: 4,
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
			data2      *postv2.PostData
			data_type  *grpcfedcel.EnumSelector
			data_type2 *grpcfedcel.EnumSelector
			post       *post.Post
			res        *post.GetPostResponse
			res2       *postv2.GetPostResponse
		}
	}
	value := &localValueType{LocalValue: grpcfed.NewLocalValue(ctx, s.celEnvOpts, "grpc.federation.private.PostArgument", req)}
	// A tree view of message dependencies is shown below.
	/*
	        res2 ─┐
	                   data2 ─┐
	   data_type ─┐           │
	              data_type2 ─┤
	         res ─┐           │
	                    post ─┤
	*/
	eg, ctx1 := grpcfed.ErrorGroupWithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "res2"
		     call {
		       method: "org.post.v2.PostService/GetPost"
		       request { field: "id", by: "$.id" }
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*postv2.GetPostResponse, *localValueType]{
			Name: `res2`,
			Type: grpcfed.CELObjectType("org.post.v2.GetPostResponse"),
			Setter: func(value *localValueType, v *postv2.GetPostResponse) error {
				value.vars.res2 = v
				return nil
			},
			Message: func(ctx context.Context, value *localValueType) (any, error) {
				args := &postv2.GetPostRequest{}
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
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.v2.PostService/GetPost", slog.Any("org.post.v2.GetPostRequest", s.logvalue_Org_Post_V2_GetPostRequest(args)))
				ret, err := s.client.Org_Post_V2_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_V2_PostService_GetPost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "data2"
		     by: "res2.post.data"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*postv2.PostData, *localValueType]{
			Name: `data2`,
			Type: grpcfed.CELObjectType("org.post.v2.PostData"),
			Setter: func(value *localValueType, v *postv2.PostData) error {
				value.vars.data2 = v
				return nil
			},
			By:           `res2.post.data`,
			ByCacheIndex: 6,
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
		     name: "data_type"
		     by: "grpc.federation.enum.select(true, org.post.PostDataType.from(org.post.PostDataType.POST_TYPE_B), org.post.v2.PostDataType.value('POST_V2_TYPE_B'))"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.EnumSelector, *localValueType]{
			Name: `data_type`,
			Type: grpcfed.CELObjectType("grpc.federation.private.EnumSelector"),
			Setter: func(value *localValueType, v *grpcfedcel.EnumSelector) error {
				value.vars.data_type = v
				return nil
			},
			By:           `grpc.federation.enum.select(true, org.post.PostDataType.from(org.post.PostDataType.POST_TYPE_B), org.post.v2.PostDataType.value('POST_V2_TYPE_B'))`,
			ByCacheIndex: 7,
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   def {
		     name: "data_type2"
		     by: "grpc.federation.enum.select(true, data_type, org.post.v2.PostDataType.value('POST_V2_TYPE_C'))"
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*grpcfedcel.EnumSelector, *localValueType]{
			Name: `data_type2`,
			Type: grpcfed.CELObjectType("grpc.federation.private.EnumSelector"),
			Setter: func(value *localValueType, v *grpcfedcel.EnumSelector) error {
				value.vars.data_type2 = v
				return nil
			},
			By:           `grpc.federation.enum.select(true, data_type, org.post.v2.PostDataType.value('POST_V2_TYPE_C'))`,
			ByCacheIndex: 8,
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
		     name: "res"
		     call {
		       method: "org.post.PostService/GetPost"
		       request: [
		         { field: "id", by: "$.id" },
		         { field: "a", by: "$.a", if: "$.a != null" },
		         { field: "b", by: "$.b", if: "$.b != null" }
		       ]
		     }
		   }
		*/
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*post.GetPostResponse, *localValueType]{
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
					CacheIndex: 9,
					Setter: func(v string) error {
						args.Id = v
						return nil
					},
				}); err != nil {
					return nil, err
				}
				// { field: "a", by: "$.a", if: "$.a != null" }
				if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
					Value:      value,
					Expr:       `$.a != null`,
					CacheIndex: 10,
					Body: func(value *localValueType) error {
						return grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*GetPostRequest_ConditionA]{
							Value:      value,
							Expr:       `$.a`,
							CacheIndex: 11,
							Setter: func(v *GetPostRequest_ConditionA) error {
								aValue, err := s.cast_Org_Federation_GetPostRequest_ConditionA__to__Org_Post_PostConditionA(v)
								if err != nil {
									return err
								}
								args.Condition = &post.GetPostRequest_A{
									A: aValue,
								}
								return nil
							},
						})
					},
				}); err != nil {
					return nil, err
				}
				// { field: "b", by: "$.b", if: "$.b != null" }
				if err := grpcfed.If(ctx, &grpcfed.IfParam[*localValueType]{
					Value:      value,
					Expr:       `$.b != null`,
					CacheIndex: 12,
					Body: func(value *localValueType) error {
						return grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*GetPostRequest_ConditionB]{
							Value:      value,
							Expr:       `$.b`,
							CacheIndex: 13,
							Setter: func(v *GetPostRequest_ConditionB) error {
								bValue, err := s.cast_Org_Federation_GetPostRequest_ConditionB__to__Org_Post_PostConditionB(v)
								if err != nil {
									return err
								}
								args.Condition = &post.GetPostRequest_B{
									B: bValue,
								}
								return nil
							},
						})
					},
				}); err != nil {
					return nil, err
				}
				grpcfed.Logger(ctx).DebugContext(ctx, "call org.post.PostService/GetPost", slog.Any("org.post.GetPostRequest", s.logvalue_Org_Post_GetPostRequest(args)))
				ret, err := s.client.Org_Post_PostServiceClient.GetPost(ctx, args)
				if err != nil {
					if err := s.errorHandler(ctx, FederationService_DependentMethod_Org_Post_PostService_GetPost, err); err != nil {
						return nil, grpcfed.NewErrorWithLogAttrs(err, slog.LevelError, grpcfed.LogAttrs(ctx))
					}
				}
				return ret, nil
			},
		}); err != nil {
			grpcfed.RecordErrorToSpan(ctx1, err)
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
		if err := grpcfed.EvalDef(ctx1, value, grpcfed.Def[*post.Post, *localValueType]{
			Name: `post`,
			Type: grpcfed.CELObjectType("org.post.Post"),
			Setter: func(value *localValueType, v *post.Post) error {
				value.vars.post = v
				return nil
			},
			By:           `res.post`,
			ByCacheIndex: 14,
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
	req.Data2 = value.vars.data2
	req.DataType = value.vars.data_type
	req.DataType2 = value.vars.data_type2
	req.Post = value.vars.post
	req.Res = value.vars.res
	req.Res2 = value.vars.res2

	// create a message value to be returned.
	ret := &Post{}

	// field binding section.
	ret.Id = value.vars.post.GetId() // { name: "post", autobind: true }
	{
		dataValue, err := s.cast_Org_Post_PostData__to__Org_Federation_PostData(value.vars.post.GetData()) // { name: "post", autobind: true }
		if err != nil {
			grpcfed.RecordErrorToSpan(ctx, err)
			return nil, err
		}
		ret.Data = dataValue
	}
	// (grpc.federation.field).by = "data2"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*postv2.PostData]{
		Value:      value,
		Expr:       `data2`,
		CacheIndex: 15,
		Setter: func(v *postv2.PostData) error {
			data2Value, err := s.cast_Org_Post_V2_PostData__to__Org_Federation_PostData(v)
			if err != nil {
				return err
			}
			ret.Data2 = data2Value
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "data_type2"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*grpcfedcel.EnumSelector]{
		Value:      value,
		Expr:       `data_type2`,
		CacheIndex: 16,
		Setter: func(v *grpcfedcel.EnumSelector) error {
			var typeValue PostType
			if v.GetCond() {
				if err := func(v *grpcfedcel.EnumSelector) error {
					if v.GetCond() {
						casted, err := s.cast_Org_Post_PostDataType__to__Org_Federation_PostType(post.PostDataType(v.GetTrueValue()))
						if err != nil {
							return err
						}
						typeValue = casted
					} else {
						casted, err := s.cast_Org_Post_V2_PostDataType__to__Org_Federation_PostType(postv2.PostDataType(v.GetFalseValue()))
						if err != nil {
							return err
						}
						typeValue = casted
					}
					return nil
				}(v.GetTrueSelector()); err != nil {
					return err
				}
			} else {
				casted, err := s.cast_Org_Post_V2_PostDataType__to__Org_Federation_PostType(postv2.PostDataType(v.GetFalseValue()))
				if err != nil {
					return err
				}
				typeValue = casted
			}
			ret.Type = typeValue
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}
	// (grpc.federation.field).by = "M{x: 'xxx'}"
	if err := grpcfed.SetCELValue(ctx, &grpcfed.SetCELValueParam[*M]{
		Value:      value,
		Expr:       `M{x: 'xxx'}`,
		CacheIndex: 17,
		Setter: func(v *M) error {
			mValue, err := s.cast_Org_Federation_M__to__Org_Post_M(v)
			if err != nil {
				return err
			}
			ret.M = mValue
			return nil
		},
	}); err != nil {
		grpcfed.RecordErrorToSpan(ctx, err)
		return nil, err
	}

	grpcfed.Logger(ctx).DebugContext(ctx, "resolved org.federation.Post", slog.Any("org.federation.Post", s.logvalue_Org_Federation_Post(ret)))
	return ret, nil
}

// cast_Org_Federation_GetPostRequest_ConditionA__to__Org_Post_PostConditionA cast from "org.federation.GetPostRequest.ConditionA" to "org.post.PostConditionA".
func (s *FederationService) cast_Org_Federation_GetPostRequest_ConditionA__to__Org_Post_PostConditionA(from *GetPostRequest_ConditionA) (*post.PostConditionA, error) {
	if from == nil {
		return nil, nil
	}

	propValue := from.GetProp()

	return &post.PostConditionA{
		Prop: propValue,
	}, nil
}

// cast_Org_Federation_GetPostRequest_ConditionB__to__Org_Post_PostConditionB cast from "org.federation.GetPostRequest.ConditionB" to "org.post.PostConditionB".
func (s *FederationService) cast_Org_Federation_GetPostRequest_ConditionB__to__Org_Post_PostConditionB(from *GetPostRequest_ConditionB) (*post.PostConditionB, error) {
	if from == nil {
		return nil, nil
	}

	return &post.PostConditionB{}, nil
}

// cast_Org_Federation_M__to__Org_Post_M cast from "org.federation.M" to "org.post.M".
func (s *FederationService) cast_Org_Federation_M__to__Org_Post_M(from *M) (*post.M, error) {
	if from == nil {
		return nil, nil
	}

	xValue := from.GetX()

	return &post.M{
		X: xValue,
	}, nil
}

// cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category cast from "org.post.PostContent.Category" to "org.federation.PostContent.Category".
func (s *FederationService) cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category(from post.PostContent_Category) (PostContent_Category, error) {
	switch from {
	case post.PostContent_CATEGORY_A:
		return PostContent_CATEGORY_A, nil
	case post.PostContent_CATEGORY_B:
		return PostContent_CATEGORY_B, nil
	default:
		return 0, nil
	}
}

// cast_Org_Post_PostContent__to__Org_Federation_PostContent cast from "org.post.PostContent" to "org.federation.PostContent".
func (s *FederationService) cast_Org_Post_PostContent__to__Org_Federation_PostContent(from *post.PostContent) (*PostContent, error) {
	if from == nil {
		return nil, nil
	}

	categoryValue, err := s.cast_Org_Post_PostContent_Category__to__Org_Federation_PostContent_Category(from.GetCategory())
	if err != nil {
		return nil, err
	}
	headValue := from.GetHead()
	bodyValue := from.GetBody()
	dupBodyValue := from.GetBody()
	countsValue := from.GetCounts()

	return &PostContent{
		Category: categoryValue,
		Head:     headValue,
		Body:     bodyValue,
		DupBody:  dupBodyValue,
		Counts:   countsValue,
	}, nil
}

// cast_Org_Post_PostDataType__to__Org_Federation_PostType cast from "org.post.PostDataType" to "org.federation.PostType".
func (s *FederationService) cast_Org_Post_PostDataType__to__Org_Federation_PostType(from post.PostDataType) (PostType, error) {
	switch from {
	case post.PostDataType_POST_TYPE_A:
		return PostType_POST_TYPE_FOO, nil
	case post.PostDataType_POST_TYPE_B:
		return PostType_POST_TYPE_BAR, nil
	case post.PostDataType_POST_TYPE_C:
		return PostType_POST_TYPE_BAR, nil
	default:
		return PostType_POST_TYPE_UNKNOWN, nil
	}
}

// cast_Org_Post_PostData__to__Org_Federation_PostData cast from "org.post.PostData" to "org.federation.PostData".
func (s *FederationService) cast_Org_Post_PostData__to__Org_Federation_PostData(from *post.PostData) (*PostData, error) {
	if from == nil {
		return nil, nil
	}

	typeValue, err := s.cast_Org_Post_PostDataType__to__Org_Federation_PostType(from.GetType())
	if err != nil {
		return nil, err
	}
	titleValue := from.GetTitle()
	contentValue, err := s.cast_Org_Post_PostContent__to__Org_Federation_PostContent(from.GetContent())
	if err != nil {
		return nil, err
	}

	return &PostData{
		Type:    typeValue,
		Title:   titleValue,
		Content: contentValue,
	}, nil
}

// cast_Org_Post_V2_PostContent_Category__to__Org_Federation_PostContent_Category cast from "org.post.v2.PostContent.Category" to "org.federation.PostContent.Category".
func (s *FederationService) cast_Org_Post_V2_PostContent_Category__to__Org_Federation_PostContent_Category(from postv2.PostContent_Category) (PostContent_Category, error) {
	switch from {
	case postv2.PostContent_CATEGORY_A:
		return PostContent_CATEGORY_A, nil
	case postv2.PostContent_CATEGORY_B:
		return PostContent_CATEGORY_B, nil
	default:
		return 0, nil
	}
}

// cast_Org_Post_V2_PostContent__to__Org_Federation_PostContent cast from "org.post.v2.PostContent" to "org.federation.PostContent".
func (s *FederationService) cast_Org_Post_V2_PostContent__to__Org_Federation_PostContent(from *postv2.PostContent) (*PostContent, error) {
	if from == nil {
		return nil, nil
	}

	categoryValue, err := s.cast_Org_Post_V2_PostContent_Category__to__Org_Federation_PostContent_Category(from.GetCategory())
	if err != nil {
		return nil, err
	}
	headValue := from.GetHead()
	bodyValue := from.GetBody()
	dupBodyValue := from.GetBody()
	countsValue := from.GetCounts()

	return &PostContent{
		Category: categoryValue,
		Head:     headValue,
		Body:     bodyValue,
		DupBody:  dupBodyValue,
		Counts:   countsValue,
	}, nil
}

// cast_Org_Post_V2_PostDataType__to__Org_Federation_PostType cast from "org.post.v2.PostDataType" to "org.federation.PostType".
func (s *FederationService) cast_Org_Post_V2_PostDataType__to__Org_Federation_PostType(from postv2.PostDataType) (PostType, error) {
	switch from {
	case postv2.PostDataType_POST_TYPE_A:
		return PostType_POST_TYPE_FOO, nil
	case postv2.PostDataType_POST_V2_TYPE_B:
		return PostType_POST_TYPE_BAR, nil
	case postv2.PostDataType_POST_V2_TYPE_C:
		return PostType_POST_TYPE_BAR, nil
	default:
		return PostType_POST_TYPE_UNKNOWN, nil
	}
}

// cast_Org_Post_V2_PostData__to__Org_Federation_PostData cast from "org.post.v2.PostData" to "org.federation.PostData".
func (s *FederationService) cast_Org_Post_V2_PostData__to__Org_Federation_PostData(from *postv2.PostData) (*PostData, error) {
	if from == nil {
		return nil, nil
	}

	typeValue, err := s.cast_Org_Post_V2_PostDataType__to__Org_Federation_PostType(from.GetType())
	if err != nil {
		return nil, err
	}
	titleValue := from.GetTitle()
	contentValue, err := s.cast_Org_Post_V2_PostContent__to__Org_Federation_PostContent(from.GetContent())
	if err != nil {
		return nil, err
	}

	return &PostData{
		Type:    typeValue,
		Title:   titleValue,
		Content: contentValue,
	}, nil
}

func (s *FederationService) logvalue_Org_Federation_GetPostRequest_ConditionA(v *GetPostRequest_ConditionA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("prop", v.GetProp()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetPostRequest_ConditionB(v *GetPostRequest_ConditionB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
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
		slog.Any("a", s.logvalue_Org_Federation_GetPostRequest_ConditionA(v.A)),
		slog.Any("condition_b", s.logvalue_Org_Federation_GetPostRequest_ConditionB(v.ConditionB)),
	)
}

func (s *FederationService) logvalue_Org_Federation_Post(v *Post) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.Any("data", s.logvalue_Org_Federation_PostData(v.GetData())),
		slog.Any("data2", s.logvalue_Org_Federation_PostData(v.GetData2())),
		slog.String("type", s.logvalue_Org_Federation_PostType(v.GetType()).String()),
		slog.Any("m", s.logvalue_Org_Post_M(v.GetM())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostArgument(v *FederationService_Org_Federation_PostArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.Id),
		slog.Any("a", s.logvalue_Org_Federation_GetPostRequest_ConditionA(v.A)),
		slog.Any("b", s.logvalue_Org_Federation_GetPostRequest_ConditionB(v.B)),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostContent(v *PostContent) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("category", s.logvalue_Org_Federation_PostContent_Category(v.GetCategory()).String()),
		slog.String("head", v.GetHead()),
		slog.String("body", v.GetBody()),
		slog.String("dup_body", v.GetDupBody()),
		slog.Any("counts", s.logvalue_Org_Federation_PostContent_CountsEntry(v.GetCounts())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostContent_Category(v PostContent_Category) slog.Value {
	switch v {
	case PostContent_CATEGORY_A:
		return slog.StringValue("CATEGORY_A")
	case PostContent_CATEGORY_B:
		return slog.StringValue("CATEGORY_B")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Federation_PostContent_CountsEntry(v map[int32]int32) slog.Value {
	attrs := make([]slog.Attr, 0, len(v))
	for key, value := range v {
		attrs = append(attrs, slog.Attr{
			Key:   grpcfed.ToLogAttrKey(key),
			Value: slog.AnyValue(value),
		})
	}
	return slog.GroupValue(attrs...)
}

func (s *FederationService) logvalue_Org_Federation_PostData(v *PostData) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("type", s.logvalue_Org_Federation_PostType(v.GetType()).String()),
		slog.String("title", v.GetTitle()),
		slog.Any("content", s.logvalue_Org_Federation_PostContent(v.GetContent())),
	)
}

func (s *FederationService) logvalue_Org_Federation_PostType(v PostType) slog.Value {
	switch v {
	case PostType_POST_TYPE_UNKNOWN:
		return slog.StringValue("POST_TYPE_UNKNOWN")
	case PostType_POST_TYPE_FOO:
		return slog.StringValue("POST_TYPE_FOO")
	case PostType_POST_TYPE_BAR:
		return slog.StringValue("POST_TYPE_BAR")
	}
	return slog.StringValue("")
}

func (s *FederationService) logvalue_Org_Post_GetPostRequest(v *post.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
		slog.Any("a", s.logvalue_Org_Post_PostConditionA(v.GetA())),
		slog.Any("b", s.logvalue_Org_Post_PostConditionB(v.GetB())),
	)
}

func (s *FederationService) logvalue_Org_Post_M(v *post.M) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("x", v.GetX()),
	)
}

func (s *FederationService) logvalue_Org_Post_PostConditionA(v *post.PostConditionA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("prop", v.GetProp()),
	)
}

func (s *FederationService) logvalue_Org_Post_PostConditionB(v *post.PostConditionB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Post_V2_GetPostRequest(v *postv2.GetPostRequest) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("id", v.GetId()),
	)
}
