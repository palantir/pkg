// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"github.com/goccy/go-yaml"
)

type YAMLOption interface {
	apply(opt *goccyYAMLLib)
}

type yamlLibOptionFunc func(opt *goccyYAMLLib)

func (f yamlLibOptionFunc) apply(opt *goccyYAMLLib) {
	f(opt)
}

func YAMLEncodeOption(encodeOption yaml.EncodeOption) YAMLOption {
	return yamlLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.encodeOptions = append(opt.encodeOptions, encodeOption)
	})
}

func UseNonFlowWhenModifyingEmptyContainer(useNonFlowWhenModifyingEmptyContainer bool) YAMLOption {
	return yamlLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.useNonFlowWhenModifyingEmptyContainer = useNonFlowWhenModifyingEmptyContainer
	})
}
