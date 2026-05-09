package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestValidateParameterRules(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Invalid Parameters"),
		option.WithVersion("1.0.0"),
	)
	r.Get("/items",
		option.CustomizeOperation(func(op *openapi.Operation) {
			op.Parameters = append(op.Parameters,
				&openapi.Parameter{Name: "q", In: "query", Schema: &openapi.Schema{Type: "string"}},
				&openapi.Parameter{Name: "raw", In: "querystring", Schema: &openapi.Schema{Type: "string"}},
			)
		}),
		option.Response(200, new([]User)),
	)

	err := r.Validate()
	assertValidationContains(t, err,
		"querystring parameter must use content",
		"querystring parameter content is required",
		"must not mix query and querystring parameters",
	)
}

func TestValidateParameterRequiresSchemaOrContent(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Invalid Parameters"),
		option.WithVersion("1.0.0"),
	)
	r.Get("/items",
		option.CustomizeOperation(func(op *openapi.Operation) {
			op.Parameters = append(op.Parameters, &openapi.Parameter{Name: "q", In: "query"})
		}),
		option.Response(200, new([]User)),
	)

	err := r.Validate()
	assertValidationContains(t, err, "parameters[0] must define schema or content")
}

func TestValidateMediaTypeEncodingRestrictions(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Invalid Media Encoding"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["/encoded"].Post.RequestBody = &openapi.RequestBody{
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema:         &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}},
						ItemSchema:     &openapi.Schema{Type: "string"},
						Encoding:       map[string]*openapi.Encoding{"field": {ContentType: "text/plain"}},
						PrefixEncoding: []*openapi.Encoding{{ContentType: "text/plain"}},
						ItemEncoding:   &openapi.Encoding{ContentType: "text/plain"},
					},
				},
			}
		}),
	)
	r.Post("/encoded", option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err,
		"encoding requires multipart or application/x-www-form-urlencoded media type",
		"prefixEncoding requires multipart media type",
		"itemEncoding requires multipart media type",
	)
}

func TestValidateMediaTypeEncodingAllowsFormAndMultipart(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Media Encoding"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["/form"].Post.RequestBody = &openapi.RequestBody{
				Content: map[string]openapi.MediaType{
					"application/x-www-form-urlencoded": {
						Schema:   &openapi.Schema{Type: "object"},
						Encoding: map[string]*openapi.Encoding{"field": {ContentType: "text/plain"}},
					},
					"multipart/mixed": {
						Schema:         &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}},
						PrefixEncoding: []*openapi.Encoding{{ContentType: "text/plain"}},
						ItemEncoding:   &openapi.Encoding{ContentType: "text/plain"},
					},
				},
			}
		}),
	)
	r.Post("/form", option.Response(204, nil))

	assert.NoError(t, r.Validate())
}

func TestValidateOperationIDAndSecurityReferences(t *testing.T) {
	r := spec.NewRouter(option.WithTitle("Invalid Operations"), option.WithVersion("1.0.0"))
	r.Get("/a", option.OperationID("same"), option.Security("missing"), option.Response(204, nil))
	r.Get("/b", option.OperationID("same"), option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err,
		`operationId "same" duplicates`,
		`references undefined security scheme "missing"`,
	)
}

