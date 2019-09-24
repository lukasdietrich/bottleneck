package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	assert.NotPanics(t, func() {
		err := Recover()(nil, func() error { panic("Stay calm!") })
		assert.Error(t, err)
		assert.Equal(t, "panic: Stay calm!", err.Error()[:17])
	})
}
