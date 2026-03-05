//go:build windows

package git

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

const (
	lockfileFailImmediately = 0x00000001
	lockfileExclusiveLock   = 0x00000002
)

var (
	kernel32ProcLockFileEx   = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")
	kernel32ProcUnlockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("UnlockFileEx")
)

func tryLockFile(file *os.File) error {
	handle := syscall.Handle(file.Fd())
	overlapped := new(syscall.Overlapped)

	return lockFileEx(
		handle,
		lockfileExclusiveLock|lockfileFailImmediately,
		0,
		1,
		0,
		overlapped,
	)
}

func unlockFile(file *os.File) error {
	handle := syscall.Handle(file.Fd())
	overlapped := new(syscall.Overlapped)

	return unlockFileEx(handle, 0, 1, 0, overlapped)
}

func isLockBusy(err error) bool {
	return errors.Is(err, syscall.Errno(33)) // ERROR_LOCK_VIOLATION
}

func lockFileEx(handle syscall.Handle, flags, reserved, bytesLow, bytesHigh uint32, overlapped *syscall.Overlapped) error {
	r1, _, callErr := kernel32ProcLockFileEx.Call(
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(overlapped)),
	)
	if r1 != 0 {
		return nil
	}

	if callErr != syscall.Errno(0) {
		return callErr
	}

	return syscall.EINVAL
}

func unlockFileEx(handle syscall.Handle, reserved, bytesLow, bytesHigh uint32, overlapped *syscall.Overlapped) error {
	r1, _, callErr := kernel32ProcUnlockFileEx.Call(
		uintptr(handle),
		uintptr(reserved),
		uintptr(bytesLow),
		uintptr(bytesHigh),
		uintptr(unsafe.Pointer(overlapped)),
	)
	if r1 != 0 {
		return nil
	}

	if callErr != syscall.Errno(0) {
		return callErr
	}

	return syscall.EINVAL
}
