package errors

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	ErrNotFound      = New("not found")
	ErrAlreadyExists = New("already exists")
	ErrInvalidInput  = New("invalid input")
	ErrUnauthorized  = New("unauthorized")
	ErrInternal      = New("internal error")
)

// Error - represents a domain error
type Error struct {
	msg   string
	cause error
	stack []uintptr
}

// New - creates a new error
func New(message string) error {
	return &Error{
		msg:   message,
		stack: callers(),
	}
}

// Wrap - wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		msg:   message,
		cause: err,
		stack: callers(),
	}
}

// Error - implements the error interface
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.cause.Error())
	}
	return e.msg
}

// Unwrap - returns the underlying error
func (e *Error) Unwrap() error {
	return e.cause
}

// Is - implements the errors.Is interface
func (e *Error) Is(target error) bool {
	if target == nil {
		return false
	}
	return e.msg == target.Error()
}

func (e *Error) StackTrace() string {
	if len(e.stack) == 0 {
		return ""
	}

	var sb strings.Builder
	frames := runtime.CallersFrames(e.stack)

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}

func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}
