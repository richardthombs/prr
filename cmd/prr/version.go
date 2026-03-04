package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show PRR version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), version)
	},
}
