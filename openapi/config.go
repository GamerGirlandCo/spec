package openapi

import (
	"reflect"

	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/config"
)

const (
	// Version300 is OpenAPI 3.0.0.
	Version300 = "3.0.0"
	// Version301 is OpenAPI 3.0.1.
	Version301 = "3.0.1"
	// Version302 is OpenAPI 3.0.2.
	Version302 = "3.0.2"
	// Version303 is OpenAPI 3.0.3.
	Version303 = "3.0.3"
	// Version304 is OpenAPI 3.0.4.
	Version304 = "3.0.4"
	// Version310 is OpenAPI 3.1.0.
	Version310 = "3.1.0"
	// Version311 is OpenAPI 3.1.1.
	Version311 = "3.1.1"
	// Version312 is OpenAPI 3.1.2.
	Version312 = "3.1.2"
	// Version320 is OpenAPI 3.2.0.
	Version320 = "3.2.0"
)

// Config is the main configuration struct for generating an OpenAPI document.
// It contains all the necessary information and options to customize the generated document, including metadata,
// server information, security schemes, and UI configuration.
type Config struct {
	OpenAPIVersion    string
	Self              string
	Title             string
	InfoSummary       string
	Version           string
	Description       *string
	JSONSchemaDialect string
	Contact           *Contact
	License           *License
	TermsOfService    *string
	Servers           []Server
	SecuritySchemes   map[string]*SecurityScheme
	Security          []SecurityRequirement
	Tags              []Tag
	ExternalDocs      *ExternalDocs

	ReflectorConfig     *ReflectorConfig
	StripTrailingSlash  bool
	PathParser          PathParser
	DocumentCustomizers []func(*Document)

	DocsPath    string
	SpecPath    string
	CacheAge    *int
	DisableDocs bool

	UIProvider              config.Provider
	SwaggerUIConfig         *config.SwaggerUI
	StoplightElementsConfig *config.StoplightElements
	ReDocConfig             *config.ReDoc
	ScalarConfig            *config.Scalar
	RapiDocConfig           *config.RapiDoc
	UIOption                specui.Option
}

// ReflectorConfig contains configuration options for the reflection process used to generate
// the OpenAPI document from Go types. It allows customization of how types are reflected, including inline references,
// stripping of definition name prefixes, and custom type mappings.
type ReflectorConfig struct {
	InlineRefs          bool
	StripDefNamePrefix  []string
	InterceptDefName    func(t reflect.Type, defaultDefName string) string
	DefNameCallerPkg    string
	TypeMappings        []TypeMapping
	ParameterTagMapping map[ParameterIn]string
}

// TypeMapping represents a mapping between a source type and a destination type. It is used in the reflection process
// to specify how certain types should be mapped when generating the OpenAPI document. The Src field represents
// the original type, while the Dst field represents the type that should be used in the generated document.
type TypeMapping struct {
	Src any
	Dst any
}

// PathParser is an interface that defines a method for parsing a path string and returning a modified version of it.
// This can be used to customize how paths are represented in the generated OpenAPI document, allowing
// for transformations such as converting path parameters to a specific format or applying
// any other necessary modifications to the path strings.
type PathParser interface {
	Parse(path string) (string, error)
}
