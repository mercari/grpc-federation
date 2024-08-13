package federation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"golang.org/x/sync/singleflight"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
)

// CELTypeHelper provides the cel.Registry needed to build a cel environment.
type CELTypeHelper struct {
	*celtypes.Registry
	structFieldMap map[string]map[string]*celtypes.FieldType
	pkgName        string
	mapMu          sync.RWMutex
	mu             sync.Mutex
}

func (h *CELTypeHelper) RegisterType(types ...ref.Type) error {
	h.mu.Lock()
	err := h.Registry.RegisterType(types...)
	h.mu.Unlock()
	return err
}

func (h *CELTypeHelper) CELRegistry() *celtypes.Registry {
	return h.Registry
}

func (h *CELTypeHelper) TypeProvider() celtypes.Provider {
	return h
}

func (h *CELTypeHelper) TypeAdapter() celtypes.Adapter {
	return h.Registry
}

func (h *CELTypeHelper) EnumValue(enumName string) ref.Val {
	return h.Registry.EnumValue(enumName)
}

func (h *CELTypeHelper) FindIdent(identName string) (ref.Val, bool) {
	return h.Registry.FindIdent(identName)
}

func (h *CELTypeHelper) FindStructType(structType string) (*celtypes.Type, bool) {
	if st, found := h.Registry.FindStructType(structType); found {
		return st, found
	}
	h.mapMu.RLock()
	defer h.mapMu.RUnlock()
	if _, exists := h.structFieldMap[structType]; exists {
		return celtypes.NewObjectType(structType), true
	}
	return nil, false
}

func (h *CELTypeHelper) FindStructFieldNames(structType string) ([]string, bool) {
	if names, found := h.Registry.FindStructFieldNames(structType); found {
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

func (h *CELTypeHelper) FindStructFieldType(structType, fieldName string) (*celtypes.FieldType, bool) {
	if field, found := h.Registry.FindStructFieldType(structType, fieldName); found {
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

func (h *CELTypeHelper) NewValue(structType string, fields map[string]ref.Val) ref.Val {
	return h.Registry.NewValue(structType, fields)
}

func NewCELFieldType(typ *celtypes.Type, fieldName string) *celtypes.FieldType {
	isSet := func(v any, fieldName string) bool {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return false
		}
		return !rv.FieldByName(fieldName).IsZero()
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

func NewOneofSelectorFieldType(typ *celtypes.Type, fieldName string, oneofTypes []reflect.Type, getterNames []string, zeroValue reflect.Value) *celtypes.FieldType {
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
			if fieldImpl.Type() != oneofType {
				continue
			}
			method := reflect.ValueOf(v).MethodByName(getterNames[idx])
			retValues := method.Call(nil)
			if len(retValues) != 1 {
				return nil, fmt.Errorf("failed to call %s for %T", "", v)
			}
			retValue := retValues[0]
			return retValue.Interface(), nil
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

type CELTypeHelperFieldMap map[string]map[string]*celtypes.FieldType

func NewCELTypeHelper(pkgName string, structFieldMap CELTypeHelperFieldMap) *CELTypeHelper {
	celRegistry := celtypes.NewEmptyRegistry()
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if err := celRegistry.RegisterDescriptor(f); err != nil {
			return false
		}
		return true
	})
	return &CELTypeHelper{
		Registry:       celRegistry,
		structFieldMap: structFieldMap,
		pkgName:        pkgName,
	}
}

func EnumAccessorOptions(enumName string, nameToValue map[string]int32, valueToName map[int32]string) []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function(
			fmt.Sprintf("%s.name", enumName),
			cel.Overload(fmt.Sprintf("%s_name_int_string", enumName), []*cel.Type{cel.IntType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return celtypes.String(valueToName[int32(self.(celtypes.Int))])
				}),
			),
		),
		cel.Function(
			fmt.Sprintf("%s.value", enumName),
			cel.Overload(fmt.Sprintf("%s_value_string_int", enumName), []*cel.Type{cel.StringType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return celtypes.Int(nameToValue[string(self.(celtypes.String))])
				}),
			),
		),
		cel.Function(
			fmt.Sprintf("%s.from", enumName),
			cel.Overload(fmt.Sprintf("%[1]s_name_int_%[1]s", enumName), []*cel.Type{cel.IntType}, celtypes.NewOpaqueType(enumName, cel.IntType),
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return self
				}),
			),
		),
	}
}

