# oaswrap/spec

[![CI](https://github.com/oaswrap/spec/actions/workflows/ci.yml/badge.svg)](https://github.com/oaswrap/spec/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/oaswrap/spec/graph/badge.svg?token=RIEIM9BAIW)](https://codecov.io/gh/oaswrap/spec)
[![Go Reference](https://pkg.go.dev/badge/github.com/oaswrap/spec.svg)](https://pkg.go.dev/github.com/oaswrap/spec)
[![Go Report Card](https://goreportcard.com/badge/github.com/oaswrap/spec)](https://goreportcard.com/report/github.com/oaswrap/spec)
[![Go Version](https://img.shields.io/github/go-mod/go-version/oaswrap/spec)](https://github.com/oaswrap/spec/blob/main/go.mod)
[![License](https://img.shields.io/github/license/oaswrap/spec)](LICENSE)

`spec` is a self-contained OpenAPI generator for Go. It generates OpenAPI `3.0.x`, `3.1.x`, and `3.2.0` documents through a router and functional options API.

The library owns its OpenAPI model and schema reflection implementation. It does not import OpenAPI or JSON Schema generator libraries. YAML serialization uses `github.com/goccy/go-yaml`; docs UI integration and tests bring additional dependencies.

## Why oaswrap/spec?

- **Native OpenAPI builder**: OpenAPI paths, operations, components, validation, and schema reflection are implemented in this repository.
- **Framework agnostic core**: Use `spec.NewRouter` for static generation, or use adapters for Chi, Echo, Gin, Fiber, net/http, and Mux.
- **Code-first route documentation**: Register routes and documentation together with Go functions and typed options.
- **Version-aware output**: Generate OpenAPI `3.0.4` by default, with support for `3.1.2` and `3.2.0` features when selected.
- **Direct model escape hatches**: Use typed `openapi` structs, `Extensions` for `x-*` fields, and `Extra` for official or future fields not wrapped by helper options yet.
- **Golden-file friendly**: The generated document is deterministic enough for snapshot tests and CI documentation checks.

## Features

- JSON and YAML generation.
- Framework-agnostic route registration with `NewRouter`, groups, route builders, and HTTP method helpers.
- Webhook registration helpers for OpenAPI `3.1.x` and `3.2.0`.
- OpenAPI `3.2.0` `QUERY` method and additional operation support.
- Request and response schema reflection from Go structs.
- Parameter reflection from `path`, `query`, `header`, `cookie`, and OpenAPI `3.2.0` `querystring` tags.
- JSON Schema/OpenAPI schema tags for examples, enums, constraints, nullability, read/write flags, content metadata, and XML metadata.
- Security helpers for API key, HTTP bearer/basic-style schemes, OAuth2, OpenID Connect, and mutual TLS.
- Duplicate response merging into `oneOf`.
- Low-level typed OpenAPI model for direct field control.
- `spec.OneOf` for explicit one-of schemas.
- `SchemaExposer` and `StaticSchemaExposer` hooks for custom reflected schemas.

## Installation

```bash
go get github.com/oaswrap/spec
```

## Quick Start

```go
package main

import (
	"log"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

func main() {
	r := spec.NewRouter(
		option.WithOpenAPIVersion(openapi.Version320),
		option.WithTitle("Users API"),
		option.WithVersion("1.0.0"),
		option.WithServer("https://api.example.com"),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("bearer")),
	)

	v1 := r.Group("/api/v1", option.GroupTags("users"))
	v1.Post("/login",
		option.Summary("Login"),
		option.Request(new(LoginRequest)),
		option.Response(200, new(LoginResponse)),
	)
	v1.Get("/users/{id}",
		option.Summary("Get user"),
		option.Request(new(GetUserRequest)),
		option.Response(200, new(User)),
	)

	if err := r.WriteSchemaTo("openapi.yaml"); err != nil {
		log.Fatal(err)
	}
}

type LoginRequest struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true" writeOnly:"true"`
}

type LoginResponse struct {
	Token string `json:"token" required:"true"`
}

type GetUserRequest struct {
	ID string `path:"id" required:"true" description:"User identifier"`
}

type User struct {
	ID   string `json:"id" required:"true"`
	Name string `json:"name"`
}
```

## Output

- `GenerateSchema()` defaults to YAML.
- `GenerateSchema("yaml")` and `GenerateSchema("yml")` return YAML.
- `GenerateSchema("json")` returns pretty JSON.
- `MarshalYAML()` validates and serializes YAML.
- `MarshalJSON()` validates and serializes pretty JSON.
- `WriteSchemaTo("openapi.yaml")`, `WriteSchemaTo("openapi.yml")`, and `WriteSchemaTo("openapi.json")` infer format from extension.
- `Document()` returns the built `*openapi.Document`.
- `Validate()` builds the document and validates OpenAPI invariants.
- `Config()` returns the effective OpenAPI configuration.

## Framework Adapters

| Framework | Package |
| --- | --- |
| Chi | [`chiopenapi`](/adapter/chiopenapi) |
| Echo v4 | [`echoopenapi`](/adapter/echoopenapi) |
| Echo v5 | [`echov5openapi`](/adapter/echov5openapi) |
| Fiber v2 | [`fiberopenapi`](/adapter/fiberopenapi) |
| Fiber v3 | [`fiberv3openapi`](/adapter/fiberv3openapi) |
| Gin | [`ginopenapi`](/adapter/ginopenapi) |
| net/http | [`httpopenapi`](/adapter/httpopenapi) |
| Mux | [`muxopenapi`](/adapter/muxopenapi) |

Use the core `spec` package for static generation and CI workflows. Use adapters when you want route registration, spec generation, and docs UI wiring in one framework-specific router.

## OpenAPI Options

Root options are passed to `spec.NewRouter` or `spec.NewGenerator`.

```go
r := spec.NewRouter(
	option.WithOpenAPIVersion(openapi.Version312),
	option.WithTitle("Payments API"),
	option.WithInfoSummary("Payments API"),
	option.WithVersion("1.4.0"),
	option.WithDescription("Payment operations."),
	option.WithTermsOfService("https://example.com/terms"),
	option.WithContact(openapi.Contact{Name: "API Team", Email: "api@example.com"}),
	option.WithLicense(openapi.License{Name: "Apache 2.0", URL: "https://www.apache.org/licenses/LICENSE-2.0.html"}),
	option.WithExternalDocs("https://docs.example.com", "Full documentation"),
	option.WithServer("https://api.example.com", option.ServerDescription("Production")),
	option.WithTag("payments", option.TagDescription("Payment operations")),
	option.WithStripTrailingSlash(),
)
```

| Option | Purpose |
| --- | --- |
| `WithOpenAPIConfig(opts...)` | Build an `*openapi.Config` with defaults and apply options. |
| `WithOpenAPIVersion(version)` | Set `openapi`; default is `openapi.Version304`. Constants are available for `3.0.0` through `3.0.4`, `3.1.0` through `3.1.2`, and `3.2.0`. |
| `WithSelf(uri)` | Set OpenAPI `3.2.0` `$self`. |
| `WithJSONSchemaDialect(uri)` | Set root `jsonSchemaDialect`. |
| `WithTitle(title)` | Set `info.title`. |
| `WithInfoSummary(summary)` | Set `info.summary`. |
| `WithVersion(version)` | Set `info.version`. |
| `WithDescription(description)` | Set `info.description`. |
| `WithContact(contact)` | Set `info.contact`. |
| `WithLicense(license)` | Set `info.license`. |
| `WithTermsOfService(url)` | Set `info.termsOfService`. |
| `WithTags(tags...)` | Add root `tags` from `openapi.Tag` values. |
| `WithTag(name, tagOpts...)` | Add one root tag without constructing `openapi.Tag` manually. |
| `WithServer(url, opts...)` | Add a root server. |
| `WithExternalDocs(url, description...)` | Set root `externalDocs`. |
| `WithSecurity(name, opts...)` | Add a reusable security scheme. |
| `WithGlobalSecurity(name, scopes...)` | Add a root security requirement. |
| `WithReflectorConfig(opts...)` | Configure schema reflection. |
| `WithStripTrailingSlash(strip...)` | Trim trailing slashes from operation paths except `/`. |
| `WithPathParser(parser)` | Convert framework paths, for example `:id` to `{id}`. |
| `WithDocument(fn)` | Mutate the final low-level document before validation and serialization. |
| `WithComponentSchema(name, schema)` | Register a reusable schema component. |
| `WithComponentResponse(name, response)` | Register a reusable response component. |
| `WithComponentParameter(name, parameter)` | Register a reusable parameter component. |
| `WithComponentExample(name, example)` | Register a reusable example component. |
| `WithComponentRequestBody(name, requestBody)` | Register a reusable request body component. |
| `WithComponentHeader(name, header)` | Register a reusable header component. |
| `WithComponentSecurityScheme(name, scheme)` | Register a reusable security scheme component. |
| `WithComponentLink(name, link)` | Register a reusable link component. |
| `WithComponentCallback(name, callback)` | Register a reusable callback component. |
| `WithComponentPathItem(name, pathItem)` | Register a reusable path item component. |
| `WithComponentMediaType(name, mediaType)` | Register a reusable media type component for OpenAPI `3.2.0`. |
| `WithDisableDocs(disable...)` | Disable adapter docs endpoints. |
| `WithDocsPath(path)` | Set adapter docs UI path. |
| `WithSpecPath(path)` | Set adapter OpenAPI spec path. |
| `WithCacheAge(cacheAge)` | Set docs/spec cache age. |
| `WithUIOption(opt)` | Set a low-level `spec-ui` option. |
| `WithSwaggerUI(cfg...)` | Use Swagger UI. |
| `WithStoplightElements(cfg...)` | Use Stoplight Elements. |
| `WithReDoc(cfg...)` | Use ReDoc. |
| `WithScalar(cfg...)` | Use Scalar. |
| `WithRapiDoc(cfg...)` | Use RapiDoc. |

Tag options:

- `TagSummary(summary)`
- `TagDescription(description)`
- `TagExternalDocs(url, description...)`
- `TagParent(parent)` for OpenAPI `3.2.0`
- `TagKind(kind)` for OpenAPI `3.2.0`

Server options:

- `ServerDescription(description)`
- `ServerVariables(variables)`

## Routes And Groups

```go
api := r.Group("/api/v1", option.GroupTags("v1"))

api.Get("/users/{id}",
	option.OperationID("getUser"),
	option.Summary("Get user"),
	option.Description("Returns one user."),
	option.Tags("users"),
	option.Security("bearerAuth"),
	option.Request(new(GetUserRequest)),
	option.Response(200, new(User), option.ContentDescription("OK")),
	option.Response(404, nil, option.ContentDescription("Not Found")),
)
```

Router methods:

- `Get(path, opts...)`
- `Post(path, opts...)`
- `Put(path, opts...)`
- `Delete(path, opts...)`
- `Patch(path, opts...)`
- `Options(path, opts...)`
- `Head(path, opts...)`
- `Trace(path, opts...)`
- `Query(path, opts...)` for OpenAPI `3.2.0`
- `Add(method, path, opts...)`
- `Webhook(name, opts...)` using POST by default, for OpenAPI `3.1.x+`
- `AddWebhook(method, name, opts...)` for OpenAPI `3.1.x+`
- `NewRoute(opts...).Method(method).Path(path).With(opts...)`
- `Group(prefix, opts...)`
- `Route(prefix, func(router), opts...)`
- `With(groupOpts...)`

Group options:

| Option | Purpose |
| --- | --- |
| `GroupTags(tags...)` | Apply tags to routes in the group. |
| `GroupSecurity(name, scopes...)` | Apply operation security to routes in the group. |
| `GroupDeprecated(deprecated...)` | Mark routes in the group deprecated. |
| `GroupHidden(hide...)` | Hide routes in the group. |

Operation options:

| Option | Purpose |
| --- | --- |
| `OperationID(id)` | Set `operationId`. |
| `Summary(summary)` | Set `summary`; also sets `description` when description is empty. |
| `Description(description)` | Set `description`. |
| `ExternalDocs(url, description...)` | Set operation `externalDocs`. |
| `Tags(tags...)` | Add operation tags. |
| `Security(name, scopes...)` | Add operation security requirement. |
| `Deprecated(deprecated...)` | Mark operation deprecated. |
| `Hidden(hide...)` | Skip operation from generated document. |
| `Request(structure, contentOpts...)` | Add parameters and/or request body from a Go value. |
| `Response(status, structure, contentOpts...)` | Add a response. Duplicate status/content pairs are merged into `oneOf`. |
| `CustomizeOperation(fn)` | Mutate the low-level `openapi.Operation`. |

Content options:

| Option | Purpose |
| --- | --- |
| `ContentType(contentType)` | Set media type. Default is `application/json`. |
| `ContentDescription(description)` | Set request/response description. |
| `ContentDefault(isDefault...)` | Mark response as `default`. |
| `ContentEncoding(prop, enc)` | Add media type encoding metadata for a property. |
| `ContentExample(value)` | Set media type `example`. |
| `ContentNamedExample(name, value, opts...)` | Add one named media type example. |
| `ContentExamples(examples)` | Set media type `examples`. |
| `ContentRequired(required...)` | Mark request body required. |
| `ContentFormat(format)` | Override reflected content schema format. |

Example options:

- `ExampleSummary(summary)`
- `ExampleDescription(description)`
- `ExampleExternalValue(url)`
- `ExampleDataValue(value)` for OpenAPI `3.2.0`
- `ExampleSerializedValue(value)` for OpenAPI `3.2.0`

## Security

```go
r := spec.NewRouter(
	option.WithSecurity("apiKey", option.SecurityAPIKey("X-API-Key", openapi.SecuritySchemeAPIKeyInHeader)),
	option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("bearer")),
	option.WithSecurity("mtls", option.SecurityMutualTLS()),
	option.WithSecurity("oauth2", option.SecurityOAuth2AuthorizationCode(
		"https://auth.example.com/oauth/authorize",
		"https://auth.example.com/oauth/token",
		map[string]string{
			"read":  "Read access",
			"write": "Write access",
		},
	)),
	option.WithSecurity("oidc", option.SecurityOpenIDConnect("https://auth.example.com/.well-known/openid-configuration")),
)
```

Security helpers:

- `SecurityAPIKey(name, in)`
- `SecurityHTTPBearer(scheme, bearerFormat...)`
- `SecurityOAuth2(flows)`
- `SecurityOAuth2Implicit(authorizationURL, scopes, flowOpts...)`
- `SecurityOAuth2Password(tokenURL, scopes, flowOpts...)`
- `SecurityOAuth2ClientCredentials(tokenURL, scopes, flowOpts...)`
- `SecurityOAuth2AuthorizationCode(authorizationURL, tokenURL, scopes, flowOpts...)`
- `SecurityOAuth2DeviceAuthorization(deviceAuthorizationURL, tokenURL, scopes, flowOpts...)` for OpenAPI `3.2.0`
- `OAuthRefreshURL(url)`
- `SecurityOpenIDConnect(url)`
- `SecurityMutualTLS()`
- `SecurityDescription(description)`
- `SecurityOAuth2MetadataURL(url)` for OpenAPI `3.2.0`
- `SecurityDeprecated(deprecated...)` for OpenAPI `3.2.0`

## Reflection Tags

Request structs are split into parameters and request body:

- Fields with `path`, `query`, `header`, `cookie`, or `querystring` tags become OpenAPI parameters.
- `querystring` is only valid for OpenAPI `3.2.0`.
- Body fields use `json` names by default.
- For `application/x-www-form-urlencoded` and `multipart/form-data`, body fields use `form` names and fall back to `json`.
- `json:"-"` skips a field.
- Path parameters are always marked required.
- Adapter packages can add framework-specific parameter tags with `ParameterTagMapping`; for example Gin uses `uri`, Fiber uses `params`, and Echo uses `param`.

```go
type SearchRequest struct {
	ID       string `path:"id" required:"true" description:"Resource ID"`
	Status   string `query:"status" enum:"active,inactive"`
	TraceID  string `header:"X-Trace-ID"`
	Session  string `cookie:"session"`
	Name     string `json:"name" minLength:"2" maxLength:"80"`
	File     []byte `form:"file" format:"binary" description:"Upload file"`
	Internal string `json:"-"`
}
```

Naming and location tags:

| Tag | Purpose |
| --- | --- |
| `json:"name"` | Body/schema property name. |
| `path:"name"` | Path parameter. |
| `query:"name"` | Query parameter. |
| `header:"name"` | Header parameter. |
| `cookie:"name"` | Cookie parameter. |
| `querystring:"name"` | OpenAPI `3.2.0` whole-query-string parameter. |
| `form:"name"` | Form body property name for form content types. |

Schema tags:

| Tag | Output |
| --- | --- |
| `required:"true"` | Adds property to parent `required`; path parameters are required automatically. |
| `type:"string,null"` | Overrides `type`; comma-separated unions are emitted only for OpenAPI `3.1.x` and `3.2.0`. |
| `title:"..."` | `title`. |
| `description:"..."` | `description`. |
| `format:"email"` | `format`. |
| `pattern:"..."` | `pattern`. |
| `default:"..."` | `default`; JSON values are decoded when valid. |
| `example:"..."` | `example`; JSON values are decoded when valid. |
| `examples:"[...]"` | `examples` for OpenAPI `3.1.x` and `3.2.0`; JSON array preferred, comma-separated fallback. |
| `enum:"a,b,c"` | `enum`; JSON array or comma-separated values. |
| `const:"..."` | `const` for OpenAPI `3.1.x` and `3.2.0`. |
| `multipleOf:"2"` | `multipleOf`. |
| `maximum:"100"` | `maximum`. |
| `minimum:"1"` | `minimum`. |
| `exclusiveMaximum:"true"` | OpenAPI `3.0.x` boolean `exclusiveMaximum`; OpenAPI `3.1.x`/`3.2.0` numeric value. |
| `exclusiveMinimum:"true"` | OpenAPI `3.0.x` boolean `exclusiveMinimum`; OpenAPI `3.1.x`/`3.2.0` numeric value. |
| `maxLength:"80"` | `maxLength`. |
| `minLength:"2"` | `minLength`. |
| `maxItems:"10"` | `maxItems`. |
| `minItems:"1"` | `minItems`. |
| `maxProperties:"10"` | `maxProperties`. |
| `minProperties:"1"` | `minProperties`. |
| `uniqueItems:"true"` | `uniqueItems`. |
| `nullable:"true"` | `nullable` for OpenAPI `3.0.x`; `type: [T, null]` for OpenAPI `3.1.x`/`3.2.0` when possible. |
| `deprecated:"true"` | `deprecated`. |
| `readOnly:"true"` | `readOnly`. |
| `writeOnly:"true"` | `writeOnly`. |
| `contentEncoding:"base64"` | `contentEncoding` for OpenAPI `3.1.x` and `3.2.0`. |
| `contentMediaType:"image/png"` | `contentMediaType` for OpenAPI `3.1.x` and `3.2.0`. |
| `xmlName:"name"` | XML object `name`. |
| `xmlNamespace:"uri"` | XML object `namespace`. |
| `xmlPrefix:"p"` | XML object `prefix`. |
| `xmlAttribute:"true"` | XML object `attribute`; skipped when OpenAPI `3.2.0` `xmlNodeType` is set. |
| `xmlWrapped:"true"` | XML object `wrapped`; skipped when OpenAPI `3.2.0` `xmlNodeType` is set. |
| `xmlNodeType:"element"` | OpenAPI `3.2.0` XML object `nodeType`. |

Reflection intentionally avoids emitting keywords that are invalid for the selected OpenAPI version. For example, `const`, `examples`, `contentEncoding`, and `contentMediaType` are not emitted for OpenAPI `3.0.x`.

## Reflected Go Types

| Go type | Schema |
| --- | --- |
| `bool` | `type: boolean` |
| signed ints except `int64` | `type: integer`, `format: int32` |
| `int64` | `type: integer`, `format: int64` |
| unsigned ints except `uint64` and `uintptr` | `type: integer`, `format: int32`, `minimum: 0` |
| `uint64`, `uintptr` | `type: integer`, `format: int64`, `minimum: 0` |
| `float32` | `type: number`, `format: float` |
| `float64` | `type: number`, `format: double` |
| `string` | `type: string` |
| `time.Time` | `type: string`, `format: date-time` |
| `[]T`, `[N]T` | `type: array`, `items: T` |
| `[]byte` | OpenAPI `3.0.x`: `type: string`, `format: byte`; OpenAPI `3.1.x`/`3.2.0`: `type: string`, `contentEncoding: base64` |
| `map[string]T` | `type: object`, `additionalProperties: T` |
| structs | `type: object`, `properties` |
| named structs in component mode | `#/components/schemas/{TypeName}` references |
| pointers | nullable schema behavior |

Custom types can expose their own schema when tags are not expressive enough:

```go
type Slug string

func (*Slug) OpenAPISchema(version string) *openapi.Schema {
	if version == openapi.Version304 {
		return &openapi.Schema{Type: "string", Format: "slug"}
	}
	return &openapi.Schema{Type: []string{"string", "null"}, Format: "slug"}
}
```

For static schemas, implement `OpenAPISchema() *openapi.Schema` instead. Field tags are still applied on top of custom schemas.

## Reflector Configuration

```go
r := spec.NewRouter(
	option.WithReflectorConfig(
		option.InlineRefs(),
		option.StripDefNamePrefix("DTO"),
		option.TypeMapping(NullString{}, new(string)),
		option.ParameterTagMapping(openapi.ParameterInPath, "params"),
		option.InterceptDefName(func(t reflect.Type, defaultName string) string {
			return defaultName
		}),
	),
)
```

| Option | Purpose |
| --- | --- |
| `InlineRefs(inline...)` | Inline schemas instead of using components for named structs. |
| `StripDefNamePrefix(prefixes...)` | Strip prefixes from generated component names. |
| `TypeMapping(src, dst)` | Reflect `src` as `dst`. |
| `ParameterTagMapping(in, sourceTag)` | Add a custom tag for a parameter location while keeping the default tag. |
| `InterceptDefName(fn)` | Customize schema component names. |

## OpenAPI 3.2

OpenAPI `3.2.0` enables:

- `Router.Query(path, opts...)`.
- custom HTTP methods through `Add`, emitted as `additionalOperations`.
- `querystring` parameter tags.
- root `$self`.
- tag `parent` and `kind`.
- security scheme metadata and deprecation fields.
- `components.mediaTypes`.
- media type and encoding fields such as `itemSchema`, `prefixEncoding`, and `itemEncoding`.
- example `dataValue` and `serializedValue`.
- XML `nodeType`.

## Low-Level OpenAPI Control

The `openapi` package exposes owned low-level OpenAPI structs. Use them when a feature does not need reflection or does not have a convenience option yet.

```go
r := spec.NewRouter(
	option.WithOpenAPIVersion(openapi.Version320),
	option.WithTitle("API"),
	option.WithVersion("1.0.0"),
	option.WithDocument(func(doc *openapi.Document) {
		doc.Webhooks = map[string]*openapi.PathItem{
			"user.created": {
				Post: &openapi.Operation{
					Responses: map[string]*openapi.Response{
						"202": {Description: "Accepted"},
					},
				},
			},
		}
		doc.Extensions = map[string]any{"x-company": "oaswrap"}
		doc.Extra = map[string]any{"futureOfficialField": true}
	}),
)
```

Use:

- `Extensions` for `x-*` specification extensions.
- `Extra` to emit official or future fields not typed yet.
- `CustomizeOperation` to mutate an operation directly.
- `WithDocument` to mutate the final document directly.

The library validates many OpenAPI invariants, but it does not attempt exhaustive semantic validation for every official rule.

## Examples

Explore complete working examples in the [`examples/`](examples/) directory:

- **[Basic](examples/basic/)**: standalone spec generation.
- **[Petstore](examples/petstore/)**: full Petstore API with routes and models.

## API Reference

Complete documentation is available at [pkg.go.dev/github.com/oaswrap/spec](https://pkg.go.dev/github.com/oaswrap/spec).

Key packages:

- [`spec`](https://pkg.go.dev/github.com/oaswrap/spec): core router and spec builder.
- [`openapi`](https://pkg.go.dev/github.com/oaswrap/spec/openapi): owned OpenAPI model.
- [`option`](https://pkg.go.dev/github.com/oaswrap/spec/option): configuration options.
- [`pkg/parser`](https://pkg.go.dev/github.com/oaswrap/spec/pkg/parser): path parsers such as `NewColonParamParser`.

## Contributing

Issues and pull requests are welcome. Please check existing issues and discussions before starting work on new features.

## License

[MIT](LICENSE)
