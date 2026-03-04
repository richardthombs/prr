package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "prr",
	Short: "PRR automates pull request review",
	Long:  "PRR automates pull request review by preparing a deterministic review bundle and sending it to a review engine.",
}

func Execute() error {
	return rootCmd.Execute()
}
