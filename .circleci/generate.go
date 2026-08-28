//go:build generate

// This program writes the inputs for the shared CircleCI and Autorelease templates for the "pkg" repository.
// Output paths are resolved relative to this source file rather than the process working directory.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const (
	templateShPath  = "template.sh"
	autoreleasePath = "../.palantir/autorelease.yml"
)

// Every module of this repository is a peer in its own directory and the root of the repository is not a godel project,
// so all of the modules are listed in "MODULES": "ADDITIONAL_MODULES" would render jobs for a module at the root of the
// repository that does not exist. "PRIMARY_BRANCH" is set because the template defaults to "develop", while this
// repository uses "master". "USE_GOPATH_WD" preserves the working directories used by the old generated configuration.
const templateShTemplateContent = `#!/usr/bin/env bash
export CIRCLECI_TEMPLATE=go-library-oss
export PRIMARY_BRANCH=master
export USE_GOPATH_WD=true
export MODULES={{join . ","}}
`

const autoreleaseTemplateContent = `version: 3

groups:
  __default__:
    paths: ["."]
    tag_prefix: "v"
{{range .}}  {{.}}:
    paths: ["{{.}}"]
    tag_prefix: "{{.}}/v"
{{end}}
intoto:
  disable: true

options:
  # Multi-group repos reject plain label releases. Release only groups with unreleased changelog entries.
  release_recommended_groups: true
  allowed_branches: ["master"]
`

var (
	templateShTemplate  = template.Must(template.New(templateShPath).Funcs(template.FuncMap{"join": strings.Join}).Parse(templateShTemplateContent))
	autoreleaseTemplate = template.Must(template.New(autoreleasePath).Parse(autoreleaseTemplateContent))
)

func main() {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to determine generator source path")
	}
	sourceDir := filepath.Dir(sourceFile)

	mods, err := modules(filepath.Dir(sourceDir))
	if err != nil {
		panic(err)
	}
	if err := writeGeneratedFile(sourceDir, templateShPath, templateShTemplate, mods, 0755); err != nil {
		panic(err)
	}
	if err := writeGeneratedFile(sourceDir, autoreleasePath, autoreleaseTemplate, mods, 0644); err != nil {
		panic(err)
	}
}

func writeGeneratedFile(sourceDir, relativePath string, tmpl *template.Template, modDirs []string, perm os.FileMode) error {
	outBuf := &bytes.Buffer{}
	if err := tmpl.Execute(outBuf, modDirs); err != nil {
		return fmt.Errorf("failed to render %s: %w", relativePath, err)
	}
	outputPath := filepath.Join(sourceDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", outputPath, err)
	}
	if err := os.WriteFile(outputPath, outBuf.Bytes(), perm); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}
	if err := os.Chmod(outputPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", outputPath, err)
	}
	return nil
}

func modules(parentDir string) ([]string, error) {
	fis, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	var dirNames []string
	for _, fi := range fis {
		if !fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		dirNames = append(dirNames, fi.Name())
	}
	return dirNames, nil
}
