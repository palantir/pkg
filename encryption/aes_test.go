package encryption_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/encryption"
)

var testKey *encryption.SymmetricKey

func TestMain(m *testing.M) {
	encoded := `AES:dqidc6FDohqBF/Kmkfvl9hm8+gYD8t/Voe4aQRpzc08=`
	var err error
	testKey, err = encryption.EncodedSymmetricKey(encoded).Decode()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestAES(t *testing.T) {
	for _, testcase := range []struct {
		secret encryption.Plaintext
	}{
		{
			secret: "",
		},
		{
			secret: encryption.Plaintext(strings.Repeat("so secret\n", 64)),
		},
	} {
		sym, err := encryption.NewSymmetricKey()
		require.NoError(t, err, "failed to generate symmetric key")

		enc, err := encryption.EncryptSym(testcase.secret, sym)
		require.NoError(t, err, "symmetric encryption failed")

		dec, err := encryption.DecryptSym(enc, sym)
		require.NoError(t, err, "symmetric decryption failed")

		assert.Equal(t, testcase.secret, dec, "AES decrypted incorrectly")
	}
}

func TestDecryptInvalid(t *testing.T) {
	cipher := encryption.Ciphertext("not valid base64")
	_, err := encryption.DecryptSym(cipher, testKey)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Encrypted value is not valid base64")
	}

	cipher = encryption.Ciphertext(encryption.Base64Encode([]byte("too short")))
	_, err = encryption.DecryptSym(cipher, testKey)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Expected encrypted value to be at least 32 bytes but was 9")
	}
}
