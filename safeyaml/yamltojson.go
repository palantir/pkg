// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safeyaml

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

// YAMLUnmarshalerToJSONBytes decodes YAML bytes (in the form of an unmarshal lambda provided by go-yaml)
// into an interface{} where all objects and maps are decoded as a map[string]interface{} for JSON
// compatibility. That object is then marshaled to JSON bytes. This can be used in an UnmarshalYAML
// implementation where the type implements custom UnmarshalJSON behavior and wants it to apply to
// YAML as well. For example:
//
//	func (o *Foo) UnmarshalYAML(unmarshal func(interface{}) error) error {
//		jsonBytes, err := safeyaml.YAMLUnmarshalerToJSONBytes(unmarshal)
//		if err != nil {
//			return err
//		}
//		return json.Unmarshal(jsonBytes, o)
//	}
func YAMLUnmarshalerToJSONBytes(unmarshal func(interface{}) error) ([]byte, error) {
	// Convert the YAML to an object.
	var yamlObj interface{}
	if err := unmarshal(&yamlObj); err != nil {
		return nil, err
	}

	// YAML objects are not completely compatible with JSON objects (e.g. YAML
	// may have non-string keys). So, convert the YAML-compatible object
	// to a JSON-compatible object, failing with an error if irrecoverable
	// incompatibilities happen along the way.
	jsonObj, err := convertToJSONableObject(yamlObj, nil)
	if err != nil {
		return nil, err
	}

	// Convert this object to JSON and return the data.
	return json.Marshal(jsonObj)
}

// YAMLtoJSONBytes converts YAML content to JSON.
// Returns an error if it encounters types that are invalid in JSON but
// valid in YAML (e.g. non-string map keys).
func YAMLtoJSONBytes(yamlBytes []byte) ([]byte, error) {
	return YAMLUnmarshalerToJSONBytes(func(i interface{}) error {
		return yaml.Unmarshal(yamlBytes, *&i)
	})
}
