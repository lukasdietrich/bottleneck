package middleware

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukasdietrich/bottleneck"
	"github.com/stretchr/testify/assert"
)

type LoggerContext struct {
	bottleneck.Context
}

func TestLogger(t *testing.T) {
	var (
		router = bottleneck.NewRouter(LoggerContext{})
		group  = bottleneck.NewGroup()
		buf    bytes.Buffer
	)

	log.SetOutput(&buf)

	group.Use(Logger())
	group.DELETE("/", func(ctx *LoggerContext) error {
		return ctx.String(http.StatusOK, "OK")
	})
	group.PUT("/err-500", func(*LoggerContext) error {
		return errors.New("standard error")
	})
	group.GET("/err-400", func(*LoggerContext) error {
		return bottleneck.NewError(http.StatusBadRequest)
	})

	router.Mount(group)

	for pattern, request := range map[string]*http.Request{
		`\| 200 \| \s*\d+\.\d+.?s \| DELETE /`:        httptest.NewRequest(http.MethodDelete, "/", nil),
		`\| 500 \| \s*\d+\.\d+.?s \| \s*PUT /err-500`: httptest.NewRequest(http.MethodPut, "/err-500", nil),
		`\| 400 \| \s*\d+\.\d+.?s \| \s*GET /err-400`: httptest.NewRequest(http.MethodGet, "/err-400", nil),
	} {
		buf.Reset()
		router.ServeHTTP(httptest.NewRecorder(), request)
		assert.Regexp(t, pattern, buf.String())
	}
}
