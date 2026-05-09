package option

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/openapi"
)

func TestWithOpenAPIConfig(t *testing.T) {
	cfg := WithOpenAPIConfig(
		WithOpenAPIVersion(openapi.Version312),
		WithSelf("https://api.test.com/openapi.yaml"),
		WithJSONSchemaDialect("https://json-schema.org/draft/2020-12/schema"),
		WithTitle("Test API"),
		WithInfoSummary("Short API summary"),
		WithVersion("2.0.0"),
		WithDescription("Test description"),
		WithTermsOfService("https://example.com/terms"),
		WithContact(openapi.Contact{Name: "Test Contact"}),
		WithLicense(openapi.License{Name: "MIT"}),
		WithTags(openapi.Tag{Name: "test"}),
		WithServer("https://api.test.com", ServerDescription("Test server")),
		WithExternalDocs("https://docs.test.com", "Test docs"),
		WithSecurity("apiKey", SecurityAPIKey("X-API-Key", "header")),
		WithGlobalSecurity("apiKey", "read", "write"),
		WithStripTrailingSlash(true),
	)

	assert.Equal(t, openapi.Version312, cfg.OpenAPIVersion)
	assert.Equal(t, "https://api.test.com/openapi.yaml", cfg.Self)
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", cfg.JSONSchemaDialect)
	assert.Equal(t, "Test API", cfg.Title)
	assert.Equal(t, "Short API summary", cfg.InfoSummary)
	assert.Equal(t, "2.0.0", cfg.Version)
	assert.Equal(t, "Test description", *cfg.Description)
	assert.Equal(t, "https://example.com/terms", *cfg.TermsOfService)
	assert.Equal(t, "Test Contact", cfg.Contact.Name)
	assert.Equal(t, "MIT", cfg.License.Name)

	if assert.Len(t, cfg.Tags, 1) {
		assert.Equal(t, "test", cfg.Tags[0].Name)
	}
	if assert.Len(t, cfg.Servers, 1) {
		assert.Equal(t, "https://api.test.com", cfg.Servers[0].URL)
		assert.Equal(t, "Test server", *cfg.Servers[0].Description)
	}
	assert.Equal(t, "https://docs.test.com", cfg.ExternalDocs.URL)
	assert.Equal(t, "Test docs", cfg.ExternalDocs.Description)
	assert.Equal(t, "apiKey", cfg.SecuritySchemes["apiKey"].Type)
	if assert.Len(t, cfg.Security, 1) {
		assert.Equal(t, "read", cfg.Security[0]["apiKey"][0])
	}
	assert.True(t, cfg.StripTrailingSlash)

	t.Run("ReflectorConfig", func(t *testing.T) {
		c := WithOpenAPIConfig(WithReflectorConfig(InlineRefs(true)))
		assert.True(t, c.ReflectorConfig.InlineRefs)
	})

	t.Run("PathParser", func(t *testing.T) {
		c := WithOpenAPIConfig(WithPathParser(nil))
		assert.Nil(t, c.PathParser)
	})

	t.Run("GlobalSecurityWithoutScopes", func(t *testing.T) {
		c := WithOpenAPIConfig(WithGlobalSecurity("apiKey"))
		if assert.Len(t, c.Security, 1) {
			assert.Empty(t, c.Security[0]["apiKey"])
		}
	})

	t.Run("WithDocument", func(t *testing.T) {
		called := false
		c := WithOpenAPIConfig(WithDocument(func(_ *openapi.Document) { called = true }))
		if assert.Len(t, c.DocumentCustomizers, 1) {
			c.DocumentCustomizers[0](nil)
			assert.True(t, called)
		}
	})
}

func TestTagOptions(t *testing.T) {
	cfg := WithOpenAPIConfig(
		WithTag("payments",
			TagSummary("Payments"),
			TagDescription("Payment operations"),
			TagExternalDocs("https://docs.test/payments", "Payment docs"),
			TagParent("commerce"),
			TagKind("nav"),
		),
	)

	if assert.Len(t, cfg.Tags, 1) {
		tag := cfg.Tags[0]
		assert.Equal(t, "payments", tag.Name)
		assert.Equal(t, "Payments", tag.Summary)
		assert.Equal(t, "Payment operations", tag.Description)
		assert.Equal(t, "https://docs.test/payments", tag.ExternalDocs.URL)
		assert.Equal(t, "Payment docs", tag.ExternalDocs.Description)
		assert.Equal(t, "commerce", tag.Parent)
		assert.Equal(t, "nav", tag.Kind)
	}
}

