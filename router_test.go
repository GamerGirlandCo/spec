package spec_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type LoginRequest struct {
	Username string `json:"username" required:"true" minLength:"3"`
	Password string `json:"password" required:"true" writeOnly:"true"`
}

type LoginResponse struct {
	Token string `json:"token" required:"true"`
}

type GetUserRequest struct {
	ID string `path:"id" required:"true" description:"User identifier"`
}

type User struct {
	ID   string `json:"id" required:"true"`
	Name string `json:"name"`
}

func TestRouter_GenerateSchema_DefaultVersion(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Users API"),
		option.WithVersion("1.2.3"),
		option.WithServer("https://api.example.com"),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("bearer")),
	)

	v1 := r.Group("/api/v1", option.GroupTags("v1"))
	v1.Post("/login",
		option.Summary("Login"),
		option.Request(new(LoginRequest)),
		option.Response(200, new(LoginResponse)),
	)
	v1.Get("/users/{id}",
		option.Summary("Get user"),
		option.Request(new(GetUserRequest)),
		option.Response(200, new(User)),
	)

	raw, err := r.GenerateSchema("json")
	require.NoError(t, err)

	var doc openapi.Document
	err = json.Unmarshal(raw, &doc)
	require.NoError(t, err, "unmarshal generated JSON:\n%s", raw)

	assert.Equal(t, openapi.Version312, doc.OpenAPI)
	assert.Equal(t, "Users API", doc.Info.Title)
	assert.Equal(t, "1.2.3", doc.Info.Version)
	assert.NotNil(t, doc.Paths["/api/v1/login"].Post)

	get := doc.Paths["/api/v1/users/{id}"].Get
	require.NotNil(t, get)
	if assert.Len(t, get.Parameters, 1) {
		assert.Equal(t, "id", get.Parameters[0].Name)
		assert.True(t, get.Parameters[0].Required)
	}

	assert.NotContains(t, doc.Components.Schemas, "user", "component names should use exported Go type names")
	assert.NotNil(t, doc.Components.Schemas["SpecTestUser"])
	assert.NotNil(t, doc.Components.Schemas["SpecTestLoginRequest"])
	assert.Equal(t, "http", doc.Components.SecuritySchemes["bearerAuth"].Type)
}

func TestRouter_OpenAPI320_Features(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Search API"),
		option.WithVersion("1.0.0"),
	)
	r.Query("/search", option.Response(200, new([]User)))
	r.Add("PURGE", "/cache", option.Response(204, nil))
	r.Add(http.MethodConnect, "/tunnel", option.Response(204, nil))

	doc := r.Document()
	require.NoError(t, r.Validate())
	assert.NotNil(t, doc.Paths["/search"].Query)
	assert.NotNil(t, doc.Paths["/cache"].AdditionalOperations["PURGE"])
	assert.NotNil(t, doc.Paths["/tunnel"].AdditionalOperations[http.MethodConnect])
}

func TestRouter_Webhooks_OpenAPI312(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Webhook API"),
		option.WithVersion("1.0.0"),
	)
	r.Webhook("user.created", option.Response(202, nil))
	r.AddWebhook("POST", "cache.invalidate", option.Response(204, nil))

	doc := r.Document()
	require.NoError(t, r.Validate())
	require.NotNil(t, doc.Webhooks["user.created"].Post)
	require.NotNil(t, doc.Webhooks["cache.invalidate"].Post)
	assert.NotNil(t, doc.Webhooks["cache.invalidate"].Post.Responses["204"])
}

func TestRouter_Webhooks_OpenAPI304_Rejection(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version304),
		option.WithTitle("Webhook API"),
		option.WithVersion("1.0.0"),
	)
	r.Webhook("user.created", option.Response(202, nil))

	err := r.Validate()
	require.Error(t, err)
	assert.ErrorContains(t, err, "webhooks require OpenAPI 3.1.x or 3.2.0")
}

func TestRouter_MergeResponses(t *testing.T) {
	type conflictA struct {
		Code string `json:"code"`
	}
	type conflictB struct {
		Message string `json:"message"`
	}

	r := spec.NewRouter(option.WithTitle("Conflicts"), option.WithVersion("1.0.0"))
	r.Get("/items",
		option.Response(409, new(conflictA)),
		option.Response(409, new(conflictB)),
	)

	doc := r.Document()
	require.NoError(t, r.Validate())
	schema := doc.Paths["/items"].Get.Responses["409"].Content["application/json"].Schema
	assert.Len(t, schema.OneOf, 2)
}

