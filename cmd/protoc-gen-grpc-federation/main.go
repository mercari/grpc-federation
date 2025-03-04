package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/mercari/grpc-federation/generator"
	grpcfed "github.com/mercari/grpc-federation/grpc/federation"
)

type option struct {
	Version bool `description:"show the version" long:"version" short:"v"`
}

const (
	exitOK  = 0
	exitERR = 1
)

func main() {
	_, opt, err := parseOpt()
	if err != nil {
		// output error message to stderr by library
		os.Exit(exitOK)
	}
	if opt.Version {
		fmt.Fprintln(os.Stdout, grpcfed.Version)
		return
	}
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	req, err := parseRequest(os.Stdin)
	if err != nil {
		return err
	}
	resp, err := generator.CreateCodeGeneratorResponse(context.Background(), req)
	if err != nil {
		return err
	}
	if resp == nil {
		return nil
	}
	if err := outputResponse(resp); err != nil {
		return err
	}
	return nil
}

func parseRequest(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func outputResponse(resp *pluginpb.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err = os.Stdout.Write(buf); err != nil {
		return err
	}
	return nil
}

func parseOpt() ([]string, *option, error) {
	var opt option
	parser := flags.NewParser(&opt, flags.Default)
	args, err := parser.Parse()
	if err != nil {
		return nil, nil, err
	}
	return args, &opt, nil
}
