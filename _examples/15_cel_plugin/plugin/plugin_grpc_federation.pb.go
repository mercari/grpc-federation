// Code generated by protoc-gen-grpc-federation. DO NOT EDIT!
// versions:
//
//	protoc-gen-grpc-federation: (devel)
//
// source: plugin/plugin.proto
package pluginpb

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"

	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
	"google.golang.org/grpc/metadata"
)

var (
	_ = reflect.Invalid // to avoid "imported and not used error"
)

type RegexpPlugin interface {
	Example_Regexp_Compile(context.Context, string) (*Regexp, error)
	Example_Regexp_NewExample(context.Context) (*Example, error)
	Example_Regexp_NewExamples(context.Context) ([]*Example, error)
	Example_Regexp_FilterExamples(context.Context, []*Example) ([]*Example, error)
	Example_Regexp_Regexp_MatchString(context.Context, *Regexp, string) (bool, error)
	Example_Regexp_Example_Concat(context.Context, *Example, []string) (string, error)
	Example_Regexp_Example_MySplit(context.Context, *Example, string, string) ([]string, error)
}

func RegisterRegexpPlugin(plug RegexpPlugin) {
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
		if content == "gc\n" {
			runtime.GC()
			continue
		}
		if content == "version\n" {
			b, _ := grpcfed.EncodeCELPluginVersion(grpcfed.CELPluginVersionSchema{
				ProtocolVersion:   grpcfed.CELPluginProtocolVersion,
				FederationVersion: "(devel)",
				Functions: []string{
					"example_regexp_compile_string_example_regexp_Regexp",
					"example_regexp_newExample_example_regexp_Example",
					"example_regexp_newExamples_repeated example_regexp_Example",
					"example_regexp_filterExamples_repeated example_regexp_Example_repeated example_regexp_Example",
					"example_regexp_Regexp_matchString_example_regexp_Regexp_string_bool",
					"example_regexp_Example_concat_example_regexp_Example_repeated string_string",
					"example_regexp_Example_mySplit_example_regexp_Example_string_string_repeated string",
				},
			})
			_, _ = os.Stdout.Write(append(b, '\n'))
			continue
		}
		res, err := handleRegexpPlugin([]byte(content), plug)
		if err != nil {
			res = grpcfed.ToErrorCELPluginResponse(err)
		}
		encoded, err := grpcfed.EncodeCELPluginResponse(res)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal error: failed to encode cel plugin response: %s\n", err.Error())
			os.Exit(1)
		}
		_, _ = os.Stdout.Write(append(encoded, '\n'))
	}
}

func handleRegexpPlugin(content []byte, plug RegexpPlugin) (*grpcfed.CELPluginResponse, error) {
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
	case "example_regexp_compile_string_example_regexp_Regexp":
		if len(req.GetArgs()) != 1 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 1)
		}
		arg0, err := grpcfed.ToString(req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Regexp_Compile(ctx, arg0)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToMessageCELPluginResponse[*Regexp](ret)
	case "example_regexp_newExample_example_regexp_Example":
		if len(req.GetArgs()) != 0 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 0)
		}
		ret, err := plug.Example_Regexp_NewExample(ctx)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToMessageCELPluginResponse[*Example](ret)
	case "example_regexp_newExamples_repeated example_regexp_Example":
		if len(req.GetArgs()) != 0 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 0)
		}
		ret, err := plug.Example_Regexp_NewExamples(ctx)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToMessageListCELPluginResponse[*Example](ret)
	case "example_regexp_filterExamples_repeated example_regexp_Example_repeated example_regexp_Example":
		if len(req.GetArgs()) != 1 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 1)
		}
		arg0, err := grpcfed.ToMessageList[*Example](req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Regexp_FilterExamples(ctx, arg0)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToMessageListCELPluginResponse[*Example](ret)
	case "example_regexp_Regexp_matchString_example_regexp_Regexp_string_bool":
		if len(req.GetArgs()) != 2 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 2)
		}
		arg0, err := grpcfed.ToMessage[*Regexp](req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		arg1, err := grpcfed.ToString(req.GetArgs()[1])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Regexp_Regexp_MatchString(ctx, arg0, arg1)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToBoolCELPluginResponse(ret)
	case "example_regexp_Example_concat_example_regexp_Example_repeated string_string":
		if len(req.GetArgs()) != 2 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 2)
		}
		arg0, err := grpcfed.ToMessage[*Example](req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		arg1, err := grpcfed.ToStringList(req.GetArgs()[1])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Regexp_Example_Concat(ctx, arg0, arg1)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToStringCELPluginResponse(ret)
	case "example_regexp_Example_mySplit_example_regexp_Example_string_string_repeated string":
		if len(req.GetArgs()) != 3 {
			return nil, fmt.Errorf("%s: invalid argument number: %d. expected number is %d", req.GetMethod(), len(req.GetArgs()), 3)
		}
		arg0, err := grpcfed.ToMessage[*Example](req.GetArgs()[0])
		if err != nil {
			return nil, err
		}
		arg1, err := grpcfed.ToString(req.GetArgs()[1])
		if err != nil {
			return nil, err
		}
		arg2, err := grpcfed.ToString(req.GetArgs()[2])
		if err != nil {
			return nil, err
		}
		ret, err := plug.Example_Regexp_Example_MySplit(ctx, arg0, arg1, arg2)
		if err != nil {
			return nil, err
		}
		return grpcfed.ToStringListCELPluginResponse(ret)
	}
	return nil, fmt.Errorf("unexpected method name: %s", req.GetMethod())
}
