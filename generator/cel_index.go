package generator

import (
	"fmt"
	"sort"

	"github.com/google/cel-go/cel"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"github.com/mercari/grpc-federation/resolver"
)

// celVariable is a variable referenced by a CEL expression.
type celVariable struct {
	Name string
	Type *cel.Type
}

// referencedCELVariables returns the user-defined variables referenced by the checked expression.
func referencedCELVariables(checked *exprv1.CheckedExpr) ([]*celVariable, error) {
	seen := map[string]struct{}{}
	notVariable := nonVariableReferenceIDs(checked.GetExpr())
	var vars []*celVariable
	for id, ref := range checked.GetReferenceMap() {
		// An identifier reference has neither overloads nor a constant value (enum values).
		if len(ref.GetOverloadId()) != 0 || ref.GetValue() != nil {
			continue
		}
		if _, skip := notVariable[id]; skip {
			continue
		}
		name := ref.GetName()
		if isServiceWideCELVariable(name) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		celType, err := cel.ExprTypeToType(checked.GetTypeMap()[id])
		if err != nil {
			return nil, err
		}
		vars = append(vars, &celVariable{Name: name, Type: celType})
	}
	sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })
	return vars, nil
}

// nonVariableReferenceIDs returns the ids of expressions that have an entry in the reference map
// but are not variables declared in the env: identifiers bound by a comprehension
// (macro iteration and accumulator variables) and the type names of message literals.
func nonVariableReferenceIDs(root *exprv1.Expr) map[int64]struct{} {
	ret := map[int64]struct{}{}
	var walk func(e *exprv1.Expr, bound map[string]struct{})
	walkAll := func(es []*exprv1.Expr, bound map[string]struct{}) {
		for _, e := range es {
			walk(e, bound)
		}
	}
	walk = func(e *exprv1.Expr, bound map[string]struct{}) {
		switch k := e.GetExprKind().(type) {
		case *exprv1.Expr_IdentExpr:
			if _, ok := bound[k.IdentExpr.GetName()]; ok {
				ret[e.GetId()] = struct{}{}
			}
		case *exprv1.Expr_SelectExpr:
			walk(k.SelectExpr.GetOperand(), bound)
		case *exprv1.Expr_CallExpr:
			walk(k.CallExpr.GetTarget(), bound)
			walkAll(k.CallExpr.GetArgs(), bound)
		case *exprv1.Expr_ListExpr:
			walkAll(k.ListExpr.GetElements(), bound)
		case *exprv1.Expr_StructExpr:
			ret[e.GetId()] = struct{}{}
			for _, entry := range k.StructExpr.GetEntries() {
				if key := entry.GetMapKey(); key != nil {
					walk(key, bound)
				}
				walk(entry.GetValue(), bound)
			}
		case *exprv1.Expr_ComprehensionExpr:
			c := k.ComprehensionExpr
			walk(c.GetIterRange(), bound)
			walk(c.GetAccuInit(), bound)
			inner := map[string]struct{}{}
			for name := range bound {
				inner[name] = struct{}{}
			}
			for _, name := range []string{c.GetIterVar(), c.GetIterVar2(), c.GetAccuVar()} {
				if name != "" {
					inner[name] = struct{}{}
				}
			}
			walk(c.GetLoopCondition(), inner)
			walk(c.GetLoopStep(), inner)
			walk(c.GetResult(), inner)
		}
	}
	walk(root, map[string]struct{}{})
	return ret
}

func isServiceWideCELVariable(name string) bool {
	switch name {
	case "error", grpcfed.ContextVariableName, "grpc.federation.env", "grpc.federation.var":
		return true
	}
	return false
}

// celIndexer assigns an index to every CEL expression of a file, which is the key of the program cache
// of the generated service, and keeps the expressions of each service for precompilation.
// Expressions are identified by pointer, so identical text in different scopes gets different indexes.
type celIndexer struct {
	next       int
	values     map[*resolver.CELValue]int
	names      map[string]int
	svcEntries map[*resolver.Service][]*CELPrecompileEntry
}

