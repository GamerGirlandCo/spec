package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestBuilder_AddOperation(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	doc := &openapi.Document{Paths: map[string]*openapi.PathItem{}}
	b := NewBuilder(cfg, doc)

	err := b.AddOperation("GET", "/test", []option.OperationOption{
		option.Summary("Test Summary"),
	})
	require.NoError(t, err)

	assert.NotNil(t, doc.Paths["/test"])
	assert.NotNil(t, doc.Paths["/test"].Get)
	assert.Equal(t, "Test Summary", doc.Paths["/test"].Get.Summary)
}

func TestBuilder_AddWebhookOperation(t *testing.T) {
	t.Run("OpenAPI 3.0.4", func(t *testing.T) {
		cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
		doc := &openapi.Document{}
		b := NewBuilder(cfg, doc)

		err := b.AddWebhookOperation("POST", "webhook", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "webhooks require OpenAPI 3.1.x or 3.2.0")
	})

	t.Run("OpenAPI 3.1.2", func(t *testing.T) {
		cfg := &openapi.Config{OpenAPIVersion: openapi.Version312}
		doc := &openapi.Document{}
		b := NewBuilder(cfg, doc)

		err := b.AddWebhookOperation("POST", "webhook", nil)
		require.NoError(t, err)
		assert.NotNil(t, doc.Webhooks["webhook"])
	})
}

func TestBuilder_Finish(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	doc := &openapi.Document{}
	b := NewBuilder(cfg, doc)

	// Simulate reflector having components
	b.Reflector.Components["User"] = &openapi.Schema{Type: "object"}

	b.Finish()

	assert.NotNil(t, doc.Components)
	assert.NotNil(t, doc.Components.Schemas["User"])
}

func TestBuilder_ComponentsEmpty(t *testing.T) {
	assert.True(t, ComponentsEmpty(nil))
	assert.True(t, ComponentsEmpty(&openapi.Components{}))
	assert.False(t, ComponentsEmpty(&openapi.Components{Schemas: map[string]*openapi.Schema{"S": {}}}))
}

func TestBuilder_SecurityRequirement(t *testing.T) {
	sr := SecurityRequirement("auth", []string{"read"})
	assert.Equal(t, []string{"read"}, sr["auth"])

	sr = SecurityRequirement("auth", nil)
	assert.Equal(t, []string{}, sr["auth"])
}
