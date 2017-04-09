// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryption_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryption"
)

func TestRSAEncryptDecrypt(t *testing.T) {
	for i, currCase := range []struct {
		keyBits int
		input   []byte
	}{
		{2048, []byte("secret message")},
		{1024, []byte("secret message")},
		{1024, []byte{byte(1), byte(3)}},
	} {
		rsaKey, err := encryption.NewRSAOAEPWithSHA256(currCase.keyBits)
		require.NoError(t, err, "Case %d", i)

		encrypted, err := rsaKey.Encrypt(currCase.input)
		require.NoError(t, err, "Case %d", i)

		decrypted, err := rsaKey.Decrypt(encrypted)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.input, decrypted, "Case %d", i)
	}

	rsaKey, err := encryption.NewRSAOAEPWithSHA256(2048)
	require.NoError(t, err)

	privKeyBytes, err := encryption.RSAPrivateKeyPKCS8Bytes(rsaKey.PrivateKey())
	require.NoError(t, err)
	fmt.Println("private key:", base64.StdEncoding.EncodeToString(privKeyBytes))

	pubKeyBytes, err := encryption.RSAPublicKeyX509PEMBytes(rsaKey.PublicKey())
	require.NoError(t, err)
	fmt.Println("public key:", string(pubKeyBytes))

	encrypted, err := rsaKey.Encrypt([]byte("secret message"))
	require.NoError(t, err)

	fmt.Println(base64.StdEncoding.EncodeToString(encrypted))
}

func TestRSADecryptUsingStoredKey(t *testing.T) {
	for i, currCase := range []struct {
		privateKeyBase64 string
		ciphertextBase64 string
		plaintext        []byte
	}{
		{
			"MIIEugIBADALBgkqhkiG9w0BAQEEggSmMIIEogIBAAKCAQEAvb0eaqI8dQA7EdPsktiyb1AX03VohIyEes/Sfc9yx8hQxbyruJPfmVTvrhC06IQhPUDWB7zUZuq69Nv2T1hyOS5NXHuAnTJCDeVOGa4IefW9f9VATgjapoh40JLnYcrm3LqZixFuyCfwjAOUGz4aTOSYO7owao06tWixDoiWNLE6t0vZclRFj+KRNFLwpmUKw7TUaXYzRZT46fQy0h52nDG99DCIR+xSekeA/6Vb2obWR++pZhbu7f4ZwtdUCnnYLp0NbUwkNfuYBcvtDtDEz4F514rP05uEdWimyiBw7Gg0WSg5gQ13ANYWvQ/5iHjumawCQ9zWbwUHwpW+SoadMQIDAQABAoIBAHMIVomFxKuWsTlUx8gb0sqDv343X1+FJcijeNVH0Snoe3f2tBGarWRzx0A75sJVYSWWymw0gn3G8AQF26YtVErwlHxucAJd9wgfrqMJDSCL6RC4hF0LJyzx7nVdqyRx9Fd8VhynPAfjlwX8IW4Gz8Ewkk8bycC/0Qks6LOMAaz9fQj8jRiNfjbNrltl2A1kK060OyDfV5+V6crsJkXD5kT2diX8uz4gVApoouNkuZ2E4C4U6uYSg6Bumu5bpl3Monnu1d6NZMjB3jiV+RI1FPgW7qMfvK9K6G6AoX8TOganXeFbYSpKdt31+4jtUdk+SwcaWl+myLJRcGb4dJ4J+rUCgYEAwyfgnEANl2QQp5HVYfegq84AiXPLxMJ6HlN4KzQhTZX5kv/O0ebIZdRlwfiZdssYo7bugXOm4CmXuMwld47tlIZmGEcrAKZBbqn0HdBpnzpOVYwOJs7wmyjGTB3ErzYX1noyZ5WqWOi9SXbIJTZpxabY2XwhVKtEM0pX6rCoNE8CgYEA+OTjV7VtGFZOSzn4f/ptBmhsy7w/w7cE6bk4F+si6Nv5bnfOkm2rFbEOWw0a+lBuSIFQrqUEEYF84BqPKt2G1mY8fKstBISfJxGNy2PCLCdQnsufgVeNl0pE6u8hDB+Lh4Uzdpx+tJtCFd/6vYCSBAPBLiQIsR6/OJQYdbQ0Nn8CgYBcgJJefZ3znGKoit9xyEZIKSToAhMb+HKZ3Uagc901QVlC7C3EIHfsjHiPMJ7NSAct0o/KnF8E8bIQzfMUcJL8S5go+dLQQ/3Pzo7/csIdzy1CL1il3ID/ipwziAiqZCF4cANkRfSwn+DY6YyE1v3byfIPZF8IGwYAqcRyYbnY0QKBgBr031KOicRLBKvNGYby3oAFK1NdjiJqXhPaRaMBioRh3sACJdNiIVd2F7Hw120o7OjZaJ2hrbEfryCuf7cKyyHJbN+rwnJs0rfuhMb0hQE3ONoZ/6qIiwNJvfEb+R5RBFGnMY21IVv6PMwVuBhyJl5c8b1HldFpHRjJoWSOIeGNAoGAaCuRuD+3pluyF4skr8gR8M1ohWrP8Oy0+v1Kavev+VuIIHjgIQga6I75oBAuksNKXewvfRV6fGmkch/vuFq7axmrm8DLFXxXCPsALg21+/sO2kKLQ+L2+wdUELoe2/iPzw6b2uRiEJwtGKbklzay+waE+mvlBI4n1VNAyF/9ncE=",
			"ksK/wRw60AoAd37pmzEMkywMxvUjQOpnuIrtuESscbaYbVtFrnx9qlTBBTsvNxTi02/intlqu2BILFvxC2Alf8hcaa5dxiD0UJjqKavwodNziH+CJWhDEH9LAQJw2FYfKgPIlc2kjE/g7VNxWdX7v6zwwYnyVkbrSd9PAIIGsQ8OitlJ/biPOXR1VI0ZqNWcMUFhyVMhI016GTyciB8UhcNmJvN0YOyEeoC4Lmh15pPdNlzOhV1ble2Aq28Dwqgsy/hCiSNCFDdzS5Z9eziycAavsI4a/5lsfTwTOHRLQ28onQQIl0/Ft09shf1Aa8DV7y4d1BbdG7/geq5afnA3tQ==",
			[]byte("secret message"),
		},
	} {
		keyBytes, err := base64.StdEncoding.DecodeString(currCase.privateKeyBase64)
		require.NoError(t, err, "Case %d", i)
		key, err := encryption.RSAOAEPWithSHA256FromPrivateKey(keyBytes)
		require.NoError(t, err, "Case %d", i)

		encrypted, err := base64.StdEncoding.DecodeString(currCase.ciphertextBase64)
		require.NoError(t, err, "Case %d", i)

		gotPlaintext, err := key.Decrypt(encrypted)
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, currCase.plaintext, gotPlaintext, "Case %d", i)
	}
}

