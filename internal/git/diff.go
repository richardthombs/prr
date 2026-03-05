package git

import (
	"context"
	"sort"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

const contributionRange = "HEAD^1..HEAD"

func (s *Service) DiffContribution(ctx context.Context, workDir string) (types.DiffOutput, error) {
	return s.DiffContributionWithOptions(ctx, workDir, EnsureOptions{})
}

func (s *Service) DiffContributionWithOptions(ctx context.Context, workDir string, opts EnsureOptions) (types.DiffOutput, error) {
	trimmedWorkDir := strings.TrimSpace(workDir)
	if trimmedWorkDir == "" {
		return types.DiffOutput{}, apperrors.WrapConfig("worktree directory is required; provide --work-dir or stdin JSON with workDir", nil)
	}

	filesOutput, err := s.runCommand(ctx, opts, "git", "-C", trimmedWorkDir, "diff", "--name-only", contributionRange)
	if err != nil {
		return types.DiffOutput{}, apperrors.WrapRuntime("failed to compute changed file list", err)
	}

	statOutput, err := s.runCommand(ctx, opts, "git", "-C", trimmedWorkDir, "diff", "--stat", contributionRange)
	if err != nil {
		return types.DiffOutput{}, apperrors.WrapRuntime("failed to compute diff stat", err)
	}

	patchOutput, err := s.runCommand(ctx, opts, "git", "-C", trimmedWorkDir, "diff", "--patch", "--binary", contributionRange)
	if err != nil {
		return types.DiffOutput{}, apperrors.WrapRuntime("failed to compute unified patch", err)
	}

	files := parseChangedFiles(filesOutput)

	return types.DiffOutput{
		WorkDir: trimmedWorkDir,
		Range:   contributionRange,
		Files:   files,
		Stat:    statOutput,
		Patch:   patchOutput,
	}, nil
}

func parseChangedFiles(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	lines := strings.Split(trimmed, "\n")
	files := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		candidate := strings.TrimSpace(line)
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		files = append(files, candidate)
	}

	sort.Strings(files)

	return files
}
