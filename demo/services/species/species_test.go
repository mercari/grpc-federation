package species_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/species"
	speciespb "github.com/mercari/grpc-federation/demo/swapi/species"
)

func TestSpeciesService(t *testing.T) {
	svc := species.NewSpeciesService()
	t.Run("GetSpecies", func(t *testing.T) {
		res, err := svc.GetSpecies(context.Background(), &speciespb.GetSpeciesRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetSpecies() == nil {
			t.Fatalf("failed to get species: %+v", res)
		}
		if res.GetSpecies().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
	})
	t.Run("ListSpeciess", func(t *testing.T) {
		res, err := svc.ListSpecies(context.Background(), &speciespb.ListSpeciesRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetSpecies()) != 2 {
			t.Fatalf("failed to get species: %+v", res)
		}
	})
}
