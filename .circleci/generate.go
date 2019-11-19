// +build generate

// This program prints the CircleCI configuration for the "pkg" repository. Standard way to run it is to run
// "go run generate.go".
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
	header = `checkout-path: &checkout-path
  checkout-path: /go/src/github.com/palantir/pkg

version: 2.1

orbs:
  go: palantir/go@0.0.14
  godel: palantir/godel@0.0.14

jobs:
  verify-circleci:
    working_directory: /go/src/github.com/palantir/pkg
    docker:
      - image: "golang:1.13.4"
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

`

	moduleTemplateContent = `      # {{.Module}}
      - godel/verify:
          name: {{.Module}}-verify
          <<: *checkout-path
          include-tests: true
          executor:
            name: go/golang
            version: {{.CurrGoVersion}}
            owner-repo: palantir/pkg/{{.Module}}
      - godel/test:
          name: {{.Module}}-test-go-{{.PrevGoMajorVersion}}
          <<: *checkout-path
          executor:
            name: go/golang
            version: {{.PrevGoVersion}}
            owner-repo: palantir/pkg/{{.Module}}
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
	configYML, err := createConfigYML(mods, "1.13.4", "1.12.13")
	if err != nil {
		panic(err)
	}
	fmt.Print(configYML)
}

var moduleTemplate *template.Template

func init() {
	var err error
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

	outBuf := &bytes.Buffer{}
	_, _ = fmt.Fprint(outBuf, header)
	for i, modDir := range modDirs {
		modJobs, err := moduleJobs(TemplateObject{
			Module:             modDir,
			CurrGoVersion:      currGoVersion,
			PrevGoVersion:      prevGoVersion,
			PrevGoMajorVersion: strings.Join(prevParts[:2], "."),
		})
		if err != nil {
			return "", fmt.Errorf("failed to generate jobs for moduleTemplate %s: %v", modDir, err)
		}
		fmt.Fprint(outBuf, modJobs)
		if i != len(modDirs)-1 {
			fmt.Fprintln(outBuf)
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

func moduleJobs(obj TemplateObject) (string, error) {
	buf := &bytes.Buffer{}
	if err := moduleTemplate.Execute(buf, obj); err != nil {
		return "", fmt.Errorf("failed to execute moduleTemplate temlpate: %v", err)
	}
	return buf.String(), nil
}
