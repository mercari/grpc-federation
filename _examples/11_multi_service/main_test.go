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
	*federation.OtherServiceUnimplementedResolver
}

var otherServicePostData = &federation.Post{
	Id:      "abcd",
	Title:   "tttt",
	Content: "xxxx",
	User: &federation.User{
		Id:   "yyyy",
		Name: "zzzz",
	},
}

func (r *Resolver) Resolve_Federation_GetResponse_Post(
	ctx context.Context,
	arg *federation.Federation_GetResponse_PostArgument[*federation.OtherServiceDependentClientSet],
) (*federation.Post, error) {
	return otherServicePostData, nil
}

func TestFederation(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	t.Run("federation", func(t *testing.T) {
		var (
			requestID = "foo"
			expected  = &federation.GetPostResponse{
				Post: &federation.Post{
					Id:      "post-id",
					Title:   "title",
					Content: "content",
					User: &federation.User{
						Id:   requestID,
						Name: "bar",
					},
				},
			}
		)

		svc, err := federation.NewFederationService(federation.FederationServiceConfig{
			Logger: logger,
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		got, err := svc.GetPost(ctx, &federation.GetPostRequest{Id: requestID})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetPostResponse{}, federation.Post{}, federation.User{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("debug", func(t *testing.T) {
		expected := &federation.GetStatusResponse{
			User: &federation.User{
				Id:   "xxxx",
				Name: "yyyy",
			},
		}
		svc, err := federation.NewDebugService(federation.DebugServiceConfig{
			Logger: logger,
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		got, err := svc.GetStatus(ctx, &federation.GetStatusRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetStatusResponse{}, federation.User{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
	t.Run("other", func(t *testing.T) {
		expected := &federation.GetResponse{
			Post: otherServicePostData,
		}
		svc, err := federation.NewOtherService(federation.OtherServiceConfig{
			Logger:   logger,
			Resolver: new(Resolver),
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		got, err := svc.Get(ctx, &federation.GetRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(
			got, expected,
			cmpopts.IgnoreUnexported(federation.GetResponse{}, federation.Post{}, federation.User{}),
		); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

}
