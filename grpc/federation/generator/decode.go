package generator

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"github.com/mercari/grpc-federation/grpc/federation/generator/plugin"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

type CodeGeneratorRequest struct {
	ProtoPath           string
	OutDir              string
	Files               []*plugin.ProtoCodeGeneratorResponse_File
	GRPCFederationFiles []*resolver.File
}

func ToCodeGeneratorRequest(r io.Reader) (*CodeGeneratorRequest, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var v plugin.CodeGeneratorRequest
	if err := proto.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return newDecoder(v.GetReference()).toCodeGeneratorRequest(&v)
}

type decoder struct {
	ref            *plugin.Reference
	fileMap        map[string]*resolver.File
	pkgMap         map[string]*resolver.Package
	enumMap        map[string]*resolver.Enum
	enumValueMap   map[string]*resolver.EnumValue
	msgMap         map[string]*resolver.Message
	fieldMap       map[string]*resolver.Field
	oneofMap       map[string]*resolver.Oneof
	svcMap         map[string]*resolver.Service
	mtdMap         map[string]*resolver.Method
	celPluginMap   map[string]*resolver.CELPlugin
	graphMap       map[string]*resolver.MessageDependencyGraph
	varDefMap      map[string]*resolver.VariableDefinition
	varDefGroupMap map[string]resolver.VariableDefinitionGroup
}

func newDecoder(ref *plugin.Reference) *decoder {
	return &decoder{
		ref:            ref,
		fileMap:        make(map[string]*resolver.File),
		pkgMap:         make(map[string]*resolver.Package),
		enumMap:        make(map[string]*resolver.Enum),
		enumValueMap:   make(map[string]*resolver.EnumValue),
		msgMap:         make(map[string]*resolver.Message),
		fieldMap:       make(map[string]*resolver.Field),
		oneofMap:       make(map[string]*resolver.Oneof),
		svcMap:         make(map[string]*resolver.Service),
		mtdMap:         make(map[string]*resolver.Method),
		celPluginMap:   make(map[string]*resolver.CELPlugin),
		graphMap:       make(map[string]*resolver.MessageDependencyGraph),
		varDefMap:      make(map[string]*resolver.VariableDefinition),
		varDefGroupMap: make(map[string]resolver.VariableDefinitionGroup),
	}
}

func (d *decoder) toCodeGeneratorRequest(req *plugin.CodeGeneratorRequest) (*CodeGeneratorRequest, error) {
	grpcFederationFiles, err := d.toFiles(req.GetGrpcFederationFileIds())
	if err != nil {
		return nil, err
	}
	return &CodeGeneratorRequest{
		Files:               req.GetFiles(),
		GRPCFederationFiles: grpcFederationFiles,
	}, nil
}

func (d *decoder) toFiles(ids []string) ([]*resolver.File, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.File, 0, len(ids))
	for _, id := range ids {
		file, err := d.toFile(id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, file)
	}
	return ret, nil
}

func (d *decoder) toFile(id string) (*resolver.File, error) {
	if id == "" {
		return nil, nil
	}
	if file, exists := d.fileMap[id]; exists {
		return file, nil
	}
	ret := &resolver.File{}
	d.fileMap[id] = ret

	file, exists := d.ref.FileMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find file reference: %s", id)
	}
	pkg, err := d.toPackage(file.GetPackage())
	if err != nil {
		return nil, err
	}
	svcs, err := d.toServices(file.GetServiceIds())
	if err != nil {
		return nil, err
	}
	msgs, err := d.toMessages(file.GetMessageIds())
	if err != nil {
		return nil, err
	}
	enums, err := d.toEnums(file.GetEnumIds())
	if err != nil {
		return nil, err
	}
	celPlugins, err := d.toCELPlugins(file.GetCelPluginIds())
	if err != nil {
		return nil, err
	}
	importFiles, err := d.toFiles(file.GetImportFileIds())
	if err != nil {
		return nil, err
	}
	ret.Name = file.GetName()
	ret.Package = pkg
	ret.GoPackage = d.toGoPackage(file.GetGoPackage())
	ret.Services = svcs
	ret.Messages = msgs
	ret.Enums = enums
	ret.CELPlugins = celPlugins
	ret.ImportFiles = importFiles
	return ret, nil
}

func (d *decoder) toPackage(pkg *plugin.Package) (*resolver.Package, error) {
	if pkg == nil {
		return nil, nil
	}
	files, err := d.toFiles(pkg.GetFileIds())
	if err != nil {
		return nil, err
	}
	return &resolver.Package{
		Name:  pkg.GetName(),
		Files: files,
	}, nil
}

func (d *decoder) toGoPackage(pkg *plugin.GoPackage) *resolver.GoPackage {
	if pkg == nil {
		return nil
	}
	return &resolver.GoPackage{
		Name:       pkg.GetName(),
		ImportPath: pkg.GetImportPath(),
		AliasName:  pkg.GetAliasName(),
	}
}

func (d *decoder) toServices(ids []string) ([]*resolver.Service, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Service, 0, len(ids))
	for _, id := range ids {
		svc, err := d.toService(id)
		if err != nil {
			return nil, err
		}
		if svc == nil {
			continue
		}
		ret = append(ret, svc)
	}
	return ret, nil
}

func (d *decoder) toService(id string) (*resolver.Service, error) {
	if id == "" {
		return nil, nil
	}
	if svc, exists := d.svcMap[id]; exists {
		return svc, nil
	}
	svc, exists := d.ref.ServiceMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find service reference: %s", id)
	}
	ret := &resolver.Service{Name: svc.GetName()}
	d.svcMap[id] = ret

	methods, err := d.toMethods(svc.GetMethodIds())
	if err != nil {
		return nil, err
	}
	file, err := d.toFile(svc.GetFileId())
	if err != nil {
		return nil, err
	}
	msgs, err := d.toMessages(svc.GetMessageIds())
	if err != nil {
		return nil, err
	}

	msgArgs, err := d.toMessages(svc.GetMessageArgIds())
	if err != nil {
		return nil, err
	}
	celPlugins, err := d.toCELPlugins(svc.GetCelPluginIds())
	if err != nil {
		return nil, err
	}
	rule, err := d.toServiceRule(svc.GetRule())
	if err != nil {
		return nil, err
	}
	ret.Methods = methods
	ret.File = file
	ret.Messages = msgs
	ret.MessageArgs = msgArgs
	ret.CELPlugins = celPlugins
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toServiceRule(rule *plugin.ServiceRule) (*resolver.ServiceRule, error) {
	if rule == nil {
		return nil, nil
	}
	env, err := d.toEnv(rule.GetEnv())
	if err != nil {
		return nil, err
	}
	vars, err := d.toServiceVariables(rule.GetVars())
	if err != nil {
		return nil, err
	}
	return &resolver.ServiceRule{
		Env:  env,
		Vars: vars,
	}, nil
}

