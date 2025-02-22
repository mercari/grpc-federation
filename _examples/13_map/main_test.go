package main_test

import (
	"context"
	"fmt"
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
	"example/post"
	"example/user"
)

const bufSize = 1024

var (
	listener   *bufconn.Listener
	postClient post.PostServiceClient
	userClient user.UserServiceClient
)

type clientConfig struct{}

func (c *clientConfig) Post_PostServiceClient(cfg federation.FederationServiceClientConfig) (post.PostServiceClient, error) {
	return postClient, nil
}

func (c *clientConfig) User_UserServiceClient(cfg federation.FederationServiceClientConfig) (user.UserServiceClient, error) {
	return userClient, nil
}

type PostServer struct {
	*post.UnimplementedPostServiceServer
}

func (s *PostServer) GetPosts(ctx context.Context, req *post.GetPostsRequest) (*post.GetPostsResponse, error) {
	var ret post.GetPostsResponse
	for _, id := range req.GetIds() {
		ret.Posts = append(ret.Posts, &post.Post{
			Id:      id,
			Title:   "title for " + id,
			Content: "content for " + id,
			UserId:  id,
		})
	}
	return &ret, nil
}

type UserServer struct {
	*user.UnimplementedUserServiceServer
}

func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	return &user.GetUserResponse{
		User: &user.User{
			Id:   req.Id,
			Name: fmt.Sprintf("name_%s", req.Id),
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
					semconv.ServiceNameKey.String("example13/map"),
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
	userClient = user.NewUserServiceClient(conn)

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
	user.RegisterUserServiceServer(grpcServer, &UserServer{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)
	res, err := client.GetPosts(ctx, &federation.GetPostsRequest{
		Ids: []string{"x", "y"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetPostsResponse{
		Posts: &federation.Posts{
			Ids:      []string{"x", "y"},
			Titles:   []string{"title for x", "title for y"},
			Contents: []string{"content for x", "content for y"},
			Users: []*federation.User{
				{
					Id:   "x",
					Name: "name_x",
				},
				{
					Id:   "y",
					Name: "name_y",
				},
			},
			Items: []*federation.Posts_PostItem{
				{Name: "item_x"},
				{Name: "item_y"},
			},
			ItemTypes: []federation.Item_ItemType{
				federation.Item_ITEM_TYPE_1,
				federation.Item_ITEM_TYPE_2,
			},
			SelectedItemTypes: []federation.Item_ItemType{
				federation.Item_ITEM_TYPE_1,
				federation.Item_ITEM_TYPE_2,
			},
		},
	}, cmpopts.IgnoreUnexported(
		federation.GetPostsResponse{},
		federation.Posts{},
		federation.User{},
		federation.Posts_PostItem{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}

	grpcServer.Stop()
}
