package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/spf13/cobra"
)

func readInputJSON(cmd *cobra.Command, target any) (bool, error) {
	reader := cmd.InOrStdin()
	if file, ok := reader.(*os.File); ok {
		info, err := file.Stat()
		if err != nil {
			return false, apperrors.WrapRuntime("failed to inspect stdin", err)
		}
		if (info.Mode() & os.ModeCharDevice) != 0 {
			return false, nil
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return false, apperrors.WrapRuntime("failed to read stdin JSON", err)
	}
	if strings.TrimSpace(string(body)) == "" {
		return false, nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return false, apperrors.WrapConfig("stdin must be valid JSON", err)
	}

	return true, nil
}
