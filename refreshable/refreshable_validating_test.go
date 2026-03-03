// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatingRefreshable(t *testing.T) {
	ctx := context.Background()
	type container struct{ Value string }
	r := refreshable.New(container{Value: "value"})
	vr, _, err := refreshable.Validate[container](ctx, r, func(ctx context.Context, i container) error {
		if len(i.Value) == 0 {
			return errors.New("empty")
		}
		return nil
	})
	require.NoError(t, err)
	v, err := vr.Validation()
	require.NoError(t, err)
	require.Equal(t, "value", v.Value)
	require.Equal(t, "value", r.Current().Value)
	require.Equal(t, "value", vr.LastCurrent().Value)

	// attempt bad update
	r.Update(container{})
	require.Equal(t, r.Current().Value, "")
	v, err = vr.Validation()
	require.EqualError(t, err, "empty", "expected validation error")
	require.Equal(t, "", v.Value, "expected invalid value from Validation")
	require.Equal(t, vr.LastCurrent().Value, "value", "expected unchanged validating refreshable")

	// attempt good update
	r.Update(container{Value: "value2"})
	v, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, "value2", v.Value)
	require.Equal(t, "value2", vr.LastCurrent().Value)
	require.Equal(t, "value2", r.Current().Value)
}

func TestMapValidatingRefreshable(t *testing.T) {
	ctx := context.Background()
	parsed, err := url.Parse("https://palantir.com:443")
	require.NoError(t, err)
	r := refreshable.New("https://palantir.com:443")
	vr, _, err := refreshable.MapWithError[string, *url.URL](ctx, r, func(_ context.Context, s string) (*url.URL, error) { return url.Parse(s) })
	require.NoError(t, err)
	val, err := vr.Validation()
	require.NoError(t, err)
	validatedHost, _, _ := refreshable.MapValidated(ctx, vr, func(ctx context.Context, u *url.URL) (string, error) {
		return u.Hostname(), nil
	})
	require.Equal(t, r.Current(), "https://palantir.com:443")
	require.Equal(t, val, parsed)
	require.Equal(t, vr.LastCurrent().Hostname(), "palantir.com")
	require.Equal(t, validatedHost.LastCurrent(), "palantir.com")

	// attempt bad update
	r.Update(":::error.com")
	assert.Equal(t, r.Current(), ":::error.com")
	val, err = vr.Validation()
	assert.Nil(t, val)
	require.EqualError(t, err, "parse \":::error.com\": missing protocol scheme", "expected err from validating refreshable")
	assert.Equal(t, vr.LastCurrent().Hostname(), "palantir.com", "expected unchanged validating refreshable")
	require.Equal(t, validatedHost.LastCurrent(), "palantir.com")
	_, err = validatedHost.Validation()
	assert.Error(t, err)

	// attempt good update
	r.Update("https://example.com")
	_, err = vr.Validation()
	require.NoError(t, err)
	require.Equal(t, r.Current(), "https://example.com")
	require.Equal(t, vr.LastCurrent().Hostname(), "example.com")
}

func TestMapValidated(t *testing.T) {
	ctx := context.Background()
	r := refreshable.New(10)
	vr, _, err := refreshable.Validate[int](ctx, r, func(_ context.Context, i int) error {
		if i < 0 {
			return errors.New("negative")
		}
		return nil
	})
	require.NoError(t, err)
	doubled, stop, err := refreshable.MapValidated(ctx, vr, func(_ context.Context, i int) (int, error) {
		return i * 2, nil
	})
	defer stop()
	require.NoError(t, err)
	require.Equal(t, 20, doubled.LastCurrent())
	val, err := doubled.Validation()
	require.NoError(t, err)
	require.Equal(t, 20, val)
	// Parent validation error propagates
	r.Update(-1)
	require.Equal(t, 20, doubled.LastCurrent(), "should retain last valid value")
	_, err = doubled.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "negative")
	// Recovery after valid update
	r.Update(5)
	require.Equal(t, 10, doubled.LastCurrent())
	val, err = doubled.Validation()
	require.NoError(t, err)
	require.Equal(t, 10, val)
}