func TestComponentOptions(t *testing.T) {
	doc := &openapi.Document{}
	cfg := WithOpenAPIConfig(
		WithComponentSchema("User", &openapi.Schema{Type: "object"}),
		WithComponentResponse("OK", &openapi.Response{Description: "OK"}),
		WithComponentParameter("TraceID", &openapi.Parameter{Name: "X-Trace-ID", In: "header"}),
		WithComponentExample("UserExample", &openapi.Example{Value: map[string]any{"id": "123"}}),
		WithComponentRequestBody("CreateUser", &openapi.RequestBody{Content: map[string]openapi.MediaType{}}),
		WithComponentHeader("RateLimit", &openapi.Header{Description: "Rate limit"}),
		WithComponentSecurityScheme("ApiKey", &openapi.SecurityScheme{Type: "apiKey", Name: "X-API-Key", In: "header"}),
		WithComponentLink("UserLink", &openapi.Link{OperationID: "getUser"}),
		WithComponentCallback("Event", &openapi.Callback{}),
		WithComponentPathItem("Users", &openapi.PathItem{}),
		WithComponentMediaType("JSON", &openapi.MediaType{Schema: &openapi.Schema{Type: "object"}}),
	)

	for _, customize := range cfg.DocumentCustomizers {
		customize(doc)
	}

	assert.Equal(t, "object", doc.Components.Schemas["User"].Type)
	assert.Equal(t, "OK", doc.Components.Responses["OK"].Description)
	assert.Equal(t, "X-Trace-ID", doc.Components.Parameters["TraceID"].Name)
	assert.NotNil(t, doc.Components.Examples["UserExample"])
	assert.NotNil(t, doc.Components.RequestBodies["CreateUser"])
	assert.Equal(t, "Rate limit", doc.Components.Headers["RateLimit"].Description)
	assert.Equal(t, "apiKey", doc.Components.SecuritySchemes["ApiKey"].Type)
	assert.Equal(t, "getUser", doc.Components.Links["UserLink"].OperationID)
	assert.NotNil(t, doc.Components.Callbacks["Event"])
	assert.NotNil(t, doc.Components.PathItems["Users"])
	assert.NotNil(t, doc.Components.MediaTypes["JSON"])
}

