package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
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
	//go:embed regexp.wasm
	wasm []byte
)

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func toSha256(v []byte) string {
	hash := sha256.Sum256(v)
	return hex.EncodeToString(hash[:])
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
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
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
		CELPlugin: &federation.FederationServiceCELPluginConfig{
			Regexp: federation.FederationServiceCELPluginWasmConfig{
				Reader: bytes.NewReader(wasm),
				Sha256: toSha256(wasm),
			},
		},
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

	client := federation.NewFederationServiceClient(conn)
	t.Run("success", func(t *testing.T) {
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
	})
	t.Run("failure", func(t *testing.T) {
		_, err := client.IsMatch(ctx, &federation.IsMatchRequest{
			Expr:   "[]",
			Target: "hello world world",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		t.Logf("expected error is %s", err)
	})
	t.Run("example", func(t *testing.T) {
		res, err := client.Example(ctx, &federation.ExampleRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(res, &federation.ExampleResponse{
			Size: 2,
			Str:  "abc",
		}, cmpopts.IgnoreUnexported(
			federation.ExampleResponse{},
		)); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}