func TestValidateCallbackObject(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Callbacks"),
		option.WithVersion("1.0.0"),
	)
	r.Post("/subscriptions",
		option.CustomizeOperation(func(op *openapi.Operation) {
			op.Callbacks = map[string]*openapi.Callback{
				"onEvent": {
					Expressions: map[string]*openapi.PathItem{
						"{$request.body#/callbackUrl}": {
							Post: &openapi.Operation{
								Responses: map[string]*openapi.Response{
									"200": {Description: "Callback accepted"},
								},
							},
						},
					},
				},
			}
		}),
		option.Response(202, nil),
	)

	raw, err := r.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"onEvent"`)
	assert.Contains(t, string(raw), `"{$request.body#/callbackUrl}"`)
}

func TestValidateHeaderAndLinkRules(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Invalid Header and Link"),
		option.WithVersion("1.0.0"),
		option.WithDocument(func(doc *openapi.Document) {
			resp := doc.Paths["/invalid"].Get.Responses["200"]
			resp.Headers = map[string]*openapi.Header{
				"X-Bad": {
					Schema:          &openapi.Schema{Type: "string"},
					Style:           "form",
					AllowEmptyValue: true,
				},
			}
			resp.Links = map[string]*openapi.Link{
				"missingTarget": {Description: "No operation target"},
			}
		}),
	)
	r.Get("/invalid", option.Response(200, nil))

	err := r.Validate()
	assertValidationContains(t, err,
		"headers.X-Bad allowEmptyValue is not allowed for headers",
		"headers.X-Bad.style must be simple for headers",
		"links.missingTarget must define operationRef or operationId",
	)
}

func TestValidate_OpenAPI320_ExampleSerializedValue(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Examples"),
		option.WithVersion("1.0.0"),
	)
	r.Get("/examples", option.Response(200, "",
		option.ContentType("text/plain"),
		option.ContentNamedExample("encoded", nil,
			option.ExampleDataValue("hello world"),
			option.ExampleSerializedValue("hello%20world"),
		),
	))

	raw, err := r.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"serializedValue": "hello%20world"`)
	assert.NotContains(t, string(raw), "serializedExample")
}

func TestValidateRejectsDeprecatedSerializedExampleField(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithDocument(func(doc *openapi.Document) {
			mt := openapi.MediaType{
				Schema: &openapi.Schema{Type: "string"},
				Examples: map[string]*openapi.Example{
					"old": {
						SerializedExample: "hello",
					},
				},
			}
			doc.Paths["/examples"].Get.Responses["200"].Content = map[string]openapi.MediaType{"text/plain": mt}
		}),
	)
	r.Get("/examples", option.Response(200, nil))

	err := r.Validate()
	assertValidationContains(t, err, "serializedExample is not an OpenAPI field; use serializedValue")
}

func TestValidationEdgeCases(t *testing.T) {
	t.Run("Headers", func(t *testing.T) {
		r := spec.NewRouter(option.WithDocument(func(doc *openapi.Document) {
			resp := doc.Paths["/headers"].Get.Responses["200"]
			resp.Headers = map[string]*openapi.Header{
				"X-Custom": {
					Description: "A custom header",
					Schema:      &openapi.Schema{Type: "string"},
				},
			}
		}))
		r.Get("/headers", option.Response(200, nil))
		assert.NoError(t, r.Validate())
	})

	t.Run("Examples", func(t *testing.T) {
		r := spec.NewRouter(option.WithDocument(func(doc *openapi.Document) {
			mt := doc.Paths["/examples"].Get.Responses["200"].Content["application/json"]
			mt.Examples = map[string]*openapi.Example{
				"test": {
					Summary: "A test example",
					Value:   map[string]string{"foo": "bar"},
				},
			}
			doc.Paths["/examples"].Get.Responses["200"].Content["application/json"] = mt
		}))
		r.Get("/examples", option.Response(200, nil))
		assert.NoError(t, r.Validate())
	})

	t.Run("Links", func(t *testing.T) {
		r := spec.NewRouter(option.WithDocument(func(doc *openapi.Document) {
			resp := doc.Paths["/links"].Get.Responses["200"]
			resp.Links = map[string]*openapi.Link{
				"userLink": {
					OperationID: "getUser",
					Description: "Link to user",
				},
			}
		}))
		r.Get("/links", option.Response(200, nil))
		assert.NoError(t, r.Validate())
	})
}

func TestValidateEncoding(t *testing.T) {
	t.Run("OpenAPI304_InvalidFields", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Post("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.RequestBody = &openapi.RequestBody{
				Content: map[string]openapi.MediaType{
					"multipart/form-data": {
						Encoding: map[string]*openapi.Encoding{
							"field": {
								PrefixEncoding: []*openapi.Encoding{{}},
								ItemEncoding:   &openapi.Encoding{},
							},
						},
					},
				},
			}
		}), option.Response(204, nil))

		err := r.Validate()
		assertValidationContains(t, err,
			"prefixEncoding requires OpenAPI 3.2.0",
			"itemEncoding requires OpenAPI 3.2.0",
		)
	})
}

