// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	stdjson "encoding/json"
	"strconv"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/require"
)

// Benchmark methodology
//
// These benchmarks ask whether the reflection-free cj codecs hold up against the
// reflection-based encoders, and where.
//
//   - The cj side always goes through the production entry points cj.Marshal /
//     cj.Unmarshal. Those reuse json/v2's pooled encoder/decoder, exactly as
//     encoding/json and json/v2 pool theirs. Driving a codec with a freshly
//     allocated jsontext.Encoder per iteration (as earlier benchmarks did)
//     measures encoder construction, not codec work, and is not comparable to a
//     pooled json.Marshal.
//   - "jsonv2" (go-json-experiment/json) is the baseline that matters long term:
//     it is the engine this package is built on and the reflection path generated
//     code falls back to. "stdlib" (encoding/json v1) is included because it is
//     what conjure-go generates today.
//   - String/List/Map use Go types with no custom JSON methods, so their baselines
//     already exercise reflection and share the runBench* helpers. The struct
//     benchmarks are written longhand instead: their baselines must encode/decode
//     reflectStruct, a method-stripped alias of testStruct (defining a new type
//     drops its methods), so they measure pure reflection rather than testStruct's
//     own hand-written methods -- the work cj's generated
//     MarshalJSONTo/UnmarshalJSONFrom replace.
//   - Data is deliberately *distinct* (no repeated strings). json/v2 carries a
//     256-entry string-interning cache on its pooled decoder (see makeString /
//     stringCache); decoding the same string repeatedly lets that cache return
//     interned copies and reports near-zero string allocations. That cache lives
//     on json/v2's internal decoder state and is unreachable through the public
//     jsontext API, so cj cannot use it. Benchmarking with repeated strings would
//     measure that cache, not the codecs; distinct data isolates the work both
//     sides actually do.
//
// What these show, against json/v2: cj is competitive on CPU time -- ahead on
// compound marshal and struct unmarshal, behind on scalar unmarshal -- and
// allocates the same for the fundamental decode work (materializing a stored
// string escapes to the heap either way). json/v2's only structural edge is the
// string cache above, which helps when the same string recurs (field names,
// enum values) and is a wash otherwise. The case for cj is therefore Conjure
// semantics and richer errors at no meaningful throughput cost, not a raw win.

// The codec is a type parameter, not a cj.Codec[T] interface value: cj.Marshal /
// cj.Unmarshal instantiate their internal wrapper with the codec's type, and a
// boxed interface would make that wrapper dispatch on a nil interface (panic).
// Passing the concrete codec type keeps the call monomorphized, as generated
// code does.

func runBenchMarshal[T any, C cj.Codec[T]](b *testing.B, data T, codec C) {
	b.Run("jsonv1", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = stdjson.Marshal(data)
		}
	})
	b.Run("jsonv2", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = jsonv2.Marshal(data, jsonv2.Deterministic(true))
		}
	})
	b.Run("cj", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = cj.Marshal(data, codec)
		}
	})
}

func runBenchUnmarshal[T any, C cj.Codec[T]](b *testing.B, data []byte, codec C) {
	b.Run("cj", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v T
			_ = cj.Unmarshal(data, &v, codec)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v T
			_ = stdjson.Unmarshal(data, &v)
		}
	})
	b.Run("jsonv2", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v T
			_ = jsonv2.Unmarshal(data, &v)
		}
	})
}

func BenchmarkMarshalString(b *testing.B) {
	data := "hello world, this is a test string for benchmarking purposes"
	codec := cj.String[string]()
	runBenchMarshal(b, data, codec)
}

func BenchmarkUnmarshalString(b *testing.B) {
	data := []byte(`"hello world, this is a test string for benchmarking purposes"`)
	codec := cj.String[string]()
	runBenchUnmarshal(b, data, codec)
}

func BenchmarkMarshalList(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = "item-" + strconv.Itoa(i)
	}
	codec := cj.List[[]string](cj.String[string]())
	runBenchMarshal(b, data, codec)
}

func BenchmarkUnmarshalList(b *testing.B) {
	src := make([]string, 100)
	for i := range src {
		src[i] = "item-" + strconv.Itoa(i)
	}
	codec := cj.List[[]string](cj.String[string]())
	data, err := cj.Marshal(src, codec)
	require.NoError(b, err)
	runBenchUnmarshal(b, data, codec)
}

func BenchmarkMarshalMap(b *testing.B) {
	data := make(map[string]int, 50)
	for i := range 50 {
		data[string(rune('a'+i))] = i
	}
	codec := cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]())
	runBenchMarshal(b, data, codec)
}

func BenchmarkUnmarshalMap(b *testing.B) {
	src := make(map[string]int, 50)
	for i := range 50 {
		src[string(rune('a'+i))] = i
	}
	codec := cj.OrderedMap[map[string]int](cj.String[string](), cj.Int32[int]())
	data, err := cj.Marshal(src, codec)
	require.NoError(b, err)
	runBenchUnmarshal(b, data, codec)
}

// reflectStruct is a method-stripped copy of testStruct, so the reflection
// baselines below encode/decode by reflection instead of dispatching to
// testStruct's handwritten JSON methods.
type reflectStruct = struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// BenchmarkMarshalStructList and BenchmarkUnmarshalStructList are the headline
// cases: a list of objects is where handwritten codecs displace the most
// reflection (one struct walk per element).
// Written longhand rather than via runBenchMarshal/runBenchUnmarshal: testStruct
// carries hand-written JSON methods, so the reflection baselines must operate on
// the method-stripped reflectStruct while cj operates on testStruct via its codec.
func BenchmarkMarshalStructList(b *testing.B) {
	const n = 50
	data := make([]testStruct, n)
	refl := make([]reflectStruct, n)
	for i := range data {
		data[i] = testStruct{Name: "person-" + strconv.Itoa(i), Age: i}
		refl[i] = reflectStruct{Name: "person-" + strconv.Itoa(i), Age: i}
	}
	codec := cj.List[[]testStruct](cj.Struct[testStruct]())
	b.Run("jsonv1", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = stdjson.Marshal(refl)
		}
	})
	b.Run("jsonv2", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = jsonv2.Marshal(refl, jsonv2.Deterministic(true))
		}
	})
	b.Run("cj", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = cj.Marshal(data, codec)
		}
	})
}

func BenchmarkUnmarshalStructList(b *testing.B) {
	const n = 50
	src := make([]testStruct, n)
	for i := range src {
		src[i] = testStruct{Name: "person-" + strconv.Itoa(i), Age: i}
	}
	codec := cj.List[[]testStruct](cj.Struct[testStruct]())
	data, err := cj.Marshal(src, codec)
	require.NoError(b, err)
	b.Run("cj", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []testStruct
			_ = cj.Unmarshal(data, &v, codec)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []reflectStruct
			_ = stdjson.Unmarshal(data, &v)
		}
	})
	b.Run("jsonv2", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []reflectStruct
			_ = jsonv2.Unmarshal(data, &v)
		}
	})
}