func TestMissingPathParamIsAutoAdded(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Invalid"), option.WithVersion("1.0.0"))
	r.Get("/users/{id}", option.Response(200, new(User)))

	require.NoError(t, r.Validate())
	doc := r.Document()
	params := doc.Paths["/users/{id}"].Get.Parameters
	require.Len(t, params, 1)
	assert.Equal(t, "id", params[0].Name)
	assert.Equal(t, "path", params[0].In)
	assert.True(t, params[0].Required)
}

func TestUnsupportedVersionFailsValidation(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion("3.3.0"),
		option.WithTitle("Invalid"),
		option.WithVersion("1.0.0"),
	)
	r.Get("/ping", option.Response(204, nil))

	err := r.Validate()
	assert.ErrorContains(t, err, "unsupported OpenAPI version")
}

func TestSupportedPatchVersions(t *testing.T) {
	versions := []string{
		openapi.Version300,
		openapi.Version301,
		openapi.Version302,
		openapi.Version303,
		openapi.Version304,
		openapi.Version310,
		openapi.Version311,
		openapi.Version312,
		openapi.Version320,
	}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			r := spec.NewRouter(
				option.WithOpenAPIVersion(version),
				option.WithTitle("Versioned"),
				option.WithVersion("1.0.0"),
			)
			r.Get("/ping", option.Response(200, nil))
			require.NoError(t, r.Validate())
			assert.Equal(t, version, r.Document().OpenAPI)
		})
	}
}

func TestSchemaSkipsJSONIgnoredFields(t *testing.T) {
	type Payload struct {
		Public string `json:"public"`
		Secret string `json:"-"`
	}

	r := spec.NewRouter(option.WithTitle("Ignored"), option.WithVersion("1.0.0"))
	r.Post("/payload", option.Request(new(Payload)), option.Response(204, nil))

	doc := r.Document()
	require.NoError(t, r.Validate())
	payload := doc.Components.Schemas["SpecTestPayload"]
	assert.NotNil(t, payload.Properties["public"])
	assert.NotContains(t, payload.Properties, "Secret")
	assert.NotContains(t, payload.Properties, "secret")
}

func TestEmptySecurityScopesRenderAsArray(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Secure API"),
		option.WithVersion("1.0.0"),
		option.WithSecurity("apiKey", option.SecurityAPIKey("X-API-Key", "header")),
	)
	r.Get("/me", option.Security("apiKey"), option.Response(204, nil))

	raw, err := r.GenerateSchema("yaml")
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "apiKey: null")
	assert.Contains(t, string(raw), "apiKey: []")
}

func TestAllowMutationAfterBuild(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Built"), option.WithVersion("1.0.0"))
	r.Get("/before", option.Response(200, nil))

	require.NoError(t, r.Validate())
	require.NotNil(t, r.Document().Paths["/before"])

	// Add more routes after first build/validate
	r.Get("/after", option.Response(200, nil))
	route := r.NewRoute()
	route.Method("POST").Path("/after-route").With(option.Response(201, nil))

	// Validate again, should succeed and include new routes
	require.NoError(t, r.Validate())
	doc := r.Document()
	require.NotNil(t, doc.Paths["/before"])
	assert.NotNil(t, doc.Paths["/after"])
	assert.NotNil(t, doc.Paths["/after-route"])
}

func TestRouterAdditionalMethods(t *testing.T) {
	r := spec.NewRouter()
	r.Patch("/patch", option.Response(204, nil))
	r.Options("/options", option.Response(204, nil))
	r.Head("/head", option.Response(204, nil))
	r.Trace("/trace", option.Response(204, nil))

	doc := r.Document()
	methods := []string{"patch", "options", "head", "trace"}
	for _, m := range methods {
		path := "/" + m
		assert.NotNil(t, doc.Paths[path], "expected path %s", path)
	}
}

func TestRouteFluentAPI(t *testing.T) {
	r := spec.NewRouter()
	r.NewRoute().
		Method(http.MethodPost).
		Path("/fluent").
		With(option.OperationID("fluentOp"), option.Response(201, nil))

	doc := r.Document()
	require.NotNil(t, doc.Paths["/fluent"])
	require.NotNil(t, doc.Paths["/fluent"].Post)
	assert.Equal(t, "fluentOp", doc.Paths["/fluent"].Post.OperationID)
}

func TestRouterGroupWith(t *testing.T) {
	r := spec.NewRouter()
	g := r.Group("/api")
	g.With(option.GroupTags("api")).Get("/ping", option.Response(204, nil))

	doc := r.Document()
	op := doc.Paths["/api/ping"].Get
	if assert.Len(t, op.Tags, 1) {
		assert.Equal(t, "api", op.Tags[0])
	}
}

