package main

import (
	"fmt"
	"log"
	"os"
	"timefind/pkg/config"
	"timefind/pkg/index"

	"github.com/pborman/getopt"
)

//
// Command-line arguments
//
// TODO should probably put these in a struct?
var configPaths []string = []string{}
var verbose bool = false

func main() {
	getopt.ListVarLong(&configPaths, "config", 'c',
		"REQUIRED: Path to configuration file (can be used multiple times)", "PATH")
	getopt.BoolVarLong(&verbose, "verbose", 'v', "Verbose progress indicators and messages")
	help := getopt.BoolLong("help", 'h', "Show this help message and exit")
	version := getopt.BoolLong("version", 0, "Prints the version")
	getopt.SetParameters("")
	getopt.Parse()

	getopt.SetUsage(func() {
		fmt.Fprintf(os.Stderr, "timefind-indexer v%s\n", IndexerVersion)
		getopt.PrintUsage(os.Stderr)
	})

	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Fprintf(os.Stderr, "timefind-indexer v%s\n", IndexerVersion)
		os.Exit(0)
	}

	if len(configPaths) == 0 {
		fmt.Fprintf(os.Stderr, "error: no configuration (-c/--config) found\n")
		getopt.Usage()
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	for _, configPath := range configPaths {

		cfg, err := config.NewConfiguration(configPath)
		if err != nil {
			log.Fatal(err)
			continue
		}

		idx, err := index.NewIndex(cfg)
		if err != nil {
			log.Print(err)
			return
		}

		if err := idx.Update(); err != nil {
			log.Print(err)
			return
		}

		if err := idx.WriteOut(); err != nil {
			log.Print(err)
			return
		}
	}
}

// vim: noet:ts=4:sw=4:tw=80
