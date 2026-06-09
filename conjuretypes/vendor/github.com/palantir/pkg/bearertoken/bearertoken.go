// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bearertoken

// Token represents a bearer token, generally sent by a REST client in a
// Authorization or Cookie header for authentication purposes.
type Token string

func (t Token) String() string {
	return string(t)
}

func (t Token) MarshalText() ([]byte, error) {
	return []byte(t), nil
}

func (t *Token) UnmarshalText(text []byte) error {
	*t = Token(text)
	return nil
}
