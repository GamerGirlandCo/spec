package reflect

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/oaswrap/spec/openapi"
)

func (r *Reflector) TypeName(t reflect.Type) string {
	if name, ok := r.Names[t]; ok {
		return name
	}
	name := SanitizeTypeName(t.Name())
	name = sanitizeDefName(t, name, r.callerPkgPath())
	for _, prefix := range r.StripPrefixes() {
		name = strings.TrimPrefix(name, prefix)
	}
	if r.Config.ReflectorConfig != nil && r.Config.ReflectorConfig.InterceptDefName != nil {
		name = r.Config.ReflectorConfig.InterceptDefName(t, name)
	}
	if name == "" {
		name = "Schema"
	}
	base := name
	i := 2
	for usedType, usedName := range r.Names {
		if usedName == name && usedType != t {
			name = fmt.Sprintf("%s%d", base, i)
			i++
		}
	}
	r.Names[t] = name
	return name
}

func sanitizeDefName(t reflect.Type, defaultDefName, callerPkgPath string) string {
	if callerPkgPath == "" || defaultDefName == "" || t == nil || t.PkgPath() == "" || t.PkgPath() == callerPkgPath {
		return defaultDefName
	}
	pkgName := path.Base(t.PkgPath())
	if pkgName == "" || pkgName == "main" {
		return defaultDefName
	}
	pkgName = strings.ToUpper(pkgName[:1]) + pkgName[1:]
	return pkgName + defaultDefName
}

func (r *Reflector) callerPkgPath() string {
	if r.Config == nil || r.Config.ReflectorConfig == nil {
		return ""
	}
	return r.Config.ReflectorConfig.DefNameCallerPkg
}

func (r *Reflector) StripPrefixes() []string {
	if r.Config.ReflectorConfig == nil {
		return nil
	}
	return r.Config.ReflectorConfig.StripDefNamePrefix
}

func (r *Reflector) InlineRefs() bool {
	return r.Config.ReflectorConfig != nil && r.Config.ReflectorConfig.InlineRefs
}

func (r *Reflector) interceptPropFn() openapi.InterceptPropFunc {
	if r.Config == nil || r.Config.ReflectorConfig == nil {
		return nil
	}
	return r.Config.ReflectorConfig.InterceptProp
}

func (r *Reflector) interceptSchemaFn() openapi.InterceptSchemaFunc {
	if r.Config == nil || r.Config.ReflectorConfig == nil {
		return nil
	}
	return r.Config.ReflectorConfig.InterceptSchema
}

func IndirectType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func ForEachField(t reflect.Type, fn func(reflect.StructField)) {
	for i := range t.NumField() {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		if field.Anonymous && IndirectType(field.Type).Kind() == reflect.Struct && TagName(field, "json") == "" {
			ForEachField(IndirectType(field.Type), fn)
			continue
		}
		fn(field)
	}
}

func LowerCamel(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

var genericNameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func SanitizeTypeName(name string) string {
	if name == "" {
		return ""
	}

	// Handle slices: []Foo -> FooList
	if strings.HasPrefix(name, "[]") {
		return SanitizeTypeName(name[2:]) + "List"
	}

	// Handle pointers
	name = strings.TrimLeft(name, "*")

	// Handle generics: BaseResponse[github.com/foo.User]
	if start := strings.Index(name, "["); start != -1 && strings.HasSuffix(name, "]") {
		base := name[:start]
		inner := name[start+1 : len(name)-1]

		// Split multiple generic params: Map[string, int]
		parts := strings.Split(inner, ",")
		for i, p := range parts {
			parts[i] = SanitizeTypeName(strings.TrimSpace(p))
		}
		return SanitizeTypeName(base) + strings.Join(parts, "")
	}

	// For names with package paths: github.com/foo.User -> User
	// Note: reflect.Type.Name() usually only contains the local name for defined types,
	// but for generic instances it includes full paths for parameters.
	if lastDot := strings.LastIndex(name, "."); lastDot != -1 {
		name = name[lastDot+1:]
	}

	// Final cleanup for any remaining characters
	return genericNameRe.ReplaceAllString(name, "")
}
