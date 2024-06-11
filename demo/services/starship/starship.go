package starship

import (
	"context"

	"github.com/peterhellberg/swapi"

	starshippb "github.com/mercari/grpc-federation/demo/swapi/starship"
	"github.com/mercari/grpc-federation/demo/util"
)

type StarshipService struct {
	*starshippb.UnimplementedStarshipServiceServer
	cli   *swapi.Client
	cache map[int64]*starshippb.Starship
}

func NewStarshipService() *StarshipService {
	return &StarshipService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*starshippb.Starship),
	}
}

func (s *StarshipService) GetStarship(ctx context.Context, req *starshippb.GetStarshipRequest) (*starshippb.GetStarshipReply, error) {
	starship, err := s.getStarship(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &starshippb.GetStarshipReply{
		Starship: starship,
	}, nil
}

func (s *StarshipService) ListStarships(ctx context.Context, req *starshippb.ListStarshipsRequest) (*starshippb.ListStarshipsReply, error) {
	starships := make([]*starshippb.Starship, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		starship, err := s.getStarship(ctx, id)
		if err != nil {
			return nil, err
		}
		starships = append(starships, starship)
	}
	return &starshippb.ListStarshipsReply{
		Starships: starships,
	}, nil
}

func (s *StarshipService) getStarship(ctx context.Context, id int64) (*starshippb.Starship, error) {
	if starship, exists := s.cache[id]; exists {
		return starship, nil
	}
	res, err := s.cli.Starship(ctx, int(id))
	if err != nil {
		return nil, err
	}
	starship, err := s.toStarship(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = starship
	return starship, nil
}

func (s *StarshipService) toStarship(id int64, st *swapi.Starship) (*starshippb.Starship, error) {
	var (
		filmIDs  []int64
		pilotIDs []int64
	)
	for _, url := range st.FilmURLs {
		id, err := util.FilmURLToID(url)
		if err != nil {
			return nil, err
		}
		filmIDs = append(filmIDs, id)
	}
	for _, url := range st.PilotURLs {
		id, err := util.PersonURLToID(url)
		if err != nil {
			return nil, err
		}
		pilotIDs = append(pilotIDs, id)
	}
	return &starshippb.Starship{
		Id:                   id,
		Name:                 st.Name,
		Model:                st.Model,
		StarshipClass:        st.StarshipClass,
		Manufacturer:         st.Manufacturer,
		CostInCredits:        st.CostInCredits,
		Length:               st.Length,
		Crew:                 st.Crew,
		Passengers:           st.Passengers,
		MaxAtmospheringSpeed: st.MaxAtmospheringSpeed,
		HyperdriveRating:     st.HyperdriveRating,
		Mglt:                 st.MGLT,
		CargoCapacity:        st.CargoCapacity,
		Consumables:          st.Consumables,
		Url:                  st.URL,
		Created:              st.Created,
		Edited:               st.Edited,
		PilotIds:             pilotIDs,
		FilmIds:              filmIDs,
	}, nil
}