func NewDefaultEnvOptions(celHelper *CELTypeHelper) []cel.EnvOption {
	return []cel.EnvOption{
		cel.StdLib(),
		cel.Lib(grpcfedcel.NewLibrary(celHelper)),
		cel.CrossTypeNumericComparisons(true),
		cel.CustomTypeAdapter(celHelper.TypeAdapter()),
		cel.CustomTypeProvider(celHelper.TypeProvider()),
		cel.Variable("error", cel.ObjectType("grpc.federation.private.Error")),
		cel.Variable(ContextVariableName, cel.ObjectType(ContextTypeName)),
		cel.Container(celHelper.pkgName),
	}
}

// CELCache used to speed up CEL evaluation from the second time onward.
// cel.Program cannot be reused to evaluate contextual libraries or plugins, so cel.Ast is reused to speed up the process.
type CELCache struct {
	program cel.Program // cache for simple expressions.
}

// CELCacheMap service-wide in-memory cache store for CEL evaluation.
// The cache key is a constant value created by code-generation.
type CELCacheMap struct {
	mu       sync.RWMutex
	cacheMap map[int]*CELCache
}

// NewCELCacheMap creates CELCacheMap instance.
func NewCELCacheMap() *CELCacheMap {
	return &CELCacheMap{
		cacheMap: make(map[int]*CELCache),
	}
}

func (m *CELCacheMap) get(index int) *CELCache {
	m.mu.RLock()
	cache := m.cacheMap[index]
	m.mu.RUnlock()
	return cache
}

func (m *CELCacheMap) set(index int, cache *CELCache) {
	m.mu.Lock()
	m.cacheMap[index] = cache
	m.mu.Unlock()
}

type LocalValue struct {
	sg         singleflight.Group
	mu         sync.RWMutex
	envOpts    []cel.EnvOption
	evalValues map[string]any
}

func NewLocalValue(ctx context.Context, envOpts []cel.EnvOption, argName string, arg any) *LocalValue {
	var newEnvOpts []cel.EnvOption
	newEnvOpts = append(
		append(newEnvOpts, envOpts...),
		cel.Variable(MessageArgumentVariableName, cel.ObjectType(argName)),
	)
	return &LocalValue{
		envOpts:    newEnvOpts,
		evalValues: map[string]any{MessageArgumentVariableName: arg},
	}
}

func (v *LocalValue) AddEnv(env any) {
	v.setEvalValue("grpc.federation.env", env)
}

type localValue interface {
	do(string, func() (any, error)) (any, error)
	rlock()
	runlock()
	lock()
	unlock()
	getEnvOpts() []cel.EnvOption
	getEvalValues(context.Context) map[string]any
	setEnvOptValue(string, *cel.Type)
	setEvalValue(string, any)
}

func (v *LocalValue) do(name string, cb func() (any, error)) (any, error) {
	ret, err, _ := v.sg.Do(name, cb)
	return ret, err
}

func (v *LocalValue) rlock() {
	v.mu.RLock()
}

func (v *LocalValue) runlock() {
	v.mu.RUnlock()
}

func (v *LocalValue) lock() {
	v.mu.Lock()
}

func (v *LocalValue) unlock() {
	v.mu.Unlock()
}

func (v *LocalValue) getEnvOpts() []cel.EnvOption {
	return v.envOpts
}

func (v *LocalValue) getEvalValues(ctx context.Context) map[string]any {
	ret := map[string]any{ContextVariableName: grpcfedcel.NewContextValue(ctx)}
	for k, v := range v.evalValues {
		ret[k] = v
	}
	return ret
}

