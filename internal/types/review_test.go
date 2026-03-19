package types

import (
	"strings"
	"testing"
)

func TestNormalizeAndValidateReviewOutputRejectsRiskScoreOutsideZeroToOne(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk: Risk{Score: 4, Reasons: []string{"r"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       7,
			Severity:   "important",
			Category:   "tests",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{"c"},
	})
	if err == nil {
		t.Fatalf("expected risk score validation failure")
	}
}

func TestNormalizeAndValidateReviewOutputStillValidatesStructure(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk: Risk{Score: 4, Reasons: []string{"r"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       7,
			Severity:   "critical",
			Category:   "tests",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{"c"},
	})
	if err == nil {
		t.Fatalf("expected structure validation failure for invalid severity")
	}
}

func TestValidateReviewInputKeepsStrictRiskScale(t *testing.T) {
	_, err := ValidateReviewInput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk: Risk{Score: 4, Reasons: []string{"r"}},
		Findings: []Finding{{
			ID:         "F001",
			File:       "a.go",
			Line:       7,
			Severity:   "important",
			Category:   "tests",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{"c"},
	})
	if err == nil {
		t.Fatalf("expected strict input validation failure")
	}
}

func TestValidateReviewRejectsMissingIssueSummary(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		PRSummary: "ok",
		Risk:      Risk{Score: 0.5, Reasons: []string{"r"}},
		Findings:  []Finding{},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing issueSummary")
	}
	if !strings.Contains(err.Error(), "missing issueSummary") {
		t.Fatalf("expected missing issueSummary diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsMissingPRSummary(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		Risk:         Risk{Score: 0.5, Reasons: []string{"r"}},
		Findings:     []Finding{},
		Checklist:    []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing prSummary")
	}
	if !strings.Contains(err.Error(), "missing prSummary") {
		t.Fatalf("expected missing prSummary diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsNilRiskReasons(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:      Risk{Score: 0.5, Reasons: nil},
		Findings:  []Finding{},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for nil risk.reasons")
	}
	if !strings.Contains(err.Error(), "missing risk.reasons") {
		t.Fatalf("expected missing risk.reasons diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsNilFindings(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:      Risk{Score: 0.5, Reasons: []string{"r"}},
		Findings:  nil,
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for nil findings")
	}
	if !strings.Contains(err.Error(), "missing findings") {
		t.Fatalf("expected missing findings diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsMissingFindingFile(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.1, Reasons: []string{"low"}},
		Findings: []Finding{{
			Severity:   "nit",
			Category:   "other",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing finding file")
	}
	if !strings.Contains(err.Error(), "finding is missing file") {
		t.Fatalf("expected missing file diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsInvalidFindingCategory(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.1, Reasons: []string{"low"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       1,
			Severity:   "nit",
			Category:   "invalid-category",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for invalid finding category")
	}
	if !strings.Contains(err.Error(), "finding has invalid category") {
		t.Fatalf("expected invalid category diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsMissingFindingMessage(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.1, Reasons: []string{"low"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       1,
			Severity:   "nit",
			Category:   "other",
			Message:    "",
			Suggestion: "s",
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing finding message")
	}
	if !strings.Contains(err.Error(), "finding is missing message") {
		t.Fatalf("expected missing message diagnostic, got %v", err)
	}
}

func TestValidateReviewRejectsMissingFindingSuggestion(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.1, Reasons: []string{"low"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       1,
			Severity:   "nit",
			Category:   "other",
			Message:    "m",
			Suggestion: "",
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing finding suggestion")
	}
	if !strings.Contains(err.Error(), "finding is missing suggestion") {
		t.Fatalf("expected missing suggestion diagnostic, got %v", err)
	}
}

func TestNormalizeAndValidateReviewOutputAutoAssignsIDAndLine(t *testing.T) {
	result, err := NormalizeAndValidateReviewOutput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.1, Reasons: []string{"low"}},
		Findings: []Finding{{
			File:       "a.go",
			Severity:   "nit",
			Category:   "other",
			Message:    "m",
			Suggestion: "s",
			// ID and Line deliberately omitted — output mode auto-assigns
		}},
		Checklist: []string{},
	})
	if err != nil {
		t.Fatalf("expected auto-assignment to succeed, got %v", err)
	}
	if result.Findings[0].ID != "F001" {
		t.Fatalf("expected auto-assigned ID F001, got %q", result.Findings[0].ID)
	}
	if result.Findings[0].Line != 1 {
		t.Fatalf("expected auto-assigned line 1, got %d", result.Findings[0].Line)
	}
}

func TestValidateReviewInputRejectsMissingFindingID(t *testing.T) {
	_, err := ValidateReviewInput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.2, Reasons: []string{"low"}},
		Findings: []Finding{{
			File:       "a.go",
			Line:       5,
			Severity:   "suggestion",
			Category:   "readability",
			Message:    "m",
			Suggestion: "s",
			// ID deliberately omitted
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for missing finding ID in strict input mode")
	}
	if !strings.Contains(err.Error(), "finding is missing id") {
		t.Fatalf("expected missing id diagnostic, got %v", err)
	}
}

func TestValidateReviewInputRejectsZeroLine(t *testing.T) {
	_, err := ValidateReviewInput(Review{
		IssueSummary: "ok",
		PRSummary:    "ok",
		Risk:    Risk{Score: 0.2, Reasons: []string{"low"}},
		Findings: []Finding{{
			ID:         "F001",
			File:       "a.go",
			Line:       0,
			Severity:   "suggestion",
			Category:   "readability",
			Message:    "m",
			Suggestion: "s",
		}},
		Checklist: []string{},
	})
	if err == nil {
		t.Fatalf("expected error for zero line in strict input mode")
	}
	if !strings.Contains(err.Error(), "finding must include a positive line") {
		t.Fatalf("expected positive line diagnostic, got %v", err)
	}
}
