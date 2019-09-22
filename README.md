# Bottleneck Go router

[![Build Status](https://dev.azure.com/lukasdietrich/bottleneck/_apis/build/status/lukasdietrich.bottleneck?branchName=master)](https://dev.azure.com/lukasdietrich/bottleneck/_build/latest?definitionId=2&branchName=master)
![Azure DevOps tests (master)](https://img.shields.io/azure-devops/tests/lukasdietrich/bottleneck/2/master)
[![codecov](https://codecov.io/gh/lukasdietrich/bottleneck/branch/master/graph/badge.svg)](https://codecov.io/gh/lukasdietrich/bottleneck)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukasdietrich/bottleneck)](https://goreportcard.com/report/github.com/lukasdietrich/bottleneck)
[![GoDoc](https://godoc.org/github.com/lukasdietrich/bottleneck?status.svg)](https://godoc.org/github.com/lukasdietrich/bottleneck)

Bottleneck is a web framework for [Go](https://golang.org/) with a focus on convenience over raw performance.
Most routers try to be the *best* by being a few cpu cycles faster or using fewer allocations per request.
I'm going to go out on a limb here by saying that in many real world applications and even more so in amateur projects
the performance benefits of the *faster* libraries do not really make a difference.

## Motivation

While working on different ideas involving web and Go I used quite a few of the libraries
the community has to offer. All of them did something better than others, but never
everything. Bottleneck is the aggregation of things I personally liked.

### Context from [gin-gonic/gin](https://github.com/gin-gonic/gin)

Many routers use a pattern that is not compatible with the standard library, but
is something I much prefer and first saw in gin. The pattern in question is to use
a single "context" instead of `*http.Request` and `http.ResponseWriter` and provide some
convenience methods on top of it. The advantage is to have lots of utilities available
in every handler and to have a slightly shorter function signature.

### Error handling from [labstack/echo](https://github.com/labstack/echo)

Most routers define route handlers as functions that do not return anything.
That starts to become annoying when you do a lot of `if err != nil { return err }`
and maybe want to handle errors at a centralized place, for example through middleware.

Echo is the first framework I came across that returns an error in every handler and
middleware which made it much more pleasant for me to write services.

### Mounting subgroups from [go-chi/chi](https://github.com/go-chi/chi)

Many routers support grouping of routes, but most of them require them to be
*defined with* the router. This makes defining groups of routes as a separate package
a lot harder / less modular. Chi allows groups to be created separately and then be
mounted onto a router later.

### Bring your own context from [gocraft/web](https://github.com/gocraft/web)

A common pattern for web services is to do some work or checks before a group of routes.
This can be achieved with middleware in many of the existing routers, but storing
something to be accessed down the chain often requires type assertions. gocraft/web allows
you to *bring your own context* and therefore makes it possible to avoid type assertions by
putting that burden on the web framework.

### Automatic request unmarshalling

Handling a request is (almost) always the same. Unmarshal and possibly validate a request. 
Do some work. Marshal a response. Done.

Marshalling a response can be hidden behind the context, but unmarshalling and validating
the request always requires some boilerplate at the start of each handler. I want Bottleneck
to do that automatically and provide the resulting value as a parameter to handlers.

## Example

```go
package main

import (
	"net/http"

	"github.com/lukasdietrich/bottleneck"
)

// Context is a custom context
type Context struct {
	bottleneck.Context        // It must embed bottleneck.Context to be valid
	User               string // Store user during request handling
}

// AmIRequest is a request definition
type AmIRequest struct {
	Name string `json:"name" validate:"required"`
}

// AmIResponse is a response definition
type AmIResponse struct {
	Yes bool `json:"yes"`
}

func main() {
	r := bottleneck.NewRouter(Context{}) // Create a new router
	r.Mount(routes())                    // Mount the routes

	// Start a http server and use the router
	http.ListenAndServe(":8080", r)
}

// routes creates a group with all /api routes.
func routes() *bottleneck.Group {
	g := bottleneck.NewGroup().WithPrefix("/api")

	// store the current user for handlers
	g.Use(func(ctx *Context, next bottleneck.Next) error {
		user, _, ok := ctx.Request().BasicAuth()
		if !ok {
			return bottleneck.NewError(http.StatusUnauthorized)
		}

		ctx.User = user
		return next()
	})

	// test if the user knows his own name
	g.POST("/test", func(ctx *Context, req *AmIRequest) error {
		return ctx.JSON(http.StatusOK, AmIResponse{
			Yes: req.Name == ctx.User,
		})
	})

	return g
}
```

## Credits

Bottleneck utilizes some great packages under the hood:

1. [dimfeld/httptreemux](https://github.com/dimfeld/httptreemux) for request routing
2. [go-playground/validator](https://github.com/go-playground/validator) for the default
   validation implemenation