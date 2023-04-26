package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Strings are populated by Goreleaser
var (
	version = "snapshot"
	commit  = "unknown"
	date    = "unknown"

	rootCommand = &cobra.Command{
		Use:   "control-api {server|controller}",
		Short: "An aggregated API and controller for APPUiO.",
	}
)

func main() {
	rootCommand.AddCommand(ControllerCommand(), APICommand(), CleanupCommand())

	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
