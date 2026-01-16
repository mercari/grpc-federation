package resolver

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
	"github.com/mercari/grpc-federation/types"
)

type CELRegistry struct {
	*celtypes.Registry
	registryFiles     *protoregistry.Files
	registeredFileMap map[string]struct{}
	messageMap        map[string]*Message
	enumTypeMap       map[*celtypes.Type]*Enum
	enumValueMap      map[string]*EnumValue
	usedEnumValueMap  map[*EnumValue]struct{}
	errs              []error
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
	return toCELType(typ)
}

func toCELType(typ *Type) *cel.Type {
	if typ.Repeated {
		t := typ.Clone()
		t.Repeated = false
		return cel.ListType(toCELType(t))
	}
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
		Registry:          celtypes.NewEmptyRegistry(),
		registryFiles:     new(protoregistry.Files),
		registeredFileMap: make(map[string]struct{}),
		messageMap:        messageMap,
		enumTypeMap:       make(map[*celtypes.Type]*Enum),
		enumValueMap:      enumValueMap,
		usedEnumValueMap:  make(map[*EnumValue]struct{}),
	}
}

func (r *CELRegistry) RegisterFiles(fds ...*descriptorpb.FileDescriptorProto) error {
	fileMap := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, fd := range fds {
		fileName := fd.GetName()
		if _, ok := fileMap[fileName]; ok {
			return fmt.Errorf("file appears multiple times: %q", fileName)
		}
		fileMap[fileName] = fd
	}
	for _, fd := range fileMap {
		if err := r.registerFileDeps(fd, fileMap); err != nil {
			return err
		}
	}
	return nil
}

func (r *CELRegistry) registerFileDeps(fd *descriptorpb.FileDescriptorProto, fileMap map[string]*descriptorpb.FileDescriptorProto) error {
	// set the entry to nil while descending into a file's dependencies to detect cycles.
	fileName := fd.GetName()
	fileMap[fileName] = nil
	for _, dep := range fd.GetDependency() {
		depFD, ok := fileMap[dep]
		if depFD == nil {
			if ok {
				return fmt.Errorf("import cycle in file: %q", dep)
			}
			continue
		}
		if err := r.registerFileDeps(depFD, fileMap); err != nil {
			return err
		}
	}
	// delete the entry once dependencies are processed.
	delete(fileMap, fileName)
	if _, exists := r.registeredFileMap[fileName]; exists {
		return nil
	}

	f, err := protodesc.NewFile(fd, r.registryFiles)
	if err != nil {
		return err
	}
	if err := r.registryFiles.RegisterFile(f); err != nil {
		return err
	}
	if err := r.Registry.RegisterDescriptor(f); err != nil {
		return err
	}
	r.registeredFileMap[fileName] = struct{}{}
	return nil
}

var (
	cachedStandardMessageType   = make(map[string]*Type)
	cachedStandardMessageTypeMu sync.Mutex
)

func NewCELStandardLibraryMessageType(pkgName, msgName string) *Type {
	cachedStandardMessageTypeMu.Lock()
	defer cachedStandardMessageTypeMu.Unlock()
	fqdn := fmt.Sprintf("%s.%s", pkgName, msgName)
	if typ, exists := cachedStandardMessageType[fqdn]; exists {
		return typ
	}
	ret := &Type{
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
	cachedStandardMessageType[fqdn] = ret
	return ret
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
		Return: ToCELType(ret),
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
	bindFunc := cel.FunctionBinding(func(args ...ref.Val) ref.Val { return nil })

	for _, fn := range plugin.Functions {
		if fn.Receiver != nil {
			// For member functions, sig.Args already includes receiver as first element
			// We need to insert context AFTER receiver: [receiver, context, arg1, arg2, ...]
			var memberOpts []grpcfedcel.BindMemberFunctionOpt
			for _, sig := range fn.Signatures() {
				if len(sig.Args) == 0 {
					continue
				}
				argsWithContext := append([]*cel.Type{sig.Args[0], cel.ObjectType(grpcfedcel.ContextTypeName)}, sig.Args[1:]...)
				memberOpts = append(memberOpts, grpcfedcel.BindMemberFunctionOpt{
					FunctionOpt: cel.MemberOverload(sig.ID, argsWithContext, sig.Return, bindFunc),
				})
			}
			opts = append(opts, grpcfedcel.BindMemberFunction(fn.Name, memberOpts...)...)
		} else {
			// For regular functions, insert context at the beginning
			var funcOpts []grpcfedcel.BindFunctionOpt
			for _, sig := range fn.Signatures() {
				// Regular function signature: [context, arg1, arg2, ...]
				argsWithContext := append([]*cel.Type{cel.ObjectType(grpcfedcel.ContextTypeName)}, sig.Args...)
				funcOpts = append(funcOpts, grpcfedcel.BindFunctionOpt{
					FunctionOpt: cel.Overload(sig.ID, argsWithContext, sig.Return, bindFunc),
				})
			}
			opts = append(opts, grpcfedcel.BindFunction(fn.Name, funcOpts...)...)
		}
	}
	return opts
}

func (plugin *CELPlugin) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
