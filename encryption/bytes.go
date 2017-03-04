package encryption

import (
	"crypto/rand"
	"fmt"
)

func RandomBytes(n int) ([]byte, error) {
	out := make([]byte, n)
	if _, err := rand.Read(out); err != nil {
		return nil, fmt.Errorf("Failed to generate %d cryptographically strong pseudo-random bytes: %v", n, err)
	}
	return out, nil
}
