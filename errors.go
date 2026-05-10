package spec

import (
	"errors"
	"strings"

	"github.com/oaswrap/spec/internal/validate"
)

// Severity represents the severity level of a validation error.
type Severity = validate.Severity

const (
	// SeverityError indicates a strict validation failure.
	SeverityError = validate.SeverityError
	// SeverityWarning indicates a validation warning that doesn't necessarily invalidate the document.
	SeverityWarning = validate.SeverityWarning
	// SeverityInfo indicates informational validation feedback.
	SeverityInfo = validate.SeverityInfo
)

// ValidationError represents a validation error with an associated severity level.
type ValidationError = validate.Error

// ValidationErrors is a collection of validation errors that can be returned by the Validate method of various structs in the spec package. It implements the error interface and can be used to aggregate multiple validation errors into a single error value.
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

// HasSeverity returns true if the collection contains at least one error with the given severity.
func (e ValidationErrors) HasSeverity(s Severity) bool {
	for _, err := range e.Errors {
		if err == nil {
			continue
		}
		var valErr validate.Error
		var valErrPtr *validate.Error
		if errors.As(err, &valErrPtr) {
			if valErrPtr.Severity == s {
				return true
			}
			continue
		}
		if errors.As(err, &valErr) {
			if valErr.Severity == s {
				return true
			}
			continue
		}
		if s == SeverityError {
			// Standard errors are treated as SeverityError
			return true
		}
	}
	return false
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
	vErrs := ValidationErrors{Errors: filtered}
	if !vErrs.HasSeverity(SeverityError) {
		return nil
	}
	return vErrs
}
