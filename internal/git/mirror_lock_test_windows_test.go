//go:build windows

package git

import (
	"os"
)

func holdTestLock(file *os.File) (func(), error) {
	if err := tryLockFile(file); err != nil {
		return nil, err
	}

	return func() {
		_ = unlockFile(file)
	}, nil
}
