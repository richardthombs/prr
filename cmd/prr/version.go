package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version   = "v0.0.0-dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var semverReleasePattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:-rc\.[0-9]+)?$`)

func resolvedVersion() string {
	if semverReleasePattern.MatchString(version) {
		return version
	}

	return "v0.0.0-dev+" + shortCommit(commit)
}

func shortCommit(value string) string {
	if value == "" || value == "unknown" {
		return "unknown"
	}

	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= 7 {
		return trimmed
	}

	return trimmed[:7]
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show PRR version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), resolvedVersion())
	},
}
