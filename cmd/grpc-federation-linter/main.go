package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
)

type option struct {
	ImportPaths []string `description:"specify the import path for loading proto file" long:"import-path" short:"I"`
}

const (
	exitOK  = 0
	exitERR = 1
)

func main() {
	args, opt, err := parseOpt()
	if err != nil {
		// output error message to stderr by library
		os.Exit(exitERR)
	}
	if len(args) == 0 {
		log.Fatal("[grpc-federation-linter]: required file path to validation")
	}
	ctx := context.Background()
	var existsErr bool
	for _, arg := range args {
		path := arg
		f, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("[grpc-federation-linter]: failed to read file %s: %v", path, err)
		}
		file, err := source.NewFile(path, f)
		if err != nil {
			log.Fatalf("[grpc-federation-linter]: %v", err)
		}

		v := validator.New()
		outs := v.Validate(ctx, file, validator.ImportPathOption(opt.ImportPaths...))
		if len(outs) == 0 {
			continue
		}
		fmt.Println(validator.Format(outs))
		if validator.ExistsError(outs) {
			existsErr = true
		}
	}
	if existsErr {
		os.Exit(exitERR)
	}
	os.Exit(exitOK)
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
