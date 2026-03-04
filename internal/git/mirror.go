package git

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

const (
	mergeRefPrefix = "refs/prr/pull/"
)

const defaultLockTimeout = 30 * time.Second

type EnsureOptions struct {
	LockTimeout time.Duration
	ForceLock   bool
	Verbose     bool
	WhatIf      bool
	Logger      func(format string, args ...any)
}

type Service struct {
	runner   Runner
	cacheDir string
}

func NewService(runner Runner) *Service {
	base := defaultMirrorCacheDir()

	return &Service{runner: runner, cacheDir: base}
}

func NewServiceWithCacheDir(runner Runner, cacheDir string) *Service {
	return &Service{runner: runner, cacheDir: cacheDir}
}

func (s *Service) EnsureMirror(ctx context.Context, repoURL string) (string, error) {
	return s.EnsureMirrorWithOptions(ctx, repoURL, EnsureOptions{LockTimeout: defaultLockTimeout})
}

func (s *Service) EnsureMirrorWithOptions(ctx context.Context, repoURL string, opts EnsureOptions) (string, error) {
	bareDir, err := s.ResolveMirrorDir(repoURL)
	if err != nil {
		return "", err
	}

	if opts.WhatIf {
		if _, statErr := os.Stat(bareDir); statErr == nil {
			_, runErr := s.runCommand(ctx, opts, "git", "-C", bareDir, "remote", "update", "--prune")
			if runErr != nil {
				return "", runErr
			}

			return bareDir, nil
		} else if !os.IsNotExist(statErr) {
			return "", apperrors.WrapRuntime("failed to inspect bare mirror path", statErr)
		}

		_, runErr := s.runCommand(ctx, opts, "git", "clone", "--mirror", strings.TrimSpace(repoURL), bareDir)
		if runErr != nil {
			return "", runErr
		}

		return bareDir, nil
	}

	if err := os.MkdirAll(filepath.Dir(bareDir), 0o755); err != nil {
		return "", apperrors.WrapRuntime("failed to create mirror cache root", err)
	}

	err = s.withRepoLock(bareDir, opts, func() error {
		if _, statErr := os.Stat(bareDir); statErr == nil {
			_, runErr := s.runCommand(ctx, opts, "git", "-C", bareDir, "remote", "update", "--prune")
			if runErr != nil {
				return apperrors.WrapRuntime("failed to update bare mirror", runErr)
			}

			return nil
		} else if !os.IsNotExist(statErr) {
			return apperrors.WrapRuntime("failed to inspect bare mirror path", statErr)
		}

		_, runErr := s.runCommand(ctx, opts, "git", "clone", "--mirror", strings.TrimSpace(repoURL), bareDir)
		if runErr != nil {
			return apperrors.WrapRuntime("failed to create bare mirror", runErr)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return bareDir, nil
}

func (s *Service) runCommand(ctx context.Context, opts EnsureOptions, name string, args ...string) (string, error) {
	if opts.Verbose && opts.Logger != nil {
		opts.Logger("exec: %s %s", name, strings.Join(args, " "))
	}

	if opts.WhatIf {
		return "", nil
	}

	return s.runner.Run(ctx, name, args...)
}

func (s *Service) FetchPRMergeRef(ctx context.Context, bareDir, remote string, prID int) (string, error) {
	trimmedBareDir := strings.TrimSpace(bareDir)
	if trimmedBareDir == "" {
		return "", apperrors.WrapConfig("bare mirror directory is required; provide --bare-dir", nil)
	}

	trimmedRemote := strings.TrimSpace(remote)
	if trimmedRemote == "" {
		trimmedRemote = "origin"
	}

	if prID <= 0 {
		return "", apperrors.WrapConfig("valid PR ID is required; provide --pr-id", nil)
	}

	mergeRef := MergeRefForPRID(prID)
	sourceRef := "pull/" + strconv.Itoa(prID) + "/merge"
	destination := sourceRef + ":" + mergeRef

	_, err := s.runner.Run(ctx, "git", "-C", trimmedBareDir, "fetch", trimmedRemote, destination)
	if err != nil {
		return "", apperrors.WrapProvider("failed to fetch PR merge ref", err)
	}

	return mergeRef, nil
}

func (s *Service) ResolveMirrorDir(repoURL string) (string, error) {
	trimmedRepo := strings.TrimSpace(repoURL)
	if trimmedRepo == "" {
		return "", apperrors.WrapConfig("repository context is required; provide --repo", nil)
	}

	hash := sha256.Sum256([]byte(trimmedRepo))
	repoHash := hex.EncodeToString(hash[:])

	return filepath.Join(s.cacheDir, repoHash+".git"), nil
}

func MergeRefForPRID(prID int) string {
	return mergeRefPrefix + strconv.Itoa(prID) + "/merge"
}

func (s *Service) withRepoLock(bareDir string, opts EnsureOptions, run func() error) error {
	if opts.ForceLock {
		return run()
	}

	lockPath := bareDir + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return apperrors.WrapRuntime("failed to open mirror lock", err)
	}
	defer lockFile.Close()

	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = defaultLockTimeout
	}

	deadline := time.Now().Add(lockTimeout)
	for {
		err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}

		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return apperrors.WrapRuntime("failed to acquire mirror lock", err)
		}

		if time.Now().After(deadline) {
			return apperrors.WrapRuntime(
				fmt.Sprintf("timed out waiting for mirror lock after %s; rerun with --force to bypass lock", lockTimeout),
				err,
			)
		}

		time.Sleep(200 * time.Millisecond)
	}

	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	if err := run(); err != nil {
		return err
	}

	return nil
}

func defaultMirrorCacheDir() string {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(".", ".prr", "repos")
	}

	return filepath.Join(userCacheDir, "prr", "repos")
}
