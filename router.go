package bottleneck

import (
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux/v5"
)

// A Handler must be a func with either 1 or 2 arguments and returning an error.
// The first must be either *bottleneck.Context or a pointer to a struct, that embeds bottleneck.Context.
// The second is optional and, if provided, must be a pointer to a struct, which is used to unmarshal and validate
// request payloads.
//
//   type LoginRequest struct {
//     Username string `json:"username"`
//     Password string `json:"password"`
//   }
//
//   type LoginResponse struct {
//     Message string `json:"message"`
//   }
//
//   func login(ctx *bottleneck.Context, req *LoginRequest) error {
//     if req.Username == "AzureDiamond" && req.Password == "hunter2" {
//       return ctx.JSON(http.StatusOK, LoginResponse{
//         Message: "Welcome home, AzureDiamond!"
//       })
//     }
//
//     return ctx.JSON(http.StatusUnauthorized, LoginResponse{
//       Message: "Nice try!"
//     })
//   }
type Handler interface{}

// A Middleware must be a func with exactly two arguments.
// The first must be either *bottleneck.Context or a pointer to a struct, that embeds bottleneck.Context.
// The second must be a Next func, that has to be called to continue the route handling.
//
//   type SessionContext struct {
//     bottleneck.Context
//
//     Username string
//   }
//
//   func checkLogin(ctx *SessionContext, next bottleneck.Next) error {
//     if ctx.Username != "AzureDiamond" {
//       return ctx.String(http.StatusForbidden, "Wrong person.")
//     }
//
//     return next()
//   }
type Middleware interface{}

// Next is the second argument for Middleware. When called it will continue the route handling and return future errors.
type Next func() error

type route struct {
	method     string
	path       string
	handler    Handler
	middleware []Middleware
}

// A Router is a multiplexer for http requests.
type Router struct {
	mux            *httptreemux.TreeMux
	contextCreator *contextCreator

	Binder    Binder
	Validator Validator
}

// NewRouter creates a new Router for a custom context. The provided contextValue is an example instance of the context,
// which is used to get its type.
//
//   type CustomContext struct {
//     bottleneck.Context
//   }
//
//   router := NewRouter(CustomContext{})
//   router.Listen(":8080")
func NewRouter(contextValue interface{}) *Router {
	return &Router{
		mux:            httptreemux.New(),
		contextCreator: newContextCreator(contextValue),

		Binder:    DefaultBinder,
		Validator: DefaultValidator,
	}
}

// Mount adds all roues of a Group to the Router.
func (r *Router) Mount(g *Group) {
	for _, route := range g.routes {
		r.mux.Handle(route.method, route.path, makeMuxHandler(r, route))
	}
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	result, _ := r.mux.Lookup(res, req)
	r.mux.ServeLookupResult(res, req, result)
}

// A Group is a collection of routes. A Group may have a prefix, that is shared across all routes.
type Group struct {
	prefix     string
	routes     []route
	middleware []Middleware
}

// NewGroup creates a new and empty Group.
func NewGroup() *Group {
	return &Group{}
}

// WithPrefix prepends the prefix to all routes of the Group.
func (g *Group) WithPrefix(prefix string) *Group {
	g.prefix = prefix
	return g
}

// Mount adds all routes of the subgroups to this Group.
func (g *Group) Mount(subgroups ...*Group) *Group {
	for _, subgroup := range subgroups {
		for _, route := range subgroup.routes {
			g.Add(route.method, route.path, route.handler, route.middleware...)
		}
	}

	return g
}

// Use adds a list of Middleware to all routes of the Group.
// Middleware added with this method are only used for routes that are added afterwards.
func (g *Group) Use(middleware ...Middleware) *Group {
	g.middleware = append(g.middleware, middleware...)
	return g
}

func (g *Group) relativePath(path string) string {
	return g.prefix + path
}

// Add adds a Handler to the Group.
func (g *Group) Add(method, path string, handler Handler, middleware ...Middleware) *Group {
	m := make([]Middleware, len(g.middleware)+len(middleware))
	copy(m, g.middleware)
	copy(m[len(g.middleware):], middleware)

	g.routes = append(g.routes, route{
		method:     method,
		path:       g.relativePath(path),
		handler:    handler,
		middleware: m,
	})

	return g
}

// GET adds a Handler to the Group with the "GET" http method.
func (g *Group) GET(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodGet, path, handler, middleware...)
}

// POST adds a Handler to the Group with the "POST" http method.
func (g *Group) POST(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodPost, path, handler, middleware...)
}

// PUT adds a Handler to the Group with the "PUT" http method.
func (g *Group) PUT(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodPut, path, handler, middleware...)
}

// DELETE adds a Handler to the Group with the "DELETE" http method.
func (g *Group) DELETE(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodDelete, path, handler, middleware...)
}

// HEAD adds a Handler to the Group with the "HEAD" http method.
func (g *Group) HEAD(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodHead, path, handler, middleware...)
}

// OPTIONS adds a Handler to the Group with the "OPTIONS" http method.
func (g *Group) OPTIONS(path string, handler Handler, middleware ...Middleware) *Group {
	return g.Add(http.MethodOptions, path, handler, middleware...)
}

// Files adds a Handler for static files using http.ServeContent.
func (g *Group) Files(path string, opts FileHandlerOptions) *Group {
	handler := newFileHandler(opts)

	g.GET(path, handler)
	g.GET(strings.TrimSuffix(path, "/")+"/*filepath", handler)

	return g
}
