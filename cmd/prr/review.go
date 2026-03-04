package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(reviewCmd)
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Run an end-to-end PR review",
	Long:  "Run an end-to-end PR review for a pull request using configured providers and engine adapters.",
}
