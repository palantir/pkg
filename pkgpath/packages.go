// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkgpath

import (
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/palantir/pkg/matcher"
)

// DefaultGoPkgExcludeMatcher returns a matcher that matches names that standard Go tools generally exclude as Go
// packages. This includes hidden directories, directories named "testdata" and directories that start with an
// underscore.
func DefaultGoPkgExcludeMatcher() matcher.Matcher {
	return matcher.Any(matcher.Hidden(), matcher.Name("testdata"), matcher.Name("_.+"))
}

type Type int

const (
	NotSupported Type = iota
	Relative
	GoPathRelative
	Absolute
)

// Convert the provided path that is of the provided mode to the path using the mode of the receiver. If the receiver or
// "providedMode" is Relative, "rootDir" will be used as the base directory to resolve the relative paths.
func (t Type) Convert(providedPath string, providedType Type, rootDir string) (string, error) {
	// if modes are identical, no conversion needed
	if t == providedType {
		return providedPath, nil
	}

	// convert provided path to absolute path
	providedPathAbs := ""
	switch providedType {
	case Absolute:
		providedPathAbs = providedPath
	case GoPathRelative:
		providedPathAbs = path.Join(os.Getenv("GOPATH"), "src", providedPath)
	case Relative:
		providedPathAbs = path.Join(rootDir, providedPath)
	default:
		return "", fmt.Errorf("unrecognized source path mode: %v", providedType)
	}

	// convert absolute path to desired path
	switch t {
	case Absolute:
		return providedPathAbs, nil
	case GoPathRelative:
		goSrcDir := path.Join(os.Getenv("GOPATH"), "src")
		return filepath.Rel(goSrcDir, providedPathAbs)
	case Relative:
		relPath, err := filepath.Rel(rootDir, providedPathAbs)
		if err != nil {
			return "", fmt.Errorf("failed to convert %s to relative path against %v: %v", providedPathAbs, rootDir, err)
		}
		return "./" + relPath, nil
	default:
		return "", fmt.Errorf("unrecognized target path mode: %v", t)
	}
}

type packages struct {
	rootDir string
	// key is absolute package path, value is package name
	pkgs map[string]string
}

type Packages interface {
	RootDir() string
	Packages(pathType Type) (map[string]string, error)
	Paths(pathType Type) ([]string, error)
	Filter(exclude matcher.Matcher) (Packages, error)
}

// Filter returns a Packages object that contains all of the packages that do not match the provided matcher.
func (p *packages) Filter(exclude matcher.Matcher) (Packages, error) {
	allPkgsRelPaths, err := p.Packages(Relative)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative paths for packages: %v", err)
	}

	filteredAbsPathPkgs := make(map[string]string)
	for currPkgRelPath, currPkg := range allPkgsRelPaths {
		if exclude == nil || !exclude.Match(currPkgRelPath) {
			filteredAbsPathPkgs[path.Join(p.rootDir, currPkgRelPath)] = currPkg
		}
	}

	return createPkgsWithValidation(p.rootDir, filteredAbsPathPkgs)
}

func (p *packages) RootDir() string {
	return p.rootDir
}

func (p *packages) Packages(pathType Type) (map[string]string, error) {
	pkgs := make(map[string]string, len(p.pkgs))
	for currPath, currPkg := range p.pkgs {
		pkgs[currPath] = currPkg
	}

	switch pathType {
	case Absolute:
		return pkgs, nil
	case GoPathRelative:
		relPathsMap := make(map[string]string, len(pkgs))
		for currAbsPath, currPkg := range pkgs {
			currRelPath, err := GoPathRelative.Convert(currAbsPath, Absolute, path.Join(os.Getenv("GOPATH"), "src"))
			if err != nil {
				return nil, fmt.Errorf("unable to get relative paths: %v", err)
			}
			relPathsMap[currRelPath] = currPkg
		}
		return relPathsMap, nil
	case Relative:
		relPathsMap := make(map[string]string, len(pkgs))
		for currAbsPath, currPkg := range pkgs {
			currRelPath, err := Relative.Convert(currAbsPath, Absolute, p.rootDir)
			if err != nil {
				return nil, fmt.Errorf("unable to get relative paths: %v", err)
			}
			relPathsMap[currRelPath] = currPkg
		}
		return relPathsMap, nil
	default:
		return nil, fmt.Errorf("unrecognized path type: %v", pathType)
	}
}

func (p *packages) Paths(pathType Type) ([]string, error) {
	pkgs, err := p.Packages(pathType)
	if err != nil {
		return nil, err
	}
	pkgPaths := make([]string, 0, len(pkgs))
	for currPath := range pkgs {
		pkgPaths = append(pkgPaths, currPath)
	}
	sort.Strings(pkgPaths)
	return pkgPaths, nil
}

