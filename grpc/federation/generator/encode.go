package generator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/mercari/grpc-federation/grpc/federation/generator/plugin"
	"github.com/mercari/grpc-federation/resolver"
	"github.com/mercari/grpc-federation/types"
)

type encoder struct {
	ref *plugin.Reference
}

func newEncoder() *encoder {
	return &encoder{
		ref: &plugin.Reference{
			FileMap:                    make(map[string]*plugin.File),
			EnumMap:                    make(map[string]*plugin.Enum),
			EnumValueMap:               make(map[string]*plugin.EnumValue),
			MessageMap:                 make(map[string]*plugin.Message),
			FieldMap:                   make(map[string]*plugin.Field),
			OneofMap:                   make(map[string]*plugin.Oneof),
			ServiceMap:                 make(map[string]*plugin.Service),
			MethodMap:                  make(map[string]*plugin.Method),
			CelPluginMap:               make(map[string]*plugin.CELPlugin),
			GraphMap:                   make(map[string]*plugin.MessageDependencyGraph),
			VariableDefinitionMap:      make(map[string]*plugin.VariableDefinition),
			VariableDefinitionGroupMap: make(map[string]*plugin.VariableDefinitionGroup),
		},
	}
}

type ActionType string

const (
	KeepAction   ActionType = "keep"
	CreateAction ActionType = "create"
	DeleteAction ActionType = "delete"
	UpdateAction ActionType = "update"
	ProtocAction ActionType = "protoc"
)

type CodeGeneratorRequestConfig struct {
	Type                ActionType
	ProtoPath           string
	OutDir              string
	Files               []*plugin.ProtoCodeGeneratorResponse_File
	GRPCFederationFiles []*resolver.File
}

func CreateCodeGeneratorRequest(cfg *CodeGeneratorRequestConfig) *plugin.CodeGeneratorRequest {
	return newEncoder().toCodeGeneratorRequest(cfg)
}

func (e *encoder) toCodeGeneratorRequest(cfg *CodeGeneratorRequestConfig) *plugin.CodeGeneratorRequest {
	ret := &plugin.CodeGeneratorRequest{
		ProtoPath: cfg.ProtoPath,
		OutDir:    cfg.OutDir,
		Reference: e.ref,
	}
	for _, file := range e.toFiles(cfg.GRPCFederationFiles) {
		ret.GrpcFederationFileIds = append(ret.GrpcFederationFileIds, file.GetId())
	}
	return ret
}

func (e *encoder) toFile(file *resolver.File) *plugin.File {
	if file == nil {
		return nil
	}
	id := e.toFileID(file)
	if file, exists := e.ref.FileMap[id]; exists {
		return file
	}
	ret := &plugin.File{
		Id:        id,
		Name:      file.Name,
		GoPackage: e.toGoPackage(file.GoPackage),
	}
	e.ref.FileMap[id] = ret

	ret.Package = e.toPackage(file.Package)
	for _, svc := range e.toServices(file.Services) {
		ret.ServiceIds = append(ret.ServiceIds, svc.GetId())
	}
	for _, msg := range e.toMessages(file.Messages) {
		ret.MessageIds = append(ret.MessageIds, msg.GetId())
	}
	for _, enum := range e.toEnums(file.Enums) {
		ret.EnumIds = append(ret.EnumIds, enum.GetId())
	}
	for _, p := range e.toCELPlugins(file.CELPlugins) {
		ret.CelPluginIds = append(ret.CelPluginIds, p.GetId())
	}
	for _, file := range e.toFiles(file.ImportFiles) {
		ret.ImportFileIds = append(ret.ImportFileIds, file.GetId())
	}
	return ret
}

func (e *encoder) toPackage(pkg *resolver.Package) *plugin.Package {
	ret := &plugin.Package{Name: pkg.Name}

	for _, file := range e.toFiles(pkg.Files) {
		ret.FileIds = append(ret.FileIds, file.GetId())
	}
	return ret
}

func (e *encoder) toGoPackage(pkg *resolver.GoPackage) *plugin.GoPackage {
	return &plugin.GoPackage{
		Name:       pkg.Name,
		ImportPath: pkg.ImportPath,
		AliasName:  pkg.AliasName,
	}
}

func (e *encoder) toFiles(files []*resolver.File) []*plugin.File {
	ret := make([]*plugin.File, 0, len(files))
	for _, file := range files {
		f := e.toFile(file)
		if f == nil {
			continue
		}
		ret = append(ret, f)
	}
	return ret
}

func (e *encoder) toServices(svcs []*resolver.Service) []*plugin.Service {
	ret := make([]*plugin.Service, 0, len(svcs))
	for _, svc := range svcs {
		s := e.toService(svc)
		if s == nil {
			continue
		}
		ret = append(ret, s)
	}
	return ret
}