func TestMapValidated_OwnError(t *testing.T) {
	ctx := context.Background()
	r := refreshable.New(10)
	vr, vrStop, err := refreshable.Validate(ctx, r, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vrStop()
	mapped, stop, err := refreshable.MapValidated(ctx, vr, func(_ context.Context, i int) (string, error) {
		if i > 100 {
			return "", errors.New("too large")
		}
		return "ok", nil
	})
	defer stop()
	require.NoError(t, err)
	require.Equal(t, "ok", mapped.LastCurrent())
	r.Update(200)
	require.Equal(t, "ok", mapped.LastCurrent(), "should retain last valid value")
	_, err = mapped.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "too large")
	r.Update(50)
	require.Equal(t, "ok", mapped.LastCurrent())
	_, err = mapped.Validation()
	require.NoError(t, err)
}

func TestMergeValidated(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New("hello")
	r2 := refreshable.New(2)
	vr1, _, err := refreshable.MapWithError(ctx, r1, func(_ context.Context, s string) (string, error) {
		if s == "" {
			return "", errors.New("empty string")
		}
		return s, nil
	})
	require.NoError(t, err)
	vr2, _, err := refreshable.MapWithError(ctx, r2, func(_ context.Context, i int) (int, error) {
		if i < 0 {
			return 0, errors.New("negative")
		}
		return i, nil
	})
	require.NoError(t, err)
	type merged struct {
		s string
		i int
	}
	m, stop := refreshable.MergeValidated(vr1, vr2, func(s string, i int) merged {
		return merged{s: s, i: i}
	})
	defer stop()
	require.Equal(t, merged{s: "hello", i: 2}, m.LastCurrent())
	_, err = m.Validation()
	require.NoError(t, err)
	// Error in first source
	r1.Update("")
	require.Equal(t, merged{s: "hello", i: 2}, m.LastCurrent(), "should retain last valid value")
	_, err = m.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty string")
	// Recovery
	r1.Update("world")
	require.Equal(t, merged{s: "world", i: 2}, m.LastCurrent())
	_, err = m.Validation()
	require.NoError(t, err)
	// Error in second source
	r2.Update(-1)
	require.Equal(t, merged{s: "world", i: 2}, m.LastCurrent(), "should retain last valid value")
	_, err = m.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "negative")
	// Recovery of second
	r2.Update(1)
	require.Equal(t, merged{s: "world", i: 1}, m.LastCurrent())
	_, err = m.Validation()
	require.NoError(t, err)
}

func TestMergeValidated_BothErrors(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New("")
	r2 := refreshable.New(-1)
	vr1, _, _ := refreshable.MapWithError(ctx, r1, func(_ context.Context, s string) (string, error) {
		if s == "" {
			return "", errors.New("empty")
		}
		return s, nil
	})
	vr2, _, _ := refreshable.MapWithError(ctx, r2, func(_ context.Context, i int) (int, error) {
		if i < 0 {
			return 0, errors.New("negative")
		}
		return i, nil
	})
	merged, stop := refreshable.MergeValidated(vr1, vr2, func(s string, i int) string {
		return s
	})
	defer stop()
	_, err := merged.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
	require.Contains(t, err.Error(), "negative")
}

