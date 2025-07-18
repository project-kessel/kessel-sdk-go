package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Sentinel errors for common scenarios
var (
	ErrConnectionFailed   = errors.New("connection failed")
	ErrTokenRetrieval     = errors.New("token retrieval failed")
	ErrTokenCacheNotFound = errors.New("cached token not found")
	ErrUnexpectedStatus   = errors.New("unexpected status code")
	ErrHTTPClientCreation = errors.New("HTTP client creation failed")
	ErrResourceClose      = errors.New("resource close failed")
)

type Error struct {
	Code    codes.Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

// Unwrap returns the underlying cause, enabling errors.Is and errors.As
func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap wraps an error with a sentinel error for easier identification
func Wrap(sentinel error, cause error, message string) error {
	if cause == nil {
		return nil
	}

	code := codes.Unknown
	if s, ok := status.FromError(cause); ok {
		code = s.Code()
	}

	return &Error{
		Code:    code,
		Message: message,
		Cause:   fmt.Errorf("%w: %v", sentinel, cause),
	}
}

// New creates a new error with a sentinel for categorization
func New(sentinel error, code codes.Code, message string) error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   sentinel,
	}
}

// Convenience functions for common error types
func NewConnectionError(cause error, message string) error {
	return Wrap(ErrConnectionFailed, cause, message)
}

func NewTokenError(cause error, message string) error {
	return Wrap(ErrTokenRetrieval, cause, message)
}

func NewTokenCacheError(message string) error {
	return New(ErrTokenCacheNotFound, codes.NotFound, message)
}

func NewStatusError(statusCode int, message string) error {
	return New(ErrUnexpectedStatus, codes.InvalidArgument,
		fmt.Sprintf("%s: status code %d", message, statusCode))
}

func NewHTTPClientError(cause error, message string) error {
	return Wrap(ErrHTTPClientCreation, cause, message)
}

func NewResourceCloseError(cause error, message string) error {
	return Wrap(ErrResourceClose, cause, message)
}

// Helper functions to check for specific error types
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

func IsTokenError(err error) bool {
	return errors.Is(err, ErrTokenRetrieval)
}

func IsTokenCacheError(err error) bool {
	return errors.Is(err, ErrTokenCacheNotFound)
}

func IsStatusError(err error) bool {
	return errors.Is(err, ErrUnexpectedStatus)
}

func IsHTTPClientError(err error) bool {
	return errors.Is(err, ErrHTTPClientCreation)
}

func IsResourceCloseError(err error) bool {
	return errors.Is(err, ErrResourceClose)
}
