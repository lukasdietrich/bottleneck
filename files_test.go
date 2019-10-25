package bottleneck

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

var (
	errOpen = errors.New("fs open error")
	errStat = errors.New("fs stat error")
)

type testFile struct {
	http.File
}

func (t testFile) Stat() (os.FileInfo, error) {
	return nil, errStat
}

type testFs struct {
	http.FileSystem
}

func (t testFs) Open(name string) (http.File, error) {
	if name == "open.error" {
		return nil, errOpen
	}

	if name == "stat.error" {
		f, _ := t.Open("fallback.html")
		return testFile{f}, nil
	}

	return t.FileSystem.Open(name)
}

func createFileHandlerTestFs(s suite.Suite) http.FileSystem {
	fs := afero.NewMemMapFs()

	for filename, content := range map[string]string{
		"folder/index.html":    "folder index",
		"fallback.html":        "fallback file",
		"very/normal/file.txt": "normal file",
	} {
		s.Require().NoError(fs.MkdirAll(path.Dir(filename), 0700))
		f, err := fs.Create(filename)
		s.Require().NoError(err)
		_, err = f.WriteString(content)
		s.Require().NoError(err)
	}

	return testFs{afero.NewHttpFs(fs)}
}

type FileHandlerTestSuite struct {
	suite.Suite
	router *Router
}

func (s *FileHandlerTestSuite) SetupTest() {
	routes := NewGroup()
	routes.Files("/with-fallback", FileHandlerOptions{
		Fs:       createFileHandlerTestFs(s.Suite),
		NotFound: "fallback.html",
	})
	routes.Files("/", FileHandlerOptions{
		Fs: createFileHandlerTestFs(s.Suite),
	})

	router := NewRouter(routerTestContext{})
	router.Mount(routes)
	s.router = router
}

func (s *FileHandlerTestSuite) TestFallback() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/with-fallback/missingno.png", nil)
		res = httptest.NewRecorder()
	)

	s.router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	s.Equal(200, res.Code)
	s.Equal("fallback file", buf.String())
}

func (s *FileHandlerTestSuite) TestNotFound() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/missingno.png", nil)
		res = httptest.NewRecorder()
	)

	s.router.ServeHTTP(res, req)
	s.Equal(404, res.Code)
}

func (s *FileHandlerTestSuite) TestRegularFile() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/very/normal/file.txt", nil)
		res = httptest.NewRecorder()
	)

	s.router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	s.Equal(200, res.Code)
	s.Equal("normal file", buf.String())
}

func (s *FileHandlerTestSuite) TestIndexFile() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/folder", nil)
		res = httptest.NewRecorder()
	)

	s.router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	s.Equal(200, res.Code)
	s.Equal("folder index", buf.String())
}

func (s *FileHandlerTestSuite) TestFsError() {
	for _, url := range []string{"/open.error", "/stat.error"} {
		var (
			req = httptest.NewRequest(http.MethodGet, url, nil)
			res = httptest.NewRecorder()
		)

		s.router.ServeHTTP(res, req)

		buf := bytes.NewBuffer(nil)
		buf.ReadFrom(res.Result().Body)

		s.Equal(500, res.Code)
	}

}

func TestFileHandler(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}
