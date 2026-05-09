package option

import "github.com/oaswrap/spec/openapi"

// OperationConfig stores effective operation-level settings.
type OperationConfig struct {
	Hide         bool
	OperationID  string
	Description  string
	Summary      string
	ExternalDocs *openapi.ExternalDocs
	Deprecated   bool
	Tags         []string
	Security     []OperationSecurityConfig
	Requests     []*openapi.ContentUnit
	Responses    []*openapi.ContentUnit
	Customizers  []func(*openapi.Operation)
}

// OperationSecurityConfig describes one operation security requirement entry.
type OperationSecurityConfig struct {
	Name   string
	Scopes []string
}

// OperationOption mutates operation generation behavior.
type OperationOption func(*OperationConfig)

// Hidden skips emitting this operation.
func Hidden(hide ...bool) OperationOption {
	return func(cfg *OperationConfig) { cfg.Hide = optional(true, hide...) }
}

// OperationID sets `operationId`.
func OperationID(id string) OperationOption {
	return func(cfg *OperationConfig) { cfg.OperationID = id }
}

// Description sets operation description.
func Description(description string) OperationOption {
	return func(cfg *OperationConfig) { cfg.Description = description }
}

// ExternalDocs sets operation external documentation.
func ExternalDocs(url string, description ...string) OperationOption {
	return func(cfg *OperationConfig) {
		docs := &openapi.ExternalDocs{URL: url}
		if len(description) > 0 {
			docs.Description = description[0]
		}
		cfg.ExternalDocs = docs
	}
}

// Summary sets operation summary and, when empty, description.
func Summary(summary string) OperationOption {
	return func(cfg *OperationConfig) {
		cfg.Summary = summary
		if cfg.Description == "" {
			cfg.Description = summary
		}
	}
}

// Deprecated marks the operation as deprecated.
func Deprecated(deprecated ...bool) OperationOption {
	return func(cfg *OperationConfig) { cfg.Deprecated = optional(true, deprecated...) }
}

// Tags appends operation tags.
func Tags(tags ...string) OperationOption {
	return func(cfg *OperationConfig) { cfg.Tags = append(cfg.Tags, tags...) }
}

// Security appends one operation security requirement.
func Security(name string, scopes ...string) OperationOption {
	return func(cfg *OperationConfig) {
		cfg.Security = append(cfg.Security, OperationSecurityConfig{Name: name, Scopes: scopes})
	}
}

// Request appends one request content unit.
func Request(structure any, opts ...ContentOption) OperationOption {
	return func(cfg *OperationConfig) {
		cu := &openapi.ContentUnit{Structure: structure}
		for _, opt := range opts {
			opt(cu)
		}
		cfg.Requests = append(cfg.Requests, cu)
	}
}

// Response appends one response content unit.
func Response(httpStatus int, structure any, opts ...ContentOption) OperationOption {
	return func(cfg *OperationConfig) {
		cu := &openapi.ContentUnit{HTTPStatus: httpStatus, Structure: structure}
		for _, opt := range opts {
			opt(cu)
		}
		cfg.Responses = append(cfg.Responses, cu)
	}
}

// CustomizeOperation applies a low-level mutation to the generated operation.
func CustomizeOperation(fn func(*openapi.Operation)) OperationOption {
	return func(cfg *OperationConfig) {
		if fn != nil {
			cfg.Customizers = append(cfg.Customizers, fn)
		}
	}
}
