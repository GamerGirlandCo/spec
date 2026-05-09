package openapi

type Document struct {
	OpenAPI           string                `json:"openapi"                     yaml:"openapi"`
	Self              string                `json:"$self,omitempty"             yaml:"$self,omitempty"`
	Info              Info                  `json:"info"                        yaml:"info"`
	JSONSchemaDialect string                `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	Servers           []Server              `json:"servers,omitempty"           yaml:"servers,omitempty"`
	Paths             map[string]*PathItem  `json:"paths"                       yaml:"paths"`
	Webhooks          map[string]*PathItem  `json:"webhooks,omitempty"          yaml:"webhooks,omitempty"`
	Components        *Components           `json:"components,omitempty"        yaml:"components,omitempty"`
	Security          []SecurityRequirement `json:"security,omitempty"          yaml:"security,omitempty"`
	Tags              []Tag                 `json:"tags,omitempty"              yaml:"tags,omitempty"`
	ExternalDocs      *ExternalDocs         `json:"externalDocs,omitempty"      yaml:"externalDocs,omitempty"`
	Extensions        map[string]any        `json:"-"                           yaml:"-"`
	Extra             map[string]any        `json:"-"                           yaml:"-"`
}

type Info struct {
	Title          string         `json:"title"                    yaml:"title"`
	Summary        string         `json:"summary,omitempty"        yaml:"summary,omitempty"`
	Description    *string        `json:"description,omitempty"    yaml:"description,omitempty"`
	TermsOfService *string        `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact       `json:"contact,omitempty"        yaml:"contact,omitempty"`
	License        *License       `json:"license,omitempty"        yaml:"license,omitempty"`
	Version        string         `json:"version"                  yaml:"version"`
	Extensions     map[string]any `json:"-"                        yaml:"-"`
	Extra          map[string]any `json:"-"                        yaml:"-"`
}

type Contact struct {
	Name       string         `json:"name,omitempty"  yaml:"name,omitempty"`
	URL        string         `json:"url,omitempty"   yaml:"url,omitempty"`
	Email      string         `json:"email,omitempty" yaml:"email,omitempty"`
	Extensions map[string]any `json:"-"               yaml:"-"`
	Extra      map[string]any `json:"-"               yaml:"-"`
}

type License struct {
	Name       string         `json:"name"                 yaml:"name"`
	Identifier string         `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string         `json:"url,omitempty"        yaml:"url,omitempty"`
	Extensions map[string]any `json:"-"                    yaml:"-"`
	Extra      map[string]any `json:"-"                    yaml:"-"`
}

type Tag struct {
	Name         string         `json:"name"                   yaml:"name"`
	Summary      string         `json:"summary,omitempty"      yaml:"summary,omitempty"`
	Description  string         `json:"description,omitempty"  yaml:"description,omitempty"`
	ExternalDocs *ExternalDocs  `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Parent       string         `json:"parent,omitempty"       yaml:"parent,omitempty"`
	Kind         string         `json:"kind,omitempty"         yaml:"kind,omitempty"`
	Extensions   map[string]any `json:"-"                      yaml:"-"`
	Extra        map[string]any `json:"-"                      yaml:"-"`
}

type ExternalDocs struct {
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string         `json:"url"                   yaml:"url"`
	Extensions  map[string]any `json:"-"                     yaml:"-"`
	Extra       map[string]any `json:"-"                     yaml:"-"`
}

type Server struct {
	URL         string                    `json:"url"                   yaml:"url"`
	Description *string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"   yaml:"variables,omitempty"`
	Extensions  map[string]any            `json:"-"                     yaml:"-"`
	Extra       map[string]any            `json:"-"                     yaml:"-"`
}

type ServerVariable struct {
	Enum        []string       `json:"enum,omitempty"        yaml:"enum,omitempty"`
	Default     string         `json:"default"               yaml:"default"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Extensions  map[string]any `json:"-"                     yaml:"-"`
	Extra       map[string]any `json:"-"                     yaml:"-"`
}

type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"         yaml:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"       yaml:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"      yaml:"parameters,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty"        yaml:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"   yaml:"requestBodies,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty"         yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]*Link           `json:"links,omitempty"           yaml:"links,omitempty"`
	Callbacks       map[string]*Callback       `json:"callbacks,omitempty"       yaml:"callbacks,omitempty"`
	PathItems       map[string]*PathItem       `json:"pathItems,omitempty"       yaml:"pathItems,omitempty"`
	MediaTypes      map[string]*MediaType      `json:"mediaTypes,omitempty"      yaml:"mediaTypes,omitempty"`
	Extensions      map[string]any             `json:"-"                         yaml:"-"`
	Extra           map[string]any             `json:"-"                         yaml:"-"`
}
