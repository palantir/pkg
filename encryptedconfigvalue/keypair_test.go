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

func TestKeyWithTypeAndEncryptedValSerDe(t *testing.T) {
	for i, currAlg := range []encryptedconfigvalue.AlgorithmType{
		encryptedconfigvalue.RSA,
		encryptedconfigvalue.AES,
	} {
		wantPlaintext := "foo"

		// create key pair and cipher
		kp, err := currAlg.GenerateKeyPair()
		require.NoError(t, err, "Case %d: %s", i, currAlg)
		encrypter := currAlg.Encrypter()

		// create encrypted value and serialize
		ev, err := encrypter.Encrypt(wantPlaintext, kp.EncryptionKey)
		require.NoError(t, err, "Case %d: %s", i, currAlg)
		evStr, err := ev.ToSerializable()
		require.NoError(t, err, "Case %d: %s", i, currAlg)

		// serialize decryption key
		encKeyStr := kp.DecryptionKey.ToSerializable()

		// deserialize encrypted value
		ev, err = encryptedconfigvalue.NewEncryptedValue(evStr)
		require.NoError(t, err, "Case %d: %s", i, currAlg)

		// deserialize key
		key, err := encryptedconfigvalue.NewKeyWithType(encKeyStr)
		require.NoError(t, err, "Case %d: %s", i, currAlg)

		// decrypt
		decryptedBytes, err := ev.Decrypt(key)
		require.NoError(t, err, "Case %d: %s", i, currAlg)

		assert.Equal(t, wantPlaintext, decryptedBytes, "Case %d: %s", i, currAlg)
	}
}

