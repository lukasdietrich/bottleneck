package middleware

import (
	"errors"
	"io"
	"net/http"

	"github.com/lukasdietrich/bottleneck"
)

var errLimitExceeded = errors.New("request size limit exceeded")

type limitedReader struct {
	reader io.ReadCloser
	limit  int64
	count  int64
}

func (l *limitedReader) Close() error {
	return l.reader.Close()
}

func (l *limitedReader) Read(b []byte) (int, error) {
	if l.count < l.limit {
		if remaining := l.limit - l.count; remaining < int64(len(b)) {
			b = b[:remaining]
		}

		n, err := l.reader.Read(b)
		l.count += int64(n)
		return n, err
	}

	return 0, bottleneck.NewError(http.StatusRequestEntityTooLarge).WithCause(errLimitExceeded)
}

// Limit creates a middleware that returns an error when a request body is larger than size.
func Limit(size int64) StandardMiddleware {
	return func(ctx *bottleneck.Context, next bottleneck.Next) error {
		if req := ctx.Request(); req.Body != nil {
			req.Body = &limitedReader{
				reader: req.Body,
				count:  0,
				limit:  size,
			}
		}

		return next()
	}
}
