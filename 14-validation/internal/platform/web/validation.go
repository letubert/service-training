package web

import (
	"errors"
	"net/http"

	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

// validator holds the settings and caches for validating request struct values.
var validatorInstance = validator.New()

// translator is a cache of locale and translation information.
var translator *ut.UniversalTranslator

func init() {

	// Instantiate the english locale for the validator library.
	enLocale := en.New()

	// Create a value using English as the fallback locale (first argument).
	// Provide one or more arguments for additional supported locales.
	translator = ut.New(enLocale, enLocale)

	// Register the english error messages for validation errors.
	en_translations.RegisterDefaultTranslations(validatorInstance, translator.GetFallback())
}

// validate must be called with a struct value. It
func validate(val interface{}) error {
	if err := validatorInstance.Struct(val); err != nil {

		// Use a type assertion to get the real error value.
		verr := err.(validator.ValidationErrors)

		// lang controls the language of the error messages. You could pass in the
		// *http.Request and look at the Accept-Language header if you intend to
		// support multiple languages.
		lang := translator.GetFallback()

		var fields []fieldError
		for field, msg := range verr.Translate(lang) {
			fields = append(
				fields,
				fieldError{Field: field, Error: msg},
			)
		}

		return &statusError{
			err:    errors.New("Field validation error"),
			status: http.StatusBadRequest,
			fields: fields,
		}
	}

	return nil
}
