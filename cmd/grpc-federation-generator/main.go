package main

import (
	"context"
	"log"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/mercari/grpc-federation/generator"
)

type option struct {
	Config    string `description:"specify the config path" long:"config" short:"c" default:"grpc-federation.yaml"`
	WatchMode bool   `description:"enable 'watch' mode to monitor changes in proto files and generate interactively" short:"w"`
}

const (
	exitOK  = 0
	exitERR = 1
)

func main() {
	args, opt, err := parseOpt()
	if err != nil {
		// output error message to stderr by library
		os.Exit(exitOK)
	}
	if err := _main(context.Background(), args, opt); err != nil {
		log.Printf("%+v", err)
		os.Exit(exitERR)
	}
	os.Exit(exitOK)
}

func _main(ctx context.Context, args []string, opt *option) error {
	cfg, err := generator.LoadConfig(opt.Config)
	if err != nil {
		return err
	}
	var protoPath string
	if len(args) != 0 {
		protoPath = args[0]
	}
	g := generator.New(cfg)
	var opts []generator.Option
	if opt.WatchMode {
		opts = append(opts, generator.WatchMode())
	}
	if err := g.Generate(ctx, protoPath, opts...); err != nil {
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
