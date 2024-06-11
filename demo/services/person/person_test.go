package person_test

import (
	"context"
	"testing"

	"github.com/mercari/grpc-federation/demo/services/person"
	personpb "github.com/mercari/grpc-federation/demo/swapi/person"
)

func TestPersonService(t *testing.T) {
	svc := person.NewPersonService()
	t.Run("GetPerson", func(t *testing.T) {
		res, err := svc.GetPerson(context.Background(), &personpb.GetPersonRequest{
			Id: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if res.GetPerson() == nil {
			t.Fatalf("failed to get person: %+v", res)
		}
		if res.GetPerson().GetId() != 1 {
			t.Fatalf("failed to get id: %+v", res)
		}
	})
	t.Run("ListPersons", func(t *testing.T) {
		res, err := svc.ListPeople(context.Background(), &personpb.ListPeopleRequest{
			Ids: []int64{1, 2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.GetPeople()) != 2 {
			t.Fatalf("failed to get persons: %+v", res)
		}
	})
}
