package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/richardthombs/prr/internal/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().String("work-dir", "", "Path to isolated review worktree")
	diffCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	diffCmd.Flags().Bool("what-if", false, "Show commands that would be executed without running them")
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Generate deterministic diff outputs from an isolated worktree",
	Long:  "Compute changed files, diff stat, and unified patch for HEAD^1..HEAD and emit a JSON payload for bundle composition.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		workDir, err := cmd.Flags().GetString("work-dir")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse work-dir flag", err)
		}
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse verbose flag", err)
		}
		whatIf, err := readWhatIfFlag(cmd)
		if err != nil {
			return err
		}

		input := types.DiffOutput{}
		_, err = readInputJSON(cmd, &input)
		if err != nil {
			return err
		}

		resolvedWorkDir := strings.TrimSpace(workDir)
		if resolvedWorkDir == "" {
			resolvedWorkDir = strings.TrimSpace(input.WorkDir)
		}
		if resolvedWorkDir == "" {
			return apperrors.WrapConfig("worktree directory is required; provide --work-dir or stdin JSON with workDir", nil)
		}

		service := mirrorServiceFactory()
		diffOutput, err := service.DiffContributionWithOptions(context.Background(), resolvedWorkDir, git.EnsureOptions{
			Verbose: verbose || whatIf,
			WhatIf:  whatIf,
			Logger: func(format string, args ...any) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[prr] "+format+"\n", args...)
			},
		})
		if err != nil {
			return err
		}

		diffOutput.PRID = input.PRID
		diffOutput.RepoURL = input.RepoURL
		diffOutput.Remote = input.Remote
		diffOutput.Provider = input.Provider
		diffOutput.BareDir = input.BareDir
		diffOutput.MergeRef = input.MergeRef
		diffOutput.WorkDir = resolvedWorkDir

		payload, err := json.Marshal(diffOutput)
		if err != nil {
			return apperrors.WrapRuntime("failed to encode diff JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}
