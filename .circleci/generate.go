//go:build generate
// +build generate

// This program prints the CircleCI configuration for the "pkg" repository.
// This program can also update the go module Go directives for all sub-modules contained in the "pkg" repository.
// to the N-1 go version, if the "-updateGoMod" flag is provided.
//
// Example Usages:
//  1. Print the CircleCI config to stdout and overwrite the existing CircleCI config.yml
//      go run generate.go -repoRoot /path/to/github.com/palantir/pkg > config.yml".
//  2. Print the CircleCI config and update all go.mod go directives to the defined prev version
//      go run generate.go -updateGoMod -repoRoot /path/to/github.com/palantir/pkg
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	goVersionLatest       = "1.17.5"
	goVersionPrev         = "1.16.12"
	headerTemplateContent = `checkout-path: &checkout-path
  checkout-path: /go/src/github.com/palantir/pkg

version: 2.1

orbs:
  go: palantir/go@0.0.18
  godel: palantir/godel@0.0.18

jobs:
  verify-circleci:
    working_directory: /go/src/github.com/palantir/pkg
    docker:
      - image: "golang:{{.CurrGoVersion}}"
    resource_class: small
    steps:
      - checkout
      - run: go version
      - run: go run .circleci/generate.go -repoRoot .
      - run: diff  <(cat .circleci/config.yml) <(go run .circleci/generate.go .)
  circle-all:
    docker:
      - image: "golang:{{.CurrGoVersion}}"
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
	updateGoMod := flag.Bool("updateGoMod", false, "update go mod versions for all modules")
	repoRoot := flag.String("repoRoot", "", "Absolute path to the repo root")
	flag.Parse()

	if *repoRoot == "" {
		panic("repo root directory must be provided")
	}

	modParentDir := *repoRoot
	mods, err := modules(modParentDir)
	if err != nil {
		panic(err)
	}
	configYML, err := createConfigYML(mods, goVersionLatest, goVersionPrev)
	if err != nil {
		panic(err)
	}
	if *updateGoMod {
		if err := updateGoModVersions(modParentDir, mods, goVersionPrev); err != nil {
			panic(err)
		}
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
	prevMajor, prevMinor, err := goMajorMinorVersion(prevGoVersion)
	if err != nil {
		return "", err
	}
	prevMajorMinorVersion := strings.Join([]string{prevMajor, prevMinor}, ".")

	jobNames := make([]string, 0, len(modDirs)*2)
	for _, modDir := range modDirs {
		jobNames = append(jobNames, modDir+"-verify", modDir+"-test-go-"+prevMajorMinorVersion)
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
			PrevGoMajorVersion: prevMajorMinorVersion,
		}); err != nil {
			return "", fmt.Errorf("failed to execute moduleTemplate template: %v", err)
		}
	}
	return outBuf.String(), nil
}

// updateGoModVersions updates the go directive defined in all modules to the provided Go version.
func updateGoModVersions(parentDir string, modDirs []string, goVersion string) error {
	for _, modDir := range modDirs {
		if err := updateGoModVersion(filepath.Join(parentDir, modDir), goVersion); err != nil {
			return err
		}
	}
	return nil
}

// updateGoModVersion runs "go mod edit -go=<major>.<minor>" for the given module to update the module version.
// For example, given a goVersion of "1.16.12" and a go directive of "go 1.15" defined in the go.mod file of the
// given directory, the go directive will be updated to be "go 1.16".
func updateGoModVersion(moduleDir, goVersion string) error {
	major, minor, err := goMajorMinorVersion(goVersion)
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "mod", "edit", fmt.Sprintf("-go=%s.%s", major, minor))
	cmd.Dir = moduleDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running command returned error: err (%v), output: (%s)", err, string(output))
	}
	return nil
}

// goMajorMinorVersion splits the provided go version on all "." and returns the major and minor version.
// For example, given the input "1.16.12", the return values would be "1" and "16".
// An error will be returned if the provided goVersion does not contain a major minor version separated by a "."
func goMajorMinorVersion(goVersion string) (string, string, error) {
	parts := strings.Split(goVersion, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("goVersion must have at least 2 parts separated by a period: %s", goVersion)
	}
	return parts[0], parts[1], nil
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
