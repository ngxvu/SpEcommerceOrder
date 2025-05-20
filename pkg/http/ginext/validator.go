package ginext

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	binding.Validator = NewValidator()
}

var (
	_ ValidatorErrors = &validationErrors{}

	// cheat coverage
	_ = NewValidator().Engine()
)

// NewValidator ...
func NewValidator() binding.StructValidator {
	vr := validator.New()
	vr.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &validatorImpl{vr}
}

// ValidatorFieldError presents a field error (with name & message)
type ValidatorFieldError interface {
	GetField() string
	GetMessage() string
	Error() string
}

// ValidatorErrors presents a validation error
// since it implements ApiError interface, it works will with ErrorHandler
type ValidatorErrors interface {
	GetErrors() []ValidatorFieldError
	Error() string
	GetErrorsMap() map[string]string
	MarshalJSON() ([]byte, error)
	Code() int
}

type validatorImpl struct {
	validate *validator.Validate
}

// ValidateStruct ...
func (v *validatorImpl) ValidateStruct(obj interface{}) error {
	return v.Struct(obj)
}

// Engine ...
func (v *validatorImpl) Engine() interface{} {
	return v.validate
}

type validatorFieldError struct {
	field   string
	message string
}

type validationErrors struct {
	fieldErrors []ValidatorFieldError
}

func (v *validationErrors) Code() int {
	return http.StatusBadRequest
}

// GetErrorsMap return a map of field => message for better responding
func (v *validationErrors) GetErrorsMap() map[string]string {
	errorsMap := make(map[string]string)
	for _, fieldErr := range v.fieldErrors {
		errorsMap[fieldErr.GetField()] = fieldErr.GetMessage()
	}

	return errorsMap
}

// MarshalJSON implements the json.Marshaller interface.
func (v *validationErrors) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.GetErrorsMap())
}

// Error ...
func (v *validationErrors) Error() string {
	return fmt.Sprintf("failed validation on %d field(s)", len(v.fieldErrors))
}

// GetErrors ...
func (v *validationErrors) GetErrors() []ValidatorFieldError {
	return v.fieldErrors
}

// GetField ...
func (v *validatorFieldError) GetField() string {
	return v.field
}

// GetMessage ...
func (v *validatorFieldError) GetMessage() string {
	return v.message
}

// Error ...
func (v *validatorFieldError) Error() string {
	return fmt.Sprintf("%s: %s", v.GetField(), v.GetMessage())
}

// Struct ...
func (v *validatorImpl) Struct(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}

	var fields []ValidatorFieldError
	for _, e := range err.(validator.ValidationErrors) {
		field := &validatorFieldError{
			field: e.Field(),
		}
		tag := ""
		if e.Param() != "" {
			tag = fmt.Sprintf(" (param: %s, value: %v)", e.Param(), e.Value())
		}

		field.message = fmt.Sprintf("failed validation on tag %s%s", e.Tag(), tag)
		fields = append(fields, field)
	}

	return &validationErrors{fieldErrors: fields}
}
