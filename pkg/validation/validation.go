package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

var validate = validator.New()

func ValidateStruct(value any) []FieldError {
	if err := validate.Struct(value); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			out := make([]FieldError, 0, len(validationErrors))
			for _, fieldErr := range validationErrors {
				out = append(out, FieldError{
					Field:   jsonFieldName(value, fieldErr.StructField()),
					Tag:     fieldErr.Tag(),
					Message: humanMessage(fieldErr.StructField(), fieldErr.Tag(), fieldErr.Param()),
				})
			}
			return out
		}
		return []FieldError{{
			Field:   "",
			Tag:     "invalid",
			Message: err.Error(),
		}}
	}
	return nil
}

func jsonFieldName(value any, structField string) string {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if field, ok := typ.FieldByName(structField); ok {
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			name := strings.Split(jsonTag, ",")[0]
			if name != "" && name != "-" {
				return name
			}
		}
	}
	return strings.ToLower(structField)
}

func humanMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	default:
		return fmt.Sprintf("%s failed %s validation", field, tag)
	}
}
