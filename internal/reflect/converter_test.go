package reflect_test

import (
	std_reflect "reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type CustomSlug string

func (*CustomSlug) OpenAPISchema(version string) *openapi.Schema {
	if strings.HasPrefix(version, "3.0.") {
		return &openapi.Schema{Type: "string", Format: "slug"}
	}
	return &openapi.Schema{Type: []string{"string", "null"}, Format: "slug"}
}

type CustomSchemaPayload struct {
	ID CustomSlug `json:"id" description:"Stable identifier"`
}

type SchemaExposerType struct{}

func (SchemaExposerType) OpenAPISchema(_ string) *openapi.Schema {
	return &openapi.Schema{Type: "string", Description: "Exposed"}
}

func TestConverter_SchemaForType(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version312}
	r := reflect.NewReflector(cfg)

	tests := []struct {
		name     string
		val      any
		expected string
	}{
		{"Int64", int64(0), "integer"},
		{"Uint32", uint32(0), "integer"},
		{"Float32", float32(0), "number"},
		{"ByteSlice", []byte{0}, "string"},
		{"Map", map[string]int{"a": 1}, "object"},
		{"Interface", any(nil), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := std_reflect.TypeOf(tt.val)
			if tt.name == "Interface" {
				typ = std_reflect.TypeFor[any]()
			}
			schema, err := r.SchemaForType(typ, reflect.SchemaInline, nil)
			require.NoError(t, err)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, schema.Type)
			}
		})
	}
}

func TestConverter_ApplyNullable_EdgeCases(t *testing.T) {
	t.Run("OpenAPI304_Ref", func(t *testing.T) {
		r := reflect.NewReflector(&openapi.Config{OpenAPIVersion: openapi.Version304})
		schema := &openapi.Schema{Ref: "#/components/schemas/User"}
		r.ApplyNullable(schema, true)
		assert.Len(t, schema.AllOf, 1)
		assert.True(t, schema.Nullable)
	})

	t.Run("OpenAPI312_Ref", func(t *testing.T) {
		r := reflect.NewReflector(&openapi.Config{OpenAPIVersion: openapi.Version312})
		schema := &openapi.Schema{Ref: "#/components/schemas/User"}
		r.ApplyNullable(schema, true)
		assert.Len(t, schema.AnyOf, 2)
		assert.Equal(t, "null", schema.AnyOf[1].Type)
	})
}

func TestConverter_CustomSchemaExposer(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Custom Schema"), option.WithVersion("1.0.0"))
	r.Post("/payload", option.Request(new(CustomSchemaPayload)), option.Response(204, nil))

	raw, err := r.GenerateSchema("json")
	require.NoError(t, err)

	id := generatedComponentProperty(t, raw, "CustomSchemaPayload", "id")
	assert.Equal(t, "slug", id["format"])
	assert.Equal(t, "Stable identifier", id["description"])
}

func TestConverter_Nullable(t *testing.T) {
	t.Run("OpenAPI304", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Post("/payload", option.Request(new(ReflectionVersionPayload)), option.Response(204, nil))

		raw, err := r.GenerateSchema("json")
		require.NoError(t, err)

		owner := generatedComponentProperty(t, raw, "ReflectionVersionPayload", "owner")
		assert.NotContains(t, owner, "$ref", "nullable component refs must not emit $ref siblings in 3.0")
		assert.Equal(t, true, owner["nullable"])
	})

	t.Run("OpenAPI312", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version312))
		r.Post("/payload", option.Request(new(ReflectionVersionPayload312)), option.Response(204, nil))

		raw, err := r.GenerateSchema("json")
		require.NoError(t, err)

		name := generatedComponentProperty(t, raw, "ReflectionVersionPayload312", "name")
		typ := name["type"].([]any)
		assert.Equal(t, []any{"string", "null"}, typ)
	})
}

func TestConverter_SchemaExposer(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Exposer"))
	r.Get("/exposer", option.Response(200, SchemaExposerType{}))
	_, err := r.GenerateSchema("yaml")
	require.NoError(t, err)
	doc := r.Document()
	assert.Equal(
		t,
		"Exposed",
		doc.Paths["/exposer"].Get.Responses["200"].Content["application/json"].Schema.Description,
	)
}
