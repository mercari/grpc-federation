package vehicle_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/vehicle"
	vehiclepb "github.com/mercari/grpc-federation/demo/swapi/vehicle"
)

func TestVehicleService(t *testing.T) {
	svc := vehicle.NewVehicleService()
	t.Run("GetVehicle", func(t *testing.T) {
		res, err := svc.GetVehicle(context.Background(), &vehiclepb.GetVehicleRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetVehicle() == nil {
			t.Fatalf("failed to get vehicle: %+v", res)
		}
		if res.GetVehicle().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
	})
	t.Run("ListVehicles", func(t *testing.T) {
		res, err := svc.ListVehicles(context.Background(), &vehiclepb.ListVehiclesRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetVehicles()) != 2 {
			t.Fatalf("failed to get vehicles: %+v", res)
		}
	})
}
