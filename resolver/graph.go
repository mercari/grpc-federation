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
	msgToNode := map[*Message]*AllMessageDependencyGraphNode{}
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
		for depIdx, depMessage := range msg.Rule.MessageDependencies {
			depNode, depNodeExists := msgToNode[depMessage.Message]
			if _, exists := childMap[depNode]; exists {
				continue
			}
			if !depNodeExists {
				if depMessage.Message == nil {
					fileName := msg.File.Name
					ctx.addError(
						ErrWithLocation(
							`undefined message specified "grpc.federation.message" option`,
							source.MessageDependencyMessageLocation(fileName, msg.Name, depIdx),
						),
					)
					continue
				}
				fileName := msg.File.Name
				depMsg := depMessage.Message
				if !depMsg.HasRuleEveryFields() {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
							source.MessageDependencyMessageLocation(fileName, msg.Name, depIdx),
						),
					)
				}
				depMessage.Message.Rule = &MessageRule{}
				depNode = &AllMessageDependencyGraphNode{Message: depMessage.Message}
			}
			node.Children = append(node.Children, depNode)
			depNode.Parent = append(depNode.Parent, node)
			childMap[depNode] = struct{}{}
		}
		for _, field := range msg.Fields {
			if field.Rule == nil {
				continue
			}
			if field.Rule.Oneof == nil {
				continue
			}
			for depIdx, depMessage := range field.Rule.Oneof.MessageDependencies {
				depNode, depNodeExists := msgToNode[depMessage.Message]
				if _, exists := childMap[depNode]; exists {
					continue
				}
				if !depNodeExists {
					if depMessage.Message == nil {
						fileName := msg.File.Name
						ctx.addError(
							ErrWithLocation(
								`undefined message specified in "grpc.federation.field.oneof" option`,
								source.MessageFieldOneofMessageDependencyMessageLocation(fileName, msg.Name, field.Name, depIdx),
							),
						)
						continue
					}
					fileName := msg.File.Name
					depMsg := depMessage.Message
					if !depMsg.HasRuleEveryFields() {
						ctx.addError(
							ErrWithLocation(
								fmt.Sprintf(`"%s.%s" message does not specify "grpc.federation.message" option`, depMsg.Package().Name, depMsg.Name),
								source.MessageFieldOneofMessageDependencyMessageLocation(fileName, msg.Name, field.Name, depIdx),
							),
						)
					}
					depMessage.Message.Rule = &MessageRule{}
					depNode = &AllMessageDependencyGraphNode{Message: depMessage.Message}
				}
				node.Children = append(node.Children, depNode)
				depNode.Parent = append(depNode.Parent, node)
				childMap[depNode] = struct{}{}
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
		return g.createMessageResolverByMessageRule(ctx, node, g.MessageRule)
	}
	if g.FieldOneofRule != nil {
		return g.createMessageResolverByFieldOneofRule(ctx, node, g.FieldOneofRule)
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message has not resolver content`, node.Message.Name),
			source.MessageLocation(ctx.fileName(), node.Message.Name),
		),
	)
	return nil
}

func (g *MessageDependencyGraph) createMessageResolverByMessageRule(ctx *context, node *MessageDependencyGraphNode, rule *MessageRule) *MessageResolver {
	if rule.MethodCall != nil {
		if rule.MethodCall.Response.Type == node.Message {
			var name string
			methodCall := rule.MethodCall
			if methodCall.Method != nil {
				name = methodCall.Method.Name
			}
			return &MessageResolver{Name: name, MethodCall: methodCall}
		}
	}
	for _, depMessage := range rule.MessageDependencies {
		if depMessage == node.MessageDependency {
			return &MessageResolver{Name: depMessage.Name, MessageDependency: depMessage}
		}
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message has not resolver content`, node.Message.Name),
			source.MessageLocation(ctx.fileName(), node.Message.Name),
		),
	)
	return nil
}

