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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
)

const bufSize = 1024

var (
	listener *bufconn.Listener
)

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func TestFederation(t *testing.T) {
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
					semconv.ServiceNameKey.String("example15/celplugin"),
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	setupServer(t, logger, true)

	client := federation.NewFederationServiceClient(conn)
	t.Run("success", func(t *testing.T) {
		testRegexSuccess(t, ctx, client)
	})
	t.Run("success complex", func(t *testing.T) {
		testRegexComplexSuccess(t, ctx, client)
	})
	t.Run("failure", func(t *testing.T) {
		testRegexFailure(t, client, ctx)
	})
}

func BenchmarkFederation(b *testing.B) {
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	conn, err := grpc.DialContext(
		ctx, "",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	setupServer(b, nil, false)

	client := federation.NewFederationServiceClient(conn)
	b.Run("success", func(t *testing.B) {
		testRegexSuccess(t, ctx, client)
	})
	b.Run("success complex", func(t *testing.B) {
		testRegexComplexSuccess(t, ctx, client)
	})
	b.Run("failure", func(t *testing.B) {
		testRegexFailure(t, client, ctx)
	})
}

func setupServer(t testing.TB, logger *slog.Logger, wasmDebugLogging bool) {
	grpcServer := grpc.NewServer()
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		CELPlugin: &federation.FederationServiceCELPluginConfig{
			Regexp: federation.FederationServiceCELPluginWasmConfig{
				Path:         "regexp.wasm",
				Sha256:       "0930ae259c7b742192327a7761fb207ea1d3b7f37f912ca4dc742b3f359af0f9",
				DebugLogging: wasmDebugLogging,
			},
		},
		Logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()
}

func testRegexSuccess(t testing.TB, ctx context.Context, client federation.FederationServiceClient) {
	res, err := client.IsMatch(ctx, &federation.IsMatchRequest{
		Expr:   "hello world",
		Target: "hello world world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.IsMatchResponse{
		Result: true,
	}, cmpopts.IgnoreUnexported(
		federation.IsMatchResponse{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func testRegexComplexSuccess(t testing.TB, ctx context.Context, client federation.FederationServiceClient) {
	res, err := client.IsMatch(ctx, &federation.IsMatchRequest{
		Expr:   `[helo]+\s+(w+\s*)+`,
		Target: "hellooleole world world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.IsMatchResponse{
		Result: true,
	}, cmpopts.IgnoreUnexported(
		federation.IsMatchResponse{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func testRegexFailure(t testing.TB, client federation.FederationServiceClient, ctx context.Context) {
	_, err := client.IsMatch(ctx, &federation.IsMatchRequest{
		Expr:   "[]",
		Target: "hello world world",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	// If testing, produce logging, otherwise do not interfere with the benchmark.
	if _, ok := t.(*testing.T); ok {
		t.Logf("expected error is %s", err)
	}
}
