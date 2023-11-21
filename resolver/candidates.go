package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/types"
)

func (r *Resolver) Candidates(loc *source.Location) []string {
	for _, file := range r.files {
		if strings.HasSuffix(file.GetName(), loc.FileName) {
			protoPkgName := file.GetPackage()
			switch {
			case loc.Message != nil:
				msg := r.cachedMessageMap[fmt.Sprintf("%s.%s", protoPkgName, loc.Message.Name)]
				return r.candidatesFromMessage(msg, loc.Message)
			case loc.Service != nil:
			}
		}
	}
	return nil
}

func (r *Resolver) candidatesFromMessage(msg *Message, loc *source.Message) []string {
	switch {
	case loc.Option != nil:
		return r.candidatesFromMessageOption(msg, loc.Option)
	case loc.Field != nil:
		return r.candidatesFromField(msg, loc.Field)
	}
	return nil
}

func (r *Resolver) candidatesFromMessageOption(msg *Message, opt *source.MessageOption) []string {
	switch {
	case opt.Resolver != nil:
		return r.candidatesFromResolver(msg, opt.Resolver)
	case opt.Messages != nil:
		return r.candidatesFromMessageDependency(msg, opt.Messages)
	}
	return []string{"resolver", "messages"}
}

func (r *Resolver) candidatesFromField(msg *Message, field *source.Field) []string {
	switch {
	case field.Option != nil:
		return r.candidatesFromFieldOption(msg, field)
	}
	return nil
}

func (r *Resolver) candidatesFromFieldOption(msg *Message, field *source.Field) []string {
	switch {
	case field.Option.By:
		f := msg.Field(field.Name)
		if f == nil {
			return nil
		}
		return r.candidatesFieldValue(f.Type, msg)
	}
	return []string{"by"}
}

func (r *Resolver) candidatesFromResolver(msg *Message, resolver *source.ResolverOption) []string {
	switch {
	case resolver.Method:
		return r.candidatesResolverMethodName(msg)
	case resolver.Request != nil:
		return r.candidatesFromRequest(msg, resolver.Request)
	case resolver.Response != nil:
		return r.candidatesFromResponse(msg, resolver.Response)
	}
	return []string{"method", "request", "response"}
}

func (r *Resolver) candidatesFromRequest(msg *Message, req *source.RequestOption) []string {
	switch {
	case req.Field:
		return r.candidatesResolverRequestField(msg)
	case req.By:
		expectedType := r.requestFieldType(msg, req.Idx)
		return r.candidatesResolverRequestValue(expectedType, msg)
	}
	return []string{"field", "by"}
}

func (r *Resolver) candidatesFromResponse(msg *Message, res *source.ResponseOption) []string {
	switch {
	case res.Name:
		return []string{}
	case res.Field:
		return r.candidatesResolverResponseField(msg)
	case res.AutoBind:
		return []string{"true", "false"}
	}
	return []string{"name", "field", "autobind"}
}

func (r *Resolver) candidatesFromMessageDependency(msg *Message, dep *source.MessageDependencyOption) []string {
	switch {
	case dep.Name:
		return []string{}
	case dep.Message:
		return r.candidatesDependencyMessageValue(msg)
	case dep.Args != nil:
		return r.candidatesFromMessageCallArguments(msg, dep.Idx, dep.Args)
	}
	return []string{"name", "message", "args"}
}

func (r *Resolver) candidatesFromMessageCallArguments(msg *Message, depIdx int, arg *source.ArgumentOption) []string {
	switch {
	case arg.Name:
		return []string{}
	case arg.By:
		expectedType := r.messageCallArgumentType(msg, depIdx, arg.Idx)
		return r.candidatesMessageCallArgumentValue(expectedType, msg, depIdx)
	case arg.Inline:
		expectedType := &Type{Type: types.Message}
		return r.candidatesMessageCallArgumentValue(expectedType, msg, depIdx)
	}
	return []string{"name", "by", "inline"}
}

