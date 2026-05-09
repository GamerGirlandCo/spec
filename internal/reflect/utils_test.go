package reflect

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
