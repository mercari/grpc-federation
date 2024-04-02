//go:build !tinygo.wasm

package federation

import (
	"context"
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
	celRegistry    *celtypes.Registry
	structFieldMap map[string]map[string]*celtypes.FieldType
	mapMu          sync.RWMutex
}

func (h *CELTypeHelper) TypeProvider() celtypes.Provider {
	return h
}

func (h *CELTypeHelper) TypeAdapter() celtypes.Adapter {
	return h.celRegistry
}

func (h *CELTypeHelper) EnumValue(enumName string) ref.Val {
	return h.celRegistry.EnumValue(enumName)
}

func (h *CELTypeHelper) FindIdent(identName string) (ref.Val, bool) {
	return h.celRegistry.FindIdent(identName)
}

func (h *CELTypeHelper) FindStructType(structType string) (*celtypes.Type, bool) {
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

func (h *CELTypeHelper) FindStructFieldNames(structType string) ([]string, bool) {
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

func (h *CELTypeHelper) FindStructFieldType(structType, fieldName string) (*celtypes.FieldType, bool) {
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

func (h *CELTypeHelper) NewValue(structType string, fields map[string]ref.Val) ref.Val {
	return h.celRegistry.NewValue(structType, fields)
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

func NewCELTypeHelper(structFieldMap map[string]map[string]*celtypes.FieldType) *CELTypeHelper {
	celRegistry := celtypes.NewEmptyRegistry()
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if err := celRegistry.RegisterDescriptor(f); err != nil {
			return false
		}
		return true
	})
	return &CELTypeHelper{
		celRegistry:    celRegistry,
		structFieldMap: structFieldMap,
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
	}
}

func NewDefaultEnvOptions(celHelper *CELTypeHelper) []cel.EnvOption {
	return []cel.EnvOption{
		cel.StdLib(),
		cel.Lib(grpcfedcel.NewLibrary()),
		cel.CrossTypeNumericComparisons(true),
		cel.CustomTypeAdapter(celHelper.TypeAdapter()),
		cel.CustomTypeProvider(celHelper.TypeProvider()),
		cel.Variable("error", cel.ObjectType("grpc.federation.private.Error")),
	}
}

type LocalValue struct {
	sg         singleflight.Group
	mu         sync.RWMutex
	env        *cel.Env
	envOpts    []cel.EnvOption
	evalValues map[string]any
}

func NewLocalValue(env *cel.Env, argName string, arg any) *LocalValue {
	return &LocalValue{
		env: env,
		envOpts: []cel.EnvOption{
			cel.Variable(
				MessageArgumentVariableName,
				cel.ObjectType(argName),
			),
		},
		evalValues: map[string]any{
			MessageArgumentVariableName: arg,
		},
	}
}

type localValue interface {
	do(string, func() (any, error)) (any, error)
	rlock()
	runlock()
	lock()
	unlock()
	getEnv() *cel.Env
	getEnvOpts() []cel.EnvOption
	getEvalValues() map[string]any
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

func (v *LocalValue) getEnv() *cel.Env {
	return v.env
}

func (v *LocalValue) getEnvOpts() []cel.EnvOption {
	return v.envOpts
}

func (v *LocalValue) getEvalValues() map[string]any {
	return v.evalValues
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

func (v *MapIteratorValue) getEnv() *cel.Env {
	return v.localValue.getEnv()
}

func (v *MapIteratorValue) getEnvOpts() []cel.EnvOption {
	return v.envOpts
}

func (v *MapIteratorValue) getEvalValues() map[string]any {
	return v.evalValues
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
	If         string
	Name       string
	Type       *cel.Type
	Setter     func(U, T)
	By         string
	Message    func(context.Context, U) (any, error)
	Validation func(context.Context, U) error
}

type DefMap[T any, U any, V localValue] struct {
	If             string
	Name           string
	Type           *cel.Type
	Setter         func(V, T)
	IteratorName   string
	IteratorType   *cel.Type
	IteratorSource func(V) []U
	Iterator       func(context.Context, *MapIteratorValue) (any, error)
	outType        T
}

func EvalDef[T any, U localValue](ctx context.Context, value U, def Def[T, U]) error {
	var (
		v    T
		err  error
		cond = true
		name = def.Name
	)
	if def.If != "" {
		c, err := EvalCEL(ctx, value, def.If, reflect.TypeOf(false))
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
				return EvalCEL(ctx, value, def.By, reflect.TypeOf(v))
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
		err = runErr
	}

	value.lock()
	def.Setter(value, v)
	value.setEnvOptValue(name, def.Type)
	value.setEvalValue(name, v)
	value.unlock()

	if err != nil {
		return err
	}
	return nil
}

func EvalDefMap[T any, U any, V localValue](ctx context.Context, value V, def DefMap[T, U, V]) error {
	var (
		v    T
		err  error
		cond = true
		name = def.Name
	)
	if def.If != "" {
		c, err := EvalCEL(ctx, value, def.If, reflect.TypeOf(false))
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
		err = runErr
	}

	value.lock()
	def.Setter(value, v)
	value.setEnvOptValue(name, def.Type)
	value.setEvalValue(name, v)
	value.unlock()

	if err != nil {
		return err
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
	for k, v := range value.getEvalValues() {
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

func If[T localValue](ctx context.Context, value T, expr string, cb func(T) error) error {
	cond, err := EvalCEL(ctx, value, expr, reflect.TypeOf(false))
	if err != nil {
		return err
	}
	if cond.(bool) {
		return cb(value)
	}
	return nil
}

func EvalCEL(ctx context.Context, value localValue, expr string, outType reflect.Type) (any, error) {
	value.rlock()
	defer value.runlock()

	var envOpts []cel.EnvOption
	envOpts = append(envOpts, value.getEnvOpts()...)
	envOpts = append(envOpts, cel.Lib(grpcfedcel.NewContextualLibrary(ctx)))
	for _, plugin := range CELPlugins(ctx) {
		plugin.SetContext(ctx)
		envOpts = append(envOpts, cel.Lib(plugin))
	}

	env, err := value.getEnv().Extend(envOpts...)
	if err != nil {
		return nil, err
	}
	expr = strings.Replace(expr, "$", MessageArgumentVariableName, -1)
	ast, iss := env.Compile(expr)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, err
	}
	out, _, err := program.ContextEval(ctx, value.getEvalValues())
	if err != nil {
		return nil, err
	}
	opt, ok := out.(*celtypes.Optional)
	if ok {
		if opt == celtypes.OptionalNone {
			return reflect.Zero(outType).Interface(), nil
		}
		out = opt.GetValue()
	}
	if outType != nil {
		return out.ConvertToNative(outType)
	}
	return out.Value(), nil
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

func SetCELValue[T any](ctx context.Context, value localValue, expr string, setter func(T)) error {
	var typ T
	out, err := EvalCEL(ctx, value, expr, reflect.TypeOf(typ))
	if err != nil {
		return err
	}

	value.lock()
	setter(out.(T))
	value.unlock()
	return nil
}
