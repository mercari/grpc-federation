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
	"path/filepath"
	"testing"

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
	//go:embed net.wasm
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
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	envValue := "TEST_ENV_CAPABILITY"
	t.Setenv("FOO", envValue)

	testFileContent := `{"hello":"world"}`
	testFilePath := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(testFilePath, []byte(testFileContent), 0o644); err != nil {
		t.Fatal(err)
	}

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
					semconv.ServiceNameKey.String("example21/wasm_net"),
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
			Net: federation.FederationServiceCELPluginWasmConfig{
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

	res, err := client.Get(ctx, &federation.GetRequest{
		Url:  "https://example.com",
		Path: testFilePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Body) == 0 {
		t.Fatalf("failed to get body")
	}
	if res.Foo != envValue {
		t.Fatalf("failed to get env value: %s", res.Foo)
	}
	if res.File != testFileContent {
		t.Fatalf("failed to get file content: %s", res.File)
	}
}
