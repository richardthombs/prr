package main

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/richardthombs/prr/internal/types"
)

func renderMarkdown(review types.Review, prID int, prURL string, issues []types.RelatedIssue) string {
	var builder strings.Builder

	builder.WriteString("## Summary\n")
	builder.WriteString(renderPRLine(prID, prURL))
	builder.WriteString("\n")
	builder.WriteString(renderIssuesLine(issues))
	builder.WriteString("\n\n")

	builder.WriteString("### Issue summary\n")
	builder.WriteString(review.IssueSummary)
	builder.WriteString("\n\n")

	builder.WriteString("### PR summary\n")
	builder.WriteString(review.PRSummary)
	builder.WriteString("\n\n")

	builder.WriteString("## Review\n")
	builder.WriteString("### A) Issue resolution assessment\n")
	builder.WriteString(fmt.Sprintf("Risk score: %.2f\n", review.Risk.Score))
	for _, reason := range review.Risk.Reasons {
		builder.WriteString("- ")
		builder.WriteString(reason)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	builder.WriteString("### B) PR change review conclusions\n")
	severityOrder := []string{"blocker", "important", "suggestion", "nit"}
	grouped := map[string][]types.Finding{}
	for _, finding := range review.Findings {
		grouped[finding.Severity] = append(grouped[finding.Severity], finding)
	}

	for _, severity := range severityOrder {
		findings := grouped[severity]
		builder.WriteString("### ")
		builder.WriteString(severityHeading(severity))
		builder.WriteString("\n")
		if len(findings) == 0 {
			builder.WriteString("- None.\n\n")
			continue
		}

		sort.SliceStable(findings, func(i, j int) bool {
			if findings[i].File == findings[j].File {
				if findings[i].Line == findings[j].Line {
					return findings[i].ID < findings[j].ID
				}
				return findings[i].Line < findings[j].Line
			}

			return findings[i].File < findings[j].File
		})

		for _, finding := range findings {
			builder.WriteString(fmt.Sprintf("- [%s] %s:%d (%s) - %s\n", finding.ID, finding.File, finding.Line, finding.Category, finding.Message))
			builder.WriteString(fmt.Sprintf("  Suggestion: %s\n", finding.Suggestion))
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

func severityHeading(severity string) string {
	switch severity {
	case "blocker":
		return "Blocker"
	case "important":
		return "Important"
	case "suggestion":
		return "Suggestion"
	case "nit":
		return "Nitpick"
	default:
		return severity
	}
}

func renderPRLine(prID int, prURL string) string {
	if prID <= 0 {
		return "PR: N/A"
	}
	if prURL == "" {
		return fmt.Sprintf("PR: #%d", prID)
	}
	return fmt.Sprintf("PR: [#%d](%s)", prID, prURL)
}

func renderIssuesLine(issues []types.RelatedIssue) string {
	if len(issues) == 0 {
		return "Issues: None"
	}

	deduped := make(map[string]types.RelatedIssue, len(issues))
	for _, issue := range issues {
		trimmedID := strings.TrimSpace(issue.ID)
		if trimmedID == "" {
			continue
		}
		if _, exists := deduped[trimmedID]; exists {
			continue
		}
		deduped[trimmedID] = issue
	}
	if len(deduped) == 0 {
		return "Issues: None"
	}

	ids := make([]string, 0, len(deduped))
	for id := range deduped {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	refs := make([]string, 0, len(ids))
	for _, id := range ids {
		issue := deduped[id]
		issueURL := strings.TrimSpace(issue.URL)
		if issueURL != "" {
			refs = append(refs, fmt.Sprintf("[#%s](%s)", id, issueURL))
		} else {
			refs = append(refs, fmt.Sprintf("#%s", id))
		}
	}
	return "Issues: " + strings.Join(refs, ", ")
}

func buildPRURL(repoURL string, prID int) string {
	if strings.TrimSpace(repoURL) == "" || prID <= 0 {
		return ""
	}
	parsed, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil {
		return ""
	}
	parsed.Path = strings.TrimSuffix(parsed.Path, "/") + "/pull/" + strconv.Itoa(prID)
	return parsed.String()
}
