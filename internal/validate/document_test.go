package validate_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestValidate_Document_OpenAPI304_RejectsNewerFields(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version304),
		option.WithTitle("Invalid 3.0"),
		option.WithVersion("1.0.0"),
		option.WithInfoSummary("summary"),
		option.WithJSONSchemaDialect("https://spec.openapis.org/oas/3.1/dialect/base"),
		option.WithLicense(openapi.License{Name: "Apache 2.0", Identifier: "Apache-2.0"}),
		option.WithComponentPathItem("Reusable", &openapi.PathItem{
			Get: &openapi.Operation{Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
		}),
		option.WithComponentMediaType("json-seq", &openapi.MediaType{
			ItemSchema: &openapi.Schema{Type: "object"},
		}),
		option.WithComponentSchema("NewerSchema", &openapi.Schema{
			Type: "object",
			Defs: map[string]*openapi.Schema{
				"ID": {Type: "string"},
			},
		}),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Webhooks = map[string]*openapi.PathItem{
				"created": {
					Post: &openapi.Operation{Responses: map[string]*openapi.Response{"202": {Description: "Accepted"}}},
				},
			}
		}),
	)
	r.Get("/users/{id}", option.Request(new(GetUserRequest)), option.Response(200, new(User)))

	err := r.Validate()
	assertValidationContains(t, err,
		"info.summary requires OpenAPI 3.1.x or 3.2.0",
		"info.license.identifier requires OpenAPI 3.1.x or 3.2.0",
		"jsonSchemaDialect requires OpenAPI 3.1.x or 3.2.0",
		"webhooks requires OpenAPI 3.1.x or 3.2.0",
		"components.pathItems requires OpenAPI 3.1.x or 3.2.0",
		"components.mediaTypes requires OpenAPI 3.2.0",
		"contains JSON Schema dialect fields",
	)
}

func TestValidate_Document_OpenAPI312_Rejects320Fields(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Invalid 3.1"),
		option.WithVersion("1.0.0"),
		option.WithComponentMediaType("json-seq", &openapi.MediaType{
			ItemSchema: &openapi.Schema{Type: "object"},
		}),
		option.WithComponentSchema("With32Extras", &openapi.Schema{
			Type: "object",
			Discriminator: &openapi.Discriminator{
				PropertyName: "kind",
				Extra:        map[string]any{"defaultMapping": "Other"},
			},
			XML: &openapi.XML{Extra: map[string]any{"nodeType": "element"}},
		}),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["/search"] = &openapi.PathItem{
				Query: &openapi.Operation{
					Responses: map[string]*openapi.Response{"200": {Description: "OK"}},
				},
			}
		}),
	)
	r.Get("/search", option.Response(200, new([]User)))

	err := r.Validate()
	assertValidationContains(t, err,
		"QUERY operation at /search requires OpenAPI 3.2.0",
		"components.mediaTypes requires OpenAPI 3.2.0",
		"components.mediaTypes.json-seq.itemSchema requires OpenAPI 3.2.0",
		"components.schemas.With32Extras.discriminator.defaultMapping requires OpenAPI 3.2.0",
		"components.schemas.With32Extras.xml.nodeType requires OpenAPI 3.2.0",
	)
}

func TestValidate_Document_OpenAPI304_RejectsEmptyWebhooksField(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version304),
		option.WithTitle("Invalid 3.0"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Webhooks = map[string]*openapi.PathItem{}
		}),
	)

	err := r.Validate()
	assertValidationContains(t, err, "webhooks requires OpenAPI 3.1.x or 3.2.0")
}

func TestValidate_Document_OpenAPI304_RejectsEmptyComponentsPathItems(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version304),
		option.WithTitle("Invalid 3.0"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Components = &openapi.Components{
				Schemas: map[string]*openapi.Schema{
					"User": {Type: "object"},
				},
				PathItems: map[string]*openapi.PathItem{},
			}
		}),
	)

	err := r.Validate()
	assertValidationContains(t, err, "components.pathItems requires OpenAPI 3.1.x or 3.2.0")
}

func TestValidate_Document_OpenAPI312_RejectsEmptyMediaTypesField(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Invalid 3.1"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Components = &openapi.Components{
				Schemas: map[string]*openapi.Schema{
					"User": {Type: "object"},
				},
				MediaTypes: map[string]*openapi.MediaType{},
			}
		}),
	)

	err := r.Validate()
	assertValidationContains(t, err, "components.mediaTypes requires OpenAPI 3.2.0")
}

