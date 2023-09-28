package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/source"
)

type MessageDependencyGraph struct {
	Roots    []*MessageDependencyGraphNode
	RootArgs map[*MessageDependencyGraphNode]*Message
}

type MessageDependencyGraphNode struct {
	Parent          []*MessageDependencyGraphNode
	Children        []*MessageDependencyGraphNode
	Message         *Message
	MessageArgument *Message
}

func (n *MessageDependencyGraphNode) ExpectedMessageArguments() []*Argument {
	var (
		args       []*Argument
		argNameMap = map[string]struct{}{}
	)
	msg := n.Message
	for _, parent := range n.Parent {
		for _, dep := range parent.Message.Rule.MessageDependencies {
			if dep.Message != msg {
				continue
			}
			for _, arg := range dep.Args {
				if _, exists := argNameMap[arg.Name]; !exists {
					args = append(args, arg)
					argNameMap[arg.Name] = struct{}{}
				}
			}
		}
	}
	return args
}

// CreateMessageDependencyGraph creates a dependency graph for all messages with  message options defined.
func CreateMessageDependencyGraph(ctx *context, msgs []*Message) *MessageDependencyGraph {
	msgToNode := map[*Message]*MessageDependencyGraphNode{}
	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}
		msgToNode[msg] = &MessageDependencyGraphNode{Message: msg}
	}
	for _, msg := range msgs {
		if msg.Rule == nil {
			continue
		}
		node := msgToNode[msg]
		for depIdx, depMessage := range msg.Rule.MessageDependencies {
			depNode, exists := msgToNode[depMessage.Message]
			if !exists {
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
				depNode = &MessageDependencyGraphNode{Message: depMessage.Message}
			}
			node.Children = append(node.Children, depNode)
			depNode.Parent = append(depNode.Parent, node)
		}
	}
	var roots []*MessageDependencyGraphNode
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
	graph := &MessageDependencyGraph{
		Roots:    roots,
		RootArgs: map[*MessageDependencyGraphNode]*Message{},
	}
	if err := validateMessageGraph(graph); err != nil {
		ctx.addError(err)
		return nil
	}
	return graph
}

type MessageRuleDependencyGraph struct {
	Rule  *MessageRule
	Roots []*MessageRuleDependencyGraphNode
}

func (g *MessageRuleDependencyGraph) MessageResolverGroups(ctx *context) []MessageResolverGroup {
	var groups []MessageResolverGroup
	for _, child := range g.uniqueChildren() {
		if group := g.createMessageResolverGroup(ctx, child); group != nil {
			groups = append(groups, group)
		}
	}
	return groups
}

func (g *MessageRuleDependencyGraph) uniqueChildren() []*MessageRuleDependencyGraphNode {
	children := g.children(g.Roots)
	uniqueMap := make(map[*MessageRuleDependencyGraphNode]struct{})
	for _, child := range children {
		uniqueMap[child] = struct{}{}
	}
	uniqueChildren := make([]*MessageRuleDependencyGraphNode, 0, len(uniqueMap))
	for child := range uniqueMap {
		uniqueChildren = append(uniqueChildren, child)
	}
	sort.Slice(uniqueChildren, func(i, j int) bool {
		return uniqueChildren[i].Message.Name < uniqueChildren[j].Message.Name
	})
	return uniqueChildren
}

func (g *MessageRuleDependencyGraph) children(nodes []*MessageRuleDependencyGraphNode) []*MessageRuleDependencyGraphNode {
	var children []*MessageRuleDependencyGraphNode
	for _, node := range nodes {
		if len(node.Children) != 0 {
			children = append(children, g.children(node.Children)...)
		} else {
			children = append(children, node)
		}
	}
	return children
}

func (g *MessageRuleDependencyGraph) createMessageResolverGroup(ctx *context, node *MessageRuleDependencyGraphNode) MessageResolverGroup {
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
		return node.Parent[i].Message.Name < node.Parent[j].Message.Name
	})
	for _, parent := range node.Parent {
		if group := g.createMessageResolverGroup(ctx, parent); group != nil {
			rg.Starts = append(rg.Starts, group)
		}
	}
	rg.End = g.createMessageResolver(ctx, node)
	return rg
}

