package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
)

//nolint:gocognit // validates top-level document rules in one traversal for coherent errors.
func ValidateDocument(doc *openapi.Document, version string) []error {
	var errs []error
	if doc.OpenAPI != "" && doc.OpenAPI != version {
		errs = append(errs, fmt.Errorf("openapi must be %s, got %s", version, doc.OpenAPI))
	}
	if doc.Self != "" && !IsURIReference(doc.Self) {
		errs = append(errs, fmt.Errorf("$self must be a URI reference"))
	}
	if doc.JSONSchemaDialect != "" && !IsURIReference(doc.JSONSchemaDialect) {
		errs = append(errs, fmt.Errorf("jsonSchemaDialect must be a URI"))
	}
	if doc.Info.Title == "" {
		errs = append(errs, fmt.Errorf("info.title is required"))
	}
	if doc.Info.Version == "" {
		errs = append(errs, fmt.Errorf("info.version is required"))
	}
	errs = append(errs, ValidateInfo(doc.Info, version)...)
	if reflect.IsOpenAPI30(version) && doc.Paths == nil {
		errs = append(errs, fmt.Errorf("paths is required"))
	}
	if reflect.IsOpenAPI30(version) {
		if doc.JSONSchemaDialect != "" {
			errs = append(errs, fmt.Errorf("jsonSchemaDialect requires OpenAPI 3.1.x or 3.2.0"))
		}
		if doc.Webhooks != nil {
			errs = append(errs, fmt.Errorf("webhooks requires OpenAPI 3.1.x or 3.2.0"))
		}
	}
	if IsOpenAPI31(version) || IsOpenAPI32(version) {
		if doc.Paths == nil && doc.Webhooks == nil && doc.Components == nil {
			errs = append(errs, fmt.Errorf("one of paths, components, or webhooks is required"))
		}
	}
	for i := range doc.Servers {
		errs = append(errs, ValidateServer(fmt.Sprintf("servers[%d]", i), &doc.Servers[i])...)
	}
	errs = append(errs, ValidateServerNames(doc.Servers)...)
	if doc.ExternalDocs != nil && doc.ExternalDocs.URL == "" {
		errs = append(errs, fmt.Errorf("externalDocs.url is required"))
	}
	if doc.ExternalDocs != nil && doc.ExternalDocs.URL != "" && !IsURIReference(doc.ExternalDocs.URL) {
		errs = append(errs, fmt.Errorf("externalDocs.url must be a URI"))
	}
	securitySchemes := map[string]*openapi.SecurityScheme{}
	componentParameters := map[string]*openapi.Parameter{}
	if doc.Components != nil {
		securitySchemes = doc.Components.SecuritySchemes
		componentParameters = doc.Components.Parameters
	}
	errs = append(errs, ValidateSecurityRequirements("security", doc.Security, securitySchemes, version)...)
	operationIDs := map[string]string{}
	normalizedPaths := map[string]string{}
	for path, item := range doc.Paths {
		if normalized := NormalizeTemplatedPath(path); normalized != path {
			if previous, exists := normalizedPaths[normalized]; exists && previous != path {
				errs = append(errs, fmt.Errorf("path %q conflicts with equivalent templated path %q", path, previous))
			} else {
				normalizedPaths[normalized] = path
			}
		}
		errs = append(
			errs,
			ValidatePathItem(path, item, version, operationIDs, securitySchemes, componentParameters)...)
	}
	for name, item := range doc.Webhooks {
		errs = append(
			errs,
			ValidatePathItemOperations(
				"webhooks."+name,
				item,
				version,
				operationIDs,
				securitySchemes,
				componentParameters,
			)...)
	}
	if doc.Components != nil {
		errs = append(
			errs,
			ValidateComponents(doc.Components, version, operationIDs, securitySchemes, componentParameters)...)
	}
	errs = append(errs, ValidateTags(doc.Tags, version)...)
	errs = append(errs, ValidateReferenceTargets(doc)...)
	return errs
}

