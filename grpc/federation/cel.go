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
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
	"github.com/google/cel-go/parser"
	"golang.org/x/sync/singleflight"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/genproto/googleapis/rpc/code"
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
		if !fieldImpl.IsValid() {
			// prevent panic if no value assigned
			return nil, fmt.Errorf("%s is invalid field", fieldName)
		}
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
					return celtypes.String(valueToName[int32(self.(celtypes.Int))]) //nolint:gosec
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
			cel.Overload(fmt.Sprintf("%s_from_int_int", enumName), []*cel.Type{cel.IntType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return self
				}),
			),
		),
	}
}

func GRPCErrorAccessorOptions(adapter celtypes.Adapter, protoName string) []cel.EnvOption {
	var opts []cel.EnvOption
	opts = append(opts,
		grpcfedcel.BindMemberFunction(
			"hasIgnoredError",
			grpcfedcel.MemberOverloadFunc(fmt.Sprintf("%s_has_ignored_error", protoName), cel.ObjectType(protoName), []*cel.Type{}, cel.BoolType,
				func(ctx context.Context, self ref.Val, _ ...ref.Val) ref.Val {
					v := self.Value()
					value := localValueFromContext(ctx)
					if value == nil {
						return celtypes.Bool(false)
					}
					return celtypes.Bool(value.grpcError(v) != nil)
				},
			),
		)...,
	)
	opts = append(opts,
		grpcfedcel.BindMemberFunction(
			"ignoredError",
			grpcfedcel.MemberOverloadFunc(
				fmt.Sprintf("%s_ignored_error", protoName), cel.ObjectType(protoName), []*cel.Type{}, cel.ObjectType("grpc.federation.private.Error"),
				func(ctx context.Context, self ref.Val, _ ...ref.Val) ref.Val {
					v := self.Value()
					value := localValueFromContext(ctx)
					if value == nil {
						return nil
					}
					return adapter.NativeToValue(value.grpcError(v))
				},
			),
		)...,
	)
	return opts
}

type EnumAttributeMap[T ~int32] map[T]EnumValueAttributeMap

type EnumValueAttributeMap map[string]string

func EnumAttrOption[T ~int32](enumName string, enumAttrMap EnumAttributeMap[T]) cel.EnvOption {
	return cel.Function(
		fmt.Sprintf("%s.attr", enumName),
		cel.Overload(fmt.Sprintf("%s_attr", enumName), []*cel.Type{cel.IntType, cel.StringType}, cel.StringType,
			cel.BinaryBinding(func(enumValue, key ref.Val) ref.Val {
				enumValueAttrMap := enumAttrMap[T(enumValue.(celtypes.Int))]
				if enumValueAttrMap == nil {
					return celtypes.NewErr(`could not find enum value attribute map from %q`, enumName)
				}
				return celtypes.String(enumValueAttrMap[string(key.(celtypes.String))])
			}),
		),
	)
}

func NewDefaultEnvOptions(celHelper *CELTypeHelper) []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.StdLib(),
		ext.TwoVarComprehensions(),
		cel.Lib(grpcfedcel.NewLibrary(celHelper)),
		cel.CrossTypeNumericComparisons(true),
		cel.CustomTypeAdapter(celHelper.TypeAdapter()),
		cel.CustomTypeProvider(celHelper.TypeProvider()),
		cel.Variable("error", cel.ObjectType("grpc.federation.private.Error")),
		cel.Variable(ContextVariableName, cel.ObjectType(ContextTypeName)),
		cel.Container(celHelper.pkgName),
	}
	opts = append(opts, EnumAccessorOptions("google.rpc.Code", code.Code_value, code.Code_name)...)
	return opts
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
	sgCache    *singleflightCache
	mu         sync.RWMutex
	envOpts    []cel.EnvOption
	evalValues map[string]any
	grpcErrMap map[any]*grpcfedcel.Error
}

