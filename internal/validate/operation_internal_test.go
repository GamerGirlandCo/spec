package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/openapi"
)

func TestValidateLink_Direct(t *testing.T) {
	errs := ValidateLink("ctx", nil, openapi.Version304)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "ctx is required")

	errs = ValidateLink("ctx", &openapi.Link{Summary: "x"}, openapi.Version304)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "summary is only allowed with $ref")

	errs = ValidateLink("ctx", &openapi.Link{OperationRef: "/a", OperationID: "id"}, openapi.Version304)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "operationRef and operationId are mutually exclusive")

	errs = ValidateLink("ctx", &openapi.Link{}, openapi.Version304)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "must define operationRef or operationId")
}

func TestValidateExample_Direct(t *testing.T) {
	errs := ValidateExample("ctx", nil, openapi.Version304)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "ctx is required")

	errs = ValidateExample("ctx", &openapi.Example{
		SerializedValue: "x",
		Value:           "y",
	}, openapi.Version320)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "serializedValue is mutually exclusive with value and externalValue")
}

func TestSecuritySchemeOAuth2MetadataURL(t *testing.T) {
	value, ok := securitySchemeOAuth2MetadataURL(&openapi.SecurityScheme{
		OAuth2MetadataURL: "https://auth.example/.well-known/oauth-authorization-server",
	})
	assert.True(t, ok)
	assert.Equal(t, "https://auth.example/.well-known/oauth-authorization-server", value)

	value, ok = securitySchemeOAuth2MetadataURL(&openapi.SecurityScheme{
		Extra: map[string]any{"oauth2MetadataUrl": "https://auth.example/metadata"},
	})
	assert.True(t, ok)
	assert.Equal(t, "https://auth.example/metadata", value)

	value, ok = securitySchemeOAuth2MetadataURL(&openapi.SecurityScheme{
		Extra: map[string]any{"oauth2MetadataUrl": 1},
	})
	assert.True(t, ok)
	assert.Empty(t, value)

	value, ok = securitySchemeOAuth2MetadataURL(&openapi.SecurityScheme{})
	assert.False(t, ok)
	assert.Empty(t, value)
}
