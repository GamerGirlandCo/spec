package reflect

import (
	"reflect"
	"time"

	"github.com/oaswrap/spec/openapi"
)

type SchemaMode int

const (
	SchemaInline SchemaMode = iota
	SchemaUseComponent
)

type Reflector struct {
	Config      *openapi.Config
	Components  map[string]*openapi.Schema
	Names       map[reflect.Type]string
	Generating  map[reflect.Type]bool
	TypeMapping map[reflect.Type]reflect.Type
}

func NewReflector(cfg *openapi.Config) *Reflector {
	r := &Reflector{
		Config:      cfg,
		Components:  map[string]*openapi.Schema{},
		Names:       map[reflect.Type]string{},
		Generating:  map[reflect.Type]bool{},
		TypeMapping: map[reflect.Type]reflect.Type{},
	}
	if cfg.ReflectorConfig != nil {
		for _, tm := range cfg.ReflectorConfig.TypeMappings {
			src := IndirectType(reflect.TypeOf(tm.Src))
			dst := IndirectType(reflect.TypeOf(tm.Dst))
			if src != nil && dst != nil {
				r.TypeMapping[src] = dst
			}
		}
	}
	return r
}

func (r *Reflector) RequestParts(
	value any,
	ct string,
) ([]*openapi.Parameter, *openapi.Schema) {
	t := IndirectType(reflect.TypeOf(value))
	if t == nil {
		return nil, nil
	}
	if mapped := r.TypeMapping[t]; mapped != nil {
		t = mapped
	}
	if t.Kind() != reflect.Struct || IsTime(t) {
		return nil, r.SchemaForType(t, SchemaUseComponent, nil)
	}

	var params []*openapi.Parameter
	bodyTag := BodyNameTag(ct)
	hasBody := false
	hasParam := false
	ForEachField(t, func(field reflect.StructField) {
		if paramIn, name, ok := r.ParameterField(field); ok {
			hasParam = true
			params = append(params, r.ParameterSchema(field, paramIn, name))
		}
		if TagName(field, bodyTag) != "" || (bodyTag != "json" && TagName(field, "json") != "") {
			hasBody = true
		}
	})
	if !hasParam {
		return nil, r.SchemaForType(t, SchemaUseComponent, nil)
	}
	if !hasBody {
		return params, nil
	}
	body := r.StructSchema(t, bodyTag, true, SchemaInline)
	if len(body.Properties) == 0 {
		body = nil
	}
	return params, body
}

func (r *Reflector) ParameterField(field reflect.StructField) (string, string, bool) {
	tagPairs := []struct {
		in  openapi.ParameterIn
		tag string
	}{
		{openapi.ParameterInPath, "path"},
		{openapi.ParameterInQuery, "query"},
		{openapi.ParameterInHeader, "header"},
		{openapi.ParameterInCookie, "cookie"},
	}
	if r.Config.OpenAPIVersion == openapi.Version320 {
		tagPairs = append(tagPairs, struct {
			in  openapi.ParameterIn
			tag string
		}{openapi.ParameterInQueryString, "querystring"})
	}
	custom := map[openapi.ParameterIn]string{}
	if r.Config.ReflectorConfig != nil {
		for in, tag := range r.Config.ReflectorConfig.ParameterTagMapping {
			custom[in] = tag
		}
	}
	for _, pair := range tagPairs {
		if tag, ok := custom[pair.in]; ok {
			delete(custom, pair.in)
			if tag != "" && tag != pair.tag {
				tagPairs = append(tagPairs, struct {
					in  openapi.ParameterIn
					tag string
				}{pair.in, tag})
			}
		}
	}
	for in, tag := range custom {
		tagPairs = append(tagPairs, struct {
			in  openapi.ParameterIn
			tag string
		}{in, tag})
	}
	for _, pair := range tagPairs {
		in, tag := pair.in, pair.tag
		if name := TagName(field, tag); name != "" {
			return string(in), name, true
		}
	}
	return "", "", false
}

