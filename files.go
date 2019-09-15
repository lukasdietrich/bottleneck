package bottleneck

import (
	"errors"
	"net/http"
	"os"
	"path"
)

// FileHandlerOptions define how static files are handled.
type FileHandlerOptions struct {
	// Fs is used to resolve files.
	Fs http.FileSystem
	// NotFound is the filename to use, when no file exists. If empty 404 is returned.
	NotFound string
}

func newFileHandler(opts FileHandlerOptions) Handler {
	return func(ctx *Context) error {
		return serveFile(&opts, ctx, ctx.Param("filepath"), opts.NotFound != "")
	}
}

func serveFile(opts *FileHandlerOptions, ctx *Context, filepath string, handleNotFound bool) error {
	f, err := opts.Fs.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if handleNotFound {
				return serveFile(opts, ctx, opts.NotFound, false)
			}

			return NewError(http.StatusNotFound).WithCause(err)
		}

		return NewError(http.StatusInternalServerError).WithCause(err)
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return NewError(http.StatusInternalServerError).WithCause(err)
	}

	if info.IsDir() {
		return serveFile(opts, ctx, path.Join(filepath, "index.html"), handleNotFound)
	}

	http.ServeContent(ctx.Response(), ctx.Request(), info.Name(), info.ModTime(), f)
	return nil
}
