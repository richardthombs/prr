package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/richardthombs/prr/internal/bundle"
	"github.com/richardthombs/prr/internal/config"
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
	reviewCmd.Flags().String("model", "", "Copilot model to use for review generation")
	reviewCmd.Flags().Bool("json", false, "Emit structured JSON output instead of Markdown")
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
		model, err := cmd.Flags().GetString("model")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse model flag", err)
		}

		useCheckoutContext := arg == "" && hasAuthoritativeCheckoutContext(stdinInput)

		prRef := types.PRRef{}
		if useCheckoutContext {
			prRef = types.PRRef{
				PRID:     stdinInput.PRID,
				RepoURL:  strings.TrimSpace(stdinInput.RepoURL),
				Remote:   strings.TrimSpace(stdinInput.Remote),
				Provider: strings.TrimSpace(stdinInput.Provider),
			}
		} else {
			resolver := provider.NewResolver(provider.NewDefaultProvider())
			prRef, err = resolveReviewPRRef(context.Background(), resolver, arg, resolveOpts, stdinInput, hasStdinInput)
			if err != nil {
				return err
			}
		}

		warnf := func(format string, args ...any) {
			if verbose || whatIf {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] warning: "+format+"\n", args...)
			}
		}
		providerClient := provider.NewDefaultProvider()
		if !useCheckoutContext {
			prRef = provider.EnrichPRRef(context.Background(), prRef, prEnricherFactory(), warnf)
		}

		service := mirrorServiceFactory(cmd.ErrOrStderr())
		commonOpts := git.EnsureOptions{
			Verbose: verbose || whatIf,
			WhatIf:  whatIf,
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		}

		bareDir := strings.TrimSpace(stdinInput.BareDir)
		mergeRef := strings.TrimSpace(stdinInput.MergeRef)
		baseRef := strings.TrimSpace(stdinInput.BaseRef)
		workDir := strings.TrimSpace(stdinInput.WorkDir)

		if useCheckoutContext {
			if verbose || whatIf {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] using checkout JSON context from stdin; skipping resolve/mirror/fetch/worktree setup")
			}

			if stdinInput.Cleanup && !keep {
				defer func() {
					cleanupErr := service.CleanupWorktree(context.Background(), bareDir, workDir, commonOpts)
					if cleanupErr != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] warning: cleanup failed: %v\n", cleanupErr)
					}
				}()
			}
		} else {
			bareDir, err = service.EnsureMirrorWithOptions(context.Background(), prRef.RepoURL, commonOpts)
			if err != nil {
				return err
			}
			mergeRef, err = service.FetchPRMergeRefWithOptions(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
			if err != nil {
				if prRef.BaseSHA == "" {
					return provider.EnrichmentRequiredError(prRef.Provider)
				}
				warnf("merge ref unavailable (closed PR?), falling back to head ref with base %s", prRef.BaseSHA[:min(len(prRef.BaseSHA), 12)])
				headRef, headErr := service.FetchPRHeadRef(context.Background(), bareDir, prRef.Remote, prRef.PRID, commonOpts)
				if headErr != nil {
					return headErr
				}
				mergeRef = headRef
				baseRef = prRef.BaseSHA
			}
			workDir, err = service.ResolveWorktreeDirFromBareDir(bareDir, prRef.PRID)
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
		}

		diffOutput, err := service.DiffContributionWithOptions(context.Background(), workDir, baseRef, commonOpts)
		if err != nil {
			return err
		}

		diffOutput.PRID = prRef.PRID
		diffOutput.RepoURL = prRef.RepoURL
		diffOutput.Remote = prRef.Remote
		diffOutput.Provider = prRef.Provider
		diffOutput.BareDir = bareDir
		diffOutput.MergeRef = mergeRef
		diffOutput.BaseRef = baseRef
		diffOutput.WorkDir = workDir
		issues, err := providerClient.DiscoverIssues(context.Background(), prRef, issueRunnerFactory())
		if err != nil {
			return err
		}
		diffOutput.Issues = issues

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

		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] starting PR review...\n")
		reviewOutput, err := reviewEngineFactory().Review(context.Background(), engine.ReviewInput{
			Bundle:             bundlePayload,
			WorkDir:            workDir,
			Model:              model,
			Verbose:            verbose,
			WhatIf:             whatIf,
			ReviewInstructions: loadReviewInstructions(),
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		})
		if err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				return err
			}

			return apperrors.WrapEngine("failed to run review engine", err)
		}

		emitJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse json flag", err)
		}

		if emitJSON {
			encoded, err := json.Marshal(reviewOutput)
			if err != nil {
				return apperrors.WrapRuntime("failed to encode review JSON", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(encoded)); err != nil {
				return apperrors.WrapRuntime("failed to write output", err)
			}
		} else {
			prURL := prRef.PRURL
			if prURL == "" {
				prURL = buildPRURL(prRef.RepoURL, prRef.PRID)
			}
			markdown := renderMarkdown(reviewOutput, prRef.PRID, prURL, diffOutput.Issues)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), markdown); err != nil {
				return apperrors.WrapRuntime("failed to write markdown output", err)
			}
		}

		return nil
	},
}

func hasAuthoritativeCheckoutContext(input checkoutOutput) bool {
	return input.PRID > 0 &&
		strings.TrimSpace(input.RepoURL) != "" &&
		strings.TrimSpace(input.Remote) != "" &&
		strings.TrimSpace(input.Provider) != "" &&
		strings.TrimSpace(input.BareDir) != "" &&
		strings.TrimSpace(input.MergeRef) != "" &&
		strings.TrimSpace(input.WorkDir) != ""
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

func loadReviewInstructions() string {
	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return config.DefaultReviewInstructions
	}

	return config.ResolveReviewInstructions(userCfg)
}
