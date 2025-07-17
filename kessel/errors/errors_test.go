package errors

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestWrap_ErrorHandling(t *testing.T) {
	// Test nil cause returns nil
	result := Wrap(ErrConnectionFailed, nil, "test message")
	if result != nil {
		t.Errorf("Wrap() with nil cause should return nil, got %v", result)
	}

	// Test wrapping error successfully
	cause := errors.New("network error")
	result = Wrap(ErrTokenRetrieval, cause, "failed to retrieve token")

	if result == nil {
		t.Fatal("Wrap() returned nil, expected error")
	}

	// Check if sentinel error is properly wrapped
	if !errors.Is(result, ErrTokenRetrieval) {
		t.Errorf("Wrap() result should wrap sentinel error %v", ErrTokenRetrieval)
	}

	// Check if it's the correct type
	var wrappedErr *Error
	if !errors.As(result, &wrappedErr) {
		t.Errorf("Wrap() returned %T, want *Error", result)
	}
}

func TestNew_ErrorCreation(t *testing.T) {
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

func TestConvenienceFunction_ErrorTypes(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(error, string) error
		checker func(error) bool
	}{
		{
			name:    "NewConnectionError",
			fn:      NewConnectionError,
			checker: IsConnectionError,
		},
		{
			name:    "NewTokenError",
			fn:      NewTokenError,
			checker: IsTokenError,
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

			if !tt.checker(err) {
				t.Errorf("checker function should return true for error type")
			}
		})
	}
}

func TestErrorCheckers_CrossTypes(t *testing.T) {
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
