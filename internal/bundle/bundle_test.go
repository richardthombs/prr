package bundle

import (
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/types"
)

func TestBuildV1ProducesBundlePayload(t *testing.T) {
	input := types.DiffOutput{
		PRID:     42,
		RepoURL:  "https://github.com/acme/repo",
		Remote:   "origin",
		Provider: "github",
		MergeRef: "refs/prr/pull/42/merge",
		Range:    "HEAD^1..HEAD",
		Files:    []string{"a.txt", "b.txt"},
		Stat:     "2 files changed",
		Patch:    "diff --git a/a.txt b/a.txt",
	}

	payload, err := BuildV1(input, Limits{})
	if err != nil {
		t.Fatalf("expected bundle build to succeed, got %v", err)
	}

	if payload.Version != "v1" {
		t.Fatalf("expected version v1, got %q", payload.Version)
	}
	if payload.ChangedFiles != 2 {
		t.Fatalf("expected changedFiles 2, got %d", payload.ChangedFiles)
	}
	if payload.PatchBytes <= 0 {
		t.Fatalf("expected patchBytes to be greater than zero, got %d", payload.PatchBytes)
	}

	if err := ValidateV1Schema(payload); err != nil {
		t.Fatalf("expected schema validation to pass, got %v", err)
	}
}

func TestBuildV1RejectsMissingRequiredFields(t *testing.T) {
	_, err := BuildV1(types.DiffOutput{Range: "HEAD^1..HEAD", Files: []string{"a.txt"}, Stat: "stat"}, Limits{})
	if err == nil {
		t.Fatalf("expected build failure for missing patch")
	}
	if !strings.Contains(err.Error(), "missing unified patch") {
		t.Fatalf("expected missing patch diagnostic, got %v", err)
	}
}

func TestBuildV1EnforcesPatchByteLimit(t *testing.T) {
	_, err := BuildV1(types.DiffOutput{
		Range: "HEAD^1..HEAD",
		Files: []string{"a.txt"},
		Stat:  "1 file changed",
		Patch: "0123456789",
	}, Limits{MaxPatchBytes: 5})
	if err == nil {
		t.Fatalf("expected patch size limit failure")
	}
	if !strings.Contains(err.Error(), "LIMIT_EXCEEDED") {
		t.Fatalf("expected LIMIT_EXCEEDED diagnostic, got %v", err)
	}
}

func TestBuildV1EnforcesChangedFilesLimit(t *testing.T) {
	_, err := BuildV1(types.DiffOutput{
		Range: "HEAD^1..HEAD",
		Files: []string{"a.txt", "b.txt"},
		Stat:  "2 files changed",
		Patch: "diff --git",
	}, Limits{MaxChangedFiles: 1})
	if err == nil {
		t.Fatalf("expected changed files limit failure")
	}
	if !strings.Contains(err.Error(), "LIMIT_EXCEEDED") {
		t.Fatalf("expected LIMIT_EXCEEDED diagnostic, got %v", err)
	}
}
