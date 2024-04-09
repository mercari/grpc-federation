package federation

import (
	"context"
	"fmt"
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
		ret   *T
		errch = make(chan error)
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
		status := grpcstatus.New(grpccodes.DeadlineExceeded, ctx.Err().Error())
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

func WithRetry[T any](b *BackOff, fn func() (*T, error)) (*T, error) {
	var res *T
	if err := backoff.Retry(func() (err error) {
		result, err := fn()
		if err != nil {
			return err
		}
		res = result
		return nil
	}, b); err != nil {
		return nil, err
	}
	return res, nil
}

func ToLogAttrKey(v any) string {
	return fmt.Sprint(v)
}
