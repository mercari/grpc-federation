package main_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
)

const bufSize = 1024

var (
	listener *bufconn.Listener
)

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func TestCELEvaluation(t *testing.T) {
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		CELPlugin: &federation.FederationServiceCELPluginConfig{
			Account: federation.FederationServiceCELPluginWasmConfig{
				Path:   "account.wasm",
				Sha256: "c7b0a1dcca329d837cfe875c83aee9ac665c2546995bc90514175aaf40719c83",
			},
		},
		Logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "foo" + fmt.Sprint(i)
			ctx := metadata.AppendToOutgoingContext(context.Background(), "id", id)
			res, err := client.Get(ctx, &federation.GetRequest{})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(res, &federation.GetResponse{
				IdFromPlugin:   id,
				IdFromMetadata: id,
			}, cmpopts.IgnoreUnexported(
				federation.GetResponse{},
			)); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		}()
	}
	wg.Wait()
}

func Benchmark_CELEvaluation(b *testing.B) {
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	conn, err := grpc.DialContext(
		ctx, "",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	grpcServer := grpc.NewServer()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		CELPlugin: &federation.FederationServiceCELPluginConfig{
			Account: federation.FederationServiceCELPluginWasmConfig{
				Path:   "account.wasm",
				Sha256: "c7b0a1dcca329d837cfe875c83aee9ac665c2546995bc90514175aaf40719c83",
			},
		},
		Logger: logger,
	})
	if err != nil {
		b.Fatal(err)
	}
	federation.RegisterFederationServiceServer(grpcServer, federationServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			b.Fatal(err)
		}
	}()

	client := federation.NewFederationServiceClient(conn)

	b.ReportAllocs()
	var wg sync.WaitGroup
	for n := 0; n < 100; n++ {
		n := n
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "foo" + fmt.Sprint(n)
			ctx := metadata.AppendToOutgoingContext(context.Background(), "id", id)
			res, err := client.Get(ctx, &federation.GetRequest{})
			if err != nil {
				b.Fatal(err)
			}
			if diff := cmp.Diff(res, &federation.GetResponse{
				IdFromPlugin:   id,
				IdFromMetadata: id,
			}, cmpopts.IgnoreUnexported(
				federation.GetResponse{},
			)); diff != "" {
				b.Errorf("(-got, +want)\n%s", diff)
			}
		}()
	}
	wg.Wait()
}
