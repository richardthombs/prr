package config

import (
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

const (
	defaultProvider = "azure-devops"
	defaultRemote   = "origin"
)

type ResolveInput struct {
	Provider string
	RepoURL  string
	Remote   string
}

type ResolveConfig struct {
	Provider string
	RepoURL  string
	Remote   string
}

func Resolve(input ResolveInput) (ResolveConfig, error) {
	provider := strings.TrimSpace(input.Provider)
	repoURL := strings.TrimSpace(input.RepoURL)
	remote := strings.TrimSpace(input.Remote)

	if provider == "" {
		provider = defaultProvider
	}
	if remote == "" {
		remote = defaultRemote
	}

	if repoURL == "" {
		return ResolveConfig{}, apperrors.WrapConfig("repository context is required; provide --repo", nil)
	}

	return ResolveConfig{
		Provider: provider,
		RepoURL:  repoURL,
		Remote:   remote,
	}, nil
}