func (d *decoder) toEnv(env *plugin.Env) (*resolver.Env, error) {
	vars, err := d.toEnvVars(env.GetVars())
	if err != nil {
		return nil, err
	}
	return &resolver.Env{
		Vars: vars,
	}, nil
}

func (d *decoder) toEnvVars(vars []*plugin.EnvVar) ([]*resolver.EnvVar, error) {
	if len(vars) == 0 {
		return nil, nil
	}
	ret := make([]*resolver.EnvVar, 0, len(vars))
	for _, envVar := range vars {
		v, err := d.toEnvVar(envVar)
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (d *decoder) toEnvVar(v *plugin.EnvVar) (*resolver.EnvVar, error) {
	if v == nil {
		return nil, nil
	}
	typ, err := d.toType(v.GetType())
	if err != nil {
		return nil, err
	}
	return &resolver.EnvVar{
		Name:   v.GetName(),
		Type:   typ,
		Option: d.toEnvVarOption(v.GetOption()),
	}, nil
}

func (d *decoder) toEnvVarOption(opt *plugin.EnvVarOption) *resolver.EnvVarOption {
	if opt == nil {
		return nil
	}
	return &resolver.EnvVarOption{
		Alternate: opt.GetAlternate(),
		Default:   opt.GetDefault(),
		Required:  opt.GetRequired(),
		Ignored:   opt.GetIgnored(),
	}
}

func (d *decoder) toServiceVariables(vars []*plugin.ServiceVariable) ([]*resolver.ServiceVariable, error) {
	if len(vars) == 0 {
		return nil, nil
	}
	ret := make([]*resolver.ServiceVariable, 0, len(vars))
	for _, svcVar := range vars {
		v, err := d.toServiceVariable(svcVar)
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (d *decoder) toServiceVariable(v *plugin.ServiceVariable) (*resolver.ServiceVariable, error) {
	if v == nil {
		return nil, nil
	}
	ret := &resolver.ServiceVariable{
		Name: v.GetName(),
	}
	ifValue, err := d.toCELValue(v.GetIf())
	if err != nil {
		return nil, err
	}
	expr, err := d.toServiceVariableExpr(v.GetExpr())
	if err != nil {
		return nil, err
	}
	ret.If = ifValue
	ret.Expr = expr
	return ret, nil
}

func (d *decoder) toServiceVariableExpr(expr *plugin.ServiceVariableExpr) (*resolver.ServiceVariableExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.ServiceVariableExpr{}
	typ, err := d.toType(expr.GetType())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValue(expr.GetBy())
	if err != nil {
		return nil, err
	}
	mapExpr, err := d.toMapExpr(expr.GetMap())
	if err != nil {
		return nil, err
	}
	msgExpr, err := d.toMessageExpr(expr.GetMessage())
	if err != nil {
		return nil, err
	}
	enumExpr, err := d.toEnumExpr(expr.GetEnum())
	if err != nil {
		return nil, err
	}
	validationExpr, err := d.toServiceVariableValidationExpr(expr.GetValidation())
	if err != nil {
		return nil, err
	}

	ret.Type = typ
	ret.By = by
	ret.Map = mapExpr
	ret.Message = msgExpr
	ret.Enum = enumExpr
	ret.Validation = validationExpr
	return ret, nil
}

func (d *decoder) toServiceVariableValidationExpr(expr *plugin.ServiceVariableValidationExpr) (*resolver.ServiceVariableValidationExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.ServiceVariableValidationExpr{}
	ifValue, err := d.toCELValue(expr.GetIf())
	if err != nil {
		return nil, err
	}
	msg, err := d.toCELValue(expr.GetMessage())
	if err != nil {
		return nil, err
	}
	ret.If = ifValue
	ret.Message = msg
	return ret, nil
}

func (d *decoder) toMessages(ids []string) ([]*resolver.Message, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Message, 0, len(ids))
	for _, id := range ids {
		msg, err := d.toMessage(id)
		if err != nil {
			return nil, err
		}
		if msg == nil {
			continue
		}
		ret = append(ret, msg)
	}
	return ret, nil
}

func (d *decoder) toMessage(id string) (*resolver.Message, error) {
	if id == "" {
		return nil, nil
	}
	if msg, exists := d.msgMap[id]; exists {
		return msg, nil
	}
	msg, exists := d.ref.MessageMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find message reference: %s", id)
	}
	ret := &resolver.Message{
		Name:       msg.GetName(),
		IsMapEntry: msg.GetIsMapEntry(),
	}
	d.msgMap[id] = ret

	file, err := d.toFile(msg.GetFileId())
	if err != nil {
		return nil, err
	}
	parent, err := d.toMessage(msg.GetParentMessageId())
	if err != nil {
		return nil, err
	}
	nestedMsgs, err := d.toMessages(msg.GetNestedMessageIds())
	if err != nil {
		return nil, err
	}
	enums, err := d.toEnums(msg.GetEnumIds())
	if err != nil {
		return nil, err
	}
	fields, err := d.toFields(msg.GetFieldIds())
	if err != nil {
		return nil, err
	}
	oneofs, err := d.toOneofs(msg.GetOneofIds())
	if err != nil {
		return nil, err
	}
	rule, err := d.toMessageRule(msg.GetRule())
	if err != nil {
		return nil, err
	}
	ret.File = file
	ret.ParentMessage = parent
	ret.NestedMessages = nestedMsgs
	ret.Enums = enums
	ret.Fields = fields
	ret.Oneofs = oneofs
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toFields(ids []string) ([]*resolver.Field, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Field, 0, len(ids))
	for _, id := range ids {
		field, err := d.toField(id)
		if err != nil {
			return nil, err
		}
		if field == nil {
			continue
		}
		ret = append(ret, field)
	}
	return ret, nil
}

func (d *decoder) toField(id string) (*resolver.Field, error) {
	if id == "" {
		return nil, nil
	}
	if field, exists := d.fieldMap[id]; exists {
		return field, nil
	}
	field, exists := d.ref.FieldMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find field reference: %s", id)
	}
	ret := &resolver.Field{Name: field.GetName()}
	d.fieldMap[id] = ret
	typ, err := d.toType(field.GetType())
	if err != nil {
		return nil, err
	}
	oneof, err := d.toOneof(field.GetOneofId())
	if err != nil {
		return nil, err
	}
	msg, err := d.toMessage(field.GetMessageId())
	if err != nil {
		return nil, err
	}
	rule, err := d.toFieldRule(field.GetRule())
	if err != nil {
		return nil, err
	}
	ret.Type = typ
	ret.Oneof = oneof
	ret.Message = msg
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toFieldRule(rule *plugin.FieldRule) (*resolver.FieldRule, error) {
	if rule == nil {
		return nil, nil
	}
	ret := &resolver.FieldRule{
		CustomResolver:        rule.GetCustomResolver(),
		MessageCustomResolver: rule.GetMessageCustomResolver(),
	}
	value, err := d.toValue(rule.GetValue())
	if err != nil {
		return nil, err
	}
	aliases, err := d.toFields(rule.GetAliasIds())
	if err != nil {
		return nil, err
	}
	autoBindField, err := d.toAutoBindField(rule.GetAutoBindField())
	if err != nil {
		return nil, err
	}
	fieldOneofRule, err := d.toFieldOneofRule(rule.GetOneofRule())
	if err != nil {
		return nil, err
	}
	ret.Value = value
	ret.Aliases = aliases
	ret.AutoBindField = autoBindField
	ret.Oneof = fieldOneofRule
	return ret, nil
}

func (d *decoder) toValue(value *plugin.Value) (*resolver.Value, error) {
	if value == nil {
		return nil, nil
	}
	ret := &resolver.Value{Inline: value.GetInline()}
	cel, err := d.toCELValue(value.GetCel())
	if err != nil {
		return nil, err
	}
	if cel == nil {
		return nil, fmt.Errorf("failed to convert cel to value: %s", value.GetCel())
	}
	ret.CEL = cel
	return ret, nil
}

func (d *decoder) toCELValues(values []*plugin.CELValue) ([]*resolver.CELValue, error) {
	if values == nil {
		return nil, nil
	}
	ret := make([]*resolver.CELValue, 0, len(values))
	for _, value := range values {
		v, err := d.toCELValue(value)
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (d *decoder) toCELValue(value *plugin.CELValue) (*resolver.CELValue, error) {
	if value == nil {
		return nil, nil
	}
	ret := &resolver.CELValue{
		Expr: value.GetExpr(),
	}
	out, err := d.toType(value.GetOut())
	if err != nil {
		return nil, err
	}
	ret.Out = out
	return ret, nil
}

func (d *decoder) toAutoBindField(field *plugin.AutoBindField) (*resolver.AutoBindField, error) {
	if field == nil {
		return nil, nil
	}
	ret := &resolver.AutoBindField{}
	def, err := d.toVariableDefinition(field.GetVariableDefinitionId())
	if err != nil {
		return nil, err
	}
	f, err := d.toField(field.GetFieldId())
	if err != nil {
		return nil, err
	}
	ret.VariableDefinition = def
	ret.Field = f
	return ret, nil
}

func (d *decoder) toVariableDefinitionSet(set *plugin.VariableDefinitionSet) (*resolver.VariableDefinitionSet, error) {
	if set == nil {
		return nil, nil
	}
	defs, err := d.toVariableDefinitions(set.GetVariableDefinitionIds())
	if err != nil {
		return nil, err
	}
	groups, err := d.toVariableDefinitionGroups(set.GetVariableDefinitionGroupIds())
	if err != nil {
		return nil, err
	}
	graph, err := d.toMessageDependencyGraph(set.GetDependencyGraphId())
	if err != nil {
		return nil, err
	}
	return &resolver.VariableDefinitionSet{
		Defs:   defs,
		Groups: groups,
		Graph:  graph,
	}, nil
}

func (d *decoder) toFieldOneofRule(rule *plugin.FieldOneofRule) (*resolver.FieldOneofRule, error) {
	if rule == nil {
		return nil, nil
	}
	ret := &resolver.FieldOneofRule{Default: rule.GetDefault()}
	ifValue, err := d.toCELValue(rule.GetIf())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValue(rule.GetBy())
	if err != nil {
		return nil, err
	}
	defSet, err := d.toVariableDefinitionSet(rule.GetDefSet())
	if err != nil {
		return nil, err
	}
	ret.If = ifValue
	ret.By = by
	ret.DefSet = defSet
	return ret, nil
}

func (d *decoder) toOneofs(ids []string) ([]*resolver.Oneof, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Oneof, 0, len(ids))
	for _, id := range ids {
		oneof, err := d.toOneof(id)
		if err != nil {
			return nil, err
		}
		if oneof == nil {
			continue
		}
		ret = append(ret, oneof)
	}
	return ret, nil
}

func (d *decoder) toOneof(id string) (*resolver.Oneof, error) {
	if id == "" {
		return nil, nil
	}
	if oneof, exists := d.oneofMap[id]; exists {
		return oneof, nil
	}
	oneof, exists := d.ref.OneofMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find oneof reference %s", id)
	}
	ret := &resolver.Oneof{Name: oneof.GetName()}
	d.oneofMap[id] = ret

	msg, err := d.toMessage(oneof.GetMessageId())
	if err != nil {
		return nil, err
	}
	fields, err := d.toFields(oneof.GetFieldIds())
	if err != nil {
		return nil, err
	}
	ret.Message = msg
	ret.Fields = fields
	return ret, nil
}

func (d *decoder) toMessageRule(rule *plugin.MessageRule) (*resolver.MessageRule, error) {
	if rule == nil {
		return nil, nil
	}
	msgArg, err := d.toMessage(rule.GetMessageArgumentId())
	if err != nil {
		return nil, err
	}
	aliases, err := d.toMessages(rule.GetAliasIds())
	if err != nil {
		return nil, err
	}
	defSet, err := d.toVariableDefinitionSet(rule.GetDefSet())
	if err != nil {
		return nil, err
	}
	return &resolver.MessageRule{
		CustomResolver:  rule.GetCustomResolver(),
		MessageArgument: msgArg,
		Aliases:         aliases,
		DefSet:          defSet,
	}, nil
}

func (d *decoder) toMessageDependencyGraph(id string) (*resolver.MessageDependencyGraph, error) {
	if id == "" {
		return nil, nil
	}
	if graph, exists := d.graphMap[id]; exists {
		return graph, nil
	}
	graph, exists := d.ref.GraphMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find message dependency graph reference: %s", id)
	}
	ret := &resolver.MessageDependencyGraph{}
	d.graphMap[id] = ret

	roots, err := d.toMessageDependencyGraphNodes(graph.GetRoots())
	if err != nil {
		return nil, err
	}
	ret.Roots = roots
	return ret, nil
}

func (d *decoder) toMessageDependencyGraphNodes(nodes []*plugin.MessageDependencyGraphNode) ([]*resolver.MessageDependencyGraphNode, error) {
	if nodes == nil {
		return nil, nil
	}
	ret := make([]*resolver.MessageDependencyGraphNode, 0, len(nodes))
	for _, node := range nodes {
		n, err := d.toMessageDependencyGraphNode(node)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func (d *decoder) toMessageDependencyGraphNode(node *plugin.MessageDependencyGraphNode) (*resolver.MessageDependencyGraphNode, error) {
	if node == nil {
		return nil, nil
	}
	ret := &resolver.MessageDependencyGraphNode{
		ParentMap:   make(map[*resolver.MessageDependencyGraphNode]struct{}),
		ChildrenMap: make(map[*resolver.MessageDependencyGraphNode]struct{}),
	}
	children, err := d.toMessageDependencyGraphNodes(node.GetChildren())
	if err != nil {
		return nil, err
	}
	baseMsg, err := d.toMessage(node.GetBaseMessageId())
	if err != nil {
		return nil, err
	}
	def, err := d.toVariableDefinition(node.GetVariableDefinitionId())
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		child.Parent = append(child.Parent, ret)
		child.ParentMap[ret] = struct{}{}
		ret.ChildrenMap[child] = struct{}{}
	}
	ret.Children = children
	ret.BaseMessage = baseMsg
	ret.VariableDefinition = def
	return ret, nil
}

func (d *decoder) toVariableDefinitions(ids []string) ([]*resolver.VariableDefinition, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.VariableDefinition, 0, len(ids))
	for _, id := range ids {
		def, err := d.toVariableDefinition(id)
		if err != nil {
			return nil, err
		}
		if def == nil {
			continue
		}
		ret = append(ret, def)
	}
	return ret, nil
}

func (d *decoder) toVariableDefinition(id string) (*resolver.VariableDefinition, error) {
	if id == "" {
		return nil, nil
	}
	if def, exists := d.varDefMap[id]; exists {
		return def, nil
	}
	def, exists := d.ref.VariableDefinitionMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find variable definition reference: %s", id)
	}
	ret := &resolver.VariableDefinition{
		Idx:      int(def.GetIndex()),
		Name:     def.GetName(),
		AutoBind: def.GetAutoBind(),
		Used:     def.GetUsed(),
	}
	d.varDefMap[id] = ret

	ifValue, err := d.toCELValue(def.GetIf())
	if err != nil {
		return nil, err
	}

	expr, err := d.toVariableExpr(def.GetExpr())
	if err != nil {
		return nil, err
	}

	ret.If = ifValue
	ret.Expr = expr
	return ret, nil
}

