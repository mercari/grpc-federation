package cel

import (
	"context"
	"math"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const MathPackageName = "math"

var _ cel.SingletonLibrary = new(MathLibrary)

type MathLibrary struct {
}

func NewMathLibrary() *MathLibrary {
	return &MathLibrary{}
}

func (lib *MathLibrary) LibraryName() string {
	return packageName(MathPackageName)
}

func createMathName(name string) string {
	return createName(MathPackageName, name)
}

func createMathID(name string) string {
	return createID(MathPackageName, name)
}

func (lib *MathLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{}

	for _, funcOpts := range [][]cel.EnvOption{
		// math package functions
		BindFunction(
			createMathName("sqrt"),
			OverloadFunc(createMathID("sqrt_double_double"), []*cel.Type{cel.DoubleType}, cel.DoubleType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Double(math.Sqrt(args[0].(types.Double).Value().(float64)))
				},
			),
		),
		BindFunction(
			createMathName("pow"),
			OverloadFunc(createMathID("pow_double_double_double"), []*cel.Type{cel.DoubleType, cel.DoubleType}, cel.DoubleType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Double(math.Pow(args[0].(types.Double).Value().(float64), args[1].(types.Double).Value().(float64)))
				},
			),
		),
		BindFunction(
			createMathName("floor"),
			OverloadFunc(createMathID("floor_double_double"), []*cel.Type{cel.DoubleType}, cel.DoubleType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Double(math.Floor(args[0].(types.Double).Value().(float64)))
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}

	return opts
}

func (lib *MathLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
