// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	stdjson "encoding/json"
	"strconv"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
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

// BenchmarkMarshalSet and BenchmarkUnmarshalSet exercise the set codec's O(n^2)
// deduplication scan, which routes through SetItemCodec.Contains. For a string
// element (a comparableCodec) Contains is slices.Contains with an inlined ==.
func BenchmarkMarshalSet(b *testing.B) {
	data := make([]string, 100)
	for i := range data {
		data[i] = "item-" + strconv.Itoa(i)
	}
	codec := cj.Set[[]string](cj.String[string]())
	runBenchMarshal(b, data, codec)
}

func BenchmarkUnmarshalSet(b *testing.B) {
	src := make([]string, 100)
	for i := range src {
		src[i] = "item-" + strconv.Itoa(i)
	}
	codec := cj.Set[[]string](cj.String[string]())
	data, err := cj.Marshal(src, codec)
	require.NoError(b, err)
	runBenchUnmarshal(b, data, codec)
}

// mapBenchSize is large enough to overflow json/v2's 256-entry string-interning
// cache, so the benchmark reflects arbitrary-key traffic (where each key is
// decoded fresh) rather than the best case where every key stays cached.
const mapBenchSize = 512

func benchMap() map[string]int {
	data := make(map[string]int, mapBenchSize)
	for i := range mapBenchSize {
		data["key-"+strconv.Itoa(i)] = i
	}
	return data
}

func BenchmarkMarshalMap(b *testing.B) {
	codec := cj.Map[map[string]int](cj.String[string](), cj.Int32[int]())
	runBenchMarshal(b, benchMap(), codec)
}

func BenchmarkUnmarshalMap(b *testing.B) {
	codec := cj.Map[map[string]int](cj.String[string](), cj.Int32[int]())
	data, err := cj.Marshal(benchMap(), codec)
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

type benchAddress struct {
	Street string `json:"street"`
	City   string `json:"city"`
	Zip    string `json:"zip"`
}

type benchPerson struct {
	Name       string            `json:"name"`
	Age        int               `json:"age"`
	Emails     []string          `json:"emails"`
	Address    benchAddress      `json:"address"`
	Attributes map[string]string `json:"attributes"`
}

// reflectPerson mirrors benchPerson's wire shape with anonymous struct types so
// it carries no JSON methods: the reflection baselines walk it field by field.
type reflectPerson = struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Emails  []string `json:"emails"`
	Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    string `json:"zip"`
	} `json:"address"`
	Attributes map[string]string `json:"attributes"`
}

func (a benchAddress) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("street")); err != nil {
		return err
	}
	if err := cj.String[string]().MarshalJSONTo(enc, a.Street); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("city")); err != nil {
		return err
	}
	if err := cj.String[string]().MarshalJSONTo(enc, a.City); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("zip")); err != nil {
		return err
	}
	if err := cj.String[string]().MarshalJSONTo(enc, a.Zip); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.EndObject)
}

func (a *benchAddress) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return cj.WrapSyntaxError(dec, err)
	}
	if kind := tok.Kind(); kind != jsontext.KindBeginObject {
		return cj.NewKindMismatchError(dec, kind, "benchAddress opening brace")
	}
	var seenStreet, seenCity, seenZip bool
	var unknownMembers []string
	for {
		key, err := dec.ReadToken()
		if err != nil {
			return cj.WrapSyntaxError(dec, err)
		}
		kind := key.Kind()
		if kind == jsontext.KindEndObject {
			break
		}
		if kind != jsontext.KindString {
			return cj.NewKindMismatchError(dec, kind, "field name")
		}
		switch key.String() {
		case "street":
			if seenStreet {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &a.Street); err != nil {
				return err
			}
			seenStreet = true
		case "city":
			if seenCity {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &a.City); err != nil {
				return err
			}
			seenCity = true
		case "zip":
			if seenZip {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &a.Zip); err != nil {
				return err
			}
			seenZip = true
		default:
			unknownMembers = append(unknownMembers, key.String())
			if err := dec.SkipValue(); err != nil {
				return err
			}
		}
	}
	var missingFields []string
	if !seenStreet {
		missingFields = append(missingFields, "street")
	}
	if !seenCity {
		missingFields = append(missingFields, "city")
	}
	if !seenZip {
		missingFields = append(missingFields, "zip")
	}
	if len(missingFields) > 0 {
		return cj.NewMissingFieldsError(dec, missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := jsonv2.GetOption(dec.Options(), jsonv2.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError(dec, unknownMembers)
		}
	}
	return nil
}

