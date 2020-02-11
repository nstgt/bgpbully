package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nstgt/bgpdbully/internal/pkg/bgpdbully"
)

func main() {
	f := flag.String("f", "", "config file")
	flag.Parse()
	if *f == "" {
		fmt.Println("Usage: bgpdbully -f configfile")
		os.Exit(1)
	}

	bgpdbully.Run(f)
}
