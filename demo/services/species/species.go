package species

import (
	"context"

	"github.com/peterhellberg/swapi"

	speciespb "github.com/mercari/grpc-federation/demo/swapi/species"
	"github.com/mercari/grpc-federation/demo/util"
)

type SpeciesService struct {
	*speciespb.UnimplementedSpeciesServiceServer
	cli   *swapi.Client
	cache map[int64]*speciespb.Species
}

func NewSpeciesService() *SpeciesService {
	return &SpeciesService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*speciespb.Species),
	}
}

func (s *SpeciesService) GetSpecies(ctx context.Context, req *speciespb.GetSpeciesRequest) (*speciespb.GetSpeciesReply, error) {
	species, err := s.getSpecies(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &speciespb.GetSpeciesReply{
		Species: species,
	}, nil
}

func (s *SpeciesService) ListSpecies(ctx context.Context, req *speciespb.ListSpeciesRequest) (*speciespb.ListSpeciesReply, error) {
	species := make([]*speciespb.Species, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		sp, err := s.getSpecies(ctx, id)
		if err != nil {
			return nil, err
		}
		species = append(species, sp)
	}
	return &speciespb.ListSpeciesReply{
		Species: species,
	}, nil
}

func (s *SpeciesService) getSpecies(ctx context.Context, id int64) (*speciespb.Species, error) {
	if species, exists := s.cache[id]; exists {
		return species, nil
	}
	res, err := s.cli.Species(ctx, int(id))
	if err != nil {
		return nil, err
	}
	species, err := s.toSpecies(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = species
	return species, nil
}

func (s *SpeciesService) toSpecies(id int64, sp *swapi.Species) (*speciespb.Species, error) {
	var (
		personIDs []int64
		filmIDs   []int64
	)
	for _, url := range sp.PeopleURLs {
		id, err := util.PersonURLToID(url)
		if err != nil {
			return nil, err
		}
		personIDs = append(personIDs, id)
	}
	for _, url := range sp.FilmURLs {
		id, err := util.FilmURLToID(url)
		if err != nil {
			return nil, err
		}
		filmIDs = append(filmIDs, id)
	}
	return &speciespb.Species{
		Id:              id,
		Name:            sp.Name,
		Classification:  sp.Classification,
		Designation:     sp.Designation,
		AverageHeight:   sp.AverageHeight,
		AverageLifespan: sp.AverageLifespan,
		EyeColors:       sp.EyeColors,
		HairColors:      sp.HairColors,
		SkinColors:      sp.SkinColors,
		Language:        sp.Language,
		Homeworld:       sp.Homeworld,
		Url:             sp.URL,
		Created:         sp.Created,
		Edited:          sp.Edited,
		PersonIds:       personIDs,
		FilmIds:         filmIDs,
	}, nil
}
