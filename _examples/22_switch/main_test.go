package main_test

import (
	"context"
	"log/slog"
	"net"
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
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
)

const bufSize = 1024

var (
	listener *bufconn.Listener
)

type clientConfig struct{}

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func TestFederation(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

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
					semconv.ServiceNameKey.String("example22/switch"),
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

	conn, err := grpc.DialContext(
		ctx, "",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		Logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer federation.CleanupFederationService(ctx, federationServer)
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		desc string
		req  *federation.GetPostRequest
		want *federation.GetPostResponse
	}{
		{
			desc: "blue",
			req:  &federation.GetPostRequest{Id: "blue"},
			want: &federation.GetPostResponse{Svar: 2, Switch: 3},
		},
		{
			desc: "red",
			req:  &federation.GetPostRequest{Id: "red"},
			want: &federation.GetPostResponse{Svar: 2, Switch: 4},
		},
		{
			desc: "default",
			req:  &federation.GetPostRequest{Id: "green"},
			want: &federation.GetPostResponse{Svar: 2, Switch: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			client := federation.NewFederationServiceClient(conn)
			res, err := client.GetPost(ctx, tt.req)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(res, tt.want, cmpopts.IgnoreUnexported(
				federation.GetPostResponse{},
			)); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
