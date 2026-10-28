// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bearertoken_test

import (
	"testing"

	"github.com/palantir/pkg/bearertoken"
)

func TestToken_AppendText(t *testing.T) {
	got, err := bearertoken.Token("tok").AppendText([]byte("prefix:"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "prefix:tok" {
		t.Fatalf("got %q", got)
	}
}
