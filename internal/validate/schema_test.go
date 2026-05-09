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

func TestValidateOpenAPI312SchemaIDReferenceTargets(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Schema ID Refs"),
		option.WithVersion("1.0.0"),
		option.WithComponentSchema("Foo", &openapi.Schema{
			ID:   "https://schemas.example/foo",
			Type: "object",
			Defs: map[string]*openapi.Schema{
				"Bar": {Type: "string"},
			},
			Properties: map[string]*openapi.Schema{
				"bar": {Ref: "#/$defs/Bar"},
			},
		}),
	)
	r.Get("/foo", option.Response(200, &openapi.Schema{Ref: "https://schemas.example/foo"}))

	assert.NoError(t, r.Validate())
}

func TestValidateOpenAPI312SchemaIDScopedPointerTarget(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Invalid Schema ID Refs"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Components.Schemas = map[string]*openapi.Schema{
				"Foo": {
					ID:   "https://schemas.example/foo",
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"bar": {Ref: "#/components/schemas/Bar"},
					},
				},
				"Bar": {Type: "string"},
			}
		}),
	)
	r.Get("/foo", option.Response(200, &openapi.Schema{Ref: "#/components/schemas/Foo"}))

	err := r.Validate()
	assertValidationContains(t, err, `$ref "#/components/schemas/Bar" points to a missing target`)
}

func TestValidateSchema_EdgeCases(t *testing.T) {
	t.Run("OpenAPI304_RefSiblings", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Get("/test", option.Response(200, &openapi.Schema{
			Ref:         "#/components/schemas/User",
			Description: "sibling",
		}))
		err := r.Validate()
		assertValidationContains(t, err, "must not define siblings with $ref in OpenAPI 3.0.x")
	})

	t.Run("OpenAPI304_ReadOnlyWriteOnly", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Get("/test", option.Response(200, &openapi.Schema{
			ReadOnly:  true,
			WriteOnly: true,
		}))
		err := r.Validate()
		assertValidationContains(t, err, "must not be both readOnly and writeOnly")
	})

	t.Run("OpenAPI312_320OnlyFields", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version312))
		r.Get("/test", option.Response(200, &openapi.Schema{
			XML: &openapi.XML{Extra: map[string]any{"nodeType": "attr"}},
		}))
		err := r.Validate()
		assertValidationContains(t, err, "xml.nodeType requires OpenAPI 3.2.0")
	})
}

func TestValidateAnySchema(t *testing.T) {
	t.Run("ValueType", func(t *testing.T) {
		r := spec.NewRouter()
		r.Get("/test", option.Response(200, &openapi.Schema{
			AdditionalProperties: openapi.Schema{Type: "string"},
		}))
		assert.NoError(t, r.Validate())
	})
}

func TestSchemaTypeIncludesArray(t *testing.T) {
	assert.False(t, validate.SchemaTypeIncludesArray(nil))
	assert.True(t, validate.SchemaTypeIncludesArray(&openapi.Schema{Type: "array"}))
	assert.False(t, validate.SchemaTypeIncludesArray(&openapi.Schema{Type: "object"}))
	assert.True(t, validate.SchemaTypeIncludesArray(&openapi.Schema{Type: []string{"object", "array"}}))
	assert.True(t, validate.SchemaTypeIncludesArray(&openapi.Schema{Type: []any{"string", "array"}}))
	assert.False(t, validate.SchemaTypeIncludesArray(&openapi.Schema{Type: []any{"string", 1}}))
}

func TestValidateSchema304Fields_Direct(t *testing.T) {
	s := &openapi.Schema{
		Schema:           "https://json-schema.org/draft/2020-12/schema",
		Examples:         []any{"x"},
		Type:             []string{"string"},
		Extra:            map[string]any{"$schema": "x"},
		ExclusiveMaximum: 1.2,
		ExclusiveMinimum: 0.1,
	}

	errs := validate.ValidateSchema304Fields("schema", s)
	assert.GreaterOrEqual(t, len(errs), 6)
	assert.Contains(t, errs[0].Error(), "contains JSON Schema dialect fields")
	assert.Contains(t, errs[1].Error(), "contains JSON Schema 2020-12 keywords")

	var joined strings.Builder
	for _, err := range errs {
		joined.WriteString(err.Error() + "\n")
	}
	assert.Contains(t, joined.String(), "schema.type must be a string in OpenAPI 3.0.x")
	assert.Contains(t, joined.String(), "schema.exclusiveMaximum must be a boolean in OpenAPI 3.0.x")
	assert.Contains(t, joined.String(), "schema.exclusiveMinimum must be a boolean in OpenAPI 3.0.x")
	assert.Contains(t, joined.String(), "contains Extra JSON Schema keywords")
}

func TestValidateSchema304Fields_TypeAnySlice(t *testing.T) {
	errs := validate.ValidateSchema304Fields("schema", &openapi.Schema{
		Type: []any{"string"},
	})
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "schema.type must be a string in OpenAPI 3.0.x")
}

func TestValidateAnySchema_Direct(t *testing.T) {
	errs := validate.ValidateAnySchema("schema", 123, openapi.Version304, map[*openapi.Schema]bool{})
	assert.Empty(t, errs)

	errs = validate.ValidateAnySchema(
		"schema",
		openapi.Schema{ReadOnly: true, WriteOnly: true},
		openapi.Version304,
		map[*openapi.Schema]bool{},
	)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "must not be both readOnly and writeOnly")
}
