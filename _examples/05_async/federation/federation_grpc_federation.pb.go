// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
package federation

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// Org_Federation_AAArgument is argument for "org.federation.AA" message.
type Org_Federation_AAArgument[T any] struct {
	Client T
}

// Org_Federation_AArgument is argument for "org.federation.A" message.
type Org_Federation_AArgument[T any] struct {
	Client T
}

// Org_Federation_ABArgument is argument for "org.federation.AB" message.
type Org_Federation_ABArgument[T any] struct {
	Client T
}

// Org_Federation_BArgument is argument for "org.federation.B" message.
type Org_Federation_BArgument[T any] struct {
	Client T
}

// Org_Federation_CArgument is argument for "org.federation.C" message.
type Org_Federation_CArgument[T any] struct {
	A      string
	Client T
}

// Org_Federation_DArgument is argument for "org.federation.D" message.
type Org_Federation_DArgument[T any] struct {
	B      string
	Client T
}

// Org_Federation_EArgument is argument for "org.federation.E" message.
type Org_Federation_EArgument[T any] struct {
	C      string
	D      string
	Client T
}

// Org_Federation_FArgument is argument for "org.federation.F" message.
type Org_Federation_FArgument[T any] struct {
	C      string
	D      string
	Client T
}

// Org_Federation_GArgument is argument for "org.federation.G" message.
type Org_Federation_GArgument[T any] struct {
	Client T
}

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type Org_Federation_GetResponseArgument[T any] struct {
	A      *A
	B      *B
	C      *C
	D      *D
	E      *E
	F      *F
	G      *G
	H      *H
	I      *I
	J      *J
	Client T
}

// Org_Federation_HArgument is argument for "org.federation.H" message.
type Org_Federation_HArgument[T any] struct {
	E      string
	F      string
	G      string
	Client T
}

// Org_Federation_IArgument is argument for "org.federation.I" message.
type Org_Federation_IArgument[T any] struct {
	Client T
}

// Org_Federation_JArgument is argument for "org.federation.J" message.
type Org_Federation_JArgument[T any] struct {
	I      string
	Client T
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
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
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
	env          *cel.Env
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
	celHelper := grpcfed.NewCELTypeHelper(map[string]map[string]*celtypes.FieldType{
		"grpc.federation.private.AAArgument": {},
		"grpc.federation.private.AArgument":  {},
		"grpc.federation.private.ABArgument": {},
		"grpc.federation.private.BArgument":  {},
		"grpc.federation.private.CArgument": {
			"a": grpcfed.NewCELFieldType(celtypes.StringType, "A"),
		},
		"grpc.federation.private.DArgument": {
			"b": grpcfed.NewCELFieldType(celtypes.StringType, "B"),
		},
		"grpc.federation.private.EArgument": {
			"c": grpcfed.NewCELFieldType(celtypes.StringType, "C"),
			"d": grpcfed.NewCELFieldType(celtypes.StringType, "D"),
		},
		"grpc.federation.private.FArgument": {
			"c": grpcfed.NewCELFieldType(celtypes.StringType, "C"),
			"d": grpcfed.NewCELFieldType(celtypes.StringType, "D"),
		},
		"grpc.federation.private.GArgument":           {},
		"grpc.federation.private.GetResponseArgument": {},
		"grpc.federation.private.HArgument": {
			"e": grpcfed.NewCELFieldType(celtypes.StringType, "E"),
			"f": grpcfed.NewCELFieldType(celtypes.StringType, "F"),
			"g": grpcfed.NewCELFieldType(celtypes.StringType, "G"),
		},
		"grpc.federation.private.IArgument": {},
		"grpc.federation.private.JArgument": {
			"i": grpcfed.NewCELFieldType(celtypes.StringType, "I"),
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
		client:       &FederationServiceDependentClientSet{},
	}, nil
}

// Get implements "org.federation.FederationService/Get" method.
func (s *FederationService) Get(ctx context.Context, req *GetRequest) (res *GetResponse, e error) {
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
		grpcfed.OutputErrorLog(ctx, s.logger, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_A resolve "org.federation.A" message.
func (s *FederationService) resolve_Org_Federation_A(ctx context.Context, req *Org_Federation_AArgument[*FederationServiceDependentClientSet]) (*A, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.A", slog.Any("message_args", s.logvalue_Org_Federation_AArgument(req)))
	var (
		sg      singleflight.Group
		valueMu sync.RWMutex
	)
	// A tree view of message dependencies is shown below.
	/*
	   aa ─┐
	   ab ─┤
	*/
	eg, ctx1 := errgroup.WithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "aa"
		     message: "AA"
		   }
		*/
		if _, err, _ := sg.Do("aa_org.federation.AA", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_AAArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_AA(ctx1, args)
		}); err != nil {
			return nil, err
		}
		valueMu.Lock()
		valueMu.Unlock()
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "ab"
		     message: "AB"
		   }
		*/
		if _, err, _ := sg.Do("ab_org.federation.AB", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_ABArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_AB(ctx1, args)
		}); err != nil {
			return nil, err
		}
		valueMu.Lock()
		valueMu.Unlock()
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// create a message value to be returned.
	ret := &A{}

	// field binding section.
	ret.Name = "a" // (grpc.federation.field).string = "a"

	s.logger.DebugContext(ctx, "resolved org.federation.A", slog.Any("org.federation.A", s.logvalue_Org_Federation_A(ret)))
	return ret, nil
}

