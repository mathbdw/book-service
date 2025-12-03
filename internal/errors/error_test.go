package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Wrap_ErrorNil(t *testing.T) {
	errNew := Wrap(nil, "Test Wrap")
	assert.Nil(t, errNew)
}

func TestError_Wrap_Success(t *testing.T) {
	err := New("old message")

	err = Wrap(err, "Test Wrap")
	assert.Contains(t, err.Error(), "Test Wrap")
}

func TestError_Unwrap(t *testing.T) {
	err := &Error{cause: New("old message")}

	unw := err.Unwrap()

	assert.Equal(t, err.cause, unw)
}

func TestError_Is_ErrorNil(t *testing.T) {
	err := &Error{}

	assert.False(t, err.Is(nil))
}

func TestError_Is_Success(t *testing.T) {
	err := &Error{msg: "Test err"}

	assert.True(t, err.Is(New("Test err")))
}

func TestError_StackTrace_Empty(t *testing.T) {
	err := &Error{msg: "Test err"}

	assert.Empty(t, err.StackTrace())
}

func TestError_StackTrace(t *testing.T) {
	err := &Error{
		msg:   "Test err",
		stack: callers(),
	}

	stack := err.StackTrace()
	assert.True(t, len(stack) > 0)
}
