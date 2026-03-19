package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUserConfigReturnsEmptyWhenFileAbsent(t *testing.T) {
	t.TempDir() // ensure a fresh home-like env
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cfg, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.ReviewInstructionsFile != "" {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestLoadUserConfigParsesReviewInstructionsFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	configData := map[string]string{"reviewInstructionsFile": "/path/to/instructions.md"}
	data, _ := json.Marshal(configData)
	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ReviewInstructionsFile != "/path/to/instructions.md" {
		t.Fatalf("expected reviewInstructionsFile, got %q", cfg.ReviewInstructionsFile)
	}
}

func TestLoadUserConfigReturnsErrorOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadUserConfig()
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
}

func TestResolveReviewInstructionsReturnsDefaultWhenNoFile(t *testing.T) {
	instructions := ResolveReviewInstructions(UserConfig{})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions, got %q", instructions)
	}
}

func TestResolveReviewInstructionsReturnsDefaultWhenFileAbsent(t *testing.T) {
	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: "/nonexistent/path/instructions.md"})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions for missing file, got %q", instructions)
	}
}

func TestResolveReviewInstructionsReturnsFileContents(t *testing.T) {
	dir := t.TempDir()
	instructionsPath := filepath.Join(dir, "review.md")
	customInstructions := "Focus on security vulnerabilities and performance bottlenecks."
	if err := os.WriteFile(instructionsPath, []byte(customInstructions), 0600); err != nil {
		t.Fatalf("failed to write instructions file: %v", err)
	}

	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: instructionsPath})
	if instructions != customInstructions {
		t.Fatalf("expected custom instructions %q, got %q", customInstructions, instructions)
	}
}

func TestResolveReviewInstructionsReturnsDefaultForEmptyFile(t *testing.T) {
	dir := t.TempDir()
	instructionsPath := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(instructionsPath, []byte("   \n  "), 0600); err != nil {
		t.Fatalf("failed to write empty instructions file: %v", err)
	}

	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: instructionsPath})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions for empty file, got %q", instructions)
	}
}
