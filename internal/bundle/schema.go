package bundle

import (
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

func ValidateV1Schema(payload types.BundleV1) error {
	if strings.TrimSpace(payload.Version) != "v1" {
		return apperrors.WrapConfig("bundle version must be v1", nil)
	}
	if strings.TrimSpace(payload.Range) == "" {
		return apperrors.WrapConfig("bundle range is required", nil)
	}
	if payload.Files == nil {
		return apperrors.WrapConfig("bundle files list is required", nil)
	}
	if strings.TrimSpace(payload.Stat) == "" {
		return apperrors.WrapConfig("bundle stat is required", nil)
	}
	if strings.TrimSpace(payload.Patch) == "" {
		return apperrors.WrapConfig("bundle patch is required", nil)
	}

	return nil
}
