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
	mapped := MapValues(ctx, input, func(ctx context.Context, _ string, value int) Validated[int] {
		v, _, _ := MapWithError(ctx, New(value), func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		return v
	})
	current := mapped.Unvalidated()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)
	_, err := mapped.Validation()
	assert.NoError(t, err)
}

func TestMapValuesAddKey(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1})
	mapped := MapValues(ctx, input, func(ctx context.Context, _ string, value int) Validated[int] {
		v, _, _ := MapWithError(ctx, New(value), func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		return v
	})
	require.Len(t, mapped.Unvalidated(), 1)
	input.Update(map[string]int{"a": 1, "b": 2})
	current := mapped.Unvalidated()
	assert.Equal(t, 2, current["a"])
	assert.Equal(t, 4, current["b"])
	assert.Len(t, current, 2)
}

func TestMapValuesRemoveKey(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1, "b": 2})
	mapped := MapValues(ctx, input, func(ctx context.Context, _ string, value int) Validated[int] {
		v, _, _ := MapWithError(ctx, New(value), func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		return v
	})
	require.Len(t, mapped.Unvalidated(), 2)
	input.Update(map[string]int{"a": 1})
	current := mapped.Unvalidated()
	assert.Equal(t, 2, current["a"])
	_, exists := current["b"]
	assert.False(t, exists)
	assert.Len(t, current, 1)
}

func TestMapValuesValidationError(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("validation failed")
	input := New(map[string]int{"a": 1, "b": 2})
	mapped := MapValues(ctx, input, func(ctx context.Context, key string, value int) Validated[int] {
		v, _, _ := Validate(ctx, New(value*2), func(_ context.Context, _ int) error {
			if key == "b" {
				return testErr
			}
			return nil
		})
		return v
	})
	val, err := mapped.Validation()
	assert.Nil(t, val)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, testErr))
	current := mapped.Unvalidated()
	assert.Equal(t, map[string]int{"a": 2, "b": 0}, current)
}

func TestMapValuesMappedRefreshableUpdates(t *testing.T) {
	ctx := context.Background()
	input := New(map[string]int{"a": 1})
	var underlying Updatable[string]
	mapped := MapValues(ctx, input, func(ctx context.Context, _ string, _ int) Validated[string] {
		underlying = New("b")
		v, _, _ := MapWithError(ctx, underlying, func(_ context.Context, s string) (string, error) {
			return s, nil
		})
		return v
	})
	assert.Equal(t, map[string]string{"a": "b"}, mapped.Unvalidated())
	underlying.Update("c")
	assert.Equal(t, map[string]string{"a": "c"}, mapped.Unvalidated())
}
