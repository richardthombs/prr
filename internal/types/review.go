package types

import (
	"fmt"
	"slices"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

var validSeverities = []string{"blocker", "important", "suggestion", "nit"}
var validCategories = []string{"correctness", "security", "performance", "readability", "api", "tests", "other"}

type Risk struct {
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

type Finding struct {
	ID         string `json:"id"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

type Review struct {
	Summary   string    `json:"summary"`
	Risk      Risk      `json:"risk"`
	Findings  []Finding `json:"findings"`
	Checklist []string  `json:"checklist"`
}

func NormalizeAndValidateReview(input Review) (Review, error) {
	review := input

	review.Summary = strings.TrimSpace(review.Summary)
	if review.Summary == "" {
		return Review{}, apperrors.WrapEngine("review output is missing summary", nil)
	}

	if review.Risk.Score < 0 || review.Risk.Score > 1 {
		return Review{}, apperrors.WrapEngine("review output risk.score must be between 0 and 1", nil)
	}
	if review.Risk.Reasons == nil {
		return Review{}, apperrors.WrapEngine("review output is missing risk.reasons", nil)
	}

	for i := range review.Risk.Reasons {
		review.Risk.Reasons[i] = strings.TrimSpace(review.Risk.Reasons[i])
	}

	if review.Findings == nil {
		return Review{}, apperrors.WrapEngine("review output is missing findings", nil)
	}

	for i := range review.Findings {
		finding := &review.Findings[i]
		if strings.TrimSpace(finding.ID) == "" {
			finding.ID = fmt.Sprintf("F%03d", i+1)
		}

		finding.File = strings.TrimSpace(finding.File)
		if finding.File == "" {
			return Review{}, apperrors.WrapEngine("review output finding is missing file", nil)
		}

		severity := strings.ToLower(strings.TrimSpace(finding.Severity))
		if !slices.Contains(validSeverities, severity) {
			return Review{}, apperrors.WrapEngine("review output finding has invalid severity", nil)
		}
		finding.Severity = severity

		category := strings.ToLower(strings.TrimSpace(finding.Category))
		if !slices.Contains(validCategories, category) {
			return Review{}, apperrors.WrapEngine("review output finding has invalid category", nil)
		}
		finding.Category = category

		finding.Message = strings.TrimSpace(finding.Message)
		if finding.Message == "" {
			return Review{}, apperrors.WrapEngine("review output finding is missing message", nil)
		}

		finding.Suggestion = strings.TrimSpace(finding.Suggestion)
		if finding.Suggestion == "" {
			return Review{}, apperrors.WrapEngine("review output finding is missing suggestion", nil)
		}
	}

	if review.Checklist == nil {
		return Review{}, apperrors.WrapEngine("review output is missing checklist", nil)
	}
	for i := range review.Checklist {
		review.Checklist[i] = strings.TrimSpace(review.Checklist[i])
	}

	return review, nil
}
