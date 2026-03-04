package provider

import (
	"context"
	"fmt"

	"github.com/richardthombs/prr/internal/types"
)

type defaultProvider struct{}

func NewDefaultProvider() PRProvider {
	return &defaultProvider{}
}

func (p *defaultProvider) Resolve(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
	repoURL := opts["repoUrl"]
	providerName := opts["provider"]
	remote := opts["remote"]

	if repoURL == "" {
		return types.PRRef{}, fmt.Errorf("missing repository URL")
	}

	return types.PRRef{
		PRID:     prID,
		RepoURL:  repoURL,
		Remote:   remote,
		Provider: providerName,
	}, nil
}