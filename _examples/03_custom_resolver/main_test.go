package main_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
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
	"google.golang.org/protobuf/types/known/durationpb"

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

func (c *clientConfig) Post_PostServiceClient(cfg federation.FederationV2DevServiceClientConfig) (post.PostServiceClient, error) {
	return postClient, nil
}

func (c *clientConfig) User_UserServiceClient(cfg federation.FederationV2DevServiceClientConfig) (user.UserServiceClient, error) {
	return userClient, nil
}

type Resolver struct {
}

func (r *Resolver) Resolve_Federation_V2Dev_User(ctx context.Context, arg *federation.FederationV2DevService_Federation_V2Dev_UserArgument) (*federation.User, error) {
	grpcfed.SetLogger(ctx, grpcfed.Logger(ctx).With(slog.String("foo", "hoge")))
	env := federation.GetFederationV2DevServiceEnv(ctx)
	if env.A != "xxx" {
		return nil, fmt.Errorf("failed to get environement variable for A: %v", env.A)
	}
	if !reflect.DeepEqual(env.B, []int64{1, 2, 3, 4}) {
		return nil, fmt.Errorf("failed to get environement variable for B: %v", env.B)
	}
	if len(env.C) != len(envCMap) {
		return nil, fmt.Errorf("failed to get environement variable for C: %v", env.C)
	}
	if env.D != 0 {
		return nil, fmt.Errorf("failed to get environement variable for D: %v", env.D)
	}
	grpcfed.Logger(ctx).Debug("print env variables", slog.Any("env", env))

	return &federation.User{
		Id:   arg.U.Id,
		Name: arg.U.Name,
	}, nil
}

func (r *Resolver) Resolve_Federation_V2Dev_PostV2Dev_User(ctx context.Context, arg *federation.FederationV2DevService_Federation_V2Dev_PostV2Dev_UserArgument) (*federation.User, error) {
	return arg.User, nil
}

func (r *Resolver) Resolve_Federation_V2Dev_Unused(_ context.Context, _ *federation.FederationV2DevService_Federation_V2Dev_UnusedArgument) (*federation.Unused, error) {
	return &federation.Unused{}, nil
}

func (r *Resolver) Resolve_Federation_V2Dev_ForNameless(_ context.Context, _ *federation.FederationV2DevService_Federation_V2Dev_ForNamelessArgument) (*federation.ForNameless, error) {
	return &federation.ForNameless{}, nil
}

func (r *Resolver) Resolve_Federation_V2Dev_User_Name(_ context.Context, arg *federation.FederationV2DevService_Federation_V2Dev_User_NameArgument) (string, error) {
	return arg.Federation_V2Dev_User.Name, nil
}

func (r *Resolver) Resolve_Federation_V2Dev_TypedNil(_ context.Context, _ *federation.FederationV2DevService_Federation_V2Dev_TypedNilArgument) (*federation.TypedNil, error) {
	return nil, nil
}

type PostServer struct {
	*post.UnimplementedPostServiceServer
}

func (s *PostServer) GetPost(ctx context.Context, req *post.GetPostRequest) (*post.GetPostResponse, error) {
	return nil, errors.New("error!!")
}

func (s *PostServer) GetPosts(ctx context.Context, req *post.GetPostsRequest) (*post.GetPostsResponse, error) {
	return nil, errors.New("error!!")
}

type UserServer struct {
	*user.UnimplementedUserServiceServer
}

func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	if req.Id == "" {
		return &user.GetUserResponse{User: &user.User{Id: "anonymous_id", Name: "anonymous"}}, nil
	}
	return &user.GetUserResponse{
		User: &user.User{
			Id:   req.Id,
			Name: fmt.Sprintf("name_%s", req.Id),
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

var envCMap map[string]grpcfed.Duration

func init() {
	x, err := time.ParseDuration("10h")
	if err != nil {
		panic(err)
	}
	y, err := time.ParseDuration("20m")
	if err != nil {
		panic(err)
	}
	z, err := time.ParseDuration("30s")
	if err != nil {
		panic(err)
	}
	envCMap = map[string]grpcfed.Duration{
		"x": grpcfed.Duration(x),
		"y": grpcfed.Duration(y),
		"z": grpcfed.Duration(z),
	}
}

func TestFederation(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	t.Setenv("YYY", "1,2,3,4")
	t.Setenv("ZZZ", "x:10h,y:20m,z:30s")
	t.Setenv("d", "2.0")

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
					semconv.ServiceNameKey.String("example03/custom_resolver"),
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
	userClient = user.NewUserServiceClient(conn)

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationV2DevService(federation.FederationV2DevServiceConfig{
		Client:   new(clientConfig),
		Resolver: new(Resolver),
		Logger:   logger,
		ErrorHandler: func(ctx context.Context, methodName string, err error) error {
			federationServiceMethodName, _ := grpc.Method(ctx)
			grpcfed.Logger(ctx).InfoContext(
				ctx,
				"error handler",
				slog.String("federation-service-method", federationServiceMethodName),
				slog.String("dependent-method", methodName),
			)
			switch methodName {
			case federation.FederationV2DevService_DependentMethod_Post_PostService_GetPost:
				return nil
			case federation.FederationV2DevService_DependentMethod_User_UserService_GetUser:
				return err
			}
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer federation.CleanupFederationV2DevService(ctx, federationServer)

	post.RegisterPostServiceServer(grpcServer, &PostServer{})
	user.RegisterUserServiceServer(grpcServer, &UserServer{})
	federation.RegisterFederationV2DevServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationV2DevServiceClient(conn)
	res, err := client.GetPostV2Dev(ctx, &federation.GetPostV2DevRequest{
		Id: "foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetPostV2DevResponse{
		Post: &federation.PostV2Dev{
			User: &federation.User{
				Id:   "anonymous_id",
				Name: "anonymous",
			},
			NullCheck: true,
		},
		EnvA:      "xxx",
		EnvB:      2,
		EnvCValue: durationpb.New(envCMap["z"]),
		Ref:       &federation.Ref{A: "xxx"},
	}, cmpopts.IgnoreUnexported(
		federation.GetPostV2DevResponse{},
		federation.PostV2Dev{},
		federation.User{},
		federation.Ref{},
		durationpb.Duration{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
