package middleware

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/lukasdietrich/bottleneck"
)

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

func Recover() MiddlewareFunc {
	return func(ctx *bottleneck.Context, next bottleneck.Next) (err error) {
		recoverNext(&err, next)
		return
	}
}
