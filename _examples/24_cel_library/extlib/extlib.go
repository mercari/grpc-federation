package extlib

import (
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func NewLibrary() cel.SingletonLibrary {
	return &library{}
}

type library struct{}

func (library) LibraryName() string { return "example.ext" }

func (library) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("example.ext.upper",
			cel.Overload(
				"example_ext_upper_string_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					s, ok := arg.Value().(string)
					if !ok {
						return types.NewErr("example.ext.upper: expected string, got %T", arg.Value())
					}
					return types.String(strings.ToUpper(s))
				}),
			),
		),
	}
}

func (library) ProgramOptions() []cel.ProgramOption { return nil }
