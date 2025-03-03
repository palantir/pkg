// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/palantir/pkg/yamlpatch/internal/yamlpatchcommon"
	"github.com/palantir/pkg/yamlpatch/yamlpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyYAMLPatch(t *testing.T) {
	yamlpatchcommon.RunApplyYAMLPatchTests(
		t,
		"goccy",
		newGoccyYAMLLibrary(
			YAMLEncodeOption(yaml.IndentSequence(false)),
		),
	)
}

func TestApplyYAMLPatch_CustomObjectTest(t *testing.T) {
	yamlpatchcommon.RunApplyYAMLPatchCustomObjectTests(
		t,
		"goccy",
		New(
			YAMLEncodeOption(yaml.IndentSequence(false)),
		),
	)
}

func TestApplyYAMLPatch_AddOrReplace(t *testing.T) {
	for _, tc := range []struct {
		name           string
		in             string
		yamllibOptions []YAMLOption
		patch          yamlpatch.Patch
		want           string
	}{
		{
			name: "add element to map that contains another entry matches indentation",
			in: `top-level-one-indent-0:
  one-indent-1:
    - one-indent-2:
        one-indent-3:
          one-indent-4: four-value
top-level-two-indent-0:
  two-indent-1:
    two-indent-2:
      two-indent-3: three-value
`,
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/top-level-two-indent-0/two-indent-1/two-indent-2/two-indent-3-val-2"),
					From:    nil,
					Value:   "two-value",
					Comment: "",
				},
			},
			want: `top-level-one-indent-0:
  one-indent-1:
    - one-indent-2:
        one-indent-3:
          one-indent-4: four-value
top-level-two-indent-0:
  two-indent-1:
    two-indent-2:
      two-indent-3: three-value
      two-indent-3-val-2: two-value
`,
		},
		{
			name: "add single element to non-flow list",
			in: `my-list:
  - one
  - two
  `,
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list:
  - one
  - two
  - three
`,
		},
		{
			name: "add single element to empty independent flow list",
			in: `[]
  `,
			yamllibOptions: []YAMLOption{
				UseNonFlowWhenModifyingEmptyContainer(false),
			},
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `[three]
`,
		},
		{
			name: "add single element to empty flow list value",
			in: `my-list: []
  `,
			yamllibOptions: []YAMLOption{
				UseNonFlowWhenModifyingEmptyContainer(false),
			},
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list: [three]
`,
		},
		{
			name: "add single element to empty flow list value on different line than key",
			in: `my-list:
[]
  `,
			yamllibOptions: []YAMLOption{
				UseNonFlowWhenModifyingEmptyContainer(false),
			},
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			// writing in non-flow style forces value on same line
			want: `my-list: [three]
`,
		},
		{
			name: "add single element to empty flow list value in non-flow mode",
			in: `my-list: []
  `,
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list:
  - three
`,
		},
		{
			name: "add single element to empty flow list value in non-flow mode with indent off",
			in: `my-list: []
  `,
			yamllibOptions: []YAMLOption{
				YAMLEncodeOption(yaml.IndentSequence(false)),
			},
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list:
- three
`,
		},
		{
			name: "add single element to empty flow list value in non-flow mode with value on new line",
			in: `my-list:
  []
  `,
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list:
  - three
`,
		},
		{
			name: "add single element to non-empty flow list value",
			in: `my-list: ["one"]
  `,
			patch: yamlpatch.Patch{
				{
					Type:    yamlpatch.OperationAdd,
					Path:    yamlpatch.MustParsePath("/my-list/-"),
					From:    nil,
					Value:   "three",
					Comment: "",
				},
			},
			want: `my-list: ["one", three]
`,
		},
		{
			name: "set element on empty flow list value in non-flow mode",
			in: `my-list: []
  `,
			patch: yamlpatch.Patch{
				{
					Type: yamlpatch.OperationReplace,
					Path: yamlpatch.MustParsePath("/my-list"),
					From: nil,
					Value: []string{
						"new",
					},
					Comment: "",
				},
			},
			want: `my-list:
  - new
`,
		},
		{
			name: "set elements on non-empty non-flow list value matches previous indent level that exists",
			in: `my-list:
  - old
  `,
			patch: yamlpatch.Patch{
				{
					Type: yamlpatch.OperationReplace,
					Path: yamlpatch.MustParsePath("/my-list"),
					From: nil,
					Value: []string{
						"new",
					},
					Comment: "",
				},
			},
			want: `my-list:
  - new
`,
		},
		{
			name: "set elements on non-empty non-flow list value matches previous indent level that is greater than set",
			in: `my-list:
    - old
  `,
			patch: yamlpatch.Patch{
				{
					Type: yamlpatch.OperationReplace,
					Path: yamlpatch.MustParsePath("/my-list"),
					From: nil,
					Value: []string{
						"new",
					},
					Comment: "",
				},
			},
			want: `my-list:
    - new
`,
		},
		{
			name: "set elements on non-empty non-flow list value matches previous indent level that doesn't indent",
			in: `my-list:
- old
  `,
			patch: yamlpatch.Patch{
				{
					Type: yamlpatch.OperationReplace,
					Path: yamlpatch.MustParsePath("/my-list"),
					From: nil,
					Value: []string{
						"new",
					},
					Comment: "",
				},
			},
			want: `my-list:
- new
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yamlPatcher := New(tc.yamllibOptions...)

			out, err := yamlPatcher.Apply([]byte(tc.in), tc.patch)
			require.NoError(t, err)

			assert.Equal(t, tc.want, string(out))
		})
	}
}
