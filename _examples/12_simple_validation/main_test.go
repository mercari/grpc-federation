package main_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"

	"example/federation"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024

var listener   *bufconn.Listener

func dialer(ctx context.Context, address string) (net.Conn, error) {
	return listener.Dial()
}

func TestFederation(t *testing.T) {
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	grpcServer := grpc.NewServer()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
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

	for _, tc := range []struct{
		desc string
		request *federation.GetPostRequest
		expected *federation.GetPostResponse
		expectedErr string
	} {
		{
			desc: "success",
			request: &federation.GetPostRequest{
				Id: "correct-id",
			},
			expected: &federation.GetPostResponse{
				Post: &federation.Post{
					Id: "some-id",
					Title: "some-title",
					Content: "some-content",
				},
			},
		},
		{
			desc: "validation failure",
			request: &federation.GetPostRequest{
				Id: "wrong-id",
			},
			expectedErr: "validation failure",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := client.GetPost(ctx, tc.request)
			if err != nil {
				if tc.expectedErr == "" {
					t.Fatalf("failed to call GetPost: %v", err)
				}

				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Fatalf("expected %q contains a substring %q", err.Error(), tc.expectedErr)
				}
				return
			}

			if tc.expectedErr != "" {
				t.Fatal("expected to receive an error but got nil")
			}

			if diff := cmp.Diff(res, tc.expected, cmpopts.IgnoreUnexported(
				federation.GetPostResponse{},
				federation.Post{},
			)); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
