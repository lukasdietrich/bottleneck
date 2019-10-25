package bottleneck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type validateTestStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"required,min=18"`
}

func TestDefaultValidatorValid(t *testing.T) {
	validStruct := validateTestStruct{
		Name: "Jake",
		Age:  35,
	}

	assert.NoError(t, DefaultValidator.Validate(nil, &validStruct))
}

func TestDefaultValidatorInvalid(t *testing.T) {
	invalidStruct := validateTestStruct{
		Name: "Joe",
		Age:  17,
	}

	assert.Error(t, DefaultValidator.Validate(nil, &invalidStruct))
}
