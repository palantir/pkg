package osutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

// MakeValidRegexPath takes a path and makes it into a valid regex string
func MakeValidRegexPath(path string) string {
	return strings.Replace(filepath.FromSlash(path), "\\", "\\\\", -1)
}

// GetNotADirErrorMsg returns the error message given if an action is
// performed on a non-existant directory.
func GetNotADirErrorMsg() string {
	if runtime.GOOS == "windows" {
		return "The system cannot find the path specified."
	}
	return "not a directory"
}

// GetNoSuchFileOrDirErrorMsg returns the error message given if an action is
// performed on a non-existant file or directory.
func GetNoSuchFileOrDirErrorMsg() string {
	if runtime.GOOS == "windows" {
		return "The system cannot find the file specified."
	}
	return "no such file or directory"
}
