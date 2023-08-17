package resolver

import (
	"github.com/mercari/grpc-federation/types"
)

var (
	AnyType       *Type
	TimestampType *Type
	DurationType  *Type
	EmptyType     *Type
)

func init() {
	files, err := New(nil).ResolveWellknownFiles()
	if err != nil {
		panic(err)
	}
	for _, file := range files.FindByPackageName("google.protobuf") {
		if msg := file.Message("Any"); msg != nil {
			AnyType = &Type{Type: types.Message, Ref: msg}
		}
		if msg := file.Message("Timestamp"); msg != nil {
			TimestampType = &Type{Type: types.Message, Ref: msg}
		}
		if msg := file.Message("Duration"); msg != nil {
			DurationType = &Type{Type: types.Message, Ref: msg}
		}
		if msg := file.Message("Empty"); msg != nil {
			EmptyType = &Type{Type: types.Message, Ref: msg}
		}
	}
}
