package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/richardthombs/prr/internal/types"
)

func renderMarkdown(review types.Review) string {
	var builder strings.Builder

	builder.WriteString("## Summary\n")
	builder.WriteString(review.Summary)
	builder.WriteString("\n\n")

	builder.WriteString("## Risk\n")
	builder.WriteString(fmt.Sprintf("Score: %.2f\n", review.Risk.Score))
	if len(review.Risk.Reasons) > 0 {
		for _, reason := range review.Risk.Reasons {
			builder.WriteString("- ")
			builder.WriteString(reason)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n")

	builder.WriteString("## Findings\n")
	severityOrder := []string{"blocker", "important", "suggestion", "nit"}
	grouped := map[string][]types.Finding{}
	for _, finding := range review.Findings {
		grouped[finding.Severity] = append(grouped[finding.Severity], finding)
	}

	for _, severity := range severityOrder {
		findings := grouped[severity]
		if len(findings) == 0 {
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

		builder.WriteString("### ")
		builder.WriteString(severityHeading(severity))
		builder.WriteString("\n")
		for _, finding := range findings {
			builder.WriteString(fmt.Sprintf("- [%s] %s:%d (%s) - %s\n", finding.ID, finding.File, finding.Line, finding.Category, finding.Message))
			builder.WriteString(fmt.Sprintf("  Suggestion: %s\n", finding.Suggestion))
		}
		builder.WriteString("\n")
	}

	if len(review.Findings) == 0 {
		builder.WriteString("No findings.\n\n")
	}

	builder.WriteString("## Checklist\n")
	for _, item := range review.Checklist {
		builder.WriteString("- [ ] ")
		builder.WriteString(item)
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
		return "Nit"
	default:
		return severity
	}
}
