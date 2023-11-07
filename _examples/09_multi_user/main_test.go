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
	"example/user"
)

const bufSize = 1024

var (
	listener   *bufconn.Listener
	userClient user.UserServiceClient
)

type clientConfig struct{}

func (c *clientConfig) User_UserServiceClient(cfg federation.FederationServiceClientConfig) (user.UserServiceClient, error) {
	return userClient, nil
}

type UserServer struct {
	*user.UnimplementedUserServiceServer
}

var nameMap = map[string]string{
	"xxx": "foo",
}

func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	return &user.GetUserResponse{
		User: &user.User{
			Id:   req.Id,
			Name: nameMap[req.Id],
		},
	}, nil
}

type resolver struct{}

func (r *resolver) Resolve_Org_Federation_Sub(
	_ context.Context,
	_ *federation.Org_Federation_SubArgument[*federation.FederationServiceDependentClientSet]) (*federation.Sub, error) {
	return &federation.Sub{}, nil
}

func (r *resolver) Resolve_Org_Federation_User_Name(
	_ context.Context,
	arg *federation.Org_Federation_User_NameArgument[*federation.FederationServiceDependentClientSet]) (string, error) {
	return arg.User.Name, nil
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
					semconv.ServiceNameKey.String("example09/multi_user"),
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

	userClient = user.NewUserServiceClient(conn)

	grpcServer := grpc.NewServer()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		Client:   new(clientConfig),
		Logger:   logger,
		Resolver: new(resolver),
	})
	if err != nil {
		t.Fatal(err)
	}
	user.RegisterUserServiceServer(grpcServer, &UserServer{})
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)
	res, err := client.Get(ctx, &federation.GetRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, &federation.GetResponse{
		User: &federation.User{
			Id:   "xxx",
			Name: "foo",
		},
		User2: &federation.User{
			Id:   "xxx",
			Name: "foo",
		},
	}, cmpopts.IgnoreUnexported(
		federation.GetResponse{},
		federation.User{},
	)); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