func TestValidateInfo(t *testing.T) {
	t.Run("Summary30", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{Summary: "foo"}, openapi.Version304)
		assertValidationErrorsContains(t, errs, "info.summary requires OpenAPI 3.1.x or 3.2.0")
	})

	t.Run("InvalidTOS", func(t *testing.T) {
		tos := "://bad"
		errs := validate.ValidateInfo(openapi.Info{TermsOfService: &tos}, openapi.Version312)
		assertValidationErrorsContains(t, errs, "info.termsOfService must be a URI")
	})

	t.Run("ContactInvalidURL", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{
			Contact: &openapi.Contact{URL: "://bad"},
		}, openapi.Version312)
		assertValidationErrorsContains(t, errs, "info.contact.url must be a URI")
	})

	t.Run("ContactInvalidEmail", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{
			Contact: &openapi.Contact{Email: "not-an-email"},
		}, openapi.Version312)
		assertValidationErrorsContains(t, errs, "info.contact.email must be an email address")
	})

	t.Run("LicenseMissingName", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{
			License: &openapi.License{},
		}, openapi.Version312)
		assertValidationErrorsContains(t, errs, "info.license.name is required")
	})

	t.Run("LicenseInvalidURL", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{
			License: &openapi.License{Name: "MIT", URL: "://bad"},
		}, openapi.Version312)
		assertValidationErrorsContains(t, errs, "info.license.url must be a URI")
	})

	t.Run("LicenseIdentifier30", func(t *testing.T) {
		errs := validate.ValidateInfo(openapi.Info{
			License: &openapi.License{Name: "MIT", Identifier: "MIT"},
		}, openapi.Version304)
		assertValidationErrorsContains(t, errs, "info.license.identifier requires OpenAPI 3.1.x or 3.2.0")
	})
}

func assertValidationErrorsContains(t *testing.T, errs []error, msg string) {
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), msg) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error %q not found in %v", msg, errs)
}

func TestValidate_Document_AllowsComponentsWithoutPaths(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Components Only"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths = nil
			doc.Components = &openapi.Components{
				Schemas: map[string]*openapi.Schema{
					"User": {Type: "object"},
				},
			}
		}),
	)

	assert.NoError(t, r.Validate())
}

func TestValidate_Document_URIFields(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithSelf("not a uri"),
		option.WithJSONSchemaDialect("relative-dialect"),
		option.WithTags(openapi.Tag{
			Name: "users",
			ExternalDocs: &openapi.ExternalDocs{
				URL: "not a uri",
			},
		}),
	)
	r.Get("/uri", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err,
		"$self requires OpenAPI 3.2.0",
		"$self must be a URI reference",
		"tags[0].externalDocs.url must be a URI",
	)
}

func TestValidate_Document_Self_AllowedIn320(t *testing.T) {
	errs := validate.ValidateDocument(&openapi.Document{
		OpenAPI: openapi.Version320,
		Self:    "https://api.example.com/openapi.json",
		Info:    openapi.Info{Title: "Test", Version: "1.0.0"},
		Paths:   map[string]*openapi.PathItem{},
	}, openapi.Version320)
	// $self should not trigger a version error for 3.2.0
	for _, e := range errs {
		if e.Error() == "$self requires OpenAPI 3.2.0" {
			t.Fatalf("unexpected $self version error in 3.2.0")
		}
	}
}

func TestValidate_Document_AllowsEmptyPathsIn312(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Empty Paths"),
		option.WithVersion("1.0.0"),
	)

	assert.NoError(t, r.Validate())
}

func TestValidate_Document_AllowsRelativeURIs(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Relative URIs"),
		option.WithVersion("1.0.0"),
		option.WithJSONSchemaDialect("schemas/dialect"),
		option.WithTermsOfService("terms"),
		option.WithContact(openapi.Contact{
			URL:   "../contact",
			Email: "api@example.com",
		}),
		option.WithLicense(openapi.License{
			Name: "MIT",
			URL:  "./license",
		}),
		option.WithExternalDocs("../docs"),
		option.WithServer("/v1"),
	)
	r.Get("/uri", option.Response(204, nil))

	assert.NoError(t, r.Validate())
}

