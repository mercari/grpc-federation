package cel

import (
	"context"
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
	return s, nil
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
	return s
}

type Rand struct {
	*rand.Rand
}

func (r *Rand) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return r, nil
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
	return r
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
	var opts []cel.EnvOption
	for _, funcOpts := range [][]cel.EnvOption{
		// Source functions
		BindFunction(
			createRandName("newSource"),
			OverloadFunc(createRandID("new_source_int_source"), []*cel.Type{cel.IntType}, SourceType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return &Source{
						Source: rand.NewSource(int64(args[0].(types.Int))),
					}
				},
			),
		),
		BindMemberFunction(
			"int63",
			MemberOverloadFunc(createRandID("int63_source_int"), SourceType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Source).Int63())
				},
			),
		),
		BindMemberFunction(
			"seed",
			MemberOverloadFunc(createRandID("seed_source_int"), SourceType, []*cel.Type{cel.IntType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					self.(*Source).Seed(int64(args[0].(types.Int)))
					return types.True
				},
			),
		),

		// Rand functions
		BindFunction(
			createRandName("new"),
			OverloadFunc(createRandID("new_source_rand"), []*cel.Type{SourceType}, RandType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return &Rand{
						Rand: rand.New(args[0].(*Source)), //nolint:gosec
					}
				},
			),
		),
		BindMemberFunction(
			"expFloat64",
			MemberOverloadFunc(createRandID("exp_float64_rand_double"), RandType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(*Rand).ExpFloat64())
				},
			),
		),
		BindMemberFunction(
			"float32",
			MemberOverloadFunc(createRandID("float32_rand_double"), RandType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(*Rand).Float32())
				},
			),
		),
		BindMemberFunction(
			"float64",
			MemberOverloadFunc(createRandID("float64_rand_double"), RandType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(*Rand).Float64())
				},
			),
		),
		BindMemberFunction(
			"int",
			MemberOverloadFunc(createRandID("int_rand_int"), RandType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int())
				},
			),
		),
		BindMemberFunction(
			"int31",
			MemberOverloadFunc(createRandID("int31_rand_int"), RandType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int31())
				},
			),
		),
		BindMemberFunction(
			"int31n",
			MemberOverloadFunc(createRandID("int31n_rand_int"), RandType, []*cel.Type{cel.IntType}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int31n(int32(args[0].(types.Int)))) //nolint:gosec
				},
			),
		),
		BindMemberFunction(
			"int63",
			MemberOverloadFunc(createRandID("int63_rand_int"), RandType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int63())
				},
			),
		),
		BindMemberFunction(
			"int63n",
			MemberOverloadFunc(createRandID("int63n_rand_int"), RandType, []*cel.Type{cel.IntType}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Int63n(int64(args[0].(types.Int))))
				},
			),
		),
		BindMemberFunction(
			"intn",
			MemberOverloadFunc(createRandID("intn_rand_int"), RandType, []*cel.Type{cel.IntType}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(*Rand).Intn(int(args[0].(types.Int))))
				},
			),
		),
		BindMemberFunction(
			"normFloat64",
			MemberOverloadFunc(createRandID("norm_float64_rand_double"), RandType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(*Rand).NormFloat64())
				},
			),
		),
		BindMemberFunction(
			"seed",
			MemberOverloadFunc(createRandID("seed_rand_int_bool"), RandType, []*cel.Type{cel.IntType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					self.(*Rand).Seed(int64(args[0].(types.Int)))
					return types.True
				},
			),
		),
		BindMemberFunction(
			"uint32",
			MemberOverloadFunc(createRandID("uint32_rand_uint"), RandType, []*cel.Type{}, cel.UintType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Uint(self.(*Rand).Uint32())
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *RandLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
