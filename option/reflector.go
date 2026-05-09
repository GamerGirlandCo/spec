package option

import (
	"reflect"

	"github.com/oaswrap/spec/openapi"
)

// ReflectorOption mutates schema reflection behavior.
type ReflectorOption func(*openapi.ReflectorConfig)

// InlineRefs toggles inlining referenced schemas.
func InlineRefs(inline ...bool) ReflectorOption {
	return func(cfg *openapi.ReflectorConfig) { cfg.InlineRefs = optional(true, inline...) }
}

// StripDefNamePrefix appends prefixes removed from reflected definition names.
func StripDefNamePrefix(prefixes ...string) ReflectorOption {
	return func(cfg *openapi.ReflectorConfig) {
		cfg.StripDefNamePrefix = append(cfg.StripDefNamePrefix, prefixes...)
	}
}

// InterceptDefName sets callback to customize reflected definition names.
func InterceptDefName(fn func(t reflect.Type, defaultDefName string) string) ReflectorOption {
	return func(cfg *openapi.ReflectorConfig) { cfg.InterceptDefName = fn }
}

// TypeMapping maps source type to destination type during reflection.
func TypeMapping(src, dst any) ReflectorOption {
	return func(cfg *openapi.ReflectorConfig) {
		cfg.TypeMappings = append(cfg.TypeMappings, openapi.TypeMapping{Src: src, Dst: dst})
	}
}

// ParameterTagMapping overrides tag source for a specific parameter location.
func ParameterTagMapping(in openapi.ParameterIn, sourceTag string) ReflectorOption {
	return func(cfg *openapi.ReflectorConfig) {
		if cfg.ParameterTagMapping == nil {
			cfg.ParameterTagMapping = map[openapi.ParameterIn]string{}
		}
		cfg.ParameterTagMapping[in] = sourceTag
	}
}
