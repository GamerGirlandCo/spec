package option

import "github.com/oaswrap/spec/openapi"

// ContentOption mutates request/response content generation.
type ContentOption func(*openapi.ContentUnit)

// ContentType sets media type, for example `application/json`.
func ContentType(contentType string) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.ContentType = contentType }
}

// ContentDescription sets request-body/response description.
func ContentDescription(description string) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Description = description }
}

// ContentSummary sets request-body/response summary.
// It is only valid for OpenAPI 3.2.0.
func ContentSummary(summary string) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Summary = summary }
}

// ContentDefault marks a response as the `default` response.
func ContentDefault(isDefault ...bool) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.IsDefault = optional(true, isDefault...) }
}

// ContentEncoding sets encoding metadata for a body property.
func ContentEncoding(prop, enc string) ContentOption {
	return func(cu *openapi.ContentUnit) {
		if cu.Encoding == nil {
			cu.Encoding = map[string]string{}
		}
		cu.Encoding[prop] = enc
	}
}

// ContentExample sets the media type example value.
func ContentExample(value any) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Example = value }
}

// ExampleOption mutates an example object.
type ExampleOption func(*openapi.Example)

// ContentNamedExample adds one named media type example.
func ContentNamedExample(name string, value any, opts ...ExampleOption) ContentOption {
	return func(cu *openapi.ContentUnit) {
		if cu.Examples == nil {
			cu.Examples = map[string]*openapi.Example{}
		}
		example := &openapi.Example{Value: value}
		for _, opt := range opts {
			opt(example)
		}
		cu.Examples[name] = example
	}
}

// ContentExamples sets the media type examples map.
func ContentExamples(examples map[string]*openapi.Example) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Examples = examples }
}

// ExampleSummary sets example summary.
func ExampleSummary(summary string) ExampleOption {
	return func(example *openapi.Example) { example.Summary = summary }
}

// ExampleDescription sets example description.
func ExampleDescription(description string) ExampleOption {
	return func(example *openapi.Example) { example.Description = description }
}

// ExampleExternalValue sets example externalValue and clears value fields.
func ExampleExternalValue(url string) ExampleOption {
	return func(example *openapi.Example) {
		example.Value = nil
		example.DataValue = nil
		example.ExternalValue = url
	}
}

// ExampleDataValue sets example dataValue and clears other value fields.
// `dataValue` is only valid for OpenAPI 3.2.0.
func ExampleDataValue(value any) ExampleOption {
	return func(example *openapi.Example) {
		example.Value = nil
		example.ExternalValue = ""
		example.DataValue = value
	}
}

// ExampleSerializedValue sets example serializedValue and clears value fields.
// `serializedValue` is only valid for OpenAPI 3.2.0.
func ExampleSerializedValue(value string) ExampleOption {
	return func(example *openapi.Example) {
		example.Value = nil
		example.DataValue = nil
		example.SerializedValue = value
	}
}

// ContentRequired marks request body as required.
func ContentRequired(required ...bool) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Required = optional(true, required...) }
}

// ContentFormat sets schema format for reflected payload.
func ContentFormat(format string) ContentOption {
	return func(cu *openapi.ContentUnit) { cu.Format = format }
}
