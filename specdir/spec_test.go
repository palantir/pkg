// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package specdir_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/osutil"
	"github.com/palantir/pkg/specdir"
)

func TestValidateSpec(t *testing.T) {
	tmpDir, cleanup, err := dirs.TempDir("", "")
	defer cleanup()
	require.NoError(t, err)

	for i, currCase := range []struct {
		dirToValidate string
		spec          specdir.LayoutSpec
		values        specdir.TemplateValues
		pathsToCreate map[string]specdir.PathType
		expectedError string
	}{
		{
			dirToValidate: "root",
			spec:          specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), ""), true),
			pathsToCreate: map[string]specdir.PathType{
				"root": specdir.DirPath,
			},
		},
		{
			dirToValidate: "root",
			spec:          specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), ""), true),
			expectedError: fmt.Sprintf(`^.+%s does not exist$`, osutil.MakeValidRegexPath("/root")),
		},
		{
			dirToValidate: "rootNotPartOfSpec",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), "",
				specdir.Dir(specdir.LiteralName("child"), ""),
			), false),
			pathsToCreate: map[string]specdir.PathType{
				"rootNotPartOfSpec":       specdir.DirPath,
				"rootNotPartOfSpec/child": specdir.DirPath,
			},
		},
		{
			dirToValidate: "dirWithSingleFile",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithSingleFile"), "",
				specdir.File(specdir.LiteralName("child"), ""),
			), false),
			pathsToCreate: map[string]specdir.PathType{
				"dirWithSingleFile":       specdir.DirPath,
				"dirWithSingleFile/child": specdir.FilePath,
			},
		},
		{
			dirToValidate: "dirWithOptionalDirMissing",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithOptionalDirMissing"), "",
				specdir.OptionalDir(specdir.LiteralName("child")),
			), true),
			pathsToCreate: map[string]specdir.PathType{
				"dirWithOptionalDirMissing": specdir.DirPath,
			},
		},
		{
			dirToValidate: "dirWithOptionalDirPresent",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithOptionalDirPresent"), "",
				specdir.OptionalDir(specdir.LiteralName("child")),
			), true),
			pathsToCreate: map[string]specdir.PathType{
				"dirWithOptionalDirPresent":       specdir.DirPath,
				"dirWithOptionalDirPresent/child": specdir.DirPath,
			},
		},
		{
			dirToValidate: "dirWithWrongChildType",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithWrongChildType"), "",
				specdir.OptionalDir(specdir.LiteralName("child")),
			), true),
			pathsToCreate: map[string]specdir.PathType{
				"dirWithWrongChildType":       specdir.DirPath,
				"dirWithWrongChildType/child": specdir.FilePath,
			},
			expectedError: fmt.Sprintf(
				`^isDir for %s returned wrong value: expected true, was false$`,
				osutil.MakeValidRegexPath("dirWithWrongChildType/child"),
			),
		},
		{
			dirToValidate: "templateKeyName",
			values: map[string]string{
				"product": "foo_product",
			},
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("templateKeyName"), "",
				specdir.Dir(specdir.TemplateName("product"), ""),
			), true),
			pathsToCreate: map[string]specdir.PathType{
				"templateKeyName":             specdir.DirPath,
				"templateKeyName/foo_product": specdir.DirPath,
			},
		},
		{
			dirToValidate: "templateName",
			values: map[string]string{
				"product": "foo_product",
				"version": "1.0.0",
			},
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("templateName"), "",
				specdir.Dir(specdir.CompositeName(specdir.TemplateName("product"), specdir.LiteralName("-"), specdir.TemplateName("version")), ""),
			), true),
			pathsToCreate: map[string]specdir.PathType{
				"templateName":                   specdir.DirPath,
				"templateName/foo_product-1.0.0": specdir.DirPath,
			},
		},
	} {
		currCaseTmpDir, err := ioutil.TempDir(tmpDir, "")
		require.NoError(t, err)

		createDirectoryStructure(t, currCaseTmpDir, currCase.pathsToCreate)
		err = currCase.spec.Validate(filepath.Join(currCaseTmpDir, currCase.dirToValidate), currCase.values)
		if currCase.expectedError == "" {
			assert.NoError(t, err, "Case %d", i)
		} else {
			assert.Regexp(t, regexp.MustCompile(currCase.expectedError), err.Error(), "Case %d", i)
		}
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	tmpDir, cleanup, err := dirs.TempDir("", "")
	defer cleanup()
	require.NoError(t, err)

	for i, currCase := range []struct {
		rootDirForCreation string
		spec               specdir.LayoutSpec
		pathsToCreate      map[string]specdir.PathType
		includeOptional    bool
		expectedPaths      map[string]specdir.PathType
		unexpectedPaths    []string
		expectedError      string
	}{
		{
			rootDirForCreation: "root",
			spec:               specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), ""), true),
			expectedPaths: map[string]specdir.PathType{
				"root": specdir.DirPath,
			},
		},
		{
			rootDirForCreation: "wrongName",
			spec:               specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), ""), true),
			expectedPaths: map[string]specdir.PathType{
				"root": specdir.DirPath,
			},
			expectedError: fmt.Sprintf(
				`^.+%s is not a path to root$`,
				osutil.MakeValidRegexPath("/wrongName"),
			),
		},
		{
			rootDirForCreation: "rootNotPartOfSpec",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("root"), "",
				specdir.Dir(specdir.LiteralName("child"), ""),
			), false),
			expectedPaths: map[string]specdir.PathType{
				"rootNotPartOfSpec":       specdir.DirPath,
				"rootNotPartOfSpec/child": specdir.DirPath,
			},
		},
		{
			rootDirForCreation: "dirWithSingleFile",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithSingleFile"), "",
				specdir.File(specdir.LiteralName("child"), ""),
			), false),
			expectedPaths: map[string]specdir.PathType{
				"dirWithSingleFile": specdir.DirPath,
			},
		},
		{
			rootDirForCreation: "dirWithOptionalDirNoCreate",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithOptionalDirNoCreate"), "",
				specdir.OptionalDir(specdir.LiteralName("child")),
			), true),
			expectedPaths: map[string]specdir.PathType{
				"dirWithOptionalDirNoCreate": specdir.DirPath,
			},
			unexpectedPaths: []string{
				"dirWithOptionalDirNoCreate/child",
			},
		},
		{
			rootDirForCreation: "dirWithOptionalDirCreate",
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("dirWithOptionalDirCreate"), "",
				specdir.OptionalDir(specdir.LiteralName("child")),
			), true),
			includeOptional: true,
			expectedPaths: map[string]specdir.PathType{
				"dirWithOptionalDirCreate":       specdir.DirPath,
				"dirWithOptionalDirCreate/child": specdir.DirPath,
			},
		},
		{
			rootDirForCreation: "failIfFileExistsWhereDirToBeCreated",
			pathsToCreate: map[string]specdir.PathType{
				"failIfFileExistsWhereDirToBeCreated":       specdir.DirPath,
				"failIfFileExistsWhereDirToBeCreated/child": specdir.FilePath,
			},
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("failIfFileExistsWhereDirToBeCreated"), "",
				specdir.Dir(specdir.LiteralName("child"), ""),
			), true),
			includeOptional: true,
			expectedError: fmt.Sprintf(
				`^failed to create directory .+%s: mkdir .*%s: %s$`,
				osutil.MakeValidRegexPath("/failIfFileExistsWhereDirToBeCreated/child"),
				osutil.MakeValidRegexPath("/failIfFileExistsWhereDirToBeCreated/child"),
				osutil.GetNotADirErrorMsg(),
			),
		},
		{
			rootDirForCreation: "okIfDirAlreadyExists",
			pathsToCreate: map[string]specdir.PathType{
				"okIfDirAlreadyExists":       specdir.DirPath,
				"okIfDirAlreadyExists/child": specdir.DirPath,
			},
			spec: specdir.NewLayoutSpec(specdir.Dir(specdir.LiteralName("okIfDirAlreadyExists"), "",
				specdir.Dir(specdir.LiteralName("child"), ""),
			), true),
		},
	} {
		currCaseTmpDir, err := ioutil.TempDir(tmpDir, "")
		require.NoError(t, err)

		rootForCreation := filepath.Join(currCaseTmpDir, currCase.rootDirForCreation)
		err = os.Mkdir(rootForCreation, 0755)
		require.NoError(t, err)

		createDirectoryStructure(t, currCaseTmpDir, currCase.pathsToCreate)

		err = currCase.spec.CreateDirectoryStructure(rootForCreation, nil, currCase.includeOptional)
		if currCase.expectedError == "" {
			assert.NoError(t, err, "Case %d", i)

			for currPath, pathType := range currCase.expectedPaths {
				info, err := os.Stat(filepath.Join(currCaseTmpDir, currPath))
				assert.NoError(t, err, "Case %d", i)
				assert.Equal(t, bool(pathType), !info.IsDir(), "Case %d", i)
			}

			for _, currPath := range currCase.unexpectedPaths {
				_, err = os.Stat(filepath.Join(currCaseTmpDir, currPath))
				assert.True(t, os.IsNotExist(err), "Case %d")
			}
		} else {
			assert.Regexp(t, regexp.MustCompile(currCase.expectedError), err.Error(), "Case %d", i)
		}
	}
}

func createDirectoryStructure(t *testing.T, tmpDir string, paths map[string]specdir.PathType) {
	for currPath, pathType := range paths {
		currPath = filepath.Join(tmpDir, currPath)

		dirToCreate := currPath
		if pathType == specdir.FilePath {
			dirToCreate = filepath.Dir(currPath)
		}
		err := os.MkdirAll(dirToCreate, 0755)
		require.NoError(t, err)

		if pathType == specdir.FilePath {
			err := ioutil.WriteFile(currPath, []byte("test file"), 0644)
			require.NoError(t, err)
		}
	}
}
