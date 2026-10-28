// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

package gotoolchain_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/palantir/pkg/gotoolchain"
)

type selectionCase struct {
	Name      string `json:"name"`
	GoMod     string `json:"go_mod"`
	Toolchain string `json:"toolchain"`
	Error     bool   `json:"error"`
}

func loadSelectionCases(t testing.TB) []selectionCase {
	content, err := os.ReadFile(filepath.Join("testdata", "cases.json"))
	requireNoError(t, err)
	var cases []selectionCase
	requireNoError(t, json.Unmarshal(content, &cases))
	if len(cases) == 0 {
		t.Fatal("expected at least one selection case")
	}
	return cases
}

func writeGoMod(t testing.TB, content string) string {
	path := filepath.Join(t.TempDir(), "go.mod")
	requireNoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func requireNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFromGoMod(t *testing.T) {
	for _, tc := range loadSelectionCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := gotoolchain.FromGoMod(writeGoMod(t, tc.GoMod))
			if tc.Error {
				if err == nil {
					t.Error("expected an error")
				}
				return
			}
			requireNoError(t, err)
			if got != tc.Toolchain {
				t.Errorf("got %q, want %q", got, tc.Toolchain)
			}
		})
	}
}

func TestFromGoModMissingFile(t *testing.T) {
	_, err := gotoolchain.FromGoMod(filepath.Join(t.TempDir(), "go.mod"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("got error %v, want %v", err, os.ErrNotExist)
	}
}

var normalizeCases = []struct{ version, exact, language string }{
	{"1.27.3", "go1.27.3", "go1.27.3"}, {"go1.27.3", "go1.27.3", "go1.27.3"},
	{"1.27", "", "go1.27.0"}, {"1.20", "go1.20", "go1.20.0"}, {"1.0", "go1.0", "go1.0.0"},
	{"1.100", "", "go1.100.0"}, {"1.27rc1", "go1.27rc1", "go1.27rc1"},
	{"1.27beta2", "go1.27beta2", "go1.27beta2"}, {"1.20.14", "go1.20.14", "go1.20.14"},
	{"", "", ""}, {"go", "", ""}, {"2.0.0", "", ""}, {"1.027", "", ""},
	{"1.27rc0", "", ""}, {"1.27alpha1", "", ""}, {"1.27.3-custom", "", ""}, {"latest", "", ""},
}

func TestNormalize(t *testing.T) {
	for _, tc := range normalizeCases {
		for mode, want := range map[gotoolchain.NormalizeMode]string{gotoolchain.NormalizeExact: tc.exact, gotoolchain.NormalizeLanguage: tc.language} {
			got, err := gotoolchain.Normalize(tc.version, mode)
			if want == "" {
				if err == nil {
					t.Errorf("Normalize(%q, %v) expected an error", tc.version, mode)
				}
				continue
			}
			requireNoError(t, err)
			if got != want {
				t.Errorf("Normalize(%q, %v) = %q, want %q", tc.version, mode, got, want)
			}
		}
	}
}

func TestSelectedToolchain(t *testing.T) {
	for _, tc := range []struct {
		d    gotoolchain.Directives
		want string
	}{
		{gotoolchain.Directives{Go: "1.27"}, "go1.27.0"},
		{gotoolchain.Directives{Go: "1.27", Toolchain: "go1.27.3"}, "go1.27.3"},
		{gotoolchain.Directives{Go: "1.27", Toolchain: "default"}, "go1.27.0"},
		{gotoolchain.Directives{Toolchain: "go1.27.3"}, ""},
		{gotoolchain.Directives{Go: "1.27", Toolchain: "go1.27"}, ""},
	} {
		got, err := tc.d.SelectedToolchain()
		if tc.want == "" {
			if err == nil {
				t.Errorf("SelectedToolchain(%+v) expected an error", tc.d)
			}
		} else {
			requireNoError(t, err)
			if got != tc.want {
				t.Errorf("SelectedToolchain(%+v) = %q, want %q", tc.d, got, tc.want)
			}
		}
	}
}
