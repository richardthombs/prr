package main

import (
	"context"
	"fmt"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pwdCmd)
	pwdCmd.Flags().String("provider", "", "Override PR provider")
	pwdCmd.Flags().String("repo", "", "Override repository URL")
	pwdCmd.Flags().String("remote", "", "Override git remote name")
}

var pwdCmd = &cobra.Command{
	Use:   "pwd <PR_URL>",
	Short: "Print the path to the PR's git worktree",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return apperrors.WrapConfig("invalid arguments", fmt.Errorf("usage: prr pwd <PR_URL>"))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		prURL := args[0]

		resolveOpts, err := resolveFlags(cmd)
		if err != nil {
			return err
		}

		resolver := provider.NewResolver(provider.NewDefaultProvider())
		prRef, err := resolver.ResolveFromPullRequestURL(context.Background(), prURL, provider.ResolveOptions{
			Provider: resolveOpts.Provider,
			RepoURL:  resolveOpts.RepoURL,
			Remote:   resolveOpts.Remote,
		})
		if err != nil {
			return err
		}

		service := mirrorServiceFactory()

		bareDir, err := service.ResolveMirrorDir(prRef.RepoURL)
		if err != nil {
			return err
		}

		workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, prRef.PRID)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), workDir)

		return nil
	},
}
