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
	header = `version: 2.1

checkout-path: &checkout-path
  checkout-path: /home/circleci/go/src/github.com/palantir/pkg

orbs:
  go: palantir/go@0.0.29
  godel: palantir/godel@0.0.29

homepath: &homepath
  homepath: /home/circleci

gopath: &gopath
  gopath: /home/circleci/go

executors:`

	executorTemplateContent = `
  circleci-go-{{.Module}}:
    docker:
      - image: cimg/go:1.19-browsers
    working_directory: /home/circleci/go/src/github.com/palantir/pkg/{{.Module}}
`

	jobsWorkflowsTemplateContent = `
jobs:
  verify-circleci:
    working_directory: /home/circleci/go/src/github.com/palantir/pkg
    docker:
      - image: cimg/go:1.19-browsers
    resource_class: small
    steps:
      - checkout
      - run: go version
      - run: go run .circleci/generate.go .
      - run: diff  <(cat .circleci/config.yml) <(go run .circleci/generate.go .)
  circle-all:
    docker:
      - image: cimg/go:1.19-browsers
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
          executor: circleci-go-{{.Module}}
          <<: *checkout-path
          <<: *homepath
          <<: *gopath
          go-version-file: "../.palantir/go-version"
          include-tests: true
      - godel/test:
          name: {{.Module}}-test-go-prev
          executor: circleci-go-{{.Module}}
          <<: *checkout-path
          <<: *homepath
          <<: *gopath
          go-version-file: "../.palantir/go-version"
          go-prev-version: 1
          requires:
            - {{.Module}}-verify
`
)

type TemplateObject struct {
	Module string
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
	configYML, err := createConfigYML(mods)
	if err != nil {
		panic(err)
	}
	fmt.Print(configYML)
}

var (
	executorTemplate,
	jobsWorkflowsTemplate,
	moduleTemplate *template.Template
)

func init() {
	var err error
	executorTemplate, err = template.New("executorTemplate").Parse(executorTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create executorTemplate template: %v", err))
	}
	jobsWorkflowsTemplate, err = template.New("jobsWorkflowsTemplate").Parse(jobsWorkflowsTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create headerTemplate template: %v", err))
	}
	moduleTemplate, err = template.New("moduleTemplate").Parse(moduleTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create moduleTemplate template: %v", err))
	}
}

func createConfigYML(modDirs []string) (string, error) {
	jobNames := make([]string, 0, len(modDirs)*2)
	for _, modDir := range modDirs {
		jobNames = append(jobNames, modDir+"-verify", modDir+"-test-go-prev")
	}
	outBuf := &bytes.Buffer{}
	outBuf.WriteString(header)

	var modTemplates []TemplateObject
	for _, modDir := range modDirs {
		modTemplates = append(modTemplates, TemplateObject{
			Module: modDir,
		})
	}

	for _, modTemplate := range modTemplates {
		if err := executorTemplate.Execute(outBuf, modTemplate); err != nil {
			return "", fmt.Errorf("failed to execute executorTemplate template: %v", err)
		}
	}

	if err := jobsWorkflowsTemplate.Execute(outBuf, map[string]interface{}{
		"JobNames": jobNames,
	}); err != nil {
		return "", fmt.Errorf("failed to execute headerTemplate template: %v", err)
	}

	for _, modTemplate := range modTemplates {
		if err := moduleTemplate.Execute(outBuf, modTemplate); err != nil {
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
