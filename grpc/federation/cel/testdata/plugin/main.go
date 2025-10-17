package main

import (
	"context"
	"fmt"
	"runtime"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"google.golang.org/grpc/metadata"
)

func main() {
	for {
		content := grpcfed.ReadPluginContent()
		if content == "exit\n" {
			return
		}
		if content == "gc\n" {
			runtime.GC()
			grpcfed.WritePluginContent([]byte{'\n'})
			continue
		}
		res, err := handle([]byte(content))
		if err != nil {
			res = grpcfed.ToErrorCELPluginResponse(err)
		}
		encoded, err := grpcfed.EncodeCELPluginResponse(res)
		if err != nil {
			panic(fmt.Sprintf("failed to encode cel plugin response: %s", err.Error()))
		}
		grpcfed.WritePluginContent(append(encoded, '\n'))
	}
}

func handle(content []byte) (res *grpcfed.CELPluginResponse, e error) {
	defer func() {
		if r := recover(); r != nil {
			res = grpcfed.ToErrorCELPluginResponse(fmt.Errorf("%v", r))
		}
	}()

	req, err := grpcfed.DecodeCELPluginRequest(content)
	if err != nil {
		return nil, err
	}
	md := make(metadata.MD)
	for _, m := range req.GetMetadata() {
		md[m.GetKey()] = m.GetValues()
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)
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
}

func foo(ctx context.Context) (int64, error) {
	return 10, nil
}
