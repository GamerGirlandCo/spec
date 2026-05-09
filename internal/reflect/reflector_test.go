package reflect_test

import (
	std_reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestReflector_ParameterSchema(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	r := reflect.NewReflector(cfg)

	type ParamStruct struct {
		ID string `path:"id" description:"User ID" required:"true" deprecated:"true"`
	}
	f, _ := std_reflect.TypeFor[ParamStruct]().FieldByName("ID")

	p := r.ParameterSchema(f, "path", "id")
	assert.Equal(t, "id", p.Name)
	assert.Equal(t, "path", p.In)
	assert.Equal(t, "User ID", p.Description)
	assert.True(t, p.Required)
	assert.True(t, p.Deprecated)
	assert.Equal(t, "string", p.Schema.Type)
}

func TestReflector_SchemaForValue(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	r := reflect.NewReflector(cfg)

	t.Run("OneOf", func(t *testing.T) {
		val := spec.OneOf(1, "two")
		schema := r.SchemaForValue(val, reflect.SchemaInline)
		assert.Len(t, schema.OneOf, 2)
	})

	t.Run("SchemaPointer", func(t *testing.T) {
		expected := &openapi.Schema{Type: "boolean"}
		schema := r.SchemaForValue(expected, reflect.SchemaInline)
		assert.Equal(t, expected, schema)
	})
}

func TestReflector_Config(t *testing.T) {
	t.Run("InterceptDefName", func(t *testing.T) {
		r := spec.NewRouter(option.WithReflectorConfig(
			option.InterceptDefName(func(_ std_reflect.Type, _ string) string {
				return "CustomName"
			}),
		))
		type NamedStruct struct{ Foo string }
		r.Get("/ping", option.Response(200, NamedStruct{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		assert.Contains(t, doc.Components.Schemas, "CustomName")
	})

	t.Run("DuplicateNames", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithReflectorConfig(option.InterceptDefName(func(_ std_reflect.Type, _ string) string {
				return "Collision"
			})),
		)

		type TypeA struct{ Foo string }
		type TypeB struct{ Bar string }

		r.Get("/a", option.Response(200, TypeA{}))
		r.Get("/b", option.Response(200, TypeB{}))

		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()

		assert.Contains(t, doc.Components.Schemas, "Collision")
		assert.Contains(t, doc.Components.Schemas, "Collision2")
	})
}

func TestReflector_ParameterField_CustomMappingKeepsDefaultTag(t *testing.T) {
	cfg := option.WithOpenAPIConfig(
		option.WithReflectorConfig(option.ParameterTagMapping(openapi.ParameterInPath, "param")),
	)
	r := reflect.NewReflector(cfg)

	type Request struct {
		ID int `path:"id" required:"true"`
	}

	params, _ := r.RequestParts(Request{}, "")
	require.Len(t, params, 1)
	assert.Equal(t, "id", params[0].Name)
	assert.Equal(t, "path", params[0].In)
	assert.True(t, params[0].Required)
}
