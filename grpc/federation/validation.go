package federation

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type PreconditionFailureViolation struct {
	Type    string
	Subject string
	Desc    string
}

func PreconditionFailure(ctx context.Context, value localValue, violations []*PreconditionFailureViolation) *errdetails.PreconditionFailure {
	logger := Logger(ctx)

	ret := &errdetails.PreconditionFailure{}
	for idx, violation := range violations {
		typ, err := EvalCEL(ctx, value, violation.Type, reflect.TypeOf(""))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating PreconditionFailure violation type",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		subject, err := EvalCEL(ctx, value, violation.Subject, reflect.TypeOf(""))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating PreconditionFailure violation subject",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		desc, err := EvalCEL(ctx, value, violation.Desc, reflect.TypeOf(""))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating PreconditionFailure violation description",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		ret.Violations = append(ret.Violations, &errdetails.PreconditionFailure_Violation{
			Type:        typ.(string),
			Subject:     subject.(string),
			Description: desc.(string),
		})
	}
	if len(ret.Violations) == 0 {
		return nil
	}
	return ret
}

type BadRequestFieldViolation struct {
	Field string
	Desc  string
}

func BadRequest(ctx context.Context, value localValue, violations []*BadRequestFieldViolation) *errdetails.BadRequest {
	logger := Logger(ctx)

	ret := &errdetails.BadRequest{}

	for idx, violation := range violations {
		field, err := EvalCEL(ctx, value, violation.Field, reflect.TypeOf(""))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating BadRequest field violation field",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		desc, err := EvalCEL(ctx, value, violation.Desc, reflect.TypeOf(""))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating BadRequest field violation description",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		ret.FieldViolations = append(ret.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       field.(string),
			Description: desc.(string),
		})
	}
	if len(ret.FieldViolations) == 0 {
		return nil
	}
	return ret
}

func LocalizedMessage(ctx context.Context, value localValue, locale, msg string) *errdetails.LocalizedMessage {
	logger := Logger(ctx)

	message, err := EvalCEL(ctx, value, msg, reflect.TypeOf(""))
	if err != nil {
		logger.ErrorContext(ctx, "failed evaluating LocalizedMessage message", slog.String("error", err.Error()))
		return nil
	}
	return &errdetails.LocalizedMessage{
		Locale:  locale,
		Message: message.(string),
	}
}

func CustomMessage(ctx context.Context, value localValue, expr string, idx int) proto.Message {
	logger := Logger(ctx)

	msg, err := EvalCEL(ctx, value, expr, reflect.TypeOf(proto.Message(nil)))
	if err != nil {
		logger.ErrorContext(
			ctx,
			"failed evaluating validation error detail message",
			slog.Int("index", idx),
			slog.String("error", err.Error()),
		)
		return nil
	}
	return msg.(proto.Message)
}
