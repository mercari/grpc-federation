//go:build !tinygo.wasm

package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/source"
)

type ValueCandidate struct {
	Name string
	Type *Type
}

type ValueCandidates []*ValueCandidate

func (c ValueCandidates) Names() []string {
	ret := make([]string, 0, len(c))
	for _, cc := range c {
		ret = append(ret, cc.Name)
	}
	return ret
}

func (c ValueCandidates) Unique() ValueCandidates {
	ret := make(ValueCandidates, 0, len(c))
	nameMap := make(map[string]struct{})
	for _, cc := range c {
		if _, exists := nameMap[cc.Name]; exists {
			continue
		}
		ret = append(ret, cc)
		nameMap[cc.Name] = struct{}{}
	}
	return ret
}

func (c ValueCandidates) Filter(typ *Type) ValueCandidates {
	ret := make(ValueCandidates, 0, len(c))
	for _, cc := range c {
		if cc.Type == nil || typ == nil {
			ret = append(ret, cc)
			continue
		}
		if cc.Type.Message == nil && cc.Type.Enum == nil {
			if cc.Type.Kind == typ.Kind {
				ret = append(ret, cc)
				continue
			}
		}
		if cc.Type == typ {
			ret = append(ret, cc)
			continue
		}
		if cc.Type.IsNumber() && typ.IsNumber() {
			ret = append(ret, cc)
			continue
		}
		if cc.Type.Message != nil && typ.Message != nil && cc.Type.Message == typ.Message {
			ret = append(ret, cc)
			continue
		}
		if cc.Type.Enum != nil && typ.Enum != nil && cc.Type.Enum == typ.Enum {
			ret = append(ret, cc)
			continue
		}
	}
	return ret
}

func (r *Resolver) Candidates(loc *source.Location) []string {
	for _, file := range r.files {
		if strings.HasSuffix(file.GetName(), loc.FileName) {
			protoPkgName := file.GetPackage()
			switch {
			case loc.Message != nil:
				msg := r.cachedMessageMap[fmt.Sprintf("%s.%s", protoPkgName, loc.Message.Name)]
				return r.candidatesMessage(msg, loc.Message)
			case loc.Service != nil:
			}
		}
	}
	return nil
}

func (r *Resolver) candidatesMessage(msg *Message, loc *source.Message) []string {
	switch {
	case loc.Option != nil:
		return r.candidatesMessageOption(msg, loc.Option)
	case loc.Field != nil:
		return r.candidatesField(msg, loc.Field)
	}
	return nil
}

func (r *Resolver) candidatesMessageOption(msg *Message, opt *source.MessageOption) []string {
	switch {
	case opt.VariableDefinitions != nil:
		return r.candidatesVariableDefinitions(msg, opt.VariableDefinitions)
	}
	return []string{"def", "custom_resolver", "alias"}
}

func (r *Resolver) candidatesField(msg *Message, field *source.Field) []string {
	switch {
	case field.Option != nil:
		return r.candidatesFieldOption(msg, field)
	}
	return nil
}

func (r *Resolver) candidatesFieldOption(msg *Message, field *source.Field) []string {
	switch {
	case field.Option.By:
		f := msg.Field(field.Name)
		if f == nil {
			return nil
		}
		if msg.Rule == nil {
			return []string{}
		}
		defIdx := len(msg.Rule.VariableDefinitions) - 1
		return r.candidatesCELValue(msg, defIdx).Unique().Filter(f.Type).Names()
	}
	return []string{"by"}
}

func (r *Resolver) candidatesVariableDefinitions(msg *Message, opt *source.VariableDefinitionOption) []string {
	switch {
	case opt.Name:
		return []string{}
	case opt.By:
		return r.candidatesCELValue(msg, opt.Idx-1).Unique().Names()
	case opt.Map != nil:
		return r.candidatesMapExpr(msg, opt.Idx, opt.Map)
	case opt.Call != nil:
		return r.candidatesCallExpr(msg, opt.Idx, opt.Call)
	case opt.Message != nil:
		return r.candidatesMessageExpr(msg, opt.Idx, opt.Message)
	}
	return []string{"name", "by", "map", "call", "message", "validation"}
}

func (r *Resolver) candidatesCallExpr(msg *Message, defIdx int, opt *source.CallExprOption) []string {
	switch {
	case opt.Method:
		return r.candidatesMethodName(msg)
	case opt.Request != nil:
		return r.candidatesRequest(msg, defIdx, opt.Request)
	case opt.Timeout:
	case opt.Retry != nil:
	}
	return []string{"method", "request", "timeout", "retry"}
}

