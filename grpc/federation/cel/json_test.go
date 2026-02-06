package cel_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"testing/quick"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/grpc/federation/cel/testdata/testpb"
)

func TestJSONFromFunction(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		err  bool
		cmp  func(ref.Val) error
	}{
		// Primitive types
		{
			name: "string",
			expr: `grpc.federation.json.from('"hello"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.String("hello")); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "string with unicode",
			expr: `grpc.federation.json.from('"日本語"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.String("日本語")); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "int",
			expr: `grpc.federation.json.from('42', 0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.Int(42)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "negative int",
			expr: `grpc.federation.json.from('-123', 0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.Int(-123)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "uint",
			expr: `grpc.federation.json.from('42', uint(0))`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.Uint(42)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "double",
			expr: `grpc.federation.json.from('3.14', 0.0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.Double(3.14)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "bool true",
			expr: `grpc.federation.json.from('true', false)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.True); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "bool false",
			expr: `grpc.federation.json.from('false', false)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.False); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "bytes (base64)",
			expr: `grpc.federation.json.from('"SGVsbG8gV29ybGQ="', b"")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := types.Bytes("Hello World")
				if diff := cmp.Diff(gotV, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// List types - using template instances with element
		{
			name: "list of int",
			expr: `grpc.federation.json.from('[1, 2, 3]', [0])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 3 {
					return fmt.Errorf("expected size 3, got %v", gotV.Size())
				}
				// With template [0], elements should be parsed as int
				for i, expected := range []types.Int{1, 2, 3} {
					if gotV.Get(types.Int(i)).(types.Int) != expected {
						return fmt.Errorf("expected %v at index %d, got %v", expected, i, gotV.Get(types.Int(i)))
					}
				}
				return nil
			},
		},
		{
			name: "list of string",
			expr: `grpc.federation.json.from('["a", "b", "c"]', [""])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 3 {
					return fmt.Errorf("expected size 3, got %v", gotV.Size())
				}
				for i, expected := range []string{"a", "b", "c"} {
					got := gotV.Get(types.Int(i)).(types.String)
					if string(got) != expected {
						return fmt.Errorf("expected %v at index %d, got %v", expected, i, got)
					}
				}
				return nil
			},
		},
		{
			name: "empty list",
			expr: `grpc.federation.json.from('[]', [0])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 0 {
					return fmt.Errorf("expected size 0, got %v", gotV.Size())
				}
				return nil
			},
		},

		// Map types - using template instances with key-value pair
		{
			name: "map of string to int",
			expr: `grpc.federation.json.from('{"a": 1, "b": 2, "c": 3}', {"": 0})`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 3 {
					return fmt.Errorf("expected size 3, got %v", gotV.Size())
				}
				// Check values
				expected := map[string]int64{"a": 1, "b": 2, "c": 3}
				for k, v := range expected {
					val := gotV.Get(types.String(k))
					if val.(types.Int) != types.Int(v) {
						return fmt.Errorf("expected %s=%d, got %v", k, v, val)
					}
				}
				return nil
			},
		},
		{
			name: "map of string to string",
			expr: `grpc.federation.json.from('{"key1": "value1", "key2": "value2"}', {"": ""})`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 2 {
					return fmt.Errorf("expected size 2, got %v", gotV.Size())
				}
				if gotV.Get(types.String("key1")).(types.String) != "value1" {
					return fmt.Errorf("expected key1=value1, got %v", gotV.Get(types.String("key1")))
				}
				return nil
			},
		},
		{
			name: "empty map",
			expr: `grpc.federation.json.from('{}', {"": 0})`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 0 {
					return fmt.Errorf("expected size 0, got %v", gotV.Size())
				}
				return nil
			},
		},
		{
			name: "map with null value",
			expr: `grpc.federation.json.from('{"a": 1, "b": null, "c": 3}', {"": 0})`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				// null should become zero value (0 for int)
				if gotV.Get(types.String("b")).(types.Int) != 0 {
					return fmt.Errorf("expected b=0, got %v", gotV.Get(types.String("b")))
				}
				return nil
			},
		},

		// Proto message types - using template instances
		// Note: CEL returns dynamicpb.Message for struct literals, so we verify fields via proto reflection
		{
			name: "proto message",
			expr: `grpc.federation.json.from('{"id": "test-id", "num": 42}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				// Verify fields via reflection (works with both concrete and dynamic types)
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				if refl.Get(idField).String() != "test-id" {
					return fmt.Errorf("expected id 'test-id', got %v", refl.Get(idField).String())
				}
				if refl.Get(numField).Int() != 42 {
					return fmt.Errorf("expected num 42, got %v", refl.Get(numField).Int())
				}
				return nil
			},
		},
		{
			name: "proto message with nested",
			expr: `grpc.federation.json.from('{"id": "outer", "num": 1, "inner": {"id": "inner-id"}}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				innerField := refl.Descriptor().Fields().ByName("inner")
				if refl.Get(idField).String() != "outer" {
					return fmt.Errorf("expected id 'outer', got %v", refl.Get(idField).String())
				}
				if refl.Get(numField).Int() != 1 {
					return fmt.Errorf("expected num 1, got %v", refl.Get(numField).Int())
				}
				innerMsg := refl.Get(innerField).Message()
				innerIdField := innerMsg.Descriptor().Fields().ByName("id")
				if innerMsg.Get(innerIdField).String() != "inner-id" {
					return fmt.Errorf("expected inner.id 'inner-id', got %v", innerMsg.Get(innerIdField).String())
				}
				return nil
			},
		},
		{
			name: "empty proto message",
			expr: `grpc.federation.json.from('{}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				// Default values for empty message
				if refl.Get(idField).String() != "" {
					return fmt.Errorf("expected empty id, got %v", refl.Get(idField).String())
				}
				if refl.Get(numField).Int() != 0 {
					return fmt.Errorf("expected num 0, got %v", refl.Get(numField).Int())
				}
				return nil
			},
		},
		// List of proto message with template element
		{
			name: "list of proto message",
			expr: `grpc.federation.json.from('[{"id": "a"}, {"id": "b"}]', [grpc.federation.cel.test.Message{}])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 2 {
					return fmt.Errorf("expected size 2, got %v", gotV.Size())
				}
				expectedIds := []string{"a", "b"}
				for i, expectedId := range expectedIds {
					elem := gotV.Get(types.Int(i))
					msg, ok := elem.Value().(proto.Message)
					if !ok {
						return fmt.Errorf("at index %d: expected proto.Message, got %T", i, elem.Value())
					}
					refl := msg.ProtoReflect()
					idField := refl.Descriptor().Fields().ByName("id")
					if refl.Get(idField).String() != expectedId {
						return fmt.Errorf("at index %d: expected id %q, got %q", i, expectedId, refl.Get(idField).String())
					}
				}
				return nil
			},
		},

		// Null handling (returns zero values)
		{
			name: "null to string",
			expr: `grpc.federation.json.from('null', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.String("")); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "null to int",
			expr: `grpc.federation.json.from('null', 0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.IntZero); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "null to bool",
			expr: `grpc.federation.json.from('null', false)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if diff := cmp.Diff(gotV, types.False); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "null to list",
			expr: `grpc.federation.json.from('null', [0])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 0 {
					return fmt.Errorf("expected size 0, got %v", gotV.Size())
				}
				return nil
			},
		},
		{
			name: "null to map",
			expr: `grpc.federation.json.from('null', {"": 0})`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Mapper)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 0 {
					return fmt.Errorf("expected size 0, got %v", gotV.Size())
				}
				return nil
			},
		},
		{
			name: "null to proto message",
			expr: `grpc.federation.json.from('null', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				// Verify zero values
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				if refl.Get(idField).String() != "" {
					return fmt.Errorf("expected empty id, got %v", refl.Get(idField).String())
				}
				if refl.Get(numField).Int() != 0 {
					return fmt.Errorf("expected num 0, got %v", refl.Get(numField).Int())
				}
				return nil
			},
		},

		// Extra/missing fields handling
		// NOTE: protojson.Unmarshal by default rejects unknown fields (strict mode)
		// This is safer as it catches typos and schema mismatches
		{
			name: "extra fields in JSON cause error (strict mode)",
			expr: `grpc.federation.json.from('{"id": "test", "unknown_field": "value"}', grpc.federation.cel.test.Message{})`,
			err:  true, // unknown fields are NOT allowed by default
		},
		{
			name: "missing fields use zero values",
			expr: `grpc.federation.json.from('{"id": "only-id"}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				innerField := refl.Descriptor().Fields().ByName("inner")
				// Provided field should have its value
				if refl.Get(idField).String() != "only-id" {
					return fmt.Errorf("expected id 'only-id', got %v", refl.Get(idField).String())
				}
				// Missing fields should have zero values
				if refl.Get(numField).Int() != 0 {
					return fmt.Errorf("expected num 0 (zero value), got %v", refl.Get(numField).Int())
				}
				// Missing message field should be nil/not set
				if refl.Has(innerField) {
					return fmt.Errorf("expected inner to be not set, but it was set")
				}
				return nil
			},
		},
		{
			name: "missing all fields uses all zero values",
			expr: `grpc.federation.json.from('{}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				numField := refl.Descriptor().Fields().ByName("num")
				// All fields should have zero values
				if refl.Get(idField).String() != "" {
					return fmt.Errorf("expected empty id, got %v", refl.Get(idField).String())
				}
				if refl.Get(numField).Int() != 0 {
					return fmt.Errorf("expected num 0, got %v", refl.Get(numField).Int())
				}
				return nil
			},
		},

		// Error cases - Malformed JSON syntax
		{
			name: "malformed json - completely invalid",
			expr: `grpc.federation.json.from('not valid json at all', "")`,
			err:  true,
		},
		{
			name: "malformed json - unclosed string",
			expr: `grpc.federation.json.from('"unclosed string', "")`,
			err:  true,
		},
		{
			name: "malformed json - unclosed object",
			expr: `grpc.federation.json.from('{"id": "test"', grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "malformed json - unclosed array",
			expr: `grpc.federation.json.from('[1, 2, 3', [0])`,
			err:  true,
		},
		{
			name: "malformed json - trailing comma in object",
			expr: `grpc.federation.json.from('{"id": "test",}', grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "malformed json - trailing comma in array",
			expr: `grpc.federation.json.from('[1, 2, 3,]', [0])`,
			err:  true,
		},
		{
			name: "malformed json - single quotes instead of double",
			expr: `grpc.federation.json.from("{'id': 'test'}", grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "malformed json - unquoted key",
			expr: `grpc.federation.json.from('{id: "test"}', grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "malformed json - multiple values",
			expr: `grpc.federation.json.from('"a" "b"', "")`,
			err:  true,
		},
		{
			name: "malformed json - empty string input",
			expr: `grpc.federation.json.from('', "")`,
			err:  true,
		},
		{
			name: "malformed json - whitespace only",
			expr: `grpc.federation.json.from('   ', "")`,
			err:  true,
		},

		// Error cases - Type mismatches
		{
			name: "type mismatch - string to int",
			expr: `grpc.federation.json.from('"hello"', 0)`,
			err:  true,
		},
		{
			name: "type mismatch - int to bool",
			expr: `grpc.federation.json.from('42', false)`,
			err:  true,
		},
		{
			name: "type mismatch - bool to int",
			expr: `grpc.federation.json.from('true', 0)`,
			err:  true,
		},
		{
			name: "type mismatch - object to string",
			expr: `grpc.federation.json.from('{"key": "value"}', "")`,
			err:  true,
		},
		{
			name: "type mismatch - array to string",
			expr: `grpc.federation.json.from('[1, 2, 3]', "")`,
			err:  true,
		},
		{
			name: "type mismatch - string to message",
			expr: `grpc.federation.json.from('"just a string"', grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "type mismatch - array to message",
			expr: `grpc.federation.json.from('[1, 2, 3]', grpc.federation.cel.test.Message{})`,
			err:  true,
		},
		{
			name: "type mismatch - object to list",
			expr: `grpc.federation.json.from('{"key": "value"}', [0])`,
			err:  true,
		},
		{
			name: "type mismatch - wrong field type in message",
			expr: `grpc.federation.json.from('{"id": 123}', grpc.federation.cel.test.Message{})`,
			err:  true, // id should be string, not number
		},
		{
			name: "type mismatch - wrong nested field type",
			expr: `grpc.federation.json.from('{"id": "test", "inner": "not an object"}', grpc.federation.cel.test.Message{})`,
			err:  true, // inner should be object, not string
		},

		// Error cases - List element type mismatches
		{
			name: "list type mismatch - strings in int list",
			expr: `grpc.federation.json.from('["a", "b", "c"]', [0])`,
			err:  true,
		},
		{
			name: "list type mismatch - mixed types in int list",
			expr: `grpc.federation.json.from('[1, "two", 3]', [0])`,
			err:  true,
		},
		{
			name: "list type mismatch - objects in int list",
			expr: `grpc.federation.json.from('[{"id": "a"}, {"id": "b"}]', [0])`,
			err:  true,
		},
		{
			name: "list type mismatch - ints in string list",
			expr: `grpc.federation.json.from('[1, 2, 3]', [""])`,
			err:  true,
		},

		// Error cases - Map type mismatches
		{
			name: "map type mismatch - array to map",
			expr: `grpc.federation.json.from('[1, 2, 3]', {"": 0})`,
			err:  true,
		},
		{
			name: "map type mismatch - string to map",
			expr: `grpc.federation.json.from('"hello"', {"": 0})`,
			err:  true,
		},
		{
			name: "map type mismatch - wrong value type",
			expr: `grpc.federation.json.from('{"a": "string"}', {"": 0})`,
			err:  true, // value should be int, not string
		},
		{
			name: "map type mismatch - mixed value types",
			expr: `grpc.federation.json.from('{"a": 1, "b": "two"}', {"": 0})`,
			err:  true,
		},

		// Edge cases - Special values
		{
			name: "special value - very large int",
			expr: `grpc.federation.json.from('9223372036854775807', 0)`, // int64 max
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if int64(gotV) != 9223372036854775807 {
					return fmt.Errorf("expected max int64, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "special value - very small int",
			expr: `grpc.federation.json.from('-9223372036854775808', 0)`, // int64 min
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if int64(gotV) != -9223372036854775808 {
					return fmt.Errorf("expected min int64, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "special value - int overflow",
			expr: `grpc.federation.json.from('9223372036854775808', 0)`, // int64 max + 1
			err:  true,
		},
		{
			name: "special value - float to int truncation",
			expr: `grpc.federation.json.from('3.14', 0)`,
			err:  true, // float cannot be parsed as int
		},
		{
			name: "special value - scientific notation int",
			expr: `grpc.federation.json.from('1e10', 0)`,
			err:  true, // scientific notation parsed as float
		},
		{
			name: "special value - zero",
			expr: `grpc.federation.json.from('0', 0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV != types.IntZero {
					return fmt.Errorf("expected 0, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "special value - negative zero double",
			expr: `grpc.federation.json.from('-0.0', 0.0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if float64(gotV) != 0.0 {
					return fmt.Errorf("expected 0.0, got %v", gotV)
				}
				return nil
			},
		},

		// Edge cases - String special characters
		{
			name: "string with escaped quotes",
			expr: `grpc.federation.json.from('"hello \\"world\\""', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := `hello "world"`
				if string(gotV) != expected {
					return fmt.Errorf("expected %q, got %q", expected, gotV)
				}
				return nil
			},
		},
		{
			name: "string with backslash",
			expr: `grpc.federation.json.from('"path\\\\to\\\\file"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := `path\to\file`
				if string(gotV) != expected {
					return fmt.Errorf("expected %q, got %q", expected, gotV)
				}
				return nil
			},
		},
		{
			name: "string with newline",
			expr: `grpc.federation.json.from('"line1\\nline2"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := "line1\nline2"
				if string(gotV) != expected {
					return fmt.Errorf("expected %q, got %q", expected, gotV)
				}
				return nil
			},
		},
		{
			name: "string with tab",
			expr: `grpc.federation.json.from('"col1\\tcol2"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := "col1\tcol2"
				if string(gotV) != expected {
					return fmt.Errorf("expected %q, got %q", expected, gotV)
				}
				return nil
			},
		},
		{
			name: "string with unicode escape",
			expr: `grpc.federation.json.from('"\\u0048\\u0065\\u006c\\u006c\\u006f"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if string(gotV) != "Hello" {
					return fmt.Errorf("expected 'Hello', got %q", gotV)
				}
				return nil
			},
		},
		{
			name: "string with null byte",
			expr: `grpc.federation.json.from('"hello\\u0000world"', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := "hello\x00world"
				if string(gotV) != expected {
					return fmt.Errorf("expected %q, got %q", expected, gotV)
				}
				return nil
			},
		},
		{
			name: "empty string",
			expr: `grpc.federation.json.from('""', "")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if string(gotV) != "" {
					return fmt.Errorf("expected empty string, got %q", gotV)
				}
				return nil
			},
		},

		// Edge cases - Arrays
		{
			name: "deeply nested array - template mismatch",
			expr: `grpc.federation.json.from('[[[[1]]]]', [0])`,
			err:  true, // nested array cannot be parsed as flat int array
		},
		{
			name: "array with null elements",
			expr: `grpc.federation.json.from('[1, null, 3]', [0])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 3 {
					return fmt.Errorf("expected size 3, got %v", gotV.Size())
				}
				// null becomes zero value (0 for int)
				if gotV.Get(types.Int(1)).(types.Int) != 0 {
					return fmt.Errorf("expected null to become 0, got %v", gotV.Get(types.Int(1)))
				}
				return nil
			},
		},
		{
			name: "large array",
			expr: `grpc.federation.json.from('[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]', [0])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 20 {
					return fmt.Errorf("expected size 20, got %v", gotV.Size())
				}
				return nil
			},
		},

		// Edge cases - Proto message specifics
		{
			name: "message with only whitespace in json",
			expr: `grpc.federation.json.from('{   }', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				idField := refl.Descriptor().Fields().ByName("id")
				if refl.Get(idField).String() != "" {
					return fmt.Errorf("expected empty id, got %v", refl.Get(idField).String())
				}
				return nil
			},
		},
		{
			name: "message with camelCase field name",
			// protojson uses camelCase by default, testing that snake_case also works
			expr: `grpc.federation.json.from('{"id": "test", "num": 42}', grpc.federation.cel.test.Message{})`,
			cmp: func(got ref.Val) error {
				msg, ok := got.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got.Value())
				}
				refl := msg.ProtoReflect()
				numField := refl.Descriptor().Fields().ByName("num")
				if refl.Get(numField).Int() != 42 {
					return fmt.Errorf("expected num 42, got %v", refl.Get(numField).Int())
				}
				return nil
			},
		},
		{
			name: "list of messages with null element",
			expr: `grpc.federation.json.from('[{"id": "a"}, null, {"id": "c"}]', [grpc.federation.cel.test.Message{}])`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(traits.Lister)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV.Size().(types.Int) != 3 {
					return fmt.Errorf("expected size 3, got %v", gotV.Size())
				}
				// First element should have id "a"
				elem0 := gotV.Get(types.Int(0))
				msg0, ok := elem0.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("expected proto.Message at index 0, got %T", elem0.Value())
				}
				refl0 := msg0.ProtoReflect()
				idField := refl0.Descriptor().Fields().ByName("id")
				if refl0.Get(idField).String() != "a" {
					return fmt.Errorf("expected id 'a' at index 0, got %v", refl0.Get(idField).String())
				}
				// Second element (null) should be empty message
				elem1 := gotV.Get(types.Int(1))
				msg1, ok := elem1.Value().(proto.Message)
				if !ok {
					return fmt.Errorf("expected proto.Message at index 1, got %T", elem1.Value())
				}
				refl1 := msg1.ProtoReflect()
				if refl1.Get(idField).String() != "" {
					return fmt.Errorf("expected empty id at index 1, got %v", refl1.Get(idField).String())
				}
				return nil
			},
		},

		// Edge cases - Bytes (base64)
		{
			name: "bytes empty",
			expr: `grpc.federation.json.from('""', b"")`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if len(gotV) != 0 {
					return fmt.Errorf("expected empty bytes, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "bytes invalid base64",
			expr: `grpc.federation.json.from('"not-valid-base64!!!"', b"")`,
			err:  true,
		},
		{
			name: "bytes with padding",
			expr: `grpc.federation.json.from('"YQ=="', b"")`, // "a" with padding
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if string(gotV) != "a" {
					return fmt.Errorf("expected 'a', got %q", string(gotV))
				}
				return nil
			},
		},

		// Edge cases - Double special values
		{
			name: "double very small",
			expr: `grpc.federation.json.from('0.000000001', 0.0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if float64(gotV) != 0.000000001 {
					return fmt.Errorf("expected 0.000000001, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "double scientific notation",
			expr: `grpc.federation.json.from('1.5e10', 0.0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if float64(gotV) != 1.5e10 {
					return fmt.Errorf("expected 1.5e10, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "double negative scientific notation",
			expr: `grpc.federation.json.from('-2.5e-5', 0.0)`,
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if float64(gotV) != -2.5e-5 {
					return fmt.Errorf("expected -2.5e-5, got %v", gotV)
				}
				return nil
			},
		},

		// Edge cases - Uint
		{
			name: "uint max value",
			expr: `grpc.federation.json.from('18446744073709551615', uint(0))`, // uint64 max
			cmp: func(got ref.Val) error {
				gotV, ok := got.(types.Uint)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if uint64(gotV) != 18446744073709551615 {
					return fmt.Errorf("expected max uint64, got %v", gotV)
				}
				return nil
			},
		},
		{
			name: "uint negative value",
			expr: `grpc.federation.json.from('-1', uint(0))`,
			err:  true, // negative numbers cannot be uint
		},
	}

	reg, err := types.NewRegistry(&testpb.Message{}, &testpb.InnerMessage{})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
				cel.Types(&testpb.Message{}, &testpb.InnerMessage{}),
				cel.Lib(cellib.NewJSONLibrary(reg)),
			)
			if err != nil {
				t.Fatal(err)
			}
			ast, iss := env.Compile(test.expr)
			if iss.Err() != nil {
				t.Fatal(iss.Err())
			}
			program, err := env.Program(ast)
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]any{cellib.ContextVariableName: cellib.NewContextValue(context.Background())}
			for k, v := range test.args {
				args[k] = v
			}
			out, _, err := program.Eval(args)
			if test.err {
				if err == nil && !types.IsError(out) {
					t.Fatalf("expected error, got %v", out)
				}
			} else if err != nil {
				t.Fatal(err)
			} else if types.IsError(out) {
				t.Fatalf("unexpected error: %v", out)
			} else {
				if err := test.cmp(out); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

// Property-based tests

func TestJSONFromRoundtrip_String(t *testing.T) {
	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, "")`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(s string) bool {
		// JSON encode the string
		jsonBytes, err := json.Marshal(s)
		if err != nil {
			return true // skip invalid inputs
		}

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		// Check that the result equals the original string
		gotStr, ok := out.(types.String)
		if !ok {
			return false
		}
		return string(gotStr) == s
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJSONFromRoundtrip_Int(t *testing.T) {
	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, 0)`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(i int64) bool {
		jsonBytes, _ := json.Marshal(i)

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		gotInt, ok := out.(types.Int)
		if !ok {
			return false
		}
		return int64(gotInt) == i
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJSONFromRoundtrip_Bool(t *testing.T) {
	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, false)`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(b bool) bool {
		jsonBytes, _ := json.Marshal(b)

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		gotBool, ok := out.(types.Bool)
		if !ok {
			return false
		}
		return bool(gotBool) == b
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJSONFromRoundtrip_Double(t *testing.T) {
	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, 0.0)`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(d float64) bool {
		// Skip NaN and Inf which are not valid JSON
		if d != d { // NaN check
			return true
		}

		jsonBytes, err := json.Marshal(d)
		if err != nil {
			return true // skip invalid inputs
		}

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		gotDouble, ok := out.(types.Double)
		if !ok {
			return false
		}
		return float64(gotDouble) == d
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJSONFromRoundtrip_Message(t *testing.T) {
	reg, err := types.NewRegistry(&testpb.Message{}, &testpb.InnerMessage{})
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Types(&testpb.Message{}, &testpb.InnerMessage{}),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, grpc.federation.cel.test.Message{})`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(id string, num int32) bool {
		msg := &testpb.Message{Id: id, Num: num}
		jsonBytes, err := protojson.Marshal(msg)
		if err != nil {
			return true // skip invalid inputs
		}

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		// Verify via reflection (works with both concrete and dynamic types)
		gotMsg, ok := out.Value().(proto.Message)
		if !ok {
			return false
		}
		refl := gotMsg.ProtoReflect()
		idField := refl.Descriptor().Fields().ByName("id")
		numField := refl.Descriptor().Fields().ByName("num")
		return refl.Get(idField).String() == id && refl.Get(numField).Int() == int64(num)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestJSONFromListLengthPreservation(t *testing.T) {
	reg, err := types.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		t.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, [0])`)
	if iss.Err() != nil {
		t.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}

	f := func(nums []int64) bool {
		jsonBytes, _ := json.Marshal(nums)

		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  string(jsonBytes),
		}
		out, _, err := program.Eval(args)
		if err != nil {
			return false
		}
		if types.IsError(out) {
			return false
		}

		gotList, ok := out.(traits.Lister)
		if !ok {
			return false
		}

		return int(gotList.Size().(types.Int)) == len(nums)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Fuzzing tests

func FuzzJSONFromString(f *testing.F) {
	// Seed corpus
	f.Add(`"hello"`)
	f.Add(`""`)
	f.Add(`"日本語"`)
	f.Add(`"\u0000"`)
	f.Add(`"\n\t\r"`)
	f.Add(`null`)
	f.Add(`invalid`)
	f.Add(`"unclosed`)
	f.Add(`{}`)
	f.Add(`[]`)
	f.Add(`123`)
	f.Add(`true`)

	reg, err := types.NewRegistry()
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, "")`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, type should be correct
		if !types.IsError(out) && out != nil {
			if _, ok := out.(types.String); !ok {
				t.Errorf("unexpected type: %T", out)
			}
		}
	})
}

func FuzzJSONFromInt(f *testing.F) {
	// Seed corpus
	f.Add(`0`)
	f.Add(`-1`)
	f.Add(`9223372036854775807`)  // int64 max
	f.Add(`-9223372036854775808`) // int64 min
	f.Add(`1.5`)
	f.Add(`"not a number"`)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`{}`)
	f.Add(`true`)

	reg, err := types.NewRegistry()
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, 0)`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, type should be correct
		if !types.IsError(out) && out != nil {
			if _, ok := out.(types.Int); !ok {
				t.Errorf("unexpected type: %T", out)
			}
		}
	})
}

func FuzzJSONFromMessage(f *testing.F) {
	// Seed corpus
	f.Add(`{}`)
	f.Add(`{"id": "test", "num": 42}`)
	f.Add(`{"id": "", "num": 0}`)
	f.Add(`{"unknown_field": "value"}`)
	f.Add(`{"id": 123}`) // wrong type
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`"string"`)
	f.Add(`123`)
	f.Add(`true`)
	f.Add(`{"id": "a", "num": 1, "inner": {"id": "b"}}`)

	reg, err := types.NewRegistry(&testpb.Message{}, &testpb.InnerMessage{})
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Types(&testpb.Message{}, &testpb.InnerMessage{}),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, grpc.federation.cel.test.Message{})`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, should be a valid proto message
		if !types.IsError(out) && out != nil {
			val := out.Value()
			if _, ok := val.(proto.Message); !ok {
				t.Errorf("unexpected type: %T", val)
			}
		}
	})
}

func FuzzJSONFromList(f *testing.F) {
	// Seed corpus
	f.Add(`[]`)
	f.Add(`[1, 2, 3]`)
	f.Add(`[1, "mixed", true]`)
	f.Add(`[[1, 2], [3, 4]]`) // nested
	f.Add(`null`)
	f.Add(`{}`)
	f.Add(`"string"`)
	f.Add(`123`)
	f.Add(`[1.5, 2.5]`)

	reg, err := types.NewRegistry()
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, [0])`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, should be a list
		if !types.IsError(out) && out != nil {
			if _, ok := out.(traits.Lister); !ok {
				t.Errorf("unexpected type: %T", out)
			}
		}
	})
}

func FuzzJSONFromBytes(f *testing.F) {
	// Seed corpus - base64 encoded strings
	f.Add(`"SGVsbG8gV29ybGQ="`) // "Hello World"
	f.Add(`""`)
	f.Add(`"YWJj"`)             // "abc"
	f.Add(`"not valid base64"`) // invalid base64 but valid json string
	f.Add(`null`)
	f.Add(`123`)
	f.Add(`[]`)
	f.Add(`{}`)

	reg, err := types.NewRegistry()
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, b"")`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, should be bytes
		if !types.IsError(out) && out != nil {
			if _, ok := out.(types.Bytes); !ok {
				t.Errorf("unexpected type: %T", out)
			}
		}
	})
}

func FuzzJSONFromMap(f *testing.F) {
	// Seed corpus
	f.Add(`{}`)
	f.Add(`{"a": 1, "b": 2}`)
	f.Add(`{"key": 0}`)
	f.Add(`{"": 0}`)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`"string"`)
	f.Add(`123`)
	f.Add(`{"a": "string"}`) // wrong value type
	f.Add(`{"a": 1, "b": "mixed"}`)

	reg, err := types.NewRegistry()
	if err != nil {
		f.Fatal(err)
	}
	lib := cellib.NewJSONLibrary(reg)

	env, err := cel.NewEnv(
		cel.Variable(cellib.ContextVariableName, cel.ObjectType(cellib.ContextTypeName)),
		cel.Variable("jsonStr", cel.StringType),
		cel.Lib(lib),
	)
	if err != nil {
		f.Fatal(err)
	}

	ast, iss := env.Compile(`grpc.federation.json.from(jsonStr, {"": 0})`)
	if iss.Err() != nil {
		f.Fatal(iss.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		args := map[string]any{
			cellib.ContextVariableName: cellib.NewContextValue(context.Background()),
			"jsonStr":                  jsonStr,
		}
		// Should not panic
		out, _, _ := program.Eval(args)

		// If not an error, should be a map
		if !types.IsError(out) && out != nil {
			if _, ok := out.(traits.Mapper); !ok {
				t.Errorf("unexpected type: %T", out)
			}
		}
	})
}

