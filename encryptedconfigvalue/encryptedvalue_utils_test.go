// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

const aesTestKey = "AES:kHCJE62tlzYsP2eIxSnstSJVIMyk74+gCu/ImrhGMq8="

func TestDecryptAllEncryptedValueStringVars(t *testing.T) {
	key, err := encryptedconfigvalue.NewKeyWithType(aesTestKey)
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
			"decode normally with multiple values",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=} is the input\n${enc:invalid}\nagain, ${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}!"),
			[]byte("game theory is the input\n${enc:invalid}\nagain, game theory!"),
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
	} {
		assert.Equal(t, string(tc.want), string(encryptedconfigvalue.DecryptAllEncryptedValueStringVars(tc.input, key)), tc.test)
	}
}

func TestNormalizeEncryptedValueStringVars(t *testing.T) {
	key, err := encryptedconfigvalue.NewKeyWithType(aesTestKey)
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
			"decode normally with multiple values",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=} is the input\n${enc:invalid}\nagain, ${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}!"),
			map[string]encryptedconfigvalue.EncryptedValue{},
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=} is the input\n${enc:invalid}\nagain, ${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=}!"),
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
			"substitute encrypted value with legacy format",
			[]byte("${enc:JS1SwxxoRi6Mr7VVjx9kNw9bT87YGzW4JbuMQkZ+WR+Afdgdkz1Dnk6DCFWCHcF2826wM/h8hIq/cWI=}"),
			map[string]encryptedconfigvalue.EncryptedValue{
				"game theory": encryptedconfigvalue.MustNewEncryptedValue("enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k="),
			},
			[]byte("${enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k=}"),
		},
		{
			"substitute encrypted value with multiple values in legacy format",
			[]byte("${enc:tgU2MoIxsb3I5v6IW6+HBEiCbH8K695/LYQK8u9YxOB/nyz2IEPbi54xsSVY78xonidFgnhcl5AR8SM=} is the input\n${enc:invalid}\nagain, ${enc:JS1SwxxoRi6Mr7VVjx9kNw9bT87YGzW4JbuMQkZ+WR+Afdgdkz1Dnk6DCFWCHcF2826wM/h8hIq/cWI=}!"),
			map[string]encryptedconfigvalue.EncryptedValue{
				"game theory": encryptedconfigvalue.MustNewEncryptedValue("enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k="),
			},
			[]byte("${enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k=} is the input\n${enc:invalid}\nagain, ${enc:G2iHNdcNFPs5Omjml/hOqCWXQPpICa7+ZFk9naewVPxBp5LVvfnoqeRWXq0L+JerLc3zG8oRqS67s+k=}!"),
		},
	} {
		assert.Equal(t, string(tc.want), string(encryptedconfigvalue.NormalizeEncryptedValueStringVars(tc.input, key, tc.substitutions)), tc.test)
	}
}
