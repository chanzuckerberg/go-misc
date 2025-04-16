package errors

import (
	"errors"
	"fmt"
)

/*
/ This package is meant to simplify the way we treat server-side errors. In particular,
/ we want to be diligent about what information is safe to return to the clients.
/ Only the public parts of errors will be returned to the caller, the rest can be logged server side.
*/

func optionalString(s string) *string {
	return &s
}

func GetPublicMessage(err error) *string {
	pubErr := &publicError{}
	ok := errors.As(err, &pubErr)

	if ok {
		return optionalString(pubErr.GetPublicMessage())
	}
	return nil
}

func GetInternalError(err error) error {
	pubErr := &publicError{}
	ok := errors.As(err, &pubErr)
	if ok {
		return pubErr.GetInternalError()
	}
	return nil
}

// PublicError represents a ca error with a message that is safe to share with users
type publicError struct {
	cause         error
	publicMessage string
}

// PublicWrapf wraps an error with a public message
func PublicWrapf(err error, publicMsgFormat string, args ...interface{}) error {
	return PublicWrap(err, fmt.Sprintf(publicMsgFormat, args...))
}

// PublicWrap wraps an error with a public message
func PublicWrap(err error, publicMsg string) error {
	if err == nil {
		return nil
	}
	return &publicError{cause: err, publicMessage: publicMsg}
}

// NewPublic return a new error with public message
func NewPublic(msg string) error {
	return &publicError{publicMessage: msg}
}

// NewPublicf returns a new formatted error with public message
func NewPublicf(format string, args ...interface{}) error {
	return NewPublic(fmt.Sprintf(format, args...))
}

func IsPublicError(err error) bool {
	pubErr := &publicError{}
	return errors.As(err, &pubErr)
}

// Error returns the public part of this error
func (e *publicError) Error() string {
	return e.GetPublicMessage()
}

// GetPublicMessage returns this error's public message
func (e *publicError) GetPublicMessage() string {
	return e.publicMessage
}

// GetInternalError returns the internal part of this error
func (e *publicError) GetInternalError() error {
	if e == nil {
		return nil
	}
	return e.cause
}
