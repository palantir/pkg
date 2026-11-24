// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datetime_test

import (
	"encoding"
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"testing"
	"time"

	"github.com/palantir/pkg/datetime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ encoding.TextAppender = datetime.DateTime{}

func TestDateTimeAppendText(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value time.Time
	}{
		{name: "zero"},
		{name: "UTC", value: time.Date(2026, 10, 23, 1, 2, 3, 123456789, time.UTC)},
		{name: "offset", value: time.Date(2026, 10, 23, 1, 2, 3, 100000000, time.FixedZone("", 19800))},
		{name: "before epoch", value: time.Unix(-1, 123456789).UTC()},
		{name: "negative year", value: time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "large year", value: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "large offset", value: time.Date(2026, 10, 23, 1, 2, 3, 0, time.FixedZone("", 25*3600))},
		{name: "offset seconds", value: time.Date(2026, 10, 23, 1, 2, 3, 0, time.FixedZone("", 45))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := datetime.DateTime(tc.value)
			want := tc.value.Format(time.RFC3339Nano)
			got, err := d.AppendText(nil)
			require.NoError(t, err)
			assert.Equal(t, want, string(got))

			buf := make([]byte, len("prefix:"), 128)
			copy(buf, "prefix:")
			got, err = d.AppendText(buf)
			require.NoError(t, err)
			assert.Equal(t, "prefix:"+want, string(got))
			assert.Equal(t, "prefix:", string(buf))

			legacy, err := json.Marshal(d)
			require.NoError(t, err)
			encoded, err := jsonv2.Marshal(d)
			require.NoError(t, err)
			assert.Equal(t, legacy, encoded)
		})
	}
}
