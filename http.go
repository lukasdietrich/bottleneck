package bottleneck

const (
	charsetUTF8 = "charset=utf-8"
)

// Some well-known http header keys.
const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderVary            = "Vary"
)

// Some well-known content types.
//
// See https://www.iana.org/assignments/media-types/media-types.xhtml.
const (
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationXML             = "application/xml"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; " + charsetUTF8
	MIMEMultipartForm              = "multipart/form-data"
	MIMEOctetStream                = "application/octet-stream"
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; " + charsetUTF8
	MIMETextXML                    = "text/xml"
	MIMETextXMLCharsetUTF8         = MIMETextXML + "; " + charsetUTF8
)