type singleflightCache struct {
	mu        sync.RWMutex
	resultMap map[string]*singleflightCacheResult
}

type singleflightCacheResult struct {
	result any
	err    error
}

func NewLocalValue(ctx context.Context, envOpts []cel.EnvOption, argName string, arg any) *LocalValue {
	var newEnvOpts []cel.EnvOption
	newEnvOpts = append(
		append(newEnvOpts, envOpts...),
		cel.Variable(MessageArgumentVariableName, cel.ObjectType(argName)),
	)
	return &LocalValue{
		sgCache: &singleflightCache{
			resultMap: make(map[string]*singleflightCacheResult),
		},
		envOpts:    newEnvOpts,
		evalValues: map[string]any{MessageArgumentVariableName: arg},
		grpcErrMap: make(map[any]*grpcfedcel.Error),
	}
}

func NewServiceVariableLocalValue(envOpts []cel.EnvOption) *LocalValue {
	newEnvOpts := append([]cel.EnvOption{}, envOpts...)
	return &LocalValue{
		sgCache: &singleflightCache{
			resultMap: make(map[string]*singleflightCacheResult),
		},
		envOpts:    newEnvOpts,
		evalValues: make(map[string]any),
	}
}

func (v *LocalValue) AddEnv(env any) {
	v.setEvalValue("grpc.federation.env", env)
}

func (v *LocalValue) AddServiceVariable(env any) {
	v.setEvalValue("grpc.federation.var", env)
}

func (v *LocalValue) SetGRPCError(retVal any, err *grpcfedcel.Error) {
	v.lock()
	v.grpcErrMap[retVal] = err
	v.unlock()
}