func newCELIndexer(file *resolver.File) (*celIndexer, error) {
	idx := &celIndexer{
		values:     make(map[*resolver.CELValue]int),
		names:      make(map[string]int),
		svcEntries: make(map[*resolver.Service][]*CELPrecompileEntry),
	}
	for _, svc := range file.Services {
		w := &celWalker{indexer: idx, seen: make(map[int]struct{})}
		w.service(svc)
		if w.err != nil {
			return nil, w.err
		}
		sort.Slice(w.entries, func(i, j int) bool { return w.entries[i].Index < w.entries[j].Index })
		idx.svcEntries[svc] = w.entries
	}
	return idx, nil
}

func (i *celIndexer) valueIndex(v *resolver.CELValue) (int, error) {
	idx, exists := i.values[v]
	if !exists {
		return 0, fmt.Errorf("grpc-federation: CEL expression %q is not reachable from any service", v.Expr)
	}
	return idx, nil
}

// nameIndex returns the index of the expression that evaluates a variable by its name.
func (i *celIndexer) nameIndex(def *resolver.VariableDefinition) (int, error) {
	if def == nil || def.Expr == nil || def.Expr.Type == nil {
		return 0, fmt.Errorf("grpc-federation: variable has no type to evaluate by name")
	}
	idx, exists := i.names[nameKey(def)]
	if !exists {
		return 0, fmt.Errorf("grpc-federation: variable %q is not reachable from any service", def.Name)
	}
	return idx, nil
}

func nameKey(def *resolver.VariableDefinition) string {
	return def.Name + "|" + toCELTypeDeclare(def.Expr.Type)
}

type celWalker struct {
	indexer *celIndexer
	entries []*CELPrecompileEntry
	seen    map[int]struct{}
	err     error
}

func (w *celWalker) add(idx int, expr string, vars []*CELPrecompileVariable) {
	if _, exists := w.seen[idx]; exists {
		return
	}
	w.seen[idx] = struct{}{}
	w.entries = append(w.entries, &CELPrecompileEntry{Index: idx, Expr: expr, Variables: vars})
}

func (w *celWalker) value(v *resolver.CELValue) {
	if v == nil || w.err != nil {
		return
	}
	idx, exists := w.indexer.values[v]
	if !exists {
		w.indexer.next++
		idx = w.indexer.next
		w.indexer.values[v] = idx
	}
	if _, exists := w.seen[idx]; exists {
		return
	}
	variables, err := referencedCELVariables(v.CheckedExpr)
	if err != nil {
		w.err = err
		return
	}
	vars := make([]*CELPrecompileVariable, 0, len(variables))
	for _, variable := range variables {
		celType, err := celTypeToNativeType(variable.Type)
		if err != nil {
			w.err = fmt.Errorf("grpc-federation: variable %q of CEL expression %q: %w", variable.Name, v.Expr, err)
			return
		}
		vars = append(vars, &CELPrecompileVariable{Name: variable.Name, CELType: celType})
	}
	w.add(idx, v.Expr, vars)
}

// name registers the expression that evaluates the variable definition by its name (see grpcfed.CustomMessage).
func (w *celWalker) name(def *resolver.VariableDefinition) {
	if def == nil || def.Name == "" || def.Expr == nil || def.Expr.Type == nil || w.err != nil {
		return
	}
	key := nameKey(def)
	idx, exists := w.indexer.names[key]
	if !exists {
		w.indexer.next++
		idx = w.indexer.next
		w.indexer.names[key] = idx
	}
	w.add(idx, def.Name, []*CELPrecompileVariable{{Name: def.Name, CELType: toCELTypeDeclare(def.Expr.Type)}})
}

func (w *celWalker) service(svc *resolver.Service) {
	if svc.Rule != nil {
		for _, svcVar := range svc.Rule.Vars {
			w.value(svcVar.If)
			w.serviceVariableExpr(svcVar.Expr)
		}
	}
	for _, msg := range svc.Messages {
		w.message(msg)
	}
}

func (w *celWalker) serviceVariableExpr(expr *resolver.ServiceVariableExpr) {
	if expr == nil {
		return
	}
	w.value(expr.By)
	w.mapExpr(expr.Map)
	w.messageExpr(expr.Message)
	w.enumExpr(expr.Enum)
	w.switchExpr(expr.Switch)
	if expr.Validation != nil {
		w.value(expr.Validation.If)
		w.value(expr.Validation.Message)
	}
}

