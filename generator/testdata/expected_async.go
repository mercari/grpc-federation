// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
package federation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// FederationServiceConfig configuration required to initialize the service that use GRPC Federation.
type FederationServiceConfig struct {
	// ErrorHandler Federation Service often needs to convert errors received from downstream services.
	// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
	ErrorHandler FederationServiceErrorHandler
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

// FederationServiceDependencyServiceClient has a gRPC client for all services on which the federation service depends.
// This is provided as an argument when implementing the custom resolver.
type FederationServiceDependencyServiceClient struct {
}

// FederationServiceResolver provides an interface to directly implement message resolver and field resolver not defined in Protocol Buffers.
type FederationServiceResolver interface {
}

// FederationServiceUnimplementedResolver a structure implemented to satisfy the Resolver interface.
// An Unimplemented error is always returned.
// This is intended for use when there are many Resolver interfaces that do not need to be implemented,
// by embedding them in a resolver structure that you have created.
type FederationServiceUnimplementedResolver struct{}

// FederationServiceErrorHandler Federation Service often needs to convert errors received from downstream services.
// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
type FederationServiceErrorHandler func(ctx context.Context, methodName string, err error) error

// FederationServiceRecoveredError represents recovered error.
type FederationServiceRecoveredError struct {
	Message string
	Stack   []string
}

func (e *FederationServiceRecoveredError) Error() string {
	return fmt.Sprintf("recovered error: %s", e.Message)
}

// FederationService represents Federation Service.
type FederationService struct {
	*UnimplementedFederationServiceServer
	cfg          FederationServiceConfig
	logger       *slog.Logger
	errorHandler FederationServiceErrorHandler
	client       *FederationServiceDependencyServiceClient
}

// Org_Federation_AAArgument is argument for "org.federation.AA" message.
type Org_Federation_AAArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_AArgument is argument for "org.federation.A" message.
type Org_Federation_AArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_ABArgument is argument for "org.federation.AB" message.
type Org_Federation_ABArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_BArgument is argument for "org.federation.B" message.
type Org_Federation_BArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_CArgument is argument for "org.federation.C" message.
type Org_Federation_CArgument struct {
	A      string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_DArgument is argument for "org.federation.D" message.
type Org_Federation_DArgument struct {
	B      string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_EArgument is argument for "org.federation.E" message.
type Org_Federation_EArgument struct {
	C      string
	D      string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_FArgument is argument for "org.federation.F" message.
type Org_Federation_FArgument struct {
	C      string
	D      string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_GArgument is argument for "org.federation.G" message.
type Org_Federation_GArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type Org_Federation_GetResponseArgument struct {
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
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_HArgument is argument for "org.federation.H" message.
type Org_Federation_HArgument struct {
	E      string
	F      string
	G      string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_IArgument is argument for "org.federation.I" message.
type Org_Federation_IArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_JArgument is argument for "org.federation.J" message.
type Org_Federation_JArgument struct {
	I      string
	Client *FederationServiceDependencyServiceClient
}

// NewFederationService creates FederationService instance by FederationServiceConfig.
func NewFederationService(cfg FederationServiceConfig) (*FederationService, error) {
	if err := validateFederationServiceConfig(cfg); err != nil {
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
	return &FederationService{
		cfg:          cfg,
		logger:       logger,
		errorHandler: errorHandler,
		client:       &FederationServiceDependencyServiceClient{},
	}, nil
}

func validateFederationServiceConfig(cfg FederationServiceConfig) error {
	return nil
}

func withTimeoutFederationService[T any](ctx context.Context, method string, timeout time.Duration, fn func(context.Context) (*T, error)) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		ret   *T
		errch = make(chan error)
	)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errch <- recoverErrorFederationService(r, debug.Stack())
			}
		}()

		res, err := fn(ctx)
		ret = res
		errch <- err
	}()
	select {
	case <-ctx.Done():
		status := grpcstatus.New(grpccodes.DeadlineExceeded, ctx.Err().Error())
		withDetails, err := status.WithDetails(&errdetails.ErrorInfo{
			Metadata: map[string]string{
				"method":  method,
				"timeout": timeout.String(),
			},
		})
		if err != nil {
			return nil, status.Err()
		}
		return nil, withDetails.Err()
	case err := <-errch:
		return ret, err
	}
}

func withRetryFederationService[T any](b backoff.BackOff, fn func() (*T, error)) (*T, error) {
	var res *T
	if err := backoff.Retry(func() (err error) {
		res, err = fn()
		return
	}, b); err != nil {
		return nil, err
	}
	return res, nil
}

func recoverErrorFederationService(v interface{}, rawStack []byte) *FederationServiceRecoveredError {
	msg := fmt.Sprint(v)
	lines := strings.Split(msg, "\n")
	if len(lines) <= 1 {
		lines := strings.Split(string(rawStack), "\n")
		stack := make([]string, 0, len(lines))
		for _, line := range lines {
			if line == "" {
				continue
			}
			stack = append(stack, strings.TrimPrefix(line, "\t"))
		}
		return &FederationServiceRecoveredError{
			Message: msg,
			Stack:   stack,
		}
	}
	// If panic occurs under singleflight, singleflight's recover catches the error and gives a stack trace.
	// Therefore, once the stack trace is removed.
	stack := make([]string, 0, len(lines))
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		stack = append(stack, strings.TrimPrefix(line, "\t"))
	}
	return &FederationServiceRecoveredError{
		Message: lines[0],
		Stack:   stack,
	}
}

func (s *FederationService) goWithRecover(eg *errgroup.Group, fn func() (interface{}, error)) {
	eg.Go(func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = recoverErrorFederationService(r, debug.Stack())
			}
		}()
		_, err := fn()
		return err
	})
}