func (p benchPerson) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("name")); err != nil {
		return err
	}
	if err := cj.String[string]().MarshalJSONTo(enc, p.Name); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("age")); err != nil {
		return err
	}
	if err := cj.Int32[int]().MarshalJSONTo(enc, p.Age); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("emails")); err != nil {
		return err
	}
	if err := cj.List[[]string](cj.String[string]()).MarshalJSONTo(enc, p.Emails); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("address")); err != nil {
		return err
	}
	if err := cj.Struct[benchAddress]().MarshalJSONTo(enc, p.Address); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String("attributes")); err != nil {
		return err
	}
	if err := cj.Map[map[string]string](cj.String[string](), cj.String[string]()).MarshalJSONTo(enc, p.Attributes); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.EndObject)
}

func (p *benchPerson) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return cj.WrapSyntaxError(dec, err)
	}
	if kind := tok.Kind(); kind != jsontext.KindBeginObject {
		return cj.NewKindMismatchError(dec, kind, "benchPerson opening brace")
	}
	var seenName, seenAge, seenEmails, seenAddress, seenAttributes bool
	var unknownMembers []string
	for {
		key, err := dec.ReadToken()
		if err != nil {
			return cj.WrapSyntaxError(dec, err)
		}
		kind := key.Kind()
		if kind == jsontext.KindEndObject {
			break
		}
		if kind != jsontext.KindString {
			return cj.NewKindMismatchError(dec, kind, "field name")
		}
		switch key.String() {
		case "name":
			if seenName {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &p.Name); err != nil {
				return err
			}
			seenName = true
		case "age":
			if seenAge {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.Int32[int]().UnmarshalJSONFrom(dec, &p.Age); err != nil {
				return err
			}
			seenAge = true
		case "emails":
			if seenEmails {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.List[[]string](cj.String[string]()).UnmarshalJSONFrom(dec, &p.Emails); err != nil {
				return err
			}
			seenEmails = true
		case "address":
			if seenAddress {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.Struct[benchAddress]().UnmarshalJSONFrom(dec, &p.Address); err != nil {
				return err
			}
			seenAddress = true
		case "attributes":
			if seenAttributes {
				return cj.NewDuplicateFieldKeyError(dec)
			}
			if err := cj.Map[map[string]string](cj.String[string](), cj.String[string]()).UnmarshalJSONFrom(dec, &p.Attributes); err != nil {
				return err
			}
			seenAttributes = true
		default:
			unknownMembers = append(unknownMembers, key.String())
			if err := dec.SkipValue(); err != nil {
				return err
			}
		}
	}
	var missingFields []string
	if !seenName {
		missingFields = append(missingFields, "name")
	}
	if !seenAge {
		missingFields = append(missingFields, "age")
	}
	if !seenEmails {
		missingFields = append(missingFields, "emails")
	}
	if !seenAddress {
		missingFields = append(missingFields, "address")
	}
	if !seenAttributes {
		missingFields = append(missingFields, "attributes")
	}
	if len(missingFields) > 0 {
		return cj.NewMissingFieldsError(dec, missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := jsonv2.GetOption(dec.Options(), jsonv2.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError(dec, unknownMembers)
		}
	}
	return nil
}

func benchPeople() ([]benchPerson, []reflectPerson) {
	const n = 50
	people := make([]benchPerson, n)
	refl := make([]reflectPerson, n)
	for i := range people {
		id := strconv.Itoa(i)
		people[i] = benchPerson{
			Name:    "person-" + id,
			Age:     i,
			Emails:  []string{id + "@a.example", id + "@b.example", id + "@c.example"},
			Address: benchAddress{Street: id + " Main St", City: "City-" + id, Zip: "Z" + id},
			Attributes: map[string]string{
				"role-" + id: "engineer",
				"team-" + id: "platform",
				"tier-" + id: "gold",
			},
		}
		refl[i].Name = people[i].Name
		refl[i].Age = people[i].Age
		refl[i].Emails = people[i].Emails
		refl[i].Address.Street = people[i].Address.Street
		refl[i].Address.City = people[i].Address.City
		refl[i].Address.Zip = people[i].Address.Zip
		refl[i].Attributes = people[i].Attributes
	}
	return people, refl
}

func BenchmarkMarshalCompound(b *testing.B) {
	people, refl := benchPeople()
	codec := cj.List[[]benchPerson](cj.Struct[benchPerson]())
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
			_, _ = cj.Marshal(people, codec)
		}
	})
}

func BenchmarkUnmarshalCompound(b *testing.B) {
	people, _ := benchPeople()
	codec := cj.List[[]benchPerson](cj.Struct[benchPerson]())
	data, err := cj.Marshal(people, codec)
	require.NoError(b, err)
	b.Run("cj", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []benchPerson
			_ = cj.Unmarshal(data, &v, codec)
		}
	})
	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []reflectPerson
			_ = stdjson.Unmarshal(data, &v)
		}
	})
	b.Run("jsonv2", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v []reflectPerson
			_ = jsonv2.Unmarshal(data, &v)
		}
	})
}
