package cel

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const JSONPackageName = "json"

var _ cel.SingletonLibrary = new(JSONLibrary)

type JSONLibrary struct {
	typeAdapter types.Adapter
}

func NewJSONLibrary(typeAdapter types.Adapter) *JSONLibrary {
	return &JSONLibrary{typeAdapter: typeAdapter}
}

func (lib *JSONLibrary) LibraryName() string {
	return packageName(JSONPackageName)
}

func createJSONName(name string) string {
	return createName(JSONPackageName, name)
}

func createJSONID(name string) string {
	return createID(JSONPackageName, name)
}

func (lib *JSONLibrary) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createJSONName("from"),
			OverloadFunc(createJSONID("from_string_dyn_dyn"),
				[]*cel.Type{cel.StringType, cel.DynType},
				cel.DynType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					jsonStr := string(args[0].(types.String))
					return lib.fromJSON(jsonStr, args[1])
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *JSONLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (lib *JSONLibrary) fromJSON(jsonStr string, template ref.Val) ref.Val {
	if jsonStr == "null" {
		return lib.zeroValue(template)
	}

	if msg, ok := template.Value().(proto.Message); ok {
		return lib.unmarshalMessage(jsonStr, msg)
	}

	if lister, ok := template.(traits.Lister); ok {
		return lib.unmarshalList(jsonStr, lister)
	}

	if mapper, ok := template.(traits.Mapper); ok {
		return lib.unmarshalMap(jsonStr, mapper)
	}

	switch template.Type() {
	case types.StringType:
		var s string
		if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.String(s)
	case types.IntType:
		var i int64
		if err := json.Unmarshal([]byte(jsonStr), &i); err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.Int(i)
	case types.UintType:
		var u uint64
		if err := json.Unmarshal([]byte(jsonStr), &u); err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.Uint(u)
	case types.DoubleType:
		var d float64
		if err := json.Unmarshal([]byte(jsonStr), &d); err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.Double(d)
	case types.BoolType:
		var b bool
		if err := json.Unmarshal([]byte(jsonStr), &b); err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.Bool(b)
	case types.BytesType:
		var s string
		if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
			return types.NewErrFromString(err.Error())
		}
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return types.NewErrFromString(err.Error())
		}
		return types.Bytes(decoded)
	default:
		return types.NewErrFromString("unsupported template type")
	}
}

func (lib *JSONLibrary) zeroValue(template ref.Val) ref.Val {
	if msg, ok := template.Value().(proto.Message); ok {
		newMsg := msg.ProtoReflect().New().Interface()
		return lib.typeAdapter.NativeToValue(newMsg)
	}
	if _, ok := template.(traits.Lister); ok {
		return types.NewRefValList(lib.typeAdapter, []ref.Val{})
	}
	if _, ok := template.(traits.Mapper); ok {
		return types.NewRefValMap(lib.typeAdapter, map[ref.Val]ref.Val{})
	}
	switch template.Type() {
	case types.StringType:
		return types.String("")
	case types.IntType:
		return types.IntZero
	case types.UintType:
		return types.Uint(0)
	case types.DoubleType:
		return types.Double(0)
	case types.BoolType:
		return types.False
	case types.BytesType:
		return types.Bytes([]byte{})
	default:
		return types.NewErrFromString("unsupported type for null")
	}
}

func (lib *JSONLibrary) unmarshalMessage(jsonStr string, template proto.Message) ref.Val {
	newMsg := template.ProtoReflect().New().Interface()

	if err := protojson.Unmarshal([]byte(jsonStr), newMsg); err != nil {
		return types.NewErrFromString(err.Error())
	}
	return lib.typeAdapter.NativeToValue(newMsg)
}

func (lib *JSONLibrary) unmarshalList(jsonStr string, lister traits.Lister) ref.Val {
	var rawList []json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &rawList); err != nil {
		return types.NewErrFromString(err.Error())
	}

	var elemTemplate ref.Val
	if lister.Size().(types.Int) > 0 {
		elemTemplate = lister.Get(types.IntZero)
	}

	results := make([]ref.Val, len(rawList))
	for i, raw := range rawList {
		var val ref.Val
		if elemTemplate != nil {
			val = lib.fromJSON(string(raw), elemTemplate)
		} else {
			val = lib.unmarshalDynamic(string(raw))
		}
		if types.IsError(val) {
			return val
		}
		results[i] = val
	}
	return types.NewRefValList(lib.typeAdapter, results)
}

func (lib *JSONLibrary) unmarshalMap(jsonStr string, mapper traits.Mapper) ref.Val {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &rawMap); err != nil {
		return types.NewErrFromString(err.Error())
	}

	var valueTemplate ref.Val
	it := mapper.Iterator()
	if it.HasNext() == types.True {
		key := it.Next()
		valueTemplate = mapper.Get(key)
	}

	results := make(map[ref.Val]ref.Val, len(rawMap))
	for k, raw := range rawMap {
		var val ref.Val
		if valueTemplate != nil {
			val = lib.fromJSON(string(raw), valueTemplate)
		} else {
			val = lib.unmarshalDynamic(string(raw))
		}
		if types.IsError(val) {
			return val
		}
		results[types.String(k)] = val
	}
	return types.NewRefValMap(lib.typeAdapter, results)
}

func (lib *JSONLibrary) unmarshalDynamic(jsonStr string) ref.Val {
	if jsonStr == "null" {
		return types.NullValue
	}

	var v any
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return types.NewErrFromString(err.Error())
	}

	return lib.typeAdapter.NativeToValue(v)
}
