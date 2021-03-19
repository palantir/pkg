// Copyright (c) 2019 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package window

import (
	"sync"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
)

// ItemWithTimestamp is a struct that stores an item and the time it was submitted.
type ItemWithTimestamp struct {
	Time time.Time
	Item interface{}
}

// TimeWindowedStore is a thread-safe struct that stores submitted items
// and supports polling for all items submitted within the last windowSize period.
// When any operation is made, all out-of-date items are pruned out of memory.
type TimeWindowedStore struct {
	items      []ItemWithTimestamp
	itemsMutex sync.Mutex
	windowSize time.Duration
}

// NewTimeWindowedStore creates a new TimeWindowedStore with the provided windowSize.
// windowSize must be a positive value, otherwise returns error.
func NewTimeWindowedStore(windowSize time.Duration) (*TimeWindowedStore, error) {
	if windowSize <= 0 {
		return nil, werror.Error("windowSize must be positive", werror.SafeParam("windowSize", windowSize))
	}
	return &TimeWindowedStore{
		windowSize: windowSize,
	}, nil
}

// GetWindowSize returns the windowSize.
func (t *TimeWindowedStore) GetWindowSize() time.Duration {
	return t.windowSize
}

func (t *TimeWindowedStore) pruneExpiredEntries() {
	currentTime := time.Now()
	newStartIndex := 0
	for _, entry := range t.items {
		if currentTime.Sub(entry.Time) <= t.windowSize {
			break
		}
		newStartIndex++
	}
	t.items = t.items[newStartIndex:]
}

// Submit prunes all out-of-date items out of memory and then adds a new one.
func (t *TimeWindowedStore) Submit(item interface{}) {
	t.itemsMutex.Lock()
	defer t.itemsMutex.Unlock()

	t.pruneExpiredEntries()
	t.items = append(t.items, ItemWithTimestamp{
		Time: time.Now(),
		Item: item,
	})
}

// ItemsInWindow prunes all out-of-date items out of memory and then returns all up-to-date items.
// The returned slice is the one used internally and must not be modified.
func (t *TimeWindowedStore) ItemsInWindow() []ItemWithTimestamp {
	t.itemsMutex.Lock()
	defer t.itemsMutex.Unlock()

	t.pruneExpiredEntries()
	return t.items
}
