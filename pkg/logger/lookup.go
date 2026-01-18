package logger

import (
	"reflect"
	"strconv"
	"strings"
)

/*
====================
 Public API
====================
*/

// ใช้เมื่อ expect ค่าเดียว
func LookupOne(path string, data any) (string, bool) {
	values := lookupStrings(path, data)
	if len(values) != 1 {
		return "", false
	}
	return values[0], true
}

// ใช้เมื่อ expect หลายค่า (array / wildcard)
func LookupMany(path string, data any) ([]string, bool) {
	values := lookupStrings(path, data)
	if len(values) == 0 {
		return nil, false
	}
	return values, true
}

/*
====================
 Internal
====================
*/

// main entry
func lookupStrings(path string, data any) []string {
	if path == "" || data == nil {
		return nil
	}

	parts := strings.Split(path, ".")
	raw := lookupValue(parts, reflect.ValueOf(data))

	return filterString(raw)
}

// recursive lookup
func lookupValue(parts []string, v reflect.Value) []any {
	v = unwrap(v)
	if !v.IsValid() {
		return nil
	}

	if len(parts) == 0 {
		return []any{v.Interface()}
	}

	part := parts[0]
	rest := parts[1:]

	switch v.Kind() {

	case reflect.Map:
		return lookupMap(part, rest, v)

	case reflect.Struct:
		return lookupStruct(part, rest, v)

	case reflect.Slice, reflect.Array:
		return lookupSlice(part, rest, v)
	}

	return nil
}

/*
====================
 Type handlers
====================
*/

func lookupMap(key string, rest []string, v reflect.Value) []any {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.String &&
			strings.EqualFold(k.String(), key) {
			return lookupValue(rest, v.MapIndex(k))
		}
	}
	return nil
}

func lookupStruct(field string, rest []string, v reflect.Value) []any {
	f := v.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, field)
	})
	if f.IsValid() {
		return lookupValue(rest, f)
	}
	return nil
}

func lookupSlice(part string, rest []string, v reflect.Value) []any {
	// wildcard
	if part == "*" {
		var out []any
		for i := 0; i < v.Len(); i++ {
			out = append(out, lookupValue(rest, v.Index(i))...)
		}
		return out
	}

	// [0,2]
	if idxs, ok := parseIndexes(part); ok {
		var out []any
		for _, i := range idxs {
			if i >= 0 && i < v.Len() {
				out = append(out, lookupValue(rest, v.Index(i))...)
			}
		}
		return out
	}

	// single index
	if i, err := strconv.Atoi(part); err == nil {
		if i >= 0 && i < v.Len() {
			return lookupValue(rest, v.Index(i))
		}
	}

	return nil
}

/*
====================
 Helpers
====================
*/

func unwrap(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func filterString(values []any) []string {
	var out []string
	for _, v := range values {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func parseIndexes(part string) ([]int, bool) {
	if !strings.HasPrefix(part, "[") || !strings.HasSuffix(part, "]") {
		return nil, false
	}

	part = strings.Trim(part, "[]")
	chunks := strings.Split(part, ",")

	var idxs []int
	for _, c := range chunks {
		i, err := strconv.Atoi(strings.TrimSpace(c))
		if err != nil {
			return nil, false
		}
		idxs = append(idxs, i)
	}

	return idxs, true
}
