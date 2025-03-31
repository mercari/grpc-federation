package federation

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	celplugin "github.com/mercari/grpc-federation/grpc/federation/cel/plugin"
)

type (
	CELPluginRequest  = celplugin.CELPluginRequest
	CELPluginResponse = celplugin.CELPluginResponse
)

func DecodeCELPluginRequest(v []byte) (*celplugin.CELPluginRequest, error) {
	var req celplugin.CELPluginRequest
	if err := protojson.Unmarshal(v, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func EncodeCELPluginVersion(v CELPluginVersionSchema) ([]byte, error) {
	return json.Marshal(v)
}

func EncodeCELPluginResponse(v *celplugin.CELPluginResponse) ([]byte, error) {
	encoded, err := protojson.Marshal(v)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func ToFloat32(v *celplugin.CELPluginValue) (float32, error) {
	return float32(v.GetDouble()), nil
}

func ToFloat32List(v *celplugin.CELPluginValue) ([]float32, error) {
	ret := make([]float32, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, float32(vv.GetDouble()))
	}
	return ret, nil
}

func ToFloat64(v *celplugin.CELPluginValue) (float64, error) {
	return v.GetDouble(), nil
}

func ToFloat64List(v *celplugin.CELPluginValue) ([]float64, error) {
	ret := make([]float64, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetDouble())
	}
	return ret, nil
}

func ToInt32(v *celplugin.CELPluginValue) (int32, error) {
	return int32(v.GetInt64()), nil //nolint:gosec
}

func ToInt32List(v *celplugin.CELPluginValue) ([]int32, error) {
	ret := make([]int32, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, int32(vv.GetInt64())) //nolint:gosec
	}
	return ret, nil
}

func ToInt64(v *celplugin.CELPluginValue) (int64, error) {
	return v.GetInt64(), nil
}

func ToInt64List(v *celplugin.CELPluginValue) ([]int64, error) {
	ret := make([]int64, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetInt64())
	}
	return ret, nil
}

func ToUint32(v *celplugin.CELPluginValue) (uint32, error) {
	return uint32(v.GetUint64()), nil //nolint:gosec
}

func ToUint32List(v *celplugin.CELPluginValue) ([]uint32, error) {
	ret := make([]uint32, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, uint32(vv.GetUint64())) //nolint:gosec
	}
	return ret, nil
}

func ToUint64(v *celplugin.CELPluginValue) (uint64, error) {
	return v.GetUint64(), nil
}

func ToUint64List(v *celplugin.CELPluginValue) ([]uint64, error) {
	ret := make([]uint64, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetUint64())
	}
	return ret, nil
}

func ToString(v *celplugin.CELPluginValue) (string, error) {
	return v.GetString_(), nil
}

func ToStringList(v *celplugin.CELPluginValue) ([]string, error) {
	ret := make([]string, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetString_())
	}
	return ret, nil
}

func ToBytes(v *celplugin.CELPluginValue) ([]byte, error) {
	return v.GetBytes(), nil
}

func ToBytesList(v *celplugin.CELPluginValue) ([][]byte, error) {
	ret := make([][]byte, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetBytes())
	}
	return ret, nil
}

func ToBool(v *celplugin.CELPluginValue) (bool, error) {
	return v.GetBool(), nil
}

func ToBoolList(v *celplugin.CELPluginValue) ([]bool, error) {
	ret := make([]bool, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, vv.GetBool())
	}
	return ret, nil
}

func ToEnum[T ~int32](v *celplugin.CELPluginValue) (T, error) {
	return T(v.GetInt64()), nil
}

func ToEnumList[T ~int32](v *celplugin.CELPluginValue) ([]T, error) {
	ret := make([]T, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		ret = append(ret, T(vv.GetInt64()))
	}
	return ret, nil
}

func ToMessage[T proto.Message](v *celplugin.CELPluginValue) (T, error) {
	msg, err := v.GetMessage().UnmarshalNew()
	if err != nil {
		var ret T
		return ret, err
	}
	return msg.(T), nil
}

func ToMessageList[T proto.Message](v *celplugin.CELPluginValue) ([]T, error) {
	ret := make([]T, 0, len(v.GetList().GetValues()))
	for _, vv := range v.GetList().GetValues() {
		msg, err := vv.GetMessage().UnmarshalNew()
		if err != nil {
			return nil, err
		}
		ret = append(ret, msg.(T))
	}
	return ret, nil
}

