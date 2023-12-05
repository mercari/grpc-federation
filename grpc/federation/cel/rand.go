package cel

import (
	"math/rand"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

const RandPackageName = "rand"

var (
	SourceType = types.NewObjectType(createRandName("Source"))
	RandType   = types.NewObjectType(createRandName("Rand"))
)

type Source struct {
	rand.Source
}

func (s *Source) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return s.Source, nil
}

func (s *Source) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("grpc.federation.rand: source type conversion does not support")
}

func (s *Source) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Source); ok {
		return types.Bool(s.Int63() == o.Int63())
	}
	return types.False
}

func (s *Source) Type() ref.Type {
	return SourceType
}

func (s *Source) Value() any {
	return s.Source
}

type Rand struct {
	*rand.Rand
}

func (r *Rand) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return r.Rand, nil
}

func (r *Rand) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("grpc.federation.rand: rand type conversion does not support")
}

func (r *Rand) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Rand); ok {
		return types.Bool(r.Int() == o.Int())
	}
	return types.False
}

func (r *Rand) Type() ref.Type {
	return RandType
}

func (r *Rand) Value() any {
	return r.Rand
}

type RandLibrary struct {
}

func (lib *RandLibrary) LibraryName() string {
	return packageName(RandPackageName)
}

func createRandName(name string) string {
	return createName(RandPackageName, name)
}

func createRandID(name string) string {
	return createID(RandPackageName, name)
}

func (lib *RandLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		// Source functions
		cel.Function(
			createRandName("newSource"),
			cel.Overload(createRandID("new_source_int_source"), []*cel.Type{cel.IntType}, SourceType,
				cel.UnaryBinding(func(seed ref.Val) ref.Val {
					return &Source{
						Source: rand.NewSource(int64(seed.(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			"int63",
			cel.MemberOverload(createRandID("int63_source_int"), []*cel.Type{SourceType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(*Source).Int63())
				}),
			),
		),
		cel.Function(
			"seed",
			cel.MemberOverload(createRandID("seed_source_int"), []*cel.Type{SourceType, cel.IntType}, cel.BoolType,
				cel.BinaryBinding(func(self, seed ref.Val) ref.Val {
					self.(*Source).Seed(int64(seed.(types.Int)))
					return types.Bool(types.True)
				}),
			),
		),

		// Rand functions
		cel.Function(
			createRandName("new"),
			cel.Overload(createRandID("new_source_rand"), []*cel.Type{SourceType}, RandType,
				cel.UnaryBinding(func(source ref.Val) ref.Val {
					return &Rand{
						Rand: rand.New(source.(*Source)),
					}
				}),
			),
		),
		cel.Function(
			"expFloat64",
			cel.MemberOverload(createRandID("exp_float64_rand_double"), []*cel.Type{RandType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(*Rand).ExpFloat64())
				}),
			),
		),
		cel.Function(
			"float32",
			cel.MemberOverload(createRandID("float32_rand_double"), []*cel.Type{RandType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(*Rand).Float32())
				}),
			),
		),
		cel.Function(
			"float64",
			cel.MemberOverload(createRandID("float64_rand_double"), []*cel.Type{RandType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(*Rand).Float64())
				}),
			),
		),
		cel.Function(
			"int",
			cel.MemberOverload(createRandID("int_rand_int"), []*cel.Type{RandType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int())
				}),
			),
		),
		cel.Function(
			"int31",
			cel.MemberOverload(createRandID("int31_rand_int"), []*cel.Type{RandType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int31())
				}),
			),
		),
		cel.Function(
			"int31n",
			cel.MemberOverload(createRandID("int31n_rand_int"), []*cel.Type{RandType, cel.IntType}, cel.IntType,
				cel.BinaryBinding(func(self, n ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int31n(int32(n.(types.Int))))
				}),
			),
		),
		cel.Function(
			"int63",
			cel.MemberOverload(createRandID("int63_rand_int"), []*cel.Type{RandType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int63())
				}),
			),
		),
		cel.Function(
			"int63n",
			cel.MemberOverload(createRandID("int63n_rand_int"), []*cel.Type{RandType, cel.IntType}, cel.IntType,
				cel.BinaryBinding(func(self, n ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int63n(int64(n.(types.Int))))
				}),
			),
		),
		cel.Function(
			"intn",
			cel.MemberOverload(createRandID("intn_rand_int"), []*cel.Type{RandType, cel.IntType}, cel.IntType,
				cel.BinaryBinding(func(self, n ref.Val) ref.Val {
					return types.Int(self.(*Rand).Intn(int(n.(types.Int))))
				}),
			),
		),
		cel.Function(
			"normFloat64",
			cel.MemberOverload(createRandID("norm_float64_rand_double"), []*cel.Type{RandType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(*Rand).NormFloat64())
				}),
			),
		),
		cel.Function(
			"seed",
			cel.MemberOverload(createRandID("seed_rand_int_bool"), []*cel.Type{RandType, cel.IntType}, cel.BoolType,
				cel.BinaryBinding(func(self, seed ref.Val) ref.Val {
					self.(*Rand).Seed(int64(seed.(types.Int)))
					return types.Bool(types.True)
				}),
			),
		),
		cel.Function(
			"uint32",
			cel.MemberOverload(createRandID("uint32_rand_uint"), []*cel.Type{RandType}, cel.UintType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Uint(self.(*Rand).Uint32())
				}),
			),
		),
	}
	return opts
}

func (lib *RandLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
