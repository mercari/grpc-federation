package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/source"
)

// CreateAllMessageDependencyGraph creates a dependency graph for all messages with  message options defined.
func CreateAllMessageDependencyGraph(ctx *context, msgs []*Message) *AllMessageDependencyGraph {
	msgToNode := make(map[*Message]*AllMessageDependencyGraphNode)
	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}
		msgToNode[msg] = &AllMessageDependencyGraphNode{Message: msg}
	}

	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}

		newAllMessageDependencyGraphNodeReferenceBuilder(msgToNode, msg).Build(ctx)
	}
	var roots []*AllMessageDependencyGraphNode
	for _, node := range msgToNode {
		if len(node.Parent) == 0 {
			roots = append(roots, node)
		}
	}
	if len(roots) == 0 {
		return nil
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Message.Name < roots[j].Message.Name
	})
	graph := &AllMessageDependencyGraph{Roots: roots}
	if err := validateAllMessageGraph(graph); err != nil {
		ctx.addError(err)
		return nil
	}
	return graph
}

func (n *AllMessageDependencyGraphNode) childMessages() []*Message {
	messages := []*Message{n.Message}
	for _, child := range n.Children {
		if child.Message.Rule == nil {
			continue
		}
		messages = append(messages, child.childMessages()...)
	}
	return messages
}

type AllMessageDependencyGraphNodeReferenceBuilder struct {
	msgToNode map[*Message]*AllMessageDependencyGraphNode
	childMap  map[*AllMessageDependencyGraphNode]struct{}
	msg       *Message
	node      *AllMessageDependencyGraphNode
}

