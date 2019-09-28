package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/lukasdietrich/bottleneck"
)

// ErrPanic is used when capturing a panic and wrapping it as an error.
// This error is never returned directly. You can test if an error is caused by a panic using the errors package:
//
//   if errors.Is(err, ErrPanic) {
//      // do something
//   }
var ErrPanic = errors.New("panic")

func recoverToError(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%w: %v\n%s", ErrPanic, r, debug.Stack())
	}
}

func recoverNext(err *error, next bottleneck.Next) {
	defer recoverToError(err)
	*err = next()
}

// Recover creates a middleware that recovers from panics and wraps them as errors.
//
// The resulting error embeds ErrPanic and has the form "panic: {panic value}\n{stacktrace}".
func Recover() StandardMiddleware {
	return func(ctx *bottleneck.Context, next bottleneck.Next) (err error) {
		recoverNext(&err, next)
		return
	}
}
