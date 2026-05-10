package validate

import (
	"errors"
	"testing"
)

func TestErrorf(t *testing.T) {
	err := Errorf("test error %d", 1)
	if err.Error() != "test error 1" {
		t.Errorf("expected 'test error 1', got %q", err.Error())
	}
	if err.Severity != SeverityError {
		t.Errorf("expected SeverityError, got %v", err.Severity)
	}
}

func TestWarningf(t *testing.T) {
	err := Warningf("test warning")
	if err.Error() != "test warning" {
		t.Errorf("expected 'test warning', got %q", err.Error())
	}
	if err.Severity != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", err.Severity)
	}
}

func TestInfof(t *testing.T) {
	err := Infof("test info")
	if err.Error() != "test info" {
		t.Errorf("expected 'test info', got %q", err.Error())
	}
	if err.Severity != SeverityInfo {
		t.Errorf("expected SeverityInfo, got %v", err.Severity)
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &Error{Err: inner, Severity: SeverityError}
	if !errors.Is(err, inner) {
		t.Error("expected err to wrap inner")
	}
	if !errors.Is(errors.Unwrap(err), inner) {
		t.Error("expected Unwrap to return inner")
	}
}
