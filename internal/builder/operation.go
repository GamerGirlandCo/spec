package builder

import (
	"strconv"

	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
)

func (b *Builder) AddRequest(op *openapi.Operation, cu *openapi.ContentUnit) {
	params, body := b.Reflector.RequestParts(cu.Structure, ContentType(cu))
	op.Parameters = append(op.Parameters, params...)

	ct := ContentType(cu)
	if body == nil {
		isDefaultJSON := ct == "application/json" || cu.ContentType == ""
		if isDefaultJSON && cu.Format == "" && cu.Example == nil && len(cu.Examples) == 0 {
			return
		}
	}

	if op.RequestBody == nil {
		op.RequestBody = &openapi.RequestBody{Content: map[string]openapi.MediaType{}}
	}
	if cu.Description != "" {
		op.RequestBody.Description = cu.Description
	}
	if cu.Required {
		op.RequestBody.Required = true
	}
	if body == nil && cu.Structure == nil {
		body = &openapi.Schema{Type: "string"}
	}
	if body != nil && cu.Format != "" {
		body.Format = cu.Format
	}
	mt := openapi.MediaType{Schema: body}
	ApplyContentExamples(&mt, cu)
	if len(cu.Encoding) > 0 {
		mt.Encoding = map[string]*openapi.Encoding{}
		for prop, enc := range cu.Encoding {
			mt.Encoding[prop] = &openapi.Encoding{ContentType: enc}
		}
	}
	op.RequestBody.Content[ct] = mt
}

func (b *Builder) AddResponse(op *openapi.Operation, cu *openapi.ContentUnit) error {
	key := strconv.Itoa(cu.HTTPStatus)
	if cu.IsDefault {
		key = "default"
	} else if cu.HTTPStatus == 0 {
		return validate.Errorf("HTTP status is required unless ContentDefault is set")
	}

	response := op.Responses[key]
	if response == nil {
		response = &openapi.Response{Description: ResponseDescription(cu)}
		op.Responses[key] = response
	}
	if cu.Summary != "" && b.Config.OpenAPIVersion == openapi.Version320 {
		response.Summary = cu.Summary
	}
	if response.Content == nil {
		response.Content = map[string]openapi.MediaType{}
	}

	ct := ContentType(cu)
	if cu.Structure != nil || cu.ContentType != "" || cu.Example != nil || len(cu.Examples) > 0 {
		schema := b.Reflector.SchemaForValue(cu.Structure, reflect.SchemaUseComponent)
		if schema == nil && cu.ContentType != "" {
			schema = &openapi.Schema{Type: "string"}
		}
		if schema != nil && cu.Format != "" {
			schema.Format = cu.Format
		}
		mt := openapi.MediaType{
			Schema: schema,
		}
		ApplyContentExamples(&mt, cu)
		response.Content[ct] = mt
	}
	return nil
}

func ApplyContentExamples(mediaType *openapi.MediaType, cu *openapi.ContentUnit) {
	if cu.Example != nil {
		mediaType.Example = cu.Example
	}
	if len(cu.Examples) > 0 {
		mediaType.Examples = cu.Examples
	}
}
