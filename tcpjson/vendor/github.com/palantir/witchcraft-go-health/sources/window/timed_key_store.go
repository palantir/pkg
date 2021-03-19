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
	"time"
)

// TimedKey is a pair of a key and a timestamp.
type TimedKey struct {
	Key     string
	Time    time.Time
	Payload interface{}
}

// TimedKeys is a list of TimedKey objects.
type TimedKeys []TimedKey

// Keys converts the list of TimedKey objects into a list of keys preserving the order.
func (t TimedKeys) Keys() []string {
	var keys []string
	for _, timedKey := range t {
		keys = append(keys, timedKey.Key)
	}
	return keys
}

// TimedKeyStore is a list of keys ordered by the time they were added or updated.
// Each key is unique within the store. Adding an already present key will cause the time of the key to be updated to
// the current time. The position within the list will be updated accordingly.
type TimedKeyStore interface {
	// Put adds a new TimedKey to the end of the list with the timestamp set to the current time.
	// Adding an already present key will cause the current TimedKey to be updated to the current and to be sent to the end of the list.
	Put(key string, payload interface{})
	// Delete removes a TimedKey from the list. If the key doesn't exist, it is a no op.
	// The second return value returns whether or not the key existed within the store.
	Delete(key string) bool
	// Get returns the TimedKey associated with the provided key if it exists. Returns empty struct otherwise.
	// The second return value returns whether or not the key exists within the store.
	Get(key string) (TimedKey, bool)
	// Get returns a list of all stored TimedKeys in increasing order of timestamps.
	List() TimedKeys
	// Oldest returns the stored TimedKey with the oldest timestamp if it exists. Returns empty struct otherwise.
	// The second return value returns whether or not such element exist.
	Oldest() (TimedKey, bool)
	// Newest returns the stored TimedKey with the newest timestamp if it exists. Returns empty struct otherwise.
	// The second return value returns whether or not such element exist.
	Newest() (TimedKey, bool)
	// PruneOldKeys removes any TimedKey from the list that was added longer than maxAge ago.
	// DEPRECATED: please use PruneKeysAboveAge instead as it uses the internal timeProvider.
	PruneOldKeys(maxAge time.Duration, timeProvider TimeProvider)
	// PruneKeysAboveAge removes any TimedKey from the list that was added longer than maxAge ago.
	PruneKeysAboveAge(maxAge time.Duration)
}

// keyNode is a node in double linked list that holds a TimedKey.
type keyNode struct {
	prev     *keyNode
	next     *keyNode
	timedKey TimedKey
}

// timedKeyStore is an implementation of a TimedKeyStore using a map and a double linked list.
// begin and end are extra nodes that are before the first element and after the last one, respectively.
// They point to each other when the list is empty.
type timedKeyStore struct {
	begin        *keyNode
	end          *keyNode
	nodeByKey    map[string]*keyNode
	timeProvider TimeProvider
}

// NewTimedKeyStore creates a TimedKeyStore that executes all operations in O(1) time except
// for List, which is O(n), where n is the number of stored keys.
// Memory consumption is O(n), where n is the number of stored keys.
// This struct is not thread safe.
func NewTimedKeyStore(timeProvider TimeProvider) TimedKeyStore {
	begin := &keyNode{}
	end := &keyNode{}
	begin.next = end
	end.prev = begin
	return &timedKeyStore{
		begin:        begin,
		end:          end,
		nodeByKey:    make(map[string]*keyNode),
		timeProvider: timeProvider,
	}
}

func (t *timedKeyStore) PruneKeysAboveAge(maxAge time.Duration) {
	t.PruneOldKeys(maxAge, t.timeProvider)
}

func (t *timedKeyStore) PruneOldKeys(maxAge time.Duration, timeProvider TimeProvider) {
	curTime := timeProvider.Now()
	for {
		oldest, exists := t.Oldest()
		if !exists {
			return
		}

		if curTime.Sub(oldest.Time) < maxAge {
			return
		}

		t.Delete(oldest.Key)
	}
}

func (t *timedKeyStore) Put(key string, payload interface{}) {
	_ = t.Delete(key)
	timedKey := TimedKey{
		Key:     key,
		Time:    t.timeProvider.Now(),
		Payload: payload,
	}
	node := &keyNode{
		prev:     t.end.prev,
		next:     t.end,
		timedKey: timedKey,
	}
	node.prev.next = node
	node.next.prev = node
	t.nodeByKey[key] = node
}

func (t *timedKeyStore) Delete(key string) bool {
	node, exists := t.nodeByKey[key]
	if !exists {
		return false
	}
	delete(t.nodeByKey, key)
	node.prev.next = node.next
	node.next.prev = node.prev
	return true
}

func (t *timedKeyStore) Get(key string) (TimedKey, bool) {
	node, exists := t.nodeByKey[key]
	if !exists {
		return TimedKey{}, false
	}
	return node.timedKey, true
}

func (t *timedKeyStore) List() TimedKeys {
	var timedKeys TimedKeys
	for node := t.begin.next; node != t.end; node = node.next {
		timedKeys = append(timedKeys, node.timedKey)
	}
	return timedKeys
}

func (t *timedKeyStore) Oldest() (TimedKey, bool) {
	if t.begin.next == t.end {
		return TimedKey{}, false
	}
	return t.begin.next.timedKey, true
}

func (t *timedKeyStore) Newest() (TimedKey, bool) {
	if t.end.prev == t.begin {
		return TimedKey{}, false
	}
	return t.end.prev.timedKey, true
}
