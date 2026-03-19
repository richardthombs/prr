package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
			_, runErr := s.runCommand(ctx, opts, "git", "-C", bareDir, "fetch", "--all", "--prune", "--progress")
			if runErr != nil {
				return "", runErr
			}

			return bareDir, nil
		} else if !os.IsNotExist(statErr) {
			return "", apperrors.WrapRuntime("failed to inspect bare mirror path", statErr)
		}

		_, runErr := s.runCommand(ctx, opts, "git", "clone", "--mirror", "--progress", strings.TrimSpace(repoURL), bareDir)
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
			_, runErr := s.runCommand(ctx, opts, "git", "-C", bareDir, "fetch", "--all", "--prune", "--progress")
			if runErr != nil {
				return apperrors.WrapRuntime("failed to update bare mirror", runErr)
			}

			return nil
		} else if !os.IsNotExist(statErr) {
			return apperrors.WrapRuntime("failed to inspect bare mirror path", statErr)
		}

		_, runErr := s.runCommand(ctx, opts, "git", "clone", "--mirror", "--progress", strings.TrimSpace(repoURL), bareDir)
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
	return s.FetchPRMergeRefWithOptions(ctx, bareDir, remote, prID, EnsureOptions{})
}

func (s *Service) FetchPRMergeRefWithOptions(ctx context.Context, bareDir, remote string, prID int, opts EnsureOptions) (string, error) {
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

	_, err := s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "fetch", "--progress", trimmedRemote, destination)
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

	repoSlug := repoSlugFromURL(trimmedRepo)

	return filepath.Join(s.cacheDir, repoSlug+".git"), nil
}

func repoSlugFromURL(rawRepoURL string) string {
	provider := "repo"
	primarySegment := "unknown"
	repoSegment := "repo"

	if parsedURL, err := url.Parse(rawRepoURL); err == nil && strings.TrimSpace(parsedURL.Host) != "" {
		host := strings.ToLower(strings.TrimSpace(parsedURL.Host))
		provider = providerFromHost(host)

		pathSegments := pathSegmentsFromRepoURL(parsedURL.Path)
		// dev.azure.com includes the org as the first path segment; skip it so
		// the slug uses project+repo, consistent with the *.visualstudio.com format.
		if host == "dev.azure.com" && len(pathSegments) > 2 {
			pathSegments = pathSegments[1:]
		}
		scope, repo := scopeAndRepoFromSegments(pathSegments)
		if scope != "" {
			primarySegment = scope
		}
		if repo != "" {
			repoSegment = repo
		}
	} else {
		slug := sanitizeSlugPart(rawRepoURL)
		if slug != "" {
			return slug
		}
	}

	return strings.Join([]string{provider, primarySegment, repoSegment}, "-")
}

func providerFromHost(host string) string {
	switch {
	case host == "github.com":
		return "github"
	case host == "dev.azure.com":
		return "azure"
	case strings.HasSuffix(host, ".visualstudio.com"):
		return "azure"
	default:
		firstLabel := host
		if idx := strings.Index(firstLabel, "."); idx > 0 {
			firstLabel = firstLabel[:idx]
		}
		sanitized := sanitizeSlugPart(firstLabel)
		if sanitized == "" {
			return "repo"
		}

		return sanitized
	}
}

func pathSegmentsFromRepoURL(path string) []string {
	rawSegments := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		trimmed := strings.TrimSpace(segment)
		if trimmed == "" || strings.EqualFold(trimmed, "_git") {
			continue
		}

		cleaned := strings.TrimSuffix(trimmed, ".git")
		cleaned = sanitizeSlugPart(cleaned)
		if cleaned != "" {
			segments = append(segments, cleaned)
		}
	}

	return segments
}

func scopeAndRepoFromSegments(segments []string) (string, string) {
	if len(segments) == 0 {
		return "", ""
	}

	if len(segments) == 1 {
		return segments[0], segments[0]
	}

	return segments[0], segments[len(segments)-1]
}

func sanitizeSlugPart(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	builder := strings.Builder{}
	builder.Grow(len(trimmed))
	lastWasDash := false
	for _, char := range trimmed {
		isAlphaNum := (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
		if isAlphaNum {
			builder.WriteRune(char)
			lastWasDash = false
			continue
		}

		if !lastWasDash {
			builder.WriteRune('-')
			lastWasDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}


func MergeRefForPRID(prID int) string {
	return mergeRefPrefix + strconv.Itoa(prID) + "/merge"
}

func HeadRefForPRID(prID int) string {
	return mergeRefPrefix + strconv.Itoa(prID) + "/head"
}

func (s *Service) FetchPRHeadRef(ctx context.Context, bareDir, remote string, prID int, opts EnsureOptions) (string, error) {
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

	headRef := HeadRefForPRID(prID)
	sourceRef := "pull/" + strconv.Itoa(prID) + "/head"
	destination := sourceRef + ":" + headRef

	_, err := s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "fetch", "--progress", trimmedRemote, destination)
	if err != nil {
		return "", apperrors.WrapProvider("failed to fetch PR head ref", err)
	}

	return headRef, nil
}

func (s *Service) ResolveMergeBase(ctx context.Context, bareDir, ref1, ref2 string, opts EnsureOptions) (string, error) {
	trimmedBareDir := strings.TrimSpace(bareDir)
	if trimmedBareDir == "" {
		return "", apperrors.WrapConfig("bare mirror directory is required; provide --bare-dir", nil)
	}

	output, err := s.runCommand(ctx, opts, "git", "-C", trimmedBareDir, "merge-base", ref1, ref2)
	if err != nil {
		return "", apperrors.WrapRuntime("failed to resolve merge base", err)
	}

	base := strings.TrimSpace(output)
	if base == "" && !opts.WhatIf {
		return "", apperrors.WrapRuntime("merge-base returned empty result", nil)
	}

	return base, nil
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
		err = tryLockFile(lockFile)
		if err == nil {
			break
		}

		if !isLockBusy(err) {
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

	defer func() {
		_ = unlockFile(lockFile)
	}()

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