func (d *decoder) toVariableDefinitionGroups(ids []string) ([]resolver.VariableDefinitionGroup, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]resolver.VariableDefinitionGroup, 0, len(ids))
	for _, id := range ids {
		group, err := d.toVariableDefinitionGroup(id)
		if err != nil {
			return nil, err
		}
		if group == nil {
			continue
		}
		ret = append(ret, group)
	}
	return ret, nil
}

func (d *decoder) toVariableDefinitionGroup(id string) (resolver.VariableDefinitionGroup, error) {
	if id == "" {
		return nil, nil
	}
	if group, exists := d.varDefGroupMap[id]; exists {
		return group, nil
	}
	group, exists := d.ref.VariableDefinitionGroupMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find variable definition group reference: %s", id)
	}
	switch {
	case group.GetSequential() != nil:
		ret := &resolver.SequentialVariableDefinitionGroup{}
		d.varDefGroupMap[id] = ret

		seq := group.GetSequential()
		start, err := d.toVariableDefinitionGroup(seq.GetStart())
		if err != nil {
			return nil, err
		}
		end, err := d.toVariableDefinition(seq.GetEnd())
		if err != nil {
			return nil, err
		}
		ret.Start = start
		ret.End = end
		return ret, nil

	case group.GetConcurrent() != nil:
		ret := &resolver.ConcurrentVariableDefinitionGroup{}
		d.varDefGroupMap[id] = ret

		conc := group.GetConcurrent()
		starts, err := d.toVariableDefinitionGroups(conc.GetStarts())
		if err != nil {
			return nil, err
		}
		end, err := d.toVariableDefinition(conc.GetEnd())
		if err != nil {
			return nil, err
		}
		ret.Starts = starts
		ret.End = end

		return ret, nil
	}
	return nil, fmt.Errorf("unexpected variable definition group type")
}