func TestRouterGenerateSchemaErrors(t *testing.T) {
	r := spec.NewRouter()
	r.Get("/ping", option.Security("missing"), option.Response(204, nil))
	_, err := r.GenerateSchema("yaml")
	require.Error(t, err)

	r = spec.NewRouter()
	r.Get("/ping", option.Response(204, nil))
	_, err = r.GenerateSchema("invalid")
	require.Error(t, err)
}

func TestRouterAutoAddsPathParameterWithoutRequestStruct(t *testing.T) {
	r := spec.NewRouter()
	r.Get("/users/{id}", option.Response(200, nil))

	require.NoError(t, r.Validate())
	doc := r.Document()
	params := doc.Paths["/users/{id}"].Get.Parameters
	require.Len(t, params, 1)
	assert.Equal(t, "id", params[0].Name)
	assert.Equal(t, "path", params[0].In)
	assert.True(t, params[0].Required)
	require.NotNil(t, params[0].Schema)
	assert.Equal(t, "string", params[0].Schema.Type)
}

func TestRouterMarshal(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Test"))
	r.Get("/ping", option.Response(204, nil))

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var out map[string]any
	err = json.Unmarshal(data, &out)
	require.NoError(t, err)

	info := out["info"].(map[string]any)
	assert.Equal(t, "Test", info["title"])
}

func TestRouterUsabilityOptions(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithJSONSchemaDialect("https://json-schema.org/draft/2020-12/schema"),
		option.WithInfoSummary("Short summary"),
		option.WithTag("commerce", option.TagSummary("Commerce"), option.TagKind("nav")),
		option.WithTag("payments", option.TagParent("commerce"), option.TagDescription("Payment operations")),
		option.WithComponentExample("UserExample", &openapi.Example{Value: map[string]any{"id": "123"}}),
		option.WithSecurity("oauth2",
			option.SecurityOAuth2ClientCredentials(
				"https://auth.example.com/token",
				map[string]string{"payments:read": "Read payments"},
			),
			option.SecurityOAuth2MetadataURL("https://auth.example.com/.well-known/oauth-authorization-server"),
		),
	)
	r.Get("/users",
		option.ExternalDocs("https://docs.example.com/users/get", "Get user docs"),
		option.Response(200, new(User),
			option.ContentNamedExample("sample", map[string]any{"id": "123"}, option.ExampleSummary("Sample user")),
		),
	)

	require.NoError(t, r.Validate())
	doc := r.Document()
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", doc.JSONSchemaDialect)
	assert.Equal(t, "Short summary", doc.Info.Summary)
	assert.Equal(t, "commerce", doc.Tags[1].Parent)
	assert.Equal(
		t,
		"https://auth.example.com/token",
		doc.Components.SecuritySchemes["oauth2"].Flows.ClientCredentials.TokenURL,
	)

	op := doc.Paths["/users"].Get
	assert.Equal(t, "https://docs.example.com/users/get", op.ExternalDocs.URL)
	example := op.Responses["200"].Content["application/json"].Examples["sample"]
	assert.Equal(t, "Sample user", example.Summary)
	assert.Equal(t, map[string]any{"id": "123"}, example.Value)
}

func TestRouterWith(t *testing.T) {
	r := spec.NewRouter()
	r.With(option.GroupTags("global")).Get("/ping", option.Response(204, nil))
	doc := r.Document()
	op := doc.Paths["/ping"].Get
	if assert.Len(t, op.Tags, 1) {
		assert.Equal(t, "global", op.Tags[0])
	}
}

func TestRouter_WriteSchemaTo(t *testing.T) {
	tmp := t.TempDir()

	t.Run("SuccessYAML", func(t *testing.T) {
		r := spec.NewRouter()
		r.Get("/ping", option.Response(204, nil))
		err := r.WriteSchemaTo(filepath.Join(tmp, "test.yaml"))
		require.NoError(t, err)
	})

	t.Run("SuccessJSON", func(t *testing.T) {
		r := spec.NewRouter()
		r.Get("/ping", option.Response(204, nil))
		err := r.WriteSchemaTo(filepath.Join(tmp, "test.json"))
		require.NoError(t, err)
	})

	t.Run("InvalidExtension", func(t *testing.T) {
		r := spec.NewRouter()
		err := r.WriteSchemaTo(filepath.Join(tmp, "test.txt"))
		require.Error(t, err)
	})
}
func TestRouter_PathHelpers(t *testing.T) {
	r := spec.NewRouter()
	r.Get("ping", option.Response(204, nil)) // should ensure leading slash
	doc := r.Document()
	assert.Contains(t, doc.Paths, "/ping")
}

func TestRouter_StripTrailingSlash(t *testing.T) {
	r := spec.NewRouter(option.WithStripTrailingSlash(true))
	r.Get("/ping/", option.Response(204, nil))
	doc := r.Document()
	assert.Contains(t, doc.Paths, "/ping")
}

