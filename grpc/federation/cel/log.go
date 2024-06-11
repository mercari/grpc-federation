package cel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"

	"github.com/mercari/grpc-federation/grpc/federation/log"
)

const LogPackageName = "log"

const (
	logDebugFunc = "grpc.federation.log.debug"
	logInfoFunc  = "grpc.federation.log.info"
	logWarnFunc  = "grpc.federation.log.warn"
	logErrorFunc = "grpc.federation.log.error"
	logAddFunc   = "grpc.federation.log.add"
)

type LogLibrary struct {
	ctx context.Context
}

func NewLogLibrary() *LogLibrary {
	return &LogLibrary{}
}

func (lib *LogLibrary) LibraryName() string {
	return packageName(LogPackageName)
}

func (lib *LogLibrary) Initialize(ctx context.Context) {
	lib.ctx = ctx
}

func (lib *LogLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
		cel.Function(logDebugFunc,
			cel.Overload("grpc_federation_log_debug",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				cel.FunctionBinding(lib.debug),
			),
			cel.Overload("grpc_federation_log_debug_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				cel.FunctionBinding(lib.debug),
			),
		),
		cel.Function(logInfoFunc,
			cel.Overload("grpc_federation_log_info",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				cel.FunctionBinding(lib.info),
			),
			cel.Overload("grpc_federation_log_info_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				cel.FunctionBinding(lib.info),
			),
		),
		cel.Function(logWarnFunc,
			cel.Overload("grpc_federation_log_warn",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				cel.FunctionBinding(lib.warn),
			),
			cel.Overload("grpc_federation_log_warn_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				cel.FunctionBinding(lib.warn),
			),
		),
		cel.Function(logErrorFunc,
			cel.Overload("grpc_federation_log_error",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				cel.FunctionBinding(lib.error),
			),
			cel.Overload("grpc_federation_log_error_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				cel.FunctionBinding(lib.error),
			),
		),
		cel.Function(logAddFunc,
			cel.Overload("grpc_federation_log_add",
				[]*cel.Type{cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				cel.UnaryBinding(lib.add),
			),
		),
	}
	return opts
}

func (lib *LogLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (lib *LogLibrary) debug(values ...ref.Val) ref.Val {
	logger := log.Logger(lib.ctx)
	return lib.log(logger.DebugContext, values...)
}

func (lib *LogLibrary) info(values ...ref.Val) ref.Val {
	logger := log.Logger(lib.ctx)
	return lib.log(logger.InfoContext, values...)
}

func (lib *LogLibrary) warn(values ...ref.Val) ref.Val {
	logger := log.Logger(lib.ctx)
	return lib.log(logger.WarnContext, values...)
}

func (lib *LogLibrary) error(values ...ref.Val) ref.Val {
	logger := log.Logger(lib.ctx)
	return lib.log(logger.ErrorContext, values...)
}

func (lib *LogLibrary) log(logFn func(ctx context.Context, msg string, args ...any), values ...ref.Val) ref.Val {
	switch len(values) {
	case 1:
		msg := values[0]
		logFn(lib.ctx, fmt.Sprint(msg.Value()))
	case 2:
		msg := values[0]
		args := values[1].(traits.Mapper)

		var attrs []any
		for it := args.Iterator(); it.HasNext() == types.True; {
			k := it.Next()
			v := args.Get(k)
			attrs = append(attrs, slog.Any(fmt.Sprint(k.Value()), lib.expandVal(v)))
		}
		logFn(lib.ctx, fmt.Sprint(msg.Value()), attrs...)
	}
	return types.True
}

func (lib *LogLibrary) expandVal(arg ref.Val) any {
	// Expand the lister and mapper type of value to output the actual value in the log
	switch typ := arg.(type) {
	case traits.Mapper:
		o := map[string]any{}
		for it := typ.Iterator(); it.HasNext() == types.True; {
			k := it.Next()
			v := typ.Get(k)
			o[fmt.Sprint(k.Value())] = lib.expandVal(v)
		}
		return o
	case traits.Lister:
		var o []any
		for it := typ.Iterator(); it.HasNext() == types.True; {
			v := it.Next()
			o = append(o, lib.expandVal(v))
		}
		return o
	}
	return arg.Value()
}

func (lib *LogLibrary) add(value ref.Val) ref.Val {
	args := value.(traits.Mapper)
	var attrs []any
	for it := args.Iterator(); it.HasNext() == types.True; {
		k := it.Next()
		v := args.Get(k)
		attrs = append(attrs, slog.Any(fmt.Sprint(k.Value()), lib.expandVal(v)))
	}

	logger := log.Logger(lib.ctx)
	log.SetLogger(lib.ctx, logger.With(attrs...))
	return types.True
}
