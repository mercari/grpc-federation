package resolver

import (
	"sort"
)

func (set *VariableDefinitionSet) Definitions() VariableDefinitions {
	if set == nil {
		return nil
	}
	return set.Defs
}

func (set *VariableDefinitionSet) DefinitionGroups() []VariableDefinitionGroup {
	if set == nil {
		return nil
	}
	return set.Groups
}

func (set *VariableDefinitionSet) DependencyGraph() *MessageDependencyGraph {
	if set == nil {
		return nil
	}
	return set.Graph
}

func (e *GRPCError) Definitions() VariableDefinitions {
	var ret VariableDefinitions
	if e.DefSet != nil {
		ret = append(ret, e.DefSet.Definitions()...)
	}
	for _, detail := range e.Details {
		ret = append(ret, detail.DefSet.Definitions()...)
		ret = append(ret, detail.Messages.Definitions()...)
	}
	return ret
}

func (e *GRPCError) DefinitionGroups() []VariableDefinitionGroup {
	var ret []VariableDefinitionGroup
	if e.DefSet != nil {
		ret = append(ret, e.DefSet.DefinitionGroups()...)
	}
	for _, detail := range e.Details {
		ret = append(ret, detail.DefSet.DefinitionGroups()...)
		ret = append(ret, detail.Messages.DefinitionGroups()...)
	}
	return ret
}

func (s *SwitchExpr) Definitions() VariableDefinitions {
	var defs VariableDefinitions
	for _, cse := range s.Cases {
		defs = append(defs, cse.DefSet.Definitions()...)
	}
	defs = append(defs, s.Default.DefSet.Definitions()...)
	return defs
}

func (s *SwitchExpr) DefinitionGroups() []VariableDefinitionGroup {
	var groups []VariableDefinitionGroup
	for _, cse := range s.Cases {
		groups = append(groups, cse.DefSet.DefinitionGroups()...)
	}
	groups = append(groups, s.Default.DefSet.DefinitionGroups()...)
	return groups
}

func (g *SequentialVariableDefinitionGroup) VariableDefinitions() VariableDefinitions {
	var defs VariableDefinitions
	if g.Start != nil {
		defs = append(defs, g.Start.VariableDefinitions()...)
	}
	defs = append(defs, g.End)
	return defs
}

func (g *ConcurrentVariableDefinitionGroup) VariableDefinitions() VariableDefinitions {
	var defs VariableDefinitions
	for _, start := range g.Starts {
		defs = append(defs, start.VariableDefinitions()...)
	}
	defs = append(defs, g.End)
	return defs
}

func (set *VariableDefinitionSet) MessageToDefsMap() map[*Message]VariableDefinitions {
	ret := make(map[*Message]VariableDefinitions)
	for _, varDef := range set.Definitions() {
		for k, v := range varDef.MessageToDefsMap() {
			ret[k] = append(ret[k], v...)
		}
	}
	return ret
}

func (def *VariableDefinition) MessageToDefsMap() map[*Message]VariableDefinitions {
	if def.Expr == nil {
		return nil
	}
	expr := def.Expr
	switch {
	case expr.Message != nil:
		msgExpr := expr.Message
		if msgExpr.Message == nil {
			return nil
		}
		return map[*Message]VariableDefinitions{msgExpr.Message: {def}}
	case expr.Map != nil:
		mapExpr := expr.Map
		if mapExpr.Expr == nil {
			return nil
		}
		if mapExpr.Expr.Message == nil {
			return nil
		}
		msgExpr := mapExpr.Expr.Message
		return map[*Message]VariableDefinitions{msgExpr.Message: {def}}
	case expr.Call != nil:
		return expr.Call.MessageToDefsMap()
	case expr.Switch != nil:
		return expr.Switch.MessageToDefsMap()
	case expr.Validation != nil:
		return expr.Validation.MessageToDefsMap()
	}
	return nil
}

func (e *CallExpr) MessageToDefsMap() map[*Message]VariableDefinitions {
	ret := make(map[*Message]VariableDefinitions)
	for _, grpcErr := range e.Errors {
		for k, v := range grpcErr.MessageToDefsMap() {
			ret[k] = append(ret[k], v...)
		}
	}
	return ret
}

func (s *SwitchExpr) MessageToDefsMap() map[*Message]VariableDefinitions {
	ret := s.Default.DefSet.MessageToDefsMap()
	for _, cse := range s.Cases {
		for k, v := range cse.DefSet.MessageToDefsMap() {
			ret[k] = append(ret[k], v...)
		}
	}
	return ret
}

func (e *ValidationExpr) MessageToDefsMap() map[*Message]VariableDefinitions {
	return e.Error.MessageToDefsMap()
}

func (e *GRPCError) MessageToDefsMap() map[*Message]VariableDefinitions {
	ret := e.DefSet.MessageToDefsMap()
	for _, detail := range e.Details {
		for k, v := range detail.MessageToDefsMap() {
			ret[k] = append(ret[k], v...)
		}
	}
	return ret
}

