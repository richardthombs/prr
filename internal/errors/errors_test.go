package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestExitCodeMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: 0},
		{name: "config", err: WrapConfig("bad config", nil), want: 2},
		{name: "provider", err: WrapProvider("provider failed", nil), want: 3},
		{name: "engine", err: WrapEngine("engine failed", nil), want: 6},
		{name: "limit", err: WrapLimit("limit exceeded", nil), want: 4},
		{name: "runtime", err: WrapRuntime("runtime failed", nil), want: 5},
		{name: "unknown", err: errors.New("unknown"), want: 5},
	}

	for _, testCase := range tests {
		if got := ExitCode(testCase.err); got != testCase.want {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.want, got)
		}
	}
}

func TestAppErrorIncludesCauseInMessage(t *testing.T) {
	cause := errors.New("adapter failed")
	err := WrapEngine("failed to run review engine", cause)

	message := err.Error()
	if !strings.Contains(message, "ENGINE_FAILURE") {
		t.Fatalf("expected class in message, got %q", message)
	}
	if !strings.Contains(message, "failed to run review engine") {
		t.Fatalf("expected top-level message, got %q", message)
	}
	if !strings.Contains(message, "adapter failed") {
		t.Fatalf("expected cause details in message, got %q", message)
	}
}
