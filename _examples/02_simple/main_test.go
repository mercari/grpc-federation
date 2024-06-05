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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

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

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	return &post.GetPostResponse{
		Post: &post.Post{
			Id:      req.Id,
			Title:   "foo",
			Content: "bar",
			UserId:  fmt.Sprintf("user:%s", req.Id),
		},
	}, nil
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

type UserServer struct {
	*user.UnimplementedUserServiceServer
}

func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	profile, err := anypb.New(&user.User{
		Name: "foo",
	})
	if err != nil {
		return nil, err
	}
	return &user.GetUserResponse{
		User: &user.User{
			Id:   req.Id,
			Name: fmt.Sprintf("name_%s", req.Id),
			Items: []*user.Item{
				{
					Name:  "item1",
					Type:  user.Item_ITEM_TYPE_1,
					Value: 1,
					Location: &user.Item_Location{
						Addr1: "foo",
						Addr2: "bar",
						Addr3: &user.Item_Location_B{
							B: &user.Item_Location_AddrB{Bar: 1},
						},
					},
				},
				{Name: "item2", Type: user.Item_ITEM_TYPE_2, Value: 2},
			},
			Profile: map[string]*anypb.Any{"user": profile},
			Attr: &user.User_B{
				B: &user.User_AttrB{
					Bar: true,
				},
			},
		},
	}, nil
}

func (s *UserServer) GetUsers(ctx context.Context, req *user.GetUsersRequest) (*user.GetUsersResponse, error) {
	var users []*user.User
	for _, id := range req.Ids {
		users = append(users, &user.User{
			Id:   id,
			Name: fmt.Sprintf("name_%s", id),
		})
	}
	return &user.GetUsersResponse{Users: users}, nil
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
					semconv.ServiceNameKey.String("example02/simple"),
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
	ctx = metadata.AppendToOutgoingContext(ctx, "key1", "value1")
	res, err := client.GetPost(ctx, &federation.GetPostRequest{
		Id: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}

	profile, err := anypb.New(&user.User{
		Name: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetPostResponse{
		Post: &federation.Post{
			Id:      "foo",
			Title:   "foo",
			Content: "bar",
			User: &federation.User{
				Id:   "user:foo",
				Name: "name_user:foo",
				Items: []*federation.Item{
					{
						Name:  "item1",
						Type:  federation.Item_ITEM_TYPE_1,
						Value: 1,
						Location: &federation.Item_Location{
							Addr1: "foo",
							Addr2: "bar",
							Addr3: &federation.Item_Location_B{
								B: &federation.Item_Location_AddrB{
									Bar: 1,
								},
							},
						},
					},
					{
						Name:  "item2",
						Type:  federation.Item_ITEM_TYPE_2,
						Value: 2,
					},
				},
				Profile: map[string]*anypb.Any{
					"user": profile,
				},
				Attr: &federation.User_B{
					B: &federation.User_AttrB{
						Bar: true,
					},
				},
			},
		},
		Str:               "hello",
		Uuid:              "daa4728d-159f-4fc2-82cf-cae915d54e08",
		Loc:               "Asia/Tokyo",
		Value1:            "value1",
		ItemTypeName:      "ITEM_TYPE_1",
		LocationTypeName:  "LOCATION_TYPE_1",
		UserItemTypeName:  "ITEM_TYPE_2",
		ItemTypeValueEnum: federation.Item_ITEM_TYPE_1,
		ItemTypeValueInt:  1,
		ItemTypeValueCast: federation.Item_ITEM_TYPE_1,
		LocationTypeValue: 1,
		UserItemTypeValue: 2,
		A: &federation.A{
			B: &federation.A_B{
				Foo: &federation.A_B_C{
					Type: "foo",
				},
				Bar: &federation.A_B_C{
					Type: "bar",
				},
			},
		},
		SortedValues: []int32{1, 2, 3, 4},
		SortedItems: []*federation.Item{
			{
				Location: &federation.Item_Location{
					Addr1: "b",
				},
			},
			{
				Location: &federation.Item_Location{
					Addr1: "a",
				},
			},
		},
		MapValue: map[int32]string{
			1: "a",
			2: "b",
			3: "c",
		},
		DoubleWrapperValue: &wrapperspb.DoubleValue{
			Value: 1.23,
		},
		FloatWrapperValue: &wrapperspb.FloatValue{
			Value: 3.45,
		},
		I64WrapperValue: &wrapperspb.Int64Value{
			Value: 1,
		},
		U64WrapperValue: &wrapperspb.UInt64Value{
			Value: 2,
		},
		I32WrapperValue: &wrapperspb.Int32Value{
			Value: 3,
		},
		U32WrapperValue: &wrapperspb.UInt32Value{
			Value: 4,
		},
		StringWrapperValue: &wrapperspb.StringValue{
			Value: "hello",
		},
		BytesWrapperValue: &wrapperspb.BytesValue{
			Value: []byte("world"),
		},
		BoolWrapperValue: &wrapperspb.BoolValue{
			Value: true,
		},
		Hello: "hello\nworld",
	}, cmpopts.IgnoreUnexported(
		federation.GetPostResponse{},
		federation.Post{},
		federation.User{},
		federation.Item{},
		federation.Item_Location{},
		federation.User_B{},
		federation.User_AttrB{},
		federation.Item_Location_B{},
		federation.Item_Location_AddrB{},
		federation.A{},
		federation.A_B{},
		federation.A_B_C{},
		anypb.Any{},
		wrapperspb.DoubleValue{},
		wrapperspb.FloatValue{},
		wrapperspb.Int64Value{},
		wrapperspb.UInt64Value{},
		wrapperspb.Int32Value{},
		wrapperspb.UInt32Value{},
		wrapperspb.StringValue{},
		wrapperspb.BytesValue{},
		wrapperspb.BoolValue{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
