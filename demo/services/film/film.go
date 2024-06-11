package film

import (
	"context"

	"github.com/peterhellberg/swapi"

	filmpb "github.com/mercari/grpc-federation/demo/swapi/film"
	"github.com/mercari/grpc-federation/demo/util"
)

type FilmService struct {
	*filmpb.UnimplementedFilmServiceServer
	cli   *swapi.Client
	cache map[int64]*filmpb.Film
}

func NewFilmService() *FilmService {
	return &FilmService{
		cli:   swapi.NewClient(),
		cache: make(map[int64]*filmpb.Film),
	}
}

func (s *FilmService) GetFilm(ctx context.Context, req *filmpb.GetFilmRequest) (*filmpb.GetFilmReply, error) {
	film, err := s.getFilm(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &filmpb.GetFilmReply{
		Film: film,
	}, nil
}

func (s *FilmService) ListFilms(ctx context.Context, req *filmpb.ListFilmsRequest) (*filmpb.ListFilmsReply, error) {
	films := make([]*filmpb.Film, 0, len(req.GetIds()))
	for _, id := range req.GetIds() {
		film, err := s.getFilm(ctx, id)
		if err != nil {
			return nil, err
		}
		films = append(films, film)
	}
	return &filmpb.ListFilmsReply{
		Films: films,
	}, nil
}

func (s *FilmService) getFilm(ctx context.Context, id int64) (*filmpb.Film, error) {
	if film, exists := s.cache[id]; exists {
		return film, nil
	}
	res, err := s.cli.Film(ctx, int(id))
	if err != nil {
		return nil, err
	}
	film, err := s.toFilm(id, &res)
	if err != nil {
		return nil, err
	}
	s.cache[id] = film
	return film, nil
}

func (s *FilmService) toFilm(id int64, f *swapi.Film) (*filmpb.Film, error) {
	var (
		speciesIDs   []int64
		starshipIDs  []int64
		vehicleIDs   []int64
		characterIDs []int64
		planetIDs    []int64
	)
	for _, url := range f.SpeciesURLs {
		id, err := util.SpeciesURLToID(url)
		if err != nil {
			return nil, err
		}
		speciesIDs = append(speciesIDs, id)
	}
	for _, url := range f.StarshipURLs {
		id, err := util.StarshipURLToID(url)
		if err != nil {
			return nil, err
		}
		starshipIDs = append(starshipIDs, id)
	}
	for _, url := range f.VehicleURLs {
		id, err := util.VehicleURLToID(url)
		if err != nil {
			return nil, err
		}
		vehicleIDs = append(vehicleIDs, id)
	}
	for _, url := range f.CharacterURLs {
		id, err := util.PersonURLToID(url)
		if err != nil {
			return nil, err
		}
		characterIDs = append(characterIDs, id)
	}
	for _, url := range f.PlanetURLs {
		id, err := util.PlanetURLToID(url)
		if err != nil {
			return nil, err
		}
		planetIDs = append(planetIDs, id)
	}
	return &filmpb.Film{
		Id:           id,
		Title:        f.Title,
		EpisodeId:    int32(f.EpisodeID),
		OpeningCrawl: f.OpeningCrawl,
		Director:     f.Director,
		Producer:     f.Producer,
		Url:          f.URL,
		Created:      f.Created,
		Edited:       f.Edited,
		SpeciesIds:   speciesIDs,
		StarshipIds:  starshipIDs,
		VehicleIds:   vehicleIDs,
		CharacterIds: characterIDs,
		PlanetIds:    planetIDs,
	}, nil
}
