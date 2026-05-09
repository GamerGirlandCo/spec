package validate

import (
	"fmt"

	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/openapi"
)

//nolint:gocognit // recursive schema validation must cover many keyword branches.
func ValidateSchema(context string, schema *openapi.Schema, version string, visited map[*openapi.Schema]bool) []error {
	var errs []error
	if schema == nil {
		return nil
	}
	if visited[schema] {
		return nil
	}
	visited[schema] = true
	if reflect.IsOpenAPI30(version) {
		if schema.Ref != "" && HasSchemaRefSiblings(schema) {
			errs = append(errs, fmt.Errorf("%s must not define siblings with $ref in OpenAPI 3.0.x", context))
		}
		if schema.ReadOnly && schema.WriteOnly {
			errs = append(errs, fmt.Errorf("%s must not be both readOnly and writeOnly", context))
		}
		errs = append(errs, ValidateSchema304Fields(context, schema)...)
	}
	if version != openapi.Version320 {
		if schema.Discriminator != nil && ExtraHas(schema.Discriminator.Extra, "defaultMapping") {
			errs = append(errs, fmt.Errorf("%s.discriminator.defaultMapping requires OpenAPI 3.2.0", context))
		}
		if schema.XML != nil && ExtraHas(schema.XML.Extra, "nodeType") {
			errs = append(errs, fmt.Errorf("%s.xml.nodeType requires OpenAPI 3.2.0", context))
		}
	}
	for name, child := range schema.Defs {
		errs = append(errs, ValidateSchema(context+".$defs."+name, child, version, visited)...)
	}
	for name, child := range schema.Properties {
		errs = append(errs, ValidateSchema(context+".properties."+name, child, version, visited)...)
	}
	for name, child := range schema.PatternProperties {
		errs = append(errs, ValidateSchema(context+".patternProperties."+name, child, version, visited)...)
	}
	errs = append(errs, ValidateSchema(context+".items", schema.Items, version, visited)...)
	for i, child := range schema.PrefixItems {
		errs = append(errs, ValidateSchema(fmt.Sprintf("%s.prefixItems[%d]", context, i), child, version, visited)...)
	}
	errs = append(errs, ValidateSchema(context+".contains", schema.Contains, version, visited)...)
	errs = append(
		errs,
		ValidateAnySchema(context+".additionalProperties", schema.AdditionalProperties, version, visited)...)
	errs = append(
		errs,
		ValidateAnySchema(context+".unevaluatedProperties", schema.UnevaluatedProperties, version, visited)...)
	errs = append(errs, ValidateSchema(context+".propertyNames", schema.PropertyNames, version, visited)...)
	for name, child := range schema.DependentSchemas {
		errs = append(errs, ValidateSchema(context+".dependentSchemas."+name, child, version, visited)...)
	}
	for i, child := range schema.AllOf {
		errs = append(errs, ValidateSchema(fmt.Sprintf("%s.allOf[%d]", context, i), child, version, visited)...)
	}
	for i, child := range schema.AnyOf {
		errs = append(errs, ValidateSchema(fmt.Sprintf("%s.anyOf[%d]", context, i), child, version, visited)...)
	}
	for i, child := range schema.OneOf {
		errs = append(errs, ValidateSchema(fmt.Sprintf("%s.oneOf[%d]", context, i), child, version, visited)...)
	}
	errs = append(errs, ValidateSchema(context+".not", schema.Not, version, visited)...)
	errs = append(errs, ValidateSchema(context+".if", schema.If, version, visited)...)
	errs = append(errs, ValidateSchema(context+".then", schema.Then, version, visited)...)
	errs = append(errs, ValidateSchema(context+".else", schema.Else, version, visited)...)
	errs = append(errs, ValidateSchema(context+".contentSchema", schema.ContentSchema, version, visited)...)
	return errs
}

func ValidateAnySchema(context string, value any, version string, visited map[*openapi.Schema]bool) []error {
	switch typed := value.(type) {
	case *openapi.Schema:
		return ValidateSchema(context, typed, version, visited)
	case openapi.Schema:
		return ValidateSchema(context, &typed, version, visited)
	default:
		return nil
	}
}

//nolint:gocyclo,cyclop // OpenAPI 3.0.x checks enumerate many incompatible JSON Schema keywords.
func ValidateSchema304Fields(context string, schema *openapi.Schema) []error {
	var errs []error
	if schema.Schema != "" || schema.ID != "" || len(schema.Defs) > 0 || schema.Anchor != "" ||
		schema.DynamicAnchor != "" ||
		schema.DynamicRef != "" ||
		len(schema.Vocabulary) > 0 ||
		schema.Comment != "" {
		errs = append(
			errs,
			fmt.Errorf("%s contains JSON Schema dialect fields that require OpenAPI 3.1.x or 3.2.0", context),
		)
	}
	if len(schema.Examples) > 0 || schema.Const != nil || len(schema.PatternProperties) > 0 ||
		len(schema.PrefixItems) > 0 ||
		schema.Contains != nil ||
		schema.MaxContains != nil ||
		schema.MinContains != nil ||
		schema.UnevaluatedProperties != nil ||
		schema.PropertyNames != nil ||
		len(schema.DependentRequired) > 0 ||
		len(schema.DependentSchemas) > 0 ||
		schema.If != nil ||
		schema.Then != nil ||
		schema.Else != nil ||
		schema.ContentEncoding != "" ||
		schema.ContentMediaType != "" ||
		schema.ContentSchema != nil {
		errs = append(
			errs,
			fmt.Errorf("%s contains JSON Schema 2020-12 keywords that require OpenAPI 3.1.x or 3.2.0", context),
		)
	}
	if _, ok := schema.Type.([]string); ok {
		errs = append(errs, fmt.Errorf("%s.type must be a string in OpenAPI 3.0.x", context))
	}
	if _, ok := schema.Type.([]any); ok {
		errs = append(errs, fmt.Errorf("%s.type must be a string in OpenAPI 3.0.x", context))
	}
	if schema.ExclusiveMaximum != nil {
		if _, ok := schema.ExclusiveMaximum.(bool); !ok {
			errs = append(errs, fmt.Errorf("%s.exclusiveMaximum must be a boolean in OpenAPI 3.0.x", context))
		}
	}
	if schema.ExclusiveMinimum != nil {
		if _, ok := schema.ExclusiveMinimum.(bool); !ok {
			errs = append(errs, fmt.Errorf("%s.exclusiveMinimum must be a boolean in OpenAPI 3.0.x", context))
		}
	}
	if ExtraHas(
		schema.Extra,
		"$schema",
		"$id",
		"$defs",
		"$anchor",
		"$dynamicAnchor",
		"$dynamicRef",
		"$vocabulary",
		"$comment",
		"examples",
		"const",
		"patternProperties",
		"prefixItems",
		"contains",
		"maxContains",
		"minContains",
		"unevaluatedProperties",
		"propertyNames",
		"dependentRequired",
		"dependentSchemas",
		"if",
		"then",
		"else",
		"contentEncoding",
		"contentMediaType",
		"contentSchema",
	) {
		errs = append(
			errs,
			fmt.Errorf("%s contains Extra JSON Schema keywords that require OpenAPI 3.1.x or 3.2.0", context),
		)
	}
	return errs
}
