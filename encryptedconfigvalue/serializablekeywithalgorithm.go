// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SerializableKeyWithAlgorithm is the representation of a KeyWithAlgorithm that is suitable for serialization. The
// encoded key is known to be syntactically valid.
type SerializableKeyWithAlgorithm struct {
	// algorithm is the string representation of the algorithm. It is the "{algorithm}" portion of
	// "{algorithm}:{base64-encoded-key}".
	algorithm AlgorithmType

	// encodedKey is the string representation of the encoded key. It is the "{base64-encoded-key}" portion of
	// "{algorithm}:{base64-encoded-key}".
	encodedKey string
}

func NewSerializableKeyWithAlgorithm(key string) (SerializableKeyWithAlgorithm, error) {
	key = strings.TrimSpace(key)
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return SerializableKeyWithAlgorithm{}, fmt.Errorf("key must be in the format <algorithm>:<base64-encoded-key>, was %q", key)
	}

	if _, ok := algorithms[AlgorithmType(parts[0])]; !ok {
		return SerializableKeyWithAlgorithm{}, fmt.Errorf("unknown key algorithm: %q", parts[0])
	}
	return SerializableKeyWithAlgorithm{
		algorithm:  AlgorithmType(parts[0]),
		encodedKey: parts[1],
	}, nil
}

func (skwa SerializableKeyWithAlgorithm) String() string {
	return skwa.SerializedStringForm()
}

func (skwa *SerializableKeyWithAlgorithm) SerializedStringForm() string {
	return fmt.Sprintf("%s:%s", string(skwa.algorithm), skwa.encodedKey)
}

func (skwa *SerializableKeyWithAlgorithm) MarshalJSON() ([]byte, error) {
	return json.Marshal(skwa.SerializedStringForm())
}

func (skwa *SerializableKeyWithAlgorithm) UnmarshalJSON(data []byte) error {
	var serializedStringForm string
	if err := json.Unmarshal(data, &serializedStringForm); err != nil {
		return err
	}
	decoded, err := NewSerializableKeyWithAlgorithm(serializedStringForm)
	if err != nil {
		return err
	}
	*skwa = decoded
	return nil
}

func (skwa *SerializableKeyWithAlgorithm) MarshalYAML() (interface{}, error) {
	return skwa.SerializedStringForm(), nil
}

func (skwa *SerializableKeyWithAlgorithm) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var serializedStringForm string
	if err := unmarshal(&serializedStringForm); err != nil {
		return err
	}
	decoded, err := NewSerializableKeyWithAlgorithm(serializedStringForm)
	if err != nil {
		return err
	}
	*skwa = decoded
	return nil
}
