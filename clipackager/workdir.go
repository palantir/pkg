// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"fmt"
	"os"
)

func ensurePkgWorkDir(path string) error {
	if path == "" {
		return fmt.Errorf("package work directory path was empty")
	}
	return ensurePkgWorkDirPlatform(path)
}

func validatePkgWorkDirEntry(path string, fi os.FileInfo) error {
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("package work directory component %s must not be a symlink", path)
	}
	if !fi.IsDir() {
		return fmt.Errorf("package work directory component %s is not a directory", path)
	}
	return nil
}
