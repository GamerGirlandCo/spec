package reflect_test

import (
	"errors"
	std_reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/reflect"
	"github.com/oaswrap/spec/internal/testutil/dto"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestReflector_ParameterSchema(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	r := reflect.NewReflector(cfg)

	type ParamStruct struct {
		ID string `path:"id" description:"User ID" required:"true" deprecated:"true"`
	}
	f, _ := std_reflect.TypeFor[ParamStruct]().FieldByName("ID")

	p := r.ParameterSchema(f, "path", "id")
	assert.Equal(t, "id", p.Name)
	assert.Equal(t, "path", p.In)
	assert.Equal(t, "User ID", p.Description)
	assert.True(t, p.Required)
	assert.True(t, p.Deprecated)
	assert.Equal(t, "string", p.Schema.Type)
}

func TestReflector_ParameterSchema_QueryString(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version320}
	r := reflect.NewReflector(cfg)

	t.Run("DefaultMediaType", func(t *testing.T) {
		type QS struct {
			Q string `querystring:"q"`
		}
		f, _ := std_reflect.TypeFor[QS]().FieldByName("Q")
		p := r.ParameterSchema(f, string(openapi.ParameterInQueryString), "q")
		assert.Nil(t, p.Schema)
		if assert.Contains(t, p.Content, "application/x-www-form-urlencoded") {
			assert.Equal(t, "string", p.Content["application/x-www-form-urlencoded"].Schema.Type)
		}
	})

	t.Run("OverrideViaTag", func(t *testing.T) {
		type QS struct {
			Q string `querystring:"q" mediaType:"application/json"`
		}
		f, _ := std_reflect.TypeFor[QS]().FieldByName("Q")
		p := r.ParameterSchema(f, string(openapi.ParameterInQueryString), "q")
		assert.Nil(t, p.Schema)
		assert.Contains(t, p.Content, "application/json")
		assert.NotContains(t, p.Content, "application/x-www-form-urlencoded")
	})
}

