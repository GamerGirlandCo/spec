package validate_test

import (
	"strings"
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
						Schema: &openapi.Schema{Type: "object", Properties: map[string]*openapi.Schema{
							"field": {Type: "string"},
						}},
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

func TestValidate_URIFields(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("URI Fields"),
		option.WithVersion("1.0.0"),
		option.WithTermsOfService("not a uri"),
		option.WithContact(openapi.Contact{URL: "not a uri", Email: "api@example.com"}),
		option.WithLicense(openapi.License{Name: "MIT", URL: "not a uri"}),
		option.WithExternalDocs("not a uri"),
		option.WithSecurity("oidc", option.SecurityOpenIDConnect("not a uri")),
		option.WithSecurity("oauth2", option.SecurityOAuth2AuthorizationCode(
			"not a uri",
			"also not a uri",
			map[string]string{"read": "Read access"},
			option.OAuthRefreshURL("also not a uri"),
		)),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["/uri"].Get.Responses["200"].Links = map[string]*openapi.Link{
				"follow": {
					OperationRef: "not a uri",
				},
			}
		}),
	)
	r.Get("/uri",
		option.ExternalDocs("not a uri"),
		option.Response(200, "",
			option.ContentType("text/plain"),
			option.ContentNamedExample("remote", nil, option.ExampleExternalValue("not a uri")),
		),
	)

	err := r.Validate()
	assertValidationContains(t, err,
		"info.termsOfService must be a URI",
		"info.contact.url must be a URI",
		"info.license.url must be a URI",
		"externalDocs.url must be a URI",
		"GET /uri.externalDocs.url must be a URI",
		"components.securitySchemes.oidc.openIdConnectUrl must be an HTTPS URI without a fragment",
		"components.securitySchemes.oauth2.flows.authorizationCode.authorizationUrl must be a URI",
		"components.securitySchemes.oauth2.flows.authorizationCode.tokenUrl must be a URI",
		"components.securitySchemes.oauth2.flows.authorizationCode.refreshUrl must be a URI",
		"GET /uri.responses.200.content.text/plain.examples.remote.externalValue must be a URI",
		"GET /uri.responses.200.links.follow.operationRef must be a URI reference",
	)
}

