package builder

import (
	"fmt"
	"strings"

	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type Builder struct {
	Config    *openapi.Config
	Doc       *openapi.Document
	Reflector *reflect.Reflector
}

func NewBuilder(cfg *openapi.Config, doc *openapi.Document) *Builder {
	return &Builder{
		Config:    cfg,
		Doc:       doc,
		Reflector: reflect.NewReflector(cfg),
	}
}

func (b *Builder) AddOperation(method, path string, opts []option.OperationOption) error {
	return b.AddOperationTo(method, path, opts, b.Doc.Paths)
}

func (b *Builder) AddWebhookOperation(method, name string, opts []option.OperationOption) error {
	if reflect.IsOpenAPI30(b.Config.OpenAPIVersion) {
		return fmt.Errorf("webhooks require OpenAPI 3.1.x or 3.2.0")
	}
	if b.Doc.Webhooks == nil {
		b.Doc.Webhooks = map[string]*openapi.PathItem{}
	}
	return b.AddOperationTo(method, name, opts, b.Doc.Webhooks)
}

func (b *Builder) AddOperationTo(
	method, target string,
	opts []option.OperationOption,
	items map[string]*openapi.PathItem,
) error {
	cfg := &option.OperationConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Hide {
		return nil
	}

	method = strings.ToUpper(method)
	if method == "QUERY" && b.Config.OpenAPIVersion != openapi.Version320 {
		return fmt.Errorf("method QUERY requires OpenAPI 3.2.0")
	}

	op := &openapi.Operation{Responses: map[string]*openapi.Response{}}
	op.OperationID = cfg.OperationID
	op.Summary = cfg.Summary
	op.Description = cfg.Description
	op.ExternalDocs = cfg.ExternalDocs
	op.Deprecated = cfg.Deprecated
	op.Tags = append(op.Tags, cfg.Tags...)
	for _, sec := range cfg.Security {
		op.Security = append(op.Security, SecurityRequirement(sec.Name, sec.Scopes))
	}

	for _, req := range cfg.Requests {
		b.AddRequest(op, req)
	}
	if len(cfg.Responses) == 0 {
		op.Responses["default"] = &openapi.Response{Description: "Default response"}
	}
	for _, resp := range MergeResponses(cfg.Responses) {
		if err := b.AddResponse(op, resp); err != nil {
			return fmt.Errorf("%s %s response: %w", method, target, err)
		}
	}
	for _, customize := range cfg.Customizers {
		customize(op)
	}

	item := items[target]
	if item == nil {
		item = &openapi.PathItem{}
		items[target] = item
	}
	return SetOperation(item, method, op, b.Config.OpenAPIVersion)
}

func (b *Builder) Finish() {
	if b.Doc.Components == nil {
		b.Doc.Components = &openapi.Components{}
	}
	if len(b.Reflector.Components) > 0 {
		if b.Doc.Components.Schemas == nil {
			b.Doc.Components.Schemas = map[string]*openapi.Schema{}
		}
		for name, schema := range b.Reflector.Components {
			b.Doc.Components.Schemas[name] = schema
		}
	}
	if ComponentsEmpty(b.Doc.Components) {
		b.Doc.Components = nil
	}
}

func ComponentsEmpty(components *openapi.Components) bool {
	if components == nil {
		return true
	}
	return len(components.Schemas) == 0 &&
		len(components.SecuritySchemes) == 0 &&
		len(components.Responses) == 0 &&
		len(components.Parameters) == 0 &&
		len(components.Examples) == 0 &&
		len(components.RequestBodies) == 0 &&
		len(components.Headers) == 0 &&
		len(components.Links) == 0 &&
		len(components.Callbacks) == 0 &&
		len(components.PathItems) == 0 &&
		len(components.MediaTypes) == 0
}

func SecurityRequirement(name string, scopes []string) openapi.SecurityRequirement {
	if scopes == nil {
		scopes = []string{}
	}
	return openapi.SecurityRequirement{name: scopes}
}
