package bottleneck

import (
	"io"
	"net/http"
)

// Response wraps a raw http.ResponseWriter and stores additional information, which is not tracked
// by the standard library.
type Response struct {
	// Status is the http status code that is set during WriteHeader.
	Status int
	// Writer is the raw http.ResponseWriter. It can be set in a middleware to change the way data is written.
	Writer http.ResponseWriter
}

// Header returns the header map that will be sent by WriteHeader.
//
// See https://golang.org/pkg/net/http/#ResponseWriter
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
//
// See https://golang.org/pkg/net/http/#ResponseWriter
func (r *Response) Write(b []byte) (int, error) {
	return r.Writer.Write(b)
}

// WriteHeader sends an HTTP response header with the provided status code.
//
// See https://golang.org/pkg/net/http/#ResponseWriter
func (r *Response) WriteHeader(status int) {
	r.Status = status
	r.Writer.WriteHeader(status)
}

// Context is the base for custom contexts. It is a container for the raw http request and response and provides
// convenience methods to access request-data and to write responses.
type Context struct {
	request  *http.Request
	response *Response
	params   map[string]string
}

func (c *Context) init(res http.ResponseWriter, req *http.Request, params map[string]string) {
	c.response = &Response{
		Writer: res,
		Status: http.StatusOK,
	}
	c.request = req
	c.params = params
}

// Request returns the raw http request.
func (c *Context) Request() *http.Request {
	return c.request
}

// Response returns the raw http response.
func (c *Context) Response() *Response {
	return c.response
}

// Param returns the path parameter of the current matched route. If the parameter does not exist, an empty string is
// returned instead.
//
//   router.GET("/users/:name", func(ctx *Context) error {
//     return ctx.String(http.StatusOK, ctx.Param("name"))
//   })
func (c *Context) Param(key string) string {
	return c.params[key]
}

// Query returns the query value for a given key. If the key does not exist, an empty string is returned instead.
//
//   // curl "localhost:8080/search?input=Does router performance matter in Go?"
//   router.GET("/search", func(ctx *Context) error {
//     return ctx.String(http.StatusOK, ctx.Query("input"))
//   })
func (c *Context) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

// Render writes a generic response using the provided Renderer after the status-code is set.
func (c *Context) Render(status int, r Renderer) error {
	r.Header(c.response.Header())
	c.response.WriteHeader(status)
	return r.Render(c.response)
}

// String writes a response using the StringRenderer.
func (c *Context) String(status int, value string) error {
	return c.Render(status, StringRenderer{String: value})
}

// JSON writes a response using the JSONRenderer.
func (c *Context) JSON(status int, value interface{}) error {
	return c.Render(status, JSONRenderer{Value: value})
}

// XML writes a response using the XMLRenderer.
func (c *Context) XML(status int, value interface{}) error {
	return c.Render(status, XMLRenderer{Value: value})
}

// Stream writes a response using the StreamRenderer.
func (c *Context) Stream(status int, contentType string, reader io.Reader) error {
	return c.Render(status, StreamRenderer{
		ContentType: contentType,
		Reader:      reader,
	})
}
