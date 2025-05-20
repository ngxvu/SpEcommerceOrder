package ginext

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type userValidation struct {
	FirstName string `validate:"required"`
	LastName  string `validate:"required" json:"last_name"`
	Age       uint8  `validate:"gte=0,lte=150" json:"-"`
	Email     string `validate:"email"`
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	_, ok := v.(*validatorImpl)
	assert.True(t, ok)
}

func TestValidateNoError(t *testing.T) {
	testData := userValidation{
		FirstName: "Tony",
		LastName:  "Huynh",
		Age:       1,
		Email:     "Tony.huynh@gmail.com",
	}
	err := NewValidator().ValidateStruct(testData)
	assert.NoError(t, err)
}

func TestValidateFailedWithJSONField(t *testing.T) {
	testData := userValidation{
		FirstName: "Tony",
		LastName:  "",
		Age:       1,
		Email:     "Tony.huynh@gmail.com",
	}
	err := NewValidator().ValidateStruct(testData)
	vErrors, _ := err.(ValidatorErrors)
	assert.Contains(t, vErrors.GetErrors()[0].Error(), "last_name: failed validation on tag required")
}

func TestValidateFailedWithParamDetailIfSet(t *testing.T) {
	testData := userValidation{
		FirstName: "Tony",
		LastName:  "Huynh",
		Age:       200,
		Email:     "Tony.huynh@gmail.com",
	}
	err := NewValidator().ValidateStruct(testData)
	vErrors, _ := err.(ValidatorErrors)
	assert.Contains(t, vErrors.GetErrors()[0].Error(), "Age: failed validation on tag lte (param: 150, value: 200)")
}

func TestReturnErrorOnInvalidStruct(t *testing.T) {
	err := NewValidator().ValidateStruct(nil)
	assert.Error(t, err)
	_, ok := err.(*validator.InvalidValidationError)
	assert.True(t, ok)
}
