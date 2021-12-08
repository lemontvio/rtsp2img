package flag

import (
	"os"
)

func parse(args []string) Options {
	var options Options

	if len(args) < 2 {
		options.Done = true
		return options
	}

	if args[1] == "-h" || args[1] == "-?" {
		options.Help = true
	}

	if args[1] == "-v" {
		options.Version = true
	}

	for index, line := range args[1:] {
		if line == "-c" && index+2 < len(args) {
			options.Config = args[index+2]
			continue
		}

		if line == "-d" {
			options.Damon = true
		}
	}

	return options
}

func Parse() string {
	options := parse(os.Args)
	if options.Done {
		Usage()
		os.Exit(1)
	}

	if options.Help {
		Help()
		os.Exit(0)
	}

	if options.Version {
		Version()
		os.Exit(0)
	}

	if options.Config == "" {
		Usage()
		os.Exit(1)
	}

	return options.Config
}
