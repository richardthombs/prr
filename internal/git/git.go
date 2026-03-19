package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// ExecRunner executes real git commands. When Stderr is non-nil, git's stderr
// output (e.g. clone/fetch progress) is forwarded to that writer in addition
// to being captured for inclusion in any error messages.
type ExecRunner struct {
	Stderr io.Writer
}

func NewExecRunner() Runner {
	return ExecRunner{}
}

// NewExecRunnerWithStderr returns an ExecRunner that forwards git stderr output
// to the provided writer. This allows callers to surface clone and fetch progress
// to the user without interfering with stdout.
func NewExecRunnerWithStderr(stderr io.Writer) Runner {
	return ExecRunner{Stderr: stderr}
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdout = &stdout
	if r.Stderr != nil {
		cmd.Stderr = io.MultiWriter(&stderrBuf, r.Stderr)
	} else {
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	if err != nil {
		combined := strings.TrimSpace(strings.Join([]string{stderrBuf.String(), stdout.String()}, "\n"))
		trimmed := strings.TrimSpace(combined)
		if trimmed == "" {
			return "", err
		}

		return trimmed, fmt.Errorf("%w: %s", err, trimmed)
	}

	return strings.TrimSpace(stdout.String()), nil
}
