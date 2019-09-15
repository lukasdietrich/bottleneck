package bottleneck

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type reflectTestContext struct {
	Context
}

func TestNewContextCreatorPanic(t *testing.T) {
	type Context struct{}

	for _, invalidContext := range []interface{}{
		"Context",
		struct{}{},
		struct{ Context string }{},
		struct{ Context }{},
	} {
		assert.Panics(t, func() { newContextCreator(invalidContext) })
	}
}

func TestContextCreatorCreate(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	var (
		res    = httptest.NewRecorder()
		req    = httptest.NewRequest(http.MethodPost, "/", nil)
		params = make(map[string]string)
	)

	ctxHolder := creator.create(res, req, params)
	assert.NotNil(t, ctxHolder)

	var (
		context     *reflectTestContext
		baseContext *Context
	)

	context = ctxHolder.unwrap(reflect.TypeOf(context)).Interface().(*reflectTestContext)
	baseContext = ctxHolder.unwrap(reflect.TypeOf(baseContext)).Interface().(*Context)

	assert.NotNil(t, context)
	assert.NotNil(t, baseContext)
	assert.Equal(t, context.Context, *baseContext)
	assert.Equal(t, req, baseContext.request)
	assert.Equal(t, res, baseContext.response)
	assert.Equal(t, params, baseContext.params)

	assert.Panics(t, func() {
		ctxHolder.unwrap(reflect.TypeOf(0))
	})
}

func TestContextCreatorValidate(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	assert.Nil(t, creator.validateTarget(reflect.TypeOf(&reflectTestContext{})))
	assert.Nil(t, creator.validateTarget(reflect.TypeOf(&Context{})))
	assert.Error(t, creator.validateTarget(reflect.TypeOf(0)))
}

func TestValidateHandlerPass(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	for _, fn := range []interface{}{
		func(*Context, *struct{}) error { return nil },
		func(*reflectTestContext, *struct{}) error { return nil },
		func(*Context) error { return nil },
		func(*reflectTestContext) error { return nil },
	} {
		assert.Nil(t, validateHandler(creator, reflect.TypeOf(fn)))
	}
}

func TestValidateHandlerFail(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	for _, fn := range []interface{}{
		0,
		func() {},
		func(Context) {},
		func(*Context) {},
		func() error { return nil },
		func(*Context, int) error { return nil },
		func(int) error { return nil },
		func(int) {},
		func(reflectTestContext, *struct{}) error { return nil },
		func(*Context, struct{}) error { return nil },
		func(*reflectTestContext, struct{}) error { return nil },
		func(*Context, *int) error { return nil },
	} {
		assert.Error(t, validateHandler(creator, reflect.TypeOf(fn)))
	}
}

func TestWrapHandler(t *testing.T) {
	router := NewRouter(reflectTestContext{})
	assert.NotNil(t, router)

	assert.NotNil(t, wrapHandler(router, func(*reflectTestContext) error { return nil }))
	assert.NotNil(t, wrapHandler(router, func(*reflectTestContext, *struct{}) error { return nil }))
	assert.Panics(t, func() { wrapHandler(router, 0) })
}

func TestValidateMiddlewarePass(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	for _, fn := range []interface{}{
		func(*Context, Next) error { return nil },
		func(*reflectTestContext, Next) error { return nil },
	} {
		assert.Nil(t, validateMiddleware(creator, reflect.TypeOf(fn)))
	}
}

func TestValidateMiddlewareFail(t *testing.T) {
	creator := newContextCreator(reflectTestContext{})
	assert.NotNil(t, creator)

	for _, fn := range []interface{}{
		0,
		func() {},
		func(Context) {},
		func(*Context) {},
		func() error { return nil },
		func(*Context, int) error { return nil },
		func(int) error { return nil },
		func(int) {},
		func(reflectTestContext, *struct{}) error { return nil },
		func(*Context, struct{}) error { return nil },
		func(*reflectTestContext, struct{}) error { return nil },
		func(*Context, *int) error { return nil },
		func(*Context, Next) {},
		func(*Context, Next) int { return 0 },
	} {
		assert.Error(t, validateMiddleware(creator, reflect.TypeOf(fn)))
	}
}

func TestWrapMiddleware(t *testing.T) {
	router := NewRouter(reflectTestContext{})
	assert.NotNil(t, router)

	assert.NotNil(t, wrapMiddleware(router, func(*reflectTestContext, Next) error { return nil }))
	assert.Panics(t, func() { wrapMiddleware(router, 0) })
}
