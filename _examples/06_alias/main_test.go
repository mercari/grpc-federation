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
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
	"example/post"
	postv2 "example/post/v2"
)

const bufSize = 1024

var (
	listener     *bufconn.Listener
	postClient   post.PostServiceClient
	postv2Client postv2.PostServiceClient
)

type clientConfig struct{}

func (c *clientConfig) Org_Post_PostServiceClient(cfg federation.FederationServiceClientConfig) (post.PostServiceClient, error) {
	return postClient, nil
}

func (c *clientConfig) Org_Post_V2_PostServiceClient(cfg federation.FederationServiceClientConfig) (postv2.PostServiceClient, error) {
	return postv2Client, nil
}

type PostServer struct {
	*post.UnimplementedPostServiceServer
}

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	return &post.GetPostResponse{
		Post: &post.Post{
			Id: req.Id,
			Data: &post.PostData{
				Type:  post.PostDataType_POST_TYPE_C,
				Title: "foo",
				Content: &post.PostContent{
					Category: post.PostContent_CATEGORY_A,
					Head:     "headhead",
					Body:     "bodybody",
				},
			},
		},
	}, nil
}

type PostV2Server struct {
	*postv2.UnimplementedPostServiceServer
}

func (s *PostV2Server) GetPost(ctx context.Context, req *postv2.GetPostRequest) (*postv2.GetPostResponse, error) {
	return &postv2.GetPostResponse{
		Post: &postv2.Post{
			Id: req.Id,
			Data: &postv2.PostData{
				Type:  postv2.PostDataType_POST_V2_TYPE_C,
				Title: "foo2",
				Content: &postv2.PostContent{
					Category: postv2.PostContent_CATEGORY_A,
					Head:     "headhead2",
					Body:     "bodybody2",
				},
			},
		},
	}, nil
}

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
					semconv.ServiceNameKey.String("example06/alias"),
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
	postv2Client = postv2.NewPostServiceClient(conn)

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
	postv2.RegisterPostServiceServer(grpcServer, &PostV2Server{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)
	res, err := client.GetPost(ctx, &federation.GetPostRequest{
		Id: "foo",
		Condition: &federation.GetPostRequest_A{
			A: &federation.GetPostRequest_ConditionA{
				Prop: "bar",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetPostResponse{
		Post: &federation.Post{
			Id: "foo",
			Data: &federation.PostData{
				Type:  federation.PostType_POST_TYPE_BAR,
				Title: "foo",
				Content: &federation.PostContent{
					Head:    "headhead",
					Body:    "bodybody",
					DupBody: "bodybody",
				},
			},
			Data2: &federation.PostData{
				Type:  federation.PostType_POST_TYPE_BAR,
				Title: "foo2",
				Content: &federation.PostContent{
					Head:    "headhead2",
					Body:    "bodybody2",
					DupBody: "bodybody2",
				},
			},
			Type: federation.PostType_POST_TYPE_BAR,
		},
	}, cmpopts.IgnoreUnexported(
		federation.GetPostResponse{},
		federation.Post{},
		federation.PostData{},
		federation.PostContent{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
