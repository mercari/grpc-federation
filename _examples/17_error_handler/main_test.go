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
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/protoadapt"

	"example/federation"
	"example/post"
)

const bufSize = 1024

var (
	listener   *bufconn.Listener
	postClient post.PostServiceClient
)

type clientConfig struct{}

func (c *clientConfig) Post_PostServiceClient(cfg federation.FederationServiceClientConfig) (post.PostServiceClient, error) {
	return postClient, nil
}

type PostServer struct {
	*post.UnimplementedPostServiceServer
}

var getPostError error

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	return nil, getPostError
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
					semconv.ServiceNameKey.String("example17/error_handler"),
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

	postClient = post.NewPostServiceClient(conn)

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
	post.RegisterPostServiceServer(grpcServer, &PostServer{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)
	t.Run("custom error", func(t *testing.T) {
		st := status.New(codes.FailedPrecondition, "failed to create post message")
		var details []protoadapt.MessageV1
		details = append(details, &errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "foo",
					Subject:     "bar",
					Description: "baz",
				},
			},
		})
		details = append(details, &errdetails.LocalizedMessage{
			Locale:  "en-US",
			Message: "hello",
		})
		details = append(details, &post.Post{Id: "xxx"})
		stWithDetails, _ := st.WithDetails(details...)
		getPostError = stWithDetails.Err()

		_, err := client.GetPost(ctx, &federation.GetPostRequest{Id: "x"})
		if err == nil {
			t.Fatal("expected error")
		}
		s, ok := status.FromError(err)
		if !ok {
			t.Fatalf("failed to extract gRPC Status from the error: %v", err)
		}
		if s.Message() != `this is custom error message` {
			t.Fatalf("got unexpected error: %v", err)
		}
		var detailNum int
		for _, detail := range s.Details() {
			if _, ok := detail.(protoadapt.MessageV1); !ok {
				t.Fatalf("failed to get proto message from error details: %T", detail)
			}
			detailNum++
		}
		if detailNum != 4 {
			t.Fatalf("failed to get error details. got detail num: %d", detailNum)
		}
	})
	t.Run("ignore error and response", func(t *testing.T) {
		st := status.New(codes.Unimplemented, "unimplemented error")
		getPostError = st.Err()

		res, err := client.GetPost(ctx, &federation.GetPostRequest{Id: ""})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(res, &federation.GetPostResponse{Post: &federation.Post{Id: "anonymous"}},
			cmpopts.IgnoreUnexported(
				federation.GetPostResponse{},
				federation.Post{},
			),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("ignore error", func(t *testing.T) {
		st := status.New(codes.FailedPrecondition, "failed to create post message")
		getPostError = st.Err()

		res, err := client.GetPost(ctx, &federation.GetPostRequest{Id: ""})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(res, &federation.GetPostResponse{Post: &federation.Post{}},
			cmpopts.IgnoreUnexported(
				federation.GetPostResponse{},
				federation.Post{},
			),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("custom log level", func(t *testing.T) {
		st := status.New(codes.InvalidArgument, "failed to create post message")
		getPostError = st.Err()

		_, err := client.GetPost(ctx, &federation.GetPostRequest{Id: "y"})
		if err == nil {
			t.Fatal("expected error")
		}
		s, ok := status.FromError(err)
		if !ok {
			t.Fatalf("failed to extract gRPC Status from the error: %v", err)
		}
		if s.Message() != `this is custom log level` {
			t.Fatalf("got unexpected error: %v", err)
		}
	})
	t.Run("pass through default error", func(t *testing.T) {
		code := codes.FailedPrecondition
		msg := "this is default error message"
		detail := &errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "x",
					Subject:     "y",
					Description: "z",
				},
			},
		}
		st, _ := status.New(code, msg).WithDetails(detail)
		getPostError = st.Err()
		_, err := client.GetPost2(ctx, &federation.GetPost2Request{Id: "y"})
		if err == nil {
			t.Fatal("expected error")
		}
		s, ok := status.FromError(err)
		if !ok {
			t.Fatalf("failed to extract gRPC Status from the error: %v", err)
		}
		if s.Code() != code {
			t.Fatalf("got unexpected error: %v", err)
		}
		if s.Message() != msg {
			t.Fatalf("got unexpected error: %v", err)
		}
		if len(s.Details()) != 1 {
			t.Fatalf("got unexpected error: %v", err)
		}
		if diff := cmp.Diff(s.Details()[0], detail,
			cmpopts.IgnoreUnexported(
				errdetails.PreconditionFailure{},
				errdetails.PreconditionFailure_Violation{},
			),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	grpcServer.Stop()
}
