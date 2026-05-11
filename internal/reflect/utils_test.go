package reflect

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/oaswrap/spec/openapi"
)

func TestSanitizeTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"Foo", "Foo"},
		{"*Foo", "Foo"},
		{"[]Foo", "FooList"},
		{"[][]Foo", "FooListList"},
		{"github.com/foo.User", "User"},
		{"BaseResponse[github.com/foo.User]", "BaseResponseUser"},
		{"Map[string, int]", "Mapstringint"},
		{"Complex[[]string, *int]", "ComplexstringListint"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, SanitizeTypeName(tt.input))
		})
	}
}

func TestLowerCamel(t *testing.T) {
	assert.Equal(t, "fooBar", LowerCamel("FooBar"))
	assert.Equal(t, "foo", LowerCamel("Foo"))
	assert.Empty(t, LowerCamel(""))
}

func TestIndirectType(t *testing.T) {
	type Foo struct{}
	// f := Foo{}
	typ := reflect.TypeFor[Foo]()
	assert.Equal(t, typ, IndirectType(typ))
	assert.Equal(t, typ, IndirectType(reflect.TypeFor[*Foo]()))
	assert.Equal(t, typ, IndirectType(reflect.TypeFor[**Foo]()))
}

func TestForEachField(t *testing.T) {
	type Inner struct {
		InnerField string `json:"inner"`
	}
	type Outer struct {
		Inner `json:",inline"`

		OuterField string `json:"outer"`
		_          string
	}

	fields := []string{}
	ForEachField(reflect.TypeFor[Outer](), func(f reflect.StructField) {
		fields = append(fields, f.Name)
	})

	assert.Contains(t, fields, "OuterField")
	assert.Contains(t, fields, "InnerField")
	assert.NotContains(t, fields, "Inner") // Inner is inlined
}

func TestInternalHelpers(t *testing.T) {
	t.Run("sanitizeDefName", func(t *testing.T) {
		assert.Equal(t, "Model", sanitizeDefName(nil, "Model", "github.com/oaswrap/spec"))
		assert.Equal(t, "Model", sanitizeDefName(reflect.TypeFor[struct{}](), "Model", "github.com/oaswrap/spec"))
		assert.Equal(t, "Model", sanitizeDefName(reflect.TypeFor[time.Time](), "Model", ""))
		assert.Equal(t, "TimeModel", sanitizeDefName(reflect.TypeFor[time.Time](), "Model", "github.com/oaswrap/spec"))
	})

	t.Run("reflector helper accessors", func(t *testing.T) {
		r := &Reflector{}
		assert.Empty(t, r.callerPkgPath())
		assert.Nil(t, r.interceptPropFn())
		assert.Nil(t, r.interceptSchemaFn())

		cfg := &openapi.Config{
			ReflectorConfig: &openapi.ReflectorConfig{
				DefNameCallerPkg: "github.com/oaswrap/spec",
				InterceptProp:    func(openapi.InterceptPropParams) error { return nil },
				InterceptSchema:  func(openapi.InterceptSchemaParams) (bool, error) { return false, nil },
			},
		}
		r = &Reflector{Config: cfg}
		assert.Equal(t, "github.com/oaswrap/spec", r.callerPkgPath())
		assert.NotNil(t, r.interceptPropFn())
		assert.NotNil(t, r.interceptSchemaFn())
	})

	t.Run("shallowCopyMap", func(t *testing.T) {
		assert.Nil(t, shallowCopyMap(nil))

		in := map[string]any{"a": 1}
		out := shallowCopyMap(in)
		assert.Equal(t, map[string]any{"a": 1}, out)

		in["a"] = 2
		assert.Equal(t, 1, out["a"])
	})

	t.Run("uniqueStrings", func(t *testing.T) {
		assert.Nil(t, uniqueStrings(nil))
		assert.Equal(t, []string{}, uniqueStrings([]string{}))
		assert.Equal(t, []string{"a", "b", "c"}, uniqueStrings([]string{"a", "b", "a", "c", "b"}))
	})
}
