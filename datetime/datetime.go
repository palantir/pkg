// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datetime

import (
	"strings"
	"time"
)

// DateTime is an alias for time.Time which implements serialization matching the
// Conjure wire specification at https://github.com/palantir/conjure/blob/master/docs/spec/wire.md
//
// Note that the wire specification states only "In accordance with ISO 8601."
// The Go implementation will marshal all output for this type using the time.RFC3339Nano format.
// The Go implementation will unmarshal all representations that adhere to the time.RFC3339Nano format and also supports
// unmarshalling any of the following ISO-8601 formats:
//   - uuuu-MM-dd'T'HH:mmXXXXX
//   - uuuu-MM-dd'T'HH:mm:ssXXXXX
//   - uuuu-MM-dd'T'HH:mm:ss.SSSXXXXX
//   - uuuu-MM-dd'T'HH:mm:ss.SSSSSSXXXXX
//   - uuuu-MM-dd'T'HH:mm:ss.SSSSSSSSSXXXXX
//
// The above formats are specifically supported because they are the formats produced by Java's OffsetDateTime type's
// ToString method (https://docs.oracle.com/javase/8/docs/api/java/time/OffsetDateTime.html#toString--).
type DateTime time.Time

func (d DateTime) String() string {
	return time.Time(d).Format(time.RFC3339Nano)
}

// MarshalText implements encoding.TextMarshaler (used by encoding/json and others).
func (d DateTime) MarshalText() ([]byte, error) {
	return time.Time(d).AppendFormat(nil, time.RFC3339Nano), nil
}

// UnmarshalText implements encoding.TextUnmarshaler (used by encoding/json and others).
func (d *DateTime) UnmarshalText(b []byte) error {
	t, err := ParseDateTime(string(b))
	if err != nil {
		return err
	}
	*d = t
	return nil
}

// This layout string is for the case where seconds are omitted. The example time in the format string is valid
// according to ISO 8601, but not according to RFC 3339. See https://ijmacd.github.io/rfc3339-iso8601/.
// It is worth supporting this case specifically because Java's OffsetDateTime's ToString method removes seconds and
// nanoseconds when they are 0 (see https://stackoverflow.com/questions/49445766/offsetdatetime-tostring-removes-seconds-and-nanoseconds-when-they-are-0).
const rfc3339SecondsOmitted = "2006-01-02T15:04Z"

// ParseDateTime parses a DateTime from a string. Supports all strings that match the time.RFC3339Nano format as well as
// ISO 8601 formats produced by Java's OffsetDateTime type. If the provided string contains an open bracket ('['), only
// parses the input before the first occurrence of the open bracket (this is to support inputs that end with an optional
// zone identifier enclosed in square brackets: for example, "2017-01-02T04:04:05.000000000+01:00[Europe/Berlin]").
func ParseDateTime(s string) (DateTime, error) {
	// If the input string ends in a ']' and contains a '[', parse the string up to '['.
	if strings.HasSuffix(s, "]") {
		if openBracketIdx := strings.LastIndex(s, "["); openBracketIdx != -1 {
			s = s[:openBracketIdx]
		}
	}
	timeVal, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		if rfc3339SecondsOmittedVal, rfc3339SecondsOmittedErr := time.Parse(rfc3339SecondsOmitted, s); rfc3339SecondsOmittedErr == nil {
			// if parsing as RFC3339Nano failed, attempt to parse using rfc3339SecondsOmitted: if this succeeds, then
			// return the result
			timeVal = rfc3339SecondsOmittedVal
		} else {
			// otherwise, return original error
			return DateTime(time.Time{}), err
		}
	}
	return DateTime(timeVal), nil
}