func (r *Resolver) candidatesResolverMethodName(msg *Message) []string {
	selfPkgName := fmt.Sprintf("%s.", msg.PackageName())
	names := make([]string, 0, len(r.cachedMethodMap))
	for name := range r.cachedMethodMap {
		if strings.HasPrefix(name, selfPkgName) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Resolver) candidatesResolverRequestField(msg *Message) []string {
	if msg.Rule == nil {
		return nil
	}

	methodCall := msg.Rule.MethodCall
	if methodCall == nil {
		return nil
	}
	if methodCall.Method == nil {
		return nil
	}
	if methodCall.Method.Request == nil {
		return nil
	}

	names := make([]string, 0, len(methodCall.Method.Request.Fields))
	for _, field := range methodCall.Method.Request.Fields {
		names = append(names, field.Name)
	}
	return names
}

func (r *Resolver) candidatesResolverResponseField(msg *Message) []string {
	if msg.Rule == nil {
		return nil
	}

	methodCall := msg.Rule.MethodCall
	if methodCall == nil {
		return nil
	}
	if methodCall.Method == nil {
		return nil
	}
	if methodCall.Method.Response == nil {
		return nil
	}

	names := make([]string, 0, len(methodCall.Method.Response.Fields))
	for _, field := range methodCall.Method.Response.Fields {
		names = append(names, field.Name)
	}
	return names
}

func (r *Resolver) candidatesResolverRequestValue(expectedType *Type, msg *Message) []string {
	var names []string
	names = append(names, r.candidatesFromMessageDependencyName(expectedType, msg)...)
	names = append(names, r.candidatesFromMessageArguments(expectedType, msg)...)
	return names
}

func (r *Resolver) requestFieldType(msg *Message, idx int) *Type {
	if msg.Rule == nil {
		return nil
	}

	methodCall := msg.Rule.MethodCall
	if methodCall == nil {
		return nil
	}
	if methodCall.Method == nil {
		return nil
	}
	if methodCall.Method.Request == nil {
		return nil
	}
	if len(methodCall.Method.Request.Fields) <= idx {
		return nil
	}
	return methodCall.Method.Request.Fields[idx].Type
}

func (r *Resolver) messageCallArgumentType(msg *Message, msgIdx, argIdx int) *Type {
	if msg.Rule == nil {
		return nil
	}
	if len(msg.Rule.MessageDependencies) <= msgIdx {
		return nil
	}
	dep := msg.Rule.MessageDependencies[msgIdx]
	if dep.Message == nil {
		return nil
	}
	if len(dep.Args) <= argIdx {
		return nil
	}
	argName := dep.Args[argIdx].Name
	if dep.Message.Rule == nil {
		return nil
	}
	if dep.Message.Rule.MessageArgument == nil {
		return nil
	}
	for _, argField := range dep.Message.Rule.MessageArgument.Fields {
		if argField.Name == argName {
			return argField.Type
		}
	}
	return nil
}

func (r *Resolver) candidatesMessageCallArgumentValue(expectedType *Type, msg *Message, depIdx int) []string {
	var names []string
	names = append(names, r.candidatesFromResolverResponseName(expectedType, msg, false)...)
	names = append(names, r.candidatesFromMessageDependencyNameWithIgnoreIdx(expectedType, msg, depIdx)...)
	names = append(names, r.candidatesFromMessageArguments(expectedType, msg)...)
	return names
}

func (r *Resolver) candidatesFieldValue(expectedType *Type, msg *Message) []string {
	var names []string
	names = append(names, r.candidatesFromResolverResponseName(expectedType, msg, true)...)
	names = append(names, r.candidatesFromMessageDependencyName(expectedType, msg)...)
	names = append(names, r.candidatesFromMessageArguments(expectedType, msg)...)
	return names
}

func (r *Resolver) candidatesDependencyMessageValue(msg *Message) []string {
	const federationPkgName = "grpc.federation."

	definedMessageNameMap := make(map[string]struct{})
	if msg.Rule != nil {
		for _, dep := range msg.Rule.MessageDependencies {
			if dep.Message == nil {
				continue
			}
			definedMessageNameMap[dep.Message.Name] = struct{}{}
		}
	}
	selfPkgName := fmt.Sprintf("%s.", msg.PackageName())
	selfName := fmt.Sprintf("%s.%s", msg.PackageName(), msg.Name)
	names := make([]string, 0, len(r.cachedMessageMap))
	for name := range r.cachedMessageMap {
		if name == selfName {
			continue
		}
		if strings.HasPrefix(name, federationPkgName) {
			continue
		}
		name = strings.TrimPrefix(name, selfPkgName)
		if _, exists := definedMessageNameMap[name]; exists {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Resolver) candidatesFromMessageArguments(expectedType *Type, msg *Message) []string {
	if msg.Rule == nil {
		return nil
	}
	if msg.Rule.MessageArgument == nil {
		return nil
	}
	arg := msg.Rule.MessageArgument
	names := make([]string, 0, len(arg.Fields))
	for _, field := range arg.Fields {
		if !r.isAssignableType(expectedType, field.Type) {
			continue
		}
		names = append(names, fmt.Sprintf("$.%s", field.Name))
	}
	return names
}

func (r *Resolver) candidatesFromMessageDependencyName(expectedType *Type, msg *Message) []string {
	return r.candidatesFromMessageDependencyNameWithIgnoreIdx(expectedType, msg, -1)
}

func (r *Resolver) candidatesFromMessageDependencyNameWithIgnoreIdx(expectedType *Type, msg *Message, ignoreIdx int) []string {
	if msg.Rule == nil {
		return nil
	}

	var names []string
	for idx, dep := range msg.Rule.MessageDependencies {
		if idx == ignoreIdx {
			continue
		}
		if dep.Name == "" {
			continue
		}
		typ := &Type{Type: types.Message, Ref: dep.Message}
		if !r.isAssignableType(expectedType, typ) {
			continue
		}
		names = append(names, dep.Name)
	}
	return names
}

func (r *Resolver) candidatesFromResolverResponseName(expectedType *Type, msg *Message, expandAutoBindField bool) []string {
	if msg.Rule == nil {
		return nil
	}
	if msg.Rule.MethodCall == nil {
		return nil
	}
	if msg.Rule.MethodCall.Response == nil {
		return nil
	}

	var names []string
	for _, field := range msg.Rule.MethodCall.Response.Fields {
		if expandAutoBindField && field.AutoBind && field.Type != nil && field.Type.Ref != nil {
			for _, autoBindField := range field.Type.Ref.Fields {
				if autoBindField.Name == "" {
					continue
				}
				if r.isAssignableType(expectedType, autoBindField.Type) {
					names = append(names, autoBindField.Name)
				}
			}
		}
		if field.Name != "" && r.isAssignableType(expectedType, field.Type) {
			names = append(names, field.Name)
		}
	}
	return names
}

func (r *Resolver) isAssignableType(expected, got *Type) bool {
	if expected == nil {
		return true
	}
	if got == nil {
		return true
	}
	if expected.Type == got.Type {
		return true
	}
	if expected.IsNumber() && got.IsNumber() {
		return true
	}
	return false
}
