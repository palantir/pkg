// Copyright (c) 2018 Palantir Technologies. All rights reserved.
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

package wlog

import (
	"maps"
	"reflect"
)

type MapLogEntry interface {
	LogEntry

	StringValues() map[string]string
	SafeLongValues() map[string]int64
	IntValues() map[string]int32
	StringListValues() map[string][]string
	StringMapValues() map[string]map[string]string
	AnyMapValues() map[string]map[string]interface{}
	ObjectValues() map[string]ObjectValue
	ObjectListValues() map[string][]any

	// Apply applies the values of all of the stored entries of this MapLogEntry to the provided LogEntry.
	Apply(logEntry LogEntry)
	// AllValues returns a single map that contains all of the keys and values stored in this entry.
	AllValues() map[string]interface{}
}

type ObjectValue struct {
	Value         interface{}
	MarshalerType reflect.Type
}

func NewMapLogEntry() MapLogEntry {
	return &mapLogEntry{
		stringValues:     make(map[string]string),
		safeLongValues:   make(map[string]int64),
		intValues:        make(map[string]int32),
		stringListValues: make(map[string][]string),
		stringMapValues:  make(map[string]map[string]string),
		anyMapValues:     make(map[string]map[string]interface{}),
		objectValues:     make(map[string]ObjectValue),
		objectListValues: make(map[string][]any),
	}
}

// mapLogEntry is an in-memory implementation of LogEntry that tracks all of the provided key/value pairs. It only
// stores a single value for a given key -- if multiple calls are made with the same key, only the last value is stored.
// If multiple map values are provided for the same key, the map values are merged.
type mapLogEntry struct {
	stringValues     map[string]string
	safeLongValues   map[string]int64
	intValues        map[string]int32
	stringListValues map[string][]string
	stringMapValues  map[string]map[string]string
	anyMapValues     map[string]map[string]interface{}
	objectValues     map[string]ObjectValue
	objectListValues map[string][]any
}

func (le *mapLogEntry) delete(k string) {
	delete(le.stringValues, k)
	delete(le.safeLongValues, k)
	delete(le.intValues, k)
	delete(le.stringListValues, k)
	delete(le.stringMapValues, k)
	delete(le.anyMapValues, k)
	delete(le.objectValues, k)
	delete(le.objectListValues, k)
}

// mapLogEntrySetKey sets the key in the provided map to be the provided value. The provided map should be a map in the
// provided mapLogEntry. Also ensure that the provided mapLogEntry is updated so that, if the key was previously
// assigned for any other value type, it is unassigned.
func mapLogEntrySetKey[ValT any](le *mapLogEntry, m map[string]ValT, k string, val ValT) {
	le.delete(k)
	m[k] = val
}

func (le *mapLogEntry) StringValues() map[string]string {
	return le.stringValues
}

func (le *mapLogEntry) SafeLongValues() map[string]int64 {
	return le.safeLongValues
}

func (le *mapLogEntry) IntValues() map[string]int32 {
	return le.intValues
}

func (le *mapLogEntry) StringListValues() map[string][]string {
	return le.stringListValues
}

func (le *mapLogEntry) StringMapValues() map[string]map[string]string {
	return le.stringMapValues
}

func (le *mapLogEntry) AnyMapValues() map[string]map[string]interface{} {
	return le.anyMapValues
}

func (le *mapLogEntry) ObjectValues() map[string]ObjectValue {
	return le.objectValues
}

func (le *mapLogEntry) ObjectListValues() map[string][]any {
	return le.objectListValues
}

func (le *mapLogEntry) StringValue(k, v string) {
	mapLogEntrySetKey(le, le.stringValues, k, v)
}

func (le *mapLogEntry) StringListValue(k string, v []string) {
	mapLogEntrySetKey(le, le.stringListValues, k, v)
}

func (le *mapLogEntry) StringListAppendValue(k string, v []string) {
	le.StringListValue(k, append(le.stringListValues[k], v...))
}

func (le *mapLogEntry) OptionalStringValue(k, v string) {
	if v == "" {
		le.delete(k)
	} else {
		le.StringValue(k, v)
	}
}

func (le *mapLogEntry) SafeLongValue(k string, v int64) {
	mapLogEntrySetKey(le, le.safeLongValues, k, v)
}

func (le *mapLogEntry) IntValue(k string, v int32) {
	mapLogEntrySetKey(le, le.intValues, k, v)
}

func (le *mapLogEntry) StringMapValue(k string, v map[string]string) {
	mapLogEntryAddValuesToMap(le, le.stringMapValues, k, v)
}

func (le *mapLogEntry) AnyMapValue(k string, v map[string]interface{}) {
	mapLogEntryAddValuesToMap(le, le.anyMapValues, k, v)
}

func (le *mapLogEntry) ObjectListValue(k string, v []any) {
	mapLogEntrySetKey(le, le.objectListValues, k, v)
}

func (le *mapLogEntry) ObjectListAppendValue(k string, v []any) {
	le.ObjectListValue(k, append(le.objectListValues[k], v...))
}

func mapLogEntryAddValuesToMap[ValT any](m *mapLogEntry, mapValues map[string]map[string]ValT, k string, v map[string]ValT) {
	entryMapVals, ok := mapValues[k]
	if !ok {
		// if entry does not exist, initialize with an empty map
		entryMapVals = make(map[string]ValT)
		mapValues[k] = entryMapVals
	}
	// add all provided elements to map
	maps.Copy(entryMapVals, v)

	// clear key from all maps
	m.delete(k)

	// set key on target map
	mapValues[k] = entryMapVals
}

func (le *mapLogEntry) ObjectValue(k string, v interface{}, marshalerType reflect.Type) {
	mapLogEntrySetKey(le, le.objectValues, k, ObjectValue{
		Value:         v,
		MarshalerType: marshalerType,
	})
}

func (le *mapLogEntry) Apply(logEntry LogEntry) {
	for k, v := range le.stringValues {
		logEntry.StringValue(k, v)
	}
	for k, v := range le.safeLongValues {
		logEntry.SafeLongValue(k, v)
	}
	for k, v := range le.intValues {
		logEntry.IntValue(k, v)
	}
	for k, v := range le.stringListValues {
		logEntry.StringListValue(k, v)
	}
	for k, v := range le.stringMapValues {
		logEntry.StringMapValue(k, v)
	}
	for k, v := range le.anyMapValues {
		logEntry.AnyMapValue(k, v)
	}
	for k, v := range le.objectValues {
		logEntry.ObjectValue(k, v.Value, v.MarshalerType)
	}
	for k, v := range le.objectListValues {
		logEntry.ObjectListValue(k, v)
	}
}

func (le *mapLogEntry) AllValues() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range le.stringValues {
		out[k] = v
	}
	for k, v := range le.safeLongValues {
		out[k] = v
	}
	for k, v := range le.intValues {
		out[k] = v
	}
	for k, v := range le.stringListValues {
		out[k] = v
	}
	for k, v := range le.stringMapValues {
		out[k] = v
	}
	for k, v := range le.anyMapValues {
		out[k] = v
	}
	for k, v := range le.objectValues {
		out[k] = v.Value
	}
	for k, v := range le.objectListValues {
		out[k] = v
	}
	return out
}