func TestReflector_SchemaForValue(t *testing.T) {
	cfg := &openapi.Config{OpenAPIVersion: openapi.Version304}
	r := reflect.NewReflector(cfg)

	t.Run("OneOf", func(t *testing.T) {
		val := spec.OneOf(1, "two")
		schema, err := r.SchemaForValue(val, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Len(t, schema.OneOf, 2)
	})

	t.Run("SchemaPointer", func(t *testing.T) {
		expected := &openapi.Schema{Type: "boolean"}
		schema, err := r.SchemaForValue(expected, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Equal(t, expected, schema)
	})
}

func TestReflector_Config(t *testing.T) {
	t.Run("InterceptDefName", func(t *testing.T) {
		r := spec.NewRouter(option.WithReflectorConfig(
			option.InterceptDefName(func(_ std_reflect.Type, _ string) string {
				return "CustomName"
			}),
		))
		type NamedStruct struct{ Foo string }
		r.Get("/ping", option.Response(200, NamedStruct{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		assert.Contains(t, doc.Components.Schemas, "CustomName")
	})

	t.Run("DuplicateNames", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithReflectorConfig(option.InterceptDefName(func(_ std_reflect.Type, _ string) string {
				return "Collision"
			})),
		)

		type TypeA struct{ Foo string }
		type TypeB struct{ Bar string }

		r.Get("/a", option.Response(200, TypeA{}))
		r.Get("/b", option.Response(200, TypeB{}))

		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()

		assert.Contains(t, doc.Components.Schemas, "Collision")
		assert.Contains(t, doc.Components.Schemas, "Collision2")
	})

	t.Run("DefaultDefNameUsesPkgPrefixExceptCallerPackage", func(t *testing.T) {
		r := spec.NewRouter()

		type SamePkgModel struct{ Foo string }
		r.Get("/same", option.Response(200, SamePkgModel{}))
		r.Get("/other", option.Response(200, dto.Pet{}))

		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()

		assert.Contains(t, doc.Components.Schemas, "SamePkgModel")
		assert.Contains(t, doc.Components.Schemas, "DtoPet")
	})

	t.Run("StripDefNamePrefixCanStripGeneratedPkgPrefix", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithReflectorConfig(
				option.StripDefNamePrefix("Dto"),
			),
		)

		r.Get("/other", option.Response(200, dto.Pet{}))

		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()

		assert.Contains(t, doc.Components.Schemas, "Pet")
		assert.NotContains(t, doc.Components.Schemas, "DtoPet")
	})
}

func TestReflector_NilConfig(t *testing.T) {
	r := reflect.NewReflector(&openapi.Config{})
	assert.Nil(t, r.StripPrefixes())
	assert.False(t, r.InlineRefs())
}

func TestReflector_StripPrefixes(t *testing.T) {
	cfg := &openapi.Config{
		ReflectorConfig: &openapi.ReflectorConfig{
			StripDefNamePrefix: []string{"prefix_"},
		},
	}
	r := reflect.NewReflector(cfg)
	assert.Equal(t, []string{"prefix_"}, r.StripPrefixes())
}

func TestReflector_ParameterField_CustomMappingKeepsDefaultTag(t *testing.T) {
	cfg := option.WithOpenAPIConfig(
		option.WithReflectorConfig(option.ParameterTagMapping(openapi.ParameterInPath, "param")),
	)
	r := reflect.NewReflector(cfg)

	type Request struct {
		ID int `path:"id" required:"true"`
	}

	params, _, err := r.RequestParts(Request{}, "")
	require.NoError(t, err)
	require.Len(t, params, 1)
	assert.Equal(t, "id", params[0].Name)
	assert.Equal(t, "path", params[0].In)
	assert.True(t, params[0].Required)
}

func TestReflector_BodyTagMapping(t *testing.T) {
	t.Run("ParameterInBody overrides json tag", func(t *testing.T) {
		cfg := option.WithOpenAPIConfig(
			option.WithReflectorConfig(option.ParameterTagMapping(openapi.ParameterInBody, "api")),
		)
		r := reflect.NewReflector(cfg)

		type Req struct {
			ID   string `path:"id"`
			Name string `api:"name"`
		}

		params, body, err := r.RequestParts(Req{}, "application/json")
		require.NoError(t, err)
		require.Len(t, params, 1)
		require.NotNil(t, body)
		assert.Contains(t, body.Properties, "name")
		assert.NotContains(t, body.Properties, "Name")
	})

	t.Run("ParameterInForm overrides form tag", func(t *testing.T) {
		cfg := option.WithOpenAPIConfig(
			option.WithReflectorConfig(option.ParameterTagMapping(openapi.ParameterInForm, "api")),
		)
		r := reflect.NewReflector(cfg)

		type Req struct {
			ID    string `path:"id"`
			Email string `api:"email"`
		}

		params, body, err := r.RequestParts(Req{}, "application/x-www-form-urlencoded")
		require.NoError(t, err)
		require.Len(t, params, 1)
		require.NotNil(t, body)
		assert.Contains(t, body.Properties, "email")
	})

	t.Run("default body tags unchanged when no mapping", func(t *testing.T) {
		cfg := option.WithOpenAPIConfig()
		r := reflect.NewReflector(cfg)

		type Req struct {
			ID   string `path:"id"`
			Name string `json:"name"`
		}

		params, body, err := r.RequestParts(Req{}, "application/json")
		require.NoError(t, err)
		require.Len(t, params, 1)
		require.NotNil(t, body)
		assert.Contains(t, body.Properties, "name")
	})
}

func TestReflector_RequestPartsAndStructSchemaBranches(t *testing.T) {
	cfg := option.WithOpenAPIConfig()
	r := reflect.NewReflector(cfg)

	t.Run("non-struct uses schema component", func(t *testing.T) {
		params, body, err := r.RequestParts(123, "")
		require.NoError(t, err)
		assert.Nil(t, params)
		require.NotNil(t, body)
		assert.Equal(t, "integer", body.Type)
	})

	t.Run("only params without body", func(t *testing.T) {
		type Req struct {
			ID string `path:"id" required:"true"`
		}
		params, body, err := r.RequestParts(Req{}, "")
		require.NoError(t, err)
		require.Len(t, params, 1)
		assert.Equal(t, "id", params[0].Name)
		assert.Nil(t, body)
	})

	t.Run("params with explicit body field", func(t *testing.T) {
		type Req struct {
			ID   string `path:"id" required:"true"`
			Name string `json:"name"`
		}
		params, body, err := r.RequestParts(Req{}, "application/json")
		require.NoError(t, err)
		require.Len(t, params, 1)
		require.NotNil(t, body)
		assert.Contains(t, body.Properties, "name")
	})

	t.Run("body tag for form media type", func(t *testing.T) {
		type Req struct {
			ID    string `path:"id" required:"true"`
			Email string `form:"email"`
		}
		params, body, err := r.RequestParts(Req{}, "application/x-www-form-urlencoded")
		require.NoError(t, err)
		require.Len(t, params, 1)
		require.NotNil(t, body)
		assert.Contains(t, body.Properties, "email")
	})

	t.Run("type mapping applied before request analysis", func(t *testing.T) {
		type Src struct {
			ID string `path:"id" required:"true"`
		}
		type Dst struct {
			Name string `json:"name"`
		}
		cfg := option.WithOpenAPIConfig(option.WithReflectorConfig(option.TypeMapping(Src{}, Dst{})))
		rr := reflect.NewReflector(cfg)
		params, body, err := rr.RequestParts(Src{}, "")
		require.NoError(t, err)
		assert.Nil(t, params)
		require.NotNil(t, body)
		assert.Equal(t, "#/components/schemas/Dst", body.Ref)
	})
}

func TestStructSchema_InterceptProp(t *testing.T) {
	type Payload struct {
		Name   string `json:"name"`
		Secret string `json:"secret"`
	}

	t.Run("PreHookSkipsProperty", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if !params.Processed && params.Name == "secret" {
						return openapi.ErrSkipProperty
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Properties, "name")
		assert.NotContains(t, schema.Properties, "secret")
	})

	t.Run("PostHookSkipsProperty", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if params.Processed && params.Name == "secret" {
						return openapi.ErrSkipProperty
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Properties, "name")
		assert.NotContains(t, schema.Properties, "secret")
	})

	t.Run("PostHookModifiesPropertySchema", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if params.Processed && params.Name == "name" {
						params.PropertySchema.Description = "intercepted"
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.NoError(t, err)
		require.Contains(t, schema.Properties, "name")
		assert.Equal(t, "intercepted", schema.Properties["name"].Description)
		assert.Empty(t, schema.Properties["secret"].Description)
	})

	t.Run("CallOrderProcessedFalseBeforeTrue", func(t *testing.T) {
		var calls []bool
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					calls = append(calls, params.Processed)
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, _ = r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.Len(t, calls, 4) // 2 fields × (pre + post)
		assert.False(t, calls[0])
		assert.True(t, calls[1])
		assert.False(t, calls[2])
		assert.True(t, calls[3])
	})

	t.Run("PostHookSkipAlsoRemovesFromRequired", func(t *testing.T) {
		type WithRequired struct {
			Name   string `json:"name" required:"true"`
			Secret string `json:"secret" required:"true"`
		}
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if params.Processed && params.Name == "secret" {
						return openapi.ErrSkipProperty
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue(WithRequired{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "name")
		assert.NotContains(t, schema.Required, "secret")
		assert.NotContains(t, schema.Properties, "secret")
	})
}

func TestReflector_RequiredPropByValidateTag(t *testing.T) {
	type Form struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"email"`
		Age   int    `json:"age"`
	}

	t.Run("DefaultTagMarksRequired", func(t *testing.T) {
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.RequiredPropByValidateTag(),
		)))
		schema, err := r.SchemaForValue(Form{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "name")
		assert.NotContains(t, schema.Required, "email")
		assert.NotContains(t, schema.Required, "age")
	})

	t.Run("CustomTagName", func(t *testing.T) {
		type BindingForm struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"email"`
		}
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.RequiredPropByValidateTag("binding"),
		)))
		schema, err := r.SchemaForValue(BindingForm{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "name")
		assert.NotContains(t, schema.Required, "email")
	})

	t.Run("CustomSeparator", func(t *testing.T) {
		type PipeForm struct {
			Name  string `json:"name" validate:"required|min=3"`
			Email string `json:"email" validate:"email"`
		}
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.RequiredPropByValidateTag("validate", "|"),
		)))
		schema, err := r.SchemaForValue(PipeForm{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "name")
		assert.NotContains(t, schema.Required, "email")
	})

	t.Run("NoValidateTagNotRequired", func(t *testing.T) {
		type Plain struct {
			Name string `json:"name"`
		}
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.RequiredPropByValidateTag(),
		)))
		schema, err := r.SchemaForValue(Plain{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Empty(t, schema.Required)
	})

	t.Run("BothRequiredTagAndValidateTagNoDuplicate", func(t *testing.T) {
		type Overlap struct {
			Name string `json:"name" required:"true" validate:"required"`
		}
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.RequiredPropByValidateTag(),
		)))
		schema, err := r.SchemaForValue(Overlap{}, reflect.SchemaInline)
		require.NoError(t, err)
		count := 0
		for _, req := range schema.Required {
			if req == "name" {
				count++
			}
		}
		assert.Equal(t, 1, count, "name must appear exactly once in Required")
	})
}

