package film_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/film"
	filmpb "github.com/mercari/grpc-federation/demo/swapi/film"
)

func TestFilmService(t *testing.T) {
	svc := film.NewFilmService()
	t.Run("GetFilm", func(t *testing.T) {
		res, err := svc.GetFilm(context.Background(), &filmpb.GetFilmRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetFilm() == nil {
			t.Fatalf("failed to get film: %+v", res)
		}
		if res.GetFilm().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
	})
	t.Run("ListFilms", func(t *testing.T) {
		res, err := svc.ListFilms(context.Background(), &filmpb.ListFilmsRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetFilms()) != 2 {
			t.Fatalf("failed to get films: %+v", res)
		}
	})
}
