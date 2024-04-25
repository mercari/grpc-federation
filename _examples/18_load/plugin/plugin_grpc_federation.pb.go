// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: dev
//
// source: plugin/plugin.proto
package pluginpb

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"reflect"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"google.golang.org/grpc/metadata"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

type AccountPlugin interface {
	Example_Account_GetId(context.Context) (string, error)
}

func RegisterAccountPlugin(plug AccountPlugin) {
	reader := bufio.NewReader(os.Stdin)
	for {
		content, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		if content == "" {
			continue
		}
		if content == "exit\n" {
			return
		}
		if content == "version\n" {
			b, _ := grpcfed.EncodeCELPluginVersion(grpcfed.CELPluginVersionSchema{
				ProtocolVersion:   grpcfed.CELPluginProtocolVersion,
				FederationVersion: "dev",
				Functions: []string{
					"example_account_get_id_string",
				},
			})
			_, _ = os.Stdout.Write(append(b, '\n'))
			continue
		}
		res, err := handleAccountPlugin([]byte(content), plug)
		if err != nil {
			res = grpcfed.ToErrorCELPluginResponse(err)
		}
		encoded, err := grpcfed.EncodeCELPluginResponse(res)
		if err != nil {
			continue
		}
		_, _ = os.Stdout.Write(append(encoded, '\n'))
	}
}

func handleAccountPlugin(content []byte, plug AccountPlugin) (*grpcfed.CELPluginResponse, error) {
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
	case "example_account_get_id_string":
		if len(req.GetArgs()) != 0 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 0)
		}
		ret, err := plug.Example_Account_GetId(ctx)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToStringCELPluginResponse(ret)
	}
	return nil, fmt.Errorf("unexpected method name: %s", req.GetMethod())
}
