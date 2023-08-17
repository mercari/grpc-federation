package testutil

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/mercari/grpc-federation/resolver"
)

func ResolverCmpOpts() []cmp.Option {
	return []cmp.Option{
		cmpopts.IgnoreFields(resolver.File{}, "Messages", "Services", "Enums"),
		cmpopts.IgnoreFields(resolver.Package{}, "Files"),
		cmpopts.IgnoreFields(resolver.Method{}, "Service"),
		cmpopts.IgnoreFields(resolver.Message{}, "File", "ParentMessage"),
		cmpopts.IgnoreFields(resolver.Enum{}, "File", "Message.Rule"),
		cmpopts.IgnoreFields(resolver.EnumRule{}, "Alias.Rule"),
		cmpopts.IgnoreFields(resolver.MessageResolver{}, "MethodCall", "MessageDependency"),
		cmpopts.IgnoreFields(resolver.MessageDependency{}, "Message.Rule"),
		cmpopts.IgnoreFields(resolver.MessageRule{}, "MessageDependencies", "Alias.Rule"),
		cmpopts.IgnoreFields(resolver.MessageRuleDependencyGraph{}, "Rule"),
		cmpopts.IgnoreFields(resolver.MessageRuleDependencyGraphNode{}, "Parent", "ParentMap", "Children", "ChildrenMap", "Message.Rule"),
		cmpopts.IgnoreFields(resolver.MessageDependencyGraph{}, "RootArgs"),
		cmpopts.IgnoreFields(resolver.MessageDependencyGraphNode{}, "Parent", "Children", "Message.Rule"),
		cmpopts.IgnoreFields(resolver.AutoBindField{}, "ResponseField", "MessageDependency"),
		cmpopts.IgnoreFields(resolver.Type{}, "Ref.Rule", "Enum.Rule"),
	}
}
