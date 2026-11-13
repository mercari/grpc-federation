package resolver

import (
	"slices"

	"github.com/google/cel-go/cel"
	exprv1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// celScope is a CEL env together with the variables declared in it.
// cel.Env does not expose the variables declared in it, so the scope records them
// to attach them to the CELValue of each expression compiled in it.
// A celScope is immutable: extending it returns a new one.
type celScope struct {
	env *cel.Env
	// vars are the variables declared in this scope on top of the env that the runtime starts every
	// expression from: the message argument and the variables defined, or iterated over, before this point.
	// The runtime already declares ctx, error, grpc.federation.env and grpc.federation.var in every env,
	// so they are not included.
	// The scope keeps all of them because a later expression may reference any of them.
	vars []*CELVariable
}

// extend declares the variable in a new scope.
func (s *celScope) extend(name string, typ *Type) (*celScope, error) {
	env, err := s.env.Extend(cel.Variable(name, ToCELType(typ)))
	if err != nil {
		return nil, err
	}
	return &celScope{
		env:  env,
		vars: append(slices.Clone(s.vars), &CELVariable{Name: name, Type: typ}),
	}, nil
}

// extendEnv adds the options to a new scope without recording the variables declared by them.
// It is for the variables that the runtime declares in every env regardless of the scope
// (like "error").
func (s *celScope) extendEnv(opts ...cel.EnvOption) (*celScope, error) {
	env, err := s.env.Extend(opts...)
	if err != nil {
		return nil, err
	}
	return &celScope{env: env, vars: s.vars}, nil
}

// referenced returns the variables of the scope that the checked expression references, in declaration order.
// The scope itself is not changed, so the variables the expression does not reference stay visible to later expressions.
//
// A CELValue gets only these, not all the variables of the scope. The scope grows with every definition above
// the expression, so keeping all of them on every CELValue would make the lists quadratic in the number of pointers
// for a message with many definitions. The only use of the list is the generator's CEL expression precompilation,
// which only needs the variables that the expression references to be declared, not all that are in scope.
//
// The reference map can contain names that are not variables, such as comprehension variables.
// That is harmless because they are only used to select from the variables of the scope.
func (s *celScope) referenced(checked *exprv1.CheckedExpr) []*CELVariable {
	names := make(map[string]struct{}, len(checked.GetReferenceMap()))
	for _, ref := range checked.GetReferenceMap() {
		// An identifier reference has neither overloads nor a constant value (enum values).
		if len(ref.GetOverloadId()) != 0 || ref.GetValue() != nil {
			continue
		}
		names[ref.GetName()] = struct{}{}
	}
	var ret []*CELVariable
	for _, v := range s.vars {
		if _, exists := names[v.Name]; exists {
			ret = append(ret, v)
		}
	}
	return ret
}
