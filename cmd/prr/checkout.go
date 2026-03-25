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

		service := mirrorServiceFactory(cmd.ErrOrStderr())
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

		// Probe which PR refs exist before fetching to avoid git printing fatal
		// errors to stderr for refs that are unavailable (e.g. draft PRs).
		mergeRefSpec := git.RemoteMergeRefSpec(prRef.PRID)
		headRefSpec := git.RemoteHeadRefSpec(prRef.PRID)
		probeRefs := []string{mergeRefSpec, headRefSpec}
		if prRef.SourceBranch != "" {
			probeRefs = append(probeRefs, "refs/heads/"+prRef.SourceBranch)
		}
		available, probeErr := service.ProbeRemoteRefs(context.Background(), bareDir, prRef.Remote, probeRefs, commonOpts)
		if probeErr != nil {
			return probeErr
		}

		var mergeRef, baseRef string
		switch {
		case available[mergeRefSpec]:
			mergeRef, err = service.FetchPRMergeRefWithOptions(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
			if err != nil {
				return err
			}
		case available[headRefSpec]:
			if prRef.BaseSHA == "" {
				return provider.EnrichmentRequiredError(prRef.Provider)
			}
			warnf("merge ref unavailable (draft PR?), falling back to head ref with base %s", prRef.BaseSHA[:min(len(prRef.BaseSHA), 12)])
			mergeRef, err = service.FetchPRHeadRef(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
			if err != nil {
				return err
			}
			baseRef = prRef.BaseSHA
		case prRef.SourceBranch != "" && available["refs/heads/"+prRef.SourceBranch]:
			if prRef.BaseSHA == "" {
				return provider.EnrichmentRequiredError(prRef.Provider)
			}
			warnf("merge/head refs unavailable, falling back to source branch %q", prRef.SourceBranch)
			mergeRef, err = service.FetchSourceBranchRef(context.Background(), bareDir, prRef.Remote, prRef.PRID, prRef.SourceBranch, commonOpts)
			if err != nil {
				return err
			}
			baseRef = prRef.BaseSHA
		default:
			if prRef.BaseSHA == "" {
				return provider.EnrichmentRequiredError(prRef.Provider)
			}
			return apperrors.WrapProvider(fmt.Sprintf("no PR refs available on remote for PR #%d", prRef.PRID), nil)
		}

		workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, prRef.PRID)
		if err != nil {
			return err
		}

		err = service.CreateWorktree(context.Background(), bareDir, mergeRef, workDir, commonOpts)
		if err != nil {
			return err
		}

		// Refine the base to the actual common ancestor so downstream diff
		// shows only what this PR contributes.
		if baseRef != "" {
			if actualBase, mergeBaseErr := service.ResolveMergeBase(context.Background(), workDir, "HEAD", baseRef, commonOpts); mergeBaseErr == nil && actualBase != "" {
				baseRef = actualBase
			}
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
