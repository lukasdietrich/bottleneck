package bottleneck

import (
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

func makeMuxHandler(router *Router, r route) httptreemux.HandlerFunc {
	var (
		handler    = wrapHandler(router, r.handler)
		middleware = wrapMiddlewareList(router, r.middleware)
		chain      = makeChain(middleware, handler)
	)

	return func(res http.ResponseWriter, req *http.Request, params map[string]string) {
		ctx := router.contextCreator.create(res, req, params)

		if err := chain(ctx, req); err != nil {
			handleError(ctx.baseContext, err)
		}
	}
}

func makeChain(middleware []wrappedMiddleware, handler wrappedHandler) wrappedHandler {
	if len(middleware) == 0 {
		return handler
	}

	var (
		head = middleware[0]
		next = makeChain(middleware[1:], handler)
	)

	return func(ctx *contextHolder, req *http.Request) error {
		return head(ctx, func() error {
			return next(ctx, req)
		})
	}
}

func handleError(ctx *Context, err error) {
	if err, ok := err.(*Error); ok {
		ctx.JSON(err.Status, err)
	} else {
		handleError(ctx, NewError(http.StatusInternalServerError).WithCause(err))
	}
}