func TestMergeValidated_Subscribe(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, vr1Stop, err := refreshable.Validate(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr1Stop()
	vr2, vr2Stop, err := refreshable.Validate(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr2Stop()
	merged, stop := refreshable.MergeValidated(vr1, vr2, func(a, b int) int {
		return a + b
	})
	defer stop()
	var received []int
	merged.SubscribeValidated(func(val int, err error) {
		received = append(received, val)
	})
	// Initial subscription callback
	require.Len(t, received, 1)
	r1.Update(10)
	require.Len(t, received, 2)
	require.Equal(t, 12, received[1])
}

func TestCollectValidated(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New("a")
	r2 := refreshable.New("b")
	r3 := refreshable.New("c")
	vr1, _, err := refreshable.MapWithError(ctx, r1, func(_ context.Context, s string) (string, error) {
		if s == "" {
			return "", errors.New("empty")
		}
		return s, nil
	})
	require.NoError(t, err)
	vr2, _, err := refreshable.MapWithError(ctx, r2, func(_ context.Context, s string) (string, error) {
		return s, nil
	})
	require.NoError(t, err)
	vr3, _, err := refreshable.MapWithError(ctx, r3, func(_ context.Context, s string) (string, error) {
		return s, nil
	})
	require.NoError(t, err)
	collected, stop := refreshable.CollectValidated(vr1, vr2, vr3)
	defer stop()
	require.Equal(t, []string{"a", "b", "c"}, collected.LastCurrent())
	_, err = collected.Validation()
	require.NoError(t, err)
	// Update one element
	r2.Update("B")
	require.Equal(t, []string{"a", "B", "c"}, collected.LastCurrent())
	_, err = collected.Validation()
	require.NoError(t, err)
	// Error in one element retains last valid slice
	r1.Update("")
	require.Equal(t, []string{"a", "B", "c"}, collected.LastCurrent(), "should retain last valid slice")
	_, err = collected.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
	// Recovery
	r1.Update("A")
	require.Equal(t, []string{"A", "B", "c"}, collected.LastCurrent())
	_, err = collected.Validation()
	require.NoError(t, err)
}

func TestCollectValidatedMutable(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, vr1Stop, err := refreshable.Validate(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr1Stop()
	vr2, vr2Stop, err := refreshable.Validate(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr2Stop()
	collected, add, stop := refreshable.CollectValidatedMutable(vr1, vr2)
	defer stop()
	require.Equal(t, []int{1, 2}, collected.LastCurrent())
	// Add a new element
	r3 := refreshable.New(3)
	vr3, vr3Stop, err := refreshable.Validate(ctx, r3, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr3Stop()
	add(vr3)
	require.Equal(t, []int{1, 2, 3}, collected.LastCurrent())
	// Update propagates
	r1.Update(10)
	require.Equal(t, []int{10, 2, 3}, collected.LastCurrent())
	// Update added element
	r3.Update(30)
	require.Equal(t, []int{10, 2, 30}, collected.LastCurrent())
}

func TestCollectValidatedMutable_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, _, err := refreshable.MapWithError(ctx, r1, func(_ context.Context, i int) (int, error) {
		if i < 0 {
			return 0, errors.New("negative")
		}
		return i, nil
	})
	require.NoError(t, err)
	vr2, vr2Stop, err := refreshable.Validate(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr2Stop()
	collected, _, stop := refreshable.CollectValidatedMutable(vr1, vr2)
	defer stop()
	require.Equal(t, []int{1, 2}, collected.LastCurrent())
	_, err = collected.Validation()
	require.NoError(t, err)
	r1.Update(-1)
	require.Equal(t, []int{1, 2}, collected.LastCurrent(), "should retain last valid slice")
	_, err = collected.Validation()
	require.Error(t, err)
	require.Contains(t, err.Error(), "negative")
	r1.Update(5)
	require.Equal(t, []int{5, 2}, collected.LastCurrent())
	_, err = collected.Validation()
	require.NoError(t, err)
}

func TestCollectValidatedMutable_RaceCondition(t *testing.T) {
	ctx := context.Background()
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	vr1, vr1Stop, err := refreshable.Validate(ctx, r1, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr1Stop()
	vr2, vr2Stop, err := refreshable.Validate(ctx, r2, func(_ context.Context, _ int) error { return nil })
	require.NoError(t, err)
	defer vr2Stop()
	collected, add, stop := refreshable.CollectValidatedMutable(vr1, vr2)
	defer stop()
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			v, _, _ := refreshable.Validate(ctx, refreshable.New(val), func(_ context.Context, _ int) error { return nil })
			add(v)
		}(i + 100)
	}
	for i := range 10 {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			r1.Update(val)
			r2.Update(val * 2)
		}(i)
	}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = collected.LastCurrent()
		}()
	}
	wg.Wait()
	assert.Eventually(t, func() bool {
		return len(collected.LastCurrent()) == 12
	}, time.Second, time.Millisecond)
}

// TestValidatingRefreshable_SubscriptionRaceCondition tests that the ValidatingRefreshable stays current
// if the underlying refreshable updates during the creation process.
func TestValidatingRefreshable_SubscriptionRaceCondition(t *testing.T) {
	ctx := context.Background()
	//r := &updateImmediatelyRefreshable{r: refreshable.New(1), newValue: 2}
	r := refreshable.New(1)
	var seen1, seen2 bool
	vr, _, err := refreshable.Validate[int](ctx, r, func(_ context.Context, i int) error {
		go r.Update(2)
		switch i {
		case 1:
			seen1 = true
		case 2:
			seen2 = true
		}
		return nil
	})
	require.NoError(t, err)
	// If this returns 1, it is likely because the VR contains a stale value
	assert.Eventually(t, func() bool {
		return vr.LastCurrent() == 2
	}, time.Second, time.Millisecond)

	assert.True(t, seen1, "expected to process 1 value")
	assert.True(t, seen2, "expected to process 2 value")
}