func (v *LocalValue) setEnvOptValue(name string, typ *cel.Type) {
	v.envOpts = append(
		v.envOpts,
		cel.Variable(name, typ),
	)
}

func (v *LocalValue) setEvalValue(name string, value any) {
	v.evalValues[name] = value
}

type MapIteratorValue struct {
	localValue localValue
	envOpts    []cel.EnvOption
	evalValues map[string]any
}

func (v *MapIteratorValue) do(name string, cb func() (any, error)) (any, error) {
	return v.localValue.do(name, cb)
}

func (v *MapIteratorValue) rlock() {
	v.localValue.rlock()
}

func (v *MapIteratorValue) runlock() {
	v.localValue.runlock()
}

func (v *MapIteratorValue) lock() {
	v.localValue.lock()
}

func (v *MapIteratorValue) unlock() {
	v.localValue.unlock()
}

func (v *MapIteratorValue) getEnvOpts() []cel.EnvOption {
	return v.envOpts
}

func (v *MapIteratorValue) getEvalValues(ctx context.Context) map[string]any {
	ret := map[string]any{ContextVariableName: grpcfedcel.NewContextValue(ctx)}
	for k, v := range v.evalValues {
		ret[k] = v
	}
	return ret
}

func (v *MapIteratorValue) setEnvOptValue(name string, typ *cel.Type) {
	v.envOpts = append(
		v.envOpts,
		cel.Variable(name, typ),
	)
}

func (v *MapIteratorValue) setEvalValue(name string, value any) {
	v.evalValues[name] = value
}

type Def[T any, U localValue] struct {
	If           string
	Name         string
	Type         *cel.Type
	Setter       func(U, T) error
	By           string
	IfCacheIndex int
	ByCacheIndex int
	Message      func(context.Context, U) (any, error)
	Validation   func(context.Context, U) error
}

type DefMap[T any, U any, V localValue] struct {
	If             string
	IfCacheIndex   int
	Name           string
	Type           *cel.Type
	Setter         func(V, T) error
	IteratorName   string
	IteratorType   *cel.Type
	IteratorSource func(V) []U
	Iterator       func(context.Context, *MapIteratorValue) (any, error)
	outType        T
}

func EvalDef[T any, U localValue](ctx context.Context, value U, def Def[T, U]) error {
	var (
		v    T
		errs []error
		cond = true
		name = def.Name
	)
	if def.If != "" {
		c, err := EvalCEL(ctx, &EvalCELRequest{
			Value:      value,
			Expr:       def.If,
			OutType:    reflect.TypeOf(false),
			CacheIndex: def.IfCacheIndex,
		})
		if err != nil {
			return err
		}
		if !c.(bool) {
			cond = false
		}
	}

	if cond {
		ret, runErr := value.do(name, func() (any, error) {
			switch {
			case def.By != "":
				return EvalCEL(ctx, &EvalCELRequest{
					Value:      value,
					Expr:       def.By,
					OutType:    reflect.TypeOf(v),
					CacheIndex: def.ByCacheIndex,
				})
			case def.Message != nil:
				return def.Message(ctx, value)
			case def.Validation != nil:
				if err := def.Validation(ctx, value); err != nil {
					if _, ok := grpcstatus.FromError(err); ok {
						return nil, err
					}
					Logger(ctx).ErrorContext(ctx, "failed running validations", slog.String("error", err.Error()))
					return nil, grpcstatus.Errorf(grpccodes.Internal, "failed running validations: %s", err)
				}
			}
			return nil, nil
		})
		if ret != nil {
			v = ret.(T)
		}
		if runErr != nil {
			errs = append(errs, runErr)
		}
	}

	value.lock()
	if err := def.Setter(value, v); err != nil {
		errs = append(errs, err)
	}
	value.setEnvOptValue(name, def.Type)
	value.setEvalValue(name, v)
	value.unlock()

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}
	return nil
}

