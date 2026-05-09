package reflect_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type XMLType struct {
	Value string `xmlName:"val" xmlNamespace:"ns" xmlPrefix:"p" xmlAttribute:"true"`
}

func TestTags_OpenAPI304(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Reflection 3.0"), option.WithVersion("1.0.0"))
	r.Post("/payload", option.Request(new(ReflectionVersionPayload)), option.Response(204, nil))

	raw, err := r.GenerateSchema("json")
	require.NoError(t, err)

	text := string(raw)
	for _, forbidden := range []string{`"const"`, `"examples"`, `"contentEncoding"`, `"contentMediaType"`} {
		assert.NotContains(t, text, forbidden, "OpenAPI 3.0.x output contains forbidden keyword")
	}

	name := generatedComponentProperty(t, raw, "ReflectionVersionPayload", "name")
	assert.Equal(t, true, name["nullable"])
	assert.Equal(t, true, name["exclusiveMaximum"])
}

func TestTags_OpenAPI312(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Reflection 3.1"),
		option.WithVersion("1.0.0"),
	)
	r.Post("/payload", option.Request(new(ReflectionVersionPayload312)), option.Response(204, nil))

	raw, err := r.GenerateSchema("json")
	require.NoError(t, err)

	name := generatedComponentProperty(t, raw, "ReflectionVersionPayload312", "name")
	assert.Equal(t, "fixed", name["const"])
	assert.Contains(t, name, "examples")
	assert.Equal(t, "base64", name["contentEncoding"])
	assert.Equal(t, "text/plain", name["contentMediaType"])

	score := generatedComponentProperty(t, raw, "ReflectionVersionPayload312", "score")
	assert.InDelta(t, 0.0, score["exclusiveMinimum"].(float64), 0.0001)
}

func TestTags_XML(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		r := spec.NewRouter(option.WithTitle("XML"))
		r.Get("/xml", option.Response(200, XMLType{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["XMLType"]
		assert.NotNil(t, schema.Properties["value"].XML)
	})

	t.Run("XMLNodeType", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version320))
		type XMLNode struct {
			Attr string `json:"attr" xmlAttribute:"true" xmlNodeType:"attribute"`
		}
		r.Get("/xml-node", option.Response(200, XMLNode{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["XMLNode"].Properties["attr"]
		if assert.NotNil(t, schema.XML) {
			assert.Equal(t, "attribute", schema.XML.Extra["nodeType"])
		}
	})
}

func TestTags_Values(t *testing.T) {
	type TagValueType struct {
		Int    int     `json:"int" default:"123"`
		Bool   bool    `json:"bool" default:"true"`
		Float  float64 `json:"float" default:"1.23"`
		String string  `json:"string" default:"foo"`
		JSON   []int   `json:"json" default:"[1,2,3]"`
	}
	r := spec.NewRouter(option.WithTitle("Tags"))
	r.Get("/tags", option.Response(200, TagValueType{}))
	_, err := r.GenerateSchema("yaml")
	require.NoError(t, err)
	doc := r.Document()
	props := doc.Components.Schemas["TagValueType"].Properties
	assert.Equal(t, int64(123), props["int"].Default)
	assert.Equal(t, true, props["bool"].Default)
}

func TestTags_Constraints(t *testing.T) {
	t.Run("OpenAPI304_Const", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		type V304Type struct {
			Const string `json:"const" const:"foo"`
		}
		r.Get("/v304", option.Response(200, V304Type{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		assert.Nil(t, doc.Components.Schemas["V304Type"].Const)
	})

	t.Run("MultiType", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version312))
		type MultiType struct {
			Value any `json:"value" type:"string,integer"`
		}
		r.Get("/multi", option.Response(200, MultiType{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["MultiType"].Properties["value"]
		assert.Equal(t, []string{"string", "integer"}, schema.Type)
	})

	t.Run("IntTags", func(t *testing.T) {
		r := spec.NewRouter()
		type IntTagType struct {
			Val string `json:"val" maxLength:"10" minLength:"5" maxItems:"3" minItems:"1" maxProperties:"4" minProperties:"2"`
		}
		r.Get("/int", option.Response(200, IntTagType{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["IntTagType"].Properties["val"]
		assert.Equal(t, 10, *schema.MaxLength)
		assert.Equal(t, 5, *schema.MinLength)
		assert.Equal(t, 3, *schema.MaxItems)
		assert.Equal(t, 1, *schema.MinItems)
		assert.Equal(t, 4, *schema.MaxProperties)
		assert.Equal(t, 2, *schema.MinProperties)
	})

	t.Run("FloatTags", func(t *testing.T) {
		r := spec.NewRouter()
		type FloatTagType struct {
			Val float64 `json:"val" multipleOf:"2.5" maximum:"10.5" minimum:"1.5"`
		}
		r.Get("/float", option.Response(200, FloatTagType{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["FloatTagType"].Properties["val"]
		assert.InDelta(t, 2.5, *schema.MultipleOf, 1e-9)
		assert.InDelta(t, 10.5, *schema.Maximum, 1e-9)
		assert.InDelta(t, 1.5, *schema.Minimum, 1e-9)
	})

	t.Run("OtherTags", func(t *testing.T) {
		r := spec.NewRouter()
		type OtherTagType struct {
			Val string `json:"val" pattern:".*" uniqueItems:"true" readOnly:"true" deprecated:"true" title:"T" example:"E"`
		}
		r.Get("/other", option.Response(200, OtherTagType{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		schema := doc.Components.Schemas["OtherTagType"].Properties["val"]
		assert.Equal(t, ".*", schema.Pattern)
		assert.True(t, *schema.UniqueItems)
		assert.True(t, schema.ReadOnly)
		assert.False(t, schema.WriteOnly)
		assert.True(t, schema.Deprecated)
		assert.Equal(t, "T", schema.Title)
		assert.Equal(t, "E", schema.Example)
	})
}
