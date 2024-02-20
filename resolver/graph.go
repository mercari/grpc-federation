//go:build !tinygo.wasm

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
		childMap := make(map[*AllMessageDependencyGraphNode]struct{})
		node := msgToNode[msg]
		for idx, varDef := range msg.Rule.DefSet.Definitions() {
			expr := varDef.Expr
			if expr == nil {
				continue
			}
			switch {
			case expr.Map != nil && expr.Map.Expr != nil && expr.Map.Expr.Message != nil:
				msgExpr := expr.Map.Expr.Message
				depMsg := msgExpr.Message
				depNode, depNodeExists := msgToNode[depMsg]
				if _, exists := childMap[depNode]; exists {
					continue
				}
				if !depNodeExists {
					if depMsg == nil {
						fileName := msg.File.Name
						ctx.addError(
							ErrWithLocation(
								`undefined message specified "grpc.federation.message" option`,
								source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).
									WithMap().
									WithMessage().WithName().Location(),
							),
						)
						continue
					}
					fileName := msg.File.Name
					if !depMsg.HasRuleEveryFields() {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
								source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).
									WithMap().
									WithMessage().WithName().Location(),
							),
						)
					}
					depNode = &AllMessageDependencyGraphNode{Message: depMsg}
				}
				if node != nil {
					node.Children = append(node.Children, depNode)
					depNode.Parent = append(depNode.Parent, node)
				}
				childMap[depNode] = struct{}{}
			case expr.Message != nil:
				depMsg := expr.Message.Message
				depNode, depNodeExists := msgToNode[depMsg]
				if _, exists := childMap[depNode]; exists {
					continue
				}
				if !depNodeExists {
					if depMsg == nil {
						fileName := msg.File.Name
						ctx.addError(
							ErrWithLocation(
								`undefined message specified "grpc.federation.message" option`,
								source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).
									WithMessage().WithName().Location(),
							),
						)
						continue
					}
					fileName := msg.File.Name
					if !depMsg.HasRuleEveryFields() {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
								source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).WithMessage().WithName().Location(),
							),
						)
					}
					depNode = &AllMessageDependencyGraphNode{Message: depMsg}
				}
				if node != nil {
					node.Children = append(node.Children, depNode)
					depNode.Parent = append(depNode.Parent, node)
				}
				childMap[depNode] = struct{}{}
			case expr.Validation != nil && expr.Validation.Error != nil:
				for dIdx, detail := range expr.Validation.Error.Details {
					for mIdx, def := range detail.Messages.Definitions() {
						depMsg := def.Expr.Message.Message
						depNode, depNodeExists := msgToNode[depMsg]
						if _, exists := childMap[depNode]; exists {
							continue
						}
						if !depNodeExists {
							if depMsg == nil {
								fileName := msg.File.Name
								ctx.addError(
									ErrWithLocation(
										`undefined message specified "grpc.federation.message" option`,
										source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).
											WithValidation().
											WithDetail(dIdx).
											WithMessage(mIdx).Location(),
									),
								)
								continue
							}
							fileName := msg.File.Name
							if !depMsg.HasRuleEveryFields() {
								ctx.addError(
									ErrWithLocation(
										fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
										source.NewMsgVarDefOptionBuilder(fileName, msg.Name, idx).
											WithValidation().
											WithDetail(dIdx).
											WithMessage(mIdx).
											WithMessage().WithName().Location(),
									),
								)
							}
							depNode = &AllMessageDependencyGraphNode{Message: depMsg}
						}
						if node != nil {
							node.Children = append(node.Children, depNode)
							depNode.Parent = append(depNode.Parent, node)
						}
						childMap[depNode] = struct{}{}
					}
				}
			}
		}
		for _, field := range msg.Fields {
			if field.Rule == nil {
				continue
			}
			if field.Rule.Oneof == nil {
				continue
			}
			for idx, varDef := range field.Rule.Oneof.DefSet.Definitions() {
				for _, msgExpr := range varDef.MessageExprs() {
					depMsg := msgExpr.Message
					depNode, depNodeExists := msgToNode[depMsg]
					if _, exists := childMap[depNode]; exists {
						continue
					}
					if !depNodeExists {
						if depMsg == nil {
							fileName := msg.File.Name
							ctx.addError(
								ErrWithLocation(
									`undefined message specified in "grpc.federation.field.oneof" option`,
									source.NewLocationBuilder(fileName).
										WithMessage(msg.Name).
										WithField(field.Name).
										WithOption().
										WithOneOf().
										WithVariableDefinitions(idx).
										WithMessage().Location(),
								),
							)
							continue
						}
						fileName := msg.File.Name
						if !depMsg.HasRuleEveryFields() {
							ctx.addError(
								ErrWithLocation(
									fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
									source.NewLocationBuilder(fileName).
										WithMessage(msg.Name).
										WithField(field.Name).
										WithOption().
										WithOneOf().
										WithVariableDefinitions(idx).
										WithMessage().Location(),
								),
							)
						}
						depNode = &AllMessageDependencyGraphNode{Message: depMsg}
					}
					if node != nil {
						node.Children = append(node.Children, depNode)
						depNode.Parent = append(depNode.Parent, node)
					}
					childMap[depNode] = struct{}{}
				}
			}
		}
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

