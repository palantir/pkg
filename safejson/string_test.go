// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/palantir/pkg/safejson"
	"github.com/stretchr/testify/require"
)

func TestQuote(t *testing.T) {
	for _, test := range []struct {
		in  string
		out string
	}{
		{"plain", `"plain"`},
		{"new\nline", `"new\nline"`},
		{"\n❤️\t", `"\n❤️\t"`},
		{"I❤️NY", `"I❤️NY"`},
		{"I❤️", `"I❤️"`},
		{"\u2028", `"\u2028"`},
		{"\u2029", `"\u2029"`},
	} {
		t.Run(test.in, func(t *testing.T) {
			out := safejson.Quote(test.in)
			require.Equal(t, test.out, out)

			ref, err := json.Marshal(test.in)
			require.NoError(t, err)
			require.Equal(t, string(ref), out)
			require.Len(t, out, safejson.QuotedLength(test.in))
		})
	}
}

// TestEncodeString_RandomData is a fuzzing test that throws random data at the QuoteString
// function looking for panics.
func TestQuoteString_RandomData(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 200)
	for i := 0; i < 100000; i++ {
		n, err := rand.Read(b[:rand.Int()%len(b)])
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err = enc.Encode(string(b[:n]))
		require.NoError(t, err)
		sm := bytes.TrimRight(buf.Bytes(), "\n")

		out := safejson.Quote(string(b[:n]))
		require.Equal(t, string(sm), out)
		require.Len(t, out, safejson.QuotedLength(string(b[:n])))
	}
}

func BenchmarkQuote(b *testing.B) {
	const stringLen = 100

	runBench := func(b *testing.B, input string) {
		b.Logf("Quoting %q", input)
		b.Run("unbuffered preallocated AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				dst := make([]byte, 0, safejson.QuotedLength(input))
				_ = safejson.AppendQuoted(dst, input)
			}
		})
		b.Run("unbuffered AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = safejson.AppendQuoted(nil, input)
			}
		})
		b.Run("buffered AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			buf := make([]byte, 200)
			for i := 0; i < b.N; i++ {
				_ = safejson.AppendQuoted(buf[:0], input)
			}
		})
		b.Run("encoding/json", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b, err := json.Marshal(input)
				if err != nil {
					panic(err)
				}
				_ = b
			}
		})
		b.Run("encoding/json.Encoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(false)
				err := enc.Encode(input)
				if err != nil {
					panic(err)
				}
			}
		})
		b.Run("pregrow Encoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf := &bytes.Buffer{}
				buf.Grow(safejson.QuotedLength(input))
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(false)
				err := enc.Encode(input)
				if err != nil {
					panic(err)
				}
			}
		})
	}

	fixedRand := rand.New(rand.NewSource(0))
	randData := make([]byte, stringLen)
	_, err := fixedRand.Read(randData)
	require.NoError(b, err)
	hexString := hex.EncodeToString(randData[:hex.DecodedLen(stringLen)])
	require.Len(b, hexString, stringLen)

	b.Run("hexadecimal string", func(b *testing.B) {
		runBench(b, hexString)
	})
	b.Run("binary string", func(b *testing.B) {
		runBench(b, string(randData))
	})
}

func BenchmarkQuotePrealloc(b *testing.B) {
	const stringLen = 100

	runBench := func(b *testing.B, input string) {
		b.Run("unbuffered preallocated AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				dst := make([]byte, 0, safejson.QuotedLength(input))
				_ = safejson.AppendQuoted(dst, input)
			}
		})
		b.Run("unbuffered AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = safejson.AppendQuoted(nil, input)
			}
		})
		b.Run("buffered AppendQuotedString", func(b *testing.B) {
			b.ReportAllocs()
			buf := make([]byte, 200)
			for i := 0; i < b.N; i++ {
				_ = safejson.AppendQuoted(buf[:0], input)
			}
		})
		b.Run("encoding/json", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b, err := json.Marshal(input)
				if err != nil {
					panic(err)
				}
				_ = b
			}
		})
		b.Run("encoding/json.Encoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(false)
				err := enc.Encode(input)
				if err != nil {
					panic(err)
				}
			}
		})
		b.Run("pregrow Encoder", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf := &bytes.Buffer{}
				buf.Grow(safejson.QuotedLength(input))
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(false)
				err := enc.Encode(input)
				if err != nil {
					panic(err)
				}
			}
		})
	}

	fixedRand := rand.New(rand.NewSource(0))
	randData := make([]byte, stringLen)
	_, err := fixedRand.Read(randData)
	require.NoError(b, err)
	hexString := hex.EncodeToString(randData[:hex.DecodedLen(stringLen)])
	require.Len(b, hexString, stringLen)

	for _, l := range []int{100, 1000, 10000, 100000, 1000000, 10000000} {
		b.Run(fmt.Sprintf("%dB", l), func(b *testing.B) {
			randData := make([]byte, l)
			_, err := fixedRand.Read(randData)
			require.NoError(b, err)
			hexString := hex.EncodeToString(randData[:hex.DecodedLen(l)])
			require.Len(b, hexString, l)

			b.Run("hexadecimal string", func(b *testing.B) {
				runBench(b, hexString)
			})
			b.Run("binary string", func(b *testing.B) {
				runBench(b, string(randData))
			})
		})
	}
}
