package person

import (
	"context"

	"github.com/peterhellberg/swapi"

	personpb "github.com/mercari/grpc-federation/demo/swapi/person"
	"github.com/mercari/grpc-federation/demo/util"
)

type PersonService struct {
	*personpb.UnimplementedPersonServiceServer
	cli   *swapi.Client
	cache map[int64]*personpb.Person
}

func NewPersonService() *PersonService {
	return &PersonService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*personpb.Person),
	}
}

func (s *PersonService) GetPerson(ctx context.Context, req *personpb.GetPersonRequest) (*personpb.GetPersonReply, error) {
	person, err := s.getPerson(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &personpb.GetPersonReply{
		Person: person,
	}, nil
}

func (s *PersonService) ListPeople(ctx context.Context, req *personpb.ListPeopleRequest) (*personpb.ListPeopleReply, error) {
	people := make([]*personpb.Person, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		person, err := s.getPerson(ctx, id)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return &personpb.ListPeopleReply{
		People: people,
	}, nil
}

func (s *PersonService) getPerson(ctx context.Context, id int64) (*personpb.Person, error) {
	if person, exists := s.cache[id]; exists {
		return person, nil
	}
	res, err := s.cli.Person(ctx, int(id))
	if err != nil {
		return nil, err
	}
	person, err := s.toPerson(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = person
	return person, nil
}

func (s *PersonService) toPerson(id int64, p *swapi.Person) (*personpb.Person, error) {
	var (
		filmIDs     []int64
		speciesIDs  []int64
		starshipIDs []int64
		vehicleIDs  []int64
	)
	for _, url := range p.FilmURLs {
		id, err := util.FilmURLToID(url)
		if err != nil {
			return nil, err
		}
		filmIDs = append(filmIDs, id)
	}
	for _, url := range p.SpeciesURLs {
		id, err := util.SpeciesURLToID(url)
		if err != nil {
			return nil, err
		}
		speciesIDs = append(speciesIDs, id)
	}
	for _, url := range p.StarshipURLs {
		id, err := util.StarshipURLToID(url)
		if err != nil {
			return nil, err
		}
		starshipIDs = append(starshipIDs, id)
	}
	for _, url := range p.VehicleURLs {
		id, err := util.VehicleURLToID(url)
		if err != nil {
			return nil, err
		}
		vehicleIDs = append(vehicleIDs, id)
	}
	return &personpb.Person{
		Id:          id,
		Name:        p.Name,
		BirthYear:   p.BirthYear,
		EyeColor:    p.EyeColor,
		Gender:      p.Gender,
		HairColor:   p.HairColor,
		Height:      p.Height,
		Mass:        p.Mass,
		SkinColor:   p.SkinColor,
		Homeworld:   p.Homeworld,
		Url:         p.URL,
		Created:     p.Created,
		Edited:      p.Edited,
		FilmIds:     filmIDs,
		SpeciesIds:  speciesIDs,
		StarshipIds: starshipIDs,
		VehicleIds:  vehicleIDs,
	}, nil
}
