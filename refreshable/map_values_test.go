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
	input := New(map[string]int{"a": 1, "b": 2})
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{validated: value * 2, unvalidated: value * 2, lastErr: nil})
		return v
	})
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)
	_, err := mapped.Validation()
	assert.NoError(t, err)
}

func TestMapValuesAddKey(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1})
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{validated: value * 2, unvalidated: value * 2, lastErr: nil})
		return v
	})
	require.Len(t, mapped.Current(), 1)
	input.Update(map[string]int{"a": 1, "b": 2})
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)
}

func TestMapValuesRemoveKey(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1, "b": 2})
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		v.r.Update(validRefreshableContainer[int]{validated: value * 2, unvalidated: value * 2, lastErr: nil})
		return v
	})
	require.Len(t, mapped.Current(), 2)
	input.Update(map[string]int{"a": 1})
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	_, exists := current["b"]
	assert.False(t, exists)
	assert.Len(t, current, 1)
}

func TestMapValuesValidationError(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("validation failed")
	input := New(map[string]int{"a": 1, "b": 2})
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[int] {
		v := newValidRefreshable[int]()
		if key == "b" {
			v.r.Update(validRefreshableContainer[int]{validated: 0, unvalidated: value * 2, lastErr: testErr})
		} else {
			v.r.Update(validRefreshableContainer[int]{validated: value * 2, unvalidated: value * 2, lastErr: nil})
		}
		return v
	})
	current := mapped.Current()
	assert.Equal(t, 2, current["a"])
	_, exists := current["b"]
	assert.False(t, exists)
	assert.Len(t, current, 1)
	_, err := mapped.Validation()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, testErr))
}

func TestMapValuesMappedRefreshableUpdates(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1})
	var refreshToOutline *validRefreshable[string]
	mapped := MapValues(ctx, input, func(_ context.Context, key string, value int) Validated[string] {
		refreshToOutline = &validRefreshable[string]{
			r: New[validRefreshableContainer[string]](validRefreshableContainer[string]{
				validated:   "b",
				unvalidated: "b",
			}),
		}
		return refreshToOutline
	})
	assert.Equal(t, map[string]string{"a": "b"}, mapped.Current())
	updateValidRefreshable[string](refreshToOutline, func() (string, error) {
		return "c", nil
	})
	assert.Equal(t, map[string]string{"a": "c"}, mapped.Current())
	// require.Equal(t, 2, mapped.Current()["a"])
	// assert.Equal(t, 100, mapped.Current()["a"])
}
