package resolver

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/types"
)

type CELRegistry struct {
	*celtypes.Registry
	messageMap       map[string]*Message
	enumTypeMap      map[*celtypes.Type]*Enum
	enumValueMap     map[string]*EnumValue
	usedEnumValueMap map[*EnumValue]struct{}
	errs             []error
}

func (r *CELRegistry) clear() {
	r.errs = r.errs[:0]
	r.usedEnumValueMap = make(map[*EnumValue]struct{})
}

func (r *CELRegistry) errors() []error {
	return r.errs
}

func (r *CELRegistry) EnumValue(enumName string) ref.Val {
	value := r.Registry.EnumValue(enumName)
	if !celtypes.IsError(value) {
		enumValue, exists := r.enumValueMap[enumName]
		if exists {
			r.usedEnumValueMap[enumValue] = struct{}{}
		}
		return value
	}
	return value
}

func (r *CELRegistry) FindStructFieldType(structType, fieldName string) (*celtypes.FieldType, bool) {
	fieldType, found := r.Registry.FindStructFieldType(structType, fieldName)
	if msg := r.messageMap[structType]; msg != nil {
		if field := msg.Field(fieldName); field != nil {
			if field.Type.Kind == types.Enum {
				// HACK: cel-go currently does not support enum types, so it will always be an int type.
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
		if typ.Message.IsMapEntry {
			return cel.MapType(toCELType(typ.Message.Fields[0].Type), toCELType(typ.Message.Fields[1].Type))
		}
		if typ.Message.IsEnumSelector() {
			return toEnumSelectorCELType(typ.Message)
		}
		return cel.ObjectType(typ.Message.FQDN())
	case types.Bytes:
		return cel.BytesType
	case types.Enum:
		return celtypes.NewOpaqueType(typ.Enum.FQDN(), cel.IntType)
	}
	return cel.NullType
}

func toEnumSelectorCELType(sel *Message) *cel.Type {
	trueType := toCELType(sel.Fields[0].Type)
	falseType := toCELType(sel.Fields[1].Type)
	return celtypes.NewOpaqueType(grpcfedcel.EnumSelectorFQDN, trueType, falseType)
}

func newCELRegistry(messageMap map[string]*Message, enumValueMap map[string]*EnumValue) *CELRegistry {
	return &CELRegistry{
		Registry:         celtypes.NewEmptyRegistry(),
		messageMap:       messageMap,
		enumTypeMap:      make(map[*celtypes.Type]*Enum),
		enumValueMap:     enumValueMap,
		usedEnumValueMap: make(map[*EnumValue]struct{}),
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

type CELPluginFunctionSignature struct {
	ID     string
	Args   []*cel.Type
	Return *cel.Type
}

// Signatures returns signature for overload function.
// If there is an enum type, we need to prepare functions for both `opaque<int>` and `int` patterns.
func (f *CELFunction) Signatures() []*CELPluginFunctionSignature {
	sigs := f.signatures(f.Args, f.Return)
	sigMap := make(map[string]struct{})
	ret := make([]*CELPluginFunctionSignature, 0, len(sigs))
	for _, sig := range sigs {
		if _, exists := sigMap[sig.ID]; exists {
			continue
		}
		ret = append(ret, sig)
		sigMap[sig.ID] = struct{}{}
	}
	return ret
}

func (f *CELFunction) signatures(args []*Type, ret *Type) []*CELPluginFunctionSignature {
	var sigs []*CELPluginFunctionSignature
	for idx, arg := range args {
		if arg.Kind == types.Enum {
			sigs = append(sigs,
				f.signatures(
					append(append(append([]*Type{}, args[:idx]...), Int32Type), args[idx+1:]...),
					ret,
				)...,
			)
		}
	}
	if ret.Kind == types.Enum {
		sigs = append(sigs, f.signatures(args, Int32Type)...)
	}
	var celArgs []*cel.Type
	for _, arg := range args {
		celArgs = append(celArgs, ToCELType(arg))
	}
	sigs = append(sigs, &CELPluginFunctionSignature{
		ID:     f.toSignatureID(append(args, ret)),
		Args:   celArgs,
		Return: toCELType(ret),
	})
	return sigs
}

func (f *CELFunction) toSignatureID(t []*Type) string {
	var typeNames []string
	for _, tt := range t {
		if tt == nil {
			continue
		}
		typeNames = append(typeNames, tt.FQDN())
	}
	return strings.ReplaceAll(strings.Join(append([]string{f.ID}, typeNames...), "_"), ".", "_")
}

func (plugin *CELPlugin) CompileOptions() []cel.EnvOption {
	var opts []cel.EnvOption
	for _, fn := range plugin.Functions {
		var (
			overload cel.FunctionOpt
			bindFunc = cel.FunctionBinding(func(args ...ref.Val) ref.Val { return nil })
		)
		for _, sig := range fn.Signatures() {
			if fn.Receiver != nil {
				overload = cel.MemberOverload(sig.ID, sig.Args, sig.Return, bindFunc)
			} else {
				overload = cel.Overload(sig.ID, sig.Args, sig.Return, bindFunc)
			}
			opts = append(opts, cel.Function(fn.Name, overload))
		}
	}
	return opts
}

func (plugin *CELPlugin) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
