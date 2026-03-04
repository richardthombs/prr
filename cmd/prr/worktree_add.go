package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/spf13/cobra"
)

type worktreeAddOutput struct {
	PRID     int    `json:"prId"`
	RepoURL  string `json:"repoUrl,omitempty"`
	Remote   string `json:"remote,omitempty"`
	Provider string `json:"provider,omitempty"`
	BareDir  string `json:"bareDir"`
	MergeRef string `json:"mergeRef"`
	WorkDir  string `json:"workDir"`
	Keep     bool   `json:"keep"`
}

func init() {
	rootCmd.AddCommand(worktreeCmd)
	worktreeCmd.AddCommand(worktreeAddCmd)

	worktreeAddCmd.Flags().Int("pr-id", 0, "Pull request ID")
	worktreeAddCmd.Flags().String("repo", "", "Repository URL")
	worktreeAddCmd.Flags().String("remote", "origin", "Git remote name")
	worktreeAddCmd.Flags().String("provider", "", "PR provider")
	worktreeAddCmd.Flags().String("bare-dir", "", "Bare mirror directory")
	worktreeAddCmd.Flags().String("merge-ref", "", "PR merge ref in PRR namespace")
	worktreeAddCmd.Flags().Bool("keep", false, "Retain worktree after review chain completion")
	worktreeAddCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	worktreeAddCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
}

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Worktree lifecycle operations",
	Long:  "Worktree lifecycle operations for isolated PR review workspace management.",
}

var worktreeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create detached isolated worktree for a PR merge ref",
	Long:  "Create a detached worktree from refs/prr/pull/<PR_ID>/merge and emit JSON including workDir.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdinInput, hasStdinInput, err := readOptionalComposeInput(cmd)
		if err != nil {
			return err
		}

		prID, err := cmd.Flags().GetInt("pr-id")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse pr-id flag", err)
		}
		if prID == 0 && hasStdinInput {
			prID = stdinInput.PRID
		}

		repoURL, err := cmd.Flags().GetString("repo")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse repo flag", err)
		}
		if repoURL == "" && hasStdinInput {
			repoURL = stdinInput.RepoURL
		}

		remote, err := cmd.Flags().GetString("remote")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse remote flag", err)
		}
		if !cmd.Flags().Changed("remote") && hasStdinInput && stdinInput.Remote != "" {
			remote = stdinInput.Remote
		}

		providerName, err := cmd.Flags().GetString("provider")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse provider flag", err)
		}
		if providerName == "" && hasStdinInput {
			providerName = stdinInput.Provider
		}

		bareDir, err := cmd.Flags().GetString("bare-dir")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse bare-dir flag", err)
		}
		if bareDir == "" && hasStdinInput {
			bareDir = stdinInput.BareDir
		}

		mergeRef, err := cmd.Flags().GetString("merge-ref")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse merge-ref flag", err)
		}
		if mergeRef == "" && hasStdinInput {
			mergeRef = stdinInput.MergeRef
		}

		if prID == 0 && mergeRef != "" {
			parsedPRID, parseErr := parsePRIDFromMergeRef(mergeRef)
			if parseErr != nil {
				return parseErr
			}
			prID = parsedPRID
		}

		if mergeRef == "" && prID > 0 {
			mergeRef = git.MergeRefForPRID(prID)
		}

		if prID <= 0 {
			return apperrors.WrapConfig("valid PR ID is required; provide --pr-id", nil)
		}

		if strings.TrimSpace(mergeRef) == "" {
			return apperrors.WrapConfig("merge ref is required; provide --merge-ref or --pr-id", nil)
		}

		keep, err := cmd.Flags().GetBool("keep")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse keep flag", err)
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

		workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, prID)
		if err != nil {
			return err
		}

		if verbose {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] worktree add: creating detached workspace for PR %d\n", prID)
			if whatIf {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] worktree add: what-if enabled, no external commands will be executed")
			}
		}

		err = service.CreateWorktree(context.Background(), bareDir, mergeRef, workDir, git.EnsureOptions{
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
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] worktree add: completed (%s)\n", workDir)
		}

		payload, err := json.Marshal(worktreeAddOutput{
			PRID:     prID,
			RepoURL:  repoURL,
			Remote:   remote,
			Provider: providerName,
			BareDir:  bareDir,
			MergeRef: mergeRef,
			WorkDir:  workDir,
			Keep:     keep,
		})
		if err != nil {
			return apperrors.WrapRuntime("failed to encode worktree add JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}

func parsePRIDFromMergeRef(mergeRef string) (int, error) {
	trimmed := strings.TrimSpace(mergeRef)
	prefix := "refs/prr/pull/"
	suffix := "/merge"

	if !strings.HasPrefix(trimmed, prefix) || !strings.HasSuffix(trimmed, suffix) {
		return 0, apperrors.WrapConfig("merge ref must match refs/prr/pull/<PR_ID>/merge", nil)
	}

	inner := strings.TrimPrefix(trimmed, prefix)
	idValue := strings.TrimSuffix(inner, suffix)
	if strings.TrimSpace(idValue) == "" {
		return 0, apperrors.WrapConfig("merge ref must contain a PR ID", nil)
	}

	prID, err := strconv.Atoi(idValue)
	if err != nil || prID <= 0 {
		return 0, apperrors.WrapConfig("merge ref must contain a valid numeric PR ID", err)
	}

	return prID, nil
}
