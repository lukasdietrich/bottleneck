package bottleneck

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createFileHandlerTestFs(t *testing.T) http.FileSystem {
	fs := afero.NewMemMapFs()

	for filename, content := range map[string]string{
		"folder/index.html":    "folder index",
		"fallback.html":        "fallback file",
		"very/normal/file.txt": "normal file",
	} {
		require.Nil(t, fs.MkdirAll(path.Dir(filename), 0700))
		f, err := fs.Create(filename)
		require.Nil(t, err)
		_, err = f.WriteString(content)
		require.Nil(t, err)
	}

	return afero.NewHttpFs(fs)
}

func TestFileHandlerWithNotFound(t *testing.T) {
	router := NewRouter(routerTestContext{})
	router.Mount(NewGroup().Files("/", FileHandlerOptions{
		Fs:       createFileHandlerTestFs(t),
		NotFound: "fallback.html",
	}))

	var (
		req = httptest.NewRequest(http.MethodGet, "/missingno.png", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body) // nolint:errcheck

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "fallback file", buf.String())
}

func TestFileHandlerWithoutNotFound(t *testing.T) {
	router := NewRouter(routerTestContext{})
	router.Mount(NewGroup().Files("/", FileHandlerOptions{
		Fs: createFileHandlerTestFs(t),
	}))

	var (
		req = httptest.NewRequest(http.MethodGet, "/missingno.png", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)
	assert.Equal(t, 404, res.Code)
}

func TestFileHandlerRegularFile(t *testing.T) {
	router := NewRouter(routerTestContext{})
	router.Mount(NewGroup().Files("/", FileHandlerOptions{
		Fs: createFileHandlerTestFs(t),
	}))

	var (
		req = httptest.NewRequest(http.MethodGet, "/very/normal/file.txt", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body) // nolint:errcheck

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "normal file", buf.String())
}

func TestFileHandlerIndexFile(t *testing.T) {
	router := NewRouter(routerTestContext{})
	router.Mount(NewGroup().Files("/", FileHandlerOptions{
		Fs: createFileHandlerTestFs(t),
	}))

	var (
		req = httptest.NewRequest(http.MethodGet, "/folder", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body) // nolint:errcheck

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "folder index", buf.String())
}