func (e *encoder) toService(svc *resolver.Service) *plugin.Service {
	if svc == nil {
		return nil
	}
	id := e.toServiceID(svc)
	if svc, exists := e.ref.ServiceMap[id]; exists {
		return svc
	}
	ret := &plugin.Service{Id: id, Name: svc.Name}
	e.ref.ServiceMap[id] = ret

	for _, mtd := range e.toMethods(svc.Methods) {
		ret.MethodIds = append(ret.MethodIds, mtd.GetId())
	}
	for _, msg := range e.toMessages(svc.Messages) {
		ret.MessageIds = append(ret.MessageIds, msg.GetId())
	}
	for _, msg := range e.toMessages(svc.MessageArgs) {
		ret.MessageArgIds = append(ret.MessageArgIds, msg.GetId())
	}
	for _, p := range e.toCELPlugins(svc.CELPlugins) {
		ret.CelPluginIds = append(ret.CelPluginIds, p.GetId())
	}
	ret.FileId = e.toFile(svc.File).GetId()
	ret.Rule = e.toServiceRule(svc.Rule)
	return ret
}

func (e *encoder) toServiceRule(rule *resolver.ServiceRule) *plugin.ServiceRule {
	if rule == nil {
		return nil
	}
	return &plugin.ServiceRule{}
}

func (e *encoder) toMessages(msgs []*resolver.Message) []*plugin.Message {
	ret := make([]*plugin.Message, 0, len(msgs))
	for _, msg := range msgs {
		m := e.toMessage(msg)
		if m == nil {
			continue
		}
		ret = append(ret, m)
	}
	return ret
}

func (e *encoder) toMessage(msg *resolver.Message) *plugin.Message {
	if msg == nil {
		return nil
	}
	id := e.toMessageID(msg)
	if msg, exists := e.ref.MessageMap[id]; exists {
		return msg
	}
	ret := &plugin.Message{
		Id:         id,
		Name:       msg.Name,
		IsMapEntry: msg.IsMapEntry,
	}
	e.ref.MessageMap[id] = ret

	ret.FileId = e.toFile(msg.File).GetId()
	ret.ParentMessageId = e.toMessage(msg.ParentMessage).GetId()
	for _, msg := range e.toMessages(msg.NestedMessages) {
		ret.NestedMessageIds = append(ret.NestedMessageIds, msg.GetId())
	}
	for _, enum := range e.toEnums(msg.Enums) {
		ret.EnumIds = append(ret.EnumIds, enum.GetId())
	}
	for _, field := range e.toFields(msg.Fields) {
		ret.FieldIds = append(ret.FieldIds, field.GetId())
	}
	for _, oneof := range e.toOneofs(msg.Oneofs) {
		ret.OneofIds = append(ret.OneofIds, oneof.GetId())
	}
	ret.Rule = e.toMessageRule(msg.Rule)
	return ret
}

func (e *encoder) toMessageRule(rule *resolver.MessageRule) *plugin.MessageRule {
	if rule == nil {
		return nil
	}
	ret := &plugin.MessageRule{
		CustomResolver: rule.CustomResolver,
	}
	var aliasIDs []string
	for _, msg := range e.toMessages(rule.Aliases) {
		aliasIDs = append(aliasIDs, msg.GetId())
	}
	ret.MessageArgumentId = e.toMessage(rule.MessageArgument).GetId()
	ret.AliasIds = aliasIDs
	ret.DefSet = e.toVariableDefinitionSet(rule.DefSet)
	return ret
}

func (e *encoder) toVariableDefinitionSet(set *resolver.VariableDefinitionSet) *plugin.VariableDefinitionSet {
	ret := &plugin.VariableDefinitionSet{}
	ret.DependencyGraphId = e.toMessageDependencyGraph(set.DependencyGraph()).GetId()
	for _, def := range e.toVariableDefinitions(set.Definitions()) {
		ret.VariableDefinitionIds = append(ret.VariableDefinitionIds, def.GetId())
	}
	for _, group := range e.toVariableDefinitionGroups(set.DefinitionGroups()) {
		ret.VariableDefinitionGroupIds = append(ret.VariableDefinitionGroupIds, group.GetId())
	}
	return ret
}

func (e *encoder) toEnums(enums []*resolver.Enum) []*plugin.Enum {
	ret := make([]*plugin.Enum, 0, len(enums))
	for _, enum := range enums {
		ev := e.toEnum(enum)
		if ev == nil {
			continue
		}
		ret = append(ret, ev)
	}
	return ret
}

func (e *encoder) toEnum(enum *resolver.Enum) *plugin.Enum {
	if enum == nil {
		return nil
	}
	id := e.toEnumID(enum)
	if enum, exists := e.ref.EnumMap[id]; exists {
		return enum
	}
	ret := &plugin.Enum{Id: id, Name: enum.Name}
	e.ref.EnumMap[id] = ret

	for _, value := range e.toEnumValues(enum.Values) {
		ret.ValueIds = append(ret.ValueIds, value.GetId())
	}
	ret.MessageId = e.toMessage(enum.Message).GetId()
	ret.FileId = e.toFile(enum.File).GetId()
	ret.Rule = e.toEnumRule(enum.Rule)
	return ret
}

func (e *encoder) toEnumRule(rule *resolver.EnumRule) *plugin.EnumRule {
	if rule == nil {
		return nil
	}
	var aliasIDs []string
	for _, alias := range rule.Aliases {
		aliasIDs = append(aliasIDs, e.toEnumID(alias))
	}
	return &plugin.EnumRule{
		AliasIds: aliasIDs,
	}
}

