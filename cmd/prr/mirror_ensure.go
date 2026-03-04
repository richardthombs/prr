package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/spf13/cobra"
)

var mirrorServiceFactory = func() *git.Service {
	return git.NewService(git.NewExecRunner())
}

type mirrorEnsureOutput struct {
	RepoURL string `json:"repoUrl"`
	BareDir string `json:"bareDir"`
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
	mirrorCmd.AddCommand(mirrorEnsureCmd)

	mirrorEnsureCmd.Flags().String("repo", "", "Repository URL to mirror")
	mirrorEnsureCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	mirrorEnsureCmd.Flags().Duration("lock-timeout", 30*time.Second, "Maximum time to wait for repository mirror lock")
	mirrorEnsureCmd.Flags().Bool("force", false, "Bypass mirror lock acquisition (unsafe; use only if lock is stuck)")
	mirrorEnsureCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
}

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror cache operations",
	Long:  "Mirror cache operations for deterministic repository state management.",
}

var mirrorEnsureCmd = &cobra.Command{
	Use:   "ensure",
	Short: "Create or update bare mirror for repository",
	Long:  "Create or update the deterministic bare mirror location for a repository and emit JSON including bareDir.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		repoURL, err := cmd.Flags().GetString("repo")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse repo flag", err)
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse verbose flag", err)
		}

		lockTimeout, err := cmd.Flags().GetDuration("lock-timeout")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse lock-timeout flag", err)
		}

		forceLock, err := cmd.Flags().GetBool("force")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse force flag", err)
		}

		whatIf, err := cmd.Flags().GetBool("what-if")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse what-if flag", err)
		}

		service := mirrorServiceFactory()

		bareDirPreview, err := service.ResolveMirrorDir(repoURL)
		if err != nil {
			return err
		}

		if verbose {
			action := "create"
			if _, statErr := os.Stat(bareDirPreview); statErr == nil {
				action = "update"
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] mirror ensure: %s mirror at %s\n", action, bareDirPreview)
			if forceLock {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] mirror ensure: --force enabled, lock bypass requested")
			} else {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] mirror ensure: lock timeout set to %s\n", lockTimeout)
			}
			if whatIf {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] mirror ensure: what-if enabled, no external commands will be executed")
			}
		}

		bareDir, err := service.EnsureMirrorWithOptions(context.Background(), repoURL, git.EnsureOptions{
			LockTimeout: lockTimeout,
			ForceLock:   forceLock,
			Verbose:     verbose || whatIf,
			WhatIf:      whatIf,
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		})
		if err != nil {
			return err
		}

		if verbose || whatIf {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] mirror ensure: completed (%s)\n", bareDir)
		}

		payload, err := json.Marshal(mirrorEnsureOutput{RepoURL: repoURL, BareDir: bareDir})
		if err != nil {
			return apperrors.WrapRuntime("failed to encode mirror ensure JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}
