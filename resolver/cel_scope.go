package resolver

import (
	"slices"

	"github.com/google/cel-go/cel"
)

// celScope is a CEL env together with the user-defined variables declared in it.
// cel.Env does not expose the variables declared in it, so the scope records them
// to attach the variables visible to each CEL expression to its CELValue.
// A celScope is immutable: extending it returns a new one.
type celScope struct {
	env  *cel.Env
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
// (like "error"), so a precompiled expression must not declare them again.
func (s *celScope) extendEnv(opts ...cel.EnvOption) (*celScope, error) {
	env, err := s.env.Extend(opts...)
	if err != nil {
		return nil, err
	}
	return &celScope{env: env, vars: s.vars}, nil
}