// PackagesFromPaths creates a Packages using the provided relative paths. If any of the relative paths end in a splat
// ("/..."), then all of the sub-directories of that directory are also considered.
func PackagesFromPaths(rootDir string, relPaths []string) (Packages, error) {
	absoluteRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s to absolute path: %v", rootDir, err)
	}

	expandedRelPaths, err := expandPaths(rootDir, relPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to expand paths %v: %v", relPaths, err)
	}

	pkgs := make(map[string]string, len(expandedRelPaths))
	for _, currPath := range expandedRelPaths {
		currAbsPath := path.Join(absoluteRoot, currPath)
		currPkg, err := getPrimaryPkgForDir(currAbsPath, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to determine package for directory %s: %v", currAbsPath, err)
		}
		pkgs[currAbsPath] = currPkg
	}

	return createPkgsWithValidation(absoluteRoot, pkgs)
}

// PackagesInDir creates a Packages that contains all of the packages rooted at the provided directory. Every directory
// rooted in the provided directory whose path does not match the provided exclude matcher is considered as a package.
func PackagesInDir(rootDir string, exclude matcher.Matcher) (Packages, error) {
	dirAbsolutePath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s to absolute path: %v", rootDir, err)
	}

	allPkgs := make(map[string]string)
	if err := filepath.Walk(dirAbsolutePath, func(currPath string, currInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk failed at path %s: %v", currPath, err)
		}

		if !currInfo.IsDir() {
			return nil
		}

		currRelPath, err := filepath.Rel(dirAbsolutePath, currPath)
		if err != nil {
			return fmt.Errorf("failed to resolve %s to relative path against base %s: %v", currPath, dirAbsolutePath, err)
		}

		// if current path matches an include and does not match the exclude, include
		if exclude != nil && exclude.Match(currRelPath) {
			return nil
		}

		// create a filter for processing package files that only passes if it does not match an exclude
		filter := func(info os.FileInfo) bool {
			// if exclude exists and matches the file, skip it
			if exclude != nil && exclude.Match(path.Join(currRelPath, info.Name())) {
				return false
			}
			// process file if it would be included in build context (handles things like build tags)
			match, _ := build.Default.MatchFile(currPath, info.Name())
			return match
		}

		pkgName, err := getPrimaryPkgForDir(currPath, filter)
		if err != nil {
			return fmt.Errorf("unable to determine package for directory %s: %v", currPath, err)
		}

		if pkgName != "" {
			allPkgs[currPath] = pkgName
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return createPkgsWithValidation(dirAbsolutePath, allPkgs)
}

func createPkgsWithValidation(rootDir string, pkgs map[string]string) (*packages, error) {
	if !path.IsAbs(rootDir) {
		return nil, fmt.Errorf("provided rootDir %s is not an absolute path", rootDir)
	}

	for currAbsPkgPath := range pkgs {
		if !path.IsAbs(currAbsPkgPath) {
			return nil, fmt.Errorf("package %s in packages %s is not an absolute path", currAbsPkgPath, currAbsPkgPath)
		}
	}

	return &packages{
		rootDir: rootDir,
		pkgs:    pkgs,
	}, nil
}

func expandPaths(rootDir string, relPaths []string) ([]string, error) {
	var expandedRelPaths []string
	for _, currRelPath := range relPaths {
		if strings.HasSuffix(currRelPath, "/...") {
			// expand splatted paths
			splatBaseDir := currRelPath[:len(currRelPath)-len("/...")]
			baseDirAbsPath := path.Join(rootDir, splatBaseDir)
			err := filepath.Walk(baseDirAbsPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return fmt.Errorf("walk failed at path %s: %v", path, err)
				}
				if info.IsDir() {
					relPath, err := filepath.Rel(rootDir, path)
					if err != nil {
						return fmt.Errorf("failed to resolve %v as a relative path against %s: %v", path, rootDir, err)
					}
					expandedRelPaths = append(expandedRelPaths, relPath)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			expandedRelPaths = append(expandedRelPaths, currRelPath)
		}
	}
	return expandedRelPaths, nil
}

func getPrimaryPkgForDir(dir string, filter func(os.FileInfo) bool) (string, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, filter, parser.PackageClauseOnly)
	if err != nil {
		return "", fmt.Errorf("failed to parse directory %s as a package: %v", dir, err)
	}

	switch len(pkgs) {
	case 0:
		return "", nil
	case 1:
		// if only one entry exists, return its package
		for _, value := range pkgs {
			return value.Name, nil
		}
	default:
		// more than 1 entry exists: filter down to unique packages (if a package ends in "_test", remove suffix)
		uniquePkgs := make(map[string]struct{})
		for _, value := range pkgs {
			uniquePkgs[strings.TrimSuffix(value.Name, "_test")] = struct{}{}
		}

		// if there is only a single package, return it
		if len(uniquePkgs) == 1 {
			for pkg := range uniquePkgs {
				return pkg, nil
			}
		}

		// more than one package exists: return error
		pkgs := make([]string, 0, len(uniquePkgs))
		for pkg := range uniquePkgs {
			pkgs = append(pkgs, pkg)
		}
		sort.Strings(pkgs)
		return "", fmt.Errorf("directory %s contains more than 1 package: %v", dir, pkgs)
	}

	return "", nil
}
