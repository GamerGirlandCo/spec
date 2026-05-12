package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/oaswrap/spec/internal/builder"
	spec_reflect "github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type generator struct {
	cfg    *openapi.Config
	prefix string
	groups []*generator
	routes []*route
	opts   []option.GroupOption
	state  *sharedState
}

var _ Generator = (*generator)(nil)

type sharedState struct {
	mu      sync.Mutex
	root    *generator
	doc     *openapi.Document
	builder *builder.Builder
	errs    []error
	dirty   bool
}

// NewRouter creates a new OpenAPI generator.
// It is an alias of NewGenerator for naming compatibility.
//
// Example:
//
//	g := NewRouter(
//		option.WithTitle("My API"),
//		option.WithVersion("1.0.0"),
//		option.WithDescription("This is my API"),
//		option.WithServer("https://api.example.com"),
//	)
//	g.Get("/users", option.Summary("Get all users"))
//	g.Post("/users", option.Summary("Create a new user"))
//	schema, err := g.GenerateSchema("yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(schema))
func NewRouter(opts ...option.OpenAPIOption) Generator {
	return NewGenerator(opts...)
}

// NewGenerator creates a new OpenAPI generator with the provided options.
//
// Example:
//
//	g := NewGenerator(
//		option.WithTitle("My API"),
//		option.WithVersion("1.0.0"),
//		option.WithDescription("This is my API"),
//		option.WithServer("https://api.example.com"),
//	)
//	g.Get("/users", option.Summary("Get all users"))
//	g.Post("/users", option.Summary("Create a new user"))
//	schema, err := g.GenerateSchema("yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(schema))
func NewGenerator(opts ...option.OpenAPIOption) Generator {
	cfg := option.WithOpenAPIConfig(opts...)
	g := &generator{
		cfg: cfg,
	}
	g.state = &sharedState{root: g, dirty: true}
	return g
}

func (g *generator) Config() *openapi.Config {
	return g.cfg
}

func (g *generator) Document() *openapi.Document {
	g.build()
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	return g.state.doc
}

func (g *generator) Get(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodGet, path, opts...)
}

func (g *generator) Post(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodPost, path, opts...)
}

func (g *generator) Put(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodPut, path, opts...)
}

func (g *generator) Delete(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodDelete, path, opts...)
}

func (g *generator) Patch(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodPatch, path, opts...)
}

func (g *generator) Options(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodOptions, path, opts...)
}

func (g *generator) Head(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodHead, path, opts...)
}

func (g *generator) Trace(path string, opts ...option.OperationOption) Route {
	return g.Add(http.MethodTrace, path, opts...)
}

func (g *generator) Query(path string, opts ...option.OperationOption) Route {
	return g.Add("QUERY", path, opts...)
}

func (g *generator) Add(method, path string, opts ...option.OperationOption) Route {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	path = joinPath(g.prefix, path)
	r := &route{prefix: g.prefix, method: method, path: path, opts: opts, isWebhook: false, state: g.state}
	g.routes = append(g.routes, r)
	g.state.dirty = true
	return r
}

func (g *generator) Webhook(name string, opts ...option.OperationOption) Route {
	return g.AddWebhook(http.MethodPost, name, opts...)
}

func (g *generator) AddWebhook(method, name string, opts ...option.OperationOption) Route {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	r := &route{prefix: g.prefix, method: method, path: name, opts: opts, isWebhook: true, state: g.state}
	g.routes = append(g.routes, r)
	g.state.dirty = true
	return r
}

func (g *generator) NewRoute(opts ...option.OperationOption) Route {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	r := &route{prefix: g.prefix, opts: opts, state: g.state}
	g.routes = append(g.routes, r)
	g.state.dirty = true
	return r
}

func (g *generator) Route(pattern string, fn func(router Router), opts ...option.GroupOption) Router {
	group := g.Group(pattern, opts...)
	fn(group)
	return group
}

