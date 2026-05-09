package option

import (
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/config"
	"github.com/oaswrap/spec-ui/rapidoc"
	"github.com/oaswrap/spec-ui/redoc"
	"github.com/oaswrap/spec-ui/scalar"
	"github.com/oaswrap/spec-ui/stoplight"
	"github.com/oaswrap/spec-ui/swaggerui"
	"github.com/oaswrap/spec/openapi"
)

// OpenAPIOption mutates root generator configuration.
type OpenAPIOption func(*openapi.Config)

// WithOpenAPIConfig builds config with defaults and applies options in order.
func WithOpenAPIConfig(opts ...OpenAPIOption) *openapi.Config {
	cfg := &openapi.Config{
		OpenAPIVersion: openapi.Version304,
		Title:          "API Documentation",
		Version:        "1.0.0",
		DocsPath:       "/docs",
		SpecPath:       "/docs/openapi.yaml",
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithOpenAPIVersion sets the OpenAPI document version.
func WithOpenAPIVersion(version string) OpenAPIOption {
	return func(c *openapi.Config) { c.OpenAPIVersion = version }
}

// WithSelf sets the OpenAPI 3.2 `$self` URI reference.
func WithSelf(self string) OpenAPIOption {
	return func(c *openapi.Config) { c.Self = self }
}

// WithJSONSchemaDialect sets the root `jsonSchemaDialect`.
func WithJSONSchemaDialect(uri string) OpenAPIOption {
	return func(c *openapi.Config) { c.JSONSchemaDialect = uri }
}

// WithTitle sets `info.title`.
func WithTitle(title string) OpenAPIOption {
	return func(c *openapi.Config) { c.Title = title }
}

// WithInfoSummary sets `info.summary`.
func WithInfoSummary(summary string) OpenAPIOption {
	return func(c *openapi.Config) { c.InfoSummary = summary }
}

// WithVersion sets `info.version`.
func WithVersion(version string) OpenAPIOption {
	return func(c *openapi.Config) { c.Version = version }
}

// WithDescription sets `info.description`.
func WithDescription(description string) OpenAPIOption {
	return func(c *openapi.Config) { c.Description = &description }
}

// WithContact sets `info.contact`.
func WithContact(contact openapi.Contact) OpenAPIOption {
	return func(c *openapi.Config) { c.Contact = &contact }
}

// WithLicense sets `info.license`.
func WithLicense(license openapi.License) OpenAPIOption {
	return func(c *openapi.Config) { c.License = &license }
}

// WithTermsOfService sets `info.termsOfService`.
func WithTermsOfService(terms string) OpenAPIOption {
	return func(c *openapi.Config) { c.TermsOfService = &terms }
}

// WithTags appends root-level tags.
func WithTags(tags ...openapi.Tag) OpenAPIOption {
	return func(c *openapi.Config) { c.Tags = append(c.Tags, tags...) }
}

// TagOption mutates a root-level tag.
type TagOption func(*openapi.Tag)

// WithTag appends one root-level tag.
func WithTag(name string, opts ...TagOption) OpenAPIOption {
	return func(c *openapi.Config) {
		tag := openapi.Tag{Name: name}
		for _, opt := range opts {
			opt(&tag)
		}
		c.Tags = append(c.Tags, tag)
	}
}

// TagSummary sets tag summary.
func TagSummary(summary string) TagOption {
	return func(tag *openapi.Tag) { tag.Summary = summary }
}

// TagDescription sets tag description.
func TagDescription(description string) TagOption {
	return func(tag *openapi.Tag) { tag.Description = description }
}

// TagExternalDocs sets tag external documentation.
func TagExternalDocs(url string, description ...string) TagOption {
	return func(tag *openapi.Tag) {
		docs := &openapi.ExternalDocs{URL: url}
		if len(description) > 0 {
			docs.Description = description[0]
		}
		tag.ExternalDocs = docs
	}
}

// TagParent sets the OpenAPI 3.2 tag parent.
func TagParent(parent string) TagOption {
	return func(tag *openapi.Tag) { tag.Parent = parent }
}

// TagKind sets the OpenAPI 3.2 tag kind.
func TagKind(kind string) TagOption {
	return func(tag *openapi.Tag) { tag.Kind = kind }
}

// WithServer appends a root server and applies server options.
func WithServer(url string, opts ...ServerOption) OpenAPIOption {
	return func(c *openapi.Config) {
		server := openapi.Server{URL: url}
		for _, opt := range opts {
			opt(&server)
		}
		c.Servers = append(c.Servers, server)
	}
}

// WithExternalDocs sets root `externalDocs`.
func WithExternalDocs(url string, description ...string) OpenAPIOption {
	return func(c *openapi.Config) {
		docs := &openapi.ExternalDocs{URL: url}
		if len(description) > 0 {
			docs.Description = description[0]
		}
		c.ExternalDocs = docs
	}
}

// WithSecurity registers a reusable named security scheme.
func WithSecurity(name string, opts ...SecurityOption) OpenAPIOption {
	return func(c *openapi.Config) {
		s := &securityConfig{}
		for _, opt := range opts {
			opt(s)
		}
		if c.SecuritySchemes == nil {
			c.SecuritySchemes = map[string]*openapi.SecurityScheme{}
		}
		scheme := s.scheme
		if scheme != nil {
			c.SecuritySchemes[name] = scheme
		}
	}
}

// WithGlobalSecurity appends root security requirements.
func WithGlobalSecurity(name string, scopes ...string) OpenAPIOption {
	return func(c *openapi.Config) {
		if scopes == nil {
			scopes = []string{}
		}
		c.Security = append(c.Security, openapi.SecurityRequirement{name: scopes})
	}
}

// WithReflectorConfig mutates schema reflection settings.
func WithReflectorConfig(opts ...ReflectorOption) OpenAPIOption {
	return func(c *openapi.Config) {
		if c.ReflectorConfig == nil {
			c.ReflectorConfig = &openapi.ReflectorConfig{}
		}
		for _, opt := range opts {
			opt(c.ReflectorConfig)
		}
	}
}

// WithStripTrailingSlash toggles trimming of trailing slashes in route paths.
func WithStripTrailingSlash(strip ...bool) OpenAPIOption {
	return func(c *openapi.Config) { c.StripTrailingSlash = optional(true, strip...) }
}

// WithPathParser sets a custom route path parser.
func WithPathParser(parser openapi.PathParser) OpenAPIOption {
	return func(c *openapi.Config) { c.PathParser = parser }
}

// WithDocument applies a low-level mutation after routes and reflected schemas
// have been added and before validation/serialization.
func WithDocument(fn func(*openapi.Document)) OpenAPIOption {
	return func(c *openapi.Config) {
		if fn != nil {
			c.DocumentCustomizers = append(c.DocumentCustomizers, fn)
		}
	}
}

// WithComponentSchema registers a reusable schema component.
func WithComponentSchema(name string, schema *openapi.Schema) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Schemas == nil {
			components.Schemas = map[string]*openapi.Schema{}
		}
		components.Schemas[name] = schema
	})
}

