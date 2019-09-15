package bottleneck

import (
	"fmt"
	"net/http"
)

// Error is a user displayable error that is returned during request handling.
type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

// NewError creates a new Error and sets the http status code, which will be set when not handeled manually.
func NewError(status int) *Error {
	return &Error{
		Status:  status,
		Message: http.StatusText(status),
	}
}

// WithMessage adds a custom message to the Error. By default http.StatusText is used to create a message.
func (e *Error) WithMessage(message string) *Error {
	e.Message = message
	return e
}

// WithCause adds an actual error to the Error, which can later be unwrapped and tested for.
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// Unwrap returns the wrapped cause. This is useful to test for specific errors.
//
//   err := NewError(http.StatusInternalServerError).WithCause(io.EOF)
//   errors.Is(err, io.EOF) // true
func (e *Error) Unwrap() error {
	return e.Cause
}

// Error formats the error as readable text.
func (e *Error) Error() string {
	return fmt.Sprintf("status=%d message=%s (caused by: %v)", e.Status, e.Message, e.Cause)
}