func (d *decoder) toVariableExpr(expr *plugin.VariableExpr) (*resolver.VariableExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.VariableExpr{}
	typ, err := d.toType(expr.GetType())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValue(expr.GetBy())
	if err != nil {
		return nil, err
	}
	mapExpr, err := d.toMapExpr(expr.GetMap())
	if err != nil {
		return nil, err
	}
	callExpr, err := d.toCallExpr(expr.GetCall())
	if err != nil {
		return nil, err
	}
	msgExpr, err := d.toMessageExpr(expr.GetMessage())
	if err != nil {
		return nil, err
	}
	enumExpr, err := d.toEnumExpr(expr.GetEnum())
	if err != nil {
		return nil, err
	}
	validationExpr, err := d.toValidationExpr(expr.GetValidation())
	if err != nil {
		return nil, err
	}

	ret.Type = typ
	ret.By = by
	ret.Map = mapExpr
	ret.Call = callExpr
	ret.Message = msgExpr
	ret.Enum = enumExpr
	ret.Validation = validationExpr
	return ret, nil
}

func (d *decoder) toMapExpr(expr *plugin.MapExpr) (*resolver.MapExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.MapExpr{}

	iter, err := d.toIterator(expr.GetIterator())
	if err != nil {
		return nil, err
	}
	iterExpr, err := d.toMapIteratorExpr(expr.GetExpr())
	if err != nil {
		return nil, err
	}
	ret.Iterator = iter
	ret.Expr = iterExpr
	return ret, nil
}