func (w *celWalker) message(msg *resolver.Message) {
	if msg.Rule != nil {
		w.defSet(msg.Rule.DefSet)
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Value != nil {
			w.value(field.Rule.Value.CEL)
		}
		if oneof := field.Rule.Oneof; oneof != nil {
			w.value(oneof.If)
			w.value(oneof.By)
			w.defSet(oneof.DefSet)
		}
	}
}

func (w *celWalker) defSet(set *resolver.VariableDefinitionSet) {
	if set == nil {
		return
	}
	for _, def := range set.Definitions() {
		w.def(def)
	}
}

func (w *celWalker) def(def *resolver.VariableDefinition) {
	if def == nil {
		return
	}
	w.value(def.If)
	expr := def.Expr
	if expr == nil {
		return
	}
	w.value(expr.By)
	w.mapExpr(expr.Map)
	w.callExpr(expr.Call)
	w.messageExpr(expr.Message)
	w.enumExpr(expr.Enum)
	w.switchExpr(expr.Switch)
	if expr.Validation != nil {
		w.grpcError(expr.Validation.Error)
	}
}

func (w *celWalker) mapExpr(expr *resolver.MapExpr) {
	if expr == nil || expr.Expr == nil {
		return
	}
	w.value(expr.Expr.By)
	w.messageExpr(expr.Expr.Message)
	w.enumExpr(expr.Expr.Enum)
}

func (w *celWalker) callExpr(expr *resolver.CallExpr) {
	if expr == nil {
		return
	}
	if expr.Request != nil {
		w.arguments(expr.Request.Args)
	}
	if expr.Retry != nil {
		w.value(expr.Retry.If)
	}
	if expr.Option != nil {
		w.def(expr.Option.Header)
		w.def(expr.Option.Trailer)
	}
	w.value(expr.Metadata)
	for _, grpcErr := range expr.Errors {
		w.grpcError(grpcErr)
	}
}

func (w *celWalker) messageExpr(expr *resolver.MessageExpr) {
	if expr == nil {
		return
	}
	w.arguments(expr.Args)
}

func (w *celWalker) arguments(args []*resolver.Argument) {
	for _, arg := range args {
		w.value(arg.If)
		if arg.Value != nil {
			w.value(arg.Value.CEL)
		}
	}
}

func (w *celWalker) enumExpr(expr *resolver.EnumExpr) {
	if expr == nil {
		return
	}
	w.value(expr.By)
}

func (w *celWalker) switchExpr(expr *resolver.SwitchExpr) {
	if expr == nil {
		return
	}
	for _, c := range expr.Cases {
		w.defSet(c.DefSet)
		w.value(c.If)
		w.value(c.By)
	}
	if expr.Default != nil {
		w.defSet(expr.Default.DefSet)
		w.value(expr.Default.By)
	}
}

func (w *celWalker) grpcError(grpcErr *resolver.GRPCError) {
	if grpcErr == nil {
		return
	}
	w.defSet(grpcErr.DefSet)
	w.value(grpcErr.If)
	w.value(grpcErr.Message)
	w.value(grpcErr.IgnoreAndResponse)
	for _, detail := range grpcErr.Details {
		w.value(detail.If)
		w.defSet(detail.DefSet)
		for _, by := range detail.By {
			w.value(by)
		}
		w.defSet(detail.Messages)
		if detail.Messages != nil {
			for _, def := range detail.Messages.Definitions() {
				w.name(def)
			}
		}
		for _, failure := range detail.PreconditionFailures {
			for _, violation := range failure.Violations {
				w.value(violation.Type)
				w.value(violation.Subject)
				w.value(violation.Description)
			}
		}
		for _, badRequest := range detail.BadRequests {
			for _, violation := range badRequest.FieldViolations {
				w.value(violation.Field)
				w.value(violation.Description)
			}
		}
		for _, msg := range detail.LocalizedMessages {
			w.value(msg.Message)
		}
	}
}
