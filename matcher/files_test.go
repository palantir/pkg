// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matcher_test

import (
	"os"
	"path"
	"testing"

	"github.com/palantir/pkg/matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFiles(t *testing.T) {
	cases := []struct {
		include       matcher.Matcher
		exclude       matcher.Matcher
		filesToCreate map[string]string
		want          []string
	}{
		{
			include: matcher.Name(`.+\.go`),
			filesToCreate: map[string]string{
				"notgo.txt":        "",
				"isgo.go":          "",
				".hidden.go":       "",
				".hidden/inner.go": "",
				"indir/isgo.go":    "",
			},
			want: []string{
				".hidden/inner.go",
				".hidden.go",
				"indir/isgo.go",
				"isgo.go",
			},
		},
		{
			include: matcher.Name(`.+\.go`),
			exclude: matcher.Hidden(),
			filesToCreate: map[string]string{
				"notgo.txt":        "",
				"isgo.go":          "",
				".hidden.go":       "",
				"indir/isgo.go":    "",
				".hidden/inner.go": "",
			},
			want: []string{
				"indir/isgo.go",
				"isgo.go",
			},
		},
	}

	for i, currCase := range cases {
		tmpDir := t.TempDir()
		createFiles(t, tmpDir, currCase.filesToCreate)

		got, err := matcher.ListFiles(tmpDir, currCase.include, currCase.exclude)
		require.NoError(t, err, "Case %d", i)

		assert.Equal(t, currCase.want, got, "Case %d", i)
	}
}

func createFiles(t *testing.T, tmpDir string, files map[string]string) {
	for currFile, currContent := range files {
		err := os.MkdirAll(path.Join(tmpDir, path.Dir(currFile)), 0755)
		require.NoError(t, err)

		err = os.WriteFile(path.Join(tmpDir, currFile), []byte(currContent), 0644)
		require.NoError(t, err)
	}
}