//nolint:gocognit // large integration test covering all InterceptSchema edge cases
func TestStructSchema_InterceptSchema(t *testing.T) {
	t.Run("PreHookStopReturnCustomSchema", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(params openapi.InterceptSchemaParams) (bool, error) {
					if !params.Processed && params.Type == std_reflect.TypeFor[int]() {
						params.Schema.Type = "string"
						params.Schema.Format = "uuid"
						return true, nil
					}
					return false, nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue(0, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Equal(t, "string", schema.Type)
		assert.Equal(t, "uuid", schema.Format)
	})

	t.Run("PostHookModifiesPrimitiveSchema", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(params openapi.InterceptSchemaParams) (bool, error) {
					if params.Processed && params.Type == std_reflect.TypeFor[string]() {
						params.Schema.Description = "intercepted"
					}
					return false, nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		schema, err := r.SchemaForValue("", reflect.SchemaInline)
		require.NoError(t, err)
		assert.Equal(t, "intercepted", schema.Description)
	})

	t.Run("PostHookModifiesComponentSchema", func(t *testing.T) {
		type Item struct {
			Name string `json:"name"`
		}
		r := spec.NewRouter(option.WithReflectorConfig(
			option.InterceptSchema(func(params openapi.InterceptSchemaParams) (bool, error) {
				if params.Processed && params.Type == std_reflect.TypeFor[Item]() {
					if params.Schema.Extensions == nil {
						params.Schema.Extensions = map[string]any{}
					}
					params.Schema.Extensions["x-intercepted"] = true
				}
				return false, nil
			}),
		))
		r.Get("/items", option.Response(200, Item{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		require.Contains(t, doc.Components.Schemas, "Item")
		assert.Equal(t, true, doc.Components.Schemas["Item"].Extensions["x-intercepted"])
	})

	t.Run("PreHookStopOnComponentSkipsStructSchema", func(t *testing.T) {
		type Skipped struct {
			Name string `json:"name"`
		}
		r := spec.NewRouter(option.WithReflectorConfig(
			option.InterceptSchema(func(params openapi.InterceptSchemaParams) (bool, error) {
				if !params.Processed && params.Type == std_reflect.TypeFor[Skipped]() {
					params.Schema.Type = "object"
					params.Schema.Description = "custom"
					return true, nil
				}
				return false, nil
			}),
		))
		r.Get("/", option.Response(200, Skipped{}))
		_, err := r.GenerateSchema("yaml")
		require.NoError(t, err)
		doc := r.Document()
		require.Contains(t, doc.Components.Schemas, "Skipped")
		assert.Equal(t, "custom", doc.Components.Schemas["Skipped"].Description)
		assert.Nil(t, doc.Components.Schemas["Skipped"].Properties) // StructSchema was skipped
	})

	t.Run("PreHookErrorPropagated", func(t *testing.T) {
		boom := errors.New("hook error")
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(params openapi.InterceptSchemaParams) (bool, error) {
					if !params.Processed && params.Type == std_reflect.TypeFor[int]() {
						return false, boom
					}
					return false, nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForValue(0, reflect.SchemaInline)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("PostHookErrorPropagated", func(t *testing.T) {
		boom := errors.New("post hook error")
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(params openapi.InterceptSchemaParams) (bool, error) {
					if params.Processed && params.Type == std_reflect.TypeFor[string]() {
						return false, boom
					}
					return false, nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForValue("", reflect.SchemaInline)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("ChainingBothHooksFire", func(t *testing.T) {
		var fired []string
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.InterceptSchema(func(params openapi.InterceptSchemaParams) (bool, error) {
				if params.Processed {
					fired = append(fired, "first")
				}
				return false, nil
			}),
			option.InterceptSchema(func(params openapi.InterceptSchemaParams) (bool, error) {
				if params.Processed {
					fired = append(fired, "second")
				}
				return false, nil
			}),
		)))
		_, err := r.SchemaForValue("", reflect.SchemaInline)
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second"}, fired)
	})

	t.Run("ChainingStopShortCircuits", func(t *testing.T) {
		secondFired := false
		r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
			option.InterceptSchema(func(_ openapi.InterceptSchemaParams) (bool, error) {
				return true, nil // stop immediately
			}),
			option.InterceptSchema(func(_ openapi.InterceptSchemaParams) (bool, error) {
				secondFired = true
				return false, nil
			}),
		)))
		_, err := r.SchemaForValue(0, reflect.SchemaInline)
		require.NoError(t, err)
		assert.False(t, secondFired)
	})

	t.Run("StopAndErrorErrorWins", func(t *testing.T) {
		boom := errors.New("stop and error")
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(_ openapi.InterceptSchemaParams) (bool, error) {
					return true, boom // both stop and error
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForValue(0, reflect.SchemaInline)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("SchemaExposerPreHookStop", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptSchema: func(params openapi.InterceptSchemaParams) (bool, error) {
					if !params.Processed {
						params.Schema.Type = "string"
						params.Schema.Format = "override"
						return true, nil
					}
					return false, nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		// SchemaExposerType implements OpenAPISchema — without the fix only post-hook fired.
		schema, err := r.SchemaForValue(SchemaExposerType{}, reflect.SchemaInline)
		require.NoError(t, err)
		assert.Equal(t, "string", schema.Type)
		assert.Equal(t, "override", schema.Format)
	})

	t.Run("ComponentCleanedUpOnPreHookError", func(t *testing.T) {
		type Target struct{ Name string }
		boom := errors.New("pre-hook fail")
		calls := 0
		cfg := &openapi.Config{
			OpenAPIVersion:  openapi.Version312,
			ReflectorConfig: &openapi.ReflectorConfig{},
		}
		cfg.ReflectorConfig.InterceptSchema = func(params openapi.InterceptSchemaParams) (bool, error) {
			if !params.Processed && params.Type == std_reflect.TypeFor[Target]() {
				calls++
				if calls == 1 {
					return false, boom
				}
			}
			return false, nil
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForType(std_reflect.TypeFor[Target](), reflect.SchemaUseComponent, nil)
		require.ErrorIs(t, err, boom)
		// Second call must retry (not hit stale empty component from first call).
		schema, err := r.SchemaForType(std_reflect.TypeFor[Target](), reflect.SchemaUseComponent, nil)
		require.NoError(t, err)
		assert.Equal(t, "#/components/schemas/Target", schema.Ref)
		assert.Equal(t, 2, calls)
	})
}

// Ensure InterceptProp wires through spec.NewRouter to StructSchema.
func TestReflector_InterceptPropViaRouter(t *testing.T) {
	_ = spec.NewRouter // import guard
	type Item struct {
		Name   string `json:"name"`
		Hidden string `json:"hidden"`
	}
	r := spec.NewRouter(option.WithReflectorConfig(
		option.InterceptProp(func(params openapi.InterceptPropParams) error {
			if params.Processed && params.Name == "hidden" {
				return openapi.ErrSkipProperty
			}
			return nil
		}),
	))
	r.Get("/items", option.Response(200, Item{}))
	_, err := r.GenerateSchema("yaml")
	require.NoError(t, err)
	doc := r.Document()
	require.Contains(t, doc.Components.Schemas, "Item")
	assert.Contains(t, doc.Components.Schemas["Item"].Properties, "name")
	assert.NotContains(t, doc.Components.Schemas["Item"].Properties, "hidden")
}

func TestStructSchema_InterceptProp_NonSkipErrorPropagated(t *testing.T) {
	type Payload struct {
		Name string `json:"name"`
	}
	hookErr := errors.New("hook internal error")

	t.Run("PreHookErrorPropagated", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if !params.Processed {
						return hookErr
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.Error(t, err)
		assert.ErrorIs(t, err, hookErr)
	})

	t.Run("PostHookErrorPropagated", func(t *testing.T) {
		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				InterceptProp: func(params openapi.InterceptPropParams) error {
					if params.Processed {
						return hookErr
					}
					return nil
				},
			},
		}
		r := reflect.NewReflector(cfg)
		_, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
		require.Error(t, err)
		assert.ErrorIs(t, err, hookErr)
	})
}

func TestInterceptProp_Chaining(t *testing.T) {
	type Payload struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Secret string `json:"secret"`
	}

	var callLog []string
	r := reflect.NewReflector(option.WithOpenAPIConfig(option.WithReflectorConfig(
		option.InterceptProp(func(params openapi.InterceptPropParams) error {
			if params.Processed {
				callLog = append(callLog, "hook1:"+params.Name)
			}
			return nil
		}),
		option.InterceptProp(func(params openapi.InterceptPropParams) error {
			if params.Processed && params.Name == "secret" {
				return openapi.ErrSkipProperty
			}
			if params.Processed {
				callLog = append(callLog, "hook2:"+params.Name)
			}
			return nil
		}),
	)))

	schema, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
	require.NoError(t, err)

	// Both hooks fired for non-skipped fields.
	assert.Contains(t, callLog, "hook1:name")
	assert.Contains(t, callLog, "hook2:name")
	assert.Contains(t, callLog, "hook1:email")
	assert.Contains(t, callLog, "hook2:email")

	// Hook1 fired for secret (before hook2 returned ErrSkipProperty).
	assert.Contains(t, callLog, "hook1:secret")

	// secret must be absent because hook2 returned ErrSkipProperty.
	assert.NotContains(t, schema.Properties, "secret")
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "email")
}

func TestStructSchema_InterceptProp_PostHookSkipRestoresParentSnapshot(t *testing.T) {
	type Payload struct {
		Name   string `json:"name"`
		Secret string `json:"secret"`
	}

	cfg := &openapi.Config{
		ReflectorConfig: &openapi.ReflectorConfig{
			InterceptProp: func(params openapi.InterceptPropParams) error {
				if params.Processed && params.Name == "secret" {
					// Mutate parent before returning ErrSkipProperty.
					params.ParentSchema.AllOf = append(params.ParentSchema.AllOf, &openapi.Schema{Type: "object"})
					params.ParentSchema.Extensions = map[string]any{"x-dirty": true}
					return openapi.ErrSkipProperty
				}
				return nil
			},
		},
	}
	r := reflect.NewReflector(cfg)
	schema, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
	require.NoError(t, err)

	// secret must be excluded.
	assert.NotContains(t, schema.Properties, "secret")

	// Mutations made to ParentSchema before returning ErrSkipProperty must be rolled back.
	assert.Empty(t, schema.AllOf, "AllOf must be restored after ErrSkipProperty")
	assert.Nil(t, schema.Extensions, "Extensions must be restored after ErrSkipProperty")
}

func TestStructSchema_InterceptProp_ParentType(t *testing.T) {
	type Embedded struct {
		EmbedField string `json:"embedField"`
	}
	type Payload struct {
		Embedded

		Name string `json:"name"`
	}

	var observedTypes []std_reflect.Type
	cfg := &openapi.Config{
		ReflectorConfig: &openapi.ReflectorConfig{
			InterceptProp: func(params openapi.InterceptPropParams) error {
				if !params.Processed {
					observedTypes = append(observedTypes, params.ParentType)
				}
				return nil
			},
		},
	}
	r := reflect.NewReflector(cfg)
	_, err := r.SchemaForValue(Payload{}, reflect.SchemaInline)
	require.NoError(t, err)

	// Should have observed two pre-hook calls (embedField + name).
	require.Len(t, observedTypes, 2)

	// ParentType must always be the top-level struct type, not the embedded struct type.
	expectedType := std_reflect.TypeFor[Payload]()
	for _, pt := range observedTypes {
		assert.Equal(t, expectedType, pt, "ParentType must always be the top-level struct type")
	}
}