const (
	rsaPrivKeyBase64 = "MIIEvAIBADALBgkqhkiG9w0BAQEEggSoMIIEpAIBAAKCAQEAtKj+bpwUCq22ABjeJLBje+mD5XWmUAc8K2NbEGNGFaWGVAE1h/2Pjgxmj+LR4Bgt3OleYOnfV99ToqMNgB+HnNOJCg5LkHfq+WD6tRwhxFQMCmt73k9i8fgg7OCb1yTWo6pLCBIVWeisO0j0b1CYeIHebRemkx+8AK0ebsv4tdrIwAlb4jJTSz2rKZpEw7rLcGr8dFOYP5pg/jLJneittODD/uJj+1lpOze/AUT3bcuF6Ku0Oh4zNIvPcmm72bbr7+61lFOJB1IbDg1ahklE9m439/OOi3OOTdqq/HOu0k/dThrvovV1eedoL6UQz6RdijHNUt3iZqiues/Mq5dLSwIDAQABAoIBAQCEQyTi/cl+d+bC83HPEoQC99bkatmzxVg7u6WzvbpVprVNUwVJ5kzvBg0gUkKs+Ya6MPAzq4Uj5BBrBUyg/HRgUE4H2qdfwSt6H5HsfggKoC2gg0hQXXZnB+2y/k2ZmRK7B7We1v5isIFHdgXeaPb3YrzgyWveUmFlbVjWbOZM3AAJ0FczP2b3DErFS/iMyzdjCY9xwwXhQediMASj24c44/VLsaRCFesPXHoAXCvLLlPmNhfaw6ZVtHblg0QlFNftOUlIXC+s9yIN2ec38C10VR/yfGqVSYz+owXqNKRpMfsqNe1jWnl3+BVaqO53vsXzkYU8n8/vHdRSRZOiKpwhAoGBAMVutRUefOcApu5iEpHK+7Jte0o1kNFIwCXqiujjZcU/DKjDj2yK90ioza7Ntp7EHI9MUgCknyyiMlI/1VtHl3KiNfi0FQ646/AxOgzfrmUZTTyUgq02ToxFnAr1XYBzwwAPHKM4p2nJrf+G/7FpXhCMhK4qwGfMJ4C+i0pCoiJ/AoGBAOpAktM4SZGmBdtPyRpp0Z8tkrHoRNwn1YK+VS7XfkKmeMrsPEev7cjesaNJnMBjtlpGrAzVJC/ycEz5lZW7gBA5i/hDOLGegLjuu1SOTKXU4IFw5+vjTe4ecMFLLRE/rTeWMR3RfslzTiV66zLKZ9zuhq4YccGi3IFKKXVp+Fk1AoGAN86kVxToH2/6v7VvJFDpNrVlvUNI7S+QSOd0XoIwuUGqNWYZ+4eIgLxeb4PslBJBNGxRXacq6zXp3X/3sjaZY6jgcq2Mqj2xS5LOoubzZ9ZwE6izC30nVNU0V5Cl3nJac4DSCn0wLWH50hn52s867JibxJOHEZAOtoCl5NbS98cCgYAWoCoOUK96a+jA6BHqhTIEB+jVWjPcd9R9jli374R4d4/POcYQvoNfFXNe7CtBwd/JFG5lxuh54RbLuIekMLoL1yMX1ZZSQZb5RcW+QwhQNCGDHx6ngAr05ufJI7O0qMvYRJ9129g9KO/xWtAA1d/2TOuhQScrpslZi4o5lwSvyQKBgQCw/nLpPPlBGeA6jA0yZOuMPDZMGStLOAsGMmhV6LnBBllE475qQRPD/1xgcoWU7+u9H6sJNBR5p/WJq58IZFHzVCFVEBijLbNXDKOF9nDaczzXID5pM2Pspoz7JPpZkIFk0D2IR73M2RfoWNxYPRJCImDaL7HOXND6SNA+p6kkMg=="
	rsaPubKeyBase64  = "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdEtqK2Jwd1VDcTIyQUJqZUpMQmoKZSttRDVYV21VQWM4SzJOYkVHTkdGYVdHVkFFMWgvMlBqZ3htaitMUjRCZ3QzT2xlWU9uZlY5OVRvcU1OZ0IrSApuTk9KQ2c1TGtIZnErV0Q2dFJ3aHhGUU1DbXQ3M2s5aThmZ2c3T0NiMXlUV282cExDQklWV2Vpc08wajBiMUNZCmVJSGViUmVta3grOEFLMGVic3Y0dGRySXdBbGI0akpUU3oycktacEV3N3JMY0dyOGRGT1lQNXBnL2pMSm5laXQKdE9ERC91SmorMWxwT3plL0FVVDNiY3VGNkt1ME9oNHpOSXZQY21tNzJiYnI3KzYxbEZPSkIxSWJEZzFhaGtsRQo5bTQzOS9PT2kzT09UZHFxL0hPdTBrL2RUaHJ2b3ZWMWVlZG9MNlVRejZSZGlqSE5VdDNpWnFpdWVzL01xNWRMClN3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K"
)

func TestReadRSAKeys(t *testing.T) {
	for i, currCase := range []struct {
		name       string
		pubKeyStr  string
		privKeyStr string
	}{
		{
			"new format",
			"RSA-PUB:" + rsaPubKeyBase64,
			"RSA-PRIV:" + rsaPrivKeyBase64,
		},
		{
			"legacy format",
			"RSA:" + rsaPubKeyBase64,
			"RSA:" + rsaPrivKeyBase64,
		},
	} {
		pubKey, err := encryptedconfigvalue.NewKeyWithType(currCase.pubKeyStr)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)
		privKey, err := encryptedconfigvalue.NewKeyWithType(currCase.privKeyStr)
		require.NoError(t, err)

		wantPlaintext := "plaintext to encrypt"
		encrypter := encryptedconfigvalue.RSA.Encrypter()
		ev, err := encrypter.Encrypt(wantPlaintext, pubKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		decrypted, err := ev.Decrypt(privKey)
		require.NoError(t, err, "Case %d: %s", i, currCase.name)
		assert.Equal(t, wantPlaintext, decrypted, "Case %d: %s", i, currCase.name)
	}
}
