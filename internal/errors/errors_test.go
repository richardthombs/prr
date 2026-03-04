package errors

import (
	"errors"
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
		{name: "runtime", err: WrapRuntime("runtime failed", nil), want: 4},
		{name: "unknown", err: errors.New("unknown"), want: 4},
	}

	for _, testCase := range tests {
		if got := ExitCode(testCase.err); got != testCase.want {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.want, got)
		}
	}
}