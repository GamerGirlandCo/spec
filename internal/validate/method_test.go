package validate_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
)

func TestAllowsOperationMethod(t *testing.T) {
	tests := []struct {
		version string
		method  string
		want    bool
	}{
		{openapi.Version304, http.MethodGet, true},
		{openapi.Version304, http.MethodPost, true},
		{openapi.Version312, http.MethodGet, true},
		{openapi.Version320, http.MethodGet, true},
		{openapi.Version304, http.MethodConnect, false},
		{openapi.Version312, http.MethodConnect, false},
		{openapi.Version320, http.MethodConnect, true},
		{openapi.Version304, "PURGE", true},
		{openapi.Version320, "PURGE", true},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.method, func(t *testing.T) {
			assert.Equal(t, tt.want, validate.AllowsOperationMethod(tt.version, tt.method))
		})
	}
}
