package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/mercari/grpc-federation/lsp/server"
)

type option struct {
	LogFile     string   `description:"specify the log file path for debugging" long:"log-file"`
	ImportPaths []string `description:"specify the import path for loading proto file" long:"import-path" short:"I"`
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("[grpc-federation-language-server]: %+v", err)
	}
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

func run(ctx context.Context) error {
	_, opt, err := parseOpt()
	if err != nil {
		// output error message to stderr by library
		return nil
	}

	var opts []server.ServerOption
	if opt.LogFile != "" {
		logFile, err := os.OpenFile(opt.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", opt.LogFile, err)
		}
		defer logFile.Close()
		opts = append(opts, server.LogFileOption(logFile))
	}
	if len(opt.ImportPaths) != 0 {
		opts = append(opts, server.ImportPathsOption(opt.ImportPaths))
	}
	server := server.New(opts...)
	server.Run(ctx)
	return nil
}