func TestValidate_Document_TagNamesUnique(t *testing.T) {
	r := spec.NewRouter(
		option.WithTags(
			openapi.Tag{Name: "users"},
			openapi.Tag{Name: "users"},
		),
	)
	r.Get("/tags", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err, `tags[1].name "users" duplicates tags[0].name`)
}

func TestValidate_Document_OpenAPI320_TagParentExists(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTags(
			openapi.Tag{Name: "users", Parent: "missing"},
		),
	)
	r.Get("/tags", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err, `tags[0].parent "missing" must reference an existing tag`)
}

func TestValidate_Document_OpenAPI320_TagParentCircular(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTags(
			openapi.Tag{Name: "users", Parent: "accounts"},
			openapi.Tag{Name: "accounts", Parent: "users"},
		),
	)
	r.Get("/tags", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err, "tags[0].parent creates a circular tag parent reference")
}

func TestValidate_Document_Server(t *testing.T) {
	t.Run("Invalid URL", func(t *testing.T) {
		r := spec.NewRouter(option.WithServer(""))
		err := r.Validate()
		assertValidationContains(t, err, "servers[0].url is required")
	})

	t.Run("Invalid Variables", func(t *testing.T) {
		// Default router uses 3.0.x where SHOULD applies — expect a warning, not an error.
		errs := validate.ValidateServer("servers[0]", &openapi.Server{
			URL: "https://{host}",
			Variables: map[string]openapi.ServerVariable{
				"host": {Default: ""},
				"port": {Default: "80", Enum: []string{"8080"}},
			},
		}, openapi.Version304)
		assertHasError(t, errs, "servers[0].variables.host.default is required")
		assertHasWarning(t, errs, "servers[0].variables.port.default should be one of enum values")
	})

	t.Run("Invalid Variables 3.1+", func(t *testing.T) {
		// In 3.1+ MUST applies — expect an error.
		errs := validate.ValidateServer("servers[0]", &openapi.Server{
			URL: "https://{host}",
			Variables: map[string]openapi.ServerVariable{
				"port": {Default: "80", Enum: []string{"8080"}},
			},
		}, openapi.Version312)
		assertHasError(t, errs, "servers[0].variables.port.default must be one of enum values")
	})

	t.Run("QueryOrFragment", func(t *testing.T) {
		r := spec.NewRouter(option.WithServer("/v1?x=1"))
		err := r.Validate()
		assertValidationContains(t, err, "servers[0].url must not contain a query or fragment")
	})
}

func TestValidate_Server_EmptyEnum(t *testing.T) {
	t.Run("3.1+ empty enum is an error", func(t *testing.T) {
		for _, version := range []string{openapi.Version312, openapi.Version320} {
			r := spec.NewRouter(
				option.WithOpenAPIVersion(version),
				option.WithServer("https://{env}.example.com", option.ServerVariables(map[string]openapi.ServerVariable{
					"env": {Default: "prod", Enum: []string{}},
				})),
			)
			err := r.Validate()
			assertValidationContains(t, err, "enum must not be empty")
		}
	})

	t.Run("3.0 empty enum is a warning only", func(t *testing.T) {
		errs := validate.ValidateServer("servers[0]", &openapi.Server{
			URL: "https://{env}.example.com",
			Variables: map[string]openapi.ServerVariable{
				"env": {Default: "prod", Enum: []string{}},
			},
		}, openapi.Version304)
		assertHasWarning(t, errs, "should not be empty")
	})
}

func TestValidate_Server_DuplicateVariableInURL(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithServer("https://{env}.{env}.example.com", option.ServerVariables(map[string]openapi.ServerVariable{
			"env": {Default: "prod"},
		})),
	)
	err := r.Validate()
	assertValidationContains(t, err, "variable {env} must not appear more than once")
}

func TestValidate_Server_NameVersionGate(t *testing.T) {
	t.Run("3.2.0 allows server name", func(t *testing.T) {
		errs := validate.ValidateServerNames([]openapi.Server{
			{URL: "https://api.example.com", Name: "production"},
		}, openapi.Version320)
		assertNoStrictErrors(t, errs)
	})

	t.Run("pre-3.2.0 rejects server name", func(t *testing.T) {
		for _, version := range []string{openapi.Version304, openapi.Version312} {
			errs := validate.ValidateServerNames([]openapi.Server{
				{URL: "https://api.example.com", Name: "production"},
			}, version)
			assertHasError(t, errs, "servers[0].name requires OpenAPI 3.2.0")
		}
	})
}
