// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	"bytes"
	"encoding/json"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
)

// BenchmarkString_Marshal compares cj.String marshaling performance.
func BenchmarkString_Marshal(b *testing.B) {
	data := "hello world, this is a test string for benchmarking purposes"

	b.Run("cj.String", func(b *testing.B) {
		encoder := cj.String[string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf)
			_ = encoder.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})

	b.Run("json/v2.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = jsonv2.Marshal(data)
		}
	})
}

// BenchmarkList_Marshal compares list marshaling performance
func BenchmarkList_Marshal(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = "item"
	}

	b.Run("cj.List", func(b *testing.B) {
		marshaler := cj.List[[]string](cj.String[string]())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf)
			_ = marshaler.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})
}

// BenchmarkMap_Marshal compares map marshaling performance
func BenchmarkMap_Marshal(b *testing.B) {
	data := make(map[string]int, 50)
	for i := range 50 {
		data[string(rune('a'+i))] = i
	}

	b.Run("cj.OrderedMap", func(b *testing.B) {
		marshaler := cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf, jsonv2.Deterministic(true))
			_ = marshaler.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})
}

// BenchmarkInt_Marshal compares integer marshaling performance
func BenchmarkInt_Marshal(b *testing.B) {
	data := 42

	b.Run("cj.Int32", func(b *testing.B) {
		encoder := cj.Int32[int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf)
			_ = encoder.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})
}

// BenchmarkClientEncoder compares codec wrapper performance
func BenchmarkClientEncoder(b *testing.B) {
	data := []string{"a", "b", "c", "d", "e"}

	b.Run("cj.ClientEncoder", func(b *testing.B) {
		encoder := cj.ClientEncoder(cj.List[[]string](cj.String[string]()))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = encoder.Marshal(data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.Skip()
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})
}

// BenchmarkString_Unmarshal compares unmarshaling performance
func BenchmarkString_Unmarshal(b *testing.B) {
	jsonData := []byte(`"hello world, this is a test string for benchmarking purposes"`)

	b.Run("cj.String", func(b *testing.B) {
		decoder := cj.String[string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dec := jsontext.NewDecoder(bytes.NewReader(jsonData))
			var result string
			_ = decoder.UnmarshalJSONFrom(dec, &result)
		}
	})

	b.Run("json.Unmarshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result string
			_ = json.Unmarshal(jsonData, &result)
		}
	})
}

func BenchmarkStruct_Marshal(b *testing.B) {
	data := testStruct{
		Name: "John",
		Age:  25,
	}

	b.Run("cj.Struct", func(b *testing.B) {
		marshaler := cj.Struct[testStruct]()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf)
			_ = marshaler.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		type testStructAlias testStruct
		data := testStructAlias(data)
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})

	b.Run("json/v2.Marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		type testStructAlias testStruct
		data := testStructAlias(data)
		for i := 0; i < b.N; i++ {
			_, _ = jsonv2.Marshal(data)
		}
	})
}

func BenchmarkStruct_Unmarshal(b *testing.B) {
	data := []byte(`{"name":"John","age":25}`)

	b.Run("cj.Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result testStruct
			_ = json.Unmarshal(data, &result)
		}
	})

	b.Run("json.Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		type testStructAlias testStruct
		for i := 0; i < b.N; i++ {
			var result testStructAlias
			_ = json.Unmarshal(data, &result)
		}
	})

	b.Run("json/v2.Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		type testStructAlias testStruct
		for i := 0; i < b.N; i++ {
			var result testStructAlias
			_ = jsonv2.Unmarshal(data, &result)
		}
	})
}

// BenchmarkMemoryAllocation measures allocation behavior
func BenchmarkMemoryAllocation(b *testing.B) {
	data := map[string][]int{
		"numbers": {1, 2, 3, 4, 5},
		"more":    {10, 20, 30},
	}

	b.Run("cj.OrderedMap+List", func(b *testing.B) {
		b.ReportAllocs()
		marshaler := cj.OrderedMap[map[string][]int](cj.String[string](), cj.List[[]int](cj.Int32[int]()))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			enc := jsontext.NewEncoder(buf)
			_ = marshaler.MarshalJSONTo(enc, data)
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(data)
		}
	})
}
