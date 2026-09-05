// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !aix && !darwin && !dragonfly && !freebsd && !illumos && !linux && !netbsd && !openbsd && !solaris

// Copyright (c) 2026 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clipackager

import (
	"os"
)

func ensurePkgWorkDirPlatform(path string) error {
	// This preserves the no-symlink/no-non-directory leaf invariant, but does not claim to validate platform ACLs.
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	return validatePkgWorkDirEntry(path, fi)
}