func TestValidate_URIFields_AllowsRelativeReferences(t *testing.T) {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version312),
		option.WithTitle("Relative URI Fields"),
		option.WithVersion("1.0.0"),
		option.WithTermsOfService("terms"),
		option.WithContact(openapi.Contact{URL: "../contact", Email: "api@example.com"}),
		option.WithLicense(openapi.License{Name: "MIT", URL: "./license"}),
		option.WithExternalDocs("../docs"),
		option.WithSecurity("oidc",
			option.SecurityOpenIDConnect("https://example.com/.well-known/openid-configuration")),
		option.WithSecurity("oauth2", option.SecurityOAuth2AuthorizationCode(
			"/authorize",
			"/token",
			map[string]string{"read": "Read access"},
			option.OAuthRefreshURL("/refresh"),
		)),
		option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["/uri"].Get.Responses["200"].Links = map[string]*openapi.Link{
				"follow": {
					OperationRef: "#/paths/~1uri/get",
				},
			}
		}),
	)
	r.Get("/uri",
		option.ExternalDocs("../operation-docs"),
		option.Response(200, "",
			option.ContentType("text/plain"),
			option.ContentNamedExample("remote", nil, option.ExampleExternalValue("examples/remote.txt")),
		),
	)

	assert.NoError(t, r.Validate())
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

	t.Run("NilExpressionPathItem", func(t *testing.T) {
		r := spec.NewRouter()
		r.Post("/sub", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Callbacks = map[string]*openapi.Callback{
				"cb": {
					Expressions: map[string]*openapi.PathItem{
						"{$request.body#/callbackUrl}": nil,
					},
				},
			}
		}), option.Response(204, nil))

		err := r.Validate()
		assertValidationContains(t, err, `callbacks.cb.{$request.body#/callbackUrl} is required`)
	})

	t.Run("NilCallbackAndRefSiblings", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version304))
		r.Post("/sub", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Callbacks = map[string]*openapi.Callback{
				"nilCb": nil,
				"refCb": {
					Ref:     "#/components/callbacks/Base",
					Summary: "legacy-sibling",
				},
			}
		}), option.Response(204, nil))

		err := r.Validate()
		assertValidationContains(t, err,
			"callbacks.nilCb is required",
			"callbacks.refCb must not define siblings with $ref",
		)
	})
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

	t.Run("ComprehensiveHeaderRules", func(t *testing.T) {
		r := spec.NewRouter(option.WithOpenAPIVersion(openapi.Version312))
		r.Get("/headers", option.CustomizeOperation(func(op *openapi.Operation) {
			op.Responses["200"] = &openapi.Response{
				Description: "OK",
				Headers: map[string]*openapi.Header{
					"X-Mixed": {
						Summary: "only-for-ref",
						Schema:  &openapi.Schema{Type: "string"},
						Content: map[string]*openapi.MediaType{
							"text/plain":       {Schema: &openapi.Schema{Type: "string"}},
							"application/json": {Schema: &openapi.Schema{Type: "string"}},
						},
						Example:       "one",
						Examples:      map[string]*openapi.Example{"two": {Value: "two"}},
						AllowReserved: true,
					},
					"X-Ref": {
						Ref:      "#/components/headers/TraceID",
						Required: true,
					},
				},
			}
		}))

		err := r.Validate()
		assertValidationContains(t, err,
			"headers.X-Mixed.summary is only allowed with $ref",
			"headers.X-Mixed schema and content are mutually exclusive",
			"headers.X-Mixed content must contain only one media type",
			"headers.X-Mixed example and examples are mutually exclusive",
			"headers.X-Mixed allowReserved is not allowed for headers",
			"headers.X-Ref must not define siblings with $ref",
		)
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

func TestValidate_MutualTLSVersionGate(t *testing.T) {
	t.Run("3.0.x rejects mutualTLS", func(t *testing.T) {
		errs := validate.ValidateSecurityScheme("securitySchemes.mtls", &openapi.SecurityScheme{
			Type: "mutualTLS",
		}, openapi.Version304)
		assertHasError(t, errs, "mutualTLS requires OpenAPI 3.1.x or 3.2.0")
	})

	t.Run("3.1.x allows mutualTLS", func(t *testing.T) {
		errs := validate.ValidateSecurityScheme("securitySchemes.mtls", &openapi.SecurityScheme{
			Type: "mutualTLS",
		}, openapi.Version312)
		assertNoStrictErrors(t, errs)
	})

	t.Run("3.2.0 allows mutualTLS", func(t *testing.T) {
		errs := validate.ValidateSecurityScheme("securitySchemes.mtls", &openapi.SecurityScheme{
			Type: "mutualTLS",
		}, openapi.Version320)
		assertNoStrictErrors(t, errs)
	})
}

func TestValidate_ContentTypeHeader_IsWarning(t *testing.T) {
	resp := &openapi.Response{
		Description: "OK",
		Headers: map[string]*openapi.Header{
			"Content-Type": {Schema: &openapi.Schema{Type: "string"}},
		},
	}
	errs := validate.ValidateResponse("responses.200", resp, openapi.Version304)
	assertHasWarning(t, errs, "is ignored by the OpenAPI spec")
	assertNoStrictErrors(t, errs)
}

func TestValidate_EncodingSchemaProperties(t *testing.T) {
	t.Run("3.1.x errors when encoding key not in schema", func(t *testing.T) {
		mt := &openapi.MediaType{
			Schema: &openapi.Schema{
				Type: "object",
				Properties: map[string]*openapi.Schema{
					"file": {Type: "string"},
				},
			},
			Encoding: map[string]*openapi.Encoding{
				"unknown": {ContentType: "application/octet-stream"},
			},
		}
		errs := validate.ValidateMediaType(
			"requestBody.content.multipart/form-data", "multipart/form-data", mt, openapi.Version312)
		assertHasError(t, errs, "must correspond to a schema property")
	})

	t.Run("3.2.0 errors when encoding key not in schema", func(t *testing.T) {
		mt := &openapi.MediaType{
			Schema: &openapi.Schema{
				Type: "object",
				Properties: map[string]*openapi.Schema{
					"file": {Type: "string"},
				},
			},
			Encoding: map[string]*openapi.Encoding{
				"unknown": {ContentType: "application/octet-stream"},
			},
		}
		errs := validate.ValidateMediaType(
			"requestBody.content.multipart/form-data", "multipart/form-data", mt, openapi.Version320)
		assertHasError(t, errs, "must correspond to a schema property")
	})

	t.Run("encoding key matching schema property is valid", func(t *testing.T) {
		mt := &openapi.MediaType{
			Schema: &openapi.Schema{
				Type: "object",
				Properties: map[string]*openapi.Schema{
					"file": {Type: "string"},
				},
			},
			Encoding: map[string]*openapi.Encoding{
				"file": {ContentType: "application/octet-stream"},
			},
		}
		errs := validate.ValidateMediaType(
			"requestBody.content.multipart/form-data", "multipart/form-data", mt, openapi.Version320)
		assertNoStrictErrors(t, errs)
	})
}

func TestValidate_AllowReserved_Header(t *testing.T) {
	t.Run("pre-3.2.0 rejects allowReserved in headers", func(t *testing.T) {
		for _, version := range []string{openapi.Version304, openapi.Version312} {
			hdr := &openapi.Header{Schema: &openapi.Schema{Type: "string"}, AllowReserved: true}
			errs := validate.ValidateHeader("headers.X-Custom", hdr, version)
			assertHasError(t, errs, "allowReserved is not allowed for headers")
		}
	})

	t.Run("3.2.0 allows allowReserved in headers", func(t *testing.T) {
		hdr := &openapi.Header{Schema: &openapi.Schema{Type: "string"}, AllowReserved: true}
		errs := validate.ValidateHeader("headers.X-Custom", hdr, openapi.Version320)
		assertNoStrictErrors(t, errs)
	})
}

func TestValidate_AllowEmptyValue_Deprecated320(t *testing.T) {
	t.Run("3.2.0 warns when allowEmptyValue is used in query param", func(t *testing.T) {
		param := &openapi.Parameter{
			Name:            "q",
			In:              "query",
			AllowEmptyValue: true,
			Schema:          &openapi.Schema{Type: "string"},
		}
		errs := validate.ValidateParameters("op", []*openapi.Parameter{param}, openapi.Version320, nil)
		assertHasWarning(t, errs, "allowEmptyValue is deprecated in OpenAPI 3.2.0")
	})

	t.Run("pre-3.2.0 does not warn for allowEmptyValue", func(t *testing.T) {
		param := &openapi.Parameter{
			Name:            "q",
			In:              "query",
			AllowEmptyValue: true,
			Schema:          &openapi.Schema{Type: "string"},
		}
		errs := validate.ValidateParameters("op", []*openapi.Parameter{param}, openapi.Version312, nil)
		for _, e := range errs {
			if strings.Contains(e.Error(), "deprecated") {
				t.Fatalf("unexpected deprecation warning for 3.1.x: %v", e)
			}
		}
	})
}

func TestValidate_EncodingHeaders_ContentType(t *testing.T) {
	t.Run("Content-Type in encoding.headers emits a warning", func(t *testing.T) {
		enc := &openapi.Encoding{
			Headers: map[string]*openapi.Header{
				"Content-Type": {Schema: &openapi.Schema{Type: "string"}},
			},
		}
		errs := validate.ValidateEncoding(
			"requestBody.content.multipart/form-data.encoding.file",
			"multipart/form-data",
			enc,
			openapi.Version312,
		)
		assertHasWarning(t, errs, "is described separately and is ignored")
	})
}