// resolve_Org_Federation_AA resolve "org.federation.AA" message.
func (s *FederationService) resolve_Org_Federation_AA(ctx context.Context, req *Org_Federation_AAArgument[*FederationServiceDependentClientSet]) (*AA, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.AA", slog.Any("message_args", s.logvalue_Org_Federation_AAArgument(req)))

	// create a message value to be returned.
	ret := &AA{}

	// field binding section.
	ret.Name = "aa" // (grpc.federation.field).string = "aa"

	s.logger.DebugContext(ctx, "resolved org.federation.AA", slog.Any("org.federation.AA", s.logvalue_Org_Federation_AA(ret)))
	return ret, nil
}

// resolve_Org_Federation_AB resolve "org.federation.AB" message.
func (s *FederationService) resolve_Org_Federation_AB(ctx context.Context, req *Org_Federation_ABArgument[*FederationServiceDependentClientSet]) (*AB, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.AB", slog.Any("message_args", s.logvalue_Org_Federation_ABArgument(req)))

	// create a message value to be returned.
	ret := &AB{}

	// field binding section.
	ret.Name = "ab" // (grpc.federation.field).string = "ab"

	s.logger.DebugContext(ctx, "resolved org.federation.AB", slog.Any("org.federation.AB", s.logvalue_Org_Federation_AB(ret)))
	return ret, nil
}

// resolve_Org_Federation_B resolve "org.federation.B" message.
func (s *FederationService) resolve_Org_Federation_B(ctx context.Context, req *Org_Federation_BArgument[*FederationServiceDependentClientSet]) (*B, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.B", slog.Any("message_args", s.logvalue_Org_Federation_BArgument(req)))

	// create a message value to be returned.
	ret := &B{}

	// field binding section.
	ret.Name = "b" // (grpc.federation.field).string = "b"

	s.logger.DebugContext(ctx, "resolved org.federation.B", slog.Any("org.federation.B", s.logvalue_Org_Federation_B(ret)))
	return ret, nil
}

// resolve_Org_Federation_C resolve "org.federation.C" message.
func (s *FederationService) resolve_Org_Federation_C(ctx context.Context, req *Org_Federation_CArgument[*FederationServiceDependentClientSet]) (*C, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.C", slog.Any("message_args", s.logvalue_Org_Federation_CArgument(req)))

	// create a message value to be returned.
	ret := &C{}

	// field binding section.
	ret.Name = "c" // (grpc.federation.field).string = "c"

	s.logger.DebugContext(ctx, "resolved org.federation.C", slog.Any("org.federation.C", s.logvalue_Org_Federation_C(ret)))
	return ret, nil
}

