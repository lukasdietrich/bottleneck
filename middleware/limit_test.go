package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lukasdietrich/bottleneck"
	"github.com/stretchr/testify/suite"
)

type LimitContext struct {
	bottleneck.Context
}

type LimitRequest struct {
	Value string
}

type LimitTestSuite struct {
	suite.Suite
	router *bottleneck.Router
}

func (s *LimitTestSuite) SetupTest() {
	var (
		router = bottleneck.NewRouter(LimitContext{})
		group  = bottleneck.NewGroup()
	)

	group.Use(Limit(16))
	group.POST("/upload", func(ctx *LimitContext, req *LimitRequest) error {
		return ctx.String(http.StatusOK, req.Value)
	})

	router.Mount(group)
	s.router = router
}

func (s *LimitTestSuite) TestWithinLimit() {
	var (
		req = httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(`{"value":"1234"}`))
		res = httptest.NewRecorder()
	)

	req.Header.Add(bottleneck.HeaderContentType, bottleneck.MIMEApplicationJSONCharsetUTF8)
	s.router.ServeHTTP(res, req)

	s.Equal(200, res.Code)
	s.Equal("1234", string(res.Body.Bytes()))
}

func (s *LimitTestSuite) TestLimitExceeded() {
	var (
		req = httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(`{"value":"12345"}`))
		res = httptest.NewRecorder()
	)

	req.Header.Add(bottleneck.HeaderContentType, bottleneck.MIMEApplicationJSONCharsetUTF8)
	s.router.ServeHTTP(res, req)

	s.Equal(413, res.Code)
}

func TestLimit(t *testing.T) {
	suite.Run(t, new(LimitTestSuite))
}