func (v *LocalValue) grpcError(retVal any) *grpcfedcel.Error {
	v.rlock()
	err := v.grpcErrMap[retVal]
	v.runlock()
	return err
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

func (v *LocalValue) getSingleflightCache(name string) (*singleflightCacheResult, bool) {
	v.sgCache.mu.RLock()
	defer v.sgCache.mu.RUnlock()

	cache, exists := v.sgCache.resultMap[name]
	return cache, exists
}

func (v *LocalValue) setSingleflightCache(name string, result *singleflightCacheResult) {
	v.sgCache.mu.Lock()
	v.sgCache.resultMap[name] = result
	v.sgCache.mu.Unlock()
}

func (v *LocalValue) do(name string, cb func() (any, error)) (any, error) {
	if cache, exists := v.getSingleflightCache(name); exists {
		return cache.result, cache.err
	}
	ret, err, _ := v.sg.Do(name, func() (any, error) {
		ret, err := cb()
		v.setSingleflightCache(name, &singleflightCacheResult{
			result: ret,
			err:    err,
		})
		return ret, err
	})
	return ret, err
}

func (v *LocalValue) WithLock(fn func()) {
	v.mu.Lock()
	defer v.mu.Unlock()
	fn()
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
	if grpcErr := getGRPCErrorValue(ctx); grpcErr != nil {
		ret["error"] = grpcErr
	}
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
	Enum         func(context.Context, U) (T, error)
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
	_, err := value.do(def.Name, func() (any, error) {
		return nil, evalDef(ctx, value, def)
	})
	return err
}

func IgnoreAndResponse[T any, U localValue](ctx context.Context, value U, def Def[T, U]) error {
	// doesn't use cache to create response variable by same name key.
	return evalDef(ctx, value, def)
}

func evalDef[T any, U localValue](ctx context.Context, value U, def Def[T, U]) error {
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
		ret, runErr := func() (any, error) {
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
			case def.Enum != nil:
				return def.Enum(ctx, value)
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
		}()
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
	_, err := value.do(def.Name, func() (any, error) {
		err := evalDefMap(ctx, value, def)
		return nil, err
	})
	return err
}

func evalDefMap[T any, U any, V localValue](ctx context.Context, value V, def DefMap[T, U, V]) error {
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
		ret, runErr := evalMap(
			ctx,
			value,
			def.IteratorName,
			def.IteratorType,
			def.IteratorSource,
			reflect.TypeOf(def.outType),
			def.Iterator,
		)
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
	if celCacheMap := getCELCacheMap(ctx); celCacheMap == nil {
		return nil, ErrCELCacheMap
	}
	if req.CacheIndex == 0 {
		return nil, ErrCELCacheIndex
	}

	req.Value.rlock()
	envOpts := req.Value.getEnvOpts()
	evalValues := req.Value.getEvalValues(ctx)
	req.Value.runlock()

	program, err := createCELProgram(ctx, req.Expr, req.CacheIndex, envOpts)
	if err != nil {
		return nil, err
	}
	out, _, err := program.ContextEval(ctx, evalValues)
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

func createCELProgram(ctx context.Context, expr string, cacheIndex int, envOpts []cel.EnvOption) (cel.Program, error) {
	if program := getCELProgramCache(ctx, cacheIndex); program != nil {
		return program, nil
	}

	env, err := NewCELEnv(envOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cel env: %w", err)
	}
	ast, err := createCELAst(expr, env)
	if err != nil {
		return nil, err
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	setCELProgramCache(ctx, cacheIndex, program)
	return program, nil
}

func createCELAst(expr string, env *cel.Env) (*cel.Ast, error) {
	replacedExpr := strings.Replace(expr, "$", MessageArgumentVariableName, -1)
	ast, iss := env.Compile(replacedExpr)
	if iss.Err() != nil {
		return nil, iss.Err()
	}

	checkedExpr, err := cel.AstToCheckedExpr(ast)
	if err != nil {
		return nil, err
	}
	if newComparingNullResolver().Resolve(checkedExpr) {
		ca, err := celast.ToAST(checkedExpr)
		if err != nil {
			return nil, err
		}
		expr, err := parser.Unparse(ca.Expr(), ca.SourceInfo())
		if err != nil {
			return nil, err
		}
		ast, iss = env.Compile(expr)
		if iss.Err() != nil {
			return nil, iss.Err()
		}
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

func ToGRPCError(ctx context.Context, err error) *grpcfedcel.Error {
	stat, ok := grpcstatus.FromError(err)
	if !ok {
		return nil
	}
	grpcErr := &grpcfedcel.Error{
		Code:    int32(stat.Code()), //nolint:gosec
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
	return grpcErr
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

// comparingNullResolver is a feature that allows to compare typed null and null value correctly.
// It parses the expression and wraps the message with grpc.federation.cast.null_value if the message is compared to null.
type comparingNullResolver struct {
	checkedExpr *exprpb.CheckedExpr
	lastID      int64
	resolved    bool
}

func newComparingNullResolver() *comparingNullResolver {
	return &comparingNullResolver{}
}

func (r *comparingNullResolver) init(checkedExpr *exprpb.CheckedExpr) {
	var lastID int64
	for k := range checkedExpr.GetReferenceMap() {
		if lastID < k {
			lastID = k
		}
	}
	for k := range checkedExpr.GetTypeMap() {
		if lastID < k {
			lastID = k
		}
	}
	r.checkedExpr = checkedExpr
	r.lastID = lastID
	r.resolved = false
}

func (r *comparingNullResolver) nextID() int64 {
	r.lastID++
	return r.lastID
}

func (r *comparingNullResolver) Resolve(checkedExpr *exprpb.CheckedExpr) bool {
	r.init(checkedExpr)
	newExprVisitor().Visit(checkedExpr.GetExpr(), func(e *exprpb.Expr) {
		switch e.GetExprKind().(type) {
		case *exprpb.Expr_CallExpr:
			r.resolveCallExpr(e)
		}
	})
	return r.resolved
}

func (r *comparingNullResolver) resolveCallExpr(e *exprpb.Expr) {
	target := r.lookupComparingNullCallExpr(e)
	if target == nil {
		return
	}
	newID := r.nextID()
	newExprKind := &exprpb.Expr_CallExpr{
		CallExpr: &exprpb.Expr_Call{
			Function: grpcfedcel.CastNullValueFunc,
			Args: []*exprpb.Expr{
				{
					Id:       target.GetId(),
					ExprKind: target.GetExprKind(),
				},
			},
		},
	}
	target.Id = newID
	target.ExprKind = newExprKind
	r.checkedExpr.GetReferenceMap()[newID] = &exprpb.Reference{
		OverloadId: []string{grpcfedcel.CastNullValueFunc},
	}
	r.checkedExpr.GetTypeMap()[newID] = &exprpb.Type{
		TypeKind: &exprpb.Type_Dyn{},
	}
	r.resolved = true
}

func (r *comparingNullResolver) lookupComparingNullCallExpr(e *exprpb.Expr) *exprpb.Expr {
	call := e.GetCallExpr()
	fnName := call.GetFunction()
	if fnName == operators.Equals || fnName == operators.NotEquals {
		lhs := call.GetArgs()[0]
		rhs := call.GetArgs()[1]
		var target *exprpb.Expr
		if _, ok := lhs.GetConstExpr().GetConstantKind().(*exprpb.Constant_NullValue); ok {
			target = rhs
		}
		if _, ok := rhs.GetConstExpr().GetConstantKind().(*exprpb.Constant_NullValue); ok {
			if target != nil {
				// maybe null == null
				return nil
			}
			target = lhs
		}
		if target == nil {
			return nil
		}
		if target.GetCallExpr() != nil && target.GetCallExpr().GetFunction() == grpcfedcel.CastNullValueFunc {
			return nil
		}
		return target
	}
	return nil
}

type exprVisitor struct {
	callback func(e *exprpb.Expr)
}

func newExprVisitor() *exprVisitor {
	return &exprVisitor{}
}

func (v *exprVisitor) Visit(e *exprpb.Expr, cb func(e *exprpb.Expr)) {
	v.callback = cb
	v.visit(e)
}

func (v *exprVisitor) visit(e *exprpb.Expr) {
	if e == nil {
		return
	}
	v.callback(e)

	switch e.GetExprKind().(type) {
	case *exprpb.Expr_SelectExpr:
		v.visitSelect(e)
	case *exprpb.Expr_CallExpr:
		v.visitCall(e)
	case *exprpb.Expr_ListExpr:
		v.visitList(e)
	case *exprpb.Expr_StructExpr:
		v.visitStruct(e)
	case *exprpb.Expr_ComprehensionExpr:
		v.visitComprehension(e)
	}
}

func (v *exprVisitor) visitSelect(e *exprpb.Expr) {
	sel := e.GetSelectExpr()
	v.visit(sel.GetOperand())
}

func (v *exprVisitor) visitCall(e *exprpb.Expr) {
	call := e.GetCallExpr()
	for _, arg := range call.GetArgs() {
		v.visit(arg)
	}
	v.visit(call.GetTarget())
}

func (v *exprVisitor) visitList(e *exprpb.Expr) {
	l := e.GetListExpr()
	for _, elem := range l.GetElements() {
		v.visit(elem)
	}
}

func (v *exprVisitor) visitStruct(e *exprpb.Expr) {
	msg := e.GetStructExpr()
	for _, ent := range msg.GetEntries() {
		v.visit(ent.GetValue())
	}
}

func (v *exprVisitor) visitComprehension(e *exprpb.Expr) {
	comp := e.GetComprehensionExpr()
	v.visit(comp.GetIterRange())
	v.visit(comp.GetAccuInit())
	v.visit(comp.GetLoopCondition())
	v.visit(comp.GetLoopStep())
	v.visit(comp.GetResult())
}
