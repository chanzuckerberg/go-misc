package config

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

var (
	once               sync.Once
	validate           *validator.Validate
	translatorMessages map[string](func() string)
	translator         ut.Translator
)

func init() {
	once.Do(func() {
		validate = validator.New()
		// err := validate.RegisterValidation("valid_env", ValidateEnvironment)
		// if err != nil {
		// 	logrus.Fatal("Failed to register custom validation")
		// }
		// err = validate.RegisterValidation("valid_env_dest", ValidateEnvironmentCopyDestination)
		// if err != nil {
		// 	logrus.Fatal("Failed to register custom validation")
		// }

		en := en.New()
		uni := ut.New(en, en)
		translator, _ = uni.GetTranslator("en")

		translatorMessages = map[string](func() string){
			// "valid_env": func() string {
			// 	envs := []string{}
			// 	for env := range model.EnvironmentMapping {
			// 		envs = append(envs, env)
			// 	}
			// 	return fmt.Sprintf("{0} must be one of %s", envs)
			// },
			// "valid_env_dest": func() string {
			// 	return "Copying configs from source env to destination env as specified is not allowed"
			// },
		}

		for tag, getMessage := range translatorMessages {
			err := validate.RegisterTranslation(
				tag,
				translator,
				func(ut ut.Translator) error {
					return ut.Add(tag, getMessage(), true) // see universal-translator for details
				},
				// use a function that returns a function here so the tag can be memoized
				func(violatedTag string) validator.TranslationFunc {
					return func(ut ut.Translator, fe validator.FieldError) string {
						t, _ := ut.T(violatedTag, fe.Field())
						return t
					}
				}(tag),
			)
			if err != nil {
				return fmt.Errorf("Failed to register custom validation error translator: %w", err)
			}
		}
	})
}

func validateConfiguration[T any](cfg *T) []ValidationError {
	var errs []ValidationError

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

			if _, ok := translatorMessages[element.Tag]; ok {
				element.Message = err.Translate(translator)
			} else {
				element.Message = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", element.FailedField, element.Tag)
			}

			errs = append(errs, &element)
		}
	}
	return errs
}
