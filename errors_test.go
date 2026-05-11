package spec_test

import (
	"errors"
	"testing"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/validate"
)

func TestValidationErrorAlias(t *testing.T) {
	// Simulate an error returned from internal/validate
	inner := errors.New("validation failed")
	err := error(validate.Error{
		Err:      inner,
		Severity: validate.SeverityWarning,
	})

	// Verify we can use the public alias and constants
	var valErr spec.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatal("expected error to be spec.ValidationError")
	}

	if valErr.Severity != spec.SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", valErr.Severity)
	}

	if !errors.Is(valErr, inner) {
		t.Error("expected ValidationError to wrap inner error")
	}
}

func TestValidationErrors_HasSeverity(t *testing.T) {
	vErrs := spec.ValidationErrors{
		Errors: []spec.ValidationError{
			{Err: errors.New("err"), Severity: spec.SeverityError},
			{Err: errors.New("warn"), Severity: spec.SeverityWarning},
		},
	}

	if !vErrs.HasSeverity(spec.SeverityError) {
		t.Error("expected to have SeverityError")
	}
	if !vErrs.HasSeverity(spec.SeverityWarning) {
		t.Error("expected to have SeverityWarning")
	}
	if vErrs.HasSeverity(spec.SeverityInfo) {
		t.Error("expected NOT to have SeverityInfo")
	}

	vWarnsOnly := spec.ValidationErrors{
		Errors: []spec.ValidationError{
			{Err: errors.New("warn"), Severity: spec.SeverityWarning},
		},
	}
	if vWarnsOnly.HasSeverity(spec.SeverityError) {
		t.Error("expected NOT to have SeverityError in warnings-only collection")
	}
}
