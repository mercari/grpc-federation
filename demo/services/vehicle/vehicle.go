package vehicle

import (
	"context"

	"github.com/peterhellberg/swapi"

	vehiclepb "github.com/mercari/grpc-federation/demo/swapi/vehicle"
	"github.com/mercari/grpc-federation/demo/util"
)

type VehicleService struct {
	*vehiclepb.UnimplementedVehicleServiceServer
	cli   *swapi.Client
	cache map[int64]*vehiclepb.Vehicle
}

func NewVehicleService() *VehicleService {
	return &VehicleService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*vehiclepb.Vehicle),
	}
}

func (s *VehicleService) GetVehicle(ctx context.Context, req *vehiclepb.GetVehicleRequest) (*vehiclepb.GetVehicleReply, error) {
	vehicle, err := s.getVehicle(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &vehiclepb.GetVehicleReply{
		Vehicle: vehicle,
	}, nil
}

func (s *VehicleService) ListVehicles(ctx context.Context, req *vehiclepb.ListVehiclesRequest) (*vehiclepb.ListVehiclesReply, error) {
	vehicles := make([]*vehiclepb.Vehicle, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		vehicle, err := s.getVehicle(ctx, id)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, vehicle)
	}
	return &vehiclepb.ListVehiclesReply{
		Vehicles: vehicles,
	}, nil
}

func (s *VehicleService) getVehicle(ctx context.Context, id int64) (*vehiclepb.Vehicle, error) {
	if vehicle, exists := s.cache[id]; exists {
		return vehicle, nil
	}
	res, err := s.cli.Vehicle(ctx, int(id))
	if err != nil {
		return nil, err
	}
	vehicle, err := s.toVehicle(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = vehicle
	return vehicle, nil
}

func (s *VehicleService) toVehicle(id int64, v *swapi.Vehicle) (*vehiclepb.Vehicle, error) {
	var (
		filmIDs  []int64
		pilotIDs []int64
	)
	for _, url := range v.FilmURLs {
		id, err := util.FilmURLToID(url)
		if err != nil {
			return nil, err
		}
		filmIDs = append(filmIDs, id)
	}
	for _, url := range v.PilotURLs {
		id, err := util.PersonURLToID(url)
		if err != nil {
			return nil, err
		}
		pilotIDs = append(pilotIDs, id)
	}
	return &vehiclepb.Vehicle{
		Id:                   id,
		Name:                 v.Name,
		Model:                v.Model,
		VehicleClass:         v.VehicleClass,
		Manufacturer:         v.Manufacturer,
		Length:               v.Length,
		CostInCredits:        v.CostInCredits,
		Crew:                 v.Crew,
		Passengers:           v.Passengers,
		MaxAtmospheringSpeed: v.MaxAtmospheringSpeed,
		CargoCapacity:        v.CargoCapacity,
		Consumables:          v.Consumables,
		Url:                  v.URL,
		Created:              v.Created,
		Edited:               v.Edited,
		FilmIds:              filmIDs,
		PilotIds:             pilotIDs,
	}, nil
}
