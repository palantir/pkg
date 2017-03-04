package testdir

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

// GoFileSpec represents the specification for a Go file.
type GoFileSpec struct {
	// The relative path to which the file should be written. For example, "foo/foo.go".
	RelPath string
	// Content of the file.
	Src string
}

// GoFile represents a Go file that has been written to disk.
type GoFile struct {
	// The absolute path to the Go file.
	Path string
	// The import path for the Go file. For example, "github.com/nmiyake/testdir".
	ImportPath string
}

// WriteGoFiles writes the files represented by the specifications in 'files' using the provided directory as the root
// directory. Returns a map of the written files where the key is the 'RelPath' field of the specification that was
// written and the value is the GoFile that was written for the specification.
func WriteGoFiles(dir string, files []GoFileSpec) (map[string]GoFile, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	importsMap := make(map[string]string, len(files))
	for _, currFile := range files {
		fullDirPath := path.Dir(path.Join(dir, currFile.RelPath))
		importPath, err := filepath.Rel(path.Join(os.Getenv("GOPATH"), "src"), fullDirPath)
		if err != nil {
			return nil, err
		}
		importsMap[currFile.RelPath] = importPath
	}

	goFiles := make(map[string]GoFile, len(files))
	for _, currFile := range files {
		filePath := path.Join(dir, currFile.RelPath)
		buf := &bytes.Buffer{}
		t := template.Must(template.New(filePath).Parse(currFile.Src))
		if err := t.Execute(buf, importsMap); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			return nil, err
		}
		goFiles[currFile.RelPath] = GoFile{
			Path:       filePath,
			ImportPath: importsMap[currFile.RelPath],
		}
	}

	return goFiles, nil
}
