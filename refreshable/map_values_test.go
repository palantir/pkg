// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapValues(t *testing.T) {
	ctx := context.Background()

	// Create input map refreshable
	input := New(map[string]int{
		"a": 1,
		"b": 2,
	})

	// Mapper that doubles the value
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{
			validated:   value * 2,
			unvalidated: value * 2,
			lastErr:     nil,
		})
		return v
	})

	// Verify initial values
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)

	_, err := mapped.Validation()
	assert.NoError(t, err)
}

func TestMapValuesAddKey(t *testing.T) {
	ctx := context.Background()

	input := New(map[string]int{
		"a": 1,
	})

	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{
			validated:   value * 2,
			unvalidated: value * 2,
			lastErr:     nil,
		})
		return v
	})

	// Verify initial state
	require.Len(t, mapped.Current(), 1)

	// Add a new key
	input.Update(map[string]int{
		"a": 1,
		"b": 2,
	})

	// Verify new key was added
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)
}

func TestMapValuesRemoveKey(t *testing.T) {
	ctx := context.Background()

	input := New(map[string]int{
		"a": 1,
		"b": 2,
	})

	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{
			validated:   value * 2,
			unvalidated: value * 2,
			lastErr:     nil,
		})
		return v
	})

	// Verify initial state
	require.Len(t, mapped.Current(), 2)

	// Remove a key
	input.Update(map[string]int{
		"a": 1,
	})

	// Verify key was removed
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	_, exists := current["b"]
	assert.False(t, exists)
	assert.Len(t, current, 1)
}

func TestMapValuesValidationError(t *testing.T) {
	ctx := context.Background()

	testErr := errors.New("validation failed")

	input := New(map[string]int{
		"a": 1,
		"b": 2,
	})

	// Mapper that returns error for key "b"
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		if key == "b" {
			v.r.Update(validRefreshableContainer[int]{
				validated:   0,
				unvalidated: value * 2,
				lastErr:     testErr,
			})
		} else {
			v.r.Update(validRefreshableContainer[int]{
				validated:   value * 2,
				unvalidated: value * 2,
				lastErr:     nil,
			})
		}
		return v
	})

	// Current should only contain valid keys
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	_, exists := current["b"]
	assert.False(t, exists)
	assert.Len(t, current, 1)

	// Validation should return the error
	_, err := mapped.Validation()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, testErr))
}

func TestMapValuesMappedRefreshableUpdates(t *testing.T) {
	ctx := context.Background()

	input := New(map[string]int{
		"a": 1,
	})

	// Store references to the created refreshables so we can update them
	createdRefreshables := make(map[string]*validRefreshable[int])

	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{
			validated:   value * 2,
			unvalidated: value * 2,
			lastErr:     nil,
		})
		createdRefreshables[key] = v
		return v
	})

	// Verify initial value
	require.Equal(t, 2, mapped.Current()["a"])

	// Update the mapped refreshable directly
	createdRefreshables["a"].r.Update(validRefreshableContainer[int]{
		validated:   100,
		unvalidated: 100,
		lastErr:     nil,
	})

	// Verify the output was updated
	assert.Equal(t, 100, mapped.Current()["a"])
}
