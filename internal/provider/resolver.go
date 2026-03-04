package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/richardthombs/prr/internal/config"
	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

type Resolver struct {
	provider PRProvider
}

func NewResolver(provider PRProvider) *Resolver {
	return &Resolver{provider: provider}
}

func (r *Resolver) Resolve(ctx context.Context, prID int, opts ResolveOptions) (types.PRRef, error) {
	resolved, err := config.Resolve(config.ResolveInput{
		Provider: opts.Provider,
		RepoURL:  opts.RepoURL,
		Remote:   opts.Remote,
	})
	if err != nil {
		return types.PRRef{}, err
	}

	prRef, err := r.provider.Resolve(ctx, prID, map[string]string{
		"provider": resolved.Provider,
		"repoUrl":  resolved.RepoURL,
		"remote":   resolved.Remote,
	})
	if err != nil {
		return types.PRRef{}, apperrors.WrapProvider("failed to resolve pull request reference", err)
	}

	return prRef, nil
}

func (r *Resolver) ResolveFromPullRequestURL(ctx context.Context, prURL string, opts ResolveOptions) (types.PRRef, error) {
	parsedContext, err := parsePullRequestURL(strings.TrimSpace(prURL))
	if err != nil {
		return types.PRRef{}, apperrors.WrapConfig(fmt.Sprintf("invalid pull request URL: %v", err), nil)
	}

	repoURL := parsedContext.RepoURL
	if strings.TrimSpace(opts.RepoURL) != "" {
		repoURL = opts.RepoURL
	}

	providerName := parsedContext.Provider
	if strings.TrimSpace(opts.Provider) != "" {
		providerName = opts.Provider
	}

	remoteName := parsedContext.Remote
	if strings.TrimSpace(opts.Remote) != "" {
		remoteName = opts.Remote
	}

	return r.Resolve(ctx, parsedContext.PRID, ResolveOptions{
		Provider: providerName,
		RepoURL:  repoURL,
		Remote:   remoteName,
	})
}