func (d *decoder) toIterator(iter *plugin.Iterator) (*resolver.Iterator, error) {
	if iter == nil {
		return nil, nil
	}
	ret := &resolver.Iterator{Name: iter.GetName()}

	src, err := d.toVariableDefinition(iter.GetSourceId())
	if err != nil {
		return nil, err
	}

	ret.Source = src
	return ret, nil
}

func (d *decoder) toMapIteratorExpr(expr *plugin.MapIteratorExpr) (*resolver.MapIteratorExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.MapIteratorExpr{}

	typ, err := d.toType(expr.GetType())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValue(expr.GetBy())
	if err != nil {
		return nil, err
	}
	msg, err := d.toMessageExpr(expr.GetMessage())
	if err != nil {
		return nil, err
	}
	enum, err := d.toEnumExpr(expr.GetEnum())
	if err != nil {
		return nil, err
	}

	ret.Type = typ
	ret.By = by
	ret.Message = msg
	ret.Enum = enum
	return ret, nil
}

func (d *decoder) toCallExpr(expr *plugin.CallExpr) (*resolver.CallExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.CallExpr{}

	mtd, err := d.toMethod(expr.GetMethodId())
	if err != nil {
		return nil, err
	}
	req, err := d.toRequest(expr.GetRequest())
	if err != nil {
		return nil, err
	}
	retry, err := d.toRetryPolicy(expr.GetRetry())
	if err != nil {
		return nil, err
	}
	if expr.Timeout != nil {
		timeout := expr.GetTimeout().AsDuration()
		ret.Timeout = &timeout
	}
	ret.Method = mtd
	ret.Request = req
	ret.Retry = retry
	return ret, nil
}

func (d *decoder) toRequest(req *plugin.Request) (*resolver.Request, error) {
	if req == nil {
		return nil, nil
	}
	ret := &resolver.Request{}

	args, err := d.toArgs(req.GetArgs())
	if err != nil {
		return nil, err
	}
	typ, err := d.toMessage(req.GetTypeId())
	if err != nil {
		return nil, err
	}
	ret.Args = args
	ret.Type = typ
	return ret, nil
}

func (d *decoder) toArgs(args []*plugin.Argument) ([]*resolver.Argument, error) {
	if args == nil {
		return nil, nil
	}
	ret := make([]*resolver.Argument, 0, len(args))
	for _, arg := range args {
		a, err := d.toArg(arg)
		if err != nil {
			return nil, err
		}
		if a == nil {
			continue
		}
		ret = append(ret, a)
	}
	return ret, nil
}

func (d *decoder) toArg(arg *plugin.Argument) (*resolver.Argument, error) {
	if arg == nil {
		return nil, nil
	}
	ret := &resolver.Argument{
		Name: arg.GetName(),
	}

	typ, err := d.toType(arg.GetType())
	if err != nil {
		return nil, err
	}
	value, err := d.toValue(arg.GetValue())
	if err != nil {
		return nil, err
	}
	ifValue, err := d.toCELValue(arg.GetIf())
	if err != nil {
		return nil, err
	}

	ret.Type = typ
	ret.Value = value
	ret.If = ifValue
	return ret, nil
}

func (d *decoder) toRetryPolicy(retry *plugin.RetryPolicy) (*resolver.RetryPolicy, error) {
	if retry == nil {
		return nil, nil
	}
	ret := &resolver.RetryPolicy{}
	ifValue, err := d.toCELValue(retry.GetIf())
	if err != nil {
		return nil, err
	}
	ret.If = ifValue
	switch {
	case retry.GetConstant() != nil:
		cons := retry.GetConstant()
		interval := cons.GetInterval().AsDuration()
		ret.Constant = &resolver.RetryPolicyConstant{
			Interval:   interval,
			MaxRetries: cons.GetMaxRetries(),
		}
		return ret, nil
	case retry.GetExponential() != nil:
		exp := retry.GetExponential()
		initialInterval := exp.GetInitialInterval().AsDuration()
		maxInterval := exp.GetMaxInterval().AsDuration()
		maxElapsedTime := exp.GetMaxElapsedTime().AsDuration()
		ret.Exponential = &resolver.RetryPolicyExponential{
			InitialInterval:     initialInterval,
			RandomizationFactor: exp.GetRandomizationFactor(),
			Multiplier:          exp.GetMultiplier(),
			MaxInterval:         maxInterval,
			MaxRetries:          exp.GetMaxRetries(),
			MaxElapsedTime:      maxElapsedTime,
		}
		return ret, nil
	}
	return nil, fmt.Errorf("unexpected retry policy")
}

func (d *decoder) toMessageExpr(expr *plugin.MessageExpr) (*resolver.MessageExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.MessageExpr{}

	msg, err := d.toMessage(expr.GetMessageId())
	if err != nil {
		return nil, err
	}
	args, err := d.toArgs(expr.GetArgs())
	if err != nil {
		return nil, err
	}

	ret.Message = msg
	ret.Args = args
	return ret, nil
}

func (d *decoder) toEnumExpr(expr *plugin.EnumExpr) (*resolver.EnumExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.EnumExpr{}

	enum, err := d.toEnum(expr.GetEnumId())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValue(expr.GetBy())
	if err != nil {
		return nil, err
	}
	ret.Enum = enum
	ret.By = by
	return ret, nil
}

func (d *decoder) toValidationExpr(expr *plugin.ValidationExpr) (*resolver.ValidationExpr, error) {
	if expr == nil {
		return nil, nil
	}
	ret := &resolver.ValidationExpr{}

	grpcErr, err := d.toGRPCError(expr.GetError())
	if err != nil {
		return nil, err
	}
	ret.Error = grpcErr
	return ret, nil
}

