package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok {
		return &Error{
			Code:    s.Code(),
			Message: message,
			Cause:   err,
		}
	}

	return &Error{
		Code:    codes.Unknown,
		Message: message,
		Cause:   err,
	}
}

func New(code codes.Code, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}
