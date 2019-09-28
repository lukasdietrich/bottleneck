package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasdietrich/bottleneck"
	"github.com/stretchr/testify/suite"
)

type CompressContext struct {
	bottleneck.Context
}

type CompressTestSuite struct {
	suite.Suite
	router *bottleneck.Router
}

func (s *CompressTestSuite) SetupTest() {
	var (
		router = bottleneck.NewRouter(CompressContext{})
		group  = bottleneck.NewGroup()
	)

	group.Use(Compress())
	group.GET("/content", func(ctx *CompressContext) error {
		return ctx.String(http.StatusOK, "Content")
	})

	router.Mount(group)
	s.router = router
}

func (s *CompressTestSuite) TestGzip() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/content", nil)
		res = httptest.NewRecorder()
		buf bytes.Buffer
	)

	req.Header.Add(bottleneck.HeaderAcceptEncoding, "gzip")
	s.router.ServeHTTP(res, req)

	r, err := gzip.NewReader(res.Body)
	s.Nil(err)
	buf.ReadFrom(r)

	s.Equal(200, res.Code)
	s.Equal("gzip", res.Header().Get(bottleneck.HeaderContentEncoding))
	s.Equal("Content", buf.String())
}

func (s *CompressTestSuite) TestDeflate() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/content", nil)
		res = httptest.NewRecorder()
		buf bytes.Buffer
	)

	req.Header.Add(bottleneck.HeaderAcceptEncoding, "deflate")
	s.router.ServeHTTP(res, req)

	buf.ReadFrom(flate.NewReader(res.Body))

	s.Equal(200, res.Code)
	s.Equal("deflate", res.Header().Get(bottleneck.HeaderContentEncoding))
	s.Equal("Content", buf.String())
}

func (s *CompressTestSuite) TestNone() {
	var (
		req = httptest.NewRequest(http.MethodGet, "/content", nil)
		res = httptest.NewRecorder()
		buf bytes.Buffer
	)

	req.Header.Add(bottleneck.HeaderAcceptEncoding, "br")
	s.router.ServeHTTP(res, req)

	buf.ReadFrom(res.Body)

	s.Equal(200, res.Code)
	s.Equal("", res.Header().Get(bottleneck.HeaderContentEncoding))
	s.Equal("Content", buf.String())
}

func TestCompress(t *testing.T) {
	suite.Run(t, new(CompressTestSuite))
}
