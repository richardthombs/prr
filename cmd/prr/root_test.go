package main

import "testing"

func TestRootCommandMetadata(t *testing.T) {
	if rootCmd.Use != "prr" {
		t.Fatalf("unexpected root command use: %q", rootCmd.Use)
	}
}

func TestPlaceholderCommandsRegistered(t *testing.T) {
	expected := map[string]bool{
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
