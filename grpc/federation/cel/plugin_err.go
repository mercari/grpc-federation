package cel

import (
	"context"
	"errors"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// ctxStatusErr attaches a gRPC status to context.Canceled /
// context.DeadlineExceeded so that callers downstream of cel-go's
// *types.Err round-trip can recover the correct grpc code via
// status.FromError. Without it, ctx errors would be converted to codes.Unknown.
type ctxStatusErr struct{ err error }

func (e *ctxStatusErr) Error() string { return e.err.Error() }
func (e *ctxStatusErr) Unwrap() error { return e.err }
func (e *ctxStatusErr) GRPCStatus() *grpcstatus.Status {
	if errors.Is(e.err, context.DeadlineExceeded) {
		return grpcstatus.New(grpccodes.DeadlineExceeded, e.err.Error())
	}
	return grpcstatus.New(grpccodes.Canceled, e.err.Error())
}

// toCELErr converts a Go error returned from a plugin Call into a CEL
// ref.Val. For context cancellation/deadline errors it preserves the
// unwrap chain (and the gRPC status via ctxStatusErr) by going through
// types.WrapErr; for any other error it falls back to NewErrFromString.
func toCELErr(err error) ref.Val {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return types.WrapErr(&ctxStatusErr{err: err})
	}
	return types.NewErrFromString(err.Error())
}
