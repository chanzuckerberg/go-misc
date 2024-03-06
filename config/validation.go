package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// var (
// 	once     sync.Once
// 	validate *validator.Validate
// )

// func init() {
// 	once.Do(func() {
// 		validate = validator.New()
// 	})
// }

func validateConfiguration[T any](cfg *T) error {
	var errs []ValidationError
	// var finalErr error = nil

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		errSlice := &validator.ValidationErrors{}
		errors.As(err, errSlice)
		for _, err := range *errSlice {
			var element ValidationError
			field, _ := reflect.ValueOf(cfg).Type().FieldByName(err.Field())
			element.FailedField = field.Tag.Get("json")
			if element.FailedField == "" {
				element.FailedField = field.Tag.Get("query")
			}
			element.Tag = err.Tag()
			element.Value = err.Param()
			element.Type = err.Kind().String()
			element.Message = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", element.FailedField, element.Tag)

			errs = append(errs, element)
		}
		return errSlice
	}

	return nil
}
