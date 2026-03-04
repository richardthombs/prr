package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandMetadata(t *testing.T) {
	if rootCmd.Use != "prr" {
		t.Fatalf("unexpected root command use: %q", rootCmd.Use)
	}
}

func TestPlaceholderCommandsRegistered(t *testing.T) {
	expected := map[string]bool{
		"checkout": false,
		"review":  false,
		"publish": false,
		"version": false,
	}

	for _, cmd := range rootCmd.Commands() {
		if _, ok := expected[cmd.Name()]; ok {
			expected[cmd.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
}

func TestRootHelpRunsSuccessfully(t *testing.T) {
	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"--help"})

	if err := Execute(); err != nil {
		t.Fatalf("expected help command to succeed, got %v", err)
	}

	helpOutput := stdout.String()
	if !strings.Contains(helpOutput, "Usage:") {
		t.Fatalf("expected help output to include Usage, got %q", helpOutput)
	}
	if !strings.Contains(helpOutput, "prr") {
		t.Fatalf("expected help output to reference prr, got %q", helpOutput)
	}
}

func TestVersionCommandOutputsVersionValue(t *testing.T) {
	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"version"})

	if err := Execute(); err != nil {
		t.Fatalf("expected version command to succeed, got %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != version {
		t.Fatalf("expected version output %q, got %q", version, output)
	}
}
