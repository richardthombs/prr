package errors

import (
	"fmt"
)

type Class string

const (
	ClassConfig   Class = "CONFIG_INVALID"
	ClassProvider Class = "PROVIDER_RESOLUTION"
	ClassLimit    Class = "LIMIT_EXCEEDED"
	ClassRuntime  Class = "RUNTIME_FAILURE"
)

const (
	exitConfig = 2
	exitAPI    = 3
	exitLimit  = 4
	exitRun    = 5
)

type AppError struct {
	Class   Class
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Class, e.Message)
	}

	return fmt.Sprintf("%s: %s", e.Class, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func WrapConfig(message string, cause error) error {
	return &AppError{Class: ClassConfig, Message: message, Cause: cause}
}

func WrapProvider(message string, cause error) error {
	return &AppError{Class: ClassProvider, Message: message, Cause: cause}
}

func WrapRuntime(message string, cause error) error {
	return &AppError{Class: ClassRuntime, Message: message, Cause: cause}
}

func WrapLimit(message string, cause error) error {
	return &AppError{Class: ClassLimit, Message: message, Cause: cause}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	appErr, ok := err.(*AppError)
	if !ok {
		return exitRun
	}

	switch appErr.Class {
	case ClassConfig:
		return exitConfig
	case ClassProvider:
		return exitAPI
	case ClassLimit:
		return exitLimit
	default:
		return exitRun
	}
}