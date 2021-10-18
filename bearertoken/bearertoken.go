// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bearertoken

import (
	"fmt"
	"unicode/utf8"
)

// Token represents a bearer token, generally sent by a REST client in a
// Authorization or Cookie header for authentication purposes.
type Token string

func (t Token) String() string {
	return string(t)
}

func (t Token) MarshalText() ([]byte, error) {
	if err := validate(string(t)); err != nil {
		return nil, err
	}
	return []byte(t), nil
}

func (t *Token) UnmarshalText(text []byte) error {
	tok, err := New(string(text))
	if err != nil {
		return err
	}
	*t = tok
	return nil
}

func New(s string) (Token, error) {
	return Token(s), validate(s)
}

func validate(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("empty bearer token")
	}
	for i := 0; i < len(s); i++ {
		if !validChars[s[i]] || (i == 0 && s[i] == '=') {
			return fmt.Errorf("invalid character '%c' for bearer token", s[i])
		}
	}
	return nil
}

var validChars = [utf8.RuneSelf]bool{}

func init() {
	for i := '0'; i <= '9'; i++ {
		validChars[i] = true
	}
	for i := 'A'; i <= 'Z'; i++ {
		validChars[i] = true
	}
	for i := 'a'; i <= 'z'; i++ {
		validChars[i] = true
	}
	validChars['+'] = true
	validChars['-'] = true
	validChars['.'] = true
	validChars['/'] = true
	validChars['='] = true
	validChars['_'] = true
	validChars['~'] = true
}
