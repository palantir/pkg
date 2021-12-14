// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.15
// +build go1.15

package datetime_test

import (
	"encoding/json"
	"testing"

	"github.com/palantir/pkg/datetime"
	"github.com/stretchr/testify/assert"
)

func TestDateTimeUnmarshalInvalid(t *testing.T) {
	for i, currCase := range []struct {
		input   string
		wantErr string
	}{
		{
			input:   `"foo"`,
			wantErr: "parsing time \"foo\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"foo\" as \"2006\"",
		},
		{
			input:   `"2017-01-02T04:04:05.000000000+01:00[Europe/Berlin"`,
			wantErr: "parsing time \"2017-01-02T04:04:05.000000000+01:00[Europe/Berlin\": extra text: \"[Europe/Berlin\"",
		},
		{
			input:   `"2017-01-02T04:04:05.000000000+01:00[[Europe/Berlin]]"`,
			wantErr: "parsing time \"2017-01-02T04:04:05.000000000+01:00[\": extra text: \"[\"",
		},
	} {
		var gotDateTime *datetime.DateTime
		err := json.Unmarshal([]byte(currCase.input), &gotDateTime)
		assert.EqualError(t, err, currCase.wantErr, "Case %d", i)
	}
}
