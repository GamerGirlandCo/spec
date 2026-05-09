package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec"
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
