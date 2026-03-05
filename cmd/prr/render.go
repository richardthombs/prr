package main

import (
	"fmt"
	"sort"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.Flags().Bool("verbose", false, "Emit progress logs to stderr")
	renderCmd.Flags().Bool("what-if", false, "Show actions that would be executed without side effects")
}

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render review JSON into Markdown",
	Long:  "Read structured review JSON from stdin and output deterministic Markdown sections for summary, risk, findings, and checklist.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return apperrors.WrapRuntime("failed to parse verbose flag", err)
		}
		whatIf, err := readWhatIfFlag(cmd)
		if err != nil {
			return err
		}

		input := types.Review{}
		parsed, err := readInputJSON(cmd, &input)
		if err != nil {
			return err
		}
		if !parsed {
			return apperrors.WrapConfig("render command requires review JSON on stdin", nil)
		}

		review, err := types.ValidateReviewInput(input)
		if err != nil {
			return err
		}

		if verbose || whatIf {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] render: transform review JSON to markdown")
		}
		if whatIf {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[prr] what-if: render stage uses no external commands")
		}

		markdown := renderMarkdown(review)
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), markdown); err != nil {
			return apperrors.WrapRuntime("failed to write markdown output", err)
		}

		return nil
	},
}

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
