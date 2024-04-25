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

func ToFloat64(v *celplugin.CELPluginValue) (float64, error) {
	return v.GetDouble(), nil
}

func ToInt32(v *celplugin.CELPluginValue) (int32, error) {
	return int32(v.GetInt64()), nil
}

func ToInt64(v *celplugin.CELPluginValue) (int64, error) {
	return v.GetInt64(), nil
}

func ToUint32(v *celplugin.CELPluginValue) (uint32, error) {
	return uint32(v.GetUint64()), nil
}

func ToUint64(v *celplugin.CELPluginValue) (uint64, error) {
	return v.GetUint64(), nil
}

func ToString(v *celplugin.CELPluginValue) (string, error) {
	return v.GetString_(), nil
}

func ToBytes(v *celplugin.CELPluginValue) ([]byte, error) {
	return v.GetBytes(), nil
}

func ToBool(v *celplugin.CELPluginValue) (bool, error) {
	return v.GetBool(), nil
}

func ToEnum[T ~int32](v *celplugin.CELPluginValue) (T, error) {
	return T(v.GetInt64()), nil
}

func ToMessage[T proto.Message](v *celplugin.CELPluginValue) (T, error) {
	msg, err := v.GetMessage().UnmarshalNew()
	if err != nil {
		var ret T
		return ret, err
	}
	return msg.(T), nil
}

func ToErrorCELPluginResponse(v error) *celplugin.CELPluginResponse {
	return &celplugin.CELPluginResponse{Error: v.Error()}
}

func ToMessageCELPluginResponse[T proto.Message](v T) (*celplugin.CELPluginResponse, error) {
	anyValue, err := anypb.New(v)
	if err != nil {
		return nil, err
	}
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Message{
				Message: anyValue,
			},
		},
	}, nil
}

func ToStringCELPluginResponse(v string) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_String_{
				String_: v,
			},
		},
	}, nil
}

func ToBytesCELPluginResponse(v []byte) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Bytes{
				Bytes: v,
			},
		},
	}, nil
}

func ToInt32CELPluginResponse(v int32) (*celplugin.CELPluginResponse, error) {
	return ToInt64CELPluginResponse(int64(v))
}

func ToInt64CELPluginResponse(v int64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Int64{
				Int64: v,
			},
		},
	}, nil
}

func ToUint32CELPluginResponse(v uint32) (*celplugin.CELPluginResponse, error) {
	return ToUint64CELPluginResponse(uint64(v))
}

func ToUint64CELPluginResponse(v uint64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Uint64{
				Uint64: v,
			},
		},
	}, nil
}

func ToFloat32CELPluginResponse(v float32) (*celplugin.CELPluginResponse, error) {
	return ToFloat64CELPluginResponse(float64(v))
}

func ToFloat64CELPluginResponse(v float64) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Double{
				Double: v,
			},
		},
	}, nil
}

func ToBoolCELPluginResponse(v bool) (*celplugin.CELPluginResponse, error) {
	return &celplugin.CELPluginResponse{
		Value: &celplugin.CELPluginValue{
			Value: &celplugin.CELPluginValue_Bool{
				Bool: v,
			},
		},
	}, nil
}
