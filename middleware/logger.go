package middleware

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/lukasdietrich/bottleneck"
)

// Logger creates a middleware that logs requests after the handler is called.
//
// If an error occurs the status is set to 500 or the status of a bottleneck.Error respectively.
// Logger should be added before Recover to log panics.
func Logger() StandardMiddleware {
	return func(ctx *bottleneck.Context, next bottleneck.Next) error {
		var (
			t0      = time.Now()
			err     = next()
			res     = ctx.Response()
			req     = ctx.Request()
			latency = time.Now().Sub(t0)
			status  = res.Status

			bErr *bottleneck.Error
			sErr string
		)

		if err != nil {
			if errors.As(err, &bErr) {
				if bErr.Cause != nil {
					sErr = bErr.Cause.Error()
				}
				status = bErr.Status
			} else {
				sErr = err.Error()
				status = http.StatusInternalServerError
			}
		}

		log.Printf("| %3d | %12s | %6s %s\n%s",
			status,
			latency,
			req.Method,
			req.URL,
			sErr)

		return err
	}
}
