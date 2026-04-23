package main

import (
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt [--debug] <id>",
	Short: "Generate a prompt for the given ticket",
	Long:  "Generate a prompt for the given ticket by executing the full merge and render pipeline. Output is written to stdout.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	promptCmd.Flags().Bool("debug", false, "Produce a detailed processing log")
	rootCmd.AddCommand(promptCmd)
}
