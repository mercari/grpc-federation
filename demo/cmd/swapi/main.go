package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	filmpb "github.com/mercari/grpc-federation/demo/swapi/film"
	personpb "github.com/mercari/grpc-federation/demo/swapi/person"
	planetpb "github.com/mercari/grpc-federation/demo/swapi/planet"
	speciespb "github.com/mercari/grpc-federation/demo/swapi/species"
	starshippb "github.com/mercari/grpc-federation/demo/swapi/starship"
	swapipb "github.com/mercari/grpc-federation/demo/swapi/swapi"
	vehiclepb "github.com/mercari/grpc-federation/demo/swapi/vehicle"
)

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

type client struct{}

func (c *client) Swapi_Film_FilmServiceClient(_ swapipb.SWAPIClientConfig) (filmpb.FilmServiceClient, error) {
	ep := os.Getenv("FILM_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return filmpb.NewFilmServiceClient(conn), nil
}

func (c *client) Swapi_Person_PersonServiceClient(_ swapipb.SWAPIClientConfig) (personpb.PersonServiceClient, error) {
	ep := os.Getenv("PERSON_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return personpb.NewPersonServiceClient(conn), nil
}

func (c *client) Swapi_Planet_PlanetServiceClient(_ swapipb.SWAPIClientConfig) (planetpb.PlanetServiceClient, error) {
	ep := os.Getenv("PLANET_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return planetpb.NewPlanetServiceClient(conn), nil
}

func (c *client) Swapi_Species_SpeciesServiceClient(_ swapipb.SWAPIClientConfig) (speciespb.SpeciesServiceClient, error) {
	ep := os.Getenv("SPECIES_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return speciespb.NewSpeciesServiceClient(conn), nil
}

func (c *client) Swapi_Starship_StarshipServiceClient(_ swapipb.SWAPIClientConfig) (starshippb.StarshipServiceClient, error) {
	ep := os.Getenv("STARSHIP_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return starshippb.NewStarshipServiceClient(conn), nil
}

func (c *client) Swapi_Vehicle_VehicleServiceClient(_ swapipb.SWAPIClientConfig) (vehiclepb.VehicleServiceClient, error) {
	ep := os.Getenv("VEHICLE_SERVICE_ENDPOINT")
	conn, err := grpc.DialContext(
		context.Background(), ep,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, err
	}
	return vehiclepb.NewVehicleServiceClient(conn), nil
}

func run(ctx context.Context) error {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		return fmt.Errorf("must be specified GRPC_PORT environment")
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		return err
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("swapi"),
				semconv.ServiceVersionKey.String("1.0.0"),
				attribute.String("environment", "dev"),
			),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer tp.Shutdown(ctx)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tp)

	grpcServer := grpc.NewServer()
	server, err := swapipb.NewSWAPI(swapipb.SWAPIConfig{
		Client: new(client),
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	})
	if err != nil {
		return err
	}
	reflection.Register(grpcServer)
	swapipb.RegisterSWAPIServer(grpcServer, server)
	log.Printf("listening gRPC: %s", port)
	if err := grpcServer.Serve(listener); err != nil {
		return err
	}
	return nil
}