func TestRSAEncryptUsingStoredPublicKeyDecryptUsingStoredPrivateKey(t *testing.T) {
	for i, currCase := range []struct {
		privateKeyBase64 string
		publicKey        string
		plaintext        []byte
	}{
		{
			"MIIEuwIBADALBgkqhkiG9w0BAQEEggSnMIIEowIBAAKCAQEAo5bpR0t2sMw2rOgxNFAjYbQFzTu/Apu7bB5qENtW/z/Ey78GqvDUzGzImBqOGLrkzygLJbqNUQP4gvIRnOi6KKerNBgVWb/hV4qn+iKeZZKIHtxCH6Vevrs0poekFORp8BWTK0SAEjCjacFt5cEMNYGk2U2/hlgH1O9rcSxKRW7UuLR+93JOLURwI6vGH4MaeqM7UGwbya4TWS+p9RTGZYKJHrV9+IRTDZm/x3d9ACkSRT5TIADE1cKR8iAX3Jiggmha6965zwZScUNNOmrjDAFNVu2axmyur6R+Hg5U9CBd2pyMbYmwbr4OofFlg7FBVD42lK0Am97cr5BwwNC0nwIDAQABAoIBAFOgK7RUcWJDopeVQsH5TXz+qBCYQDa3IYJNse7YEYr+MD3vcxsjbcTqv0hyGr2tnJYBZGFvEhqeKwXVdQd/ONrbi3tf7Foq6qjzqpcF528JDyinc+31fY+G541RvaEoerdOcCMoK0ghMQg451MR8onPslObmRK2IZrKoWQDPhxrkHARVzMeCn/SZJC2Z79s6n7lHf6+B87oTfTtd4YGtYzWavOybj+h7SC38Q3xMAbDS7RxD9dJCtfkXH26vFplxAnO8jdr25MN96M+/OtOBkNnXWfpg8E10S/1YA0tTkEc2OLjH2vaSR2yQBRzUrEf1v+rFjZeWgMfGBmfgE1LAcECgYEAyEhIJ1kK3B478AxkUfqu/cxZOM0Swx+F+flwZ9hbFRJ/KA+czs/6Xb8/Wwzs0cZDKfUv6tiHL3mDz+s/69uz6x0IWHEuF2a7PsNDTIB4njmtjLDgWoXDOStxQjRT7xUnLgDHBXk4kxsd39hsKDZcc74fca12NeZE6Uii+IbF5PECgYEA0RlvmmKfw94u13QZQCs2LnimXxi4vI9xETgMEMSyBiWEVf5SHe05u9KxXm7gFXyob9871y7RNpnDk6DDRZAPfyk67wlx4EuUpHFx/rjapScSiQKM/HAWXD2XA7bjOOf7U5V1mHIQnUIONFRWZ5X3OfG9OIVPFp9XYqtxX/t68o8CgYBalibcdS+uQ5aEinZNhf7kGCs6v7Z+vqFQYPvwXDFGJKmSqw0XlYX+JOQ5AG2UrAHw1k8n/2uVk3aE8jhlK2gDLYx6xCY+u15xksu7rFfh6OCQQ+gVyW51Syrc8OINvxmLexqJZTyrfJZUioTQ41WJnDKIrhZLZq1AfnQHyJ11YQKBgG+Fy3piI6gJ9p2/NEB/S5SZkNKjktQvBTUT2YuP/Qs+M0jSLeX8QpCknSkqpaWQKR4RyA7Kz5b5h9BTLBML9Nfzm5UmSnBAn3TddNlQqnzvS/l7PMre1W45AzRd9O7C+87mpiO6opXdR0otuS/iUku7XRqqLzZ3odnkasGwlTCNAoGBAI0Rz/RuooXpzoul5P1LBd16WMq0HSnl0k80vgwFgluCZ14T4XDaKUaKvFZW58/nLm77Z05oESx0zyxAoFZe8xGAmoTJjE2kjyGJP1K517qgRbCgWDs1yjwjXHtMs3Gk8YDFm8JwdZHXq78CyoKfJpoW3bKnBWEE4TkBtwLhZfrR",
			`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAo5bpR0t2sMw2rOgxNFAj
YbQFzTu/Apu7bB5qENtW/z/Ey78GqvDUzGzImBqOGLrkzygLJbqNUQP4gvIRnOi6
KKerNBgVWb/hV4qn+iKeZZKIHtxCH6Vevrs0poekFORp8BWTK0SAEjCjacFt5cEM
NYGk2U2/hlgH1O9rcSxKRW7UuLR+93JOLURwI6vGH4MaeqM7UGwbya4TWS+p9RTG
ZYKJHrV9+IRTDZm/x3d9ACkSRT5TIADE1cKR8iAX3Jiggmha6965zwZScUNNOmrj
DAFNVu2axmyur6R+Hg5U9CBd2pyMbYmwbr4OofFlg7FBVD42lK0Am97cr5BwwNC0
nwIDAQAB
-----END RSA PUBLIC KEY-----`,
			[]byte("secret message"),
		},
	} {
		// create public key from stored material and encrypt plaintext
		pubKey, err := encryption.RSAOAEPWithSHA256FromPublicKey([]byte(currCase.publicKey))
		require.NoError(t, err, "Case %d", i)
		encrypted, err := pubKey.Encrypt(currCase.plaintext)
		require.NoError(t, err, "Case %d", i)

		// create private key from stored material and decrypt ciphertext
		privKeyBytes, err := base64.StdEncoding.DecodeString(currCase.privateKeyBase64)
		require.NoError(t, err, "Case %d", i)
		privKey, err := encryption.RSAOAEPWithSHA256FromPrivateKey(privKeyBytes)
		require.NoError(t, err, "Case %d", i)
		gotPlaintext, err := privKey.Decrypt(encrypted)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.plaintext, gotPlaintext, "Case %d", i)
	}
}