// resolve_Org_Federation_D resolve "org.federation.D" message.
func (s *FederationService) resolve_Org_Federation_D(ctx context.Context, req *Org_Federation_DArgument[*FederationServiceDependentClientSet]) (*D, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.D", slog.Any("message_args", s.logvalue_Org_Federation_DArgument(req)))

	// create a message value to be returned.
	ret := &D{}

	// field binding section.
	ret.Name = "d" // (grpc.federation.field).string = "d"

	s.logger.DebugContext(ctx, "resolved org.federation.D", slog.Any("org.federation.D", s.logvalue_Org_Federation_D(ret)))
	return ret, nil
}

// resolve_Org_Federation_E resolve "org.federation.E" message.
func (s *FederationService) resolve_Org_Federation_E(ctx context.Context, req *Org_Federation_EArgument[*FederationServiceDependentClientSet]) (*E, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.E", slog.Any("message_args", s.logvalue_Org_Federation_EArgument(req)))

	// create a message value to be returned.
	ret := &E{}

	// field binding section.
	ret.Name = "e" // (grpc.federation.field).string = "e"

	s.logger.DebugContext(ctx, "resolved org.federation.E", slog.Any("org.federation.E", s.logvalue_Org_Federation_E(ret)))
	return ret, nil
}

// resolve_Org_Federation_F resolve "org.federation.F" message.
func (s *FederationService) resolve_Org_Federation_F(ctx context.Context, req *Org_Federation_FArgument[*FederationServiceDependentClientSet]) (*F, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.F", slog.Any("message_args", s.logvalue_Org_Federation_FArgument(req)))

	// create a message value to be returned.
	ret := &F{}

	// field binding section.
	ret.Name = "f" // (grpc.federation.field).string = "f"

	s.logger.DebugContext(ctx, "resolved org.federation.F", slog.Any("org.federation.F", s.logvalue_Org_Federation_F(ret)))
	return ret, nil
}

