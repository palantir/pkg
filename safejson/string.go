// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

import (
	"unicode/utf8"
)

// String implements json.Marshaler and JSONAppender by quoting the underlying string.
type String string

func (s String) MarshalJSON() ([]byte, error) {
	dst := make([]byte, 0, QuotedLength(string(s)))
	return AppendQuoted(dst, string(s)), nil
}

func (s String) AppendJSON(dst []byte) ([]byte, error) {
	return AppendQuoted(dst, string(s)), nil
}

// Quote quotes and JSON-escapes s.
func Quote(s string) string {
	return string(appendQuoted(nil, nil, s))
}

// QuotedLength returns the length in bytes of the s when quoted and escaped.
// This is useful for pre-allocating memory before AppendQuoted.
func QuotedLength(s string) int {
	return lengthQuoted(nil, s)
}

// AppendQuoted quotes and JSON-escapes s and appends the result to dst.
// The resulting slice is returned in case it was resized by append().
func AppendQuoted(dst []byte, s string) []byte {
	return appendQuoted(dst, nil, s)
}

// AppendQuotedBytes quotes and JSON-escapes b and appends the result to dst.
// The resulting slice is returned in case it was resized by append().
func AppendQuotedBytes(dst []byte, b []byte) []byte {
	return appendQuoted(dst, b, "")
}

// QuotedBytesLength returns the length in bytes of the b when quoted and escaped.
// This is useful for pre-allocating memory before AppendQuotedBytes.
func QuotedBytesLength(b []byte) int {
	return lengthQuoted(b, "")
}

// appendQuoted is inspired by json.Marshal's private implementation: https://github.com/golang/go/blob/go1.19.1/src/encoding/json/encode.go#L1102-L1171
func appendQuoted(dst []byte, b []byte, s string) []byte {
	if b == nil && s != "" {
		b = []byte(s) // compiler detects that this does not escape; not an allocation.
	}
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(b); {
		if b[i] < utf8.RuneSelf {
			repl := jsonReplace[b[i]]
			if repl == nil {
				i++
				continue
			}
			if start < i {
				dst = append(dst, b[start:i]...)
			}
			dst = append(dst, repl...)
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(b[i:])
		switch {
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		case c == '\u2028':
			if start < i {
				dst = append(dst, b[start:i]...)
			}
			dst = append(dst, `\u2028`...)
			i += size
			start = i
		case c == '\u2029':
			if start < i {
				dst = append(dst, b[start:i]...)
			}
			dst = append(dst, `\u2029`...)
			i += size
			start = i
		case c == utf8.RuneError && size == 1:
			if start < i {
				dst = append(dst, b[start:i]...)
			}
			dst = append(dst, `\ufffd`...)
			i += size
			start = i
		default:
			i += size
		}
	}
	if start < len(b) {
		dst = append(dst, b[start:]...)
	}
	dst = append(dst, '"')
	return dst
}

func lengthQuoted(b []byte, s string) int {
	if b == nil && s != "" {
		b = []byte(s)
	}
	out := 2 // open/close quotes
	for i := 0; i < len(b); {
		if b[i] < utf8.RuneSelf {
			repl := jsonReplace[b[i]]
			if repl == nil {
				out++
			} else {
				out += len(repl)
			}
			i++
			continue
		}
		c, size := utf8.DecodeRune(b[i:])
		i += size
		switch {
		case c == utf8.RuneError && size == 1:
			out += len(`\ufffd`)
		case c == '\u2028', c == '\u2029':
			out += len(`\u2028`)
		default:
			out += size
		}
	}
	return out
}

var (
	// jsonReplace holds the values below 128 which require replacement in JSON strings.
	// If an entry is nil, the rune can be used as-is.
	// All values are nil except for the ASCII control characters (0-31), the
	// double quote ("), and the backslash character ("\").
	jsonReplace = [utf8.RuneSelf][]byte{
		'\\': []byte(`\\`),
		'"':  []byte(`\"`),
		'\n': []byte(`\n`),
		'\r': []byte(`\r`),
		'\t': []byte(`\t`),
	}
)

func init() {
	const hex = "0123456789abcdef"
	for i := 0; i < ' '; i++ {
		switch i {
		case '\n', '\r', '\t':
		default:
			// This encodes bytes < 0x20 except for \t, \n and \r.
			jsonReplace[i] = []byte{'\\', 'u', '0', '0', hex[i>>4], hex[i&0xF]}
		}
	}
}
