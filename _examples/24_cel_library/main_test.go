package main_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"example/extlib"
	"example/federation"
)

const bufSize = 1024

var listener *bufconn.Listener

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func TestFederation(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		Logger: logger,
		CELLibraries: []grpcfed.CELSingletonLibrary{
			extlib.NewLibrary(),
		},
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
	res, err := client.Greet(ctx, &federation.GreetRequest{Name: "world"})
	if err != nil {
		t.Fatal(err)
	}
	want := &federation.GreetResponse{Message: "hello, WORLD"}
	if diff := cmp.Diff(res, want, cmpopts.IgnoreUnexported(federation.GreetResponse{})); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
