package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/cel-go/cel"
	celtypes "github.com/google/cel-go/common/types"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	grpcfedcel "github.com/mercari/grpc-federation/grpc/federation/cel"
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

// celTypeToNativeType renders the Go source that declares the given CEL type in the runtime env.
// It must produce the same type as toCELNativeType does for the variable's declaration.
func celTypeToNativeType(t *cel.Type) (string, error) {
	if name := t.DeclaredTypeName(); strings.HasPrefix(name, "wrapper(") {
		wrappers := map[string]string{
			"wrapper(bool)":   "google.protobuf.BoolValue",
			"wrapper(bytes)":  "google.protobuf.BytesValue",
			"wrapper(double)": "google.protobuf.DoubleValue",
			"wrapper(int)":    "google.protobuf.Int64Value",
			"wrapper(string)": "google.protobuf.StringValue",
			"wrapper(uint)":   "google.protobuf.UInt64Value",
		}
		if msg, ok := wrappers[name]; ok {
			return fmt.Sprintf("grpcfed.CELObjectType(%q)", msg), nil
		}
	}
	switch t.Kind() {
	case celtypes.BoolKind:
		return "grpcfed.CELBoolType", nil
	case celtypes.BytesKind:
		return "grpcfed.CELBytesType", nil
	case celtypes.DoubleKind:
		return "grpcfed.CELDoubleType", nil
	case celtypes.IntKind:
		return "grpcfed.CELIntType", nil
	case celtypes.UintKind:
		return "grpcfed.CELUintType", nil
	case celtypes.StringKind:
		return "grpcfed.CELStringType", nil
	case celtypes.NullTypeKind:
		return "grpcfed.CELNullType", nil
	case celtypes.AnyKind:
		return `grpcfed.CELObjectType("google.protobuf.Any")`, nil
	case celtypes.DurationKind:
		return `grpcfed.CELObjectType("google.protobuf.Duration")`, nil
	case celtypes.TimestampKind:
		return `grpcfed.CELObjectType("google.protobuf.Timestamp")`, nil
	case celtypes.StructKind:
		return fmt.Sprintf("grpcfed.CELObjectType(%q)", t.TypeName()), nil
	case celtypes.OpaqueKind:
		// The runtime env declares an enum selector as a plain object type.
		if t.TypeName() == grpcfedcel.EnumSelectorFQDN {
			return fmt.Sprintf("grpcfed.CELObjectType(%q)", t.TypeName()), nil
		}
		// An enum is declared as an opaque type with an int parameter, but the runtime env declares it as int.
		if params := t.Parameters(); len(params) == 1 && params[0].Kind() == celtypes.IntKind && t.TypeName() != "optional_type" {
			return "grpcfed.CELIntType", nil
		}
	case celtypes.ListKind:
		elem, err := celTypeToNativeType(t.Parameters()[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("grpcfed.CELListType(%s)", elem), nil
	case celtypes.MapKind:
		key, err := celTypeToNativeType(t.Parameters()[0])
		if err != nil {
			return "", err
		}
		value, err := celTypeToNativeType(t.Parameters()[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("grpcfed.NewCELMapType(%s, %s)", key, value), nil
	}
	return "", fmt.Errorf("unsupported CEL type %s", t)
}

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