func (d *decoder) toGRPCError(e *plugin.GRPCError) (*resolver.GRPCError, error) {
	if e == nil {
		return nil, nil
	}
	ret := &resolver.GRPCError{
		Code:   e.Code,
		Ignore: e.GetIgnore(),
	}

	ifValue, err := d.toCELValue(e.GetIf())
	if err != nil {
		return nil, err
	}
	msgValue, err := d.toCELValue(e.GetMessage())
	if err != nil {
		return nil, err
	}
	details, err := d.toGRPCErrorDetails(e.GetDetails())
	if err != nil {
		return nil, err
	}
	ignoreAndResponse, err := d.toCELValue(e.GetIgnoreAndResponse())
	if err != nil {
		return nil, err
	}
	ret.If = ifValue
	ret.Message = msgValue
	ret.Details = details
	ret.IgnoreAndResponse = ignoreAndResponse
	return ret, nil
}

func (d *decoder) toGRPCErrorDetails(details []*plugin.GRPCErrorDetail) ([]*resolver.GRPCErrorDetail, error) {
	if details == nil {
		return nil, nil
	}
	ret := make([]*resolver.GRPCErrorDetail, 0, len(details))
	for _, detail := range details {
		v, err := d.toGRPCErrorDetail(detail)
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (d *decoder) toGRPCErrorDetail(detail *plugin.GRPCErrorDetail) (*resolver.GRPCErrorDetail, error) {
	ret := &resolver.GRPCErrorDetail{}

	defSet, err := d.toVariableDefinitionSet(detail.GetDefSet())
	if err != nil {
		return nil, err
	}
	ifValue, err := d.toCELValue(detail.GetIf())
	if err != nil {
		return nil, err
	}
	by, err := d.toCELValues(detail.GetBy())
	if err != nil {
		return nil, err
	}
	msgs, err := d.toVariableDefinitionSet(detail.GetMessages())
	if err != nil {
		return nil, err
	}
	preconditionFailures, err := d.toPreconditionFailures(detail.GetPreconditionFailures())
	if err != nil {
		return nil, err
	}
	badRequests, err := d.toBadRequests(detail.GetBadRequests())
	if err != nil {
		return nil, err
	}
	localizedMsgs, err := d.toLocalizedMessages(detail.GetLocalizedMessages())
	if err != nil {
		return nil, err
	}
	ret.DefSet = defSet
	ret.If = ifValue
	ret.By = by
	ret.Messages = msgs
	ret.PreconditionFailures = preconditionFailures
	ret.BadRequests = badRequests
	ret.LocalizedMessages = localizedMsgs
	return ret, nil
}

func (d *decoder) toPreconditionFailures(v []*plugin.PreconditionFailure) ([]*resolver.PreconditionFailure, error) {
	if v == nil {
		return nil, nil
	}
	ret := make([]*resolver.PreconditionFailure, 0, len(v))
	for _, vv := range v {
		preconditionFailure, err := d.toPreconditionFailure(vv)
		if err != nil {
			return nil, err
		}
		if preconditionFailure == nil {
			continue
		}
		ret = append(ret, preconditionFailure)
	}
	return ret, nil
}

func (d *decoder) toPreconditionFailure(v *plugin.PreconditionFailure) (*resolver.PreconditionFailure, error) {
	if v == nil {
		return nil, nil
	}
	ret := &resolver.PreconditionFailure{}
	violations, err := d.toPreconditionFailureViolations(v.GetViolations())
	if err != nil {
		return nil, err
	}
	ret.Violations = violations
	return ret, nil
}

func (d *decoder) toPreconditionFailureViolations(v []*plugin.PreconditionFailureViolation) ([]*resolver.PreconditionFailureViolation, error) {
	if v == nil {
		return nil, nil
	}
	ret := make([]*resolver.PreconditionFailureViolation, 0, len(v))
	for _, vv := range v {
		violation, err := d.toPreconditionFailureViolation(vv)
		if err != nil {
			return nil, err
		}
		if violation == nil {
			continue
		}
		ret = append(ret, violation)
	}
	return ret, nil
}

func (d *decoder) toPreconditionFailureViolation(v *plugin.PreconditionFailureViolation) (*resolver.PreconditionFailureViolation, error) {
	if v == nil {
		return nil, nil
	}
	ret := &resolver.PreconditionFailureViolation{}

	typ, err := d.toCELValue(v.GetType())
	if err != nil {
		return nil, err
	}
	subject, err := d.toCELValue(v.GetSubject())
	if err != nil {
		return nil, err
	}
	desc, err := d.toCELValue(v.GetDescription())
	if err != nil {
		return nil, err
	}

	ret.Type = typ
	ret.Subject = subject
	ret.Description = desc
	return ret, nil
}

func (d *decoder) toBadRequests(v []*plugin.BadRequest) ([]*resolver.BadRequest, error) {
	if v == nil {
		return nil, nil
	}
	ret := make([]*resolver.BadRequest, 0, len(v))
	for _, vv := range v {
		req, err := d.toBadRequest(vv)
		if err != nil {
			return nil, err
		}
		if req == nil {
			continue
		}
		ret = append(ret, req)
	}
	return ret, nil
}

func (d *decoder) toBadRequest(req *plugin.BadRequest) (*resolver.BadRequest, error) {
	if req == nil {
		return nil, nil
	}
	ret := &resolver.BadRequest{}

	violations, err := d.toBadRequestFieldViolations(req.GetFieldViolations())
	if err != nil {
		return nil, err
	}

	ret.FieldViolations = violations
	return ret, nil
}

func (d *decoder) toBadRequestFieldViolations(v []*plugin.BadRequestFieldViolation) ([]*resolver.BadRequestFieldViolation, error) {
	if v == nil {
		return nil, nil
	}
	ret := make([]*resolver.BadRequestFieldViolation, 0, len(v))
	for _, vv := range v {
		violation, err := d.toBadRequestFieldViolation(vv)
		if err != nil {
			return nil, err
		}
		if violation == nil {
			continue
		}
		ret = append(ret, violation)
	}
	return ret, nil
}

func (d *decoder) toBadRequestFieldViolation(v *plugin.BadRequestFieldViolation) (*resolver.BadRequestFieldViolation, error) {
	if v == nil {
		return nil, nil
	}
	ret := &resolver.BadRequestFieldViolation{}

	field, err := d.toCELValue(v.GetField())
	if err != nil {
		return nil, err
	}
	desc, err := d.toCELValue(v.GetDescription())
	if err != nil {
		return nil, err
	}

	ret.Field = field
	ret.Description = desc
	return ret, nil
}

func (d *decoder) toLocalizedMessages(v []*plugin.LocalizedMessage) ([]*resolver.LocalizedMessage, error) {
	if v == nil {
		return nil, nil
	}
	ret := make([]*resolver.LocalizedMessage, 0, len(v))
	for _, vv := range v {
		msg, err := d.toLocalizedMessage(vv)
		if err != nil {
			return nil, err
		}
		if msg == nil {
			continue
		}
		ret = append(ret, msg)
	}
	return ret, nil
}

func (d *decoder) toLocalizedMessage(v *plugin.LocalizedMessage) (*resolver.LocalizedMessage, error) {
	if v == nil {
		return nil, nil
	}
	ret := &resolver.LocalizedMessage{Locale: v.GetLocale()}
	msg, err := d.toCELValue(v.GetMessage())
	if err != nil {
		return nil, err
	}
	ret.Message = msg
	return ret, nil
}

func (d *decoder) toMethods(ids []string) ([]*resolver.Method, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Method, 0, len(ids))
	for _, id := range ids {
		mtd, err := d.toMethod(id)
		if err != nil {
			return nil, err
		}
		if mtd == nil {
			continue
		}
		ret = append(ret, mtd)
	}
	return ret, nil
}

func (d *decoder) toMethod(id string) (*resolver.Method, error) {
	if id == "" {
		return nil, nil
	}
	if mtd, exists := d.mtdMap[id]; exists {
		return mtd, nil
	}
	mtd, exists := d.ref.MethodMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find method reference: %s", id)
	}
	ret := &resolver.Method{Name: mtd.GetName()}
	d.mtdMap[id] = ret

	request, err := d.toMessage(mtd.GetRequestId())
	if err != nil {
		return nil, err
	}
	response, err := d.toMessage(mtd.GetResponseId())
	if err != nil {
		return nil, err
	}
	svc, err := d.toService(mtd.GetServiceId())
	if err != nil {
		return nil, err
	}
	rule, err := d.toMethodRule(mtd.GetRule())
	if err != nil {
		return nil, err
	}
	ret.Request = request
	ret.Response = response
	ret.Service = svc
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toMethodRule(rule *plugin.MethodRule) (*resolver.MethodRule, error) {
	if rule == nil {
		return nil, nil
	}
	ret := &resolver.MethodRule{}
	if rule.Timeout != nil {
		timeout := rule.GetTimeout().AsDuration()
		ret.Timeout = &timeout
	}
	if rule.GetResponseId() != "" {
		response, err := d.toMessage(rule.GetResponseId())
		if err != nil {
			return nil, err
		}
		ret.Response = response
	}
	return ret, nil
}

func (d *decoder) toEnums(ids []string) ([]*resolver.Enum, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.Enum, 0, len(ids))
	for _, id := range ids {
		enum, err := d.toEnum(id)
		if err != nil {
			return nil, err
		}
		if enum == nil {
			continue
		}
		ret = append(ret, enum)
	}
	return ret, nil
}