// WithComponentResponse registers a reusable response component.
func WithComponentResponse(name string, response *openapi.Response) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Responses == nil {
			components.Responses = map[string]*openapi.Response{}
		}
		components.Responses[name] = response
	})
}

// WithComponentParameter registers a reusable parameter component.
func WithComponentParameter(name string, parameter *openapi.Parameter) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Parameters == nil {
			components.Parameters = map[string]*openapi.Parameter{}
		}
		components.Parameters[name] = parameter
	})
}

// WithComponentExample registers a reusable example component.
func WithComponentExample(name string, example *openapi.Example) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Examples == nil {
			components.Examples = map[string]*openapi.Example{}
		}
		components.Examples[name] = example
	})
}

// WithComponentRequestBody registers a reusable request body component.
func WithComponentRequestBody(name string, requestBody *openapi.RequestBody) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.RequestBodies == nil {
			components.RequestBodies = map[string]*openapi.RequestBody{}
		}
		components.RequestBodies[name] = requestBody
	})
}

// WithComponentHeader registers a reusable header component.
func WithComponentHeader(name string, header *openapi.Header) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Headers == nil {
			components.Headers = map[string]*openapi.Header{}
		}
		components.Headers[name] = header
	})
}

// WithComponentSecurityScheme registers a reusable security scheme component.
func WithComponentSecurityScheme(name string, scheme *openapi.SecurityScheme) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.SecuritySchemes == nil {
			components.SecuritySchemes = map[string]*openapi.SecurityScheme{}
		}
		components.SecuritySchemes[name] = scheme
	})
}

// WithComponentLink registers a reusable link component.
func WithComponentLink(name string, link *openapi.Link) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Links == nil {
			components.Links = map[string]*openapi.Link{}
		}
		components.Links[name] = link
	})
}

// WithComponentCallback registers a reusable callback component.
func WithComponentCallback(name string, callback *openapi.Callback) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.Callbacks == nil {
			components.Callbacks = map[string]*openapi.Callback{}
		}
		components.Callbacks[name] = callback
	})
}

