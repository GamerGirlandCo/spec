package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestValidate_OpenAPI320_SecurityRequirements(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithSecurity("apiKeyAuth", option.SecurityAPIKey("X-API-Key", "header")),
		option.WithGlobalSecurity("apiKeyAuth", "admin"),
		option.WithGlobalSecurity("https://security.example/schemes/custom", "operator"),
	)
	r.Get("/secure", option.Response(200, nil))

	assert.NoError(t, r.Validate())
}

func TestValidate_OpenAPI320_OAuthDeviceAuth(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithSecurity("device", option.SecurityOAuth2(openapi.OAuthFlows{
			DeviceAuthorization: &openapi.OAuthFlow{
				DeviceAuthorizationURL: "https://auth.example/device",
				TokenURL:               "https://auth.example/token",
				Scopes:                 map[string]string{},
			},
		})),
		option.WithGlobalSecurity("device"),
	)
	r.Get("/device", option.Response(200, nil))

	assert.NoError(t, r.Validate())
}

func TestValidateSecurityScheme_Errors(t *testing.T) {
	t.Run("MissingRequiredFields", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithComponentSecurityScheme("apiKey", &openapi.SecurityScheme{Type: "apiKey"}),
			option.WithComponentSecurityScheme("http", &openapi.SecurityScheme{Type: "http"}),
			option.WithComponentSecurityScheme("oauth2", &openapi.SecurityScheme{Type: "oauth2"}),
			option.WithComponentSecurityScheme("oidc", &openapi.SecurityScheme{Type: "openIdConnect"}),
		)
		err := r.Validate()
		assertValidationContains(t, err,
			"components.securitySchemes.apiKey.name is required for apiKey",
			"components.securitySchemes.apiKey.in must be query, header, or cookie for apiKey",
			"components.securitySchemes.http.scheme is required for http",
			"components.securitySchemes.oauth2.flows is required for oauth2",
			"components.securitySchemes.oidc.openIdConnectUrl is required for openIdConnect",
		)
	})

	t.Run("InvalidType", func(t *testing.T) {
		r := spec.NewRouter(option.WithComponentSecurityScheme("bad", &openapi.SecurityScheme{Type: "invalid"}))
		err := r.Validate()
		assertValidationContains(t, err, "type must be one of apiKey, http, mutualTLS, oauth2, or openIdConnect")
	})

	t.Run("OpenAPI304_OAuth2Metadata", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithOpenAPIVersion(openapi.Version304),
			option.WithComponentSecurityScheme(
				"oa",
				&openapi.SecurityScheme{Type: "oauth2", OAuth2MetadataURL: "https://ex.com"},
			),
		)
		err := r.Validate()
		assertValidationContains(t, err, "oauth2MetadataUrl and deprecated require OpenAPI 3.2.0")
	})
}

func TestValidateOAuthFlows_Errors(t *testing.T) {
	t.Run("MissingFlows", func(t *testing.T) {
		r := spec.NewRouter(option.WithSecurity("oa", option.SecurityOAuth2(openapi.OAuthFlows{})))
		err := r.Validate()
		assertValidationContains(t, err, "must define at least one OAuth flow")
	})

	t.Run("OpenAPI312_DeviceAuth", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithOpenAPIVersion(openapi.Version312),
			option.WithSecurity("oa", option.SecurityOAuth2(openapi.OAuthFlows{
				DeviceAuthorization: &openapi.OAuthFlow{Scopes: map[string]string{}},
			})),
		)
		err := r.Validate()
		assertValidationContains(t, err, "deviceAuthorization requires OpenAPI 3.2.0")
	})

	t.Run("MissingURLs", func(t *testing.T) {
		r := spec.NewRouter(option.WithSecurity("oa", option.SecurityOAuth2(openapi.OAuthFlows{
			Implicit:          &openapi.OAuthFlow{Scopes: map[string]string{}},
			Password:          &openapi.OAuthFlow{Scopes: map[string]string{}},
			ClientCredentials: &openapi.OAuthFlow{Scopes: map[string]string{}},
			AuthorizationCode: &openapi.OAuthFlow{Scopes: map[string]string{}},
		})))
		err := r.Validate()
		assertValidationContains(t, err,
			"implicit.authorizationUrl is required",
			"password.tokenUrl is required",
			"clientCredentials.tokenUrl is required",
			"authorizationCode.authorizationUrl is required",
			"authorizationCode.tokenUrl is required",
		)
	})
}
