package main_test

import (
	"context"
	"example/favorite"
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
	*federation.OtherServiceUnimplementedResolver
}

var otherServicePostData = &federation.Post{
	Id:      "abcd",
	Title:   "tttt",
	Content: "xxxx",
	User: &federation.User{
		Id:   "yyyy",
		Name: "zzzz",
	},
}

func (r *Resolver) Resolve_Federation_GetResponse_Post(_ context.Context, _ *federation.OtherService_Federation_GetResponse_PostArgument) (*federation.Post, error) {
	return otherServicePostData, nil
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
					semconv.ServiceNameKey.String("example11/multi_service"),
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
	t.Run("federation", func(t *testing.T) {
		var (
			requestID           = "foo"
			expectedGetPostResp = &federation.GetPostResponse{
				Post: &federation.Post{
					Id:      "post-id",
					Title:   "title",
					Content: "content",
					User: &federation.User{
						Id:   requestID,
						Name: "bar",
					},
					Reaction: &federation.Reaction{
						FavoriteType:    favorite.FavoriteType_TYPE1,
						FavoriteTypeStr: "TYPE1",
						Cmp:             true,
					},
					FavoriteValue: federation.MyFavoriteType_TYPE1,
					Cmp:           true,
				},
			}
			expectedGetNameResp = &federation.GetNameResponse{
				Name: "federation",
			}
		)

		svc, err := federation.NewFederationService(federation.FederationServiceConfig{
			Logger: logger,
		})
		if err != nil {
			t.Fatal(err)
		}
		gotGetPostResp, err := svc.GetPost(ctx, &federation.GetPostRequest{Id: requestID})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			gotGetPostResp, expectedGetPostResp,
			cmpopts.IgnoreUnexported(federation.GetPostResponse{}, federation.Post{}, federation.User{}, federation.Reaction{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}

		gotGetNameResp, err := svc.GetName(ctx, &federation.GetNameRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			gotGetNameResp, expectedGetNameResp,
			cmpopts.IgnoreUnexported(federation.GetNameResponse{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("debug", func(t *testing.T) {
		expected := &federation.GetStatusResponse{
			User: &federation.User{
				Id:   "xxxx",
				Name: "yyyy",
			},
		}
		svc, err := federation.NewDebugService(federation.DebugServiceConfig{
			Logger: logger,
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := svc.GetStatus(ctx, &federation.GetStatusRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetStatusResponse{}, federation.User{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("private", func(t *testing.T) {
		expected := &federation.GetNameResponse{
			Name: "private",
		}
		svc, err := federation.NewPrivateService(federation.PrivateServiceConfig{
			Logger: logger,
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := svc.GetName(ctx, &federation.GetNameRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetNameResponse{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("other", func(t *testing.T) {
		expected := &federation.GetResponse{
			Post: otherServicePostData,
		}
		svc, err := federation.NewOtherService(federation.OtherServiceConfig{
			Logger:   logger,
			Resolver: new(Resolver),
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := svc.Get(ctx, &federation.GetRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetResponse{}, federation.Post{}, federation.User{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

}
