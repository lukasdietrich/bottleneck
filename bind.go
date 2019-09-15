package bottleneck

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
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
	defer r.Body.Close()

	switch contentType := r.Header.Get(HeaderContentType); contentType {
	case MIMEApplicationJSON, MIMEApplicationJSONCharsetUTF8:
		return json.NewDecoder(r.Body).Decode(v)

	case MIMETextXML, MIMETextXMLCharsetUTF8, MIMEApplicationXML, MIMEApplicationXMLCharsetUTF8:
		return xml.NewDecoder(r.Body).Decode(v)

	default:
		return fmt.Errorf("%w: %s", ErrBindUnsupportedContentType, contentType)
	}
}