func TestValidateExample_EdgeCases(t *testing.T) {
	t.Run("VersionRestrictions", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Get("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Responses["200"] = &openapi.Response{
				Description: "OK",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Examples: map[string]*openapi.Example{
							"ex": {
								DataValue:       "data",
								SerializedValue: "serialized",
							},
						},
					},
				},
			}
		}))

		err := r.Validate()
		assertValidationContains(t, err,
			"dataValue requires OpenAPI 3.2.0",
			"serializedValue requires OpenAPI 3.2.0",
		)
	})

	t.Run("MutuallyExclusive", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version320))
		r.Get("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Responses["200"] = &openapi.Response{
				Description: "OK",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Examples: map[string]*openapi.Example{
							"ex": {
								Value:         "v",
								ExternalValue: "ev",
								DataValue:     "dv",
							},
						},
					},
				},
			}
		}))

		err := r.Validate()
		assertValidationContains(t, err,
			"value and externalValue are mutually exclusive",
			"dataValue and value are mutually exclusive",
		)
	})
}

func TestValidateCallback_Errors(t *testing.T) {
	r := spec.NewRouter()
	r.Post("/sub", option.CustomizeOperation(func(op *openapi.Operation) {
		op.Callbacks = map[string]*openapi.Callback{
			"cb": {Expressions: map[string]*openapi.PathItem{}},
		}
	}), option.Response(204, nil))

	err := r.Validate()
	assertValidationContains(t, err, "must define at least one callback expression")
}

func TestValidateRequestBody_Errors(t *testing.T) {
	t.Run("OpenAPI304_RefSiblings", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Post("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.RequestBody = &openapi.RequestBody{
				Ref:         "#/components/requestBodies/Body",
				Description: "sibling",
			}
		}), option.Response(204, nil))
		err := r.Validate()
		assertValidationContains(t, err, "requestBody must not define siblings with $ref")
	})

	t.Run("EmptyContent", func(t *testing.T) {
		r := spec.NewRouter()
		r.Post("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.RequestBody = &openapi.RequestBody{Content: map[string]openapi.MediaType{}}
		}), option.Response(204, nil))
		err := r.Validate()
		assertValidationContains(t, err, "requestBody.content is required")
	})
}

func TestValidateHeader_Errors(t *testing.T) {
	t.Run("RequiredField", func(t *testing.T) {
		r := spec.NewRouter()
		r.Get("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Responses["200"] = &openapi.Response{
				Description: "OK",
				Headers: map[string]*openapi.Header{
					"X-Trace": {},
				},
			}
		}))
		err := r.Validate()
		assertValidationContains(t, err, "headers.X-Trace must define schema or content")
	})
}

func TestValidateLink_Errors(t *testing.T) {
	t.Run("MultipleTargets", func(t *testing.T) {
		r := spec.NewRouter()
		r.Get("/test", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Responses["200"] = &openapi.Response{
				Description: "OK",
				Links: map[string]*openapi.Link{
					"bad": {OperationID: "op", OperationRef: "/ref"},
				},
			}
		}))
		err := r.Validate()
		assertValidationContains(t, err, "links.bad operationRef and operationId are mutually exclusive")
	})
}

func TestValidateParameterSerializationHelper(t *testing.T) {
	assert.True(t, validate.ValidParameterStyle("path", "matrix", openapi.Version304))
	assert.True(t, validate.ValidParameterStyle("query", "deepObject", openapi.Version304))
	assert.True(t, validate.ValidParameterStyle("header", "simple", openapi.Version304))
	assert.True(t, validate.ValidParameterStyle("cookie", "form", openapi.Version312))
	assert.False(t, validate.ValidParameterStyle("cookie", "cookie", openapi.Version312))
	assert.True(t, validate.ValidParameterStyle("cookie", "cookie", openapi.Version320))
	assert.False(t, validate.ValidParameterStyle("unknown", "form", openapi.Version320))

	errs := validate.ValidateParameterSerialization("parameters[0]", &openapi.Parameter{
		In:              string(openapi.ParameterInHeader),
		AllowEmptyValue: true,
		Style:           "form",
	}, openapi.Version304)
	assert.Len(t, errs, 2)

	errs = validate.ValidateParameterSerialization("parameters[1]", &openapi.Parameter{
		In: string(openapi.ParameterInQuery),
	}, openapi.Version304)
	assert.Empty(t, errs)
}
