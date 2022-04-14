//go:build generate
// +build generate

// This program prints the CircleCI configuration for the "pkg" repository. Standard way to run it is to run
// "go run generate.go {{parentDir}} > config.yml".
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const (
	headerTemplateContent = `version: 2.1

orbs:
  go: palantir/go@0.0.29
  godel: palantir/godel@0.0.29

homepath: &homepath
  homepath: /home/circleci

gopath: &gopath
  gopath: /home/circleci/go

working_directory: &working_directory
  working_directory: /home/circleci/go/src/github.com/palantir/pkg

executors:
  circleci-go:
    docker:
      - image: cimg/go:1.16-browsers
    <<: *working_directory

jobs:
  verify-circleci:
    <<: *working_directory
    docker:
      - image: cimg/go:1.16-browsers
    resource_class: small
    steps:
      - checkout
      - run: go version
      - run: go run .circleci/generate.go .
      - run: diff  <(cat .circleci/config.yml) <(go run .circleci/generate.go .)
  circle-all:
    docker:
      - image: cimg/go:1.16-browsers
    resource_class: small
    steps:
      - run: echo "All required jobs run successfully"

workflows:
  version: 2
  verify-test:
    jobs:
      - verify-circleci
      - circle-all:
          requires: [ verify-circleci{{ range $job := .JobNames }}, {{ $job }}{{end}} ]
`

	moduleTemplateContent = `
      # {{.Module}}
      - godel/verify:
          name: {{.Module}}-verify
          executor: circleci-go
          <<: *homepath
          <<: *gopath
          include-tests: true
      - godel/test:
          name: {{.Module}}-test-go-prev
          <<: *checkout-path
          executor: circleci-go
          <<: *homepath
          <<: *gopath
          go-prev-version: 1
          requires:
            - {{.Module}}-verify
`
)

type TemplateObject struct {
	Module             string
	CurrGoVersion      string
	PrevGoVersion      string
	PrevGoMajorVersion string
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
	configYML, err := createConfigYML(mods, "1.17.5", "1.16.12")
	if err != nil {
		panic(err)
	}
	fmt.Print(configYML)
}

var headerTemplate, moduleTemplate *template.Template

func init() {
	var err error
	headerTemplate, err = template.New("headerTemplate").Parse(headerTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create headerTemplate template: %v", err))
	}
	moduleTemplate, err = template.New("moduleTemplate").Parse(moduleTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create moduleTemplate template: %v", err))
	}
}

func createConfigYML(modDirs []string, currGoVersion, prevGoVersion string) (string, error) {
	prevParts := strings.Split(prevGoVersion, ".")
	if len(prevParts) < 2 {
		return "", fmt.Errorf("prevGoVersion must have at least 2 parts separated by a period: %s", prevGoVersion)
	}
	prevGoMajorVersion := strings.Join(prevParts[:2], ".")

	jobNames := make([]string, 0, len(modDirs)*2)
	for _, modDir := range modDirs {
		jobNames = append(jobNames, modDir+"-verify", modDir+"-test-go-"+prevGoMajorVersion)
	}
	outBuf := &bytes.Buffer{}
	if err := headerTemplate.Execute(outBuf, map[string]interface{}{
		"CurrGoVersion": currGoVersion,
		"JobNames":      jobNames,
	}); err != nil {
		return "", fmt.Errorf("failed to execute headerTemplate template: %v", err)
	}
	for _, modDir := range modDirs {
		if err := moduleTemplate.Execute(outBuf, TemplateObject{
			Module:             modDir,
			CurrGoVersion:      currGoVersion,
			PrevGoVersion:      prevGoVersion,
			PrevGoMajorVersion: prevGoMajorVersion,
		}); err != nil {
			return "", fmt.Errorf("failed to execute moduleTemplate template: %v", err)
		}
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
