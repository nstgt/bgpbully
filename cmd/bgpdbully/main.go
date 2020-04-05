package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/nstgt/bgpdbully/internal/pkg/bgpdbully"
)

func main() {
	var opts struct {
		ConfigFile string `short:"f" long:"config-file" description:"specifying a config file"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	bgpdbully.Run(opts.ConfigFile)
}
