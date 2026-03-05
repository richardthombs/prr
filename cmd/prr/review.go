package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/richardthombs/prr/internal/bundle"
	"github.com/richardthombs/prr/internal/engine"
	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/richardthombs/prr/internal/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reviewCmd)
	reviewCmd.Flags().String("provider", "", "Override PR provider")
	reviewCmd.Flags().String("repo", "", "Override repository URL")
	reviewCmd.Flags().String("remote", "", "Override git remote name")
	reviewCmd.Flags().Bool("keep", false, "Retain worktree after review completion")
	reviewCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	reviewCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
	reviewCmd.Flags().Int("max-patch-bytes", 0, "Maximum allowed patch size in bytes (0 disables limit)")
	reviewCmd.Flags().Int("max-files", 0, "Maximum allowed changed file count (0 disables limit)")
}

var reviewEngineFactory = func() engine.ReviewEngine {
	return engine.NewDefaultAdapter()
}

var reviewCmd = &cobra.Command{
	Use:   "review [PR_ID|PR_URL]",
	Short: "Run an end-to-end PR review",
	Long:  "Run an end-to-end PR review for a pull request using configured providers and engine adapters.",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) > 1 {
			return apperrors.WrapConfig("invalid arguments", fmt.Errorf("usage: prr review [PR_ID|PR_URL]"))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stdinInput := checkoutOutput{}
		hasStdinInput, err := readInputJSON(cmd, &stdinInput)
		if err != nil {
			return err
		}

		resolveOpts, err := resolveFlags(cmd)
		if err != nil {
			return err
		}

		arg := ""
		if len(args) == 1 {
			arg = strings.TrimSpace(args[0])
		}

		resolver := provider.NewResolver(provider.NewDefaultProvider())
		prRef, err := resolveReviewPRRef(context.Background(), resolver, arg, resolveOpts, stdinInput, hasStdinInput)
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
		maxPatchBytes, err := cmd.Flags().GetInt("max-patch-bytes")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse max-patch-bytes flag", err)
		}
		maxFiles, err := cmd.Flags().GetInt("max-files")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse max-files flag", err)
		}

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
		if err != nil {
			return err
		}
		workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, prRef.PRID)
		if err != nil {
			return err
		}

		if err := service.CreateWorktree(context.Background(), bareDir, mergeRef, workDir, commonOpts); err != nil {
			return err
		}

		if !keep {
			defer func() {
				cleanupErr := service.CleanupWorktree(context.Background(), bareDir, workDir, commonOpts)
				if cleanupErr != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] warning: cleanup failed: %v\n", cleanupErr)
				}
			}()
		}

		diffOutput, err := service.DiffContributionWithOptions(context.Background(), workDir, commonOpts)
		if err != nil {
			return err
		}

		diffOutput.PRID = prRef.PRID
		diffOutput.RepoURL = prRef.RepoURL
		diffOutput.Remote = prRef.Remote
		diffOutput.Provider = prRef.Provider
		diffOutput.BareDir = bareDir
		diffOutput.MergeRef = mergeRef
		diffOutput.WorkDir = workDir

		if whatIf {
			if strings.TrimSpace(diffOutput.Stat) == "" {
				diffOutput.Stat = "what-if: diff stat not executed"
			}
			if strings.TrimSpace(diffOutput.Patch) == "" {
				diffOutput.Patch = "what-if: unified patch not executed"
			}
		}

		bundlePayload, err := bundle.BuildV1(diffOutput, bundle.Limits{
			MaxPatchBytes:   maxPatchBytes,
			MaxChangedFiles: maxFiles,
		})
		if err != nil {
			return err
		}
		if err := bundle.ValidateV1Schema(bundlePayload); err != nil {
			return err
		}

		reviewOutput, err := reviewEngineFactory().Review(context.Background(), bundlePayload)
		if err != nil {
			return apperrors.WrapEngine("failed to run review engine", err)
		}

		validatedReview, err := types.NormalizeAndValidateReview(reviewOutput)
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(validatedReview)
		if err != nil {
			return apperrors.WrapRuntime("failed to encode review JSON", err)
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(encoded)); err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}

func resolveReviewPRRef(
	ctx context.Context,
	resolver *provider.Resolver,
	arg string,
	resolveOpts resolveOptions,
	stdinInput checkoutOutput,
	hasStdinInput bool,
) (types.PRRef, error) {
	if arg != "" {
		if looksLikePRURL(arg) {
			return resolver.ResolveFromPullRequestURL(ctx, arg, provider.ResolveOptions{
				Provider: resolveOpts.Provider,
				RepoURL:  resolveOpts.RepoURL,
				Remote:   resolveOpts.Remote,
			})
		}

		prID, err := strconv.Atoi(arg)
		if err != nil || prID <= 0 {
			return types.PRRef{}, apperrors.WrapConfig("argument must be a PR_ID integer or PR_URL", err)
		}

		return resolver.Resolve(ctx, prID, provider.ResolveOptions{
			Provider: firstNonEmpty(resolveOpts.Provider, stdinInput.Provider),
			RepoURL:  firstNonEmpty(resolveOpts.RepoURL, stdinInput.RepoURL),
			Remote:   firstNonEmpty(resolveOpts.Remote, stdinInput.Remote),
		})
	}

	if !hasStdinInput {
		return types.PRRef{}, apperrors.WrapConfig("review requires PR_ID, PR_URL, or checkout-style JSON on stdin", nil)
	}

	if stdinInput.PRID <= 0 {
		return types.PRRef{}, apperrors.WrapConfig("stdin JSON must include prId", nil)
	}

	return resolver.Resolve(ctx, stdinInput.PRID, provider.ResolveOptions{
		Provider: firstNonEmpty(resolveOpts.Provider, stdinInput.Provider),
		RepoURL:  firstNonEmpty(resolveOpts.RepoURL, stdinInput.RepoURL),
		Remote:   firstNonEmpty(resolveOpts.Remote, stdinInput.Remote),
	})
}

func looksLikePRURL(value string) bool {
	parsedURL, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func firstNonEmpty(primary, fallback string) string {
	trimmedPrimary := strings.TrimSpace(primary)
	if trimmedPrimary != "" {
		return trimmedPrimary
	}

	return strings.TrimSpace(fallback)
}
