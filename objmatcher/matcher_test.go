// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objmatcher_test

import (
	"testing"

	"github.com/palantir/pkg/objmatcher"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	label string
	num   int
}

func TestEqualsMatcher(t *testing.T) {
	for i, currCase := range []struct {
		name        string
		matcherWant interface{}
		given       interface{}
		wantErr     string
	}{
		{
			name:        "strings match",
			matcherWant: "foo",
			given:       "foo",
		},
		{
			name:        "strings mismatch",
			matcherWant: "foo",
			given:       "bar",
			wantErr:     "want: string(foo)\ngot:  string(bar)",
		},
		{
			name:        "strings mismatch",
			matcherWant: "foo",
			given:       5,
			wantErr:     "want: string(foo)\ngot:  int(5)",
		},
		{
			name: "structs match",
			matcherWant: testStruct{
				label: "foo",
				num:   13,
			},
			given: testStruct{
				label: "foo",
				num:   13,
			},
		},
		{
			name: "structs mismatch",
			matcherWant: testStruct{
				label: "foo",
				num:   13,
			},
			given: testStruct{
				label: "bar",
				num:   13,
			},
			wantErr: "want: objmatcher_test.testStruct({label:foo num:13})\ngot:  objmatcher_test.testStruct({label:bar num:13})",
		},
		{
			name: "maps match",
			matcherWant: map[string]interface{}{
				"outer-foo": "bar",
				"outer-num": 5,
				"struct": testStruct{
					label: "inner-bar",
					num:   13,
				},
			},
			given: map[string]interface{}{
				"outer-foo": "bar",
				"outer-num": 5,
				"struct": testStruct{
					label: "inner-bar",
					num:   13,
				},
			},
		},
	} {
		gotErr := objmatcher.NewEqualsMatcher(currCase.matcherWant).Matches(currCase.given)
		if currCase.wantErr == "" {
			assert.NoError(t, gotErr, "Case %d: %v", i, currCase.name)
		} else {
			assert.EqualError(t, gotErr, currCase.wantErr, "Case %d: %v", i, currCase.name)
		}
	}
}

func TestMapMatcher(t *testing.T) {
	for i, tc := range []struct {
		name    string
		matcher objmatcher.MapMatcher
		given   interface{}
		wantErr string
	}{
		{
			name: "map matches",
			matcher: objmatcher.MapMatcher(map[string]objmatcher.Matcher{
				"key": objmatcher.NewEqualsMatcher("value"),
			}),
			given: map[string]interface{}{
				"key": "value",
			},
		},
	} {
		gotErr := tc.matcher.Matches(tc.given)
		if tc.wantErr == "" {
			assert.NoError(t, gotErr, "Case %d: %v", i, tc.name)
		} else {
			assert.EqualError(t, gotErr, tc.wantErr, "Case %d: %v", i, tc.name)
		}
	}
}

func TestSliceMatcher(t *testing.T) {
	for i, tc := range []struct {
		name    string
		matcher objmatcher.SliceMatcher
		given   interface{}
		wantErr string
	}{
		{
			name: "strings slice matches",
			matcher: objmatcher.SliceMatcher([]objmatcher.Matcher{
				objmatcher.NewEqualsMatcher("hello"),
			}),
			given: []string{
				"hello",
			},
		},
		{
			name: "unequal slice length fails",
			matcher: objmatcher.SliceMatcher([]objmatcher.Matcher{
				objmatcher.NewEqualsMatcher("hello"),
			}),
			given: []string{
				"hello",
				"goodbye",
			},
			wantErr: "want: [equals(string(hello))]\ngot:  [hello goodbye]\nsize 1 != 2",
		},
	} {
		gotErr := tc.matcher.Matches(tc.given)
		if tc.wantErr == "" {
			assert.NoError(t, gotErr, "Case %d: %v", i, tc.name)
		} else {
			assert.EqualError(t, gotErr, tc.wantErr, "Case %d: %v", i, tc.name)
		}
	}
}
