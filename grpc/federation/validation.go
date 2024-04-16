package federation

import (
	"context"
	"log/slog"
	"reflect"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/protoadapt"
)

type PreconditionFailureViolation struct {
	Type                  string
	TypeUseContextLibrary bool
	TypeCacheIndex        int

	Subject                  string
	SubjectUseContextLibrary bool
	SubjectCacheIndex        int

	Desc                  string
	DescUseContextLibrary bool
	DescCacheIndex        int
}

func PreconditionFailure(ctx context.Context, value localValue, violations []*PreconditionFailureViolation) *errdetails.PreconditionFailure {
	logger := Logger(ctx)

	ret := &errdetails.PreconditionFailure{}
	for idx, violation := range violations {
		typ, err := EvalCEL(ctx, &EvalCELRequest{
			Value:             value,
			Expr:              violation.Type,
			UseContextLibrary: violation.TypeUseContextLibrary,
			OutType:           reflect.TypeOf(""),
			CacheIndex:        violation.TypeCacheIndex,
		})
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating PreconditionFailure violation type",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		subject, err := EvalCEL(ctx, &EvalCELRequest{
			Value:             value,
			Expr:              violation.Subject,
			UseContextLibrary: violation.SubjectUseContextLibrary,
			OutType:           reflect.TypeOf(""),
			CacheIndex:        violation.SubjectCacheIndex,
		})
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating PreconditionFailure violation subject",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		desc, err := EvalCEL(ctx, &EvalCELRequest{
			Value:             value,
			Expr:              violation.Desc,
			UseContextLibrary: violation.DescUseContextLibrary,
			OutType:           reflect.TypeOf(""),
			CacheIndex:        violation.DescCacheIndex,
		})
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
	Field                  string
	FieldUseContextLibrary bool
	FieldCacheIndex        int

	Desc                  string
	DescUseContextLibrary bool
	DescCacheIndex        int
}

func BadRequest(ctx context.Context, value localValue, violations []*BadRequestFieldViolation) *errdetails.BadRequest {
	logger := Logger(ctx)

	ret := &errdetails.BadRequest{}

	for idx, violation := range violations {
		field, err := EvalCEL(ctx, &EvalCELRequest{
			Value:             value,
			Expr:              violation.Field,
			UseContextLibrary: violation.FieldUseContextLibrary,
			OutType:           reflect.TypeOf(""),
			CacheIndex:        violation.FieldCacheIndex,
		})
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed evaluating BadRequest field violation field",
				slog.Int("index", idx),
				slog.String("error", err.Error()),
			)
			continue
		}
		desc, err := EvalCEL(ctx, &EvalCELRequest{
			Value:             value,
			Expr:              violation.Desc,
			UseContextLibrary: violation.DescUseContextLibrary,
			OutType:           reflect.TypeOf(""),
			CacheIndex:        violation.DescCacheIndex,
		})
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

type LocalizedMessageParam struct {
	Value             localValue
	Locale            string
	Message           string
	UseContextLibrary bool
	CacheIndex        int
}

func LocalizedMessage(ctx context.Context, param *LocalizedMessageParam) *errdetails.LocalizedMessage {
	logger := Logger(ctx)

	message, err := EvalCEL(ctx, &EvalCELRequest{
		Value:             param.Value,
		Expr:              param.Message,
		UseContextLibrary: param.UseContextLibrary,
		OutType:           reflect.TypeOf(""),
		CacheIndex:        param.CacheIndex,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed evaluating LocalizedMessage message", slog.String("error", err.Error()))
		return nil
	}
	return &errdetails.LocalizedMessage{
		Locale:  param.Locale,
		Message: message.(string),
	}
}

type CustomMessageParam struct {
	Value            localValue
	MessageValueName string
	CacheIndex       int
	MessageIndex     int
}

func CustomMessage(ctx context.Context, param *CustomMessageParam) protoadapt.MessageV1 {
	logger := Logger(ctx)

	msg, err := EvalCEL(ctx, &EvalCELRequest{
		Value:             param.Value,
		Expr:              param.MessageValueName,
		UseContextLibrary: false,
		OutType:           reflect.TypeOf(protoadapt.MessageV1(nil)),
		CacheIndex:        param.CacheIndex,
	})
	if err != nil {
		logger.ErrorContext(
			ctx,
			"failed evaluating validation error detail message",
			slog.Int("index", param.MessageIndex),
			slog.String("error", err.Error()),
		)
		return nil
	}
	return msg.(protoadapt.MessageV1)
}