func (g *MessageDependencyGraph) VariableDefinitionGroups(ctx *context) []VariableDefinitionGroup {
	var groups []VariableDefinitionGroup
	for _, child := range g.uniqueChildren() {
		if group := g.createVariableDefinitionGroup(ctx, child); group != nil {
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

func (g *MessageDependencyGraph) createVariableDefinitionGroup(ctx *context, node *MessageDependencyGraphNode) VariableDefinitionGroup {
	if node == nil {
		return nil
	}
	if len(node.Parent) == 0 {
		return &SequentialVariableDefinitionGroup{End: node.VariableDefinition}
	}
	if len(node.Parent) == 1 {
		return &SequentialVariableDefinitionGroup{
			Start: g.createVariableDefinitionGroup(ctx, node.Parent[0]),
			End:   node.VariableDefinition,
		}
	}
	rg := new(ConcurrentVariableDefinitionGroup)
	sort.Slice(node.Parent, func(i, j int) bool {
		return node.Parent[i].FQDN() < node.Parent[j].FQDN()
	})
	for _, parent := range node.Parent {
		if group := g.createVariableDefinitionGroup(ctx, parent); group != nil {
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
	// Maps a variable name to the MessageDependencyGraphNode which returns the variable
	nameToNode := make(map[string]*MessageDependencyGraphNode)

	var rootDefNodes []*MessageDependencyGraphNode
	for _, varDef := range defSet.Definitions() {
		varDefNode := newMessageDependencyGraphNode(baseMsg, varDef)
		if varDef.Name != "" {
			nameToNode[varDef.Name] = varDefNode
		}

		refs := varDef.ReferenceNames()
		if len(refs) == 0 {
			rootDefNodes = append(rootDefNodes, varDefNode)
			continue
		}
		if varDefNode == nil {
			continue
		}
		var iterName string
		if varDef.Expr.Map != nil && varDef.Expr.Map.Iterator != nil {
			iterName = varDef.Expr.Map.Iterator.Name
		}
		for _, ref := range refs {
			if ref == iterName {
				continue
			}

			node, exists := nameToNode[ref]
			if !exists {
				// not found name in current scope.
				continue
			}
			if _, exists := node.ChildrenMap[varDefNode]; !exists {
				node.Children = append(node.Children, varDefNode)
				node.ChildrenMap[varDefNode] = struct{}{}
			}
			if _, exists := varDefNode.ParentMap[node]; !exists {
				varDefNode.Parent = append(varDefNode.Parent, node)
				varDefNode.ParentMap[node] = struct{}{}
			}
		}
	}

	var roots []*MessageDependencyGraphNode
	for _, node := range nameToNode {
		if len(node.Parent) == 0 {
			roots = append(roots, node)
		}
	}
	roots = append(roots, rootDefNodes...)
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].FQDN() < roots[j].FQDN()
	})
	if len(roots) == 0 {
		return
	}

	graph := &MessageDependencyGraph{Roots: roots}
	if err := validateMessageGraph(graph); err != nil {
		ctx.addError(err)
		return
	}
	defSet.Graph = graph
	defSet.Groups = graph.VariableDefinitionGroups(ctx)
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
		for idx, varDef := range msg.Rule.DefSet.Definitions() {
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
					source.NewMsgVarDefOptionBuilder(msg.File.Name, msg.Name, idx).WithMessage().Location(),
				)
			}
		}
		return ErrWithLocation(
			fmt.Sprintf(`found cyclic dependency for "%s.%s" message. dependency path: %s`, target.BaseMessage.PackageName(), target.BaseMessage.Name, dependencyPath),
			source.NewLocationBuilder(target.BaseMessage.File.Name).WithMessage(target.BaseMessage.Name).Location(),
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
			source.NewLocationBuilder(target.Message.File.Name).WithMessage(target.Message.Name).Location(),
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
