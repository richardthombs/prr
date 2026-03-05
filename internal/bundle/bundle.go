package bundle

import (
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

type Limits struct {
	MaxPatchBytes   int
	MaxChangedFiles int
}

func BuildV1(input types.DiffOutput, limits Limits) (types.BundleV1, error) {
	if err := validateInput(input); err != nil {
		return types.BundleV1{}, err
	}

	bundle := types.BundleV1{
		Version:      "v1",
		PRID:         input.PRID,
		RepoURL:      strings.TrimSpace(input.RepoURL),
		Remote:       strings.TrimSpace(input.Remote),
		Provider:     strings.TrimSpace(input.Provider),
		MergeRef:     strings.TrimSpace(input.MergeRef),
		Range:        strings.TrimSpace(input.Range),
		Files:        input.Files,
		Stat:         strings.TrimSpace(input.Stat),
		Patch:        input.Patch,
		ChangedFiles: len(input.Files),
		PatchBytes:   len([]byte(input.Patch)),
	}

	if err := validateLimits(bundle, limits); err != nil {
		return types.BundleV1{}, err
	}

	return bundle, nil
}

func validateInput(input types.DiffOutput) error {
	if strings.TrimSpace(input.Range) == "" {
		return apperrors.WrapConfig("bundle input is missing diff range", nil)
	}
	if strings.TrimSpace(input.Stat) == "" {
		return apperrors.WrapConfig("bundle input is missing diff stat", nil)
	}
	if strings.TrimSpace(input.Patch) == "" {
		return apperrors.WrapConfig("bundle input is missing unified patch", nil)
	}
	if input.Files == nil {
		return apperrors.WrapConfig("bundle input is missing changed files list", nil)
	}

	return nil
}

func validateLimits(bundle types.BundleV1, limits Limits) error {
	if limits.MaxPatchBytes < 0 {
		return apperrors.WrapConfig("max patch bytes must be zero or positive", nil)
	}
	if limits.MaxChangedFiles < 0 {
		return apperrors.WrapConfig("max changed files must be zero or positive", nil)
	}

	if limits.MaxPatchBytes > 0 && bundle.PatchBytes > limits.MaxPatchBytes {
		return apperrors.WrapLimit(
			"patch exceeds maxPatchBytes; reduce diff scope or increase limit",
			nil,
		)
	}

	if limits.MaxChangedFiles > 0 && bundle.ChangedFiles > limits.MaxChangedFiles {
		return apperrors.WrapLimit(
			"changed files exceed maxChangedFiles; reduce diff scope or increase limit",
			nil,
		)
	}

	return nil
}
