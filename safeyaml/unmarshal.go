// Copyright (c) 2015-2016 Michael Persson
// Copyright (c) 2012–2015 Elasticsearch <http://www.elastic.co>
//
// The bulk of this code is copied from from https://github.com/go-yaml/yaml/issues/139#issuecomment-220072190
//
// Originally distributed as part of "beats" repository (https://github.com/elastic/beats).
// Modified specifically for "iodatafmt" package.
//
// Distributed underneath "Apache License, Version 2.0" which is compatible with the LICENSE for this package.

package safeyaml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Unmarshal YAML to map[string]interface{} instead of map[interface{}]interface{}.
func Unmarshal(in []byte, out interface{}) error {
	var res interface{}

	if err := yaml.Unmarshal(in, &res); err != nil {
		return err
	}
	cleaned, err := cleanupMapValue(res)
	if err != nil {
		return err
	}
	*out.(*interface{}) = cleaned

	return nil
}

func cleanupInterfaceArray(in []interface{}) ([]interface{}, error) {
	res := make([]interface{}, len(in))
	for i, v := range in {
		cleaned, err := cleanupMapValue(v)
		if err != nil {
			return nil, err
		}
		res[i] = cleaned
	}
	return res, nil
}

func cleanupInterfaceMap(in map[interface{}]interface{}) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for k, v := range in {
		cleaned, err := cleanupMapValue(v)
		if err != nil {
			return nil, err
		}
		newKey := fmt.Sprintf("%v", k)
		if _, exists := res[newKey]; exists {
			return nil, errors.New(fmt.Sprintf("conflicting key %q encountered while unmarshaling yaml", newKey))
		}
		res[newKey] = cleaned
	}
	return res, nil
}

func cleanupMapValue(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	default:
		return v, nil
	}
}
