package resolver_test

import (
	"testing"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

func TestCELFunction(t *testing.T) {
	enumType := &resolver.Type{
		Kind: types.Enum,
		Enum: &resolver.Enum{
			File: &resolver.File{},
			Name: "EnumType",
		},
	}
	celEnumType := celtypes.NewOpaqueType(enumType.Enum.FQDN(), cel.IntType)
	fn := resolver.CELFunction{
		Name:   "test",
		ID:     "test",
		Args:   []*resolver.Type{enumType, enumType},
		Return: enumType,
	}
	if diff := cmp.Diff(fn.Signatures(), []*resolver.CELPluginFunctionSignature{
		{
			ID:     "test_int32_int32_int32",
			Args:   []*cel.Type{cel.IntType, cel.IntType},
			Return: cel.IntType,
		},
		{
			ID:     "test_int32_int32__EnumType",
			Args:   []*cel.Type{cel.IntType, cel.IntType},
			Return: celEnumType,
		},
		{
			ID:     "test_int32__EnumType_int32",
			Args:   []*cel.Type{cel.IntType, celEnumType},
			Return: cel.IntType,
		},
		{
			ID:     "test_int32__EnumType__EnumType",
			Args:   []*cel.Type{cel.IntType, celEnumType},
			Return: celEnumType,
		},
		{
			ID:     "test__EnumType_int32_int32",
			Args:   []*cel.Type{celEnumType, cel.IntType},
			Return: cel.IntType,
		},
		{
			ID:     "test__EnumType_int32__EnumType",
			Args:   []*cel.Type{celEnumType, cel.IntType},
			Return: celEnumType,
		},
		{
			ID:     "test__EnumType__EnumType_int32",
			Args:   []*cel.Type{celEnumType, celEnumType},
			Return: cel.IntType,
		},
		{
			ID:     "test__EnumType__EnumType__EnumType",
			Args:   []*cel.Type{celEnumType, celEnumType},
			Return: celEnumType,
		},
	}, cmpopts.IgnoreUnexported(cel.Type{})); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
