package encryption

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestSymmetricKey is a static symmetric key exposed for test code.
var testSymmetricKey = &SymmetricKey{
	k: []byte("\xa6\x75\x6f\x1b\xff\xda\x22\x3c\x61\x68\x40\xa1\xb0\x43\x5e\x67\xca\x24\x95\x35\x0a\x8f\x9d\x5c\x94\x40\x7a\xfb\x2a\x42\x9d\x82"),
}

func TestSymmetric(t *testing.T) {
	enc := testSymmetricKey.Encode()
	expected := "AES:pnVvG//aIjxhaEChsENeZ8oklTUKj51clEB6+ypCnYI="
	assert.Equal(t, expected, string(enc), "wrong encoded symmetric key")

	dec, err := enc.Decode()
	require.NoError(t, err, "failed to decode symmetric key")
	assert.Equal(t, testSymmetricKey, dec, "symmetric key decoded incorrectly")

	enc32 := testSymmetricKey.Encode32()
	expected32 := "AES32:UZ2W6G773IRDYYLIICQ3AQ26M7FCJFJVBKHZ2XEUIB5PWKSCTWBA===="
	assert.Equal(t, expected32, string(enc32), "wrong base32 encoded symmetric key")

	dec, err = enc32.Decode()
	require.NoError(t, err, "failed to decode base32 symmetric key")
	assert.Equal(t, testSymmetricKey, dec, "base32 symmetric key decoded incorrectly")
}

func TestSymmetricJSON(t *testing.T) {
	b, err := json.Marshal(testSymmetricKey)
	require.NoError(t, err, "failed to marshal symmetric key")

	expected := `"AES:pnVvG//aIjxhaEChsENeZ8oklTUKj51clEB6+ypCnYI="`
	assert.Equal(t, expected, string(b))

	var sym2 *SymmetricKey
	err = json.Unmarshal(b, &sym2)
	require.NoError(t, err, "failed to unmarshal symmetric key")

	assert.Equal(t, testSymmetricKey.Encode(), sym2.Encode())
}

func TestSymmetricYAML(t *testing.T) {
	b, err := yaml.Marshal(testSymmetricKey)
	require.NoError(t, err, "failed to marshal symmetric key")

	expected := "AES:pnVvG//aIjxhaEChsENeZ8oklTUKj51clEB6+ypCnYI=\n"
	assert.Equal(t, expected, string(b))

	var sym2 *SymmetricKey
	err = yaml.Unmarshal(b, &sym2)
	require.NoError(t, err, "failed to unmarshal symmetric key")

	assert.Equal(t, testSymmetricKey.Encode(), sym2.Encode())
}