func TestOperationOptions(t *testing.T) {
	cfg := &OperationConfig{}
	opts := []OperationOption{
		Hidden(true),
		OperationID("testOp"),
		Summary("Test Summary"),
		Description("Test Description"),
		ExternalDocs("https://docs.test/operation", "Operation docs"),
		Deprecated(true),
		Tags("tag1", "tag2"),
		Security("auth", "scope1"),
		Request(struct{}{}, ContentType("application/json")),
		Response(200, struct{}{}, ContentDescription("OK")),
		CustomizeOperation(func(op *openapi.Operation) { op.Summary = "Custom" }),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.True(t, cfg.Hide)
	assert.Equal(t, "testOp", cfg.OperationID)
	assert.Equal(t, "Test Summary", cfg.Summary)
	assert.Equal(t, "Test Description", cfg.Description)
	assert.Equal(t, "https://docs.test/operation", cfg.ExternalDocs.URL)
	assert.Equal(t, "Operation docs", cfg.ExternalDocs.Description)
	assert.True(t, cfg.Deprecated)
	assert.Equal(t, []string{"tag1", "tag2"}, cfg.Tags)
	if assert.Len(t, cfg.Security, 1) {
		assert.Equal(t, "auth", cfg.Security[0].Name)
		assert.Equal(t, "scope1", cfg.Security[0].Scopes[0])
	}
	if assert.Len(t, cfg.Requests, 1) {
		assert.Equal(t, "application/json", cfg.Requests[0].ContentType)
	}
	if assert.Len(t, cfg.Responses, 1) {
		assert.Equal(t, "OK", cfg.Responses[0].Description)
	}
	assert.Len(t, cfg.Customizers, 1)
}

func TestGroupOptions(t *testing.T) {
	cfg := &GroupConfig{}
	opts := []GroupOption{
		GroupHidden(true),
		GroupDeprecated(true),
		GroupTags("g1", "g2"),
		GroupSecurity("gauth", "gscope"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.True(t, cfg.Hide)
	assert.True(t, cfg.Deprecated)
	assert.Equal(t, []string{"g1", "g2"}, cfg.Tags)
	if assert.Len(t, cfg.Security, 1) {
		assert.Equal(t, "gauth", cfg.Security[0].Name)
		assert.Equal(t, "gscope", cfg.Security[0].Scopes[0])
	}
}

func TestReflectorOptions(t *testing.T) {
	cfg := &openapi.ReflectorConfig{}
	opts := []ReflectorOption{
		InlineRefs(true),
		StripDefNamePrefix("Pre"),
		InterceptDefName(func(_ reflect.Type, _ string) string { return "Intercepted" }),
		TypeMapping(1, "one"),
		ParameterTagMapping(openapi.ParameterInQuery, "q"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.True(t, cfg.InlineRefs)
	if assert.Len(t, cfg.StripDefNamePrefix, 1) {
		assert.Equal(t, "Pre", cfg.StripDefNamePrefix[0])
	}
	assert.Equal(t, "Intercepted", cfg.InterceptDefName(nil, ""))
	if assert.Len(t, cfg.TypeMappings, 1) {
		assert.Equal(t, 1, cfg.TypeMappings[0].Src)
		assert.Equal(t, "one", cfg.TypeMappings[0].Dst)
	}
	assert.Equal(t, "q", cfg.ParameterTagMapping[openapi.ParameterInQuery])
}

func TestSecurityOptions(t *testing.T) {
	t.Run("APIKey", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityAPIKey("api_key", "query")(cfg)
		assert.Equal(t, "apiKey", cfg.scheme.Type)
		assert.Equal(t, "api_key", cfg.scheme.Name)
		assert.Equal(t, openapi.SecuritySchemeAPIKeyInQuery, cfg.scheme.In)
	})

	t.Run("HTTPBearer", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityHTTPBearer("bearer", "JWT")(cfg)
		assert.Equal(t, "http", cfg.scheme.Type)
		assert.Equal(t, "bearer", cfg.scheme.Scheme)
		assert.Equal(t, "JWT", *cfg.scheme.BearerFormat)
	})

	t.Run("OAuth2", func(t *testing.T) {
		cfg := &securityConfig{}
		flows := openapi.OAuthFlows{Implicit: &openapi.OAuthFlow{AuthorizationURL: "https://auth.com"}}
		SecurityOAuth2(flows)(cfg)
		assert.Equal(t, "oauth2", cfg.scheme.Type)
		assert.Equal(t, "https://auth.com", cfg.scheme.Flows.Implicit.AuthorizationURL)
	})

	t.Run("OAuth2FlowHelpers", func(t *testing.T) {
		scopes := map[string]string{"read": "Read access"}

		cfg := &securityConfig{}
		SecurityOAuth2Implicit("https://auth.com/authorize", scopes, OAuthRefreshURL("https://auth.com/refresh"))(cfg)
		assert.Equal(t, "https://auth.com/authorize", cfg.scheme.Flows.Implicit.AuthorizationURL)
		assert.Equal(t, "https://auth.com/refresh", *cfg.scheme.Flows.Implicit.RefreshURL)

		cfg = &securityConfig{}
		SecurityOAuth2Password("https://auth.com/token", nil)(cfg)
		assert.Equal(t, "https://auth.com/token", cfg.scheme.Flows.Password.TokenURL)
		assert.Empty(t, cfg.scheme.Flows.Password.Scopes)

		cfg = &securityConfig{}
		SecurityOAuth2ClientCredentials("https://auth.com/token", scopes)(cfg)
		assert.Equal(t, "https://auth.com/token", cfg.scheme.Flows.ClientCredentials.TokenURL)

		cfg = &securityConfig{}
		SecurityOAuth2AuthorizationCode("https://auth.com/authorize", "https://auth.com/token", scopes)(cfg)
		assert.Equal(t, "https://auth.com/authorize", cfg.scheme.Flows.AuthorizationCode.AuthorizationURL)
		assert.Equal(t, "https://auth.com/token", cfg.scheme.Flows.AuthorizationCode.TokenURL)

		cfg = &securityConfig{}
		SecurityOAuth2DeviceAuthorization("https://auth.com/device", "https://auth.com/token", scopes)(cfg)
		assert.Equal(t, "https://auth.com/device", cfg.scheme.Flows.DeviceAuthorization.DeviceAuthorizationURL)
		assert.Equal(t, "https://auth.com/token", cfg.scheme.Flows.DeviceAuthorization.TokenURL)
	})

	t.Run("MutualTLS", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityMutualTLS()(cfg)
		assert.Equal(t, "mutualTLS", cfg.scheme.Type)
	})

	t.Run("OpenIDConnect", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityOpenIDConnect("https://oidc.com")(cfg)
		assert.Equal(t, "openIdConnect", cfg.scheme.Type)
		assert.Equal(t, "https://oidc.com", cfg.scheme.OpenIDConnectURL)
	})

	t.Run("Description", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityDescription("desc")(cfg)
		SecurityAPIKey("k", "h")(cfg)
		assert.Equal(t, "desc", *cfg.scheme.Description)
	})

	t.Run("OAuth2MetadataURL", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityOAuth2MetadataURL("https://issuer.test/.well-known/oauth-authorization-server")(cfg)
		SecurityHTTPBearer("bearer")(cfg)
		assert.Equal(t, "https://issuer.test/.well-known/oauth-authorization-server", cfg.scheme.OAuth2MetadataURL)
	})

	t.Run("Deprecated", func(t *testing.T) {
		cfg := &securityConfig{}
		SecurityDeprecated()(cfg)
		assert.True(t, cfg.scheme.Deprecated)
		SecurityDeprecated(false)(cfg)
		assert.False(t, cfg.scheme.Deprecated)
	})
}

