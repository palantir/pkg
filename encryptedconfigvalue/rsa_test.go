// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryptedconfigvalue"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		input []byte
	}{
		{[]byte("secret message")},
		{[]byte{byte(1), byte(3)}},
	} {
		kp, err := encryptedconfigvalue.RSA.Definition().NewKeyPair()
		require.NoError(t, err, "Case %d", i)

		encryptedBytes, err := kp.PublicKey().Encrypt(currCase.input)
		require.NoError(t, err, "Case %d", i)

		decryptedBytes, err := kp.PrivateKey().Decrypt(encryptedBytes)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.input, decryptedBytes, "Case %d", i)
	}
}

func TestRSALoadStoredPublicKeyAndWrite(t *testing.T) {
	for i, currCase := range []struct {
		serializedPublicKey string
	}{
		{"RSA:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBMTlQaEk2RER5L3RFZlRURlpRL0IKOTVISEw2YjJWK2J2ZHBBbGNiU2ZQMnRTVVpxNFhZUk1JOVpIVklmWnpQM2ZOMmdFZ0ZiMm9SeWVOaVNvK0grYQpKTzdGMTM5eDFkZENsczdVQVJ5b05PTllDdUFCVFZlZTFhN1J5VUFXS3pkckF3VmpXcy9BblcwQTFUWXhSTFlTCjBjR0x0MjZ1YVg0NG1kNFdNbGYvT1FPSGRqSTNVQ05YUEJjdmRGOVZQdlVmbDgvNWFQKzBkd0IvMVZwa0g0clAKb0Zmd2VaMytYd0JZWDJtcTFocW5aRHYzWVlNTU9Wd0ltVDA5WmtQNkUvYmVROW5xSXBTY242TktQZEVaUE9pOQp2MXNzVVVXWHh3UnZZVXlBcTlWYUNaOHZFbW9QaWZCS3VGeW03bUFoN1RLSi95anJxckRoWXc3aVFGdHRsemxJClV3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K"},
	} {
		kwa, err := encryptedconfigvalue.NewKeyWithAlgorithm(currCase.serializedPublicKey)
		require.NoError(t, err, "Case %d", i)

		serialized, err := kwa.ToSerializable()
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.serializedPublicKey, serialized.SerializedStringForm(), "Case %d", i)
	}
}

func TestRSALoadStoredPrivateKeyAndWrite(t *testing.T) {
	for i, currCase := range []struct {
		serializedPrivateKey string
	}{
		{"RSA:MIIEugIBADALBgkqhkiG9w0BAQEEggSmMIIEogIBAAKCAQEAvb0eaqI8dQA7EdPsktiyb1AX03VohIyEes/Sfc9yx8hQxbyruJPfmVTvrhC06IQhPUDWB7zUZuq69Nv2T1hyOS5NXHuAnTJCDeVOGa4IefW9f9VATgjapoh40JLnYcrm3LqZixFuyCfwjAOUGz4aTOSYO7owao06tWixDoiWNLE6t0vZclRFj+KRNFLwpmUKw7TUaXYzRZT46fQy0h52nDG99DCIR+xSekeA/6Vb2obWR++pZhbu7f4ZwtdUCnnYLp0NbUwkNfuYBcvtDtDEz4F514rP05uEdWimyiBw7Gg0WSg5gQ13ANYWvQ/5iHjumawCQ9zWbwUHwpW+SoadMQIDAQABAoIBAHMIVomFxKuWsTlUx8gb0sqDv343X1+FJcijeNVH0Snoe3f2tBGarWRzx0A75sJVYSWWymw0gn3G8AQF26YtVErwlHxucAJd9wgfrqMJDSCL6RC4hF0LJyzx7nVdqyRx9Fd8VhynPAfjlwX8IW4Gz8Ewkk8bycC/0Qks6LOMAaz9fQj8jRiNfjbNrltl2A1kK060OyDfV5+V6crsJkXD5kT2diX8uz4gVApoouNkuZ2E4C4U6uYSg6Bumu5bpl3Monnu1d6NZMjB3jiV+RI1FPgW7qMfvK9K6G6AoX8TOganXeFbYSpKdt31+4jtUdk+SwcaWl+myLJRcGb4dJ4J+rUCgYEAwyfgnEANl2QQp5HVYfegq84AiXPLxMJ6HlN4KzQhTZX5kv/O0ebIZdRlwfiZdssYo7bugXOm4CmXuMwld47tlIZmGEcrAKZBbqn0HdBpnzpOVYwOJs7wmyjGTB3ErzYX1noyZ5WqWOi9SXbIJTZpxabY2XwhVKtEM0pX6rCoNE8CgYEA+OTjV7VtGFZOSzn4f/ptBmhsy7w/w7cE6bk4F+si6Nv5bnfOkm2rFbEOWw0a+lBuSIFQrqUEEYF84BqPKt2G1mY8fKstBISfJxGNy2PCLCdQnsufgVeNl0pE6u8hDB+Lh4Uzdpx+tJtCFd/6vYCSBAPBLiQIsR6/OJQYdbQ0Nn8CgYBcgJJefZ3znGKoit9xyEZIKSToAhMb+HKZ3Uagc901QVlC7C3EIHfsjHiPMJ7NSAct0o/KnF8E8bIQzfMUcJL8S5go+dLQQ/3Pzo7/csIdzy1CL1il3ID/ipwziAiqZCF4cANkRfSwn+DY6YyE1v3byfIPZF8IGwYAqcRyYbnY0QKBgBr031KOicRLBKvNGYby3oAFK1NdjiJqXhPaRaMBioRh3sACJdNiIVd2F7Hw120o7OjZaJ2hrbEfryCuf7cKyyHJbN+rwnJs0rfuhMb0hQE3ONoZ/6qIiwNJvfEb+R5RBFGnMY21IVv6PMwVuBhyJl5c8b1HldFpHRjJoWSOIeGNAoGAaCuRuD+3pluyF4skr8gR8M1ohWrP8Oy0+v1Kavev+VuIIHjgIQga6I75oBAuksNKXewvfRV6fGmkch/vuFq7axmrm8DLFXxXCPsALg21+/sO2kKLQ+L2+wdUELoe2/iPzw6b2uRiEJwtGKbklzay+waE+mvlBI4n1VNAyF/9ncE="},
	} {
		kwa, err := encryptedconfigvalue.NewKeyWithAlgorithm(currCase.serializedPrivateKey)
		require.NoError(t, err, "Case %d", i)

		serialized, err := kwa.ToSerializable()
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.serializedPrivateKey, serialized.SerializedStringForm(), "Case %d", i)
	}
}

func TestRSADecryptUsingStoredKey(t *testing.T) {
	for i, currCase := range []struct {
		serializedKey    string
		ciphertextBase64 string
		plaintext        []byte
	}{
		{
			"AES:s0u/5zMvOz9bd2/7QSJ0yaRpav9kgAmLh6GyXkttwC4=",
			"KDXbSDZnEOYfSF6sL28Eh60HP0Lo7GZBIKPufqveF/aso9hEw/2F80ox5s8kkK+7e/jQE1vUZ3f+n33x4OM=",
			[]byte("secret message"),
		},
	} {
		kwa, err := encryptedconfigvalue.NewKeyWithAlgorithm(currCase.serializedKey)
		require.NoError(t, err, "Case %d", i)

		wantCiphertext, err := base64.StdEncoding.DecodeString(currCase.ciphertextBase64)
		require.NoError(t, err, "Case %d", i)

		gotPlaintext, err := kwa.Decrypt(wantCiphertext)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.plaintext, gotPlaintext)
	}
}
