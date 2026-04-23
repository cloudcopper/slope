package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "slope",
	Short:         "Slope project management tool",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