func (s *FederationService) outputErrorLog(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if status, ok := grpcstatus.FromError(err); ok {
		s.logger.ErrorContext(ctx, status.Message(),
			slog.Group("grpc_status",
				slog.String("code", status.Code().String()),
				slog.Any("details", status.Details()),
			),
		)
		return
	}
	var recoveredErr *FederationServiceRecoveredError
	if errors.As(err, &recoveredErr) {
		trace := make([]interface{}, 0, len(recoveredErr.Stack))
		for idx, stack := range recoveredErr.Stack {
			trace = append(trace, slog.String(fmt.Sprint(idx+1), stack))
		}
		s.logger.ErrorContext(ctx, recoveredErr.Message, slog.Group("stack_trace", trace...))
		return
	}
	s.logger.ErrorContext(ctx, err.Error())
}

// Get implements "org.federation.FederationService/Get" method.
func (s *FederationService) Get(ctx context.Context, req *GetRequest) (res *GetResponse, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = recoverErrorFederationService(r, debug.Stack())
			s.outputErrorLog(ctx, e)
		}
	}()
	res, err := s.resolve_Org_Federation_GetResponse(ctx, &Org_Federation_GetResponseArgument{
		Client: s.client,
	})
	if err != nil {
		s.outputErrorLog(ctx, err)
		return nil, err
	}
	return res, nil
}

