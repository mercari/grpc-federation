package main

import (
	"context"
	pluginpb "example/plugin"
	"fmt"

	"google.golang.org/grpc/metadata"
)

type plugin struct{}

func (_ *plugin) Example_Account_GetId(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to get metadata")
	}
	values := md["id"]
	if len(values) == 0 {
		return "", fmt.Errorf("failed to find id key in metadata")
	}
	return values[0], nil
}

func main() {
	pluginpb.RegisterAccountPlugin(&plugin{})
}
