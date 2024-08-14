package cel

import (
	"context"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

const (
	ContextVariableName = "__CTX__"
	ContextTypeName     = "grpc.federation.private.Context"
)

type ContextValue struct {
	context.Context
}

func NewContextValue(c context.Context) *ContextValue {
	return &ContextValue{Context: c}
}

func (c *ContextValue) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return c.Context, nil
}

func (c *ContextValue) ConvertToType(typeValue ref.Type) ref.Val {
	return nil
}

func (c *ContextValue) Equal(other ref.Val) ref.Val {
	return other
}

func (c *ContextValue) Type() ref.Type {
	return cel.ObjectType(ContextTypeName)
}

func (c *ContextValue) Value() any {
	return c.Context
}
