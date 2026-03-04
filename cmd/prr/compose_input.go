package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/spf13/cobra"
)

type composeInput struct {
	PRID     int    `json:"prId"`
	RepoURL  string `json:"repoUrl"`
	Remote   string `json:"remote"`
	Provider string `json:"provider"`
	BareDir  string `json:"bareDir"`
	MergeRef string `json:"mergeRef"`
}

func readOptionalComposeInput(cmd *cobra.Command) (composeInput, bool, error) {
	input := cmd.InOrStdin()
	if stdinFile, ok := input.(*os.File); ok {
		info, err := stdinFile.Stat()
		if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			return composeInput{}, false, nil
		}
	}

	raw, err := io.ReadAll(input)
	if err != nil {
		return composeInput{}, false, apperrors.WrapRuntime("failed to read stdin input", err)
	}

	if strings.TrimSpace(string(raw)) == "" {
		return composeInput{}, false, nil
	}

	var parsed composeInput
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return composeInput{}, false, apperrors.WrapConfig("invalid stdin JSON payload", err)
	}

	return parsed, true, nil
}
