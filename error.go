package midas

import (
	"errors"
	"fmt"
)

const (
	ErrUnauthorized = "unauthorized"
	ErrInternal     = "internal"
	ErrInvalid      = "invalid"
	ErrUnaccepted   = "unaccepted"
	ErrRegistry     = "registry"
)

// Error represents an application-specific error. App errors can be
// unwrapped to extract out the code & message.
//
// Any non-application errors (such as a disk error) should be reported
// as en ErrInternal error, so the end-user should only see "internal error".
// These error details should be logged to the operator of application.
type Error struct {
	// Machine-readable error code
	Code string

	// Human-readable error message
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("server error: code: %s message: %s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
// Non-application errors always return ErrInternal.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}

	return ErrInternal
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error"
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}

	return "Internal error."
}

// Errorf is a helper function to return an Error with a given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
