package main

import (
	"fmt"
	"os"

	"github.com/osrcmap/wallgen/cmd"
)

// Set by GoReleaser via -ldflags -X at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cmd.Root()
	root.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
