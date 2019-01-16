// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safeyaml

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// JSONtoYAML converts json data to yaml while preserving order of object fields.
// Invoke yaml.Marshal on the returned interface{}.
func JSONtoYAML(jsonBytes []byte) (interface{}, error) {
	dec := json.NewDecoder(bytes.NewReader(jsonBytes))
	dec.UseNumber()
	return tokenizerToYaml(dec)
}

var errClosingArrayDelim = fmt.Errorf("unexpected ']' delimiter")
var errClosingObjectDelim = fmt.Errorf("unexpected '}' delimiter")

func tokenizerToYaml(dec *json.Decoder) (interface{}, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return nil, nil
	}
	switch v := tok.(type) {
	case string:
		return v, nil
	case bool:
		return v, nil
	case float64:
		return v, nil
	case json.Number:
		if numI, err := v.Int64(); err == nil {
			return numI, nil
		}
		if numF, err := v.Float64(); err == nil {
			return numF, nil
		}
		return v.String(), nil
	case json.Delim:
		switch v {
		case '[':
			arr := make([]interface{}, 0)
			for {
				elem, err := tokenizerToYaml(dec)
				if err == errClosingArrayDelim {
					break
				}
				if err != nil {
					return nil, err
				}
				arr = append(arr, elem)
			}
			return arr, nil
		case '{':
			obj := make(yaml.MapSlice, 0)
			for {
				objectKeyI, err := tokenizerToYaml(dec)
				if err == errClosingObjectDelim {
					break
				}
				if err != nil {
					return nil, err
				}
				objectValueI, err := tokenizerToYaml(dec)
				if err != nil {
					return nil, err
				}
				obj = append(obj, yaml.MapItem{Key: objectKeyI, Value: objectValueI})
			}
			return obj, nil
		case ']':
			return nil, errClosingArrayDelim
		case '}':
			return nil, errClosingObjectDelim
		default:
			return nil, fmt.Errorf("unrecognized delimiter")
		}
	default:
		return nil, fmt.Errorf("unrecognized token type %T", tok)
	}
}
