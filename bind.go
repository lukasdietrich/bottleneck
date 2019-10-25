package bottleneck

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
)

const (
	structTagQuery = "query"
	structTagForm  = "form"
)

var (
	// DefaultBinder is the default Binder implementation.
	// It handels unmarshalling of JSON and XML encoded payloads depending on the Content-Type header of the request.
	// If the Content-Type is neither JSON nor XML ErrBindSupportedContentType is returned.
	DefaultBinder Binder = defaultBinder{}

	// ErrBindUnsupportedContentType indicates that a request could not be bound, because the Content-Type is not
	// supported by the Binder implementation.
	ErrBindUnsupportedContentType = errors.New("cannot bind content type")
)

// A Binder is used to unmarshal incoming requests.
type Binder interface {
	// Bind unmarshalls an incoming request. The first argument is the raw *http.Request.The second argument is the
	// target struct.
	Bind(*http.Request, interface{}) error
}

type defaultBinder struct{}

func (defaultBinder) Bind(r *http.Request, v interface{}) error {
	if r.Body != nil {
		defer r.Body.Close()
	}

	if r.Method == http.MethodGet {
		return decodeValues(structTagQuery, r.URL.Query(), v)
	}

	switch contentType := r.Header.Get(HeaderContentType); contentType {
	case MIMEApplicationJSON, MIMEApplicationJSONCharsetUTF8:
		return decodeJSON(r.Body, v)

	case MIMETextXML, MIMETextXMLCharsetUTF8, MIMEApplicationXML, MIMEApplicationXMLCharsetUTF8:
		return decodeXML(r.Body, v)

	case MIMEApplicationForm:
		if err := r.ParseForm(); err != nil {
			return err
		}

		return decodeValues(structTagForm, r.Form, v)

	default:
		return fmt.Errorf("%w: %s", ErrBindUnsupportedContentType, contentType)
	}
}

func decodeJSON(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func decodeXML(r io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(r)
	decoder.Strict = true
	return decoder.Decode(v)
}

func decodeValues(structTag string, values url.Values, v interface{}) error {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(false)
	decoder.SetAliasTag(structTag)
	return decoder.Decode(v, values)
}
