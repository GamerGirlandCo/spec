package validate

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/openapi"
)

func TestNormalizeTemplatedPath(t *testing.T) {
	assert.Equal(t, "/users/{}", NormalizeTemplatedPath("/users/{id}"))
	assert.Equal(t, "/orgs/{}/repos/{}", NormalizeTemplatedPath("/orgs/{org}/repos/{repo}"))
	assert.Equal(t, "/static", NormalizeTemplatedPath("/static"))
}

func TestMediaTypeBase(t *testing.T) {
	assert.Equal(t, "application/json", MediaTypeBase("application/json"))
	assert.Equal(t, "application/json", MediaTypeBase("application/json; charset=utf-8"))
	assert.Equal(t, "application/json", MediaTypeBase("  APPLICATION/JSON  ; foo=bar"))
}

func TestMediaTypeIsMultipart(t *testing.T) {
	assert.True(t, MediaTypeIsMultipart("multipart/form-data"))
	assert.True(t, MediaTypeIsMultipart("multipart/mixed"))
	assert.False(t, MediaTypeIsMultipart("application/json"))
}

func TestResolveJSONPointer(t *testing.T) {
	root := map[string]any{
		"foo": []any{"bar", "baz"},
		"qux": map[string]any{
			"a/b": 1,
			"c%d": 2,
			"e~f": 3,
			"g~h": 4,
		},
	}

	assert.Equal(t, root, ResolveJSONPointer(root, ""))
	assert.Equal(t, root["foo"], ResolveJSONPointer(root, "/foo"))
	assert.Equal(t, "bar", ResolveJSONPointer(root, "/foo/0"))
	assert.Equal(t, "baz", ResolveJSONPointer(root, "/foo/1"))
	assert.Nil(t, ResolveJSONPointer(root, "/foo/2"))
	assert.Equal(t, 1, ResolveJSONPointer(root, "/qux/a~1b"))
	assert.Equal(t, 3, ResolveJSONPointer(root, "/qux/e~0f"))
}

func TestIsNonRelativeURI(t *testing.T) {
	assert.True(t, IsNonRelativeURI("https://example.com"))
	assert.True(t, IsNonRelativeURI("https://example.com#frag"))
	assert.True(t, IsNonRelativeURI("mailto:foo@example.com"))
	assert.False(t, IsNonRelativeURI("/local/path"))
	assert.False(t, IsNonRelativeURI("relative"))
}

func TestIsHTTPSURI(t *testing.T) {
	assert.True(t, IsHTTPSURI("https://example.com"))
	assert.False(t, IsHTTPSURI("http://example.com"))
	assert.False(t, IsHTTPSURI("ftp://example.com"))
}

func TestIsURIReference(t *testing.T) {
	assert.True(t, IsURIReference("https://example.com"))
	assert.True(t, IsURIReference("/path"))
	assert.False(t, IsURIReference("not a uri with spaces"))
}

func TestResolveURIReference(t *testing.T) {
	tests := []struct {
		base, ref string
		expected  string
	}{
		{"", "https://example.com", "https://example.com"},
		{"https://example.com/a/b", "c", "https://example.com/a/c"},
		{"https://example.com/a/b", "/c", "https://example.com/c"},
	}
	for _, tt := range tests {
		got, ok := ResolveURIReference(tt.base, tt.ref)
		assert.True(t, ok)
		assert.Equal(t, tt.expected, got)
	}
}

func TestExtraHas(t *testing.T) {
	extra := map[string]any{"foo": 1, "bar": 2}
	assert.True(t, ExtraHas(extra, "foo"))
	assert.True(t, ExtraHas(extra, "baz", "bar"))
	assert.False(t, ExtraHas(extra, "qux"))
}

func TestWithoutFragment(t *testing.T) {
	assert.Equal(t, "https://example.com/path", WithoutFragment("https://example.com/path#frag"))
	assert.Equal(t, "https://example.com/path", WithoutFragment("https://example.com/path"))
	assert.Equal(t, ":// bad", WithoutFragment(":// bad"))
}

func TestResolveURIReference_InvalidInput(t *testing.T) {
	_, ok := ResolveURIReference("https://example.com", "://bad")
	assert.False(t, ok)

	_, ok = ResolveURIReference("://bad", "x")
	assert.False(t, ok)
}

func TestRegisterSchemaResourceAndAnchor(t *testing.T) {
	schema := &openapi.Schema{
		ID:            "https://schemas.example.com/user",
		Type:          "object",
		Anchor:        "user-anchor",
		DynamicAnchor: "user-dyn",
	}

	resources := map[string]any{}
	RegisterSchemaResource(reflect.ValueOf(*schema), schema.ID, resources)
	RegisterSchemaResource(reflect.ValueOf(*schema), schema.ID, resources) // idempotent

	assert.Contains(t, resources, "https://schemas.example.com/user")
	assert.Contains(t, resources, "https://schemas.example.com/user#user-anchor")
	assert.Contains(t, resources, "https://schemas.example.com/user#user-dyn")
	assert.Len(t, resources, 3)

	// Empty anchor should be ignored.
	emptySchema := openapi.Schema{}
	RegisterSchemaAnchor(reflect.ValueOf(emptySchema), "", "Anchor", resources)
	assert.Len(t, resources, 3)
}

func TestIsLocalReferenceAndReferenceTargetExists(t *testing.T) {
	resources := map[string]any{
		"https://example.com/schemas/user": map[string]any{
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		},
		"https://example.com/schemas/user#UserAnchor": map[string]any{"type": "object"},
		"": map[string]any{
			"components": map[string]any{
				"schemas": map[string]any{
					"User": map[string]any{"type": "object"},
				},
			},
		},
	}

	assert.True(t, IsLocalReference("https://example.com/schemas/user", resources))
	assert.True(t, IsLocalReference("https://example.com/schemas/user#UserAnchor", resources))
	assert.False(t, IsLocalReference("https://example.com/schemas/user#Missing", resources))
	assert.False(t, IsLocalReference("://bad", resources))

	assert.True(t, ReferenceTargetExists("https://example.com/schemas/user", resources))
	assert.True(t, ReferenceTargetExists("https://example.com/schemas/user#UserAnchor", resources))
	assert.True(t, ReferenceTargetExists("https://example.com/schemas/user#/properties/name", resources))
	assert.False(t, ReferenceTargetExists("https://example.com/schemas/user#/properties/missing", resources))
	assert.False(t, ReferenceTargetExists("https://example.com/schemas/user#Missing", resources))
	assert.False(t, ReferenceTargetExists("://bad", resources))
}

func TestMarshalAny(t *testing.T) {
	assert.Equal(t, map[string]any{"x": float64(1)}, MarshalAny(map[string]int{"x": 1}))
	assert.Nil(t, MarshalAny(func() {}))
}
