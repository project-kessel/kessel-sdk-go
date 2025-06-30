package errors

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error with cause",
			err: &Error{
				Code:    codes.Internal,
				Message: "test error",
				Cause:   errors.New("underlying cause"),
			},
			expected: "test error: underlying cause",
		},
		{
			name: "error without cause",
			err: &Error{
				Code:    codes.Internal,
				Message: "test error",
				Cause:   nil,
			},
			expected: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestError_GRPCStatus(t *testing.T) {
	err := &Error{
		Code:    codes.InvalidArgument,
		Message: "invalid input",
		Cause:   nil,
	}

	status := err.GRPCStatus()
	if status.Code() != codes.InvalidArgument {
		t.Errorf("GRPCStatus().Code() = %v, want %v", status.Code(), codes.InvalidArgument)
	}
	if status.Message() != "invalid input" {
		t.Errorf("GRPCStatus().Message() = %v, want %v", status.Message(), "invalid input")
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying cause")
	err := &Error{
		Code:    codes.Internal,
		Message: "test error",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Error.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		sentinel  error
		cause     error
		message   string
		expectNil bool
	}{
		{
			name:      "nil cause returns nil",
			sentinel:  ErrConnectionFailed,
			cause:     nil,
			message:   "test message",
			expectNil: true,
		},
		{
			name:      "wraps error successfully",
			sentinel:  ErrTokenRetrieval,
			cause:     errors.New("network error"),
			message:   "failed to retrieve token",
			expectNil: false,
		},
		{
			name:      "wraps gRPC status error",
			sentinel:  ErrConnectionFailed,
			cause:     status.Error(codes.Unavailable, "service unavailable"),
			message:   "connection failed",
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.sentinel, tt.cause, tt.message)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Wrap() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Wrap() returned nil, expected error")
			}

			// Check if the wrapped error is of correct type
			var wrappedErr *Error
			if !errors.As(result, &wrappedErr) {
				t.Errorf("Wrap() returned %T, want *Error", result)
			}

			// Check if sentinel error is properly wrapped
			if !errors.Is(result, tt.sentinel) {
				t.Errorf("Wrap() result should wrap sentinel error %v", tt.sentinel)
			}
		})
	}
}

func TestNew(t *testing.T) {
	sentinel := ErrTokenRetrieval
	code := codes.Unauthenticated
	message := "authentication failed"

	err := New(sentinel, code, message)

	var wrappedErr *Error
	if !errors.As(err, &wrappedErr) {
		t.Fatalf("New() returned %T, want *Error", err)
	}

	if wrappedErr.Code != code {
		t.Errorf("New() Code = %v, want %v", wrappedErr.Code, code)
	}
	if wrappedErr.Message != message {
		t.Errorf("New() Message = %v, want %v", wrappedErr.Message, message)
	}
	if wrappedErr.Cause != sentinel {
		t.Errorf("New() Cause = %v, want %v", wrappedErr.Cause, sentinel)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(error, string) error
		sentinel error
		checker  func(error) bool
	}{
		{
			name:     "NewConnectionError",
			fn:       NewConnectionError,
			sentinel: ErrConnectionFailed,
			checker:  IsConnectionError,
		},
		{
			name:     "NewTokenError",
			fn:       NewTokenError,
			sentinel: ErrTokenRetrieval,
			checker:  IsTokenError,
		},
		{
			name:     "NewHTTPClientError",
			fn:       NewHTTPClientError,
			sentinel: ErrHTTPClientCreation,
			checker:  IsHTTPClientError,
		},
		{
			name:     "NewResourceCloseError",
			fn:       NewResourceCloseError,
			sentinel: ErrResourceClose,
			checker:  IsResourceCloseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := errors.New("test cause")
			message := "test message"

			err := tt.fn(cause, message)

			if err == nil {
				t.Fatal("convenience function returned nil")
			}

			if !errors.Is(err, tt.sentinel) {
				t.Errorf("convenience function should wrap sentinel error %v", tt.sentinel)
			}

			if !tt.checker(err) {
				t.Errorf("checker function should return true for error type")
			}
		})
	}
}

func TestNewTokenCacheError(t *testing.T) {
	message := "token not found in cache"
	err := NewTokenCacheError(message)

	if !IsTokenCacheError(err) {
		t.Error("NewTokenCacheError should create token cache error")
	}

	var wrappedErr *Error
	if !errors.As(err, &wrappedErr) {
		t.Fatalf("NewTokenCacheError returned %T, want *Error", err)
	}

	if wrappedErr.Code != codes.NotFound {
		t.Errorf("NewTokenCacheError Code = %v, want %v", wrappedErr.Code, codes.NotFound)
	}
}

func TestNewStatusError(t *testing.T) {
	statusCode := 404
	message := "resource not found"

	err := NewStatusError(statusCode, message)

	if !IsStatusError(err) {
		t.Error("NewStatusError should create status error")
	}

	// The error message will include the sentinel error because NewStatusError uses New()
	// which sets the sentinel as the cause
	expectedMessage := fmt.Sprintf("%s: status code %d: %s", message, statusCode, ErrUnexpectedStatus.Error())
	if err.Error() != expectedMessage {
		t.Errorf("NewStatusError message = %v, want %v", err.Error(), expectedMessage)
	}
}

func TestErrorCheckers(t *testing.T) {
	// Test that checkers return false for wrong error types
	connectionErr := NewConnectionError(errors.New("test"), "test")
	tokenErr := NewTokenError(errors.New("test"), "test")

	if IsTokenError(connectionErr) {
		t.Error("IsTokenError should return false for connection error")
	}

	if IsConnectionError(tokenErr) {
		t.Error("IsConnectionError should return false for token error")
	}

	// Test with regular errors
	regularErr := errors.New("regular error")
	if IsConnectionError(regularErr) {
		t.Error("IsConnectionError should return false for regular error")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that all sentinel errors are properly defined
	sentinels := []error{
		ErrConnectionFailed,
		ErrTokenRetrieval,
		ErrTokenCacheNotFound,
		ErrUnexpectedStatus,
		ErrHTTPClientCreation,
		ErrResourceClose,
	}

	for i, sentinel := range sentinels {
		if sentinel == nil {
			t.Errorf("Sentinel error %d is nil", i)
		}
		if sentinel.Error() == "" {
			t.Errorf("Sentinel error %d has empty message", i)
		}
	}
}