func newAllMessageDependencyGraphNodeReferenceBuilder(msgToNode map[*Message]*AllMessageDependencyGraphNode, msg *Message) *AllMessageDependencyGraphNodeReferenceBuilder {
	return &AllMessageDependencyGraphNodeReferenceBuilder{
		msgToNode: msgToNode,
		childMap:  make(map[*AllMessageDependencyGraphNode]struct{}),
		msg:       msg,
		node:      msgToNode[msg],
	}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) Build(ctx *context) {
	msg := b.msg
	fileName := msg.File.Name
	for _, varDef := range msg.Rule.DefSet.Definitions() {
		b.buildVariableDefinition(ctx, varDef, varDef.builder)
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for idx, varDef := range field.Rule.Oneof.DefSet.Definitions() {
			builder := source.NewLocationBuilder(fileName).
				WithMessage(b.msg.Name).
				WithField(field.Name).
				WithOption().
				WithOneOf().
				WithDef(idx)
			b.buildVariableDefinition(ctx, varDef, builder)
		}
	}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildVariableDefinition(ctx *context, def *VariableDefinition, builder *source.VariableDefinitionOptionBuilder) {
	expr := def.Expr
	if expr == nil {
		return
	}
	switch {
	case expr.Call != nil:
		b.buildCall(ctx, expr.Call, builder.WithCall())
	case expr.Message != nil:
		b.buildMessage(ctx, expr.Message, builder.WithMessage())
	case expr.Map != nil:
		b.buildMap(ctx, expr.Map, builder.WithMap())
	case expr.Validation != nil:
		b.buildValidation(ctx, expr.Validation, builder.WithValidation())
	}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildCall(ctx *context, expr *CallExpr, builder *source.CallExprOptionBuilder) {
	for idx, grpcErr := range expr.Errors {
		b.buildGRPCError(ctx, grpcErr, builder.WithError(idx))
	}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildMessage(ctx *context, expr *MessageExpr, builder *source.MessageExprOptionBuilder) {
	depMsg := expr.Message
	depNode, depNodeExists := b.msgToNode[depMsg]
	if _, exists := b.childMap[depNode]; exists {
		return
	}
	if !depNodeExists {
		if depMsg == nil {
			ctx.addError(
				ErrWithLocation(
					`undefined message specified`,
					builder.WithName().Location(),
				),
			)
			return
		}
		if !depMsg.HasRuleEveryFields() {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
					builder.WithName().Location(),
				),
			)
		}
		depNode = &AllMessageDependencyGraphNode{Message: depMsg}
	}
	if b.node != nil {
		b.node.Children = append(b.node.Children, depNode)
		depNode.Parent = append(depNode.Parent, b.node)
	}
	b.childMap[depNode] = struct{}{}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildMap(ctx *context, expr *MapExpr, builder *source.MapExprOptionBuilder) {
	if expr.Expr == nil {
		return
	}
	if expr.Expr.Message == nil {
		return
	}
	msgExpr := expr.Expr.Message
	depMsg := msgExpr.Message
	depNode, depNodeExists := b.msgToNode[depMsg]
	if _, exists := b.childMap[depNode]; exists {
		return
	}
	if !depNodeExists {
		if depMsg == nil {
			ctx.addError(
				ErrWithLocation(
					`undefined message`,
					builder.WithMessage().WithName().Location(),
				),
			)
			return
		}
		if !depMsg.HasRuleEveryFields() {
			ctx.addError(
				ErrWithLocation(
					fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
					builder.WithMessage().WithName().Location(),
				),
			)
		}
		depNode = &AllMessageDependencyGraphNode{Message: depMsg}
	}
	if b.node != nil {
		b.node.Children = append(b.node.Children, depNode)
		depNode.Parent = append(depNode.Parent, b.node)
	}
	b.childMap[depNode] = struct{}{}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildValidation(ctx *context, expr *ValidationExpr, builder *source.ValidationExprOptionBuilder) {
	b.buildGRPCError(ctx, expr.Error, builder.WithError())
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildGRPCError(ctx *context, grpcErr *GRPCError, builder *source.GRPCErrorOptionBuilder) {
	if grpcErr == nil {
		return
	}
	for idx, def := range grpcErr.DefSet.Definitions() {
		b.buildVariableDefinition(ctx, def, builder.WithDef(idx))
	}
	for idx, detail := range grpcErr.Details {
		b.buildGRPCErrorDetail(ctx, detail, builder.WithDetail(idx))
	}
}

func (b *AllMessageDependencyGraphNodeReferenceBuilder) buildGRPCErrorDetail(ctx *context, detail *GRPCErrorDetail, builder *source.GRPCErrorDetailOptionBuilder) {
	if detail == nil {
		return
	}
	for idx, def := range detail.DefSet.Definitions() {
		b.buildVariableDefinition(ctx, def, builder.WithDef(idx))
	}
	for idx, def := range detail.Messages.Definitions() {
		b.buildVariableDefinition(ctx, def, builder.WithDef(idx))
	}
}

func (g *MessageDependencyGraph) VariableDefinitionGroups() []VariableDefinitionGroup {
	var groups []VariableDefinitionGroup
	for _, child := range g.uniqueChildren() {
		if group := g.createVariableDefinitionGroup(child); group != nil {
			groups = append(groups, group)
		}
	}
	return groups
}

func (g *MessageDependencyGraph) uniqueChildren() []*MessageDependencyGraphNode {
	children := g.children(g.Roots)
	uniqueMap := make(map[*MessageDependencyGraphNode]struct{})
	for _, child := range children {
		uniqueMap[child] = struct{}{}
	}
	uniqueChildren := make([]*MessageDependencyGraphNode, 0, len(uniqueMap))
	for child := range uniqueMap {
		uniqueChildren = append(uniqueChildren, child)
	}
	sort.Slice(uniqueChildren, func(i, j int) bool {
		return uniqueChildren[i].FQDN() < uniqueChildren[j].FQDN()
	})
	return uniqueChildren
}

func (g *MessageDependencyGraph) children(nodes []*MessageDependencyGraphNode) []*MessageDependencyGraphNode {
	var children []*MessageDependencyGraphNode
	for _, node := range nodes {
		if len(node.Children) != 0 {
			children = append(children, g.children(node.Children)...)
		} else {
			children = append(children, node)
		}
	}
	return children
}

func (g *MessageDependencyGraph) createVariableDefinitionGroup(node *MessageDependencyGraphNode) VariableDefinitionGroup {
	if node == nil {
		return nil
	}
	if len(node.Parent) == 0 {
		return &SequentialVariableDefinitionGroup{End: node.VariableDefinition}
	}
	if len(node.Parent) == 1 {
		return &SequentialVariableDefinitionGroup{
			Start: g.createVariableDefinitionGroup(node.Parent[0]),
			End:   node.VariableDefinition,
		}
	}
	rg := new(ConcurrentVariableDefinitionGroup)
	sort.Slice(node.Parent, func(i, j int) bool {
		return node.Parent[i].FQDN() < node.Parent[j].FQDN()
	})
	for _, parent := range node.Parent {
		if group := g.createVariableDefinitionGroup(parent); group != nil {
			rg.Starts = append(rg.Starts, group)
		}
	}
	rg.End = node.VariableDefinition
	return rg
}

func newMessageDependencyGraphNode(baseMsg *Message, def *VariableDefinition) *MessageDependencyGraphNode {
	return &MessageDependencyGraphNode{
		BaseMessage:        baseMsg,
		VariableDefinition: def,
		ParentMap:          make(map[*MessageDependencyGraphNode]struct{}),
		ChildrenMap:        make(map[*MessageDependencyGraphNode]struct{}),
	}
}

// setupVariableDefinitionSet create a MessageDependencyGraph from VariableDefinitions of VariableDefinitionSet.
// Also, it creates VariableDefinitionGroups value and set it to VariableDefinitionSet.
// If a circular reference occurs, add an error to context.
func setupVariableDefinitionSet(ctx *context, baseMsg *Message, defSet *VariableDefinitionSet) {
	var nodes []*MessageDependencyGraphNode
	for _, varDef := range defSet.Definitions() {
		node := newMessageDependencyGraphNode(baseMsg, varDef)
		nodes = append(nodes, node)
	}

	setupVariableDependencyByReferenceName(nodes)
	setupVariableDependencyByValidation(nodes)

	var roots []*MessageDependencyGraphNode
	for _, node := range nodes {
		if len(node.Parent) == 0 {
			roots = append(roots, node)
		}
	}
	if len(roots) == 0 {
		return
	}

	graph := &MessageDependencyGraph{Roots: roots}
	if err := validateMessageGraph(graph); err != nil {
		ctx.addError(err)
		return
	}
	defSet.Graph = graph
	defSet.Groups = graph.VariableDefinitionGroups()
}

func setupVariableDependencyByReferenceName(nodes []*MessageDependencyGraphNode) {
	nameToNode := make(map[string]*MessageDependencyGraphNode)
	for _, node := range nodes {
		def := node.VariableDefinition
		if def.Name != "" {
			nameToNode[def.Name] = node
		}

		refs := def.ReferenceNames()
		if len(refs) == 0 {
			continue
		}
		var iterName string
		if def.Expr.Map != nil && def.Expr.Map.Iterator != nil {
			iterName = def.Expr.Map.Iterator.Name
		}
		for _, ref := range refs {
			if ref == iterName {
				continue
			}

			refNode, exists := nameToNode[ref]
			if !exists {
				// not found name in current scope.
				continue
			}
			if _, exists := refNode.ChildrenMap[node]; !exists {
				refNode.Children = append(refNode.Children, node)
				refNode.ChildrenMap[node] = struct{}{}
			}
			if _, exists := node.ParentMap[refNode]; !exists {
				node.Parent = append(node.Parent, refNode)
				node.ParentMap[refNode] = struct{}{}
			}
		}
	}
}

func setupVariableDependencyByValidation(nodes []*MessageDependencyGraphNode) {
	setupVariableDependencyByValidationRecursive(make(map[*MessageDependencyGraphNode]struct{}), nodes)
}

func setupVariableDependencyByValidationRecursive(parentMap map[*MessageDependencyGraphNode]struct{}, nodes []*MessageDependencyGraphNode) {
	if len(nodes) == 0 {
		return
	}
	lastIdx := 0
	for idx, node := range nodes {
		parentMap[node] = struct{}{}
		if !node.VariableDefinition.IsValidation() {
			continue
		}
		validationNode := node
		validationParents := make([]*MessageDependencyGraphNode, 0, len(validationNode.Parent))
		for parent := range validationNode.ParentMap {
			if parent.VariableDefinition.IsValidation() {
				validationParents = append(validationParents, parent)
			} else {
				// If a parent node exists in the parentMap, add it as a parent element of the validation node.
				// At this time, since all child elements of the validation node are executed after the evaluation of the validation node,
				// all parent nodes of the validation node have already been evaluated.
				// For this reason, the parent node is removed from the parentMap.
				// If it does not exist in the parentMap, it is determined to be an already evaluated parent element and is removed from validationNode.ParentMap.
				if _, exists := parentMap[parent]; exists {
					validationParents = append(validationParents, parent)
					delete(parentMap, parent)
				} else {
					delete(validationNode.ParentMap, parent)
				}
			}
		}
		validationNode.Parent = validationParents

		for i := idx + 1; i < len(nodes); i++ {
			curNode := nodes[i]

			// All nodes following the validationNode become child nodes of the validationNode.
			// This indicates a sequential relationship.
			if _, exists := validationNode.ChildrenMap[curNode]; !exists {
				validationNode.Children = append(validationNode.Children, curNode)
				validationNode.ChildrenMap[curNode] = struct{}{}
			}

			// If a validation node is already set as a parent node,
			// remove all of them so that only the current validation node is set as the parent.
			// This ensures that only the last validation node is retained as the parent.
			curParents := make([]*MessageDependencyGraphNode, 0, len(curNode.Parent))
			for parent := range curNode.ParentMap {
				if parent.VariableDefinition.IsValidation() {
					delete(curNode.ParentMap, parent)
				} else {
					curParents = append(curParents, parent)
				}
			}
			curNode.ParentMap[validationNode] = struct{}{}
			curParents = append(curParents, validationNode)
			curNode.Parent = curParents
		}
		lastIdx = idx
		break
	}
	// Recursively evaluate from the node following the last found validation node.
	setupVariableDependencyByValidationRecursive(parentMap, nodes[lastIdx+1:])
}

func validateMessageGraph(graph *MessageDependencyGraph) *LocationError {
	for _, root := range graph.Roots {
		if err := validateMessageNode(root); err != nil {
			return err
		}
	}
	return nil
}

func validateMessageNode(node *MessageDependencyGraphNode) *LocationError {
	if err := validateMessageNodeCyclicDependency(node, make(map[*MessageDependencyGraphNode]struct{}), []*MessageDependencyGraphNode{}); err != nil {
		return err
	}
	return nil
}

func validateMessageNodeCyclicDependency(target *MessageDependencyGraphNode, visited map[*MessageDependencyGraphNode]struct{}, path []*MessageDependencyGraphNode) *LocationError {
	path = append(path, target)
	if _, exists := visited[target]; exists {
		var messages []string
		for _, node := range path {
			messages = append(messages, node.BaseMessage.Name)
		}
		dependencyPath := strings.Join(messages, " => ")

		msg := target.BaseMessage
		for _, varDef := range msg.Rule.DefSet.Definitions() {
			if varDef.Expr == nil {
				continue
			}
			if varDef.Expr.Message == nil {
				continue
			}
			if varDef.Expr.Message.Message == target.BaseMessage {
				return ErrWithLocation(
					fmt.Sprintf(
						`found cyclic dependency for "%s.%s" message in "%s.%s. dependency path: %s"`,
						target.BaseMessage.PackageName(), target.BaseMessage.Name,
						msg.PackageName(), msg.Name, dependencyPath,
					),
					varDef.builder.WithMessage().Location(),
				)
			}
		}
		return ErrWithLocation(
			fmt.Sprintf(`found cyclic dependency for "%s.%s" message. dependency path: %s`, target.BaseMessage.PackageName(), target.BaseMessage.Name, dependencyPath),
			newMessageBuilderFromMessage(target.BaseMessage).Location(),
		)
	}
	visited[target] = struct{}{}
	for _, child := range target.Children {
		if err := validateMessageNodeCyclicDependency(child, visited, path); err != nil {
			return err
		}
	}
	delete(visited, target)
	return nil
}

func validateAllMessageGraph(graph *AllMessageDependencyGraph) *LocationError {
	for _, root := range graph.Roots {
		if err := validateAllMessageGraphNode(root); err != nil {
			return err
		}
	}
	return nil
}

func validateAllMessageGraphNode(node *AllMessageDependencyGraphNode) *LocationError {
	if err := validateAllMessageGraphNodeCyclicDependency(node, make(map[*AllMessageDependencyGraphNode]struct{}), []*AllMessageDependencyGraphNode{}); err != nil {
		return err
	}
	return nil
}

func validateAllMessageGraphNodeCyclicDependency(target *AllMessageDependencyGraphNode, visited map[*AllMessageDependencyGraphNode]struct{}, path []*AllMessageDependencyGraphNode) *LocationError {
	path = append(path, target)
	if _, exists := visited[target]; exists {
		var messages []string
		for _, node := range path {
			messages = append(messages, node.Message.Name)
		}
		dependencyPath := strings.Join(messages, " => ")

		return ErrWithLocation(
			fmt.Sprintf(`found cyclic dependency in "%s.%s" message. dependency path: %s`, target.Message.PackageName(), target.Message.Name, dependencyPath),
			newMessageBuilderFromMessage(target.Message).Location(),
		)
	}
	visited[target] = struct{}{}
	for _, child := range target.Children {
		if err := validateAllMessageGraphNodeCyclicDependency(child, visited, path); err != nil {
			return err
		}
	}
	delete(visited, target)
	return nil
}