func (detail *GRPCErrorDetail) MessageToDefsMap() map[*Message]VariableDefinitions {
	ret := detail.DefSet.MessageToDefsMap()
	for k, v := range detail.Messages.MessageToDefsMap() {
		ret[k] = append(ret[k], v...)
	}
	return ret
}

func (set *VariableDefinitionSet) MessageExprs() []*MessageExpr {
	var ret []*MessageExpr
	for _, varDef := range set.Definitions() {
		ret = append(ret, varDef.MessageExprs()...)
	}
	return ret
}

func (def *VariableDefinition) IsValidation() bool {
	if def.Expr == nil {
		return false
	}
	return def.Expr.Validation != nil
}

func (def *VariableDefinition) MessageExprs() []*MessageExpr {
	if def.Expr == nil {
		return nil
	}
	expr := def.Expr
	switch {
	case expr.Call != nil:
		return expr.Call.MessageExprs()
	case expr.Map != nil:
		return expr.Map.MessageExprs()
	case expr.Message != nil:
		return []*MessageExpr{expr.Message}
	case expr.Switch != nil:
		return expr.Switch.MessageExprs()
	case expr.Validation != nil:
		return expr.Validation.MessageExprs()
	}
	return nil
}

func (e *CallExpr) MessageExprs() []*MessageExpr {
	var ret []*MessageExpr
	for _, grpcErr := range e.Errors {
		ret = append(ret, grpcErr.MessageExprs()...)
	}
	return ret
}

func (e *MapExpr) MessageExprs() []*MessageExpr {
	if e.Expr == nil {
		return nil
	}
	if e.Expr.Message == nil {
		return nil
	}
	return []*MessageExpr{e.Expr.Message}
}

func (s *SwitchExpr) MessageExprs() []*MessageExpr {
	var ret []*MessageExpr
	for _, cse := range s.Cases {
		ret = append(ret, cse.DefSet.MessageExprs()...)
	}
	ret = append(ret, s.Default.DefSet.MessageExprs()...)
	return ret
}

func (e *ValidationExpr) MessageExprs() []*MessageExpr {
	return e.Error.MessageExprs()
}

func (e *GRPCError) MessageExprs() []*MessageExpr {
	ret := e.DefSet.MessageExprs()
	for _, detail := range e.Details {
		ret = append(ret, detail.MessageExprs()...)
	}
	return ret
}

func (detail *GRPCErrorDetail) MessageExprs() []*MessageExpr {
	return append(detail.DefSet.MessageExprs(), detail.Messages.MessageExprs()...)
}

func (set *VariableDefinitionSet) ReferenceNames() []string {
	var names []string
	for _, def := range set.Definitions() {
		names = append(names, def.ReferenceNames()...)
	}
	return toUniqueReferenceNames(names)
}

func (def *VariableDefinition) ReferenceNames() []string {
	var names []string
	if def.If != nil {
		names = append(names, def.If.ReferenceNames()...)
	}
	names = append(names, def.Expr.ReferenceNames()...)
	return toUniqueReferenceNames(names)
}

func (e *VariableExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}
	switch {
	case e.By != nil:
		return e.By.ReferenceNames()
	case e.Map != nil:
		return e.Map.ReferenceNames()
	case e.Call != nil:
		return e.Call.ReferenceNames()
	case e.Message != nil:
		return e.Message.ReferenceNames()
	case e.Enum != nil:
		return e.Enum.ReferenceNames()
	case e.Validation != nil:
		return e.Validation.Error.ReferenceNames()
	case e.Switch != nil:
		return e.Switch.ReferenceNames()
	}
	return nil
}

// ReferenceNames returns all the unique reference names in the error definition.
func (e *GRPCError) ReferenceNames() []string {
	names := e.If.ReferenceNames()
	for _, def := range e.DefSet.Definitions() {
		names = append(names, def.ReferenceNames()...)
	}
	for _, detail := range e.Details {
		names = append(names, detail.ReferenceNames()...)
	}
	return toUniqueReferenceNames(names)
}

func (detail *GRPCErrorDetail) ReferenceNames() []string {
	names := detail.If.ReferenceNames()
	for _, def := range detail.DefSet.Definitions() {
		names = append(names, def.ReferenceNames()...)
	}
	for _, def := range detail.Messages.Definitions() {
		names = append(names, def.ReferenceNames()...)
	}
	for _, by := range detail.By {
		names = append(names, by.ReferenceNames()...)
	}
	for _, failure := range detail.PreconditionFailures {
		for _, violation := range failure.Violations {
			names = append(names, violation.Type.ReferenceNames()...)
			names = append(names, violation.Subject.ReferenceNames()...)
			names = append(names, violation.Description.ReferenceNames()...)
		}
	}
	for _, req := range detail.BadRequests {
		for _, violation := range req.FieldViolations {
			names = append(names, violation.Field.ReferenceNames()...)
			names = append(names, violation.Description.ReferenceNames()...)
		}
	}
	for _, msg := range detail.LocalizedMessages {
		names = append(names, msg.Message.ReferenceNames()...)
	}
	return toUniqueReferenceNames(names)
}

