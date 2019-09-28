package middleware

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lukasdietrich/bottleneck"
)

var (
	errUnknownEncoding = errors.New("unknown encoding")
	knownEncodings     = map[string]bool{
		"gzip":    true,
		"deflate": true,
	}
)

func chooseEncoding(acceptEncoding string) string {
	for _, encoding := range strings.Fields(acceptEncoding) {
		if knownEncodings[encoding] {
			return encoding
		}
	}

	return ""
}

func applyEncoding(encoding string, w io.Writer) (io.Writer, error) {
	switch encoding {
	case "gzip":
		return gzip.NewWriterLevel(w, gzip.DefaultCompression)

	case "deflate":
		return flate.NewWriter(w, flate.DefaultCompression)

	default:
		return nil, fmt.Errorf("%w: %s", errUnknownEncoding, encoding)
	}
}

type compressedResponse struct {
	http.ResponseWriter
	writer    io.Writer
	committed bool
}

func (r *compressedResponse) Write(b []byte) (int, error) {
	r.committed = true
	return r.writer.Write(b)
}

func (r *compressedResponse) finalize(ctx *bottleneck.Context) {
	if r.committed {
		if closer, ok := r.writer.(io.Closer); ok {
			closer.Close()
		}
	} else {
		ctx.Response().Header().Del(bottleneck.HeaderContentEncoding)
	}
}

// Compress creates a middleware that compresses the response, if the request accepts it.
//
// Supported encodings are: gzip and deflate.
//
// The weighted preference of accepted encodings is not respected. The first supported encoding
// in the list of accepted encodings will be applied.
func Compress() StandardMiddleware {
	return func(ctx *bottleneck.Context, next bottleneck.Next) error {
		res := ctx.Response()

		res.Header().Add(bottleneck.HeaderVary, bottleneck.HeaderAcceptEncoding)

		if encoding := chooseEncoding(ctx.Request().Header.Get(bottleneck.HeaderAcceptEncoding)); encoding != "" {
			res.Header().Set(bottleneck.HeaderContentEncoding, encoding)

			original := res.Writer
			encoder, err := applyEncoding(encoding, original)
			if err != nil {
				return err
			}

			cr := compressedResponse{
				ResponseWriter: original,
				writer:         encoder,
			}

			defer func() {
				cr.finalize(ctx)
				res.Writer = original
			}()
			res.Writer = &cr
		}

		return next()
	}
}
