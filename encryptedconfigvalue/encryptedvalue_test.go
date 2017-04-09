// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

const aesTestKey = "AES:kHCJE62tlzYsP2eIxSnstSJVIMyk74+gCu/ImrhGMq8="

func TestEncryptedValue_UnmarshalJSON(t *testing.T) {
	for i, currCase := range []struct {
		input []byte
		want  encryptedconfigvalue.EncryptedValue
	}{
		{
			input: []byte(`"${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}"`),
			want:  encryptedconfigvalue.MustNewEncryptedValue("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}"),
		},
	} {
		var got encryptedconfigvalue.EncryptedValue
		err := json.Unmarshal(currCase.input, &got)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.want, got)
	}
}

func TestEncryptedValue_UnmarshalYAML(t *testing.T) {
	type yamlConf struct {
		EncryptedValue encryptedconfigvalue.EncryptedValue `yaml:"encrypted-value"`
	}

	for i, currCase := range []struct {
		input []byte
		want  func() yamlConf
	}{
		{
			input: []byte(`---
encrypted-value: ${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}
`),
			want: func() yamlConf {
				ev, err := encryptedconfigvalue.NewEncryptedValue("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}")
				require.NoError(t, err)
				return yamlConf{
					EncryptedValue: ev,
				}
			},
		},
	} {
		var got yamlConf
		err := yaml.Unmarshal(currCase.input, &got)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.want(), got)
	}
}

func TestDecryptAll(t *testing.T) {
	skwa, err := encryptedconfigvalue.NewSerializableKeyWithAlgorithm(aesTestKey)
	require.NoError(t, err)
	key, err := encryptedconfigvalue.AES.Definition().ToPublicKey(skwa)
	require.NoError(t, err)

	for _, tc := range []struct {
		test  string
		input []byte
		want  []byte
	}{
		{
			"decode normally",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}"),
			[]byte("game theory"),
		},
		{
			"decode failure",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5A8S=}"),
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5A8S=}"),
		},
		{
			"no match",
			[]byte("ok"),
			[]byte("ok"),
		},
		{
			"decode substring",
			[]byte("a_lovely_${enc:USone9TsxkC4VtfhNwQsslJwVLQZZs5qAOwaUPjxHu7o0SBlvbBkp+e+RPmHfVDdfU78qFo=}"),
			[]byte("a_lovely_party"),
		},
		{
			"decode two",
			[]byte("${enc:USone9TsxkC4VtfhNwQsslJwVLQZZs5qAOwaUPjxHu7o0SBlvbBkp+e+RPmHfVDdfU78qFo=}" +
				"${enc:MyC9nEfercncpbfD2wKd8Aw6Y6As9L67peQJvd/TOqCtVfc2YjnkNWiM6GnlHglbwUb++5nAQKE=}"),
			[]byte("party+animals"),
		},
	} {
		got := encryptedconfigvalue.DecryptAllEncryptedValues(tc.input, key)
		require.Equal(t, tc.want, got, tc.test)
	}
}

func TestNormalizeEncryptedValues(t *testing.T) {
	skwa, err := encryptedconfigvalue.NewSerializableKeyWithAlgorithm(aesTestKey)
	require.NoError(t, err)
	key, err := encryptedconfigvalue.AES.Definition().ToPublicKey(skwa)
	require.NoError(t, err)

	for _, tc := range []struct {
		test          string
		input         []byte
		substitutions map[string]encryptedconfigvalue.EncryptedValue
		want          []byte
	}{
		{
			"decode normally",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}"),
			map[string]encryptedconfigvalue.EncryptedValue{},
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}"),
		},
		{
			"decode failure",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5A8S=}"),
			map[string]encryptedconfigvalue.EncryptedValue{},
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5A8S=}"),
		},
		{
			"no match",
			[]byte("ok"),
			map[string]encryptedconfigvalue.EncryptedValue{},
			[]byte("ok"),
		},
		{
			"substitute encrypted value",
			[]byte("${enc:JS1SwxxoRi6Mr7VVjx9kNw9bT87YGzW4JbuMQkZ+WR+Afdgdkz1Dnk6DCFWCHcF2826wM/h8hIq/cWI=}"),
			map[string]encryptedconfigvalue.EncryptedValue{
				"game theory": encryptedconfigvalue.MustNewEncryptedValue("${enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k=}"),
			},
			[]byte("${enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k=}"),
		},
	} {
		assert.Equal(t, tc.want, encryptedconfigvalue.NormalizeEncryptedValues(tc.input, key, tc.substitutions), tc.test)
	}
}