func EvalDefMap[T any, U any, V localValue](ctx context.Context, value V, def DefMap[T, U, V]) error {
	var (
		v    T
		errs []error
		cond = true
		name = def.Name
	)
	if def.If != "" {
		c, err := EvalCEL(ctx, &EvalCELRequest{
			Value:      value,
			Expr:       def.If,
			OutType:    reflect.TypeOf(false),
			CacheIndex: def.IfCacheIndex,
		})
		if err != nil {
			return err
		}
		if !c.(bool) {
			cond = false
		}
	}

	if cond {
		ret, runErr := value.do(name, func() (any, error) {
			return evalMap(
				ctx,
				value,
				def.IteratorName,
				def.IteratorType,
				def.IteratorSource,
				reflect.TypeOf(def.outType),
				def.Iterator,
			)
		})
		if ret != nil {
			v = ret.(T)
		}
		if runErr != nil {
			errs = append(errs, runErr)
		}
	}

	value.lock()
	if err := def.Setter(value, v); err != nil {
		errs = append(errs, err)
	}
	value.setEnvOptValue(name, def.Type)
	value.setEvalValue(name, v)
	value.unlock()

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}
	return nil
}

func evalMap[T localValue, U any](
	ctx context.Context,
	value T,
	name string,
	typ *cel.Type,
	srcFunc func(T) []U,
	iterOutType reflect.Type,
	cb func(context.Context, *MapIteratorValue) (any, error)) (any, error) {
	value.rlock()
	iterValue := &MapIteratorValue{
		localValue: value,
		evalValues: make(map[string]any),
	}
	envOpts := value.getEnvOpts()
	for k, v := range value.getEvalValues(ctx) {
		iterValue.evalValues[k] = v
	}
	src := srcFunc(value)
	value.runlock()

	ret := reflect.MakeSlice(iterOutType, 0, len(src))
	for _, iter := range src {
		iterValue.envOpts = append(append([]cel.EnvOption{}, envOpts...), cel.Variable(name, typ))
		iterValue.evalValues[name] = iter
		v, err := cb(ctx, iterValue)
		if err != nil {
			return nil, err
		}
		ret = reflect.Append(ret, reflect.ValueOf(v))
	}
	return ret.Interface(), nil
}

type IfParam[T localValue] struct {
	Value      T
	Expr       string
	CacheIndex int
	Body       func(T) error
}

func If[T localValue](ctx context.Context, param *IfParam[T]) error {
	cond, err := EvalCEL(ctx, &EvalCELRequest{
		Value:      param.Value,
		Expr:       param.Expr,
		OutType:    reflect.TypeOf(false),
		CacheIndex: param.CacheIndex,
	})
	if err != nil {
		return err
	}
	if cond.(bool) {
		return param.Body(param.Value)
	}
	return nil
}

type EvalCELRequest struct {
	Value      localValue
	Expr       string
	OutType    reflect.Type
	CacheIndex int
}

func EvalCEL(ctx context.Context, req *EvalCELRequest) (any, error) {
	req.Value.lock()
	defer req.Value.unlock()

	if celCacheMap := getCELCacheMap(ctx); celCacheMap == nil {
		return nil, ErrCELCacheMap
	}
	if req.CacheIndex == 0 {
		return nil, ErrCELCacheIndex
	}

	program, err := createCELProgram(ctx, req)
	if err != nil {
		return nil, err
	}
	out, _, err := program.ContextEval(ctx, req.Value.getEvalValues(ctx))
	if err != nil {
		return nil, err
	}
	opt, ok := out.(*celtypes.Optional)
	if ok {
		if opt == celtypes.OptionalNone {
			return reflect.Zero(req.OutType).Interface(), nil
		}
		out = opt.GetValue()
	}
	if _, ok := out.(celtypes.Null); ok {
		if req.OutType == nil {
			return nil, nil
		}
		return reflect.Zero(req.OutType).Interface(), nil
	}
	if req.OutType != nil {
		return out.ConvertToNative(req.OutType)
	}
	return out.Value(), nil
}

