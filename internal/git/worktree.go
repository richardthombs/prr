package git

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

func (s *Service) CreateWorktree(ctx context.Context, bareDir, mergeRef, workDir string, opts EnsureOptions) error {
	trimmedBareDir := strings.TrimSpace(bareDir)
	if trimmedBareDir == "" {
		return apperrors.WrapConfig("bare mirror directory is required; provide --bare-dir", nil)
	}

	trimmedMergeRef := strings.TrimSpace(mergeRef)
	if trimmedMergeRef == "" {
		return apperrors.WrapConfig("merge ref is required; provide --merge-ref or --pr-id", nil)
	}

	trimmedWorkDir := strings.TrimSpace(workDir)
	if trimmedWorkDir == "" {
		return apperrors.WrapConfig("worktree directory is required", nil)
	}

	if !opts.WhatIf {
		if err := os.MkdirAll(filepath.Dir(trimmedWorkDir), 0o755); err != nil {
			return apperrors.WrapRuntime("failed to create worktree parent directory", err)
		}
	}

	_, err := s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "worktree", "add", "--detach", trimmedWorkDir, trimmedMergeRef)
	if err != nil {
		return apperrors.WrapRuntime("failed to create detached worktree", err)
	}

	return nil
}

func (s *Service) CleanupWorktree(ctx context.Context, bareDir, workDir string, opts EnsureOptions) error {
	trimmedBareDir := strings.TrimSpace(bareDir)
	if trimmedBareDir == "" {
		return apperrors.WrapConfig("bare mirror directory is required; provide --bare-dir", nil)
	}

	trimmedWorkDir := strings.TrimSpace(workDir)
	if trimmedWorkDir == "" {
		return apperrors.WrapConfig("worktree directory is required", nil)
	}

	_, err := s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "worktree", "remove", "--force", trimmedWorkDir)
	if err != nil {
		return apperrors.WrapRuntime("failed to remove worktree", err)
	}

	_, err = s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "worktree", "prune")
	if err != nil {
		return apperrors.WrapRuntime("failed to prune worktrees", err)
	}

	return nil
}

func (s *Service) ResolveWorktreeDirFromBareDir(bareDir string, prID int) (string, error) {
	trimmedBareDir := strings.TrimSpace(bareDir)
	if trimmedBareDir == "" {
		return "", apperrors.WrapConfig("bare mirror directory is required; provide --bare-dir", nil)
	}

	if prID <= 0 {
		return "", apperrors.WrapConfig("valid PR ID is required; provide --pr-id", nil)
	}

	repoHash := strings.TrimSuffix(filepath.Base(trimmedBareDir), ".git")
	if strings.TrimSpace(repoHash) == "" || repoHash == "." || repoHash == string(filepath.Separator) {
		return "", apperrors.WrapConfig("could not determine repository hash from bare mirror path", nil)
	}

	runID := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	return filepath.Join(defaultWorktreeCacheDir(), repoHash, "pr-"+strconv.Itoa(prID), runID), nil
}

func defaultWorktreeCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".", ".prr", "work")
	}

	return filepath.Join(userCacheDir, "prr", "work")
}
