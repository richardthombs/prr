package main

import (
	"context"
	"encoding/json"
	"fmt"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/spf13/cobra"
)

type prrefFetchOutput struct {
	PRID     int    `json:"prId"`
	RepoURL  string `json:"repoUrl,omitempty"`
	Remote   string `json:"remote"`
	Provider string `json:"provider,omitempty"`
	BareDir  string `json:"bareDir"`
	MergeRef string `json:"mergeRef"`
}

func init() {
	rootCmd.AddCommand(prrefCmd)
	prrefCmd.AddCommand(prrefFetchCmd)

	prrefFetchCmd.Flags().Int("pr-id", 0, "Pull request ID")
	prrefFetchCmd.Flags().String("repo", "", "Repository URL")
	prrefFetchCmd.Flags().String("remote", "origin", "Git remote name")
	prrefFetchCmd.Flags().String("provider", "", "PR provider")
	prrefFetchCmd.Flags().String("bare-dir", "", "Explicit bare mirror directory; defaults to deterministic repo mirror path")
	prrefFetchCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	prrefFetchCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
}

var prrefCmd = &cobra.Command{
	Use:   "prref",
	Short: "PR reference operations",
	Long:  "PR reference operations for fetching deterministic merge refs into the PRR namespace.",
}

var prrefFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch PR merge ref into PRR namespace",
	Long:  "Fetch the provider PR merge ref into refs/prr/pull/<PR_ID>/merge and emit JSON including mergeRef.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		prID, err := cmd.Flags().GetInt("pr-id")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse pr-id flag", err)
		}

		repoURL, err := cmd.Flags().GetString("repo")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse repo flag", err)
		}

		remote, err := cmd.Flags().GetString("remote")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse remote flag", err)
		}

		providerName, err := cmd.Flags().GetString("provider")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse provider flag", err)
		}

		bareDir, err := cmd.Flags().GetString("bare-dir")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse bare-dir flag", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse verbose flag", err)
		}

		whatIf, err := cmd.Flags().GetBool("what-if")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse what-if flag", err)
		}

		service := mirrorServiceFactory()
		if bareDir == "" {
			resolvedDir, resolveErr := service.ResolveMirrorDir(repoURL)
			if resolveErr != nil {
				return resolveErr
			}
			bareDir = resolvedDir
		}

		if verbose {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] prref fetch: fetching pull/%d/merge from %s into PRR namespace\n", prID, remote)
			if whatIf {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] prref fetch: what-if enabled, no external commands will be executed")
			}
		}

		mergeRef, err := service.FetchPRMergeRefWithOptions(context.Background(), bareDir, remote, prID, git.EnsureOptions{
			Verbose: verbose || whatIf,
			WhatIf:  whatIf,
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		})
		if err != nil {
			return err
		}

		if verbose || whatIf {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] prref fetch: completed (%s)\n", mergeRef)
		}

		payload, err := json.Marshal(prrefFetchOutput{
			PRID:     prID,
			RepoURL:  repoURL,
			Remote:   remote,
			Provider: providerName,
			BareDir:  bareDir,
			MergeRef: mergeRef,
		})
		if err != nil {
			return apperrors.WrapRuntime("failed to encode prref fetch JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}