func (g *MessageDependencyGraph) createMessageResolverByFieldOneofRule(ctx *context, node *MessageDependencyGraphNode, rule *FieldOneofRule) *MessageResolver {
	for _, depMessage := range rule.MessageDependencies {
		if depMessage == node.MessageDependency {
			return &MessageResolver{Name: depMessage.Name, MessageDependency: depMessage}
		}
	}
	ctx.addError(
		ErrWithLocation(
			fmt.Sprintf(`%q message has not resolver content`, node.Message.Name),
			source.MessageLocation(ctx.fileName(), node.Message.Name),
		),
	)
	return nil
}

type MessageDependencyGraphNode struct {
	Parent            []*MessageDependencyGraphNode
	Children          []*MessageDependencyGraphNode
	ParentMap         map[*MessageDependencyGraphNode]struct{}
	ChildrenMap       map[*MessageDependencyGraphNode]struct{}
	BaseMessage       *Message
	Message           *Message
	Response          *Response
	MessageDependency *MessageDependency
}

func (n *MessageDependencyGraphNode) FQDN() string {
	if n.MessageDependency != nil {
		return fmt.Sprintf("%s_%s", n.Message.FQDN(), n.MessageDependency.Name)
	}
	return n.Message.FQDN()
}

func newMessageDependencyGraphNodeByResponse(baseMsg *Message, response *Response) *MessageDependencyGraphNode {
	return &MessageDependencyGraphNode{
		BaseMessage: baseMsg,
		Message:     response.Type,
		Response:    response,
		ParentMap:   make(map[*MessageDependencyGraphNode]struct{}),
		ChildrenMap: make(map[*MessageDependencyGraphNode]struct{}),
	}
}

func newMessageDependencyGraphNodeByMessageDependency(baseMsg *Message, dep *MessageDependency) *MessageDependencyGraphNode {
	return &MessageDependencyGraphNode{
		BaseMessage:       baseMsg,
		Message:           dep.Message,
		MessageDependency: dep,
		ParentMap:         make(map[*MessageDependencyGraphNode]struct{}),
		ChildrenMap:       make(map[*MessageDependencyGraphNode]struct{}),
	}
}