func ToErrorCELPluginResponse(v error) *celplugin.CELPluginResponse {
	return &celplugin.CELPluginResponse{Error: v.Error()}
}

func ToMessageCELPluginResponse[T proto.Message](v T) (*celplugin.CELPluginResponse, error) {
	value, err := toMessageCELPluginValue(v)
	if err != nil {
		return nil, err
	}
	return &celplugin.CELPluginResponse{Value: value}, nil
}

func ToMessageListCELPluginResponse[T proto.Message](v []T) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		value, err := toMessageCELPluginValue(vv)
		if err != nil {
			return nil, err
		}
		ret = append(ret, value)
	}
	return toListResponse(ret), nil
}

func toMessageCELPluginValue[T proto.Message](v T) (*celplugin.CELPluginValue, error) {
	anyValue, err := anypb.New(v)
	if err != nil {
		return nil, err
	}
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Message{Message: anyValue}}, nil
}

func ToStringCELPluginResponse(v string) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toStringCELPluginValue(v)}, nil
}

func ToStringListCELPluginResponse(v []string) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toStringCELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toStringCELPluginValue(v string) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_String_{String_: v}}
}

func ToBytesCELPluginResponse(v []byte) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toBytesCELPluginValue(v)}, nil
}

func ToBytesListCELPluginResponse(v [][]byte) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toBytesCELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toBytesCELPluginValue(v []byte) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Bytes{Bytes: v}}
}

func ToInt32CELPluginResponse(v int32) (*celplugin.CELPluginResponse, error) {
	return ToInt64CELPluginResponse(int64(v))
}

func ToInt32ListCELPluginResponse(v []int32) (*celplugin.CELPluginResponse, error) {
	i64 := make([]int64, 0, len(v))
	for _, vv := range v {
		i64 = append(i64, int64(vv))
	}
	return ToInt64ListCELPluginResponse(i64)
}

func ToInt64CELPluginResponse(v int64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toInt64CELPluginValue(v)}, nil
}

func ToInt64ListCELPluginResponse(v []int64) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toInt64CELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toInt64CELPluginValue(v int64) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Int64{Int64: v}}
}

func ToUint32CELPluginResponse(v uint32) (*celplugin.CELPluginResponse, error) {
	return ToUint64CELPluginResponse(uint64(v))
}

func ToUint32ListCELPluginResponse(v []uint32) (*celplugin.CELPluginResponse, error) {
	u64 := make([]uint64, 0, len(v))
	for _, vv := range v {
		u64 = append(u64, uint64(vv))
	}
	return ToUint64ListCELPluginResponse(u64)
}

func ToUint64CELPluginResponse(v uint64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toUint64CELPluginValue(v)}, nil
}

func ToUint64ListCELPluginResponse(v []uint64) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toUint64CELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toUint64CELPluginValue(v uint64) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Uint64{Uint64: v}}
}

func ToFloat32CELPluginResponse(v float32) (*celplugin.CELPluginResponse, error) {
	return ToFloat64CELPluginResponse(float64(v))
}

func ToFloat32ListCELPluginResponse(v []float32) (*celplugin.CELPluginResponse, error) {
	f64 := make([]float64, 0, len(v))
	for _, vv := range v {
		f64 = append(f64, float64(vv))
	}
	return ToFloat64ListCELPluginResponse(f64)
}

func ToFloat64CELPluginResponse(v float64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toFloat64CELPluginValue(v)}, nil
}

func ToFloat64ListCELPluginResponse(v []float64) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toFloat64CELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toFloat64CELPluginValue(v float64) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Double{Double: v}}
}

func ToBoolCELPluginResponse(v bool) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{Value: toBoolCELPluginValue(v)}, nil
}

func ToBoolListCELPluginResponse(v []bool) (*celplugin.CELPluginResponse, error) {
	ret := make([]*celplugin.CELPluginValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, toBoolCELPluginValue(vv))
	}
	return toListResponse(ret), nil
}

func toBoolCELPluginValue(v bool) *celplugin.CELPluginValue {
	return &celplugin.CELPluginValue{Value: &celplugin.CELPluginValue_Bool{Bool: v}}
}

func toListResponse(v []*celplugin.CELPluginValue) *celplugin.CELPluginResponse {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_List{List: &celplugin.CELPluginListValue{Values: v}},
		},
	}
}
