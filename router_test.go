package bottleneck

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type routerTestContext struct {
	Context
	Method string
	ID     string
}

type routerTestRequest struct {
	Payload string
}

type routerTestResponse struct {
	Method  string
	ID      string
	Payload string
}

func TestRouterMountGroup(t *testing.T) {
	router := NewRouter(routerTestContext{})
	assert.NotNil(t, router)

	group := NewGroup()
	group.Use(func(*Context, Next) error { return nil })
	group.GET("/", func(*routerTestContext) error { return nil })
	group.POST("/", func(*routerTestContext) error { return nil })
	group.PUT("/", func(*routerTestContext) error { return nil })
	group.DELETE("/", func(*routerTestContext) error { return nil })
	group.OPTIONS("/", func(*routerTestContext) error { return nil })
	group.HEAD("/", func(*routerTestContext) error { return nil })

	group.WithPrefix("/prefix").Mount(group)

	router.Mount(group)
}

func TestRouterServeWithMiddleware(t *testing.T) {
	router := NewRouter(routerTestContext{})
	assert.NotNil(t, router)

	group := NewGroup()

	group.Use(func(ctx *routerTestContext, next Next) error {
		ctx.Method = ctx.Request().Method
		return next()
	})

	group.Use(func(ctx *routerTestContext, next Next) error {
		ctx.ID = ctx.Param("id")
		return next()
	})

	group.POST("/:id", func(ctx *routerTestContext, req *routerTestRequest) error {
		return ctx.JSON(http.StatusOK, routerTestResponse{
			Method:  ctx.Method,
			ID:      ctx.ID,
			Payload: req.Payload,
		})
	})

	router.Mount(group)

	var (
		req = httptest.NewRequest(http.MethodPost, "/42", strings.NewReader(`{"payload": "Jake"}`))
		res = httptest.NewRecorder()
	)

	req.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, MIMEApplicationJSONCharsetUTF8, res.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, `{"Method":"POST","ID":"42","Payload":"Jake"}`, buf.String())
}

func TestRouterServeGenericError(t *testing.T) {
	router := NewRouter(routerTestContext{})
	assert.NotNil(t, router)

	group := NewGroup()
	group.POST("/", func(ctx *routerTestContext) error {
		return errors.New("generic error")
	})

	router.Mount(group)

	var (
		req = httptest.NewRequest(http.MethodPost, "/", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	assert.Equal(t, 500, res.Code)
	assert.Equal(t, MIMEApplicationJSONCharsetUTF8, res.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, `{"status":500,"message":"Internal Server Error"}`, buf.String())
}

func TestRouterServeBottleneckError(t *testing.T) {
	router := NewRouter(routerTestContext{})
	assert.NotNil(t, router)

	group := NewGroup()
	group.POST("/", func(ctx *routerTestContext) error {
		return NewError(http.StatusLengthRequired).WithMessage("Custom Error Message")
	})

	router.Mount(group)

	var (
		req = httptest.NewRequest(http.MethodPost, "/", nil)
		res = httptest.NewRecorder()
	)

	router.ServeHTTP(res, req)

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(res.Result().Body)

	assert.Equal(t, 411, res.Code)
	assert.Equal(t, MIMEApplicationJSONCharsetUTF8, res.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, `{"status":411,"message":"Custom Error Message"}`, buf.String())
}
