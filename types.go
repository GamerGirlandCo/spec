package spec

import (
	"github.com/oaswrap/spec/internal/builder"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

// Contact is an alias of openapi.Contact.
type Contact = openapi.Contact

// License is an alias of openapi.License.
type License = openapi.License

// Tag is an alias of openapi.Tag.
type Tag = openapi.Tag

// ExternalDocs is an alias of openapi.ExternalDocs.
type ExternalDocs = openapi.ExternalDocs

// Server is an alias of openapi.Server.
type Server = openapi.Server

// ServerVariable is an alias of openapi.ServerVariable.
type ServerVariable = openapi.ServerVariable

// SecurityScheme is an alias of openapi.SecurityScheme.
type SecurityScheme = openapi.SecurityScheme

// OAuthFlows is an alias of openapi.OAuthFlows.
type OAuthFlows = openapi.OAuthFlows

// OAuthFlow is an alias of openapi.OAuthFlow.
type OAuthFlow = openapi.OAuthFlow

// Schema is an alias of openapi.Schema.
type Schema = openapi.Schema

// Document is an alias of openapi.Document.
type Document = openapi.Document

// OneOf returns a value that represents multiple possible schemas.
func OneOf(values ...any) any {
	return builder.OneOf(values...)
}

// SchemaExposer lets custom Go types provide their own OpenAPI Schema Object.
// The selected OpenAPI version is passed so implementations can return
// version-specific schema keywords.
type SchemaExposer interface {
	OpenAPISchema(version string) *openapi.Schema
}

// StaticSchemaExposer lets custom Go types provide an OpenAPI Schema Object
// when they do not need version-specific output.
type StaticSchemaExposer interface {
	OpenAPISchema() *openapi.Schema
}

// Generator builds, validates, and serializes an OpenAPI document.
type Generator interface {
	Router
	// Config returns the effective OpenAPI configuration.
	Config() *openapi.Config
	// Document returns the built in-memory OpenAPI document.
	Document() *openapi.Document
	// GenerateSchema serializes the document in YAML by default, or JSON/YAML
	// when explicitly requested.
	GenerateSchema(formats ...string) ([]byte, error)
	// MarshalYAML validates and serializes the document as YAML.
	MarshalYAML() ([]byte, error)
	// MarshalJSON validates and serializes the document as pretty JSON.
	MarshalJSON() ([]byte, error)
	// Validate builds the document and validates OpenAPI invariants.
	Validate() error
	// WriteSchemaTo serializes and writes the document based on file extension.
	WriteSchemaTo(path string) error
}

// Router registers operations and route groups that are converted to OpenAPI
// Paths and Webhooks during document generation.
type Router interface {
	// Get registers a GET operation for a path.
	Get(path string, opts ...option.OperationOption) Route
	// Post registers a POST operation for a path.
	Post(path string, opts ...option.OperationOption) Route
	// Put registers a PUT operation for a path.
	Put(path string, opts ...option.OperationOption) Route
	// Delete registers a DELETE operation for a path.
	Delete(path string, opts ...option.OperationOption) Route
	// Patch registers a PATCH operation for a path.
	Patch(path string, opts ...option.OperationOption) Route
	// Options registers an OPTIONS operation for a path.
	Options(path string, opts ...option.OperationOption) Route
	// Head registers a HEAD operation for a path.
	Head(path string, opts ...option.OperationOption) Route
	// Trace registers a TRACE operation for a path.
	Trace(path string, opts ...option.OperationOption) Route
	// Query registers an OpenAPI 3.2 QUERY operation for a path.
	Query(path string, opts ...option.OperationOption) Route
	// Add registers an operation for an arbitrary HTTP method.
	Add(method, path string, opts ...option.OperationOption) Route
	// Webhook registers a webhook entry using POST as the default method.
	// Webhooks require OpenAPI 3.1.x or 3.2.0.
	Webhook(name string, opts ...option.OperationOption) Route
	// AddWebhook registers a webhook entry with an explicit HTTP method.
	// Webhooks require OpenAPI 3.1.x or 3.2.0.
	AddWebhook(method, name string, opts ...option.OperationOption) Route
	// NewRoute creates a route that can be configured incrementally.
	NewRoute(opts ...option.OperationOption) Route
	// Route creates a grouped router and executes registrations in fn.
	Route(pattern string, fn func(router Router), opts ...option.GroupOption) Router
	// Group creates a grouped router with a path prefix and group options.
	Group(pattern string, opts ...option.GroupOption) Router
	// With appends group options to the current router scope.
	With(opts ...option.GroupOption) Router
}

// Route lets callers set method/path/options incrementally.
type Route interface {
	// Method sets the HTTP method.
	Method(method string) Route
	// Path sets the route path or webhook name.
	Path(path string) Route
	// With appends operation options.
	With(opts ...option.OperationOption) Route
}
