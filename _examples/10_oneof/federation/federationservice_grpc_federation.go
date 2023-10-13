// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
package federation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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
	env          *cel.Env
	client       *FederationServiceDependencyServiceClient
}

// Org_Federation_GetResponseArgument is argument for "org.federation.GetResponse" message.
type Org_Federation_GetResponseArgument struct {
	Sel    *UserSelection
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_MArgument is argument for "org.federation.M" message.
type Org_Federation_MArgument struct {
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_UserArgument is argument for "org.federation.User" message.
type Org_Federation_UserArgument struct {
	UserId string
	Client *FederationServiceDependencyServiceClient
}

// Org_Federation_UserSelectionArgument is argument for "org.federation.UserSelection" message.
type Org_Federation_UserSelectionArgument struct {
	M      *M
	Value  string
	Client *FederationServiceDependencyServiceClient
}

// FederationServiceCELTypeHelper
type FederationServiceCELTypeHelper struct {
	celRegistry    *celtypes.Registry
	structFieldMap map[string]map[string]*celtypes.FieldType
	mapMu          sync.RWMutex
}

func (h *FederationServiceCELTypeHelper) TypeProvider() celtypes.Provider {
	return h
}

func (h *FederationServiceCELTypeHelper) TypeAdapter() celtypes.Adapter {
	return h.celRegistry
}

func (h *FederationServiceCELTypeHelper) EnumValue(enumName string) ref.Val {
	return h.celRegistry.EnumValue(enumName)
}

func (h *FederationServiceCELTypeHelper) FindIdent(identName string) (ref.Val, bool) {
	return h.celRegistry.FindIdent(identName)
}

func (h *FederationServiceCELTypeHelper) FindStructType(structType string) (*celtypes.Type, bool) {
	if st, found := h.celRegistry.FindStructType(structType); found {
		return st, found
	}
	h.mapMu.RLock()
	defer h.mapMu.RUnlock()
	if _, exists := h.structFieldMap[structType]; exists {
		return celtypes.NewObjectType(structType), true
	}
	return nil, false
}

func (h *FederationServiceCELTypeHelper) FindStructFieldNames(structType string) ([]string, bool) {
	if names, found := h.celRegistry.FindStructFieldNames(structType); found {
		return names, found
	}

	h.mapMu.RLock()
	defer h.mapMu.RUnlock()
	fieldMap, exists := h.structFieldMap[structType]
	if !exists {
		return nil, false
	}
	fieldNames := make([]string, 0, len(fieldMap))
	for fieldName := range fieldMap {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)
	return fieldNames, true
}

func (h *FederationServiceCELTypeHelper) FindStructFieldType(structType, fieldName string) (*celtypes.FieldType, bool) {
	if field, found := h.celRegistry.FindStructFieldType(structType, fieldName); found {
		return field, found
	}

	h.mapMu.RLock()
	defer h.mapMu.RUnlock()
	fieldMap, exists := h.structFieldMap[structType]
	if !exists {
		return nil, false
	}
	field, found := fieldMap[fieldName]
	return field, found
}

func (h *FederationServiceCELTypeHelper) NewValue(structType string, fields map[string]ref.Val) ref.Val {
	return h.celRegistry.NewValue(structType, fields)
}

func newFederationServiceCELTypeHelper() *FederationServiceCELTypeHelper {
	celRegistry := celtypes.NewEmptyRegistry()
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if err := celRegistry.RegisterDescriptor(f); err != nil {
			return false
		}
		return true
	})
	newFieldType := func(typ *celtypes.Type, fieldName string) *celtypes.FieldType {
		isSet := func(v any, fieldName string) bool {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return false
			}
			return rv.FieldByName(fieldName).IsValid()
		}
		getFrom := func(v any, fieldName string) (any, error) {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return nil, fmt.Errorf("%T is not struct type", v)
			}
			value := rv.FieldByName(fieldName)
			return value.Interface(), nil
		}
		return &celtypes.FieldType{
			Type: typ,
			IsSet: func(v any) bool {
				return isSet(v, fieldName)
			},
			GetFrom: func(v any) (any, error) {
				return getFrom(v, fieldName)
			},
		}
	}
	newOneofSelectorFieldType := func(typ *celtypes.Type, fieldName string, oneofTypes []reflect.Type, getterNames []string, zeroValue reflect.Value) *celtypes.FieldType {
		isSet := func(_ any) bool {
			return false
		}
		getFrom := func(v any) (any, error) {
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return nil, fmt.Errorf("%T is not struct type", v)
			}
			field := rv.FieldByName(fieldName)
			fieldImpl := reflect.ValueOf(field.Interface())
			for idx, oneofType := range oneofTypes {
				if fieldImpl.Type() == oneofType {
					method := reflect.ValueOf(v).MethodByName(getterNames[idx])
					retValues := method.Call(nil)
					if len(retValues) != 1 {
						return nil, fmt.Errorf("failed to call %s for %T", "", v)
					}
					retValue := retValues[0]
					return retValue.Interface(), nil
				}
			}
			return zeroValue.Interface(), nil
		}
		return &celtypes.FieldType{
			Type: typ,
			IsSet: func(v any) bool {
				return isSet(v)
			},
			GetFrom: func(v any) (any, error) {
				return getFrom(v)
			},
		}
	}
	return &FederationServiceCELTypeHelper{
		celRegistry: celRegistry,
		structFieldMap: map[string]map[string]*celtypes.FieldType{
			"grpc.federation.private.GetResponseArgument": map[string]*celtypes.FieldType{},
			"grpc.federation.private.MArgument":           map[string]*celtypes.FieldType{},
			"grpc.federation.private.UserArgument": map[string]*celtypes.FieldType{
				"user_id": newFieldType(celtypes.StringType, "UserId"),
			},
			"grpc.federation.private.UserSelectionArgument": map[string]*celtypes.FieldType{
				"value": newFieldType(celtypes.StringType, "Value"),
			},
			"org.federation.UserSelection": map[string]*celtypes.FieldType{
				"user": newOneofSelectorFieldType(
					celtypes.NewObjectType("org.federation.User"), "User",
					[]reflect.Type{reflect.TypeOf((*UserSelection_UserA)(nil)), reflect.TypeOf((*UserSelection_UserB)(nil))},
					[]string{"GetUserA", "GetUserB"},
					reflect.Zero(reflect.TypeOf((*User)(nil))),
				),
			},
		},
	}
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
	celHelper := newFederationServiceCELTypeHelper()
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

