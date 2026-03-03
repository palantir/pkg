// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRefreshable(t *testing.T) {
	type container struct {
		Value string
	}

	v := &container{Value: "original"}
	r := refreshable.New(v)
	require.Equal(t, r.Current(), v)

	t.Run("Update", func(t *testing.T) {
		v2 := &container{Value: "updated"}
		r.Update(v2)
		require.Equal(t, r.Current(), v2)
	})

	t.Run("Subscribe", func(t *testing.T) {
		var v1, v2 container
		unsub1 := r.Subscribe(func(i *container) {
			v1 = *i
		})
		_ = r.Subscribe(func(i *container) {
			v2 = *i
		})
		require.Equal(t, v1.Value, "updated")
		require.Equal(t, v2.Value, "updated")
		r.Update(&container{Value: "value"})
		require.Equal(t, v1.Value, "value")
		require.Equal(t, v2.Value, "value")

		unsub1()
		r.Update(&container{Value: "value2"})
		require.Equal(t, v1.Value, "value", "should be unchanged after unsubscribing")
		require.Equal(t, v2.Value, "value2", "should be updated after unsubscribing other")
	})

	t.Run("Map", func(t *testing.T) {
		r.Update(&container{Value: "value"})
		rLen, stop := refreshable.Map[*container, int](r, func(i *container) int {
			return len(i.Value)
		})
		defer stop()
		require.Equal(t, 5, rLen.Current())

		rLenUpdates := 0
		rLen.Subscribe(func(int) { rLenUpdates++ })
		require.Equal(t, 1, rLenUpdates)
		// update to new value with same length and ensure the
		// equality check prevented unnecessary subscriber updates.
		r.Update(&container{Value: "VALUE"})
		require.Equal(t, "VALUE", r.Current().Value)
		require.Equal(t, 1, rLenUpdates)

		r.Update(&container{Value: "updated"})
		require.Equal(t, "updated", r.Current().Value)
		require.Equal(t, 7, rLen.Current())
		require.Equal(t, 2, rLenUpdates)
	})
}

func TestCollectMutable_RaceCondition(t *testing.T) {
	r1 := refreshable.New(1)
	r2 := refreshable.New(2)
	collected, add, stop := refreshable.CollectMutable(r1, r2)
	defer stop()

	var wg sync.WaitGroup
	// Concurrent adds
	for i := range 10 {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			add(refreshable.New(val))
		}(i + 100)
	}
	// Concurrent updates to initial refreshables
	for i := range 10 {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			r1.Update(val)
			r2.Update(val * 2)
		}(i)
	}
	// Concurrent reads
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = collected.Current()
		}()
	}
	wg.Wait()

	// Verify we have all 12 elements (2 initial + 10 added)
	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		assert.Len(t, collected.Current(), 12)
	}, 5*time.Second, 10*time.Millisecond)
}

func TestNewFileRefreshable_MapJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tmpdir := t.TempDir()
	ticker := make(chan time.Time, 1)
	ticker <- time.Now()
	filename := filepath.Join(tmpdir, "file.json")
	fileRefreshable := refreshable.NewFileRefreshableWithTicker(ctx, filename, ticker)

	jsonRefreshable, _ := refreshable.Map2(fileRefreshable, func(in []byte, err error) (map[string]string, error) {
		if err != nil {
			return nil, err
		}
		var out map[string]string
		err = json.Unmarshal(in, &out)
		return out, err
	})

	currFile, err := fileRefreshable.Current()
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Empty(t, currFile)

	currJSON, err := jsonRefreshable.Current()
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
	require.Empty(t, currJSON)

	inMap := map[string]string{"foo": "bar"}
	inRefreshable := refreshable.New(inMap)
	inRefreshableBytes, _ := refreshable.MapTo2(inRefreshable, func(in map[string]string) ([]byte, error) { return json.Marshal(in) })
	inWriteErr, _ := refreshable.MapFrom2(inRefreshableBytes, func(bytes []byte, err error) error {
		if err != nil {
			return err
		}
		return os.WriteFile(filename, bytes, 0644)
	})
	require.NoError(t, inWriteErr.Current())

	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		currFile, err := fileRefreshable.Current()
		require.NoError(t, err)
		require.Equal(t, `{"foo":"bar"}`, string(currFile))

		currJSON, err := jsonRefreshable.Current()
		require.NoError(t, err)
		require.Equal(t, inMap, currJSON)
	}, time.Second, time.Millisecond)

	inRefreshable.Update(map[string]string{"baz": "qux"})
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		currFile, err := fileRefreshable.Current()
		require.NoError(t, err)
		require.Equal(t, `{"baz":"qux"}`, string(currFile))

		currJSON, err := jsonRefreshable.Current()
		require.NoError(t, err)
		require.Equal(t, map[string]string{"baz": "qux"}, currJSON)
	}, time.Second, time.Millisecond)

	// Write bad JSON to file
	require.NoError(t, os.WriteFile(filename, []byte(`"not a map"`), 0644))
	ticker <- time.Now()
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		currJSON, err := jsonRefreshable.Current()
		require.Error(t, err)
		require.Empty(t, currJSON)
	}, time.Second, time.Millisecond)
}
