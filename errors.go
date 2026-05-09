package spec

import "strings"

type ValidationErrors struct { //nolint:errname // ValidationErrors is a better name than ErrorsError or ValidationError here
	Errors []error
}

func (e ValidationErrors) Error() string {
	parts := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}
	return strings.Join(parts, "; ")
}

func (e ValidationErrors) Unwrap() []error {
	return e.Errors
}

func joinErrors(errs []error) error {
	var filtered []error
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return ValidationErrors{Errors: filtered}
}
