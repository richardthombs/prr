package main

import (
	"context"
	"encoding/json"
	"fmt"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/spf13/cobra"
)

type resolveOptions struct {
	Provider string
	RepoURL  string
	Remote   string
}

func init() {
	rootCmd.AddCommand(resolveCmd)
	resolveCmd.Flags().String("provider", "", "Override PR provider")
	resolveCmd.Flags().String("repo", "", "Override repository URL")
	resolveCmd.Flags().String("remote", "", "Override git remote name")
}

var resolveCmd = &cobra.Command{
	Use:   "resolve <PR_URL>",
	Short: "Resolve PR reference context",
	Long:  "Resolve PR reference context from a pull request URL and emit a deterministic PRRef JSON payload.",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return apperrors.WrapConfig("invalid arguments", fmt.Errorf("usage: prr resolve <PR_URL>"))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		prURL := args[0]

		opts, err := resolveFlags(cmd)
		if err != nil {
			return err
		}

		resolver := provider.NewResolver(provider.NewDefaultProvider())
		prRef, err := resolver.ResolveFromPullRequestURL(context.Background(), prURL, provider.ResolveOptions{
			Provider: opts.Provider,
			RepoURL:  opts.RepoURL,
			Remote:   opts.Remote,
		})
		if err != nil {
			return err
		}

		payload, err := json.Marshal(prRef)
		if err != nil {
			return apperrors.WrapRuntime("failed to encode PRRef JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}

func resolveFlags(cmd *cobra.Command) (resolveOptions, error) {
	providerValue, err := cmd.Flags().GetString("provider")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse provider flag", err)
	}
	repoValue, err := cmd.Flags().GetString("repo")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse repo flag", err)
	}
	remoteValue, err := cmd.Flags().GetString("remote")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse remote flag", err)
	}

	return resolveOptions{
		Provider: providerValue,
		RepoURL:  repoValue,
		Remote:   remoteValue,
	}, nil
}