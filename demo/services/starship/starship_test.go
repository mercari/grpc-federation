package starship_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/starship"
	starshippb "github.com/mercari/grpc-federation/demo/swapi/starship"
)

func TestStarshipService(t *testing.T) {
	svc := starship.NewStarshipService()
	t.Run("GetStarship", func(t *testing.T) {
		res, err := svc.GetStarship(context.Background(), &starshippb.GetStarshipRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetStarship() == nil {
			t.Fatalf("failed to get starship: %+v", res)
		}
		if res.GetStarship().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
	})
	t.Run("ListStarships", func(t *testing.T) {
		res, err := svc.ListStarships(context.Background(), &starshippb.ListStarshipsRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetStarships()) != 2 {
			t.Fatalf("failed to get starships: %+v", res)
		}
	})
}
