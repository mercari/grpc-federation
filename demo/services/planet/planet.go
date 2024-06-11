package planet

import (
	"context"

	"github.com/peterhellberg/swapi"

	planetpb "github.com/mercari/grpc-federation/demo/swapi/planet"
	"github.com/mercari/grpc-federation/demo/util"
)

type PlanetService struct {
	*planetpb.UnimplementedPlanetServiceServer
	cli   *swapi.Client
	cache map[int64]*planetpb.Planet
}

func NewPlanetService() *PlanetService {
	return &PlanetService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*planetpb.Planet),
	}
}

func (s *PlanetService) GetPlanet(ctx context.Context, req *planetpb.GetPlanetRequest) (*planetpb.GetPlanetReply, error) {
	planet, err := s.getPlanet(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &planetpb.GetPlanetReply{
		Planet: planet,
	}, nil
}

func (s *PlanetService) ListPlanets(ctx context.Context, req *planetpb.ListPlanetsRequest) (*planetpb.ListPlanetsReply, error) {
	planets := make([]*planetpb.Planet, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		planet, err := s.getPlanet(ctx, id)
		if err != nil {
			return nil, err
		}
		planets = append(planets, planet)
	}
	return &planetpb.ListPlanetsReply{
		Planets: planets,
	}, nil
}

func (s *PlanetService) getPlanet(ctx context.Context, id int64) (*planetpb.Planet, error) {
	if planet, exists := s.cache[id]; exists {
		return planet, nil
	}
	res, err := s.cli.Planet(ctx, int(id))
	if err != nil {
		return nil, err
	}
	planet, err := s.toPlanet(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = planet
	return planet, nil
}

func (s *PlanetService) toPlanet(id int64, p *swapi.Planet) (*planetpb.Planet, error) {
	var (
		residentIDs []int64
		filmIDs     []int64
	)
	for _, url := range p.ResidentURLs {
		id, err := util.PersonURLToID(url)
		if err != nil {
			return nil, err
		}
		residentIDs = append(residentIDs, id)
	}
	for _, url := range p.FilmURLs {
		id, err := util.FilmURLToID(url)
		if err != nil {
			return nil, err
		}
		filmIDs = append(filmIDs, id)
	}
	return &planetpb.Planet{
		Id:             id,
		Name:           p.Name,
		Diameter:       p.Diameter,
		RotationPeriod: p.RotationPeriod,
		OrbitalPeriod:  p.OrbitalPeriod,
		Climate:        p.Climate,
		Gravity:        p.Gravity,
		Terrain:        p.Terrain,
		SurfaceWater:   p.SurfaceWater,
		Population:     p.Population,
		Created:        p.Created,
		Edited:         p.Edited,
		Url:            p.URL,
		ResidentIds:    residentIDs,
		FilmIds:        filmIDs,
	}, nil
}
