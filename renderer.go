package bottleneck

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// A Renderer is a container type that wraps an abstract http response.
type Renderer interface {
	// Header is called before Render and is used to set http headers.
	// The standard implementations use this method to set the response Content-Type.
	Header(http.Header)

	// Render is called after Header and is used to write the raw http body.
	Render(io.Writer) error
}

// StringRenderer implements the Renderer interface.
type StringRenderer struct {
	// String is the exact value written in Render
	String string
}

// Header sets the Content-Type to "text/plain; charset=UTF8".
func (StringRenderer) Header(h http.Header) {
	h.Add(HeaderContentType, MIMETextPlainCharsetUTF8)
}

// Render writes the String as is to w.
func (r StringRenderer) Render(w io.Writer) error {
	_, err := io.WriteString(w, r.String)
	return err
}

// JSONRenderer implements the Renderer interface.
type JSONRenderer struct {
	Value interface{}
}

// Header sets the Content-Type to "application/json; charset=UTF8".
func (JSONRenderer) Header(h http.Header) {
	h.Add(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
}

// Render marshals the Value as JSON and then writes it to w.
func (r JSONRenderer) Render(w io.Writer) error {
	b, err := json.Marshal(r.Value)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

// XMLRenderer implements the Renderer interface.
type XMLRenderer struct {
	Value interface{}
}

// Header sets the Content-Type to "text/xml; charset=UTF8".
func (XMLRenderer) Header(h http.Header) {
	h.Add(HeaderContentType, MIMETextXMLCharsetUTF8)
}

// Render marshals the Value as XML and then writes it w.
func (r XMLRenderer) Render(w io.Writer) error {
	b, err := xml.Marshal(r.Value)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

// StreamRenderer implements the Renderer interface.
type StreamRenderer struct {
	ContentType string
	Reader      io.Reader
}

// Header sets the Content-Type to the value of ContentType.
func (r StreamRenderer) Header(h http.Header) {
	h.Add(HeaderContentType, r.ContentType)
}

// Render pipes the contents of Reader to w.
func (r StreamRenderer) Render(w io.Writer) error {
	_, err := io.Copy(w, r.Reader)
	return err
}
