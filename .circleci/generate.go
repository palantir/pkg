//go:build generate

// This program prints the "template.sh" that renders the CircleCI configuration for the "pkg" repository using the
// shared "go-library-oss" CircleCI template. Standard way to run it is to run "go run generate.go ../ > template.sh"
// from the directory that contains this file (corresponds to "go run generate.go {{parentDir}} > template.sh").
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

// Every module of this repository is a peer in its own directory and the root of the repository is not a godel project,
// so all of the modules are listed in "MODULES": "ADDITIONAL_MODULES" would render jobs for a module at the root of the
// repository that does not exist. "PRIMARY_BRANCH" is set because the template defaults to "develop", while this
// repository uses "master". "USE_GOPATH_WD" is set because the modules of this repository are verified in their GOPATH
// locations (that is, in "/home/circleci/go/src/github.com/palantir/pkg/{{module}}").
const templateShTemplateContent = `#!/usr/bin/env bash
export CIRCLECI_TEMPLATE=go-library-oss
export PRIMARY_BRANCH=master
export USE_GOPATH_WD=true
export MODULES={{.Modules}}
`

var templateShTemplate *template.Template

func init() {
	var err error
	templateShTemplate, err = template.New("templateShTemplate").Parse(templateShTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create templateShTemplate template: %v", err))
	}
}

func main() {
	if len(os.Args) < 2 {
		panic("parent directory must be provided as argument")
	}
	modParentDir := os.Args[1]
	mods, err := modules(modParentDir)
	if err != nil {
		panic(err)
	}
	templateSh, err := createTemplateSh(mods)
	if err != nil {
		panic(err)
	}
	fmt.Print(templateSh)
}

func createTemplateSh(modDirs []string) (string, error) {
	outBuf := &bytes.Buffer{}
	// the value of a parameter in "template.sh" must be a single token that does not contain any whitespace, so the
	// modules are joined using commas with no spaces
	if err := templateShTemplate.Execute(outBuf, map[string]interface{}{
		"Modules": strings.Join(modDirs, ","),
	}); err != nil {
		return "", fmt.Errorf("failed to execute templateShTemplate template: %v", err)
	}
	return outBuf.String(), nil
}

func modules(parentDir string) ([]string, error) {
	fis, err := ioutil.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
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
