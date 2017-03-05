package safejson

import (
	"fmt"
	"reflect"
	"strings"
)

// FromYAML converts nested map[interface{}]interface{} to map[string]interface{}.
func FromYAML(y interface{}) (interface{}, error) {
	return fromYAMLValue(reflect.ValueOf(y), "")
}

func fromYAMLValue(v reflect.Value, path string) (interface{}, error) {
	switch v.Kind() {
	case reflect.Map:
		return fromYAMLMap(v, path)
	case reflect.Slice, reflect.Array:
		return fromYAMLArray(v, path)
	case reflect.Interface, reflect.Ptr:
		return fromYAMLValue(v.Elem(), path)
	case reflect.Invalid:
		return nil, nil
	default:
		return v.Interface(), nil
	}
}

func fromYAMLMap(v reflect.Value, path string) (interface{}, error) {
	m := make(map[string]interface{}, v.Len())
	for _, entry := range v.MapKeys() {
		k, err := fromYAMLKey(entry, path)
		if err != nil {
			return nil, err
		}
		v, err := fromYAMLValue(v.MapIndex(entry), fmt.Sprintf("%s.%s", path, k))
		if err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

func fromYAMLArray(v reflect.Value, path string) (interface{}, error) {
	a := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		v, err := fromYAMLValue(v.Index(i), fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return nil, err
		}
		a[i] = v
	}
	return a, nil
}

func fromYAMLKey(k reflect.Value, path string) (string, error) {
	switch k.Kind() {
	case reflect.String:
		return k.String(), nil
	case reflect.Interface, reflect.Ptr:
		return fromYAMLKey(k.Elem(), path)
	default:
		return "", expectedString(k, path)
	}
}

func expectedString(k reflect.Value, path string) error {
	var valStr string
	if k.IsValid() {
		valStr = fmt.Sprintf("%v: %v", k.Type(), k.Interface())
	} else {
		valStr = "null"
	}
	if path == "" {
		return fmt.Errorf("Expected map key to be a string but was %s", valStr)
	}
	path = strings.TrimPrefix(path, ".") // no leading dot
	return fmt.Errorf("Expected map key inside %s to be a string but was %s", path, valStr)
}
