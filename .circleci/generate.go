//go:build generate
// +build generate

// This program prints the CircleCI configuration for the "pkg" repository. Standard way to run it is to run
// "go run generate.go ../ > config.yml" from this directory (corresponds to "go run generate.go {{parentDir}} > config.yml").
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

orbs:
  go-jobs: palantir/go-jobs@0.8.0

image-version: &image-version "cimg/go:1.25.1-browsers"

checkout-path: &checkout-path
  path: /home/circleci/go/src/github.com/palantir/pkg

# Filter that matches all tags (will run on every build).
all-tags-filter: &all-tags-filter
  filters:
    tags:
      only: /.*/

# Filter that matches any branch besides primary branch and ignores all tags except for release candidates
pull-request-filter: &pull-request-filter
  filters:
    tags:
      only: /.*-rc.*/
    branches:
      ignore:
        - master

executors:`

	executorTemplateContent = `
  standard-executor-{{.Module}}:
    docker:
      - image: *image-version
    environment:
      GOTOOLCHAIN: local
    working_directory: /home/circleci/go/src/github.com/palantir/pkg/{{.Module}}
`

	requiresHeader = `
# The set of jobs that should be run on every build
requires_jobs: &requires_jobs
`

	requiresTemplateContent = `  - {{.Module}}-verify
`

	jobsWorkflowsTemplateContent = `
jobs:
  verify-circleci:
    working_directory: /home/circleci/go/src/github.com/palantir/pkg
    docker:
      - image: *image-version
    environment:
      GOTOOLCHAIN: local
    resource_class: small
    steps:
      - checkout
      - run: go version
      - run: go run .circleci/generate.go .
      - run: diff  <(cat .circleci/config.yml) <(go run .circleci/generate.go .)

workflows:
  version: 2
  verify-test:
    jobs:
      - verify-circleci
      - go-jobs/circle_all:
          name: "circle-all"
          image: "busybox:1.36.1"
          requires: *requires_jobs
          <<: *pull-request-filter
`

	moduleTemplateContent = `
      # {{.Module}}
      - go-jobs/godel_verify:
          name: {{.Module}}-verify
          executor: standard-executor-{{.Module}}
          setup_steps:
            - go-jobs/default_setup_steps:
                checkout_steps:
                  - checkout:
                      <<: *checkout-path
          <<: *all-tags-filter
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
	requiresTemplate,
	jobsWorkflowsTemplate,
	moduleTemplate *template.Template
)

func init() {
	var err error
	executorTemplate, err = template.New("executorTemplate").Parse(executorTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create executorTemplate template: %v", err))
	}
	requiresTemplate, err = template.New("requiresTemplate").Parse(requiresTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create requiresTemplate template: %v", err))
	}
	jobsWorkflowsTemplate, err = template.New("jobsWorkflowsTemplate").Parse(jobsWorkflowsTemplateContent)
	if err != nil {
		panic(fmt.Sprintf("failed to create jobsWorkflowsTemplate template: %v", err))
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

	outBuf.WriteString(requiresHeader)
	for _, modTemplate := range modTemplates {
		if err := requiresTemplate.Execute(outBuf, modTemplate); err != nil {
			return "", fmt.Errorf("failed to execute requiresTemplate template: %v", err)
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
