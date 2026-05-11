package validate_test

import (
	"errors"
	"testing"

	"github.com/oaswrap/spec/internal/validate"

	"github.com/oaswrap/spec/openapi"
)

func checkWarning(err error, expectedMessage string) bool {
	var vErr *validate.Error
	if errors.As(err, &vErr) && vErr.Severity == validate.SeverityWarning {
		return vErr.Error() == expectedMessage
	}
	return false
}

func checkInfo(err error, expectedMessage string) bool {
	var vErr *validate.Error
	if errors.As(err, &vErr) && vErr.Severity == validate.SeverityInfo {
		return vErr.Error() == expectedMessage
	}
	return false
}

//nolint:gocognit // multiple test cases inherently increase cognitive complexity.
func TestValidateWarnings(t *testing.T) {
	t.Run("DocumentRecommendations", func(t *testing.T) {
		doc := &openapi.Document{
			Info: openapi.Info{
				Title:   "Test",
				Version: "1.0",
			},
		}
		errs := validate.ValidateDocument(doc, openapi.Version310)

		hasContactInfo := false
		hasLicenseInfo := false
		hasServersInfo := false
		for _, err := range errs {
			if checkInfo(err, "info.contact is recommended") {
				hasContactInfo = true
			}
			if checkInfo(err, "info.license is recommended") {
				hasLicenseInfo = true
			}
			if checkInfo(err, "servers array is empty") {
				hasServersInfo = true
			}
		}
		if !hasContactInfo {
			t.Error("expected contact recommendation info")
		}
		if !hasLicenseInfo {
			t.Error("expected license recommendation info")
		}
		if !hasServersInfo {
			t.Error("expected servers recommendation info")
		}
	})

	t.Run("OperationRecommendations", func(t *testing.T) {
		op := &openapi.Operation{
			Responses: map[string]*openapi.Response{
				"200": {Description: "OK"},
			},
		}
		errs := validate.ValidateOperation("test", op, openapi.Version310, map[string]string{}, nil, nil)

		hasSummaryWarning := false
		hasDescriptionWarning := false
		hasTagsInfo := false
		hasOpIDWarning := false
		for _, err := range errs {
			if checkWarning(err, "test.summary is recommended") {
				hasSummaryWarning = true
			}
			if checkWarning(err, "test.description is recommended") {
				hasDescriptionWarning = true
			}
			if checkWarning(err, "test.operationId is recommended") {
				hasOpIDWarning = true
			}
			if checkInfo(err, "test.tags are missing") {
				hasTagsInfo = true
			}
		}
		if !hasSummaryWarning {
			t.Error("expected summary recommendation warning")
		}
		if !hasDescriptionWarning {
			t.Error("expected description recommendation warning")
		}
		if !hasTagsInfo {
			t.Error("expected tags recommendation info")
		}
		if !hasOpIDWarning {
			t.Error("expected operationId recommendation warning")
		}
	})

	t.Run("DeprecationWarnings", func(t *testing.T) {
		op := &openapi.Operation{
			Deprecated: true,
			Responses: map[string]*openapi.Response{
				"200": {Description: "OK"},
			},
			Parameters: []*openapi.Parameter{
				{
					Name:       "old-param",
					In:         "query",
					Deprecated: true,
					Schema:     &openapi.Schema{Type: "string"},
				},
			},
		}
		errs := validate.ValidateOperation("test", op, openapi.Version310, nil, nil, nil)

		hasOpDeprecation := false
		hasParamDeprecation := false
		for _, err := range errs {
			if checkWarning(err, "test is deprecated") {
				hasOpDeprecation = true
			}
			if checkWarning(err, "test.parameters[0] is deprecated") {
				hasParamDeprecation = true
			}
		}
		if !hasOpDeprecation {
			t.Error("expected operation deprecation warning")
		}
		if !hasParamDeprecation {
			t.Error("expected parameter deprecation warning")
		}
	})

	t.Run("SchemaDeprecationWarning", func(t *testing.T) {
		schema := &openapi.Schema{
			Type:       "string",
			Deprecated: true,
		}
		errs := validate.ValidateSchema("test-schema", schema, openapi.Version310, map[*openapi.Schema]bool{})

		hasSchemaDeprecation := false
		for _, err := range errs {
			if checkWarning(err, "test-schema is deprecated") {
				hasSchemaDeprecation = true
			}
		}
		if !hasSchemaDeprecation {
			t.Error("expected schema deprecation warning")
		}
	})
}