// CreateMessageRuleDependencyGraph construct a dependency graph using the name-based reference dependencies used in the method calls
// and the arguments used to retrieve the dependency messages.
// Requires reference resolution for arguments that use prior name-based references.
// If a circular reference occurs, add an error to context.
func CreateMessageDependencyGraph(ctx *context, baseMsg *Message) *MessageDependencyGraph {
	nameToNode := make(map[string]*MessageDependencyGraphNode)
	msgToNodes := make(map[*Message][]*MessageDependencyGraphNode)
	rule := baseMsg.Rule
	if rule.MethodCall != nil && rule.MethodCall.Response != nil {
		msg := rule.MethodCall.Response.Type
		node := newMessageDependencyGraphNodeByResponse(baseMsg, rule.MethodCall.Response)
		msgToNodes[msg] = append(msgToNodes[msg], node)
		for _, field := range rule.MethodCall.Response.Fields {
			nameToNode[field.Name] = node
		}
	}
	for _, depMessage := range rule.MessageDependencies {
		msg := depMessage.Message
		if msg == nil {
			continue
		}
		node := newMessageDependencyGraphNodeByMessageDependency(baseMsg, depMessage)
		msgToNodes[msg] = append(msgToNodes[msg], node)
		nameToNode[depMessage.Name] = node
	}

	// build dependencies from name reference.
	if rule.MethodCall != nil && rule.MethodCall.Request != nil && rule.MethodCall.Response != nil {
		for idx, arg := range rule.MethodCall.Request.Args {
			refs := arg.Value.ReferenceNames()
			if len(refs) == 0 {
				continue
			}
			for _, ref := range refs {
				node, exists := nameToNode[ref]
				if !exists {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`%q name does not exist`, ref),
							source.RequestByLocation(ctx.fileName(), baseMsg.Name, idx),
						),
					)
					continue
				}
				for _, responseTypeNode := range msgToNodes[rule.MethodCall.Response.Type] {
					if _, exists := node.ChildrenMap[responseTypeNode]; !exists {
						node.Children = append(node.Children, responseTypeNode)
						node.ChildrenMap[responseTypeNode] = struct{}{}
					}
					if _, exists := responseTypeNode.ParentMap[node]; !exists {
						responseTypeNode.Parent = append(responseTypeNode.Parent, node)
						responseTypeNode.ParentMap[node] = struct{}{}
					}
				}
			}
		}
	}
	for msgIdx, depMessage := range rule.MessageDependencies {
		for argIdx, arg := range depMessage.Args {
			refs := arg.Value.ReferenceNames()
			if len(refs) == 0 {
				continue
			}
			for _, ref := range refs {
				node, exists := nameToNode[ref]
				if !exists {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`%q name does not exist`, ref),
							source.MessageDependencyArgumentByLocation(ctx.fileName(), baseMsg.Name, msgIdx, argIdx),
						),
					)
					continue
				}
				for _, depMessageNode := range msgToNodes[depMessage.Message] {
					if _, exists := node.ChildrenMap[depMessageNode]; !exists {
						node.Children = append(node.Children, depMessageNode)
						node.ChildrenMap[depMessageNode] = struct{}{}
					}
					if _, exists := depMessageNode.ParentMap[node]; !exists {
						depMessageNode.Parent = append(depMessageNode.Parent, node)
						depMessageNode.ParentMap[node] = struct{}{}
					}
				}
			}
		}
	}

	var roots []*MessageDependencyGraphNode
	for _, nodes := range msgToNodes {
		for _, node := range nodes {
			if len(node.Parent) == 0 {
				roots = append(roots, node)
			}
		}
	}
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
	msgToNodes := make(map[*Message][]*MessageDependencyGraphNode)

	fieldOneof := field.Rule.Oneof
	allReferences := baseMsg.ReferenceNames()
	allRefMap := make(map[string]struct{})
	for _, ref := range allReferences {
		allRefMap[ref] = struct{}{}
	}
	for _, depMessage := range fieldOneof.MessageDependencies {
		msg := depMessage.Message
		if msg == nil {
			continue
		}
		node := newMessageDependencyGraphNodeByMessageDependency(baseMsg, depMessage)
		msgToNodes[msg] = append(msgToNodes[msg], node)
		nameToNode[depMessage.Name] = node
	}

	// build dependencies from name reference.
	for depIdx, depMessage := range fieldOneof.MessageDependencies {
		for argIdx, arg := range depMessage.Args {
			refs := arg.Value.ReferenceNames()
			if len(refs) == 0 {
				continue
			}
			for _, ref := range refs {
				if _, exists := allRefMap[ref]; !exists {
					ctx.addError(
						ErrWithLocation(
							fmt.Sprintf(`%q name does not exist`, ref),
							source.MessageFieldOneofMessageDependencyArgumentByLocation(
								ctx.fileName(),
								baseMsg.Name,
								field.Name,
								depIdx,
								argIdx,
							),
						),
					)
					continue
				}
				node, exists := nameToNode[ref]
				if !exists {
					// this referenced name was created by message option.
					continue
				}
				for _, depMessageNode := range msgToNodes[depMessage.Message] {
					if _, exists := node.ChildrenMap[depMessageNode]; !exists {
						node.Children = append(node.Children, depMessageNode)
						node.ChildrenMap[depMessageNode] = struct{}{}
					}
					if _, exists := depMessageNode.ParentMap[node]; !exists {
						depMessageNode.Parent = append(depMessageNode.Parent, node)
						depMessageNode.ParentMap[node] = struct{}{}
					}
				}
			}
		}
	}

	var roots []*MessageDependencyGraphNode
	for _, nodes := range msgToNodes {
		for _, node := range nodes {
			if len(node.Parent) == 0 {
				roots = append(roots, node)
			}
		}
	}
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
		for idx, dep := range msg.Rule.MessageDependencies {
			if dep.Message == target.Message {
				return ErrWithLocation(
					fmt.Sprintf(
						`found cyclic dependency for "%s.%s" message in "%s.%s. dependency path: %s"`,
						target.Message.PackageName(), target.Message.Name,
						msg.PackageName(), msg.Name, dependencyPath,
					),
					source.MessageDependencyMessageLocation(msg.File.Name, msg.Name, idx),
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
