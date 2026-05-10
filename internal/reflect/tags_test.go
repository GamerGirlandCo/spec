package reflect_test

import (
	std_reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
)

func TestIgnoredField(t *testing.T) {
	type TestStruct struct {
		IgnoredJSON        string `json:"-"`
		IgnoredCustom      string `custom:"-"`
		NotIgnored         string `json:"foo"`
		JSONDashNotIgnored string `json:"'-',omitempty"`
	}

	typ := std_reflect.TypeFor[TestStruct]()

	f, _ := typ.FieldByName("IgnoredJSON")
	assert.True(t, reflect.IgnoredField(f, "json"))
	assert.True(t, reflect.IgnoredField(f, "custom"))

	f, _ = typ.FieldByName("IgnoredCustom")
	assert.False(t, reflect.IgnoredField(f, "json"))
	assert.True(t, reflect.IgnoredField(f, "custom"))

	f, _ = typ.FieldByName("NotIgnored")
	assert.False(t, reflect.IgnoredField(f, "json"))

	f, _ = typ.FieldByName("JSONDashNotIgnored")
	assert.False(t, reflect.IgnoredField(f, "json"))
}

func TestParseTypeTag(t *testing.T) {
	assert.Equal(t, "string", reflect.ParseTypeTag("string", "3.0.4"))
	assert.Equal(t, "string", reflect.ParseTypeTag("string,null", "3.0.4"))
	assert.Equal(t, []string{"string", "null"}, reflect.ParseTypeTag("string,null", "3.1.0"))
	assert.Equal(t, "string", reflect.ParseTypeTag("string", "3.1.0"))
}

func TestNormalizeTagValue(t *testing.T) {
	assert.Equal(t, int64(123), reflect.NormalizeTagValue(123.0))
	assert.InDelta(t, 123.45, reflect.NormalizeTagValue(123.45), 0.0001)
	assert.Equal(t, "foo", reflect.NormalizeTagValue("foo"))
	assert.Equal(t, []any{int64(1)}, reflect.NormalizeTagValue([]any{1.0}))
	assert.Equal(t, map[string]any{"k": int64(1)}, reflect.NormalizeTagValue(map[string]any{"k": 1.0}))
}

func TestIntTag(t *testing.T) {
	assert.Nil(t, reflect.IntTag(""))
	assert.Nil(t, reflect.IntTag("abc"))
	i := reflect.IntTag("123")
	require.NotNil(t, i)
	assert.Equal(t, 123, *i)
}

func TestApplyXMLTags(t *testing.T) {
	r304 := reflect.NewReflector(&openapi.Config{OpenAPIVersion: openapi.Version304})
	r320 := reflect.NewReflector(&openapi.Config{OpenAPIVersion: openapi.Version320})

	t.Run("OpenAPI 3.0.4", func(t *testing.T) {
		s := &openapi.Schema{}
		tag := std_reflect.StructTag(`xmlName:"foo" xmlAttribute:"true" xmlWrapped:"true"`)
		r304.ApplyXMLTags(s, tag)
		require.NotNil(t, s.XML)
		assert.Equal(t, "foo", s.XML.Name)
		assert.True(t, s.XML.Attribute)
		assert.True(t, s.XML.Wrapped)
	})

	t.Run("OpenAPI 3.2.0 NodeType", func(t *testing.T) {
		s := &openapi.Schema{}
		tag := std_reflect.StructTag(`xmlNodeType:"attribute"`)
		r320.ApplyXMLTags(s, tag)
		require.NotNil(t, s.XML)
		assert.Equal(t, "attribute", s.XML.NodeType)
	})

	t.Run("OpenAPI 3.2.0 Attribute Fallback", func(t *testing.T) {
		s := &openapi.Schema{}
		tag := std_reflect.StructTag(`xmlAttribute:"true"`)
		r320.ApplyXMLTags(s, tag)
		assert.Equal(t, "attribute", s.XML.NodeType)
	})

	t.Run("OpenAPI 3.2.0 Wrapped Fallback", func(t *testing.T) {
		s := &openapi.Schema{}
		tag := std_reflect.StructTag(`xmlWrapped:"true"`)
		r320.ApplyXMLTags(s, tag)
		assert.Equal(t, "element", s.XML.NodeType)
	})

	t.Run("Empty", func(t *testing.T) {
		s := &openapi.Schema{}
		r304.ApplyXMLTags(s, "")
		assert.Nil(t, s.XML)
	})
}
