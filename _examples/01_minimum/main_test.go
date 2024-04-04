package main_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"example/federation"
)

type Resolver struct {
	*federation.FederationServiceUnimplementedResolver
}

var (
	requestID    = "foo"
	testResponse = &federation.GetPostResponse{
		Post: &federation.Post{
			Id:      requestID,
			Title:   "xxx",
			Content: "yyy",
			User: &federation.User{
				Id:   requestID,
				Name: "zzz",
			},
		},
	}
)

func (r *Resolver) Resolve_Federation_GetPostResponse(ctx context.Context, arg *federation.Federation_GetPostResponseArgument) (*federation.GetPostResponse, error) {
	return testResponse, nil
}

func TestFederation(t *testing.T) {
	ctx := context.Background()

	if os.Getenv("ENABLE_JAEGER") != "" {
		exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(
				resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNameKey.String("example01/minimum"),
					semconv.ServiceVersionKey.String("1.0.0"),
					attribute.String("environment", "dev"),
				),
			),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		defer tp.Shutdown(ctx)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTracerProvider(tp)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	svc, err := federation.NewFederationService(federation.FederationServiceConfig{
		Resolver: new(Resolver),
		Logger:   logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.GetPost(ctx, &federation.GetPostRequest{Id: requestID})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(
		testResponse, res,
		cmpopts.IgnoreUnexported(federation.GetPostResponse{}, federation.Post{}, federation.User{}),
	); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