func TestRouterConfigAndRoute(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Configured"), option.WithVersion("2.0.0"))
	cfg := r.Config()
	assert.Equal(t, "Configured", cfg.Title)
	assert.Equal(t, "2.0.0", cfg.Version)

	r.Route("/api", func(router spec.Router) {
		router.Get("/ping", option.Response(204, nil))
	}, option.GroupTags("v1"))

	doc := r.Document()
	require.Contains(t, doc.Paths, "/api/ping")
	if assert.NotNil(t, doc.Paths["/api/ping"].Get) {
		assert.Equal(t, []string{"v1"}, doc.Paths["/api/ping"].Get.Tags)
	}
}

func TestRouter_EscapeHatches(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Official Surface"),
		option.WithVersion("1.0.0"),
		option.WithSecurity("mtls", option.SecurityMutualTLS()),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Extensions = map[string]any{"x-root": "ok"}
			doc.Webhooks = map[string]*openapi.PathItem{
				"user.created": {
					Post: &openapi.Operation{
						Responses: map[string]*openapi.Response{
							"202": {Description: "Accepted"},
						},
					},
				},
			}
			doc.Components.MediaTypes = map[string]*openapi.MediaType{
				"json-seq": {
					ItemSchema: &openapi.Schema{Ref: "#/components/schemas/SpecTestUser"},
					ItemEncoding: &openapi.Encoding{
						ContentType: "application/json",
						Extensions:  map[string]any{"x-encoding": true},
					},
				},
			}
			doc.Components.Parameters = map[string]*openapi.Parameter{
				"TraceID": {
					Name:     "X-Trace-ID",
					In:       "header",
					Required: false,
					Schema:   &openapi.Schema{Type: "string"},
				},
			}
		}),
	)
	r.Get("/users/{id}",
		option.Request(new(GetUserRequest)),
		option.Response(200, new(User)),
		option.CustomizeOperation(func(op *openapi.Operation) {
			op.Extensions = map[string]any{"x-operation": "ok"}
			op.Parameters = append(op.Parameters, &openapi.Parameter{Ref: "#/components/parameters/TraceID"})
		}),
	)

	raw, err := r.GenerateSchema("json")
	require.NoError(t, err)

	var doc map[string]any
	err = json.Unmarshal(raw, &doc)
	require.NoError(t, err, "unmarshal generated JSON:\n%s", raw)

	assert.Equal(t, "ok", doc["x-root"])

	components := doc["components"].(map[string]any)
	security := components["securitySchemes"].(map[string]any)
	mtls := security["mtls"].(map[string]any)
	assert.Equal(t, "mutualTLS", mtls["type"])

	assert.Contains(t, doc["webhooks"].(map[string]any), "user.created")

	mediaTypes := components["mediaTypes"].(map[string]any)
	assert.Contains(t, mediaTypes, "json-seq")

	paths := doc["paths"].(map[string]any)
	get := paths["/users/{id}"].(map[string]any)["get"].(map[string]any)
	assert.Equal(t, "ok", get["x-operation"])
}

func TestRouter_Errors_Unwrap(t *testing.T) {
	r := spec.NewRouter()
	r.Get("/ping", option.Security("missing"), option.Response(204, nil))
	err := r.Validate()
	require.Error(t, err)

	if assert.Implements(t, (*interface{ Unwrap() []error })(nil), err) {
		u := err.(interface{ Unwrap() []error })
		assert.NotEmpty(t, u.Unwrap())
	}
}

func TestRouter_ValidateReport(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle(""),   // Empty title triggers error
		option.WithVersion(""), // Empty version triggers error
	)

	err := r.ValidateReport()
	require.Error(t, err)

	var valErrs spec.ValidationErrors
	require.ErrorAs(t, err, &valErrs)

	// Check if errors are as expected
	foundTitleErr := false
	foundVersionErr := false
	for _, e := range valErrs.Errors {
		if strings.Contains(e.Error(), "info.title is required") {
			foundTitleErr = true
		}
		if strings.Contains(e.Error(), "info.version is required") {
			foundVersionErr = true
		}
	}
	assert.True(t, foundTitleErr, "should have found title error")
	assert.True(t, foundVersionErr, "should have found version error")

	// Test nil case - providing all recommended fields
	r2 := spec.NewRouter(
		option.WithTitle("Valid"),
		option.WithVersion("1.0.0"),
		option.WithContact(openapi.Contact{Name: "Support"}),
		option.WithLicense(openapi.License{Name: "MIT"}),
		option.WithServer("https://example.com"),
	)

	assert.NoError(t, r2.ValidateReport())
}
