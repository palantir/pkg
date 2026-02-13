// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"context"
)

type mContextKey string

const (
	registryKey = mContextKey("metrics-registry")
)

var DefaultMetricsRegistry = NewRootMetricsRegistry()

func WithRegistry(ctx context.Context, registry Registry) context.Context {
	if container, ok := ctx.Value(registryKey).(*registryContainer); ok {
		return context.WithValue(ctx, registryKey, &registryContainer{
			Registry: registry,
			Tags:     container.Tags,
		})
	}
	return context.WithValue(ctx, registryKey, &registryContainer{
		Registry: registry,
	})
}

func FromContext(ctx context.Context) Registry {
	prev, ok := ctx.Value(registryKey).(*registryContainer)
	if !ok {
		return DefaultMetricsRegistry
	}
	registry, ok := prev.Registry.(*rootRegistry)
	if !ok {
		return prev.Registry
	}
	if len(prev.Tags) == 0 {
		return registry
	}
	return &childRegistry{
		root: registry,
		tags: prev.Tags,
	}
}

// AddTags adds the provided tags to the provided context. If no tags are provided, the context is returned unchanged.
// Otherwise, a new context is returned with the new tags appended to any tags stored on the parent context.
// This function does not perform any de-duplication (that is, if a tag in the provided tags has the
// same key as an existing one, it will still be appended).
func AddTags(ctx context.Context, tags ...Tag) context.Context {
	if len(tags) == 0 {
		return ctx
	}
	container, ok := ctx.Value(registryKey).(*registryContainer)
	if !ok || container == nil {
		return context.WithValue(ctx, registryKey, &registryContainer{
			Registry: DefaultMetricsRegistry,
			Tags:     container.Tags,
		})
	}
	newTags := make(Tags, len(container.Tags)+len(tags))
	copy(newTags, container.Tags)
	copy(newTags[len(container.Tags):], tags)
	return context.WithValue(ctx, registryKey, &registryContainer{
		Registry: container.Registry,
		Tags:     newTags,
	})
}

// TagsFromContext returns the tags stored on the provided context. May be nil if no tags have been set on the context.
func TagsFromContext(ctx context.Context) Tags {
	if container, ok := ctx.Value(registryKey).(*registryContainer); ok {
		return container.Tags
	}
	return nil
}

type registryContainer struct {
	Registry Registry
	Tags     Tags
}
