package common

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrForbidden         = errors.New("operation forbidden")
	ErrConflict          = errors.New("resource conflict")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrVersionConflict   = errors.New("optimistic version conflict")
	ErrUnsafeContent     = errors.New("unsafe script content")
	ErrGraphBlocked      = errors.New("dependency graph is blocked")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type CodedError struct {
	Code    string
	Message string
	Fields  []FieldError
	Cause   error
}

func (e *CodedError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *CodedError) Unwrap() error { return e.Cause }

func NewCoded(code, message string, cause error) *CodedError {
	return &CodedError{Code: code, Message: message, Cause: cause}
}
