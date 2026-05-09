package reflect

import (
	"reflect"

	"github.com/oaswrap/spec/openapi"
)

//nolint:funlen // covers full OpenAPI scalar/collection/struct mapping in one switch for readability.
func (r *Reflector) SchemaForType(t reflect.Type, mode SchemaMode, field *reflect.StructField) *openapi.Schema {
	nullable := false
	for t != nil && t.Kind() == reflect.Pointer {
		nullable = true
		t = t.Elem()
	}
	if t == nil {
		return &openapi.Schema{}
	}
	if mapped := r.TypeMapping[t]; mapped != nil {
		t = mapped
	}
	if schema := r.SchemaFromTypeExposer(t); schema != nil {
		r.ApplyNullable(schema, nullable)
		if field != nil {
			r.ApplySchemaTags(schema, *field)
		}
		return schema
	}
	if mode == SchemaUseComponent && IsComponentType(t) && !r.InlineRefs() {
		schema := r.RefSchema(t)
		r.ApplyNullable(schema, nullable)
		if field != nil {
			r.ApplySchemaTags(schema, *field)
		}
		return schema
	}

	var schema *openapi.Schema
	switch t.Kind() { //nolint:exhaustive // only interested in types supported by OpenAPI
	case reflect.Bool:
		schema = &openapi.Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		schema = &openapi.Schema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		schema = &openapi.Schema{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		minVal := 0.0
		schema = &openapi.Schema{Type: "integer", Format: "int32", Minimum: &minVal}
	case reflect.Uint64, reflect.Uintptr:
		minVal := 0.0
		schema = &openapi.Schema{Type: "integer", Format: "int64", Minimum: &minVal}
	case reflect.Float32:
		schema = &openapi.Schema{Type: "number", Format: "float"}
	case reflect.Float64:
		schema = &openapi.Schema{Type: "number", Format: "double"}
	case reflect.String:
		schema = &openapi.Schema{Type: "string"}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			if IsOpenAPI30(r.Config.OpenAPIVersion) {
				schema = &openapi.Schema{Type: "string", Format: "byte"}
			} else {
				schema = &openapi.Schema{Type: "string", ContentEncoding: "base64"}
			}
			break
		}
		schema = &openapi.Schema{Type: "array", Items: r.SchemaForType(t.Elem(), SchemaUseComponent, nil)}
	case reflect.Map:
		schema = &openapi.Schema{
			Type:                 "object",
			AdditionalProperties: r.SchemaForType(t.Elem(), SchemaUseComponent, nil),
		}
	case reflect.Struct:
		if IsTime(t) {
			schema = &openapi.Schema{Type: "string", Format: "date-time"}
		} else {
			schema = r.StructSchema(t, "json", false, mode)
		}
	case reflect.Interface:
		schema = &openapi.Schema{}
	default:
		schema = &openapi.Schema{}
	}
	r.ApplyNullable(schema, nullable)
	if field != nil {
		r.ApplySchemaTags(schema, *field)
	}
	return schema
}

func (r *Reflector) ApplyNullable(schema *openapi.Schema, nullable bool) {
	if !nullable || schema == nil {
		return
	}
	if IsOpenAPI30(r.Config.OpenAPIVersion) {
		if schema.Ref != "" {
			ref := schema.Ref
			*schema = openapi.Schema{
				AllOf:    []*openapi.Schema{{Ref: ref}},
				Nullable: true,
			}
			return
		}
		schema.Nullable = true
		return
	}
	if schema.Ref != "" {
		ref := schema.Ref
		*schema = openapi.Schema{
			AnyOf: []*openapi.Schema{
				{Ref: ref},
				{Type: "null"},
			},
		}
		return
	}
	if typ, ok := schema.Type.(string); ok && typ != "" {
		schema.Type = []string{typ, "null"}
	}
}

func IsOpenAPI30(version string) bool {
	return version == openapi.Version300 || version == openapi.Version301 || version == openapi.Version302 ||
		version == openapi.Version303 || version == openapi.Version304
}

func IsComponentType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.Name() != "" && !IsTime(t)
}