func TestContentOptions(t *testing.T) {
	cu := &openapi.ContentUnit{}
	opts := []ContentOption{
		ContentType("text/plain"),
		ContentDescription("Text"),
		ContentDefault(true),
		ContentEncoding("prop", "enc"),
		ContentExample(map[string]any{"id": "123"}),
		ContentNamedExample("named", map[string]any{"id": "456"}, ExampleSummary("Named")),
		ContentRequired(true),
		ContentFormat("binary"),
	}

	for _, opt := range opts {
		opt(cu)
	}

	assert.Equal(t, "text/plain", cu.ContentType)
	assert.Equal(t, "Text", cu.Description)
	assert.True(t, cu.IsDefault)
	assert.Equal(t, "enc", cu.Encoding["prop"])
	assert.Equal(t, map[string]any{"id": "123"}, cu.Example)
	assert.Equal(t, "Named", cu.Examples["named"].Summary)
	assert.Equal(t, map[string]any{"id": "456"}, cu.Examples["named"].Value)
	assert.True(t, cu.Required)
	assert.Equal(t, "binary", cu.Format)

	t.Run("ContentExamples", func(t *testing.T) {
		content := &openapi.ContentUnit{}
		examples := map[string]*openapi.Example{"external": {}}
		ContentExamples(examples)(content)
		assert.Same(t, examples["external"], content.Examples["external"])
	})

	t.Run("ExampleExternalValue", func(t *testing.T) {
		content := &openapi.ContentUnit{}
		ContentNamedExample("external", "ignored", ExampleExternalValue("https://example.com/user.json"))(content)
		assert.Nil(t, content.Examples["external"].Value)
		assert.Equal(t, "https://example.com/user.json", content.Examples["external"].ExternalValue)
	})

	t.Run("ExampleSerializedValue", func(t *testing.T) {
		content := &openapi.ContentUnit{}
		ContentNamedExample("serialized", "ignored", ExampleSerializedValue(`{"id":"123"}`))(content)
		assert.Nil(t, content.Examples["serialized"].Value)
		assert.JSONEq(t, `{"id":"123"}`, content.Examples["serialized"].SerializedValue)
	})
}

func TestServerOptions(t *testing.T) {
	s := &openapi.Server{}
	opts := []ServerOption{
		ServerDescription("Desc"),
		ServerVariables(map[string]openapi.ServerVariable{"v": {Default: "d"}}),
	}

	for _, opt := range opts {
		opt(s)
	}

	assert.Equal(t, "Desc", *s.Description)
	assert.Equal(t, "d", s.Variables["v"].Default)
}

func TestOptional(t *testing.T) {
	assert.Equal(t, 1, optional(1))
	assert.Equal(t, 2, optional(1, 2))
	assert.Equal(t, "b", optional("a", "b"))
}
