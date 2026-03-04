package main

import (
	"os"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

func main() {
	if err := Execute(); err != nil {
		os.Exit(apperrors.ExitCode(err))
	}
}