func (d *decoder) toEnum(id string) (*resolver.Enum, error) {
	if id == "" {
		return nil, nil
	}
	if enum, exists := d.enumMap[id]; exists {
		return enum, nil
	}
	enum, exists := d.ref.EnumMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find enum reference: %s", id)
	}
	ret := &resolver.Enum{Name: enum.GetName()}
	d.enumMap[id] = ret

	values, err := d.toEnumValues(enum.GetValueIds())
	if err != nil {
		return nil, err
	}
	msg, err := d.toMessage(enum.GetMessageId())
	if err != nil {
		return nil, err
	}
	file, err := d.toFile(enum.GetFileId())
	if err != nil {
		return nil, err
	}
	rule, err := d.toEnumRule(enum.GetRule())
	if err != nil {
		return nil, err
	}
	ret.Values = values
	ret.Message = msg
	ret.File = file
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toEnumRule(rule *plugin.EnumRule) (*resolver.EnumRule, error) {
	if rule == nil {
		return nil, nil
	}
	aliases, err := d.toEnums(rule.GetAliasIds())
	if err != nil {
		return nil, err
	}
	return &resolver.EnumRule{
		Aliases: aliases,
	}, nil
}

func (d *decoder) toEnumValueAliases(aliases []*plugin.EnumValueAlias) ([]*resolver.EnumValueAlias, error) {
	ret := make([]*resolver.EnumValueAlias, 0, len(aliases))
	for _, alias := range aliases {
		v, err := d.toEnumValueAlias(alias)
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (d *decoder) toEnumValueAttributes(attrs []*plugin.EnumValueAttribute) []*resolver.EnumValueAttribute {
	ret := make([]*resolver.EnumValueAttribute, 0, len(attrs))
	for _, attr := range attrs {
		v := d.toEnumValueAttribute(attr)
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}

func (d *decoder) toEnumValues(ids []string) ([]*resolver.EnumValue, error) {
	ret := make([]*resolver.EnumValue, 0, len(ids))
	for _, id := range ids {
		ev, err := d.toEnumValue(id)
		if err != nil {
			return nil, err
		}
		if ev == nil {
			continue
		}
		ret = append(ret, ev)
	}
	return ret, nil
}

func (d *decoder) toEnumValueAlias(alias *plugin.EnumValueAlias) (*resolver.EnumValueAlias, error) {
	if alias == nil {
		return nil, nil
	}
	enumAlias, err := d.toEnum(alias.GetEnumAliasId())
	if err != nil {
		return nil, err
	}
	enumValues, err := d.toEnumValues(alias.GetAliasIds())
	if err != nil {
		return nil, err
	}
	return &resolver.EnumValueAlias{
		EnumAlias: enumAlias,
		Aliases:   enumValues,
	}, nil
}

func (d *decoder) toEnumValueAttribute(attr *plugin.EnumValueAttribute) *resolver.EnumValueAttribute {
	if attr == nil {
		return nil
	}
	return &resolver.EnumValueAttribute{
		Name:  attr.GetName(),
		Value: attr.GetValue(),
	}
}

func (d *decoder) toEnumValue(id string) (*resolver.EnumValue, error) {
	if id == "" {
		return nil, nil
	}
	if value, exists := d.enumValueMap[id]; exists {
		return value, nil
	}
	value, exists := d.ref.EnumValueMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find enum value reference: %s", id)
	}
	ret := &resolver.EnumValue{Value: value.GetValue()}
	d.enumValueMap[id] = ret

	enum, err := d.toEnum(value.GetEnumId())
	if err != nil {
		return nil, err
	}
	rule, err := d.toEnumValueRule(value.GetRule())
	if err != nil {
		return nil, err
	}
	ret.Enum = enum
	ret.Rule = rule
	return ret, nil
}

func (d *decoder) toEnumValueRule(rule *plugin.EnumValueRule) (*resolver.EnumValueRule, error) {
	if rule == nil {
		return nil, nil
	}
	aliases, err := d.toEnumValueAliases(rule.GetAliases())
	if err != nil {
		return nil, err
	}
	attrs := d.toEnumValueAttributes(rule.GetAttrs())
	return &resolver.EnumValueRule{
		Default: rule.GetDefault(),
		Aliases: aliases,
		Attrs:   attrs,
	}, nil
}

func (d *decoder) toCELPlugins(ids []string) ([]*resolver.CELPlugin, error) {
	if ids == nil {
		return nil, nil
	}
	ret := make([]*resolver.CELPlugin, 0, len(ids))
	for _, id := range ids {
		p, err := d.toCELPlugin(id)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (d *decoder) toCELPlugin(id string) (*resolver.CELPlugin, error) {
	if id == "" {
		return nil, nil
	}
	if p, exists := d.celPluginMap[id]; exists {
		return p, nil
	}
	p, exists := d.ref.CelPluginMap[id]
	if !exists {
		return nil, fmt.Errorf("failed to find cel plugin reference: %s", id)
	}
	ret := &resolver.CELPlugin{
		Name: p.GetName(),
		Desc: p.GetDescription(),
	}
	d.celPluginMap[id] = ret

	funcs, err := d.toCELFunctions(p.GetFunctions())
	if err != nil {
		return nil, err
	}
	ret.Functions = funcs
	return ret, nil
}

func (d *decoder) toCELFunctions(funcs []*plugin.CELFunction) ([]*resolver.CELFunction, error) {
	ret := make([]*resolver.CELFunction, 0, len(funcs))
	for _, fn := range funcs {
		f, err := d.toCELFunction(fn)
		if err != nil {
			return nil, err
		}
		if f == nil {
			return nil, nil
		}
		ret = append(ret, f)
	}
	return ret, nil
}

func (d *decoder) toCELFunction(fn *plugin.CELFunction) (*resolver.CELFunction, error) {
	args, err := d.toTypes(fn.GetArgs())
	if err != nil {
		return nil, err
	}
	ret, err := d.toType(fn.GetReturn())
	if err != nil {
		return nil, err
	}
	receiver, err := d.toMessage(fn.GetReceiverId())
	if err != nil {
		return nil, err
	}
	return &resolver.CELFunction{
		Name:     fn.GetName(),
		ID:       fn.GetId(),
		Args:     args,
		Return:   ret,
		Receiver: receiver,
	}, nil
}

func (d *decoder) toTypes(t []*plugin.Type) ([]*resolver.Type, error) {
	if t == nil {
		return nil, nil
	}
	ret := make([]*resolver.Type, 0, len(t))
	for _, tt := range t {
		typ, err := d.toType(tt)
		if err != nil {
			return nil, err
		}
		if typ == nil {
			continue
		}
		ret = append(ret, typ)
	}
	return ret, nil
}

func (d *decoder) toType(t *plugin.Type) (*resolver.Type, error) {
	if t == nil {
		return nil, nil
	}
	msg, err := d.toMessage(t.GetMessageId())
	if err != nil {
		return nil, err
	}
	enum, err := d.toEnum(t.GetEnumId())
	if err != nil {
		return nil, err
	}
	oneofField, err := d.toOneofField(t.GetOneofFieldId())
	if err != nil {
		return nil, err
	}
	return &resolver.Type{
		Kind:       d.toTypeKind(t.GetKind()),
		Repeated:   t.GetRepeated(),
		IsNull:     t.GetIsNull(),
		Message:    msg,
		Enum:       enum,
		OneofField: oneofField,
	}, nil
}

func (d *decoder) toTypeKind(kind plugin.TypeKind) types.Kind {
	switch kind {
	case plugin.TypeKind_DOUBLE_TYPE:
		return types.Double
	case plugin.TypeKind_FLOAT_TYPE:
		return types.Float
	case plugin.TypeKind_INT64_TYPE:
		return types.Int64
	case plugin.TypeKind_UINT64_TYPE:
		return types.Uint64
	case plugin.TypeKind_INT32_TYPE:
		return types.Int32
	case plugin.TypeKind_FIXED64_TYPE:
		return types.Fixed64
	case plugin.TypeKind_FIXED32_TYPE:
		return types.Fixed32
	case plugin.TypeKind_BOOL_TYPE:
		return types.Bool
	case plugin.TypeKind_STRING_TYPE:
		return types.String
	case plugin.TypeKind_GROUP_TYPE:
		return types.Group
	case plugin.TypeKind_MESSAGE_TYPE:
		return types.Message
	case plugin.TypeKind_BYTES_TYPE:
		return types.Bytes
	case plugin.TypeKind_UINT32_TYPE:
		return types.Uint32
	case plugin.TypeKind_ENUM_TYPE:
		return types.Enum
	case plugin.TypeKind_SFIXED32_TYPE:
		return types.Sfixed32
	case plugin.TypeKind_SFIXED64_TYPE:
		return types.Sfixed64
	case plugin.TypeKind_SINT32_TYPE:
		return types.Sint32
	case plugin.TypeKind_SINT64_TYPE:
		return types.Sint64
	}
	return types.Unknown
}

func (d *decoder) toOneofField(id string) (*resolver.OneofField, error) {
	field, err := d.toField(id)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, nil
	}
	return &resolver.OneofField{Field: field}, nil
}