// WithComponentPathItem registers a reusable path item component.
func WithComponentPathItem(name string, pathItem *openapi.PathItem) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.PathItems == nil {
			components.PathItems = map[string]*openapi.PathItem{}
		}
		components.PathItems[name] = pathItem
	})
}

// WithComponentMediaType registers a reusable media type component.
func WithComponentMediaType(name string, mediaType *openapi.MediaType) OpenAPIOption {
	return WithDocument(func(doc *openapi.Document) {
		components := ensureComponents(doc)
		if components.MediaTypes == nil {
			components.MediaTypes = map[string]*openapi.MediaType{}
		}
		components.MediaTypes[name] = mediaType
	})
}

func ensureComponents(doc *openapi.Document) *openapi.Components {
	if doc.Components == nil {
		doc.Components = &openapi.Components{}
	}
	return doc.Components
}

func WithDisableDocs(disable ...bool) OpenAPIOption {
	return func(c *openapi.Config) {
		c.DisableDocs = optional(true, disable...)
	}
}

func WithDocsPath(path string) OpenAPIOption {
	return func(c *openapi.Config) {
		c.DocsPath = path
	}
}

// WithSpecPath sets the path for the OpenAPI specification.
//
// This is the path where the OpenAPI specification will be served.
// The default is "/docs/openapi.yaml".
func WithSpecPath(path string) OpenAPIOption {
	return func(c *openapi.Config) {
		c.SpecPath = path
	}
}

// WithCacheAge sets the cache age for OpenAPI specification responses.
func WithCacheAge(cacheAge int) OpenAPIOption {
	return func(c *openapi.Config) {
		c.CacheAge = &cacheAge
	}
}

// WithUIOption sets a custom spec-ui option.
//
// This enables consumers to import only the specific provider package they need,
// improving linker tree-shaking.
func WithUIOption(opt specui.Option) OpenAPIOption {
	return func(c *openapi.Config) {
		c.UIOption = opt
	}
}

// WithSwaggerUI sets the UI documentation to Swagger UI (CDN mode).
func WithSwaggerUI(cfg ...config.SwaggerUI) OpenAPIOption {
	return func(c *openapi.Config) {
		uiCfg := config.SwaggerUI{}
		if len(cfg) > 0 {
			uiCfg = cfg[0]
		}
		c.UIProvider = config.ProviderSwaggerUI
		c.SwaggerUIConfig = &uiCfg
		c.UIOption = swaggerui.WithUI(uiCfg)
	}
}

// WithStoplightElements sets the UI documentation to Stoplight Elements (CDN mode).
func WithStoplightElements(cfg ...config.StoplightElements) OpenAPIOption {
	return func(c *openapi.Config) {
		uiCfg := config.StoplightElements{}
		if len(cfg) > 0 {
			uiCfg = cfg[0]
		}
		c.UIProvider = config.ProviderStoplightElements
		c.StoplightElementsConfig = &uiCfg
		c.UIOption = stoplight.WithUI(uiCfg)
	}
}

// WithReDoc sets the UI documentation to ReDoc (CDN mode).
func WithReDoc(cfg ...config.ReDoc) OpenAPIOption {
	return func(c *openapi.Config) {
		uiCfg := config.ReDoc{}
		if len(cfg) > 0 {
			uiCfg = cfg[0]
		}
		c.UIProvider = config.ProviderReDoc
		c.ReDocConfig = &uiCfg
		c.UIOption = redoc.WithUI(uiCfg)
	}
}

// WithScalar sets the UI documentation to Scalar (CDN mode).
func WithScalar(cfg ...config.Scalar) OpenAPIOption {
	return func(c *openapi.Config) {
		uiCfg := config.Scalar{}
		if len(cfg) > 0 {
			uiCfg = cfg[0]
		}
		c.UIProvider = config.ProviderScalar
		c.ScalarConfig = &uiCfg
		c.UIOption = scalar.WithUI(uiCfg)
	}
}

// WithRapiDoc sets the UI documentation to RapiDoc (CDN mode).
func WithRapiDoc(cfg ...config.RapiDoc) OpenAPIOption {
	return func(c *openapi.Config) {
		uiCfg := config.RapiDoc{}
		if len(cfg) > 0 {
			uiCfg = cfg[0]
		}
		c.UIProvider = config.ProviderRapiDoc
		c.RapiDocConfig = &uiCfg
		c.UIOption = rapidoc.WithUI(uiCfg)
	}
}
