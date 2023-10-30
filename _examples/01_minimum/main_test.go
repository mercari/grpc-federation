package main_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"example/federation"
)

type Resolver struct {
	*federation.FederationServiceUnimplementedResolver
}

var (
	requestID    = "foo"
	testResponse = &federation.GetPostResponse{
		Post: &federation.Post{
			Id:      requestID,
			Title:   "xxx",
			Content: "yyy",
			User: &federation.User{
				Id:   requestID,
				Name: "zzz",
			},
		},
	}
)

func (r *Resolver) Resolve_Federation_GetPostResponse(
	ctx context.Context,
	arg *federation.Federation_GetPostResponseArgument[*federation.FederationServiceDependentClientSet]) (*federation.GetPostResponse, error) {
	return testResponse, nil
}

func TestFederation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	svc, err := federation.NewFederationService(federation.FederationServiceConfig{
		Resolver: new(Resolver),
		Logger:   logger,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := svc.GetPost(ctx, &federation.GetPostRequest{Id: requestID})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(
		testResponse, res,
		cmpopts.IgnoreUnexported(federation.GetPostResponse{}, federation.Post{}, federation.User{}),
	); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
