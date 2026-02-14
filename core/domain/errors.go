package domain

import "fmt"

// Code represents a domain error classification.
type Code string

const (
	CodeNotFound   Code = "NOT_FOUND"
	CodeConflict   Code = "CONFLICT"
	CodeValidation Code = "VALIDATION"
	CodeForbidden  Code = "FORBIDDEN"
	CodeInvariant  Code = "INVARIANT_VIOLATED"
)

// Error is the domain error type used across all layers.
// It carries a classification code, a human-readable message,
// and an optional wrapped error for chain compatibility.
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is supports errors.Is() matching by code.
// Two domain errors are considered equal if they share the same Code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewError creates a domain error with the given code and message.
func NewError(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// WrapError creates a domain error that wraps an underlying error.
func WrapError(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Sentinel errors for use with errors.Is().
var (
	ErrNotFound   = &Error{Code: CodeNotFound}
	ErrConflict   = &Error{Code: CodeConflict}
	ErrValidation = &Error{Code: CodeValidation}
	ErrForbidden  = &Error{Code: CodeForbidden}
	ErrInvariant  = &Error{Code: CodeInvariant}
)
