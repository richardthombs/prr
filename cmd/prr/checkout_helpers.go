package main

import (
	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/git"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/spf13/cobra"
)

var mirrorServiceFactory = func() *git.Service {
	return git.NewService(git.NewExecRunner())
}

var prEnricherFactory = func() provider.CLIRunner {
	return git.NewExecRunner()
}

type resolveOptions struct {
	Provider string
	RepoURL  string
	Remote   string
}

type checkoutOutput struct {
	PRID     int    `json:"prId"`
	RepoURL  string `json:"repoUrl,omitempty"`
	Remote   string `json:"remote,omitempty"`
	Provider string `json:"provider,omitempty"`
	BareDir  string `json:"bareDir"`
	MergeRef string `json:"mergeRef"`
	BaseRef  string `json:"baseRef,omitempty"`
	WorkDir  string `json:"workDir"`
	Keep     bool   `json:"keep"`
	Cleanup  bool   `json:"cleanup"`
}

func resolveFlags(cmd *cobra.Command) (resolveOptions, error) {
	providerValue, err := cmd.Flags().GetString("provider")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse provider flag", err)
	}
	repoValue, err := cmd.Flags().GetString("repo")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse repo flag", err)
	}
	remoteValue, err := cmd.Flags().GetString("remote")
	if err != nil {
		return resolveOptions{}, apperrors.WrapRuntime("failed to parse remote flag", err)
	}

	return resolveOptions{
		Provider: providerValue,
		RepoURL:  repoValue,
		Remote:   remoteValue,
	}, nil
}

func readWhatIfFlag(cmd *cobra.Command) (bool, error) {
	whatIf, err := cmd.Flags().GetBool("what-if")
	if err != nil {
		return false, apperrors.WrapRuntime("failed to parse what-if flag", err)
	}

	return whatIf, nil
}
