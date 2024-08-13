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

type LogLibrary struct{}

func NewLogLibrary() *LogLibrary {
	return &LogLibrary{}
}

func (lib *LogLibrary) LibraryName() string {
	return packageName(LogPackageName)
}

func (lib *LogLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.OptionalTypes(),
	}
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			logInfoFunc,
			OverloadFunc("grpc_federation_log_info",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				lib.info,
			),
			OverloadFunc("grpc_federation_log_info_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				lib.info,
			),
		),
		BindFunction(
			logDebugFunc,
			OverloadFunc("grpc_federation_log_debug",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				lib.debug,
			),
			OverloadFunc("grpc_federation_log_debug_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				lib.debug,
			),
		),
		BindFunction(
			logWarnFunc,
			OverloadFunc("grpc_federation_log_warn",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				lib.warn,
			),
			OverloadFunc("grpc_federation_log_warn_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				lib.warn,
			),
		),
		BindFunction(
			logErrorFunc,
			OverloadFunc("grpc_federation_log_error",
				[]*cel.Type{cel.StringType}, cel.BoolType,
				lib.error,
			),
			OverloadFunc("grpc_federation_log_error_args",
				[]*cel.Type{cel.StringType, cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				lib.error,
			),
		),
		BindFunction(
			logAddFunc,
			OverloadFunc("grpc_federation_log_add",
				[]*cel.Type{cel.MapType(cel.StringType, cel.DynType)}, cel.BoolType,
				lib.add,
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *LogLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func (lib *LogLibrary) debug(ctx context.Context, values ...ref.Val) ref.Val {
	logger := log.Logger(ctx)
	return lib.log(ctx, logger.DebugContext, values...)
}

func (lib *LogLibrary) info(ctx context.Context, values ...ref.Val) ref.Val {
	logger := log.Logger(ctx)
	return lib.log(ctx, logger.InfoContext, values...)
}

func (lib *LogLibrary) warn(ctx context.Context, values ...ref.Val) ref.Val {
	logger := log.Logger(ctx)
	return lib.log(ctx, logger.WarnContext, values...)
}

func (lib *LogLibrary) error(ctx context.Context, values ...ref.Val) ref.Val {
	logger := log.Logger(ctx)
	return lib.log(ctx, logger.ErrorContext, values...)
}

func (lib *LogLibrary) log(ctx context.Context, logFn func(ctx context.Context, msg string, args ...any), values ...ref.Val) ref.Val {
	switch len(values) {
	case 1:
		msg := values[0]
		logFn(ctx, fmt.Sprint(msg.Value()))
	case 2:
		msg := values[0]
		args := values[1].(traits.Mapper)

		var attrs []any
		for it := args.Iterator(); it.HasNext() == types.True; {
			k := it.Next()
			v := args.Get(k)
			attrs = append(attrs, slog.Any(fmt.Sprint(k.Value()), lib.expandVal(v)))
		}
		logFn(ctx, fmt.Sprint(msg.Value()), attrs...)
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

func (lib *LogLibrary) add(ctx context.Context, values ...ref.Val) ref.Val {
	value := values[0]
	args := value.(traits.Mapper)
	var attrs []slog.Attr
	for it := args.Iterator(); it.HasNext() == types.True; {
		k := it.Next()
		v := args.Get(k)
		attrs = append(attrs, slog.Any(fmt.Sprint(k.Value()), lib.expandVal(v)))
	}

	log.AddAttrs(ctx, attrs)
	return types.True
}
