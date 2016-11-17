// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/matcher"
)

func TestMatcherCreationFunction(t *testing.T) {
	for i, currFn := range []func(matchers ...matcher.Matcher) matcher.Matcher{
		matcher.Any,
		matcher.All,
	} {
		m := currFn()
		assert.False(t, m.Match("foo"), "Case %d: unexpected match for empty call", i)

		m = currFn(nil)
		assert.False(t, m.Match("foo"), "Case %d: unexpected match for nil call", i)
	}
}

func TestNameMatcher(t *testing.T) {
	for i, currCase := range []struct {
		matcherArgs []string
		path        string
		want        bool
	}{
		// name match includes subdirectories
		{[]string{"foo"}, "foo/bar/regular", true},
		// name match matches on subdirectories
		{[]string{"foo"}, "bar/foo/inner", true},
		// matches if any matcher matches
		{[]string{"bar", "foo"}, "foo/bar", true},
		// full match required
		{[]string{"foo"}, "fooLongerName/inner", false},
		// regexps work
		{[]string{"foo.*"}, "fooLongerName/inner", true},
		// matches occur on name parts only (does not match across directory boundaries). "foo/bar" is checked
		// against "bar" and "foo", and therefore does not match.
		{[]string{"foo/bar"}, "foo/bar", false},
	} {
		m := matcher.Name(currCase.matcherArgs...)
		got := m.Match(currCase.path)
		assert.Equal(t, currCase.want, got, "Case %d", i)
	}
}

func TestPathMatcher(t *testing.T) {
	for i, currCase := range []struct {
		matcherArgs []string
		path        string
		want        bool
	}{
		{[]string{"foo"}, "foo/bar/regular", true},
		{[]string{"foo"}, "bar/foo/inner", false},
		// full match required
		{[]string{"foo"}, "fooLongerName/inner", false},
		// glob matching
		{[]string{"foo*"}, "fooName", true},
		// glob matching matches subdirectories
		{[]string{"foo*"}, "fooLongerName/inner", true},
		// glob matching matches subdirectories
		{[]string{"foo/*/baz"}, "foo/bar/baz/inner", true},
		// globs do not match through separators
		{[]string{"foo*/bar"}, "fooz/baz/bar", false},
		{[]string{"foo/bar"}, "foo/bar", true},
		{[]string{"foo/bar"}, "/foo/bar", false},
	} {
		m := matcher.Path(currCase.matcherArgs...)
		got := m.Match(currCase.path)
		assert.Equal(t, currCase.want, got, "Case %d", i)
	}
}

func TestHiddenMatcher(t *testing.T) {
	m := matcher.Hidden()

	for i, currCase := range []struct {
		path string
		want bool
	}{
		{"foo/bar/regular", false},
		{"foo/bar/.hidden", true},
		{"foo/.bar/inHidden", true},
		{"foo/.bar/inHidden", true},
	} {
		got := m.Match(currCase.path)
		assert.Equal(t, currCase.want, got, "Case %d", i)
	}
}

func TestNotMatcher(t *testing.T) {
	for i, currCase := range []struct {
		m    matcher.Matcher
		path string
		want bool
	}{
		{matcher.Any(matcher.Name("foo")), "foo", false},
		{matcher.Any(matcher.Name("foo")), "bar", true},
		{matcher.Any(matcher.Name("foo"), matcher.Name("bar")), "bar", false},
		{matcher.Any(matcher.Name("foo"), matcher.Name("bar")), "baz", true},
		{matcher.Any(matcher.Path("foo/bar")), "foo/bar/baz", false},
		{matcher.Any(matcher.Path("foo/bar")), "baz/foo/bar", true},
	} {
		m := matcher.Not(currCase.m)
		got := m.Match(currCase.path)
		assert.Equal(t, currCase.want, got, "Case %d", i)
	}
}
