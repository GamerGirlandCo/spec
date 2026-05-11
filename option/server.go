package option

import "github.com/oaswrap/spec/openapi"

// ServerOption mutates a server entry.
type ServerOption func(*openapi.Server)

// ServerDescription sets server description.
func ServerDescription(description string) ServerOption {
	return func(server *openapi.Server) { server.Description = &description }
}

// ServerVariables sets server variables map.
func ServerVariables(variables map[string]openapi.ServerVariable) ServerOption {
	return func(server *openapi.Server) { server.Variables = variables }
}

// ServerName sets server name.
// It is only valid for OpenAPI 3.2.0.
func ServerName(name string) ServerOption {
	return func(server *openapi.Server) { server.Name = name }
}
