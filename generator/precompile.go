package generator

import (
	"fmt"
	"sort"

	"github.com/mercari/grpc-federation/resolver"
)

// This file generates the table of CEL expressions that a service compiles ahead of the first request.
// Every expression gets an index, the key of the program cache of the generated service,
// by walking the service once. The templates look the index up instead of issuing it,
// so an expression the walk does not reach fails the generation.

// CELPrecompileEntry is a CEL expression compiled by the generated precompileCEL method.
type CELPrecompileEntry struct {
	Index     int
	Expr      string
	Variables []*CELPrecompileVariable
}

// CELPrecompileVariable is a variable referenced by a precompiled expression.
type CELPrecompileVariable struct {
	Name    string
	CELType string
}

// CELPrecompileEntries returns every CEL expression used by the service, ordered by index.
func (s *Service) CELPrecompileEntries() ([]*CELPrecompileEntry, error) {
	indexer, err := s.file.celIndexer()
	if err != nil {
		return nil, err
	}
	return indexer.svcEntries[s.Service], nil
}

// celIndexer returns the CEL expression indexes of the file. They are computed once per File,
// so rendering the file template more than once gives the same indexes.
func (f *File) celIndexer() (*celIndexer, error) {
	if f.indexer != nil {
		return f.indexer, nil
	}
	indexer, err := newCELIndexer(f.File)
	if err != nil {
		return nil, err
	}
	f.indexer = indexer
	return indexer, nil
}

// cacheIndex is the template function that returns the program cache index of the CEL expression.
func (f *File) cacheIndex(v *resolver.CELValue) (int, error) {
	indexer, err := f.celIndexer()
	if err != nil {
		return 0, err
	}
	return indexer.valueIndex(v)
}

// nameCacheIndex is the template function that returns the program cache index of the expression
// that evaluates a variable definition by its name.
func (f *File) nameCacheIndex(def any) (int, error) {
	indexer, err := f.celIndexer()
	if err != nil {
		return 0, err
	}
	switch d := def.(type) {
	case *VariableDefinition:
		return indexer.nameIndex(d.VariableDefinition)
	case *resolver.VariableDefinition:
		return indexer.nameIndex(d)
	}
	return 0, fmt.Errorf("grpc-federation: unexpected variable definition type %T", def)
}

// celIndexer assigns an index to every CEL expression of a file, which is the key of the program cache
// of the generated service, and keeps the expressions of each service for precompilation.
// Expressions are identified by pointer, so identical text in different scopes gets different indexes.
type celIndexer struct {
	next   int
	values map[*resolver.CELValue]int
	// names holds the indexes of the expressions that evaluate a variable by its name (see nameCacheIndex).
	//
	// The message blocks of an error detail are resolved to variable definitions named like
	// "_def0_err_detail0_msg0" (see resolveGRPCDetailMessages). The generated code evaluates such a definition
	// with evalDefSet, which stores the result in the local value, and then reads it back with grpcfed.CustomMessage.
	// CustomMessage does this by evaluating the CEL expression "_def0_err_detail0_msg0", with only that
	// variable declared. This expression is written by the template, not by the user, so the resolver never
	// creates a *resolver.CELValue for it and it cannot be looked up by pointer as values can.
	//
	// The compiled program depends only on the expression text (the variable name) and the declared type
	// of the variable, so the key is exactly those two (see nameKey). Definitions with the same key share one
	// index and one precompile entry. Definitions that differ in either get separate ones. The key does not need to
	// identify a definition, and it is not unique per definition: the names are built from counters of the
	// enclosing context, so different messages can have a definition with the same name and a different type.
	// The type is rendered with toCELTypeDeclare, which is what the lazy path declares, so a type that maps to the
	// same declaration (for example an enum and an int64) correctly shares a program.
	//
	// TODO: CustomMessage could read the variable from the value directly instead of evaluating CEL. Then this map,
	// nameKey, celWalker.name and CustomMessageParam.CacheIndex would not be needed.
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
	// The compiler fills in CheckedExpr, and Variables with it, when it resolves the expression.
	if v.CheckedExpr == nil {
		w.err = fmt.Errorf("grpc-federation: CEL expression %q was not resolved by the compiler", v.Expr)
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
	// Variables are the user-defined variables the compiler recorded as visible to the expression.
	vars := make([]*CELPrecompileVariable, 0, len(v.Variables))
	for _, variable := range v.Variables {
		if variable == nil || variable.Name == "" || variable.Type == nil {
			continue
		}
		vars = append(vars, &CELPrecompileVariable{Name: variable.Name, CELType: toCELNativeType(variable.Type)})
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
