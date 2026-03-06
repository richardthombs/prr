package types

import "testing"

func TestNormalizeAndValidateReviewOutputRejectsRiskScoreOutsideZeroToOne(t *testing.T) {
	_, err := NormalizeAndValidateReviewOutput(Review{
		Summary: "ok",
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
		Summary: "ok",
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
		Summary: "ok",
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
