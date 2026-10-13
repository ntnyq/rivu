// Package problem carries language-independent failure codes and context.
package problem

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"syscall"

	"github.com/ntnyq/rivu/internal/i18n"
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
	Key     string
	Args    []any
}

func (e *Error) Error() string {
	if e.Key != "" {
		return fmt.Sprintf(strings.ReplaceAll(i18n.Translate("en", e.Key), "%w", "%v"), e.Args...)
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
		copy.Args = slices.Clone(p.Args)
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
func Message(code Code, key string, args ...any) *Error {
	e := &Error{Code: code, Key: key, Args: slices.Clone(args)}
	for _, arg := range args {
		if cause, ok := arg.(error); ok {
			e.Cause = cause
			break
		}
	}
	return e
}
