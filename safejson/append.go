// Copyright (c) 2022 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package safejson

// JSONAppender is similar to json.Marshaler, but AppendJSON accepts a byte slice to which the encoded json is appended.
type JSONAppender interface {
	AppendJSON([]byte) ([]byte, error)
}

// JSONLengther is an optional interface types may implement which returns the encoded length of the object.
// If implemented, will be used to preallocate []byte space to avoid growing a slice from empty via serial appends.
type JSONLengther interface {
	LengthJSON() (int, error)
}

// JSONAppendFunc is a convenience type that implements json.Marshaler and JSONAppender from an AppendJSON closure.
type JSONAppendFunc func([]byte) ([]byte, error)

func (f JSONAppendFunc) AppendJSON(b []byte) ([]byte, error) {
	return f(b)
}

func (f JSONAppendFunc) MarshalJSON() ([]byte, error) {
	return JSONAppenderMarshaler{f}.MarshalJSON()
}

// JSONAppenderMarshaler wraps a JSONAppender to implement json.Marshaler.
type JSONAppenderMarshaler struct{ JSONAppender }

func (a JSONAppenderMarshaler) MarshalJSON() ([]byte, error) {
	var out []byte
	if lengther, ok := a.JSONAppender.(JSONLengther); ok {
		length, err := lengther.LengthJSON()
		if err != nil {
			return nil, err
		}
		out = make([]byte, 0, length)
	}
	return a.JSONAppender.AppendJSON(out)
}