func (g *MessageRuleDependencyGraph) createMessageResolver(ctx *context, node *MessageRuleDependencyGraphNode) *MessageResolver {
	if g.Rule.MethodCall != nil {
		if g.Rule.MethodCall.Response.Type == node.Message {
			var name string
			methodCall := g.Rule.MethodCall
			if methodCall.Method != nil {
				name = methodCall.Method.Name
			}
			return &MessageResolver{Name: name, MethodCall: methodCall}
		}
	}
	for _, depMessage := range g.Rule.MessageDependencies {
		if node.Message == depMessage.Message {
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

type MessageRuleDependencyGraphNode struct {
	Parent      []*MessageRuleDependencyGraphNode
	Children    []*MessageRuleDependencyGraphNode
	ParentMap   map[*MessageRuleDependencyGraphNode]struct{}
	ChildrenMap map[*MessageRuleDependencyGraphNode]struct{}
	BaseMessage *Message
	Message     *Message
}

// CreateMessageRuleDependencyGraph construct a dependency graph using the name-based reference dependencies used in the method calls
// and the arguments used to retrieve the dependency messages.
// Requires reference resolution for arguments that use prior name-based references.
// If a circular reference occurs, return an error.
func CreateMessageRuleDependencyGraph(ctx *context, baseMsg *Message, rule *MessageRule) *MessageRuleDependencyGraph {
	msgToNode := map[*Message]*MessageRuleDependencyGraphNode{}
	if rule.MethodCall != nil && rule.MethodCall.Response != nil {
		msg := rule.MethodCall.Response.Type
		if msg != nil {
			msgToNode[msg] = &MessageRuleDependencyGraphNode{
				BaseMessage: baseMsg,
				Message:     msg,
				ParentMap:   map[*MessageRuleDependencyGraphNode]struct{}{},
				ChildrenMap: map[*MessageRuleDependencyGraphNode]struct{}{},
			}
		}
	}
	for _, depMessage := range rule.MessageDependencies {
		msg := depMessage.Message
		if msg == nil {
			continue
		}
		msgToNode[msg] = &MessageRuleDependencyGraphNode{
			BaseMessage: baseMsg,
			Message:     msg,
			ParentMap:   map[*MessageRuleDependencyGraphNode]struct{}{},
			ChildrenMap: map[*MessageRuleDependencyGraphNode]struct{}{},
		}
	}
	if rule.MethodCall != nil && rule.MethodCall.Request != nil && rule.MethodCall.Response != nil {
		for _, arg := range rule.MethodCall.Request.Args {
			if arg.Value == nil {
				continue
			}
			if arg.Value.PathType == MessageArgumentPathType {
				continue
			}
			typeRef := arg.Value.Ref
			if typeRef == nil {
				continue
			}
			msg := typeRef.Ref
			if msg == nil {
				continue
			}
			node, exists := msgToNode[msg]
			if !exists {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`%q message does not exist for resolving message argument`, msg.Name),
						source.MessageOptionLocation(ctx.fileName(), baseMsg.Name),
					),
				)
				continue
			}
			responseTypeNode := msgToNode[rule.MethodCall.Response.Type]
			if responseTypeNode != nil {
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
	for _, depMessage := range rule.MessageDependencies {
		for _, arg := range depMessage.Args {
			if arg.Value == nil {
				continue
			}
			if arg.Value.PathType == MessageArgumentPathType {
				continue
			}
			typeRef := arg.Value.Ref
			if typeRef == nil {
				continue
			}
			msg := typeRef.Ref
			if msg == nil {
				continue
			}
			node, exists := msgToNode[msg]
			if !exists {
				ctx.addError(
					ErrWithLocation(
						fmt.Sprintf(`%q message does not exist for resolving message argument`, msg.Name),
						source.MessageOptionLocation(ctx.fileName(), baseMsg.Name),
					),
				)
				continue
			}
			depMessageNode := msgToNode[depMessage.Message]
			if depMessageNode != nil {
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
	var roots []*MessageRuleDependencyGraphNode
	msgs := make([]*Message, 0, len(msgToNode))
	for msg, node := range msgToNode {
		msgs = append(msgs, msg)
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Message.Name < node.Children[j].Message.Name
		})
		sort.Slice(node.Parent, func(i, j int) bool {
			return node.Parent[i].Message.Name < node.Parent[j].Message.Name
		})
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Name < msgs[j].Name
	})
	for _, msg := range msgs {
		node := msgToNode[msg]
		if len(node.Parent) == 0 {
			roots = append(roots, node)
		}
	}
	if len(roots) == 0 && !rule.CustomResolver && rule.Alias == nil {
		ctx.addError(
			ErrWithLocation(
				"root message does not exist in message rule dependency graph",
				source.MessageOptionLocation(ctx.fileName(), baseMsg.Name),
			),
		)
		return nil
	}
	graph := &MessageRuleDependencyGraph{
		Rule:  rule,
		Roots: roots,
	}
	if err := validateMessageRuleGraph(graph); err != nil {
		ctx.addError(err)
		return nil
	}
	return graph
}

func validateMessageRuleGraph(graph *MessageRuleDependencyGraph) *LocationError {
	for _, root := range graph.Roots {
		if err := validateMessageRuleNode(root); err != nil {
			return err
		}
	}
	return nil
}

func validateMessageRuleNode(node *MessageRuleDependencyGraphNode) *LocationError {
	if err := validateMessageRuleNodeCyclicDependency(node, make(map[*MessageRuleDependencyGraphNode]struct{}), []*MessageRuleDependencyGraphNode{}); err != nil {
		return err
	}
	return nil
}

func validateMessageRuleNodeCyclicDependency(target *MessageRuleDependencyGraphNode, visited map[*MessageRuleDependencyGraphNode]struct{}, path []*MessageRuleDependencyGraphNode) *LocationError {
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
		if err := validateMessageRuleNodeCyclicDependency(child, visited, path); err != nil {
			return err
		}
	}
	delete(visited, target)
	return nil
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

		return ErrWithLocation(
			fmt.Sprintf(`found cyclic dependency in "%s.%s" message. dependency path: %s`, target.Message.PackageName(), target.Message.Name, dependencyPath),
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
