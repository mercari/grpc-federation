package main_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

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
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
	"example/post"
)

const bufSize = 1024

var (
	listener         *bufconn.Listener
	postClient       post.PostServiceClient
	calledUpdatePost bool
	updateDone       = make(chan struct{})
	blockCh          = make(chan struct{})
)

type clientConfig struct{}

func (c *clientConfig) Post_PostServiceClient(cfg federation.FederationServiceClientConfig) (post.PostServiceClient, error) {
	return postClient, nil
}

type PostServer struct {
	*post.UnimplementedPostServiceServer
}

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	<-blockCh
	return &post.GetPostResponse{
		Post: &post.Post{
			Id:      req.Id,
			Title:   "foo",
			Content: "bar",
			UserId:  fmt.Sprintf("user:%s", req.Id),
		},
	}, nil
}

func (s *PostServer) UpdatePost(ctx context.Context, req *post.UpdatePostRequest) (*post.UpdatePostResponse, error) {
	calledUpdatePost = true
	time.Sleep(2 * time.Second)
	updateDone <- struct{}{}
	return nil, nil
}

func (s *PostServer) DeletePost(ctx context.Context, req *post.DeletePostRequest) (*post.DeletePostResponse, error) {
	time.Sleep(1 * time.Second)
	return nil, status.New(codes.Internal, "failed to delete").Err()
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
					semconv.ServiceNameKey.String("example04/timeout"),
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

	t.Run("GetPost", func(t *testing.T) {
		if _, err := client.GetPost(ctx, &federation.GetPostRequest{
			Id: "foo",
		}); err == nil {
			t.Fatal("expected error")
		} else {
			if status.Code(err) != codes.DeadlineExceeded {
				t.Fatalf("unexpected status code: %v", err)
			}
		}
	})
	t.Run("UpdatePost", func(t *testing.T) {
		_, err := client.UpdatePost(ctx, &federation.UpdatePostRequest{
			Id: "foo",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !calledUpdatePost {
			t.Fatalf("failed to call update post")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatal("failed to get gRPC status error")
		}
		if st.Code() != codes.Internal {
			t.Fatalf("failed to get status code: %s", st.Code())
		}
	})
	blockCh <- struct{}{}
	<-updateDone
	time.Sleep(100 * time.Millisecond)
}
