package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/internal/validate"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func TestValidatePathParameterReferences(t *testing.T) {
	r := spec.NewRouter(
		option.WithTitle("Referenced Path Parameters"),
		option.WithVersion("1.0.0"),
		option.WithComponentParameter("UserID", &openapi.Parameter{
			Name:     "id",
			In:       "path",
			Required: true,
			Schema:   &openapi.Schema{Type: "string"},
		}),
	)
	r.Get("/users/{id}",
		option.CustomizeOperation(func(op *openapi.Operation) {
			op.Parameters = append(op.Parameters, &openapi.Parameter{Ref: "#/components/parameters/UserID"})
		}),
		option.Response(200, new(User)),
	)

	assert.NoError(t, r.Validate())
}

func TestValidatePathItem_Errors(t *testing.T) {
	t.Run("NoLeadingSlash", func(t *testing.T) {
		r := spec.NewRouter(option.WithDocument(func(doc *openapi.Document) {
			doc.Paths["invalid"] = &openapi.PathItem{
				Get: &openapi.Operation{Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
			}
		}))
		err := r.Validate()
		assertValidationContains(t, err, `path "invalid" must start with /`)
	})

	t.Run("QUERY_OpenAPI312", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithOpenAPIVersion(openapi.Version312),
			option.WithDocument(func(doc *openapi.Document) {
				doc.Paths["/search"] = &openapi.PathItem{
					Query: &openapi.Operation{Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
				}
			}),
		)
		err := r.Validate()
		assertValidationContains(t, err, "QUERY operation at /search requires OpenAPI 3.2.0")
	})

	t.Run("AdditionalOperations_FixedMethod", func(t *testing.T) {
		r := spec.NewRouter(
			option.WithOpenAPIVersion(openapi.Version320),
			option.WithDocument(func(doc *openapi.Document) {
				doc.Paths["/test"] = &openapi.PathItem{
					AdditionalOperations: map[string]*openapi.Operation{
						"GET": {Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
					},
				}
			}),
		)
		err := r.Validate()
		assertValidationContains(t, err, "additionalOperations at /test must not contain fixed method GET")
	})
}

func TestIsFixedMethodAndIsValidParameterIn(t *testing.T) {
	assert.True(t, validate.IsFixedMethod("GET"))
	assert.True(t, validate.IsFixedMethod("patch"))
	assert.True(t, validate.IsFixedMethod("QUERY"))
	assert.False(t, validate.IsFixedMethod("PURGE"))

	assert.True(t, validate.IsValidParameterIn("query"))
	assert.True(t, validate.IsValidParameterIn("querystring"))
	assert.False(t, validate.IsValidParameterIn("body"))
}

func TestValidatePathParams_Direct(t *testing.T) {
	errs := validate.ValidatePathParams("/users/{id}", "get", []*openapi.Parameter{
		{Name: "other", In: "path", Required: true},
	})
	assert.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), `path parameter "other" must match a path template`)
	assert.Contains(t, errs[1].Error(), `missing path parameter "id"`)

	errs = validate.ValidatePathParams("/users/{id}", "get", []*openapi.Parameter{
		{Name: "id", In: "path", Required: false},
	})
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), `path parameter "id" must be required`)

	errs = validate.ValidatePathParams("users/{id}", "get", []*openapi.Parameter{
		{Name: "id", In: "path", Required: true},
	})
	assert.Empty(t, errs)
}

func TestHasParameterRefSiblings(t *testing.T) {
	assert.True(t, validate.HasParameterRefSiblings(&openapi.Parameter{Summary: "s"}, openapi.Version304))
	assert.True(t, validate.HasParameterRefSiblings(&openapi.Parameter{Name: "n"}, openapi.Version312))
	assert.False(t, validate.HasParameterRefSiblings(&openapi.Parameter{}, openapi.Version312))
}

func TestValidatePathItemOperations_AdditionalOpsRequires320(t *testing.T) {
	item := &openapi.PathItem{
		AdditionalOperations: map[string]*openapi.Operation{
			"PURGE": {Responses: map[string]*openapi.Response{"200": {Description: "OK"}}},
		},
	}
	errs := validate.ValidatePathItemOperations("/cache", item, openapi.Version312, map[string]string{}, nil, nil)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "additionalOperations at /cache requires OpenAPI 3.2.0")
}
