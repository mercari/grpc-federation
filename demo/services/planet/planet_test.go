package planet_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/planet"
	planetpb "github.com/mercari/grpc-federation/demo/swapi/planet"
)

func TestPlanetService(t *testing.T) {
	svc := planet.NewPlanetService()
	t.Run("GetPlanet", func(t *testing.T) {
		res, err := svc.GetPlanet(context.Background(), &planetpb.GetPlanetRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetPlanet() == nil {
			t.Fatalf("failed to get planet: %+v", res)
		}
		if res.GetPlanet().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
		if len(res.GetPlanet().GetResidentIds()) == 0 {
			t.Fatalf("failed to get resident ids: %+v", res)
		}
		if len(res.GetPlanet().GetFilmIds()) == 0 {
			t.Fatalf("failed to get film ids: %+v", res)
		}
	})
	t.Run("ListPlanets", func(t *testing.T) {
		res, err := svc.ListPlanets(context.Background(), &planetpb.ListPlanetsRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetPlanets()) != 2 {
			t.Fatalf("failed to get planets: %+v", res)
		}
	})
}
