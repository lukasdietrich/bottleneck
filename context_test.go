package bottleneck

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextRaw(t *testing.T) {
	ctx := Context{
		request:  httptest.NewRequest(http.MethodGet, "/", nil),
		response: httptest.NewRecorder(),
	}

	assert.Equal(t, ctx.request, ctx.Request())
	assert.Equal(t, ctx.response, ctx.Response())
}

func TestContextParams(t *testing.T) {
	ctx := Context{
		params: map[string]string{
			"name": "Joe",
		},
	}

	assert.Equal(t, "Joe", ctx.Param("name"))
	assert.Equal(t, "", ctx.Param("age"))
}

func TestContextQuery(t *testing.T) {
	ctx := Context{
		request: httptest.NewRequest(http.MethodGet, "/?name=Jake", nil),
	}

	assert.Equal(t, "Jake", ctx.Query("name"))
	assert.Equal(t, "", ctx.Query("age"))
}

func TestContextRenderString(t *testing.T) {
	var (
		recorder = httptest.NewRecorder()
		ctx      = Context{
			response: recorder,
		}
	)

	assert.Nil(t, ctx.String(http.StatusTeapot, "Hello World"))

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(recorder.Result().Body)

	assert.Equal(t, "Hello World", buf.String())
}

func TestContextRenderJSON(t *testing.T) {
	var (
		recorder = httptest.NewRecorder()
		ctx      = Context{
			response: recorder,
		}
	)

	assert.Nil(t, ctx.JSON(http.StatusTeapot, bindTestStruct{
		Name: "Joe",
	}))

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(recorder.Result().Body)

	assert.Equal(t, http.StatusTeapot, recorder.Code)
	assert.Equal(t, MIMEApplicationJSONCharsetUTF8, recorder.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, `{"name":"Joe"}`, buf.String())
}

func TestContextRenderXML(t *testing.T) {
	var (
		recorder = httptest.NewRecorder()
		ctx      = Context{
			response: recorder,
		}
	)

	assert.Nil(t, ctx.XML(http.StatusUpgradeRequired, bindTestStruct{
		Name: "Jake",
	}))

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(recorder.Result().Body)

	assert.Equal(t, http.StatusUpgradeRequired, recorder.Code)
	assert.Equal(t, MIMETextXMLCharsetUTF8, recorder.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, `<person><name>Jake</name></person>`, buf.String())
}

func TestContextRenderStream(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := Context{
		response: recorder,
	}

	assert.Nil(t, ctx.Stream(http.StatusTooManyRequests, MIMEOctetStream, strings.NewReader("Hello World")))

	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(recorder.Result().Body)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, MIMEOctetStream, recorder.HeaderMap.Get(HeaderContentType))
	assert.Equal(t, "Hello World", buf.String())
}