func (r *Resolver) candidatesRequest(msg *Message, defIdx int, opt *source.RequestOption) []string {
	switch {
	case opt.Field:
		return r.candidatesRequestField(msg, defIdx)
	case opt.By:
		typ := r.requestFieldType(msg, defIdx, opt.Idx)
		return r.candidatesCELValue(msg, defIdx).Unique().Filter(typ).Names()
	}
	return []string{"field", "by"}
}

func (r *Resolver) candidatesMessageExpr(msg *Message, defIdx int, opt *source.MessageExprOption) []string {
	switch {
	case opt.Name:
		return r.candidatesMessageName(msg)
	case opt.Args != nil:
		return r.candidatesMessageExprArguments(msg, defIdx, opt.Args)
	}
	return []string{"name", "args"}
}

func (r *Resolver) candidatesMapExpr(msg *Message, defIdx int, opt *source.MapExprOption) []string {
	return nil
}

func (r *Resolver) candidatesMessageExprArguments(msg *Message, defIdx int, arg *source.ArgumentOption) []string {
	switch {
	case arg.Name:
		return []string{}
	case arg.By, arg.Inline:
		return r.candidatesCELValue(msg, defIdx).Unique().Names()
	}
	return []string{"name", "by", "inline"}
}

func (r *Resolver) candidatesMethodName(msg *Message) []string {
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

func (r *Resolver) candidatesRequestField(msg *Message, defIdx int) []string {
	if msg.Rule == nil {
		return nil
	}

	if len(msg.Rule.VariableDefinitions) <= defIdx {
		return nil
	}

	def := msg.Rule.VariableDefinitions[defIdx]
	if def.Expr == nil {
		return nil
	}
	if def.Expr.Call == nil {
		return nil
	}
	if def.Expr.Call.Method == nil {
		return nil
	}
	if def.Expr.Call.Method.Request == nil {
		return nil
	}

	fields := def.Expr.Call.Method.Request.Fields
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name)
	}
	return names
}

func (r *Resolver) requestFieldType(msg *Message, defIdx, reqIdx int) *Type {
	if msg.Rule == nil {
		return nil
	}
	if len(msg.Rule.VariableDefinitions) <= defIdx {
		return nil
	}

	def := msg.Rule.VariableDefinitions[defIdx]
	if def.Expr == nil {
		return nil
	}
	if def.Expr.Call == nil {
		return nil
	}
	if def.Expr.Call.Method == nil {
		return nil
	}
	if def.Expr.Call.Method.Request == nil {
		return nil
	}

	fields := def.Expr.Call.Method.Request.Fields
	if len(fields) <= reqIdx {
		return nil
	}
	return fields[reqIdx].Type
}

func (r *Resolver) candidatesMessageName(msg *Message) []string {
	const federationPkgName = "grpc.federation."

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
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Resolver) candidatesCELValue(msg *Message, defIdx int) ValueCandidates {
	var ret ValueCandidates
	ret = append(ret, r.candidatesVariableName(msg, defIdx)...)
	ret = append(ret, r.candidatesMessageArguments(msg)...)
	return ret
}

func (r *Resolver) candidatesVariableName(msg *Message, defIdx int) []*ValueCandidate {
	if msg.Rule == nil {
		return nil
	}
	if len(msg.Rule.VariableDefinitions) <= defIdx {
		return nil
	}

	var ret []*ValueCandidate
	for i := 0; i < defIdx+1; i++ {
		def := msg.Rule.VariableDefinitions[defIdx]
		name := def.Name
		if name == "" || strings.HasPrefix(name, "_") {
			continue
		}
		if def.Expr == nil {
			continue
		}
		ret = append(ret, &ValueCandidate{
			Name: name,
			Type: def.Expr.Type,
		})
	}
	return ret
}

func (r *Resolver) candidatesMessageArguments(msg *Message) []*ValueCandidate {
	if msg.Rule == nil {
		return nil
	}
	if msg.Rule.MessageArgument == nil {
		return nil
	}
	arg := msg.Rule.MessageArgument
	names := make(ValueCandidates, 0, len(arg.Fields))
	for _, field := range arg.Fields {
		names = append(names, &ValueCandidate{
			Name: fmt.Sprintf("$.%s", field.Name),
			Type: field.Type,
		})
	}
	return names
}