func (s *FederationService) evalCEL(expr string, vars []cel.EnvOption, args map[string]any, outType reflect.Type) (any, error) {
	env, err := s.env.Extend(vars...)
	if err != nil {
		return nil, err
	}
	expr = strings.Replace(expr, "$", grpcfed.MessageArgumentVariableName, -1)
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	out, _, err := program.Eval(args)
	if err != nil {
		return nil, err
	}
	if outType != nil {
		return out.ConvertToNative(outType)
	}
	return out.Value(), nil
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

// resolve_Org_Federation_GetResponse resolve "org.federation.GetResponse" message.
func (s *FederationService) resolve_Org_Federation_GetResponse(ctx context.Context, req *Org_Federation_GetResponseArgument) (*GetResponse, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.GetResponse", slog.Any("message_args", s.logvalue_Org_Federation_GetResponseArgument(req)))
	var (
		sg       singleflight.Group
		valueMu  sync.RWMutex
		valueSel *UserSelection
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.GetResponseArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

	// This section's codes are generated by the following proto definition.
	/*
	   {
	     name: "sel"
	     message: "UserSelection"
	     args { name: "value", string: "foo" }
	   }
	*/
	resUserSelectionIface, err, _ := sg.Do("sel_org.federation.UserSelection", func() (interface{}, error) {
		valueMu.RLock()
		args := &Org_Federation_UserSelectionArgument{
			Client: s.client,
			Value:  "foo", // { name: "value", string: "foo" }
		}
		valueMu.RUnlock()
		return s.resolve_Org_Federation_UserSelection(ctx, args)
	})
	if err != nil {
		return nil, err
	}
	resUserSelection := resUserSelectionIface.(*UserSelection)
	valueMu.Lock()
	valueSel = resUserSelection // { name: "sel", message: "UserSelection" ... }
	envOpts = append(envOpts, cel.Variable("sel", cel.ObjectType("org.federation.UserSelection")))
	evalValues["sel"] = valueSel
	valueMu.Unlock()

	// assign named parameters to message arguments to pass to the custom resolver.
	req.Sel = valueSel

	// create a message value to be returned.
	ret := &GetResponse{}

	// field binding section.
	// (grpc.federation.field).by = "sel.user"
	{
		_value, err := s.evalCEL("sel.user", envOpts, evalValues, nil)
		if err != nil {
			return nil, err
		}
		ret.User = _value.(*User)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.GetResponse", slog.Any("org.federation.GetResponse", s.logvalue_Org_Federation_GetResponse(ret)))
	return ret, nil
}

// resolve_Org_Federation_M resolve "org.federation.M" message.
func (s *FederationService) resolve_Org_Federation_M(ctx context.Context, req *Org_Federation_MArgument) (*M, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.M", slog.Any("message_args", s.logvalue_Org_Federation_MArgument(req)))

	// create a message value to be returned.
	ret := &M{}

	// field binding section.
	ret.Value = "foo" // (grpc.federation.field).string = "foo"

	s.logger.DebugContext(ctx, "resolved org.federation.M", slog.Any("org.federation.M", s.logvalue_Org_Federation_M(ret)))
	return ret, nil
}

// resolve_Org_Federation_User resolve "org.federation.User" message.
func (s *FederationService) resolve_Org_Federation_User(ctx context.Context, req *Org_Federation_UserArgument) (*User, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.User", slog.Any("message_args", s.logvalue_Org_Federation_UserArgument(req)))
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.UserArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

	// create a message value to be returned.
	ret := &User{}

	// field binding section.
	// (grpc.federation.field).by = "$.user_id"
	{
		_value, err := s.evalCEL("$.user_id", envOpts, evalValues, reflect.TypeOf(ret.Id))
		if err != nil {
			return nil, err
		}
		ret.Id = _value.(string)
	}

	s.logger.DebugContext(ctx, "resolved org.federation.User", slog.Any("org.federation.User", s.logvalue_Org_Federation_User(ret)))
	return ret, nil
}

// resolve_Org_Federation_UserSelection resolve "org.federation.UserSelection" message.
func (s *FederationService) resolve_Org_Federation_UserSelection(ctx context.Context, req *Org_Federation_UserSelectionArgument) (*UserSelection, error) {
	s.logger.DebugContext(ctx, "resolve  org.federation.UserSelection", slog.Any("message_args", s.logvalue_Org_Federation_UserSelectionArgument(req)))
	var (
		sg      singleflight.Group
		valueM  *M
		valueMu sync.RWMutex
		valueUa *User
		valueUb *User
	)
	envOpts := []cel.EnvOption{cel.Variable(grpcfed.MessageArgumentVariableName, cel.ObjectType("grpc.federation.private.UserSelectionArgument"))}
	evalValues := map[string]any{grpcfed.MessageArgumentVariableName: req}

	// This section's codes are generated by the following proto definition.
	/*
	   {
	     name: "m"
	     message: "M"
	   }
	*/
	resMIface, err, _ := sg.Do("m_org.federation.M", func() (interface{}, error) {
		valueMu.RLock()
		args := &Org_Federation_MArgument{
			Client: s.client,
		}
		valueMu.RUnlock()
		return s.resolve_Org_Federation_M(ctx, args)
	})
	if err != nil {
		return nil, err
	}
	resM := resMIface.(*M)
	valueMu.Lock()
	valueM = resM // { name: "m", message: "M" ... }
	envOpts = append(envOpts, cel.Variable("m", cel.ObjectType("org.federation.M")))
	evalValues["m"] = valueM
	valueMu.Unlock()

	// assign named parameters to message arguments to pass to the custom resolver.
	req.M = valueM

	// create a message value to be returned.
	ret := &UserSelection{}

	// field binding section.

	oneof_UserA, err := s.evalCEL("m.value != $.value", envOpts, evalValues, reflect.TypeOf(true))
	if err != nil {
		return nil, err
	}
	oneof_UserB, err := s.evalCEL("m.value == $.value", envOpts, evalValues, reflect.TypeOf(true))
	if err != nil {
		return nil, err
	}
	switch {
	case oneof_UserA.(bool):

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "ua"
		     message: "User"
		     args { name: "user_id", string: "a" }
		   }
		*/
		resUserIface, err, _ := sg.Do("ua_org.federation.User", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_UserArgument{
				Client: s.client,
				UserId: "a", // { name: "user_id", string: "a" }
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_User(ctx, args)
		})
		if err != nil {
			return nil, err
		}
		resUser := resUserIface.(*User)
		valueMu.Lock()
		valueUa = resUser // { name: "ua", message: "User" ... }
		envOpts = append(envOpts, cel.Variable("ua", cel.ObjectType("org.federation.User")))
		evalValues["ua"] = valueUa
		valueMu.Unlock()
		_value, err := s.evalCEL("ua", envOpts, evalValues, nil)
		if err != nil {
			return nil, err
		}
		ret.User = &UserSelection_UserA{UserA: _value.(*User)}
	case oneof_UserB.(bool):

		// This section's codes are generated by the following proto definition.
		/*
		   {
		     name: "ub"
		     message: "User"
		     args { name: "user_id", string: "b" }
		   }
		*/
		resUserIface, err, _ := sg.Do("ub_org.federation.User", func() (interface{}, error) {
			valueMu.RLock()
			args := &Org_Federation_UserArgument{
				Client: s.client,
				UserId: "b", // { name: "user_id", string: "b" }
			}
			valueMu.RUnlock()
			return s.resolve_Org_Federation_User(ctx, args)
		})
		if err != nil {
			return nil, err
		}
		resUser := resUserIface.(*User)
		valueMu.Lock()
		valueUb = resUser // { name: "ub", message: "User" ... }
		envOpts = append(envOpts, cel.Variable("ub", cel.ObjectType("org.federation.User")))
		evalValues["ub"] = valueUb
		valueMu.Unlock()
		_value, err := s.evalCEL("ub", envOpts, evalValues, nil)
		if err != nil {
			return nil, err
		}
		ret.User = &UserSelection_UserB{UserB: _value.(*User)}
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

func (s *FederationService) logvalue_Org_Federation_GetResponseArgument(v *Org_Federation_GetResponseArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue()
}

func (s *FederationService) logvalue_Org_Federation_M(v *M) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.GetValue()),
	)
}

func (s *FederationService) logvalue_Org_Federation_MArgument(v *Org_Federation_MArgument) slog.Value {
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

func (s *FederationService) logvalue_Org_Federation_UserArgument(v *Org_Federation_UserArgument) slog.Value {
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
	)
}

func (s *FederationService) logvalue_Org_Federation_UserSelectionArgument(v *Org_Federation_UserSelectionArgument) slog.Value {
	if v == nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(
		slog.String("value", v.Value),
	)
}
