package cel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/cel-go/common/types"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestToCELErr_PreservesContextCanceled(t *testing.T) {
	val := toCELErr(context.Canceled)
	err, ok := val.(error)
	if !ok {
		t.Fatalf("expected ref.Val to satisfy error, got %T", val)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("errors.Is(err, context.Canceled) = false, want true")
	}
	st, ok := grpcstatus.FromError(err)
	if !ok {
		t.Fatalf("status.FromError did not find a status")
	}
	if st.Code() != grpccodes.Canceled {
		t.Fatalf("grpc code = %s, want Canceled", st.Code())
	}
}

func TestToCELErr_PreservesContextDeadlineExceeded(t *testing.T) {
	val := toCELErr(context.DeadlineExceeded)
	err := val.(error)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("errors.Is(err, context.DeadlineExceeded) = false, want true")
	}
	st, _ := grpcstatus.FromError(err)
	if st.Code() != grpccodes.DeadlineExceeded {
		t.Fatalf("grpc code = %s, want DeadlineExceeded", st.Code())
	}
}

func TestToCELErr_PreservesWrappedContextCanceled(t *testing.T) {
	wrapped := fmt.Errorf("plugin Acquire failed: %w", context.Canceled)
	val := toCELErr(wrapped)
	err := val.(error)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("errors.Is on wrapped chain = false, want true")
	}
	st, _ := grpcstatus.FromError(err)
	if st.Code() != grpccodes.Canceled {
		t.Fatalf("grpc code = %s, want Canceled", st.Code())
	}
}

func TestToCELErr_PassesThroughOtherErrors(t *testing.T) {
	val := toCELErr(errors.New("plugin blew up"))
	err := val.(error)

	if errors.Is(err, context.Canceled) {
		t.Fatalf("non-context error should not match context.Canceled")
	}
	st, _ := grpcstatus.FromError(err)
	if st.Code() != grpccodes.Unknown {
		t.Fatalf("grpc code = %s, want Unknown for non-context error", st.Code())
	}
	if err.Error() != "plugin blew up" {
		t.Fatalf("message = %q, want %q", err.Error(), "plugin blew up")
	}
}

func TestToCELErr_ResultIsCELError(t *testing.T) {
	val := toCELErr(context.Canceled)
	if !types.IsError(val) {
		t.Fatalf("types.IsError(val) = false, want true")
	}
}
