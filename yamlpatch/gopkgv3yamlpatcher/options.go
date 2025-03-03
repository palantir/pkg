// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gopkgv3yamlpatcher

type YAMLOption interface {
	apply(opt *goyamlYAMLLib)
}

type yamlLibOptionFunc func(opt *goyamlYAMLLib)

func (f yamlLibOptionFunc) apply(opt *goyamlYAMLLib) {
	f(opt)
}

func IndentSpaces(indentSpaces int) YAMLOption {
	return yamlLibOptionFunc(func(opt *goyamlYAMLLib) {
		opt.indentSpaces = indentSpaces
	})
}
