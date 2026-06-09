// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cj_test

import (
	stdjson "encoding/json"
	"maps"
	"slices"
	"strconv"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/palantir/pkg/conjuretypes/cj"
	"github.com/stretchr/testify/require"
)

// Benchmark methodology
//
//   - cj always goes through cj.Marshal / cj.Unmarshal, which reuse json/v2's
//     pooled encoder/decoder just as encoding/json and json/v2 do. Driving a
//     codec with a freshly allocated jsontext.Encoder per iteration would measure
//     encoder construction, not codec work, and would not be comparable.
//   - "jsonv2" is the long-term baseline (the engine this package is built on and
//     the reflection fallback for generated code); "stdlib" (encoding/json) is
//     what conjure-go generates today.
//   - Data is deliberately distinct (no repeated strings). json/v2's pooled
//     decoder carries a 256-entry string-interning cache (makeString /
//     stringCache) that returns interned copies for repeated strings and reports
//     near-zero string allocations; it is unreachable through the public jsontext
//     API, so cj cannot use it. Distinct data isolates codec work from that cache,
//     which only helps when a string recurs (field names, enum values).

// The codec is a type parameter, not a boxed cj.Codec[T]: cj.Marshal / cj.Unmarshal
// instantiate their internal wrapper with the codec's type, and a nil interface
// would make that wrapper dispatch on nil (panic). The concrete type also keeps the
// call monomorphized, as generated code does.

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

// The Set benchmarks exercise the codec's O(n^2) dedup scan via
// SetItemCodec.Contains, which for a comparable element is slices.Contains.
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

// mapBenchSize overflows json/v2's 256-entry string-interning cache, so keys are
// decoded fresh rather than served from cache.
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

// reflectStruct is a method-stripped copy of testStruct so the reflection
// baselines encode/decode by reflection rather than dispatching to testStruct's
// handwritten JSON methods.
type reflectStruct = struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// The StructList benchmarks are the headline cases: a list of objects is where
// handwritten codecs displace the most reflection (one struct walk per element).
// They are written longhand rather than via runBench* because the baselines must
// use reflectStruct while cj uses testStruct's codec.
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

func (a benchAddress) Equal(other benchAddress) bool {
	return a.Street == other.Street && a.City == other.City && a.Zip == other.Zip
}

type benchPerson struct {
	Name       string            `json:"name"`
	Age        int               `json:"age"`
	Emails     []string          `json:"emails"`
	Address    benchAddress      `json:"address"`
	Attributes map[string]string `json:"attributes"`
}

func (p benchPerson) Equal(other benchPerson) bool {
	return p.Name == other.Name &&
		p.Age == other.Age &&
		slices.Equal(p.Emails, other.Emails) &&
		p.Address.Equal(other.Address) &&
		maps.Equal(p.Attributes, other.Attributes)
}

// reflectPerson mirrors benchPerson's wire shape with no JSON methods, so the
// reflection baselines walk it field by field.
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
		return cj.NewKindError[benchAddress](dec, kind, "benchAddress opening brace")
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
			return cj.NewKindError[benchAddress](dec, kind, "field name")
		}
		switch key.String() {
		case "street":
			if seenStreet {
				return cj.NewDuplicateFieldKeyError[benchAddress](dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &a.Street); err != nil {
				return err
			}
			seenStreet = true
		case "city":
			if seenCity {
				return cj.NewDuplicateFieldKeyError[benchAddress](dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &a.City); err != nil {
				return err
			}
			seenCity = true
		case "zip":
			if seenZip {
				return cj.NewDuplicateFieldKeyError[benchAddress](dec)
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
		return cj.NewMissingFieldsError[benchAddress](dec, missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := jsonv2.GetOption(dec.Options(), jsonv2.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError[benchAddress](dec, unknownMembers)
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
		return cj.NewKindError[benchPerson](dec, kind, "benchPerson opening brace")
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
			return cj.NewKindError[benchPerson](dec, kind, "field name")
		}
		switch key.String() {
		case "name":
			if seenName {
				return cj.NewDuplicateFieldKeyError[benchPerson](dec)
			}
			if err := cj.String[string]().UnmarshalJSONFrom(dec, &p.Name); err != nil {
				return err
			}
			seenName = true
		case "age":
			if seenAge {
				return cj.NewDuplicateFieldKeyError[benchPerson](dec)
			}
			if err := cj.Int32[int]().UnmarshalJSONFrom(dec, &p.Age); err != nil {
				return err
			}
			seenAge = true
		case "emails":
			if seenEmails {
				return cj.NewDuplicateFieldKeyError[benchPerson](dec)
			}
			if err := cj.List[[]string](cj.String[string]()).UnmarshalJSONFrom(dec, &p.Emails); err != nil {
				return err
			}
			seenEmails = true
		case "address":
			if seenAddress {
				return cj.NewDuplicateFieldKeyError[benchPerson](dec)
			}
			if err := cj.Struct[benchAddress]().UnmarshalJSONFrom(dec, &p.Address); err != nil {
				return err
			}
			seenAddress = true
		case "attributes":
			if seenAttributes {
				return cj.NewDuplicateFieldKeyError[benchPerson](dec)
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
		return cj.NewMissingFieldsError[benchPerson](dec, missingFields)
	}
	if len(unknownMembers) > 0 {
		if strict, _ := jsonv2.GetOption(dec.Options(), jsonv2.RejectUnknownMembers); strict {
			return cj.NewUnknownFieldsError[benchPerson](dec, unknownMembers)
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
