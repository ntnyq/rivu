// Package problem carries language-independent failure codes and context.
package problem

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/ntnyq/rivu/locales"
)

type Code string

const (
	DiskFull        Code = "disk_full"
	Destination     Code = "destination"
	Metadata        Code = "metadata"
	Persistence     Code = "persistence"
	RestoreDisabled Code = "restore_disabled"
	Conflict        Code = "conflict"
	Engine          Code = "engine"
)

type Error struct {
	Code    Code
	Context string
	Cause   error
	Message func(locales.Messages) string
}

func (e *Error) Error() string {
	if e.Message != nil {
		return e.Message(locales.In("en"))
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Context, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Context)
}
func (e *Error) Unwrap() error { return e.Cause }
func New(code Code, context string, cause error) *Error {
	return &Error{Code: code, Context: context, Cause: cause}
}
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var p *Error
	if errors.As(err, &p) && p != nil {
		copy := *p
		return &copy
	}
	code := Engine
	if errors.Is(err, syscall.ENOSPC) {
		code = DiskFull
	} else if errors.Is(err, os.ErrNotExist) || errors.Is(err, os.ErrPermission) {
		code = Destination
	}
	return New(code, "", err)
}

// Message defers translation until presentation and preserves wrapped OS errors.
func Message(code Code, message func(locales.Messages) string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}
