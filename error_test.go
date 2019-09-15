package bottleneck

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorDefaultMessage(t *testing.T) {
	err := NewError(http.StatusNotFound)

	assert.Equal(t, 404, err.Status)
	assert.Equal(t, "Not Found", err.Message)
}

func TestErrorWithMessage(t *testing.T) {
	err := NewError(http.StatusTeapot).WithMessage("Hello World")

	assert.Equal(t, 418, err.Status)
	assert.Equal(t, "Hello World", err.Message)
}

func TestErrorWithCause(t *testing.T) {
	err := NewError(http.StatusInternalServerError).WithCause(io.EOF)

	assert.Equal(t, 500, err.Status)
	assert.Equal(t, io.EOF, errors.Unwrap(err))
	assert.False(t, errors.Is(err, io.ErrClosedPipe))
}

func TestErrorInterface(t *testing.T) {
	err := error(NewError(http.StatusGone).WithCause(io.EOF))

	assert.Equal(t, "status=410 message=Gone (caused by: EOF)", err.Error())
}
