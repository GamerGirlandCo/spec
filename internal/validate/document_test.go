package validate_test

import (
	"testing"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestValidate_Document_OpenAPI304_RejectsNewerFields(t *testing.T) {
	r := spec.NewRouter(
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

func TestValidate_Document_URIFields(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithSelf("not a uri"),
		option.WithJSONSchemaDialect("relative-dialect"),
	)
	r.Get("/uri", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err,
		"$self must be a URI reference",
		"jsonSchemaDialect must be a URI",
	)
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
		r := spec.NewRouter(
			option.WithServer("https://{host}", option.ServerVariables(map[string]openapi.ServerVariable{
				"host": {Default: ""},
				"port": {Default: "80", Enum: []string{"8080"}},
			})),
		)
		err := r.Validate()
		assertValidationContains(t, err,
			"servers[0].variables.host.default is required",
			"servers[0].variables.port.default must be one of enum values",
		)
	})
}
