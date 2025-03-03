// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yamlpatch

type Patcher interface {
	Apply(originalBytes []byte, patch Patch) ([]byte, error)
}
