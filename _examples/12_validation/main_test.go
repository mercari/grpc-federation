package main_test

import (
	"context"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/goleak"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"example/federation"
)

type Resolver struct{}

func (r *Resolver) Resolve_Org_Federation_CustomHandlerMessage(_ context.Context, _ *federation.FederationService_Org_Federation_CustomHandlerMessageArgument) (*federation.CustomHandlerMessage, error) {
	return &federation.CustomHandlerMessage{}, nil
}

const bufSize = 1024

var listener *bufconn.Listener

func dialer(_ context.Context, _ string) (net.Conn, error) {
	return listener.Dial()
}

func TestFederation(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := context.Background()
	listener = bufconn.Listen(bufSize)

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	federationServer, err := federation.NewFederationService(federation.FederationServiceConfig{
		Resolver: &Resolver{},
		Logger:   logger,
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

	type errStatus struct {
		code    codes.Code
		message string
		details []any
	}
	for _, tc := range []struct {
		desc        string
		request     *federation.GetPostRequest
		expected    *federation.GetPostResponse
		expectedErr *errStatus
	}{
		{
			desc: "success",
			request: &federation.GetPostRequest{
				Id: "correct-id",
			},
			expected: &federation.GetPostResponse{
				Post: &federation.Post{
					Id:      "some-id",
					Title:   "some-title",
					Content: "some-content",
					Item: &federation.Item{
						ItemId: 2,
						Name:   "item-name2",
					},
				},
			},
		},
		{
			desc: "validation failure",
			request: &federation.GetPostRequest{
				Id: "wrong-id",
			},
			expectedErr: &errStatus{
				code:    codes.FailedPrecondition,
				message: "validation3 failed!",
				details: []any{
					&federation.CustomMessage{
						Message: "message1",
					},
					&federation.CustomMessage{
						Message: "message2",
					},
					&federation.CustomMessage{
						Message: "foo",
					},
					&errdetails.PreconditionFailure{
						Violations: []*errdetails.PreconditionFailure_Violation{
							{
								Type:        "type1",
								Subject:     "some-id",
								Description: "description1",
							},
						},
					},
					&errdetails.BadRequest{
						FieldViolations: []*errdetails.BadRequest_FieldViolation{
							{
								Field:       "some-id",
								Description: "description2",
							},
						},
					},
					&errdetails.LocalizedMessage{
						Locale:  "en-US",
						Message: "some-content",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := client.GetPost(ctx, tc.request)
			if err != nil {
				if tc.expectedErr == nil {
					t.Fatalf("failed to call GetPost: %v", err)
				}

				s, ok := status.FromError(err)
				if !ok {
					t.Fatalf("failed to extract gRPC Status from the error: %v", err)
				}

				if got := s.Code(); got != tc.expectedErr.code {
					t.Errorf("invalid a gRPC status code: got: %v, expected: %v", got, tc.expectedErr.code)
				}

				if got := s.Message(); got != tc.expectedErr.message {
					t.Errorf("invalid a gRPC status message: got: %v, expected: %v", got, tc.expectedErr.message)
				}
				if diff := cmp.Diff(s.Details(), tc.expectedErr.details, cmpopts.IgnoreUnexported(
					federation.CustomMessage{},
					errdetails.PreconditionFailure{},
					errdetails.PreconditionFailure_Violation{},
					errdetails.BadRequest{},
					errdetails.BadRequest_FieldViolation{},
					errdetails.LocalizedMessage{},
				)); diff != "" {
					t.Errorf("(-got, +want)\n%s", diff)
				}
				return
			}

			if tc.expectedErr != nil {
				t.Fatal("expected to receive an error but got nil")
			}

			if diff := cmp.Diff(res, tc.expected, cmpopts.IgnoreUnexported(
				federation.GetPostResponse{},
				federation.Post{},
				federation.Item{},
			)); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}
