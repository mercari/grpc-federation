package federation

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func WithTimeout[T any](ctx context.Context, method string, timeout time.Duration, fn func(context.Context) (*T, error)) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		ret *T

		// If the channel buffer is empty and a timeout occurs first,
		// the select statement will complete without receiving from `errch`,
		// causing the goroutine to wait indefinitely for a receiver at the end and preventing it from terminating.
		// Therefore, setting the buffer size to 1 ensures that the function can exit even if there is no receiver.
		errch = make(chan error, 1)
	)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errch <- RecoverError(r, debug.Stack())
			}
		}()

		res, err := fn(ctx)
		ret = res
		errch <- err
	}()
	select {
	case <-ctx.Done():
		ctxErr := ctx.Err()

		// If the parent context is canceled,
		// `ctxErr` will reach this condition in the state of context.Canceled.
		// In that case, return an error with the Cancel status.
		if ctxErr == context.Canceled {
			return nil, grpcstatus.New(grpccodes.Canceled, ctxErr.Error()).Err()
		}

		status := grpcstatus.New(grpccodes.DeadlineExceeded, ctxErr.Error())
		withDetails, err := status.WithDetails(&errdetails.ErrorInfo{
			Metadata: map[string]string{
				"method":  method,
				"timeout": timeout.String(),
			},
		})
		if err != nil {
			return nil, status.Err()
		}
		return nil, withDetails.Err()
	case err := <-errch:
		return ret, err
	}
}

type BackOff struct {
	backoff.BackOff
}

func NewConstantBackOff(d time.Duration) *BackOff {
	return &BackOff{
		BackOff: backoff.NewConstantBackOff(d),
	}
}

type ExponentialBackOffConfig struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	MaxElapsedTime      time.Duration
}

func NewExponentialBackOff(cfg *ExponentialBackOffConfig) *BackOff {
	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = cfg.InitialInterval
	eb.RandomizationFactor = cfg.RandomizationFactor
	eb.Multiplier = cfg.Multiplier
	eb.MaxInterval = cfg.MaxInterval
	eb.MaxElapsedTime = cfg.MaxElapsedTime
	return &BackOff{
		BackOff: eb,
	}
}

type RetryParam[T any] struct {
	Value      localValue
	If         string
	CacheIndex int
	BackOff    *BackOff
	Body       func() (*T, error)
}

func WithRetry[T any](ctx context.Context, param *RetryParam[T]) (*T, error) {
	var res *T
	if err := backoff.Retry(func() (err error) {
		result, err := param.Body()
		if err != nil {
			ctx = WithGRPCError(ctx, ToGRPCError(ctx, err))
			cond, evalErr := EvalCEL(ctx, &EvalCELRequest{
				Value:      param.Value,
				Expr:       param.If,
				OutType:    reflect.TypeOf(false),
				CacheIndex: param.CacheIndex,
			})
			if evalErr != nil {
				return backoff.Permanent(evalErr)
			}
			if !cond.(bool) {
				return backoff.Permanent(err)
			}
			return err
		}
		res = result
		return nil
	}, param.BackOff); err != nil {
		return nil, err
	}
	return res, nil
}

func ToLogAttrKey(v any) string {
	return fmt.Sprint(v)
}
