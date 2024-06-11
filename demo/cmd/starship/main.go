package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/mercari/grpc-federation/demo/services/starship"
	starshippb "github.com/mercari/grpc-federation/demo/swapi/starship"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		return fmt.Errorf("must be specified GRPC_PORT environment")
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	starshippb.RegisterStarshipServiceServer(grpcServer, starship.NewStarshipService())
	log.Printf("listening gRPC: %s", port)
	if err := grpcServer.Serve(listener); err != nil {
		return err
	}
	return nil
}