func (g *generator) Group(pattern string, opts ...option.GroupOption) Router {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	group := &generator{
		cfg:    g.cfg,
		prefix: joinPath(g.prefix, pattern),
		opts:   opts,
		state:  g.state,
	}
	g.groups = append(g.groups, group)
	g.state.dirty = true
	return group
}

func (g *generator) With(opts ...option.GroupOption) Router {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	g.opts = append(g.opts, opts...)
	g.state.dirty = true
	return g
}

func (g *generator) GenerateSchema(formats ...string) ([]byte, error) {
	format := optional("yaml", formats...)
	if !slices.Contains([]string{"json", "yaml", "yml"}, format) {
		return nil, fmt.Errorf("unsupported format: %s, expected one of json, yaml, yml", format)
	}
	if format == "json" {
		return g.MarshalJSON()
	}
	return g.MarshalYAML()
}

func (g *generator) MarshalYAML() ([]byte, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	g.state.mu.Lock()
	doc := g.state.doc
	g.state.mu.Unlock()
	return openapi.MarshalYAML(doc)
}

func (g *generator) MarshalJSON() ([]byte, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	g.state.mu.Lock()
	doc := g.state.doc
	g.state.mu.Unlock()
	raw, err := openapi.MarshalJSON(doc)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err = json.Indent(&out, raw, "", "  "); err != nil {
		return nil, err
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func (g *generator) WriteSchemaTo(path string) error {
	format := "yaml"
	if strings.HasSuffix(path, ".json") {
		format = "json"
	} else if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		return fmt.Errorf("unsupported file extension: %s, expected .json, .yaml, or .yml", path)
	}
	schema, err := g.GenerateSchema(format)
	if err != nil {
		return err
	}
	return os.WriteFile(path, schema, 0o600)
}

func (g *generator) Validate() error {
	g.build()
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	return joinErrors(append([]error(nil), g.state.errs...))
}

func (g *generator) ValidateReport() error {
	g.build()
	g.state.mu.Lock()
	defer g.state.mu.Unlock()
	return joinAllErrors(append([]error(nil), g.state.errs...))
}

func (g *generator) build() {
	g.state.mu.Lock()
	defer g.state.mu.Unlock()

	if !g.state.dirty {
		g.cfg.Logger.Debug("skip build: document not dirty")
		return
	}

	g.cfg.Logger.Debug("start build", "openapi_version", g.cfg.OpenAPIVersion)
	g.state.errs = nil
	g.state.doc = newDocument(g.cfg)
	g.state.builder = builder.NewBuilder(g.cfg, g.state.doc)

	if !supportedVersion(g.cfg.OpenAPIVersion) {
		g.state.errs = append(g.state.errs, fmt.Errorf("unsupported OpenAPI version: %s", g.cfg.OpenAPIVersion))
		return
	}
	routes := g.state.root.collectRoutes(nil)
	for _, item := range routes {
		if item.method == "" || item.path == "" {
			continue
		}
		if item.isWebhook {
			if err := g.state.builder.AddWebhookOperation(item.method, item.path, item.opts); err != nil {
				g.state.errs = append(g.state.errs, err)
			}
			continue
		}
		path := item.path
		if g.cfg.PathParser != nil {
			parsed, err := g.cfg.PathParser.Parse(path)
			if err != nil {
				g.state.errs = append(g.state.errs, fmt.Errorf("failed to parse path %q: %w", path, err))
				continue
			}
			path = parsed
		}
		if g.cfg.StripTrailingSlash && len(path) > 1 {
			path = strings.TrimRight(path, "/")
		}
		if err := g.state.builder.AddOperation(item.method, path, item.opts); err != nil {
			g.state.errs = append(g.state.errs, err)
		}
	}
	g.state.builder.Finish()
	if g.state.doc.Components == nil {
		g.state.doc.Components = &openapi.Components{}
	}
	for _, customize := range g.cfg.DocumentCustomizers {
		customize(g.state.doc)
	}
	if builder.ComponentsEmpty(g.state.doc.Components) {
		g.state.doc.Components = nil
	}
	g.state.errs = append(g.state.errs, validate.ValidateDocument(g.state.doc, g.cfg.OpenAPIVersion)...)
	g.state.dirty = false
	if len(g.state.errs) > 0 {
		g.cfg.Logger.Warn("finish build", "routes", len(routes), "errors", len(g.state.errs))
	} else {
		g.cfg.Logger.Debug("finish build", "routes", len(routes), "errors", 0)
	}
}

type routeItem struct {
	method    string
	path      string
	opts      []option.OperationOption
	isWebhook bool
}

func (g *generator) collectRoutes(parent []option.GroupOption) []routeItem {
	groupOpts := append(append([]option.GroupOption{}, parent...), g.opts...)
	var result []routeItem
	opOpts, hidden := groupOperationOptions(groupOpts)
	if !hidden {
		for _, r := range g.routes {
			opts := append([]option.OperationOption{}, opOpts...)
			opts = append(opts, r.opts...)
			result = append(result, routeItem{method: r.method, path: r.path, opts: opts, isWebhook: r.isWebhook})
		}
	}
	for _, group := range g.groups {
		result = append(result, group.collectRoutes(groupOpts)...)
	}
	return result
}

func groupOperationOptions(opts []option.GroupOption) ([]option.OperationOption, bool) {
	cfg := &option.GroupConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Hide {
		return nil, true
	}
	var out []option.OperationOption
	if cfg.Deprecated {
		out = append(out, option.Deprecated())
	}
	if len(cfg.Tags) > 0 {
		out = append(out, option.Tags(cfg.Tags...))
	}
	for _, sec := range cfg.Security {
		out = append(out, option.Security(sec.Name, sec.Scopes...))
	}
	return out, false
}

type route struct {
	prefix    string
	method    string
	path      string
	opts      []option.OperationOption
	isWebhook bool
	state     *sharedState
}

func (r *route) Method(method string) Route {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	r.method = method
	r.state.dirty = true
	return r
}

func (r *route) Path(path string) Route {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	if !r.isWebhook && r.prefix != "" {
		path = joinPath(r.prefix, path)
	}
	r.path = path
	r.state.dirty = true
	return r
}

func (r *route) With(opts ...option.OperationOption) Route {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	r.opts = append(r.opts, opts...)
	r.state.dirty = true
	return r
}

func newDocument(cfg *openapi.Config) *openapi.Document {
	doc := &openapi.Document{
		OpenAPI: cfg.OpenAPIVersion,
		Self:    cfg.Self,
		Info: openapi.Info{
			Title:          cfg.Title,
			Summary:        cfg.InfoSummary,
			Description:    cfg.Description,
			TermsOfService: cfg.TermsOfService,
			Contact:        cfg.Contact,
			License:        cfg.License,
			Version:        cfg.Version,
		},
		Servers:      cfg.Servers,
		Paths:        map[string]*openapi.PathItem{},
		Security:     cfg.Security,
		Tags:         cfg.Tags,
		ExternalDocs: cfg.ExternalDocs,
	}
	if cfg.JSONSchemaDialect != "" {
		doc.JSONSchemaDialect = cfg.JSONSchemaDialect
	}
	if len(cfg.SecuritySchemes) > 0 {
		doc.Components = &openapi.Components{SecuritySchemes: cfg.SecuritySchemes}
	}
	return doc
}

func supportedVersion(version string) bool {
	return spec_reflect.IsOpenAPI30(version) || validate.IsOpenAPI31(version) || validate.IsOpenAPI32(version)
}

func joinPath(prefix, path string) string {
	if prefix == "" {
		if path == "" {
			return "/"
		}
		return ensureLeadingSlash(path)
	}
	if path == "" || path == "/" {
		return ensureLeadingSlash(prefix)
	}
	return strings.TrimRight(ensureLeadingSlash(prefix), "/") + "/" + strings.TrimLeft(path, "/")
}

func ensureLeadingSlash(path string) string {
	if path == "" {
		return "/"
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func optional[T any](fallback T, values ...T) T {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}
