package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type option struct {
	Version VersionCommand `description:"print version" command:"version"`
	Plugin  PluginCommand  `description:"manage plugin" command:"plugin"`
}

const (
	exitOK  = 0
	exitERR = 1
)

var opt option

func main() {
	parser := flags.NewParser(&opt, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if !flags.WroteHelp(err) {
			os.Exit(exitERR)
		}
	}
	os.Exit(exitOK)
}
