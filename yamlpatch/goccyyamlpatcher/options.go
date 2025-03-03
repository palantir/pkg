// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goccyyamlpatcher

import (
	"github.com/goccy/go-yaml"
)

type GoccyYAMLOption interface {
	apply(opt *goccyYAMLLib)
}

type goccyYAMLLibOptionFunc func(opt *goccyYAMLLib)

func (f goccyYAMLLibOptionFunc) apply(opt *goccyYAMLLib) {
	f(opt)
}

func GoccyYAMLEncodeOption(encodeOption yaml.EncodeOption) GoccyYAMLOption {
	return goccyYAMLLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.encodeOptions = append(opt.encodeOptions, encodeOption)
	})
}

func GoccyUseNonFlowWhenModifyingEmptyContainer(useNonFlowWhenModifyingEmptyContainer bool) GoccyYAMLOption {
	return goccyYAMLLibOptionFunc(func(opt *goccyYAMLLib) {
		opt.useNonFlowWhenModifyingEmptyContainer = useNonFlowWhenModifyingEmptyContainer
	})
}
