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
	"google.golang.org/grpc/test/bufconn"

	"example/content"
	"example/federation"
)

const bufSize = 1024

var (
	listener      *bufconn.Listener
	contentClient content.ContentServiceClient
)

type clientConfig struct{}

func (c *clientConfig) Content_ContentServiceClient(cfg federation.FederationServiceClientConfig) (content.ContentServiceClient, error) {
	return contentClient, nil
}

type ContentServer struct {
	*content.UnimplementedContentServiceServer
}

func (s *ContentServer) GetContent(ctx context.Context, req *content.GetContentRequest) (*content.GetContentResponse, error) {
	return &content.GetContentResponse{
		Content: &content.Content{
			ByField:          req.GetByField(),
			DoubleField:      req.GetDoubleField(),
			DoublesField:     req.GetDoublesField(),
			FloatField:       req.GetFloatField(),
			FloatsField:      req.GetFloatsField(),
			Int32Field:       req.GetInt32Field(),
			Int32SField:      req.GetInt32SField(),
			Int64Field:       req.GetInt64Field(),
			Int64SField:      req.GetInt64SField(),
			Uint32Field:      req.GetUint32Field(),
			Uint32SField:     req.GetUint32SField(),
			Uint64Field:      req.GetUint64Field(),
			Uint64SField:     req.GetUint64SField(),
			Sint32Field:      req.GetSint32Field(),
			Sint32SField:     req.GetSint32SField(),
			Sint64Field:      req.GetSint64Field(),
			Sint64SField:     req.GetSint64SField(),
			Fixed32Field:     req.GetFixed32Field(),
			Fixed32SField:    req.GetFixed32SField(),
			Fixed64Field:     req.GetFixed64Field(),
			Fixed64SField:    req.GetFixed64SField(),
			Sfixed32Field:    req.GetSfixed32Field(),
			Sfixed32SField:   req.GetSfixed32SField(),
			Sfixed64Field:    req.GetSfixed64Field(),
			Sfixed64SField:   req.GetSfixed64SField(),
			BoolField:        req.GetBoolField(),
			BoolsField:       req.GetBoolsField(),
			StringField:      req.GetStringField(),
			StringsField:     req.GetStringsField(),
			ByteStringField:  req.GetByteStringField(),
			ByteStringsField: req.GetByteStringsField(),
			EnumField:        req.GetEnumField(),
			EnumsField:       req.GetEnumsField(),
			EnvField:         req.GetEnvField(),
			EnvsField:        req.GetEnvsField(),
			MessageField:     req.GetMessageField(),
			MessagesField:    req.GetMessagesField(),
		},
	}, nil
}

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
					semconv.ServiceNameKey.String("example08/literal"),
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

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	contentClient = content.NewContentServiceClient(conn)

	grpcServer := grpc.NewServer()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		Client: new(clientConfig),
		Logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	content.RegisterContentServiceServer(grpcServer, &ContentServer{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	t.Setenv("foo", "foo-value")
	t.Setenv("bar", "bar-value")

	client := federation.NewFederationServiceClient(conn)
	res, err := client.Get(ctx, &federation.GetRequest{
		Id: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetResponse{
		Content: &federation.Content{
			ByField:          "foo",
			DoubleField:      1.23,
			DoublesField:     []float64{4.56, 7.89},
			FloatField:       4.56,
			FloatsField:      []float32{7.89, 1.23},
			Int32Field:       -1,
			Int32SField:      []int32{-2, -3},
			Int64Field:       -4,
			Int64SField:      []int64{-5, -6},
			Uint32Field:      1,
			Uint32SField:     []uint32{2, 3},
			Uint64Field:      4,
			Uint64SField:     []uint64{5, 6},
			Sint32Field:      -7,
			Sint32SField:     []int32{-8, -9},
			Sint64Field:      -10,
			Sint64SField:     []int64{-11, -12},
			Fixed32Field:     10,
			Fixed32SField:    []uint32{11, 12},
			Fixed64Field:     13,
			Fixed64SField:    []uint64{14, 15},
			Sfixed32Field:    -14,
			Sfixed32SField:   []int32{-15, -16},
			Sfixed64Field:    -17,
			Sfixed64SField:   []int64{-18, -19},
			BoolField:        true,
			BoolsField:       []bool{true, false},
			StringField:      "foo",
			StringsField:     []string{"hello", "world"},
			ByteStringField:  []byte("foo"),
			ByteStringsField: [][]byte{[]byte("foo"), []byte("bar")},
			EnumField:        federation.ContentType_CONTENT_TYPE_1,
			EnumsField:       []federation.ContentType{federation.ContentType_CONTENT_TYPE_2, federation.ContentType_CONTENT_TYPE_3},
			MessageField: &federation.Content{
				DoubleField:  1.23,
				DoublesField: []float64{4.56, 7.89},
			},
			MessagesField: []*federation.Content{{}, {}},
		},
		CelExpr: -8,
	}, cmpopts.IgnoreUnexported(
		federation.GetResponse{},
		federation.Content{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
	grpcServer.Stop()
}
