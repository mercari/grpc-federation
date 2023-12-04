package federation

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
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

func EvalCEL(env *cel.Env, expr string, vars []cel.EnvOption, args map[string]any, outType reflect.Type) (any, error) {
	env, err := env.Extend(vars...)
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
	out, _, err := program.Eval(args)
	if err != nil {
		return nil, err
	}
	opt, ok := out.(*types.Optional)
	if ok {
		if opt == types.OptionalNone {
			return reflect.Zero(outType).Interface(), nil
		}
		out = opt.GetValue()
	}
	if outType != nil {
		return out.ConvertToNative(outType)
	}
	return out.Value(), nil
}
