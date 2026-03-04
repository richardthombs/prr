package provider

import (
	"context"

	"github.com/richardthombs/prr/internal/types"
)

type ResolveOptions struct {
	Provider string
	RepoURL  string
	Remote   string
}

type PRProvider interface {
	Resolve(ctx context.Context, prID int, opts map[string]string) (types.PRRef, error)
}