func (e *encoder) toEnumValueAliases(aliases []*resolver.EnumValueAlias) []*plugin.EnumValueAlias {
	ret := make([]*plugin.EnumValueAlias, 0, len(aliases))
	for _, alias := range aliases {
		v := e.toEnumValueAlias(alias)
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}

func (e *encoder) toEnumValueAttributes(attrs []*resolver.EnumValueAttribute) []*plugin.EnumValueAttribute {
	ret := make([]*plugin.EnumValueAttribute, 0, len(attrs))
	for _, attr := range attrs {
		v := e.toEnumValueAttribute(attr)
		if v == nil {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}

func (e *encoder) toEnumValues(values []*resolver.EnumValue) []*plugin.EnumValue {
	ret := make([]*plugin.EnumValue, 0, len(values))
	for _, value := range values {
		enumValue := e.toEnumValue(value)
		if enumValue == nil {
			continue
		}
		ret = append(ret, enumValue)
	}
	return ret
}

func (e *encoder) toEnumValueAlias(alias *resolver.EnumValueAlias) *plugin.EnumValueAlias {
	if alias == nil {
		return nil
	}
	var valueIDs []string
	for _, valueAlias := range alias.Aliases {
		valueIDs = append(valueIDs, e.toEnumValueID(valueAlias))
	}
	return &plugin.EnumValueAlias{
		EnumAliasId: e.toEnumID(alias.EnumAlias),
		AliasIds:    valueIDs,
	}
}

func (e *encoder) toEnumValueAttribute(attr *resolver.EnumValueAttribute) *plugin.EnumValueAttribute {
	if attr == nil {
		return nil
	}
	return &plugin.EnumValueAttribute{
		Name:  attr.Name,
		Value: attr.Value,
	}
}

func (e *encoder) toEnumValue(value *resolver.EnumValue) *plugin.EnumValue {
	if value == nil {
		return nil
	}
	id := e.toEnumValueID(value)
	if value, exists := e.ref.EnumValueMap[id]; exists {
		return value
	}
	ret := &plugin.EnumValue{Id: id, Value: value.Value}
	e.ref.EnumValueMap[id] = ret

	ret.EnumId = e.toEnum(value.Enum).GetId()
	ret.Rule = e.toEnumValueRule(value.Rule)
	return ret
}

func (e *encoder) toEnumValueRule(rule *resolver.EnumValueRule) *plugin.EnumValueRule {
	if rule == nil {
		return nil
	}
	ret := &plugin.EnumValueRule{
		Default: rule.Default,
		Aliases: e.toEnumValueAliases(rule.Aliases),
		Attrs:   e.toEnumValueAttributes(rule.Attrs),
	}
	return ret
}

func (e *encoder) toFields(fields []*resolver.Field) []*plugin.Field {
	ret := make([]*plugin.Field, 0, len(fields))
	for _, field := range fields {
		f := e.toField(field)
		if f == nil {
			continue
		}
		ret = append(ret, f)
	}
	return ret
}

func (e *encoder) toField(field *resolver.Field) *plugin.Field {
	if field == nil {
		return nil
	}
	id := e.toFieldID(field)
	if field, exists := e.ref.FieldMap[id]; exists {
		return field
	}
	ret := &plugin.Field{Id: id, Name: field.Name}
	e.ref.FieldMap[id] = ret

	ret.Type = e.toType(field.Type)
	ret.OneofId = e.toOneof(field.Oneof).GetId()
	ret.Rule = e.toFieldRule(field.Rule)
	ret.MessageId = e.toMessage(field.Message).GetId()
	return ret
}

func (e *encoder) toFieldRule(rule *resolver.FieldRule) *plugin.FieldRule {
	if rule == nil {
		return nil
	}
	var aliasIds []string
	for _, field := range e.toFields(rule.Aliases) {
		aliasIds = append(aliasIds, field.GetId())
	}
	return &plugin.FieldRule{
		Value:                 e.toValue(rule.Value),
		CustomResolver:        rule.CustomResolver,
		MessageCustomResolver: rule.MessageCustomResolver,
		AliasIds:              aliasIds,
		AutoBindField:         e.toAutoBindField(rule.AutoBindField),
		OneofRule:             e.toFieldOneofRule(rule.Oneof),
	}
}

func (e *encoder) toAutoBindField(field *resolver.AutoBindField) *plugin.AutoBindField {
	if field == nil {
		return nil
	}
	return &plugin.AutoBindField{
		VariableDefinitionId: e.toVariableDefinition(field.VariableDefinition).GetId(),
		FieldId:              e.toField(field.Field).GetId(),
	}
}

func (e *encoder) toFieldOneofRule(rule *resolver.FieldOneofRule) *plugin.FieldOneofRule {
	if rule == nil {
		return nil
	}
	ret := &plugin.FieldOneofRule{Default: rule.Default}
	ret.If = e.toCELValue(rule.If)
	ret.By = e.toCELValue(rule.By)
	ret.DefSet = e.toVariableDefinitionSet(rule.DefSet)
	return ret
}

func (e *encoder) toOneofs(oneofs []*resolver.Oneof) []*plugin.Oneof {
	ret := make([]*plugin.Oneof, 0, len(oneofs))
	for _, oneof := range oneofs {
		o := e.toOneof(oneof)
		if o == nil {
			continue
		}
		ret = append(ret, o)
	}
	return ret
}

func (e *encoder) toOneof(oneof *resolver.Oneof) *plugin.Oneof {
	if oneof == nil {
		return nil
	}
	id := e.toOneofID(oneof)
	if oneof, exists := e.ref.OneofMap[id]; exists {
		return oneof
	}
	ret := &plugin.Oneof{Id: id, Name: oneof.Name}
	e.ref.OneofMap[id] = ret

	ret.MessageId = e.toMessage(oneof.Message).GetId()
	for _, field := range e.toFields(oneof.Fields) {
		ret.FieldIds = append(ret.FieldIds, field.GetId())
	}
	return ret
}

func (e *encoder) toMethods(mtds []*resolver.Method) []*plugin.Method {
	ret := make([]*plugin.Method, 0, len(mtds))
	for _, mtd := range mtds {
		m := e.toMethod(mtd)
		if m == nil {
			continue
		}
		ret = append(ret, m)
	}
	return ret
}

func (e *encoder) toMethod(mtd *resolver.Method) *plugin.Method {
	if mtd == nil {
		return nil
	}
	id := e.toMethodID(mtd)
	if mtd, exists := e.ref.MethodMap[id]; exists {
		return mtd
	}
	ret := &plugin.Method{Id: id, Name: mtd.Name}
	e.ref.MethodMap[id] = ret

	ret.RequestId = e.toMessage(mtd.Request).GetId()
	ret.ResponseId = e.toMessage(mtd.Response).GetId()
	ret.ServiceId = e.toService(mtd.Service).GetId()
	ret.Rule = e.toMethodRule(mtd.Rule)
	return ret
}

func (e *encoder) toMethodRule(rule *resolver.MethodRule) *plugin.MethodRule {
	if rule == nil {
		return nil
	}
	var timeout *durationpb.Duration
	if rule.Timeout != nil {
		timeout = durationpb.New(*rule.Timeout)
	}
	var response string
	if rule.Response != nil {
		response = e.toMessage(rule.Response).GetId()
	}
	return &plugin.MethodRule{
		Timeout:    timeout,
		ResponseId: response,
	}
}

func (e *encoder) toCELPlugins(plugins []*resolver.CELPlugin) []*plugin.CELPlugin {
	ret := make([]*plugin.CELPlugin, 0, len(plugins))
	for _, plugin := range plugins {
		p := e.toCELPlugin(plugin)
		if p == nil {
			continue
		}
		ret = append(ret, p)
	}
	return ret
}

func (e *encoder) toCELPlugin(p *resolver.CELPlugin) *plugin.CELPlugin {
	if p == nil {
		return nil
	}
	id := e.toCELPluginID(p)
	if plug, exists := e.ref.CelPluginMap[id]; exists {
		return plug
	}
	ret := &plugin.CELPlugin{
		Id:          id,
		Name:        p.Name,
		Description: p.Desc,
		Functions:   e.toCELFunctions(p.Functions),
	}
	e.ref.CelPluginMap[id] = ret
	return ret
}

func (e *encoder) toCELFunctions(fns []*resolver.CELFunction) []*plugin.CELFunction {
	ret := make([]*plugin.CELFunction, 0, len(fns))
	for _, fn := range fns {
		f := e.toCELFunction(fn)
		if f == nil {
			continue
		}
		ret = append(ret, f)
	}
	return ret
}

func (e *encoder) toCELFunction(fn *resolver.CELFunction) *plugin.CELFunction {
	if fn == nil {
		return nil
	}
	return &plugin.CELFunction{
		Name:       fn.Name,
		Id:         fn.ID,
		Args:       e.toTypes(fn.Args),
		Return:     e.toType(fn.Return),
		ReceiverId: e.toMessage(fn.Receiver).GetId(),
	}
}

func (e *encoder) toTypes(t []*resolver.Type) []*plugin.Type {
	ret := make([]*plugin.Type, 0, len(t))
	for _, tt := range t {
		ret = append(ret, e.toType(tt))
	}
	return ret
}

func (e *encoder) toType(t *resolver.Type) *plugin.Type {
	if t == nil {
		return nil
	}
	ret := &plugin.Type{
		Kind:     e.toTypeKind(t.Kind),
		Repeated: t.Repeated,
		IsNull:   t.IsNull,
	}
	switch {
	case t.Message != nil:
		ret.Ref = &plugin.Type_MessageId{
			MessageId: e.toMessage(t.Message).GetId(),
		}
	case t.Enum != nil:
		ret.Ref = &plugin.Type_EnumId{
			EnumId: e.toEnum(t.Enum).GetId(),
		}
	case t.OneofField != nil:
		ret.Ref = &plugin.Type_OneofFieldId{
			OneofFieldId: e.toField(t.OneofField.Field).GetId(),
		}
	}
	return ret
}

func (e *encoder) toTypeKind(kind types.Kind) plugin.TypeKind {
	switch kind {
	case types.Double:
		return plugin.TypeKind_DOUBLE_TYPE
	case types.Float:
		return plugin.TypeKind_FLOAT_TYPE
	case types.Int64:
		return plugin.TypeKind_INT64_TYPE
	case types.Uint64:
		return plugin.TypeKind_UINT64_TYPE
	case types.Int32:
		return plugin.TypeKind_INT32_TYPE
	case types.Fixed64:
		return plugin.TypeKind_FIXED64_TYPE
	case types.Fixed32:
		return plugin.TypeKind_FIXED32_TYPE
	case types.Bool:
		return plugin.TypeKind_BOOL_TYPE
	case types.String:
		return plugin.TypeKind_STRING_TYPE
	case types.Group:
		return plugin.TypeKind_GROUP_TYPE
	case types.Message:
		return plugin.TypeKind_MESSAGE_TYPE
	case types.Bytes:
		return plugin.TypeKind_BYTES_TYPE
	case types.Uint32:
		return plugin.TypeKind_UINT32_TYPE
	case types.Enum:
		return plugin.TypeKind_ENUM_TYPE
	case types.Sfixed32:
		return plugin.TypeKind_SFIXED32_TYPE
	case types.Sfixed64:
		return plugin.TypeKind_SFIXED64_TYPE
	case types.Sint32:
		return plugin.TypeKind_SINT32_TYPE
	case types.Sint64:
		return plugin.TypeKind_SINT64_TYPE
	}
	return plugin.TypeKind_UNKNOWN_TYPE
}

func (e *encoder) toMessageDependencyGraph(graph *resolver.MessageDependencyGraph) *plugin.MessageDependencyGraph {
	if graph == nil {
		return nil
	}
	id := e.toDependencyGraphID(graph)
	if g, exists := e.ref.GraphMap[id]; exists {
		return g
	}
	ret := &plugin.MessageDependencyGraph{Id: id}
	e.ref.GraphMap[id] = ret

	ret.Roots = e.toMessageDependencyGraphNodes(graph.Roots)
	return ret
}

func (e *encoder) toMessageDependencyGraphNodes(nodes []*resolver.MessageDependencyGraphNode) []*plugin.MessageDependencyGraphNode {
	ret := make([]*plugin.MessageDependencyGraphNode, 0, len(nodes))
	for _, node := range nodes {
		ret = append(ret, e.toMessageDependencyGraphNode(node))
	}
	return ret
}

func (e *encoder) toMessageDependencyGraphNode(n *resolver.MessageDependencyGraphNode) *plugin.MessageDependencyGraphNode {
	if n == nil {
		return nil
	}
	return &plugin.MessageDependencyGraphNode{
		Children:             e.toMessageDependencyGraphNodes(n.Children),
		BaseMessageId:        e.toMessage(n.BaseMessage).GetId(),
		VariableDefinitionId: e.toVariableDefinition(n.VariableDefinition).GetId(),
	}
}

func (e *encoder) toVariableDefinitions(defs []*resolver.VariableDefinition) []*plugin.VariableDefinition {
	ret := make([]*plugin.VariableDefinition, 0, len(defs))
	for _, def := range defs {
		d := e.toVariableDefinition(def)
		if d == nil {
			continue
		}
		ret = append(ret, d)
	}
	return ret
}

func (e *encoder) toVariableDefinitionGroups(groups []resolver.VariableDefinitionGroup) []*plugin.VariableDefinitionGroup {
	ret := make([]*plugin.VariableDefinitionGroup, 0, len(groups))
	for _, group := range groups {
		g := e.toVariableDefinitionGroup(group)
		if g == nil {
			continue
		}
		ret = append(ret, g)
	}
	return ret
}

func (e *encoder) toVariableDefinition(def *resolver.VariableDefinition) *plugin.VariableDefinition {
	if def == nil {
		return nil
	}
	id := e.toVariableDefinitionID(def)
	if def, exists := e.ref.VariableDefinitionMap[id]; exists {
		return def
	}
	ret := &plugin.VariableDefinition{
		Id:       id,
		Index:    int64(def.Idx),
		Name:     def.Name,
		AutoBind: def.AutoBind,
		Used:     def.Used,
	}
	e.ref.VariableDefinitionMap[id] = ret

	ret.If = e.toCELValue(def.If)
	ret.Expr = e.toVariableExpr(def.Expr)
	return ret
}

func (e *encoder) toVariableDefinitionGroup(group resolver.VariableDefinitionGroup) *plugin.VariableDefinitionGroup {
	if group == nil {
		return nil
	}
	id := e.toVariableDefinitionGroupID(group)
	if g, exists := e.ref.VariableDefinitionGroupMap[id]; exists {
		return g
	}
	ret := &plugin.VariableDefinitionGroup{Id: id}
	e.ref.VariableDefinitionGroupMap[id] = ret

	switch group.Type() {
	case resolver.SequentialVariableDefinitionGroupType:
		ret.Group = &plugin.VariableDefinitionGroup_Sequential{
			Sequential: e.toSequentialVariableDefinitionGroup(group.(*resolver.SequentialVariableDefinitionGroup)),
		}
	case resolver.ConcurrentVariableDefinitionGroupType:
		ret.Group = &plugin.VariableDefinitionGroup_Concurrent{
			Concurrent: e.toConcurrentVariableDefinitionGroup(group.(*resolver.ConcurrentVariableDefinitionGroup)),
		}
	}
	return ret
}

func (e *encoder) toSequentialVariableDefinitionGroup(g *resolver.SequentialVariableDefinitionGroup) *plugin.SequentialVariableDefinitionGroup {
	return &plugin.SequentialVariableDefinitionGroup{
		Start: e.toVariableDefinitionGroup(g.Start).GetId(),
		End:   e.toVariableDefinition(g.End).GetId(),
	}
}

func (e *encoder) toConcurrentVariableDefinitionGroup(g *resolver.ConcurrentVariableDefinitionGroup) *plugin.ConcurrentVariableDefinitionGroup {
	ret := &plugin.ConcurrentVariableDefinitionGroup{}

	for _, group := range e.toVariableDefinitionGroups(g.Starts) {
		ret.Starts = append(ret.Starts, group.GetId())
	}
	ret.End = e.toVariableDefinition(g.End).GetId()
	return ret
}

func (e *encoder) toCELValues(v []*resolver.CELValue) []*plugin.CELValue {
	ret := make([]*plugin.CELValue, 0, len(v))
	for _, vv := range v {
		ret = append(ret, e.toCELValue(vv))
	}
	return ret
}

func (e *encoder) toCELValue(v *resolver.CELValue) *plugin.CELValue {
	if v == nil {
		return nil
	}
	return &plugin.CELValue{
		Expr: v.Expr,
		Out:  e.toType(v.Out),
	}
}

func (e *encoder) toVariableExpr(expr *resolver.VariableExpr) *plugin.VariableExpr {
	if expr == nil {
		return nil
	}
	ret := &plugin.VariableExpr{
		Type: e.toType(expr.Type),
	}
	switch {
	case expr.By != nil:
		ret.Expr = &plugin.VariableExpr_By{
			By: e.toCELValue(expr.By),
		}
	case expr.Map != nil:
		ret.Expr = &plugin.VariableExpr_Map{
			Map: e.toMapExpr(expr.Map),
		}
	case expr.Call != nil:
		ret.Expr = &plugin.VariableExpr_Call{
			Call: e.toCallExpr(expr.Call),
		}
	case expr.Message != nil:
		ret.Expr = &plugin.VariableExpr_Message{
			Message: e.toMessageExpr(expr.Message),
		}
	case expr.Enum != nil:
		ret.Expr = &plugin.VariableExpr_Enum{
			Enum: e.toEnumExpr(expr.Enum),
		}
	case expr.Validation != nil:
		ret.Expr = &plugin.VariableExpr_Validation{
			Validation: e.toValidationExpr(expr.Validation),
		}
	}
	return ret
}

func (e *encoder) toMapExpr(expr *resolver.MapExpr) *plugin.MapExpr {
	if expr == nil {
		return nil
	}
	return &plugin.MapExpr{
		Iterator: e.toIterator(expr.Iterator),
		Expr:     e.toMapIteratorExpr(expr.Expr),
	}
}

func (e *encoder) toIterator(iter *resolver.Iterator) *plugin.Iterator {
	if iter == nil {
		return nil
	}
	return &plugin.Iterator{
		Name:     iter.Name,
		SourceId: e.toVariableDefinitionID(iter.Source),
	}
}

func (e *encoder) toMapIteratorExpr(expr *resolver.MapIteratorExpr) *plugin.MapIteratorExpr {
	if expr == nil {
		return nil
	}
	ret := &plugin.MapIteratorExpr{
		Type: e.toType(expr.Type),
	}
	switch {
	case expr.By != nil:
		ret.Expr = &plugin.MapIteratorExpr_By{
			By: e.toCELValue(expr.By),
		}
	case expr.Message != nil:
		ret.Expr = &plugin.MapIteratorExpr_Message{
			Message: e.toMessageExpr(expr.Message),
		}
	case expr.Enum != nil:
		ret.Expr = &plugin.MapIteratorExpr_Enum{
			Enum: e.toEnumExpr(expr.Enum),
		}
	}
	return ret
}

func (e *encoder) toCallExpr(expr *resolver.CallExpr) *plugin.CallExpr {
	if expr == nil {
		return nil
	}
	var timeout *durationpb.Duration
	if expr.Timeout != nil {
		timeout = durationpb.New(*expr.Timeout)
	}
	return &plugin.CallExpr{
		MethodId: e.toMethod(expr.Method).GetId(),
		Request:  e.toRequest(expr.Request),
		Timeout:  timeout,
		Retry:    e.toRetryPolicy(expr.Retry),
	}
}

func (e *encoder) toRequest(req *resolver.Request) *plugin.Request {
	if req == nil {
		return nil
	}
	return &plugin.Request{
		Args:   e.toArgs(req.Args),
		TypeId: e.toMessage(req.Type).GetId(),
	}
}

func (e *encoder) toRetryPolicy(policy *resolver.RetryPolicy) *plugin.RetryPolicy {
	if policy == nil {
		return nil
	}
	ret := &plugin.RetryPolicy{
		If: e.toCELValue(policy.If),
	}
	switch {
	case policy.Constant != nil:
		ret.Policy = &plugin.RetryPolicy_Constant{
			Constant: e.toRetryConstant(policy.Constant),
		}
	case policy.Exponential != nil:
		ret.Policy = &plugin.RetryPolicy_Exponential{
			Exponential: e.toRetryExponential(policy.Exponential),
		}
	}
	return ret
}

func (e *encoder) toRetryConstant(cons *resolver.RetryPolicyConstant) *plugin.RetryPolicyConstant {
	if cons == nil {
		return nil
	}
	return &plugin.RetryPolicyConstant{
		Interval:   durationpb.New(cons.Interval),
		MaxRetries: cons.MaxRetries,
	}
}

func (e *encoder) toRetryExponential(exp *resolver.RetryPolicyExponential) *plugin.RetryPolicyExponential {
	if exp == nil {
		return nil
	}
	return &plugin.RetryPolicyExponential{
		InitialInterval:     durationpb.New(exp.InitialInterval),
		RandomizationFactor: exp.RandomizationFactor,
		Multiplier:          exp.Multiplier,
		MaxInterval:         durationpb.New(exp.MaxInterval),
		MaxRetries:          exp.MaxRetries,
		MaxElapsedTime:      durationpb.New(exp.MaxElapsedTime),
	}
}

func (e *encoder) toMessageExpr(expr *resolver.MessageExpr) *plugin.MessageExpr {
	if expr == nil {
		return nil
	}
	return &plugin.MessageExpr{
		MessageId: e.toMessage(expr.Message).GetId(),
		Args:      e.toArgs(expr.Args),
	}
}

func (e *encoder) toArgs(args []*resolver.Argument) []*plugin.Argument {
	ret := make([]*plugin.Argument, 0, len(args))
	for _, arg := range args {
		ret = append(ret, e.toArg(arg))
	}
	return ret
}

func (e *encoder) toArg(arg *resolver.Argument) *plugin.Argument {
	return &plugin.Argument{
		Name:  arg.Name,
		Type:  e.toType(arg.Type),
		Value: e.toValue(arg.Value),
		If:    e.toCELValue(arg.If),
	}
}

func (e *encoder) toValue(value *resolver.Value) *plugin.Value {
	if value == nil {
		return nil
	}
	return &plugin.Value{
		Inline: value.Inline,
		Cel:    e.toCELValue(value.CEL),
	}
}

func (e *encoder) toEnumExpr(expr *resolver.EnumExpr) *plugin.EnumExpr {
	if expr == nil {
		return nil
	}
	return &plugin.EnumExpr{
		EnumId: e.toEnum(expr.Enum).GetId(),
		By:     e.toCELValue(expr.By),
	}
}

func (e *encoder) toValidationExpr(expr *resolver.ValidationExpr) *plugin.ValidationExpr {
	if expr == nil {
		return nil
	}
	return &plugin.ValidationExpr{
		Error: e.toGRPCError(expr.Error),
	}
}

func (e *encoder) toGRPCError(err *resolver.GRPCError) *plugin.GRPCError {
	if err == nil {
		return nil
	}
	return &plugin.GRPCError{
		Code:              err.Code,
		If:                e.toCELValue(err.If),
		Message:           e.toCELValue(err.Message),
		Details:           e.toGRPCErrorDetails(err.Details),
		Ignore:            err.Ignore,
		IgnoreAndResponse: e.toCELValue(err.IgnoreAndResponse),
	}
}

func (e *encoder) toGRPCErrorDetails(details []*resolver.GRPCErrorDetail) []*plugin.GRPCErrorDetail {
	ret := make([]*plugin.GRPCErrorDetail, 0, len(details))
	for _, detail := range details {
		ret = append(ret, e.toGRPCErrorDetail(detail))
	}
	return ret
}

func (e *encoder) toGRPCErrorDetail(detail *resolver.GRPCErrorDetail) *plugin.GRPCErrorDetail {
	if detail == nil {
		return nil
	}
	ret := &plugin.GRPCErrorDetail{}
	ret.DefSet = e.toVariableDefinitionSet(detail.DefSet)
	ret.If = e.toCELValue(detail.If)
	ret.By = e.toCELValues(detail.By)
	ret.PreconditionFailures = e.toPreconditionFailures(detail.PreconditionFailures)
	ret.BadRequests = e.toBadRequests(detail.BadRequests)
	ret.LocalizedMessages = e.toLocalizedMessages(detail.LocalizedMessages)
	ret.Messages = e.toVariableDefinitionSet(detail.Messages)
	return ret
}

func (e *encoder) toPreconditionFailures(failures []*resolver.PreconditionFailure) []*plugin.PreconditionFailure {
	ret := make([]*plugin.PreconditionFailure, 0, len(failures))
	for _, failure := range failures {
		ret = append(ret, e.toPreconditionFailure(failure))
	}
	return ret
}

func (e *encoder) toPreconditionFailure(failure *resolver.PreconditionFailure) *plugin.PreconditionFailure {
	if failure == nil {
		return nil
	}
	return &plugin.PreconditionFailure{
		Violations: e.toPreconditionFailureViolations(failure.Violations),
	}
}

func (e *encoder) toPreconditionFailureViolations(v []*resolver.PreconditionFailureViolation) []*plugin.PreconditionFailureViolation {
	ret := make([]*plugin.PreconditionFailureViolation, 0, len(v))
	for _, vv := range v {
		ret = append(ret, e.toPreconditionFailureViolation(vv))
	}
	return ret
}

func (e *encoder) toPreconditionFailureViolation(v *resolver.PreconditionFailureViolation) *plugin.PreconditionFailureViolation {
	if v == nil {
		return nil
	}
	return &plugin.PreconditionFailureViolation{
		Type:        e.toCELValue(v.Type),
		Subject:     e.toCELValue(v.Subject),
		Description: e.toCELValue(v.Description),
	}
}

func (e *encoder) toBadRequests(reqs []*resolver.BadRequest) []*plugin.BadRequest {
	ret := make([]*plugin.BadRequest, 0, len(reqs))
	for _, req := range reqs {
		ret = append(ret, e.toBadRequest(req))
	}
	return ret
}

func (e *encoder) toBadRequest(req *resolver.BadRequest) *plugin.BadRequest {
	if req == nil {
		return nil
	}
	return &plugin.BadRequest{
		FieldViolations: e.toBadRequestFieldViolations(req.FieldViolations),
	}
}

func (e *encoder) toBadRequestFieldViolations(v []*resolver.BadRequestFieldViolation) []*plugin.BadRequestFieldViolation {
	ret := make([]*plugin.BadRequestFieldViolation, 0, len(v))
	for _, vv := range v {
		ret = append(ret, e.toBadRequestFieldViolation(vv))
	}
	return ret
}

func (e *encoder) toBadRequestFieldViolation(v *resolver.BadRequestFieldViolation) *plugin.BadRequestFieldViolation {
	if v == nil {
		return nil
	}
	return &plugin.BadRequestFieldViolation{
		Field:       e.toCELValue(v.Field),
		Description: e.toCELValue(v.Description),
	}
}

func (e *encoder) toLocalizedMessages(msgs []*resolver.LocalizedMessage) []*plugin.LocalizedMessage {
	ret := make([]*plugin.LocalizedMessage, 0, len(msgs))
	for _, msg := range msgs {
		ret = append(ret, e.toLocalizedMessage(msg))
	}
	return ret
}

func (e *encoder) toLocalizedMessage(msg *resolver.LocalizedMessage) *plugin.LocalizedMessage {
	if msg == nil {
		return nil
	}
	return &plugin.LocalizedMessage{
		Locale:  msg.Locale,
		Message: e.toCELValue(msg.Message),
	}
}

func (e *encoder) toFileID(file *resolver.File) string {
	if file == nil {
		return ""
	}
	return file.Package.Name + ":" + file.Name
}

func (e *encoder) toServiceID(svc *resolver.Service) string {
	if svc == nil {
		return ""
	}
	return svc.FQDN()
}

func (e *encoder) toMethodID(mtd *resolver.Method) string {
	if mtd == nil {
		return ""
	}
	return mtd.FQDN()
}

func (e *encoder) toMessageID(msg *resolver.Message) string {
	if msg == nil {
		return ""
	}
	return msg.FQDN()
}

func (e *encoder) toFieldID(field *resolver.Field) string {
	if field == nil {
		return ""
	}
	return field.FQDN()
}

func (e *encoder) toOneofID(oneof *resolver.Oneof) string {
	if oneof == nil {
		return ""
	}
	return fmt.Sprintf("%s.%s", oneof.Message.FQDN(), oneof.Name)
}

func (e *encoder) toEnumID(enum *resolver.Enum) string {
	if enum == nil {
		return ""
	}
	return enum.FQDN()
}

func (e *encoder) toEnumValueID(value *resolver.EnumValue) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%s.%s", value.Enum.FQDN(), value.Value)
}

func (e *encoder) toCELPluginID(celPlugin *resolver.CELPlugin) string {
	if celPlugin == nil {
		return ""
	}
	return celPlugin.Name
}

func (e *encoder) toDependencyGraphID(graph *resolver.MessageDependencyGraph) string {
	if graph == nil {
		return ""
	}
	roots := make([]string, 0, len(graph.Roots))
	for _, root := range graph.Roots {
		roots = append(roots, root.FQDN())
	}
	return strings.Join(roots, ":")
}

func (e *encoder) toVariableDefinitionID(def *resolver.VariableDefinition) string {
	if def == nil {
		return ""
	}
	return def.Name
}

func (e *encoder) toVariableDefinitionGroupID(group resolver.VariableDefinitionGroup) string {
	if group == nil {
		return ""
	}
	var defIDs []string
	for _, def := range group.VariableDefinitions() {
		defIDs = append(defIDs, e.toVariableDefinitionID(def))
	}
	return fmt.Sprintf("%s/%s", group.Type(), strings.Join(defIDs, ":"))
}
