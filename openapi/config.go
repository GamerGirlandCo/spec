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

type ReflectorConfig struct {
	InlineRefs          bool
	StripDefNamePrefix  []string
	InterceptDefName    func(t reflect.Type, defaultDefName string) string
	TypeMappings        []TypeMapping
	ParameterTagMapping map[ParameterIn]string
}

type TypeMapping struct {
	Src any
	Dst any
}

type PathParser interface {
	Parse(path string) (string, error)
}
