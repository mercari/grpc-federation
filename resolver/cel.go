//go:build !tinygo.wasm

package resolver

import (
	"fmt"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/types"
)

type CELRegistry struct {
	*celtypes.Registry
	messageMap  map[string]*Message
	enumTypeMap map[*celtypes.Type]*Enum
	errs        []error
}

func (r *CELRegistry) clearErrors() {
	r.errs = r.errs[:0]
}

func (r *CELRegistry) errors() []error {
	return r.errs
}

func (r *CELRegistry) FindStructFieldType(structType, fieldName string) (*celtypes.FieldType, bool) {
	fieldType, found := r.Registry.FindStructFieldType(structType, fieldName)
	if msg := r.messageMap[structType]; msg != nil {
		if field := msg.Field(fieldName); field != nil {
			if field.Type.Kind == types.Enum {
				// HACK: cel-go currently does not support enum types,s o it will always be an int type.
				// Therefore, in case of enum type, copy the *Type of int type, and then map the created *Type to the enum type.
				// Finally, at `fromCELType` phase, lookup enum type from *Type address.
				copiedType := *fieldType.Type
				fieldType.Type = &copiedType
				r.enumTypeMap[fieldType.Type] = field.Type.Enum
			}
		}
		oneof := msg.Oneof(fieldName)
		if !found && oneof != nil {
			if !oneof.IsSameType() {
				r.errs = append(r.errs, fmt.Errorf(
					`"%[1]s" type has "%[2]s" as oneof name, but "%[2]s" has a difference type and cannot be accessed directly, so "%[2]s" becomes an undefined field`,
					structType, fieldName,
				))
				return fieldType, found
			}
			// If we refer directly to the name of oneof and all fields in oneof have the same type,
			// we can refer to the value of a field that is not nil.
			return &celtypes.FieldType{
				Type: ToCELType(oneof.Fields[0].Type),
			}, true
		}
	}
	return fieldType, found
}

func (r *CELRegistry) LookupEnum(t *celtypes.Type) (*Enum, bool) {
	enum, found := r.enumTypeMap[t]
	return enum, found
}

func ToCELType(typ *Type) *cel.Type {
	if typ.Repeated {
		return cel.ListType(toCELType(typ))
	}
	return toCELType(typ)
}

func toCELType(typ *Type) *cel.Type {
	switch typ.Kind {
	case types.Double, types.Float:
		return cel.DoubleType
	case types.Int32, types.Int64, types.Sfixed32, types.Sfixed64, types.Sint32, types.Sint64:
		return cel.IntType
	case types.Uint32, types.Uint64, types.Fixed32, types.Fixed64:
		return cel.UintType
	case types.Bool:
		return cel.BoolType
	case types.String:
		return cel.StringType
	case types.Group, types.Message:
		if typ.Message == nil {
			return cel.NullType
		}
		return cel.ObjectType(typ.Message.FQDN())
	case types.Bytes:
		return cel.BytesType
	case types.Enum:
		return cel.IntType
	}
	return cel.NullType
}

func newCELRegistry(messageMap map[string]*Message) *CELRegistry {
	return &CELRegistry{
		Registry:    celtypes.NewEmptyRegistry(),
		messageMap:  messageMap,
		enumTypeMap: make(map[*celtypes.Type]*Enum),
	}
}

func (r *CELRegistry) RegisterFiles(files ...*descriptorpb.FileDescriptorProto) error {
	registryFiles, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: files,
	})
	if err != nil {
		return err
	}
	for _, file := range files {
		rf, err := protodesc.NewFile(file, registryFiles)
		if err != nil {
			return err
		}
		if err := r.Registry.RegisterDescriptor(rf); err != nil {
			return err
		}
	}
	return nil
}

func NewCELStandardLibraryMessageType(pkgName, msgName string) *Type {
	return &Type{
		Kind: types.Message,
		Message: &Message{
			File: &File{
				Package: &Package{
					Name: "grpc.federation." + pkgName,
				},
				GoPackage: &GoPackage{
					Name:       "grpcfedcel",
					ImportPath: "github.com/mercari/grpc-federation/grpc/federation/cel",
					AliasName:  "grpcfedcel",
				},
			},
			Name: msgName,
		},
	}
}

func (plugin *CELPlugin) LibraryName() string {
	return plugin.Name
}

func (f *CELFunction) CELArgs() []*cel.Type {
	ret := make([]*cel.Type, 0, len(f.Args))
	for _, arg := range f.Args {
		ret = append(ret, ToCELType(arg))
	}
	return ret
}

func (f *CELFunction) CELReturn() *cel.Type {
	return ToCELType(f.Return)
}

func (plugin *CELPlugin) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, fn := range plugin.Functions {
		var (
			overload cel.FunctionOpt
			bindFunc = cel.FunctionBinding(func(args ...ref.Val) ref.Val { return nil })
		)
		if fn.Receiver != nil {
			overload = cel.MemberOverload(fn.ID, fn.CELArgs(), fn.CELReturn(), bindFunc)
		} else {
			overload = cel.Overload(fn.ID, fn.CELArgs(), fn.CELReturn(), bindFunc)
		}
		opts = append(opts, cel.Function(fn.Name, overload))
	}
	return opts
}

func (plugin *CELPlugin) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
