package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/openapi"
)

func TestSetOperation(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		version string
		wantErr bool
	}{
		{"GET", http.MethodGet, openapi.Version304, false},
		{"POST", http.MethodPost, openapi.Version304, false},
		{"PUT", http.MethodPut, openapi.Version304, false},
		{"DELETE", http.MethodDelete, openapi.Version304, false},
		{"PATCH", http.MethodPatch, openapi.Version304, false},
		{"HEAD", http.MethodHead, openapi.Version304, false},
		{"OPTIONS", http.MethodOptions, openapi.Version304, false},
		{"TRACE", http.MethodTrace, openapi.Version304, false},
		{"QUERY 3.2.0", "QUERY", openapi.Version320, false},
		{
			"QUERY 3.1.2",
			"QUERY",
			openapi.Version312,
			false,
		}, // QUERY is handled as a standard method in SetOperation but it's not in net/http.
		{"Custom method 3.2.0", "CUSTOM", openapi.Version320, false},
		{"Custom method 3.0.4", "CUSTOM", openapi.Version304, true},
		{"Duplicate GET", http.MethodGet, openapi.Version304, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &openapi.PathItem{}
			op := &openapi.Operation{}
			if tt.name == "Duplicate GET" {
				item.Get = &openapi.Operation{}
			}

			err := SetOperation(item, tt.method, op, tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentType(t *testing.T) {
	assert.Equal(t, "application/json", ContentType(nil))
	assert.Equal(t, "application/json", ContentType(&openapi.ContentUnit{}))
	assert.Equal(t, "application/xml", ContentType(&openapi.ContentUnit{ContentType: "application/xml"}))
}

func TestResponseDescription(t *testing.T) {
	assert.Equal(t, "OK", ResponseDescription(&openapi.ContentUnit{HTTPStatus: http.StatusOK}))
	assert.Equal(t, "Custom", ResponseDescription(&openapi.ContentUnit{Description: "Custom"}))
	assert.Equal(t, "Default response", ResponseDescription(&openapi.ContentUnit{IsDefault: true}))
	assert.Equal(t, "HTTP 999 response", ResponseDescription(&openapi.ContentUnit{HTTPStatus: 999}))
}

func TestOneOf(t *testing.T) {
	v := OneOf(1, "two")
	ov, ok := v.(oneOfValue)
	assert.True(t, ok)
	assert.Equal(t, []any{1, "two"}, ov.GetValues())
}

func TestMergeResponses(t *testing.T) {
	responses := []*openapi.ContentUnit{
		{HTTPStatus: 200, ContentType: "application/json", Structure: 1},
		{HTTPStatus: 200, ContentType: "application/json", Structure: "two"},
		{HTTPStatus: 400, ContentType: "application/json", Structure: "err"},
	}

	merged := MergeResponses(responses)
	assert.Len(t, merged, 2)
	assert.Equal(t, 200, merged[0].HTTPStatus)
	assert.Equal(t, 400, merged[1].HTTPStatus)

	ov, ok := merged[0].Structure.(oneOfValue)
	assert.True(t, ok)
	assert.Equal(t, []any{1, "two"}, ov.GetValues())
}
