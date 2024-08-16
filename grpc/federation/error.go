package federation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"golang.org/x/sync/errgroup"
	grpcstatus "google.golang.org/grpc/status"
)

// ErrorHandler Federation Service often needs to convert errors received from downstream services.
// If an error occurs during method execution in the Federation Service, this error handler is called and the returned error is treated as a final error.
type ErrorHandler func(ctx context.Context, methodName string, err error) error

// RecoveredError represents recovered error.
type RecoveredError struct {
	Message string
	Stack   []string
}

func (e *RecoveredError) Error() string {
	return fmt.Sprintf("recovered error: %s", e.Message)
}

func GoWithRecover(eg *errgroup.Group, fn func() (any, error)) {
	eg.Go(func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = RecoverError(r, debug.Stack())
			}
		}()
		_, err := fn()
		return err
	})
}

type ErrorWithLogAttrs struct {
	err   error
	level slog.Level
	attrs []slog.Attr
}

func (e *ErrorWithLogAttrs) Error() string {
	return e.err.Error()
}

func (e *ErrorWithLogAttrs) Unwrap() error {
	return e.err
}

func (e *ErrorWithLogAttrs) GRPCStatus() *grpcstatus.Status {
	return grpcstatus.Convert(e.err)
}

func NewErrorWithLogAttrs(err error, level slog.Level, attrs []slog.Attr) error {
	if err == nil {
		return nil
	}
	return &ErrorWithLogAttrs{
		err:   err,
		level: level,
		attrs: attrs,
	}
}

func OutputErrorLog(ctx context.Context, err error) {
	if err == nil {
		return
	}
	logger := Logger(ctx)

	logLevel := slog.LevelError
	var logArgs []slog.Attr
	var errWithAttrs *ErrorWithLogAttrs
	if errors.As(err, &errWithAttrs) {
		logLevel = errWithAttrs.level
		logArgs = errWithAttrs.attrs
	}
	if status, ok := grpcstatus.FromError(err); ok {
		logArgs = append(logArgs, slog.Group("grpc_status",
			slog.String("code", status.Code().String()),
			slog.Any("details", status.Details()),
		))
		logger.LogAttrs(ctx, logLevel, status.Message(), logArgs...)
		return
	}
	var recoveredErr *RecoveredError
	if errors.As(err, &recoveredErr) {
		trace := make([]interface{}, 0, len(recoveredErr.Stack))
		for idx, stack := range recoveredErr.Stack {
			trace = append(trace, slog.String(fmt.Sprint(idx+1), stack))
		}
		logger.ErrorContext(ctx, recoveredErr.Message, slog.Group("stack_trace", trace...))
		return
	}
	logger.ErrorContext(ctx, err.Error())
}

func RecoverError(v interface{}, rawStack []byte) *RecoveredError {
	msg := fmt.Sprint(v)
	msgLines := strings.Split(msg, "\n")
	if len(msgLines) <= 1 {
		lines := strings.Split(string(rawStack), "\n")
		stack := make([]string, 0, len(lines))
		for _, line := range lines {
			if line == "" {
				continue
			}
			stack = append(stack, strings.TrimPrefix(line, "\t"))
		}
		return &RecoveredError{
			Message: msg,
			Stack:   stack,
		}
	}
	// If panic occurs under singleflight, singleflight's recover catches the error and gives a stack trace.
	// Therefore, once the stack trace is removed.
	stack := make([]string, 0, len(msgLines))
	for _, line := range msgLines[1:] {
		if line == "" {
			continue
		}
		stack = append(stack, strings.TrimPrefix(line, "\t"))
	}
	return &RecoveredError{
		Message: msgLines[0],
		Stack:   stack,
	}
}

var (
	ErrClientConfig           = errors.New("grpc-federation: Client field is not set. this field must be set")
	ErrResolverConfig         = errors.New("grpc-federation: Resolver field is not set. this field must be set")
	ErrCELPluginConfig        = errors.New("grpc-federation: CELPlugin field is not set. this field must be set")
	ErrCELCacheMap            = errors.New("grpc-federation: CELCacheMap is not found")
	ErrCELCacheIndex          = errors.New("grpc-federation: CELCacheIndex is not set")
	ErrOverflowTypeConversion = errors.New("grpc-federation: overflow type conversion was detected")
)
