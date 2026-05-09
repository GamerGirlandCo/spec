package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec/openapi"
)

func TestBuilder_AddRequest(t *testing.T) {
	t.Run("Path Parameter", func(t *testing.T) {
		cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
		doc := &openapi.Document{}
		b := NewBuilder(cfg, doc)
		op := &openapi.Operation{}

		type Request struct {
			ID string `path:"id"`
		}
		cu := &openapi.ContentUnit{Structure: Request{}}
		b.AddRequest(op, cu)

		assert.Len(t, op.Parameters, 1)
		assert.Equal(t, "id", op.Parameters[0].Name)
	})

	t.Run("Body and Description", func(t *testing.T) {
		cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
		doc := &openapi.Document{}
		b := NewBuilder(cfg, doc)
		op := &openapi.Operation{}

		cu := &openapi.ContentUnit{
			Structure:   map[string]string{"foo": "bar"},
			Description: "Body description",
			Required:    true,
			Format:      "custom",
			Encoding:    map[string]string{"foo": "text/plain"},
		}
		b.AddRequest(op, cu)

		require.NotNil(t, op.RequestBody)
		assert.Equal(t, "Body description", op.RequestBody.Description)
		assert.True(t, op.RequestBody.Required)
		mt := op.RequestBody.Content["application/json"]
		assert.Equal(t, "custom", mt.Schema.Format)
		assert.Equal(t, "text/plain", mt.Encoding["foo"].ContentType)
	})

	t.Run("Default String Body", func(t *testing.T) {
		cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
		doc := &openapi.Document{}
		b := NewBuilder(cfg, doc)
		op := &openapi.Operation{}

		cu := &openapi.ContentUnit{ContentType: "text/plain"}
		b.AddRequest(op, cu)

		require.NotNil(t, op.RequestBody)
		assert.Equal(t, "string", op.RequestBody.Content["text/plain"].Schema.Type)
	})
}

func TestBuilder_AddResponse(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	doc := &openapi.Document{}
	b := NewBuilder(cfg, doc)
	op := &openapi.Operation{Responses: map[string]*openapi.Response{}}

	cu := &openapi.ContentUnit{
		HTTPStatus: 200,
		Structure:  map[string]string{"foo": "bar"},
	}

	err := b.AddResponse(op, cu)
	require.NoError(t, err)

	assert.NotNil(t, op.Responses["200"])
	assert.NotNil(t, op.Responses["200"].Content["application/json"])
}

func TestApplyContentExamples(t *testing.T) {
	mt := &openapi.MediaType{}
	cu := &openapi.ContentUnit{
		Example:  "ex",
		Examples: map[string]*openapi.Example{"e1": {Value: "v1"}},
	}

	ApplyContentExamples(mt, cu)

	assert.Equal(t, "ex", mt.Example)
	assert.Equal(t, "v1", mt.Examples["e1"].Value)
}