// resolve_Org_Federation_A resolve "org.federation.A" message.
func (s *FederationService) resolve_Org_Federation_A(ctx context.Context, req *Org_Federation_AArgument) (*A, error) {
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
	eg, egCtx := errgroup.WithContext(ctx)

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "aa"
		     message: "AA"
		   }
		*/
		if _, err, _ := sg.Do("aa_org.federation.AA", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_AAArgument{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_AA(egCtx, args)
		}); err != nil {
			return nil, err
		}
		valueMu.Lock()
		valueMu.Unlock()
		return nil, nil
	})

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "ab"
		     message: "AB"
		   }
		*/
		if _, err, _ := sg.Do("ab_org.federation.AB", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_ABArgument{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_AB(egCtx, args)
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
func (s *FederationService) resolve_Org_Federation_AA(ctx context.Context, req *Org_Federation_AAArgument) (*AA, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.AA", slog.Any("message_args", s.logvalue_Org_Federation_AAArgument(req)))

	// create a message value to be returned.
	ret := &AA{}

	// field binding section.
	ret.Name = "aa" // (grpc.federation.field).string = "aa"

	s.logger.DebugContext(ctx, "resolved org.federation.AA", slog.Any("org.federation.AA", s.logvalue_Org_Federation_AA(ret)))
	return ret, nil
}

// resolve_Org_Federation_AB resolve "org.federation.AB" message.
func (s *FederationService) resolve_Org_Federation_AB(ctx context.Context, req *Org_Federation_ABArgument) (*AB, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.AB", slog.Any("message_args", s.logvalue_Org_Federation_ABArgument(req)))

	// create a message value to be returned.
	ret := &AB{}

	// field binding section.
	ret.Name = "ab" // (grpc.federation.field).string = "ab"

	s.logger.DebugContext(ctx, "resolved org.federation.AB", slog.Any("org.federation.AB", s.logvalue_Org_Federation_AB(ret)))
	return ret, nil
}

// resolve_Org_Federation_B resolve "org.federation.B" message.
func (s *FederationService) resolve_Org_Federation_B(ctx context.Context, req *Org_Federation_BArgument) (*B, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.B", slog.Any("message_args", s.logvalue_Org_Federation_BArgument(req)))

	// create a message value to be returned.
	ret := &B{}

	// field binding section.
	ret.Name = "b" // (grpc.federation.field).string = "b"

	s.logger.DebugContext(ctx, "resolved org.federation.B", slog.Any("org.federation.B", s.logvalue_Org_Federation_B(ret)))
	return ret, nil
}

// resolve_Org_Federation_C resolve "org.federation.C" message.
func (s *FederationService) resolve_Org_Federation_C(ctx context.Context, req *Org_Federation_CArgument) (*C, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.C", slog.Any("message_args", s.logvalue_Org_Federation_CArgument(req)))

	// create a message value to be returned.
	ret := &C{}

	// field binding section.
	ret.Name = "c" // (grpc.federation.field).string = "c"

	s.logger.DebugContext(ctx, "resolved org.federation.C", slog.Any("org.federation.C", s.logvalue_Org_Federation_C(ret)))
	return ret, nil
}

// resolve_Org_Federation_D resolve "org.federation.D" message.
func (s *FederationService) resolve_Org_Federation_D(ctx context.Context, req *Org_Federation_DArgument) (*D, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.D", slog.Any("message_args", s.logvalue_Org_Federation_DArgument(req)))

	// create a message value to be returned.
	ret := &D{}

	// field binding section.
	ret.Name = "d" // (grpc.federation.field).string = "d"

	s.logger.DebugContext(ctx, "resolved org.federation.D", slog.Any("org.federation.D", s.logvalue_Org_Federation_D(ret)))
	return ret, nil
}

// resolve_Org_Federation_E resolve "org.federation.E" message.
func (s *FederationService) resolve_Org_Federation_E(ctx context.Context, req *Org_Federation_EArgument) (*E, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.E", slog.Any("message_args", s.logvalue_Org_Federation_EArgument(req)))

	// create a message value to be returned.
	ret := &E{}

	// field binding section.
	ret.Name = "e" // (grpc.federation.field).string = "e"

	s.logger.DebugContext(ctx, "resolved org.federation.E", slog.Any("org.federation.E", s.logvalue_Org_Federation_E(ret)))
	return ret, nil
}

// resolve_Org_Federation_F resolve "org.federation.F" message.
func (s *FederationService) resolve_Org_Federation_F(ctx context.Context, req *Org_Federation_FArgument) (*F, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.F", slog.Any("message_args", s.logvalue_Org_Federation_FArgument(req)))

	// create a message value to be returned.
	ret := &F{}

	// field binding section.
	ret.Name = "f" // (grpc.federation.field).string = "f"

	s.logger.DebugContext(ctx, "resolved org.federation.F", slog.Any("org.federation.F", s.logvalue_Org_Federation_F(ret)))
	return ret, nil
}

// resolve_Org_Federation_G resolve "org.federation.G" message.
func (s *FederationService) resolve_Org_Federation_G(ctx context.Context, req *Org_Federation_GArgument) (*G, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.G", slog.Any("message_args", s.logvalue_Org_Federation_GArgument(req)))

	// create a message value to be returned.
	ret := &G{}

	// field binding section.
	ret.Name = "g" // (grpc.federation.field).string = "g"

	s.logger.DebugContext(ctx, "resolved org.federation.G", slog.Any("org.federation.G", s.logvalue_Org_Federation_G(ret)))
	return ret, nil
}

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *Org_Federation_GetResponseArgument) (*GetResponse, error) {
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
	eg, egCtx := errgroup.WithContext(ctx)

	s.goWithRecover(eg, func() (interface{}, error) {
		eg, egCtx0 := errgroup.WithContext(egCtx)
		s.goWithRecover(eg, func() (interface{}, error) {
			eg, egCtx00 := errgroup.WithContext(egCtx0)
			s.goWithRecover(eg, func() (interface{}, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "a"
				     message: "A"
				   }
				*/
				resAIface, err, _ := sg.Do("a_org.federation.A", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_AArgument{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_A(egCtx00, args)
				})
				if err != nil {
					return nil, err
				}
				resA := resAIface.(*A)
				valueMu.Lock()
				valueA = resA // { name: "a", message: "A" ... }
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "c"
				     message: "C"
				     args { name: "a", by: "a.name" }
				   }
				*/
				resCIface, err, _ := sg.Do("c_org.federation.C", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_CArgument{
						Client: s.client,
						A:      valueA.GetName(), // { name: "a", by: "a.name" }
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_C(egCtx00, args)
				})
				if err != nil {
					return nil, err
				}
				resC := resCIface.(*C)
				valueMu.Lock()
				valueC = resC // { name: "c", message: "C" ... }
				valueMu.Unlock()
				return nil, nil
			})
			s.goWithRecover(eg, func() (interface{}, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "b"
				     message: "B"
				   }
				*/
				resBIface, err, _ := sg.Do("b_org.federation.B", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_BArgument{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_B(egCtx00, args)
				})
				if err != nil {
					return nil, err
				}
				resB := resBIface.(*B)
				valueMu.Lock()
				valueB = resB // { name: "b", message: "B" ... }
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "d"
				     message: "D"
				     args { name: "b", by: "b.name" }
				   }
				*/
				resDIface, err, _ := sg.Do("d_org.federation.D", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_DArgument{
						Client: s.client,
						B:      valueB.GetName(), // { name: "b", by: "b.name" }
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_D(egCtx00, args)
				})
				if err != nil {
					return nil, err
				}
				resD := resDIface.(*D)
				valueMu.Lock()
				valueD = resD // { name: "d", message: "D" ... }
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
			resEIface, err, _ := sg.Do("e_org.federation.E", func() (interface{}, error) {
				valueMu.RLock()
				args := &Org_Federation_EArgument{
					Client: s.client,
					C:      valueC.GetName(), // { name: "c", by: "c.name" }
					D:      valueD.GetName(), // { name: "d", by: "d.name" }
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_E(egCtx0, args)
			})
			if err != nil {
				return nil, err
			}
			resE := resEIface.(*E)
			valueMu.Lock()
			valueE = resE // { name: "e", message: "E" ... }
			valueMu.Unlock()
			return nil, nil
		})
		s.goWithRecover(eg, func() (interface{}, error) {
			eg, egCtx01 := errgroup.WithContext(egCtx0)
			s.goWithRecover(eg, func() (interface{}, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "a"
				     message: "A"
				   }
				*/
				resAIface, err, _ := sg.Do("a_org.federation.A", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_AArgument{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_A(egCtx01, args)
				})
				if err != nil {
					return nil, err
				}
				resA := resAIface.(*A)
				valueMu.Lock()
				valueA = resA // { name: "a", message: "A" ... }
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "c"
				     message: "C"
				     args { name: "a", by: "a.name" }
				   }
				*/
				resCIface, err, _ := sg.Do("c_org.federation.C", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_CArgument{
						Client: s.client,
						A:      valueA.GetName(), // { name: "a", by: "a.name" }
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_C(egCtx01, args)
				})
				if err != nil {
					return nil, err
				}
				resC := resCIface.(*C)
				valueMu.Lock()
				valueC = resC // { name: "c", message: "C" ... }
				valueMu.Unlock()
				return nil, nil
			})
			s.goWithRecover(eg, func() (interface{}, error) {

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "b"
				     message: "B"
				   }
				*/
				resBIface, err, _ := sg.Do("b_org.federation.B", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_BArgument{
						Client: s.client,
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_B(egCtx01, args)
				})
				if err != nil {
					return nil, err
				}
				resB := resBIface.(*B)
				valueMu.Lock()
				valueB = resB // { name: "b", message: "B" ... }
				valueMu.Unlock()

				// This section's codes are generated by the following proto definition.
				/*
				   {
				     name: "d"
				     message: "D"
				     args { name: "b", by: "b.name" }
				   }
				*/
				resDIface, err, _ := sg.Do("d_org.federation.D", func() (interface{}, error) {
					valueMu.RLock()
					args := &Org_Federation_DArgument{
						Client: s.client,
						B:      valueB.GetName(), // { name: "b", by: "b.name" }
					}
					valueMu.RUnlock()
					return s.resolve_Org_Federation_D(egCtx01, args)
				})
				if err != nil {
					return nil, err
				}
				resD := resDIface.(*D)
				valueMu.Lock()
				valueD = resD // { name: "d", message: "D" ... }
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
			resFIface, err, _ := sg.Do("f_org.federation.F", func() (interface{}, error) {
				valueMu.RLock()
				args := &Org_Federation_FArgument{
					Client: s.client,
					C:      valueC.GetName(), // { name: "c", by: "c.name" }
					D:      valueD.GetName(), // { name: "d", by: "d.name" }
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_F(egCtx0, args)
			})
			if err != nil {
				return nil, err
			}
			resF := resFIface.(*F)
			valueMu.Lock()
			valueF = resF // { name: "f", message: "F" ... }
			valueMu.Unlock()
			return nil, nil
		})
		s.goWithRecover(eg, func() (interface{}, error) {

			// This section's codes are generated by the following proto definition.
			/*
			   {
			     name: "g"
			     message: "G"
			   }
			*/
			resGIface, err, _ := sg.Do("g_org.federation.G", func() (interface{}, error) {
				valueMu.RLock()
				args := &Org_Federation_GArgument{
					Client: s.client,
				}
				valueMu.RUnlock()
				return s.resolve_Org_Federation_G(egCtx0, args)
			})
			if err != nil {
				return nil, err
			}
			resG := resGIface.(*G)
			valueMu.Lock()
			valueG = resG // { name: "g", message: "G" ... }
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
		resHIface, err, _ := sg.Do("h_org.federation.H", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_HArgument{
				Client: s.client,
				E:      valueE.GetName(), // { name: "e", by: "e.name" }
				F:      valueF.GetName(), // { name: "f", by: "f.name" }
				G:      valueG.GetName(), // { name: "g", by: "g.name" }
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_H(egCtx, args)
		})
		if err != nil {
			return nil, err
		}
		resH := resHIface.(*H)
		valueMu.Lock()
		valueH = resH // { name: "h", message: "H" ... }
		valueMu.Unlock()
		return nil, nil
	})

	s.goWithRecover(eg, func() (interface{}, error) {

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "i"
		     message: "I"
		   }
		*/
		resIIface, err, _ := sg.Do("i_org.federation.I", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_IArgument{
				Client: s.client,
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_I(egCtx, args)
		})
		if err != nil {
			return nil, err
		}
		resI := resIIface.(*I)
		valueMu.Lock()
		valueI = resI // { name: "i", message: "I" ... }
		valueMu.Unlock()

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "j"
		     message: "J"
		     args { name: "i", by: "i.name" }
		   }
		*/
		resJIface, err, _ := sg.Do("j_org.federation.J", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_JArgument{
				Client: s.client,
				I:      valueI.GetName(), // { name: "i", by: "i.name" }
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_J(egCtx, args)
		})
		if err != nil {
			return nil, err
		}
		resJ := resJIface.(*J)
		valueMu.Lock()
		valueJ = resJ // { name: "j", message: "J" ... }
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
	ret.Hname = valueH.GetName() // (grpc.federation.field).by = "h.name"
	ret.Jname = valueJ.GetName() // (grpc.federation.field).by = "j.name"

	s.logger.DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_H resolve "org.federation.H" message.
func (s *FederationService) resolve_Org_Federation_H(ctx context.Context, req *Org_Federation_HArgument) (*H, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.H", slog.Any("message_args", s.logvalue_Org_Federation_HArgument(req)))

	// create a message value to be returned.
	ret := &H{}

	// field binding section.
	ret.Name = "h" // (grpc.federation.field).string = "h"

	s.logger.DebugContext(ctx, "resolved org.federation.H", slog.Any("org.federation.H", s.logvalue_Org_Federation_H(ret)))
	return ret, nil
}

// resolve_Org_Federation_I resolve "org.federation.I" message.
func (s *FederationService) resolve_Org_Federation_I(ctx context.Context, req *Org_Federation_IArgument) (*I, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.I", slog.Any("message_args", s.logvalue_Org_Federation_IArgument(req)))

	// create a message value to be returned.
	ret := &I{}

	// field binding section.
	ret.Name = "i" // (grpc.federation.field).string = "i"

	s.logger.DebugContext(ctx, "resolved org.federation.I", slog.Any("org.federation.I", s.logvalue_Org_Federation_I(ret)))
	return ret, nil
}

// resolve_Org_Federation_J resolve "org.federation.J" message.
func (s *FederationService) resolve_Org_Federation_J(ctx context.Context, req *Org_Federation_JArgument) (*J, error) {
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

func (s *FederationService) logvalue_Org_Federation_AAArgument(v *Org_Federation_AAArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_AArgument(v *Org_Federation_AArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_ABArgument(v *Org_Federation_ABArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_BArgument(v *Org_Federation_BArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_CArgument(v *Org_Federation_CArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_DArgument(v *Org_Federation_DArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_EArgument(v *Org_Federation_EArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_FArgument(v *Org_Federation_FArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_GArgument(v *Org_Federation_GArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *Org_Federation_GetResponseArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_HArgument(v *Org_Federation_HArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_IArgument(v *Org_Federation_IArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_JArgument(v *Org_Federation_JArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("i", v.I),
	)
}
