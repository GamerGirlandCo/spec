package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
)

func TestValidate_ServerNameUniqueness(t *testing.T) {
	doc := &openapi.Document{
		OpenAPI: openapi.Version320,
		Info:    openapi.Info{Title: "Test", Version: "1.0.0"},
		Servers: []openapi.Server{
			{URL: "https://v1.example.com", Name: "prod"},
			{URL: "https://v2.example.com", Name: "prod"},
		},
		Paths: map[string]*openapi.PathItem{},
	}
	errs := validate.ValidateDocument(doc, openapi.Version320)
	assert.NotEmpty(t, errs)
	assertHasError(t, errs, "duplicates servers[0].name")
}

func TestValidate_ExclusiveBoundaryTypes(t *testing.T) {
	t.Run("OpenAPI 3.0.4 - boolean required", func(t *testing.T) {
		schema := &openapi.Schema{
			ExclusiveMaximum: 100.0,
		}
		errs := validate.ValidateSchema("schema", schema, openapi.Version304, map[*openapi.Schema]bool{})
		assert.NotEmpty(t, errs)
		assertHasError(t, errs, "must be a boolean in OpenAPI 3.0.x")
	})

	t.Run("OpenAPI 3.1.0 - number required", func(t *testing.T) {
		schema := &openapi.Schema{
			ExclusiveMaximum: true,
		}
		errs := validate.ValidateSchema("schema", schema, openapi.Version310, map[*openapi.Schema]bool{})
		assert.NotEmpty(t, errs)
		assertHasError(t, errs, "must be a number in OpenAPI 3.1.x or 3.2.0")
	})

	t.Run("OpenAPI 3.2.0 - number valid", func(t *testing.T) {
		schema := &openapi.Schema{
			ExclusiveMaximum: 100.0,
		}
		errs := validate.ValidateSchema("schema", schema, openapi.Version320, map[*openapi.Schema]bool{})
		assertNoStrictErrors(t, errs)
	})
}

func TestValidate_320Fields_NewNodeTypeAndDefaultMapping(t *testing.T) {
	t.Run("Discriminator DefaultMapping", func(t *testing.T) {
		schema := &openapi.Schema{
			Discriminator: &openapi.Discriminator{
				PropertyName:   "type",
				DefaultMapping: "DefaultSchema",
			},
		}
		errs := validate.ValidateSchema("schema", schema, openapi.Version312, map[*openapi.Schema]bool{})
		assert.NotEmpty(t, errs)
		assertHasError(t, errs, "requires OpenAPI 3.2.0")
	})

	t.Run("XML NodeType", func(t *testing.T) {
		schema := &openapi.Schema{
			XML: &openapi.XML{
				NodeType: "element",
			},
		}
		errs := validate.ValidateSchema("schema", schema, openapi.Version312, map[*openapi.Schema]bool{})
		assert.NotEmpty(t, errs)
		assertHasError(t, errs, "requires OpenAPI 3.2.0")
	})
}

func TestValidate_ForbiddenHeaderNames(t *testing.T) {
	t.Run("Parameter headers", func(t *testing.T) {
		op := &openapi.Operation{
			Responses: map[string]*openapi.Response{"200": {Description: "OK"}},
			Parameters: []*openapi.Parameter{
				{Name: "Authorization", In: "header", Schema: &openapi.Schema{Type: "string"}},
			},
		}
		errs := validate.ValidateOperation("op", op, openapi.Version320, map[string]string{}, nil, nil)
		assert.NotEmpty(t, errs)
		assertHasError(t, errs, "not allowed for header parameters")
	})

	t.Run("Response headers", func(t *testing.T) {
		resp := &openapi.Response{
			Description: "OK",
			Headers: map[string]*openapi.Header{
				"Content-Type": {Schema: &openapi.Schema{Type: "string"}},
			},
		}
		errs := validate.ValidateResponse("resp", resp, openapi.Version320)
		assert.NotEmpty(t, errs)
		assertHasWarning(t, errs, "is ignored by the OpenAPI spec")
	})
}

func TestValidate_DiscriminatorUsage(t *testing.T) {
	schema := &openapi.Schema{
		Discriminator: &openapi.Discriminator{PropertyName: "type"},
		Type:          "object",
	}
	errs := validate.ValidateSchema("schema", schema, openapi.Version320, map[*openapi.Schema]bool{})
	assert.NotEmpty(t, errs)
	assertHasError(t, errs, "only allowed with anyOf, oneOf, or allOf")
}

func TestValidate_EncodingContext(t *testing.T) {
	mediaType := &openapi.MediaType{
		Schema: &openapi.Schema{Type: "object"},
		Encoding: map[string]*openapi.Encoding{
			"prop": {Style: "form"},
		},
	}
	errs := validate.ValidateMediaType("mt", "application/json", mediaType, openapi.Version320)
	assert.NotEmpty(t, errs)
	assertHasError(t, errs, "requires multipart or application/x-www-form-urlencoded")
}

func TestOptions_320NewFields(t *testing.T) {
	// Using internal package because we can't import spec from validate_test (cyclic)
	// Actually we are in validate_test, so we can't import spec.
	// But we can check if the structs are correctly populated if we had a builder.
	// Since we are in internal/validate, let's just test the document structure.
	doc := &openapi.Document{
		OpenAPI: openapi.Version320,
		Info:    openapi.Info{Title: "Test", Version: "1.0.0"},
		Servers: []openapi.Server{
			{URL: "https://example.com", Name: "prod"},
		},
		Paths: map[string]*openapi.PathItem{
			"/test": {
				Get: &openapi.Operation{
					Summary: "Test Summary",
					RequestBody: &openapi.RequestBody{
						Description: "Request Description",
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{Type: "string"},
							},
						},
					},
					Responses: map[string]*openapi.Response{
						"200": {
							Summary:     "Response Summary",
							Description: "OK",
						},
					},
				},
			},
		},
	}
	errs := validate.ValidateDocument(doc, openapi.Version320)
	assertNoStrictErrors(t, errs)
}
