package main_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

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

var (
	getPostErr  error
	calledCount int
)

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	calledCount++
	return nil, getPostErr
}

func (s *PostServer) GetPosts(ctx context.Context, req *post.GetPostsRequest) (*post.GetPostsResponse, error) {
	var posts []*post.Post
	for _, id := range req.Ids {
		posts = append(posts, &post.Post{
			Id:      id,
			Title:   "foo",
			Content: "bar",
			UserId:  fmt.Sprintf("user:%s", id),
		})
	}
	return &post.GetPostsResponse{Posts: posts}, nil
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
					semconv.ServiceNameKey.String("example19/retry"),
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
	defer grpcServer.Stop()

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
	defer federation.CleanupFederationService(ctx, federationServer)

	post.RegisterPostServiceServer(grpcServer, &PostServer{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)

	getPostErr = status.New(codes.Unimplemented, "unimplemented get post").Err()
	if _, err := client.GetPost(ctx, &federation.GetPostRequest{Id: "foo"}); err != nil {
		t.Logf("err: %v", err)
	} else {
		t.Fatal("expected error")
	}
	if calledCount != 1 {
		t.Fatalf("failed to retry count. expected 1 but got %d", calledCount)
	}

	calledCount = 0
	getPostErr = status.New(codes.Unavailable, "unavailable get post").Err()
	if _, err := client.GetPost(ctx, &federation.GetPostRequest{Id: "foo"}); err != nil {
		t.Logf("err: %v", err)
	} else {
		t.Fatal("expected error")
	}

	if calledCount != 4 {
		t.Fatalf("failed to retry count. expected 4 but got %d", calledCount)
	}
}
