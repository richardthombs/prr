package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/richardthombs/prr/internal/bundle"
	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	bundleCmd.Flags().Bool("what-if", false, "Show actions that would be executed without side effects")
	bundleCmd.Flags().Int("max-patch-bytes", 0, "Maximum allowed patch size in bytes (0 disables limit)")
	bundleCmd.Flags().Int("max-files", 0, "Maximum allowed changed file count (0 disables limit)")
	bundleCmd.Flags().Int("pr-id", 0, "Override PR identifier")
	bundleCmd.Flags().String("repo", "", "Override repository URL")
	bundleCmd.Flags().String("remote", "", "Override git remote name")
	bundleCmd.Flags().String("provider", "", "Override PR provider")
	bundleCmd.Flags().String("merge-ref", "", "Override merge ref")
}

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Build validated v1 review bundle payload",
	Long:  "Assemble and validate the v1 bundle contract from diff JSON input and enforce size limits before engine invocation.",
	RunE: func(cmd *cobra.Command, _ []string) error {
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

		input := types.DiffOutput{}
		parsed, err := readInputJSON(cmd, &input)
		if err != nil {
			return err
		}
		if !parsed {
			return apperrors.WrapConfig("bundle command requires diff JSON on stdin", nil)
		}

		if verbose || whatIf {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] bundle: validate diff input and build v1 payload")
		}
		if whatIf {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] what-if: bundle stage uses no external commands")
		}

		prIDOverride, err := cmd.Flags().GetInt("pr-id")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse pr-id flag", err)
		}
		if prIDOverride > 0 {
			input.PRID = prIDOverride
		}

		repoOverride, err := cmd.Flags().GetString("repo")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse repo flag", err)
		}
		if strings.TrimSpace(repoOverride) != "" {
			input.RepoURL = strings.TrimSpace(repoOverride)
		}

		remoteOverride, err := cmd.Flags().GetString("remote")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse remote flag", err)
		}
		if strings.TrimSpace(remoteOverride) != "" {
			input.Remote = strings.TrimSpace(remoteOverride)
		}

		providerOverride, err := cmd.Flags().GetString("provider")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse provider flag", err)
		}
		if strings.TrimSpace(providerOverride) != "" {
			input.Provider = strings.TrimSpace(providerOverride)
		}

		mergeRefOverride, err := cmd.Flags().GetString("merge-ref")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse merge-ref flag", err)
		}
		if strings.TrimSpace(mergeRefOverride) != "" {
			input.MergeRef = strings.TrimSpace(mergeRefOverride)
		}

		payload, err := bundle.BuildV1(input, bundle.Limits{
			MaxPatchBytes:   maxPatchBytes,
			MaxChangedFiles: maxFiles,
		})
		if err != nil {
			return err
		}

		if err := bundle.ValidateV1Schema(payload); err != nil {
			return err
		}

		encoded, err := json.Marshal(payload)
		if err != nil {
			return apperrors.WrapRuntime("failed to encode bundle JSON", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
		if err != nil {
			return apperrors.WrapRuntime("failed to write output", err)
		}

		return nil
	},
}
