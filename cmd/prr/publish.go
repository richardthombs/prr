package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(publishCmd)
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish review output",
	Long:  "Publish previously generated review output back to the pull request when provider support is enabled.",
}
