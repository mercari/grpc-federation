package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/source"
)

type AllMessageDependencyGraph struct {
	Roots []*AllMessageDependencyGraphNode
}

type AllMessageDependencyGraphNode struct {
	Parent   []*AllMessageDependencyGraphNode
	Children []*AllMessageDependencyGraphNode
	Message  *Message
}

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
		for idx, varDef := range msg.Rule.VariableDefinitions {
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
								source.MapExprMessageNameLocation(fileName, msg.Name, idx),
							),
						)
						continue
					}
					fileName := msg.File.Name
					if !depMsg.HasRuleEveryFields() {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
								source.MapExprMessageNameLocation(fileName, msg.Name, idx),
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
								source.MessageExprNameLocation(fileName, msg.Name, idx),
							),
						)
						continue
					}
					fileName := msg.File.Name
					if !depMsg.HasRuleEveryFields() {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
								source.MessageExprNameLocation(fileName, msg.Name, idx),
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
					for mIdx, message := range detail.Messages {
						depMsg := message.Expr.Message.Message
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
										source.VariableDefinitionValidationDetailMessageLocation(fileName, msg.Name, idx, dIdx, mIdx),
									),
								)
								continue
							}
							fileName := msg.File.Name
							if !depMsg.HasRuleEveryFields() {
								ctx.addError(
									ErrWithLocation(
										fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
										source.VariableDefinitionValidationDetailMessageNameLocation(fileName, msg.Name, idx, dIdx, mIdx),
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
			for idx, varDef := range field.Rule.Oneof.VariableDefinitions {
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
									source.MessageFieldOneofDefMessageLocation(fileName, msg.Name, field.Name, idx),
								),
							)
							continue
						}
						fileName := msg.File.Name
						if !depMsg.HasRuleEveryFields() {
							ctx.addError(
								ErrWithLocation(
									fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
									source.MessageFieldOneofDefMessageLocation(fileName, msg.Name, field.Name, idx),
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

type MessageDependencyGraph struct {
	MessageRule    *MessageRule
	FieldOneofRule *FieldOneofRule
	Roots          []*MessageDependencyGraphNode
}

func (g *MessageDependencyGraph) MessageResolverGroups(ctx *context) []MessageResolverGroup {
	var groups []MessageResolverGroup
	for _, child := range g.uniqueChildren() {
		if group := g.createMessageResolverGroup(ctx, child); group != nil {
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

func (g *MessageDependencyGraph) createMessageResolverGroup(ctx *context, node *MessageDependencyGraphNode) MessageResolverGroup {
	if node == nil {
		return nil
	}
	if len(node.Parent) == 0 {
		return &SequentialMessageResolverGroup{End: g.createMessageResolver(ctx, node)}
	}
	if len(node.Parent) == 1 {
		return &SequentialMessageResolverGroup{
			Start: g.createMessageResolverGroup(ctx, node.Parent[0]),
			End:   g.createMessageResolver(ctx, node),
		}
	}
	rg := new(ConcurrentMessageResolverGroup)
	sort.Slice(node.Parent, func(i, j int) bool {
		return node.Parent[i].FQDN() < node.Parent[j].FQDN()
	})
	for _, parent := range node.Parent {
		if group := g.createMessageResolverGroup(ctx, parent); group != nil {
			rg.Starts = append(rg.Starts, group)
		}
	}
	rg.End = g.createMessageResolver(ctx, node)
	return rg
}

func (g *MessageDependencyGraph) createMessageResolver(ctx *context, node *MessageDependencyGraphNode) *MessageResolver {
	if g.MessageRule != nil {
		return g.createMessageResolverByNode(ctx, node)
	}
	if g.FieldOneofRule != nil {
		return g.createMessageResolverByFieldOneofRule(ctx, node, g.FieldOneofRule)
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message does not have resolver content`, node.BaseMessage.Name),
			source.MessageLocation(ctx.fileName(), node.BaseMessage.Name),
		),
	)
	return nil
}

func (g *MessageDependencyGraph) createMessageResolverByNode(ctx *context, node *MessageDependencyGraphNode) *MessageResolver {
	if varDef := node.VariableDefinition; varDef != nil {
		return &MessageResolver{Name: varDef.Name, VariableDefinition: varDef}
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message does not have resolver content`, node.BaseMessage.Name),
			source.MessageLocation(ctx.fileName(), node.BaseMessage.Name),
		),
	)
	return nil
}

func (g *MessageDependencyGraph) createMessageResolverByFieldOneofRule(ctx *context, node *MessageDependencyGraphNode, rule *FieldOneofRule) *MessageResolver {
	for _, def := range rule.VariableDefinitions {
		return &MessageResolver{Name: def.Name, VariableDefinition: def}
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message does not have resolver content`, node.BaseMessage.Name),
			source.MessageLocation(ctx.fileName(), node.BaseMessage.Name),
		),
	)
	return nil
}

type MessageDependencyGraphNode struct {
	Parent             []*MessageDependencyGraphNode
	Children           []*MessageDependencyGraphNode
	ParentMap          map[*MessageDependencyGraphNode]struct{}
	ChildrenMap        map[*MessageDependencyGraphNode]struct{}
	BaseMessage        *Message
	Message            *Message
	VariableDefinition *VariableDefinition
}

func (n *MessageDependencyGraphNode) FQDN() string {
	if n.VariableDefinition != nil {
		return fmt.Sprintf("%s_%s", n.BaseMessage.FQDN(), n.VariableDefinition.Name)
	}
	return n.Message.FQDN()
}

func newMessageDependencyGraphNodeByVariableDefinition(baseMsg *Message, def *VariableDefinition) *MessageDependencyGraphNode {
	return &MessageDependencyGraphNode{
		BaseMessage:        baseMsg,
		VariableDefinition: def,
		ParentMap:          make(map[*MessageDependencyGraphNode]struct{}),
		ChildrenMap:        make(map[*MessageDependencyGraphNode]struct{}),
	}
}

// CreateMessageDependencyGraph constructs a dependency graph from name references in the method calls,
// the arguments, and the validations.
// Name references in the arguments must be resolved first.
// If a circular reference occurs, add an error to context.
func CreateMessageDependencyGraph(ctx *context, baseMsg *Message) *MessageDependencyGraph {
	// Maps a variable name to the MessageDependencyGraphNode which returns the variable
	nameToNode := make(map[string]*MessageDependencyGraphNode)
	rule := baseMsg.Rule

	var rootDefNodes []*MessageDependencyGraphNode
	for _, varDef := range rule.VariableDefinitions {
		varDefNode := newMessageDependencyGraphNodeByVariableDefinition(baseMsg, varDef)
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
				// It can be either a reference registered directly from a file descriptor such as an enum value
				// or just an unknown reference name. cel-go detects the later case and returns an error, so we
				// can safely ignore it here
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
		return nil
	}

	graph := &MessageDependencyGraph{
		MessageRule: rule,
		Roots:       roots,
	}
	if err := validateMessageGraph(graph); err != nil {
		ctx.addError(err)
		return nil
	}
	return graph
}

func CreateMessageDependencyGraphByFieldOneof(ctx *context, baseMsg *Message, field *Field) *MessageDependencyGraph {
	nameToNode := make(map[string]*MessageDependencyGraphNode)

	fieldOneof := field.Rule.Oneof
	allRefMap := make(map[string]struct{})
	for _, varDef := range baseMsg.Rule.VariableDefinitions {
		for _, ref := range varDef.ReferenceNames() {
			allRefMap[ref] = struct{}{}
		}
	}

	var rootDefNodes []*MessageDependencyGraphNode
	for _, varDef := range fieldOneof.VariableDefinitions {
		varDefNode := newMessageDependencyGraphNodeByVariableDefinition(baseMsg, varDef)
		if varDef.Name != "" {
			nameToNode[varDef.Name] = varDefNode
		}

		refs := varDef.ReferenceNames()
		if len(refs) == 0 {
			rootDefNodes = append(rootDefNodes, varDefNode)
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
			if _, exists := allRefMap[ref]; !exists {
				// It can be either a reference registered directly from a file descriptor such as an enum value
				// or just an unknown reference name. cel-go detects the later case and returns an error, so we
				// can safely ignore it here
				continue
			}

			node, exists := nameToNode[ref]
			if !exists {
				// not oneof field reference name
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
		for _, ref := range refs {
			allRefMap[ref] = struct{}{}
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
		return nil
	}

	graph := &MessageDependencyGraph{
		FieldOneofRule: fieldOneof,
		Roots:          roots,
	}
	if err := validateMessageGraph(graph); err != nil {
		ctx.addError(err)
		return nil
	}
	return graph
}

func CreateMessageDependencyGraphByValidationErrorDetailMessages(baseMsg *Message, messages VariableDefinitions) *MessageDependencyGraph {
	// Each node won't depend on each other, so they all should be a root
	var roots []*MessageDependencyGraphNode
	for _, msg := range messages {
		node := newMessageDependencyGraphNodeByVariableDefinition(baseMsg, msg)
		roots = append(roots, node)
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].FQDN() < roots[j].FQDN()
	})
	if len(roots) == 0 {
		return nil
	}

	graph := &MessageDependencyGraph{
		MessageRule: &MessageRule{
			VariableDefinitions: messages,
		},
		Roots: roots,
	}
	return graph
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
			messages = append(messages, node.Message.Name)
		}
		dependencyPath := strings.Join(messages, " => ")

		msg := target.BaseMessage
		for idx, varDef := range msg.Rule.VariableDefinitions {
			if varDef.Expr == nil {
				continue
			}
			if varDef.Expr.Message == nil {
				continue
			}
			if varDef.Expr.Message.Message == target.Message {
				return ErrWithLocation(
					fmt.Sprintf(
						`found cyclic dependency for "%s.%s" message in "%s.%s. dependency path: %s"`,
						target.Message.PackageName(), target.Message.Name,
						msg.PackageName(), msg.Name, dependencyPath,
					),
					source.MessageExprLocation(msg.File.Name, msg.Name, idx),
				)
			}
		}
		return ErrWithLocation(
			fmt.Sprintf(`found cyclic dependency for "%s.%s" message. dependency path: %s`, target.Message.PackageName(), target.Message.Name, dependencyPath),
			source.MessageLocation(target.Message.File.Name, target.Message.Name),
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
			source.MessageLocation(target.Message.File.Name, target.Message.Name),
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