func ValidateTags(tags []openapi.Tag, version string) []error {
	var errs []error
	tagByName := make(map[string]int, len(tags))
	for i, tag := range tags {
		if tag.Name == "" {
			errs = append(errs, fmt.Errorf("tags[%d].name is required", i))
		} else if previous, exists := tagByName[tag.Name]; exists {
			errs = append(errs, fmt.Errorf("tags[%d].name %q duplicates tags[%d].name", i, tag.Name, previous))
		} else {
			tagByName[tag.Name] = i
		}
		if version != openapi.Version320 && (tag.Summary != "" || tag.Parent != "" || tag.Kind != "") {
			errs = append(errs, fmt.Errorf("tags[%d] summary, parent, and kind require OpenAPI 3.2.0", i))
		}
		if tag.ExternalDocs != nil && tag.ExternalDocs.URL == "" {
			errs = append(errs, fmt.Errorf("tags[%d].externalDocs.url is required", i))
		}
		if tag.ExternalDocs != nil && tag.ExternalDocs.URL != "" && !IsURIReference(tag.ExternalDocs.URL) {
			errs = append(errs, fmt.Errorf("tags[%d].externalDocs.url must be a URI", i))
		}
	}
	if version == openapi.Version320 {
		errs = append(errs, ValidateTagParents(tags, tagByName)...)
	}
	return errs
}

func ValidateTagParents(tags []openapi.Tag, tagByName map[string]int) []error {
	var errs []error
	for i, tag := range tags {
		if tag.Name == "" || tag.Parent == "" {
			continue
		}
		if _, exists := tagByName[tag.Parent]; !exists {
			errs = append(errs, fmt.Errorf("tags[%d].parent %q must reference an existing tag", i, tag.Parent))
			continue
		}
		seen := map[string]bool{tag.Name: true}
		for parent := tag.Parent; parent != ""; {
			if seen[parent] {
				errs = append(errs, fmt.Errorf("tags[%d].parent creates a circular tag parent reference", i))
				break
			}
			seen[parent] = true
			parentIndex := tagByName[parent]
			parent = tags[parentIndex].Parent
		}
	}
	return errs
}

func ValidateInfo(info openapi.Info, version string) []error {
	var errs []error
	if reflect.IsOpenAPI30(version) && info.Summary != "" {
		errs = append(errs, fmt.Errorf("info.summary requires OpenAPI 3.1.x or 3.2.0"))
	}
	if info.TermsOfService != nil && !IsURIReference(*info.TermsOfService) {
		errs = append(errs, fmt.Errorf("info.termsOfService must be a URI"))
	}
	if info.Contact != nil && info.Contact.URL != "" && !IsURIReference(info.Contact.URL) {
		errs = append(errs, fmt.Errorf("info.contact.url must be a URI"))
	}
	if info.Contact != nil && info.Contact.Email != "" && !strings.Contains(info.Contact.Email, "@") {
		errs = append(errs, fmt.Errorf("info.contact.email must be an email address"))
	}
	if info.License != nil {
		if info.License.Name == "" {
			errs = append(errs, fmt.Errorf("info.license.name is required"))
		}
		if info.License.URL != "" && !IsURIReference(info.License.URL) {
			errs = append(errs, fmt.Errorf("info.license.url must be a URI"))
		}
		if reflect.IsOpenAPI30(version) && info.License.Identifier != "" {
			errs = append(errs, fmt.Errorf("info.license.identifier requires OpenAPI 3.1.x or 3.2.0"))
		}
		if info.License.Identifier != "" && info.License.URL != "" {
			errs = append(errs, fmt.Errorf("info.license.identifier and info.license.url are mutually exclusive"))
		}
	}
	return errs
}

func ValidateServerNames(servers []openapi.Server) []error {
	var errs []error
	serverNames := map[string]int{}
	for i, server := range servers {
		if server.Name != "" {
			if previous, exists := serverNames[server.Name]; exists {
				errs = append(errs,
					fmt.Errorf("servers[%d].name %q duplicates servers[%d].name", i, server.Name, previous))
			} else {
				serverNames[server.Name] = i
			}
		}
	}
	return errs
}

func ValidateServer(context string, server *openapi.Server) []error {
	var errs []error
	if server == nil {
		return nil
	}
	if server.URL == "" {
		errs = append(errs, fmt.Errorf("%s.url is required", context))
	} else if !IsServerURL(server.URL) {
		errs = append(errs, fmt.Errorf("%s.url must not contain a query or fragment", context))
	}
	for name, variable := range server.Variables {
		if variable.Default == "" {
			errs = append(errs, fmt.Errorf("%s.variables.%s.default is required", context, name))
		}
		if len(variable.Enum) > 0 && !slices.Contains(variable.Enum, variable.Default) {
			errs = append(errs, fmt.Errorf("%s.variables.%s.default must be one of enum values", context, name))
		}
	}
	return errs
}
