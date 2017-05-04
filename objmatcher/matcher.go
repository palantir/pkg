// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objmatcher

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Matcher interface {
	Matches(in interface{}) error
}

type EqualsMatcher struct {
	Want interface{}
}

func NewEqualsMatcher(want interface{}) Matcher {
	wantMap, ok := want.(map[string]interface{})
	if ok {
		// if desired value is a map, return a mapMatcher that uses an equalsMatcher for all of the provided
		// key/value pairs. Results in more descriptive error messages on match failures.
		matcherMap := make(map[string]Matcher)
		for k, v := range wantMap {
			matcherMap[k] = &EqualsMatcher{Want: v}
		}
		return MapMatcher(matcherMap)
	}
	return &EqualsMatcher{Want: want}
}

func (m *EqualsMatcher) Matches(in interface{}) error {
	if !reflect.DeepEqual(m.Want, in) {
		return fmt.Errorf("want: %T(%+v)\ngot:  %T(%+v)", m.Want, m.Want, in, in)
	}
	return nil
}

func (m *EqualsMatcher) String() string {
	return fmt.Sprintf("equals(%T(%+v))", m.Want, m.Want)
}

type RegExpMatcher struct {
	WantRegexp string
}

func NewRegExpMatcher(want string) Matcher {
	return &RegExpMatcher{WantRegexp: want}
}

func (m *RegExpMatcher) Matches(in interface{}) error {
	str, ok := in.(string)
	if !ok {
		return fmt.Errorf("want to match regexp %s, but %T(%+v) is not a string", m.WantRegexp, in, in)
	}
	if !regexp.MustCompile(m.WantRegexp).MatchString(str) {
		return fmt.Errorf("regexp %s does not match %s", m.WantRegexp, str)
	}
	return nil
}

func (m RegExpMatcher) String() string {
	return fmt.Sprintf("matchesRegexp(%s)", m.WantRegexp)
}

type MapMatcher map[string]Matcher

func (m MapMatcher) Matches(in interface{}) error {
	inMap, ok := in.(map[string]interface{})
	if !ok {
		return fmt.Errorf("want: %+v\ngot:  %+v\n%T(%+v)is not a map", m, in, in, in)
	}
	if len(m) != len(inMap) {
		return fmt.Errorf("want: %+v\ngot:  %+v\nsize %d != %d", m, inMap, len(m), len(inMap))
	}
	for wantK, wantV := range m {
		gotV, ok := inMap[wantK]
		if !ok {
			return fmt.Errorf("want: %+v\ngot:  %+v\nexpected key %q is not present", m, inMap, wantK)
		}
		if err := wantV.Matches(gotV); err != nil {
			indented := strings.Replace("\n"+err.Error(), "\n", "\n\t", -1)
			return fmt.Errorf("want: %+v\ngot:  %+v\nvalue for key %q did not match:%s", m, inMap, wantK, indented)
		}
	}
	return nil
}
