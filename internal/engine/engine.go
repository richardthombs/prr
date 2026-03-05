package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/richardthombs/prr/internal/types"
)

type ReviewEngine interface {
	Review(ctx context.Context, bundle types.BundleV1) (types.Review, error)
}

type DefaultAdapter struct{}

func NewDefaultAdapter() ReviewEngine {
	return &DefaultAdapter{}
}

func (a *DefaultAdapter) Review(_ context.Context, bundle types.BundleV1) (types.Review, error) {
	summary := "Automated review generated"
	if bundle.PRID > 0 {
		summary = fmt.Sprintf("Automated review generated for PR #%d", bundle.PRID)
	}

	reasons := []string{"Diff analysed from deterministic bundle payload."}
	if bundle.ChangedFiles > 50 {
		reasons = append(reasons, "Large changed file count increases review risk.")
	}

	riskScore := 0.2
	if bundle.ChangedFiles > 20 {
		riskScore = 0.5
	}
	if bundle.ChangedFiles > 50 {
		riskScore = 0.8
	}

	findings := []types.Finding{}
	if strings.TrimSpace(bundle.Patch) != "" {
		findings = append(findings, types.Finding{
			ID:         "",
			File:       firstFile(bundle.Files),
			Line:       0,
			Severity:   "suggestion",
			Category:   "readability",
			Message:    "Review generated from bundle; replace default adapter with a provider-specific engine for richer findings.",
			Suggestion: "Integrate an engine adapter that analyses patch semantics and emits domain-specific findings.",
		})
	}

	return types.Review{
		Summary: summary,
		Risk: types.Risk{
			Score:   riskScore,
			Reasons: reasons,
		},
		Findings: findings,
		Checklist: []string{
			"Verify findings against repository context.",
			"Run full test suite before merge.",
		},
	}, nil
}

func firstFile(files []string) string {
	if len(files) == 0 {
		return "unknown"
	}

	return files[0]
}
