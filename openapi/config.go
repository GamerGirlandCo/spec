package openapi

import (
	"reflect"

	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/config"
)

const (
	Version300 = "3.0.0"
	Version301 = "3.0.1"
	Version302 = "3.0.2"
	Version303 = "3.0.3"
	Version304 = "3.0.4"
	Version310 = "3.1.0"
	Version311 = "3.1.1"
	Version312 = "3.1.2"
	Version320 = "3.2.0"
)

// Config is the main configuration struct for generating an OpenAPI document.
// It contains all the necessary information and options to customize the generated document, including metadata,
// server information, security schemes, and UI configuration.
type Config struct {
	OpenAPIVersion    string                     // The version of the OpenAPI Specification to use for the generated document (e.g., "3.0.0", "3.1.0").
	Self              string                     // The URL or path where the generated OpenAPI document will be served. This is used to set the "servers" field in the OpenAPI document and can be a relative path (e.g., "/openapi.json") or an absolute URL (e.g., "https://api.example.com/openapi.json").
	Title             string                     // The title of the API, which will be displayed in the generated OpenAPI document and UI.
	InfoSummary       string                     // A brief summary of the API, providing a high-level overview of its purpose and functionality.
	Version           string                     // The version of the API, which will be included in the generated OpenAPI document and can be used for versioning and documentation purposes.
	Description       *string                    // A detailed description of the API, providing more information about its features, usage, and any other relevant details. This can be a longer text that gives users a better understanding of the API.
	JSONSchemaDialect string                     // The JSON Schema dialect to use for the generated OpenAPI document. This specifies the version of JSON Schema that should be used for validating the schema definitions in the document (e.g., "http://json-schema.org/draft-07/schema#", "https://json-schema.org/draft/2020-12/schema#").
	Contact           *Contact                   // Contact information for the API, including name, URL, and email. This can be used to provide users with a way to get in touch with the API maintainers or support team.
	License           *License                   // License information for the API, including name and URL. This can be used to specify the terms under which the API is licensed and any restrictions or permissions associated with its use.
	TermsOfService    *string                    // The terms of service for the API, which can include any legal or usage terms that users must agree to when using the API. This can be a URL pointing to the full terms of service document or a brief summary of the terms.
	Servers           []Server                   // A list of server objects that specify the base URLs for the API. This can include multiple servers for different environments (e.g., production, staging) or different regions.
	SecuritySchemes   map[string]*SecurityScheme // A map of security scheme names to their corresponding security scheme objects. This defines the various authentication and authorization methods that can be used with the API (e.g., API keys, OAuth2, HTTP authentication).
	Security          []SecurityRequirement      // A list of security requirement objects that specify the security requirements for the API. This can include multiple security requirements that apply to different operations or endpoints.
	Tags              []Tag                      // A list of tag objects that can be used to categorize and organize the API operations. Tags can be used to group related operations together and provide additional metadata for the API documentation.
	ExternalDocs      *ExternalDocs              // A reference to external documentation for the API, which can include a description and a URL pointing to the full documentation. This can be used to provide users with additional resources and information about the API beyond what is included in the generated OpenAPI document.

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
