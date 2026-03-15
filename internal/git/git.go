package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type ExecRunner struct{}

func NewExecRunner() Runner {
	return ExecRunner{}
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		combined := strings.TrimSpace(strings.Join([]string{stderr.String(), stdout.String()}, "\n"))
		trimmed := strings.TrimSpace(combined)
		if trimmed == "" {
			return "", err
		}

		return trimmed, fmt.Errorf("%w: %s", err, trimmed)
	}

	return strings.TrimSpace(stdout.String()), nil
}