func createCELProgram(ctx context.Context, req *EvalCELRequest) (cel.Program, error) {
	if program := getCELProgramCache(ctx, req.CacheIndex); program != nil {
		return program, nil
	}

	env, err := NewCELEnv(req.Value.getEnvOpts()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cel env: %w", err)
	}
	ast, err := createCELAst(req, env)
	if err != nil {
		return nil, err
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	setCELProgramCache(ctx, req.CacheIndex, program)
	return program, nil
}

func createCELAst(req *EvalCELRequest, env *cel.Env) (*cel.Ast, error) {
	expr := strings.Replace(req.Expr, "$", MessageArgumentVariableName, -1)
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return ast, nil
}

func getCELProgramCache(ctx context.Context, cacheIndex int) cel.Program {
	cache := getCELCacheMap(ctx).get(cacheIndex)
	if cache == nil {
		return nil
	}
	return cache.program
}

func setCELProgramCache(ctx context.Context, cacheIndex int, program cel.Program) {
	getCELCacheMap(ctx).set(cacheIndex, &CELCache{program: program})
}

func SetGRPCError(ctx context.Context, value localValue, err error) {
	stat, ok := grpcstatus.FromError(err)
	if !ok {
		return
	}
	grpcErr := &grpcfedcel.Error{
		Code:    int32(stat.Code()),
		Message: stat.Message(),
	}
	for _, detail := range stat.Details() {
		protoMsg, ok := detail.(proto.Message)
		if !ok {
			Logger(ctx).ErrorContext(
				ctx,
				"failed to convert error detail to proto message",
				slog.String("detail", fmt.Sprintf("%T", detail)),
			)
			continue
		}
		anyValue, err := anypb.New(protoMsg)
		if err != nil {
			Logger(ctx).ErrorContext(
				ctx,
				"failed to create proto.Any instance from proto message",
				slog.String("proto", fmt.Sprintf("%T", protoMsg)),
			)
			continue
		}
		grpcErr.Details = append(grpcErr.Details, anyValue)
		switch m := protoMsg.(type) {
		case *errdetails.ErrorInfo:
			grpcErr.ErrorInfo = append(grpcErr.ErrorInfo, m)
		case *errdetails.RetryInfo:
			grpcErr.RetryInfo = append(grpcErr.RetryInfo, m)
		case *errdetails.DebugInfo:
			grpcErr.DebugInfo = append(grpcErr.DebugInfo, m)
		case *errdetails.QuotaFailure:
			grpcErr.QuotaFailures = append(grpcErr.QuotaFailures, m)
		case *errdetails.PreconditionFailure:
			grpcErr.PreconditionFailures = append(grpcErr.PreconditionFailures, m)
		case *errdetails.BadRequest:
			grpcErr.BadRequests = append(grpcErr.BadRequests, m)
		case *errdetails.RequestInfo:
			grpcErr.RequestInfo = append(grpcErr.RequestInfo, m)
		case *errdetails.ResourceInfo:
			grpcErr.ResourceInfo = append(grpcErr.ResourceInfo, m)
		case *errdetails.Help:
			grpcErr.Helps = append(grpcErr.Helps, m)
		case *errdetails.LocalizedMessage:
			grpcErr.LocalizedMessages = append(grpcErr.LocalizedMessages, m)
		default:
			grpcErr.CustomMessages = append(grpcErr.CustomMessages, anyValue)
		}
	}
	value.lock()
	value.setEvalValue("error", grpcErr)
	value.unlock()
}

type SetCELValueParam[T any] struct {
	Value      localValue
	Expr       string
	CacheIndex int
	Setter     func(T) error
}

func SetCELValue[T any](ctx context.Context, param *SetCELValueParam[T]) error {
	var typ T
	out, err := EvalCEL(ctx, &EvalCELRequest{
		Value:      param.Value,
		Expr:       param.Expr,
		OutType:    reflect.TypeOf(typ),
		CacheIndex: param.CacheIndex,
	})
	if err != nil {
		return err
	}

	param.Value.lock()
	defer param.Value.unlock()

	if err := param.Setter(out.(T)); err != nil {
		return err
	}
	return nil
}
