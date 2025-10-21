package main

import (
	"context"
	"fmt"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

func main() {
	grpcfed.PluginMainLoop(
		grpcfed.CELPluginVersionSchema{},
		func(ctx context.Context, req *grpcfed.CELPluginRequest) (*grpcfed.CELPluginResponse, error) {
			switch req.GetMethod() {
			case "foo":
				if len(req.GetArgs()) != 0 {
					return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 0)
				}
				ret, err := foo(ctx)
				if err != nil {
					return nil, err
				}
				return grpcfed.ToInt64CELPluginResponse(ret)
			}
			return nil, fmt.Errorf("unexpected method name: %s", req.GetMethod())
		},
	)
}

func foo(ctx context.Context) (int64, error) {
	return 10, nil
}
