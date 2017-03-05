package safejson

import (
	"bytes"
	"math/big"
	"strconv"
	"testing"
)

// encodeTests are shared by TestEncoder and TestMarshal to guarantee that their
// behaviors are consistent.
var encodeTests = map[string]struct {
	in   interface{}
	want string
}{
	"unescaped HTML characters": {
		in:   `< this string contains & HTML characters that should not be escaped>`,
		want: `"< this string contains & HTML characters that should not be escaped>"`,
	},
	"big.Float as json.Number": {
		in:   big.NewFloat(3.14),
		want: `"3.14"`,
	},
}

func TestEncoder(t *testing.T) {
	for name, tt := range encodeTests {
		// Test Encoder
		var got bytes.Buffer
		if err := NewEncoder(&got).Encode(tt.in); err != nil {
			t.Errorf("failed to encode %s: %v", name, tt.in)
			continue
		}
		if got.String() != tt.want+"\n" { // Encode writes a newline after the JSON
			t.Errorf("wrong encoding for %s: got %q but wanted %q", name, got.String(), tt.want)
		}
	}
}

func TestMarshal(t *testing.T) {
	for name, tt := range encodeTests {
		// Test Marshal
		got, err := Marshal(tt.in)
		if err != nil {
			t.Errorf("failed to marshal %s: %v", name, tt.in)
			continue
		}
		if string(got) != tt.want {
			t.Errorf("wrong encoding for %s: got %q but wanted %q", name, string(got), tt.want)
		}
	}
}

func TestConcurrentMarshal(t *testing.T) {
	var (
		old = 532173
		new = 23589217
	)
	got, err := Marshal(old)
	if err != nil {
		t.Fatalf("failed to marshal %d", old)
	}
	if _, err := Marshal(new); err != nil {
		t.Fatalf("failed to marshal %d", new)
	}
	if string(got) != strconv.Itoa(old) {
		t.Errorf("buffer reuse: got %s but wanted %s", string(got), strconv.Itoa(old))
	}
}