func (r *Reflector) ParameterSchema(field reflect.StructField, in, name string) *openapi.Parameter {
	schema := r.SchemaForType(field.Type, SchemaInline, &field)
	return &openapi.Parameter{
		Name:        name,
		In:          in,
		Description: field.Tag.Get("description"),
		Required:    in == string(openapi.ParameterInPath) || BoolTag(field.Tag.Get("required")),
		Deprecated:  BoolTag(field.Tag.Get("deprecated")),
		Schema:      schema,
	}
}

func (r *Reflector) SchemaForValue(value any, mode SchemaMode) *openapi.Schema {
	if ov, ok := value.(OneOfValue); ok {
		values := ov.GetValues()
		schemas := make([]*openapi.Schema, 0, len(values))
		for _, item := range values {
			schemas = append(schemas, r.SchemaForValue(item, mode))
		}
		return &openapi.Schema{OneOf: schemas}
	}
	if schema, ok := value.(*openapi.Schema); ok {
		return schema
	}
	if schema := r.SchemaFromValueExposer(value); schema != nil {
		return schema
	}
	return r.SchemaForType(IndirectType(reflect.TypeOf(value)), mode, nil)
}

func (r *Reflector) RefSchema(t reflect.Type) *openapi.Schema {
	name := r.TypeName(t)
	if _, ok := r.Components[name]; ok {
		return &openapi.Schema{Ref: "#/components/schemas/" + name}
	}
	if r.Generating[t] {
		return &openapi.Schema{Ref: "#/components/schemas/" + name}
	}
	r.Generating[t] = true
	r.Components[name] = &openapi.Schema{}
	r.Components[name] = r.StructSchema(t, "json", false, SchemaInline)
	delete(r.Generating, t)
	return &openapi.Schema{Ref: "#/components/schemas/" + name}
}

func (r *Reflector) StructSchema(
	t reflect.Type,
	nameTag string,
	onlyTagged bool,
	mode SchemaMode,
) *openapi.Schema {
	schema := &openapi.Schema{Type: "object", Properties: map[string]*openapi.Schema{}}
	ForEachField(t, func(field reflect.StructField) {
		if IgnoredField(field, nameTag) {
			return
		}
		name := TagName(field, nameTag)
		if name == "" && nameTag != "json" {
			name = TagName(field, "json")
		}
		if name == "" {
			if onlyTagged {
				return
			}
			name = LowerCamel(field.Name)
		}
		prop := r.SchemaForType(field.Type, mode, &field)
		schema.Properties[name] = prop
		if BoolTag(field.Tag.Get("required")) {
			schema.Required = append(schema.Required, name)
		}
	})
	if len(schema.Properties) == 0 {
		schema.Properties = nil
	}
	return schema
}

// SchemaExposer lets a value provide an OpenAPI schema for a specific version.
type SchemaExposer interface {
	OpenAPISchema(version string) *openapi.Schema
}

// StaticSchemaExposer lets a value provide a version-independent OpenAPI schema.
type StaticSchemaExposer interface {
	OpenAPISchema() *openapi.Schema
}

// OneOfValue represents a reflected one-of union container.
type OneOfValue interface {
	GetValues() []any
}

func (r *Reflector) SchemaFromValueExposer(value any) *openapi.Schema {
	if value == nil {
		return nil
	}
	if exposer, ok := value.(SchemaExposer); ok {
		return exposer.OpenAPISchema(r.Config.OpenAPIVersion)
	}
	if exposer, ok := value.(StaticSchemaExposer); ok {
		return exposer.OpenAPISchema()
	}
	return nil
}

func (r *Reflector) SchemaFromTypeExposer(t reflect.Type) *openapi.Schema {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Interface {
		return nil
	}
	value := reflect.New(t).Interface()
	return r.SchemaFromValueExposer(value)
}

func IsTime(t reflect.Type) bool {
	return t == reflect.TypeFor[time.Time]()
}
