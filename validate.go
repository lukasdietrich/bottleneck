package bottleneck

import (
	"net/http"

	"gopkg.in/go-playground/validator.v9"
)

// DefaultValidator is the default Validator implementation using https://github.com/go-playground/validator.
var DefaultValidator Validator = defaultValidator{validator.New()}

// A Validator is used to validate incoming requests.
type Validator interface {
	// Validate validates an incoming request. The first argument is the raw *http.Request. The second argument is the
	// unmarshalled payload.
	Validate(*http.Request, interface{}) error
}

type defaultValidator struct {
	validate *validator.Validate
}

func (d defaultValidator) Validate(_ *http.Request, v interface{}) error {
	return d.validate.Struct(v)
}
