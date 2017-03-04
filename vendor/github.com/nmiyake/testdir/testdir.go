package testdir

import (
	"fmt"
	"io/ioutil"
	"os"
)

// CreateTempDir creates a directory using ioutil.TempDir. If the ioutil.TempDir call is successful, returns the result
// and a function that removes the directory. The returned function is suitable for use in a 'defer' call.
func CreateTempDir(dir, prefix string) (string, func(), error) {
	path, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return "", nil, err
	}
	return path, RemoveAllFunc(path), nil
}

// RemoveAllFunc returns a function that calls "os.RemoveAll" on the specified path and prints any error that is
// encountered. This function is useful to use as the "defer" call to clean up a directory created in a test.
func RemoveAllFunc(path string) func() {
	return func() {
		if err := os.RemoveAll(path); err != nil {
			fmt.Printf("Failed to remove directory %v in defer: %v", path, err)
		}
	}
}
