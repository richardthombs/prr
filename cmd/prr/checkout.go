package main

import (
	"context"
	"encoding/json"
	"fmt"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().String("provider", "", "Override PR provider")
	checkoutCmd.Flags().String("repo", "", "Override repository URL")
	checkoutCmd.Flags().String("remote", "", "Override git remote name")
	checkoutCmd.Flags().Bool("keep", false, "Retain worktree after review chain completion")
	checkoutCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	checkoutCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout <PR_URL>",
	Short: "Resolve, mirror, fetch, and prepare worktree in one step",
	Long:  "Run resolve, mirror ensure, prref fetch, and worktree add as one composable operation and emit the final workspace JSON payload.",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return apperrors.WrapConfig("invalid arguments", fmt.Errorf("usage: prr checkout <PR_URL>"))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		prURL := args[0]

		resolveOpts, err := resolveFlags(cmd)
		if err != nil {
			return err
		}

		keep, err := cmd.Flags().GetBool("keep")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse keep flag", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse verbose flag", err)
		}

		whatIf, err := readWhatIfFlag(cmd)
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

		warnf := func(format string, args ...any) {
			if verbose || whatIf {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] warning: "+format+"\n", args...)
			}
		}
		prRef = provider.EnrichPRRef(context.Background(), prRef, prEnricherFactory(), warnf)

		service := mirrorServiceFactory()
		commonOpts := git.EnsureOptions{
			Verbose: verbose || whatIf,
			WhatIf:  whatIf,
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		}

		bareDir, err := service.EnsureMirrorWithOptions(context.Background(), prRef.RepoURL, commonOpts)
		if err != nil {
			return err
		}

		mergeRef, err := service.FetchPRMergeRefWithOptions(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
		var baseRef string
		if err != nil {
			if prRef.BaseSHA == "" {
				return provider.EnrichmentRequiredError(prRef.Provider)
			}
			headRef, headErr := service.FetchPRHeadRef(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
			if headErr != nil {
				return headErr
			}
			mergeRef = headRef
			baseRef = prRef.BaseSHA
		}

		workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, prRef.PRID)
		if err != nil {
			return err
		}

		err = service.CreateWorktree(context.Background(), bareDir, mergeRef, workDir, commonOpts)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(checkoutOutput{
			PRID:     prRef.PRID,
			RepoURL:  prRef.RepoURL,
			Remote:   prRef.Remote,
			Provider: prRef.Provider,
			BareDir:  bareDir,
			MergeRef: mergeRef,
			BaseRef:  baseRef,
			WorkDir:  workDir,
			Keep:     keep,
			Cleanup:  !keep,
		})
		if err != nil {
			return apperrors.WrapRuntime("failed to encode checkout JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}
