package main

import (
	"fmt"
	"os"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

type VersionCommand struct {
}

func (c *VersionCommand) Execute(_ []string) error {
	fmt.Fprintln(os.Stdout, grpcfed.Version)
	return nil
}