func (e *CallExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}
	var names []string
	for _, arg := range e.Request.Args {
		names = append(names, arg.Value.ReferenceNames()...)
	}
	for _, grpcErr := range e.Errors {
		names = append(names, grpcErr.ReferenceNames()...)
	}
	if e.Option != nil {
		if e.Option.Header != nil && e.Option.Header.Name != "" {
			names = append(names, e.Option.Header.Name)
		}
		if e.Option.Trailer != nil && e.Option.Trailer.Name != "" {
			names = append(names, e.Option.Trailer.Name)
		}
	}
	if e.Metadata != nil {
		names = append(names, e.Metadata.ReferenceNames()...)
	}
	return toUniqueReferenceNames(names)
}

func (e *MapExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}

	var names []string
	if e.Iterator != nil {
		if e.Iterator.Name != "" {
			names = append(names, e.Iterator.Name)
		}
		if e.Iterator.Source != nil && e.Iterator.Source.Name != "" {
			names = append(names, e.Iterator.Source.Name)
		}
	}
	names = append(names, e.Expr.ReferenceNames()...)
	return toUniqueReferenceNames(names)
}

func (e *MapIteratorExpr) ReferenceNames() []string {
	if e == nil {
		return nil
	}

	var names []string
	switch {
	case e.By != nil:
		names = append(names, e.By.ReferenceNames()...)
	case e.Message != nil:
		names = append(names, e.Message.ReferenceNames()...)
	}
	return toUniqueReferenceNames(names)
}

func (s *SwitchExpr) ReferenceNames() []string {
	if s == nil {
		return nil
	}
	var names []string
	for _, cse := range s.Cases {
		names = append(names, cse.DefSet.ReferenceNames()...)
		names = append(names, cse.If.ReferenceNames()...)
		names = append(names, cse.By.ReferenceNames()...)
	}
	names = append(names, s.Default.DefSet.ReferenceNames()...)
	names = append(names, s.Default.By.ReferenceNames()...)
	return toUniqueReferenceNames(names)
}

func (set *VariableDefinitionSet) MarkUsed(nameRefMap map[string]struct{}) {
	for _, varDef := range set.Definitions() {
		varDef.MarkUsed(nameRefMap)
	}
}

func (def *VariableDefinition) MarkUsed(nameRefMap map[string]struct{}) {
	if _, exists := nameRefMap[def.Name]; exists {
		def.Used = true
	}
	if def.Expr == nil {
		return
	}
	expr := def.Expr
	switch {
	case expr.Call != nil:
		expr.Call.MarkUsed(nameRefMap)
	case expr.Map != nil:
		expr.Map.MarkUsed(nameRefMap)
	case expr.Switch != nil:
		expr.Switch.MarkUsed(nameRefMap)
	case expr.Validation != nil:
		expr.Validation.MarkUsed(nameRefMap)
	}
}

func (e *CallExpr) MarkUsed(nameRefMap map[string]struct{}) {
	for _, grpcErr := range e.Errors {
		grpcErr.MarkUsed(nameRefMap)
	}
}

func (e *MapExpr) MarkUsed(_ map[string]struct{}) {}

func (s *SwitchExpr) MarkUsed(nameRefMap map[string]struct{}) {
	for _, cse := range s.Cases {
		cse.DefSet.MarkUsed(nameRefMap)
	}
	s.Default.DefSet.MarkUsed(nameRefMap)
}

func (e *ValidationExpr) MarkUsed(nameRefMap map[string]struct{}) {
	e.Error.MarkUsed(nameRefMap)
}

func (e *GRPCError) MarkUsed(nameRefMap map[string]struct{}) {
	e.DefSet.MarkUsed(nameRefMap)
	for _, detail := range e.Details {
		detail.MarkUsed(nameRefMap)
	}
}

func (detail *GRPCErrorDetail) MarkUsed(nameRefMap map[string]struct{}) {
	detail.DefSet.MarkUsed(nameRefMap)
	detail.Messages.MarkUsed(nameRefMap)
}

func (e *MapIteratorExpr) ToVariableExpr() *VariableExpr {
	return &VariableExpr{
		Type:    e.Type,
		By:      e.By,
		Message: e.Message,
	}
}

func toUniqueReferenceNames(names []string) []string {
	nameMap := make(map[string]struct{})
	for _, name := range names {
		nameMap[name] = struct{}{}
	}
	ret := make([]string, 0, len(nameMap))
	for name := range nameMap {
		ret = append(ret, name)
	}
	sort.Strings(ret)
	return ret
}