// resolve_Org_Federation_G resolve "org.federation.G" message.
func (s *FederationService) resolve_Org_Federation_G(ctx context.Context, req *Org_Federation_GArgument[*FederationServiceDependentClientSet]) (*G, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.G", slog.Any("message_args", s.logvalue_Org_Federation_GArgument(req)))

	// create a message value to be returned.
	ret := &G{}

	// field binding section.
	ret.Name = "g" // (grpc.federation.field).string = "g"

	s.logger.DebugContext(ctx, "resolved org.federation.G", slog.Any("org.federation.G", s.logvalue_Org_Federation_G(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *Org_Federation_GetResponseArgument[*FederationServiceDependentClientSet]) (*GetResponse, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))
	var (
		sg      singleflight.Group
		valueA  *A
		valueB  *B
		valueC  *C
		valueD  *D
		valueE  *E
		valueF  *F
		valueG  *G
		valueH  *H
		valueI  *I
		valueJ  *J
		valueMu sync.RWMutex
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.GetResponseArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}
	// A tree view of message dependencies is shown below.
	/*
	   a ─┐
	      c ─┐
	   b ─┐  │
	      d ─┤
	         e ─┐
	   a ─┐     │
	      c ─┐  │
	   b ─┐  │  │
	      d ─┤  │
	         f ─┤
	         g ─┤
	            h ─┐
	         i ─┐  │
	            j ─┤
	*/
	eg, ctx1 := errgroup.WithContext(ctx)

	grpcfed.GoWithRecover(eg, func() (any, error) {
		eg, ctx2 := errgroup.WithContext(ctx1)
		grpcfed.GoWithRecover(eg, func() (any, error) {
			eg, ctx3 := errgroup.WithContext(ctx2)
			grpcfed.GoWithRecover(eg, func() (any, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "a"
				     message: "A"
				   }
				*/
				resAIface, err, _ := sg.Do("a_org.federation.A", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_AArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_A(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resA := resAIface.(*A)
				valueMu.Lock()
				valueA = resA // { name: "a", message: "A" ... }
				envOpts = append(envOpts, cel.Variable("a", cel.ObjectType("org.federation.A")))
				evalValues["a"] = valueA
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "c"
				     message: "C"
				     args { name: "a", by: "a.name" }
				   }
				*/
				resCIface, err, _ := sg.Do("c_org.federation.C", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_CArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					// { name: "a", by: "a.name" }
					{
						_value, err := grpcfed.EvalCEL(s.env, "a.name", envOpts, evalValues, reflect.TypeOf(""))
						if err != nil {
							return nil, err
						}
						args.A = _value.(string)
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_C(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resC := resCIface.(*C)
				valueMu.Lock()
				valueC = resC // { name: "c", message: "C" ... }
				envOpts = append(envOpts, cel.Variable("c", cel.ObjectType("org.federation.C")))
				evalValues["c"] = valueC
				valueMu.Unlock()
				return nil, nil
			})
			grpcfed.GoWithRecover(eg, func() (any, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "b"
				     message: "B"
				   }
				*/
				resBIface, err, _ := sg.Do("b_org.federation.B", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_BArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_B(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resB := resBIface.(*B)
				valueMu.Lock()
				valueB = resB // { name: "b", message: "B" ... }
				envOpts = append(envOpts, cel.Variable("b", cel.ObjectType("org.federation.B")))
				evalValues["b"] = valueB
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "d"
				     message: "D"
				     args { name: "b", by: "b.name" }
				   }
				*/
				resDIface, err, _ := sg.Do("d_org.federation.D", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_DArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					// { name: "b", by: "b.name" }
					{
						_value, err := grpcfed.EvalCEL(s.env, "b.name", envOpts, evalValues, reflect.TypeOf(""))
						if err != nil {
							return nil, err
						}
						args.B = _value.(string)
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_D(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resD := resDIface.(*D)
				valueMu.Lock()
				valueD = resD // { name: "d", message: "D" ... }
				envOpts = append(envOpts, cel.Variable("d", cel.ObjectType("org.federation.D")))
				evalValues["d"] = valueD
				valueMu.Unlock()
				return nil, nil
			})
			if err := eg.Wait(); err != nil {
				return nil, err
			}

			// This section's codes are generated by the following proto definition.
			/*
			   {
			     name: "e"
			     message: "E"
			     args: [
			       { name: "c", by: "c.name" },
			       { name: "d", by: "d.name" }
			     ]
			   }
			*/
			resEIface, err, _ := sg.Do("e_org.federation.E", func() (any, error) {
				valueMu.RLock()
				args := &Org_Federation_EArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
				}
				// { name: "c", by: "c.name" }
				{
					_value, err := grpcfed.EvalCEL(s.env, "c.name", envOpts, evalValues, reflect.TypeOf(""))
					if err != nil {
						return nil, err
					}
					args.C = _value.(string)
				}
				// { name: "d", by: "d.name" }
				{
					_value, err := grpcfed.EvalCEL(s.env, "d.name", envOpts, evalValues, reflect.TypeOf(""))
					if err != nil {
						return nil, err
					}
					args.D = _value.(string)
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_E(ctx2, args)
			})
			if err != nil {
				return nil, err
			}
			resE := resEIface.(*E)
			valueMu.Lock()
			valueE = resE // { name: "e", message: "E" ... }
			envOpts = append(envOpts, cel.Variable("e", cel.ObjectType("org.federation.E")))
			evalValues["e"] = valueE
			valueMu.Unlock()
			return nil, nil
		})
		grpcfed.GoWithRecover(eg, func() (any, error) {
			eg, ctx3 := errgroup.WithContext(ctx2)
			grpcfed.GoWithRecover(eg, func() (any, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "a"
				     message: "A"
				   }
				*/
				resAIface, err, _ := sg.Do("a_org.federation.A", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_AArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_A(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resA := resAIface.(*A)
				valueMu.Lock()
				valueA = resA // { name: "a", message: "A" ... }
				envOpts = append(envOpts, cel.Variable("a", cel.ObjectType("org.federation.A")))
				evalValues["a"] = valueA
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "c"
				     message: "C"
				     args { name: "a", by: "a.name" }
				   }
				*/
				resCIface, err, _ := sg.Do("c_org.federation.C", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_CArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					// { name: "a", by: "a.name" }
					{
						_value, err := grpcfed.EvalCEL(s.env, "a.name", envOpts, evalValues, reflect.TypeOf(""))
						if err != nil {
							return nil, err
						}
						args.A = _value.(string)
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_C(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resC := resCIface.(*C)
				valueMu.Lock()
				valueC = resC // { name: "c", message: "C" ... }
				envOpts = append(envOpts, cel.Variable("c", cel.ObjectType("org.federation.C")))
				evalValues["c"] = valueC
				valueMu.Unlock()
				return nil, nil
			})
			grpcfed.GoWithRecover(eg, func() (any, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "b"
				     message: "B"
				   }
				*/
				resBIface, err, _ := sg.Do("b_org.federation.B", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_BArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_B(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resB := resBIface.(*B)
				valueMu.Lock()
				valueB = resB // { name: "b", message: "B" ... }
				envOpts = append(envOpts, cel.Variable("b", cel.ObjectType("org.federation.B")))
				evalValues["b"] = valueB
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "d"
				     message: "D"
				     args { name: "b", by: "b.name" }
				   }
				*/
				resDIface, err, _ := sg.Do("d_org.federation.D", func() (any, error) {
					valueMu.RLock()
					args := &Org_Federation_DArgument[*FederationServiceDependentClientSet]{
						Client: s.client,
					}
					// { name: "b", by: "b.name" }
					{
						_value, err := grpcfed.EvalCEL(s.env, "b.name", envOpts, evalValues, reflect.TypeOf(""))
						if err != nil {
							return nil, err
						}
						args.B = _value.(string)
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_D(ctx3, args)
				})
				if err != nil {
					return nil, err
				}
				resD := resDIface.(*D)
				valueMu.Lock()
				valueD = resD // { name: "d", message: "D" ... }
				envOpts = append(envOpts, cel.Variable("d", cel.ObjectType("org.federation.D")))
				evalValues["d"] = valueD
				valueMu.Unlock()
				return nil, nil
			})
			if err := eg.Wait(); err != nil {
				return nil, err
			}

			// This section's codes are generated by the following proto definition.
			/*
			   {
			     name: "f"
			     message: "F"
			     args: [
			       { name: "c", by: "c.name" },
			       { name: "d", by: "d.name" }
			     ]
			   }
			*/
			resFIface, err, _ := sg.Do("f_org.federation.F", func() (any, error) {
				valueMu.RLock()
				args := &Org_Federation_FArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
				}
				// { name: "c", by: "c.name" }
				{
					_value, err := grpcfed.EvalCEL(s.env, "c.name", envOpts, evalValues, reflect.TypeOf(""))
					if err != nil {
						return nil, err
					}
					args.C = _value.(string)
				}
				// { name: "d", by: "d.name" }
				{
					_value, err := grpcfed.EvalCEL(s.env, "d.name", envOpts, evalValues, reflect.TypeOf(""))
					if err != nil {
						return nil, err
					}
					args.D = _value.(string)
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_F(ctx2, args)
			})
			if err != nil {
				return nil, err
			}
			resF := resFIface.(*F)
			valueMu.Lock()
			valueF = resF // { name: "f", message: "F" ... }
			envOpts = append(envOpts, cel.Variable("f", cel.ObjectType("org.federation.F")))
			evalValues["f"] = valueF
			valueMu.Unlock()
			return nil, nil
		})
		grpcfed.GoWithRecover(eg, func() (any, error) {

			// This section's codes are generated by the following proto definition.
			/*
			   {
			     name: "g"
			     message: "G"
			   }
			*/
			resGIface, err, _ := sg.Do("g_org.federation.G", func() (any, error) {
				valueMu.RLock()
				args := &Org_Federation_GArgument[*FederationServiceDependentClientSet]{
					Client: s.client,
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_G(ctx2, args)
			})
			if err != nil {
				return nil, err
			}
			resG := resGIface.(*G)
			valueMu.Lock()
			valueG = resG // { name: "g", message: "G" ... }
			envOpts = append(envOpts, cel.Variable("g", cel.ObjectType("org.federation.G")))
			evalValues["g"] = valueG
			valueMu.Unlock()
			return nil, nil
		})
		if err := eg.Wait(); err != nil {
			return nil, err
		}

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "h"
		     message: "H"
		     args: [
		       { name: "e", by: "e.name" },
		       { name: "f", by: "f.name" },
		       { name: "g", by: "g.name" }
		     ]
		   }
		*/
		resHIface, err, _ := sg.Do("h_org.federation.H", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_HArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			// { name: "e", by: "e.name" }
			{
				_value, err := grpcfed.EvalCEL(s.env, "e.name", envOpts, evalValues, reflect.TypeOf(""))
				if err != nil {
					return nil, err
				}
				args.E = _value.(string)
			}
			// { name: "f", by: "f.name" }
			{
				_value, err := grpcfed.EvalCEL(s.env, "f.name", envOpts, evalValues, reflect.TypeOf(""))
				if err != nil {
					return nil, err
				}
				args.F = _value.(string)
			}
			// { name: "g", by: "g.name" }
			{
				_value, err := grpcfed.EvalCEL(s.env, "g.name", envOpts, evalValues, reflect.TypeOf(""))
				if err != nil {
					return nil, err
				}
				args.G = _value.(string)
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_H(ctx1, args)
		})
		if err != nil {
			return nil, err
		}
		resH := resHIface.(*H)
		valueMu.Lock()
		valueH = resH // { name: "h", message: "H" ... }
		envOpts = append(envOpts, cel.Variable("h", cel.ObjectType("org.federation.H")))
		evalValues["h"] = valueH
		valueMu.Unlock()
		return nil, nil
	})

	grpcfed.GoWithRecover(eg, func() (any, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "i"
		     message: "I"
		   }
		*/
		resIIface, err, _ := sg.Do("i_org.federation.I", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_IArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_I(ctx1, args)
		})
		if err != nil {
			return nil, err
		}
		resI := resIIface.(*I)
		valueMu.Lock()
		valueI = resI // { name: "i", message: "I" ... }
		envOpts = append(envOpts, cel.Variable("i", cel.ObjectType("org.federation.I")))
		evalValues["i"] = valueI
		valueMu.Unlock()

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "j"
		     message: "J"
		     args { name: "i", by: "i.name" }
		   }
		*/
		resJIface, err, _ := sg.Do("j_org.federation.J", func() (any, error) {
			valueMu.RLock()
			args := &Org_Federation_JArgument[*FederationServiceDependentClientSet]{
				Client: s.client,
			}
			// { name: "i", by: "i.name" }
			{
				_value, err := grpcfed.EvalCEL(s.env, "i.name", envOpts, evalValues, reflect.TypeOf(""))
				if err != nil {
					return nil, err
				}
				args.I = _value.(string)
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_J(ctx1, args)
		})
		if err != nil {
			return nil, err
		}
		resJ := resJIface.(*J)
		valueMu.Lock()
		valueJ = resJ // { name: "j", message: "J" ... }
		envOpts = append(envOpts, cel.Variable("j", cel.ObjectType("org.federation.J")))
		evalValues["j"] = valueJ
		valueMu.Unlock()
		return nil, nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// assign named parameters to message arguments to pass to the custom resolver.
	req.A = valueA
	req.B = valueB
	req.C = valueC
	req.D = valueD
	req.E = valueE
	req.F = valueF
	req.G = valueG
	req.H = valueH
	req.I = valueI
	req.J = valueJ

	// create a message value to be returned.
	ret := &GetResponse{}

	// field binding section.
	// (grpc.federation.field).by = "h.name"
	{
		_value, err := grpcfed.EvalCEL(s.env, "h.name", envOpts, evalValues, reflect.TypeOf(""))
		if err != nil {
			return nil, err
		}
		ret.Hname = _value.(string)
	}
	// (grpc.federation.field).by = "j.name"
	{
		_value, err := grpcfed.EvalCEL(s.env, "j.name", envOpts, evalValues, reflect.TypeOf(""))
		if err != nil {
			return nil, err
		}
		ret.Jname = _value.(string)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_H resolve "org.federation.H" message.
func (s *FederationService) resolve_Org_Federation_H(ctx context.Context, req *Org_Federation_HArgument[*FederationServiceDependentClientSet]) (*H, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.H", slog.Any("message_args", s.logvalue_Org_Federation_HArgument(req)))

	// create a message value to be returned.
	ret := &H{}

	// field binding section.
	ret.Name = "h" // (grpc.federation.field).string = "h"

	s.logger.DebugContext(ctx, "resolved org.federation.H", slog.Any("org.federation.H", s.logvalue_Org_Federation_H(ret)))
	return ret, nil
}

// resolve_Org_Federation_I resolve "org.federation.I" message.
func (s *FederationService) resolve_Org_Federation_I(ctx context.Context, req *Org_Federation_IArgument[*FederationServiceDependentClientSet]) (*I, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.I", slog.Any("message_args", s.logvalue_Org_Federation_IArgument(req)))

	// create a message value to be returned.
	ret := &I{}

	// field binding section.
	ret.Name = "i" // (grpc.federation.field).string = "i"

	s.logger.DebugContext(ctx, "resolved org.federation.I", slog.Any("org.federation.I", s.logvalue_Org_Federation_I(ret)))
	return ret, nil
}

// resolve_Org_Federation_J resolve "org.federation.J" message.
func (s *FederationService) resolve_Org_Federation_J(ctx context.Context, req *Org_Federation_JArgument[*FederationServiceDependentClientSet]) (*J, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.J", slog.Any("message_args", s.logvalue_Org_Federation_JArgument(req)))

	// create a message value to be returned.
	ret := &J{}

	// field binding section.
	ret.Name = "j" // (grpc.federation.field).string = "j"

	s.logger.DebugContext(ctx, "resolved org.federation.J", slog.Any("org.federation.J", s.logvalue_Org_Federation_J(ret)))
	return ret, nil
}

func (s *FederationService) logvalue_Org_Federation_A(v *A) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_AA(v *AA) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_AAArgument(v *Org_Federation_AAArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_AArgument(v *Org_Federation_AArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_AB(v *AB) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_ABArgument(v *Org_Federation_ABArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_B(v *B) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_BArgument(v *Org_Federation_BArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_C(v *C) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_CArgument(v *Org_Federation_CArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("a", v.A),
	)
}

func (s *FederationService) logvalue_Org_Federation_D(v *D) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_DArgument(v *Org_Federation_DArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("b", v.B),
	)
}

func (s *FederationService) logvalue_Org_Federation_E(v *E) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_EArgument(v *Org_Federation_EArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("c", v.C),
		slog.String("d", v.D),
	)
}

func (s *FederationService) logvalue_Org_Federation_F(v *F) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_FArgument(v *Org_Federation_FArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("c", v.C),
		slog.String("d", v.D),
	)
}

func (s *FederationService) logvalue_Org_Federation_G(v *G) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GArgument(v *Org_Federation_GArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_GetResponse(v *GetResponse) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("hname", v.GetHname()),
		slog.String("jname", v.GetJname()),
	)
}

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *Org_Federation_GetResponseArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_H(v *H) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_HArgument(v *Org_Federation_HArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("e", v.E),
		slog.String("f", v.F),
		slog.String("g", v.G),
	)
}

func (s *FederationService) logvalue_Org_Federation_I(v *I) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_IArgument(v *Org_Federation_IArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_J(v *J) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("name", v.GetName()),
	)
}

func (s *FederationService) logvalue_Org_Federation_JArgument(v *Org_Federation_JArgument[*FederationServiceDependentClientSet]) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("i", v.I),
	)
}