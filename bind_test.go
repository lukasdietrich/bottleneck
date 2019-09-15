package bottleneck

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bindTestStruct struct {
	XMLName xml.Name `xml:"person" json:"-"`
	Name    string   `xml:"name" json:"name"`
}

func TestBindJSON(t *testing.T) {
	var (
		expected = bindTestStruct{
			Name: "Jake",
		}
		raw = `{"name":"Jake"}`
	)

	for _, contentType := range []string{
		MIMEApplicationJSON,
		MIMEApplicationJSONCharsetUTF8,
	} {

		var actual bindTestStruct

		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
		r.Header.Add(HeaderContentType, contentType)

		assert.Nil(t, DefaultBinder.Bind(r, &actual))
		assert.Equal(t, expected, actual)
	}
}

func TestBindXML(t *testing.T) {
	var (
		expected = bindTestStruct{
			XMLName: xml.Name{Space: "", Local: "person"},
			Name:    "Joe",
		}
		raw = `<person><name>Joe</name></person>`
	)

	for _, contentType := range []string{
		MIMETextXML,
		MIMETextXMLCharsetUTF8,
		MIMEApplicationXML,
		MIMEApplicationXMLCharsetUTF8,
	} {
		var actual bindTestStruct

		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))
		r.Header.Add(HeaderContentType, contentType)

		assert.Nil(t, DefaultBinder.Bind(r, &actual))
		assert.Equal(t, expected, actual)
	}
}

func TestBindUnsupported(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add(HeaderContentType, MIMETextPlain)

	err := DefaultBinder.Bind(r, &bindTestStruct{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrBindUnsupportedContentType))
}
