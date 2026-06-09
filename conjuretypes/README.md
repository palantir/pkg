# `cj`: Conjure JSON

Package `cj` provides JSON codecs for Conjure wire types. It is intended for
generated `conjure-go` code, where the generated type already knows the exact
Conjure shape and can select a concrete codec chain at compile time.

The package is built on `github.com/go-json-experiment/json` and
`github.com/go-json-experiment/json/jsontext`. Codecs implement methods
compatible with the JSON v2 `MarshalerTo` and `UnmarshalerFrom` style APIs, but
generated code normally calls codec methods directly instead of relying on
reflective dispatch for every nested value.

See the [Conjure Wire Specification's JSON format](https://github.com/palantir/conjure/blob/master/docs/spec/wire.md#5-json-format)
for the wire-level rules.

## Core Interfaces

`Codec[T]` is the unit of composition:

```go
type Codec[T any] interface {
    MarshalJSONTo(enc *jsontext.Encoder, receiver T) error
    UnmarshalJSONFrom(dec *jsontext.Decoder, receiver *T) error
    Equal(T, T) bool
}
```

`MapKeyCodec[K]` adds in-place key sorting so map encoding is deterministic;
`SetItemCodec[T]` adds a membership test used to drop duplicate set elements.
Each keeps its extra method off the base `Codec` so a single map or set codec
serves every element type at full speed (see the Map Ordering and Performance
sections):

```go
type MapKeyCodec[K comparable] interface {
    Codec[K]
    Sort(keys []K)
}

type SetItemCodec[T any] interface {
    Codec[T]
    Contains(set []T, item T) bool
}
```

Public codec constructors return unexported, zero-size implementation values.
Constructor arguments on container codecs are type witnesses for generic
inference only; they are not stored. Codec implementations should remain
stateless.

## Provided Constructors

Primitive and scalar codecs:

- `String[T ~string]()`
- `Int32[T ~int | ~int8 | ~int16 | ~int32 | ~int64]()`
- `Int32MapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]()`
- `SafeLong[T ~int | ~int8 | ~int16 | ~int32 | ~int64]()`
- `SafeLongMapKey[T ~int | ~int8 | ~int16 | ~int32 | ~int64]()`
- `Float[T ~float64]()`
- `FloatMapKey[T ~float64]()`
- `Boolean[T ~bool]()`
- `BooleanMapKey[T ~bool]()`
- `BearerToken[T ~string]()`
- `DateTime[T time.Time | datetime.DateTime]()`
- `UUID[T ~[16]byte]()`
- `RID[T ridConstraint]()`
- `Binary[T ~[]byte]()`
- `BinaryMapKey[T binary.Binary]()`
- `Any[T any]()`

Container and delegation codec call shapes:

- `Optional[*T](itemCodec)`
- `List[[]T](itemCodec)`
- `Set[[]T](itemCodec)`
- `Map[map[K]V](keyCodec, valueCodec)`
- `Struct[T]()` — `T` must implement `MarshalJSONTo`, `UnmarshalJSONFrom`, and `Equal`
- `Text[T encoding.TextMarshaler]()`

Map-key codecs for non-string scalar wire types encode and decode quoted object
names. For example, `Int32MapKey` reads `"42"` for a Go integer key, while
`Int32` reads `42` for a Go integer value.

## Top-Level Helpers

Use `Marshal` and `Unmarshal` when the top-level value needs a Conjure codec:

```go
values := []string{"a", "b"}
codec := cj.List[[]string](cj.String[string]())

data, err := cj.Marshal(values, codec)
if err != nil {
    return err
}

var decoded []string
err = cj.Unmarshal(data, &decoded, codec)
```

`MarshalWrite` and `UnmarshalRead` provide the same behavior for streams.
`NewAnonymousType(receiver, codec)` exposes the underlying anonymous wrapper
when integrating directly with `json.Marshal`, `json.Unmarshal`, or APIs that
expect JSON v2 method implementations.

`ClientEncoder`, `ServerEncoder`, `ClientDecoder`, and `ServerDecoder` are thin
runtime adapters. Server decoders add `json.RejectUnknownMembers(true)`.
Client decoders allow unknown members for forward compatibility.

```go
encoder := cj.ClientEncoder(cj.List[[]int](cj.Int32[int]()))

payload, err := encoder.Marshal([]int{1, 2, 3})
```

## Generated Codec Chains

Generated code chooses concrete codec chains. A `[]string` value uses:

```go
cj.List[[]string](cj.String[string]())
```

A `map[string]*CustomObject` value uses a nested chain:

```go
cj.Map[map[string]*CustomObject](
    cj.String[string](),
    cj.Optional[*CustomObject](cj.Struct[CustomObject]()),
)
```

The outer type argument usually remains explicit. The element, key, value, and
nested codec types are inferred from constructor arguments.

## Performance

The package is designed so codec selection happens at compile time rather than
through per-value reflection. Each constructor returns a zero-size type, and a
chain such as `List[[]string](String[string]())` is a fully monomorphized
generic type: every nested encode/decode is a direct method call (frequently
inlined and devirtualized), not a reflective walk over `reflect.Type` and
`reflect.Value`. This is the structural difference from `encoding/json` and from
the reflection fallback in `json/v2`.

Three properties follow from that design:

- **No reflection per value.** The standard encoders re-walk type information for
  every nested value; `cj` replaces that with generated, direct calls. The
  advantage grows with structural depth, so lists, maps, and objects benefit
  most. A lone scalar has little reflection to remove.
- **Shared pooling.** `Marshal` and `Unmarshal` run through `json/v2`'s pooled
  encoder and decoder, exactly as `encoding/json` and `json/v2` pool theirs, so
  there is no per-call encoder construction.
- **Allocation discipline.** Container decoders decode in place to avoid forcing
  values onto the heap: `List` and `Set` grow their backing array and decode each
  element into it, and the map decoder reuses one key/value pair across entries
  (reset between entries so a reference-typed value cannot alias the previous
  one). The remaining allocations are unavoidable — a decoded string the caller
  stores escapes to the heap regardless of encoder.

The one structural advantage upstream `json/v2` retains is an internal
256-entry string-interning cache on its pooled decoder. When the same string
recurs (repeated field names, enum values, a small key space) it returns
interned copies and reports near-zero string allocations. That cache lives on
`json/v2`'s private decoder state and is unreachable through the public
`jsontext` API, so `cj` cannot use it; for distinct, arbitrary strings — the
common case for map keys and free-form values — the cache misses and both sides
allocate equally.

Net: on compound and nested types `cj` matches or beats `json/v2` on both time
and allocations, and is roughly twice as fast as `encoding/json`; on a single
scalar value `json/v2`'s pooled fast paths and string cache can edge it out. The
benchmarks in `benchmark_test.go` cover scalars, lists, maps, and lists of
objects against both packages. The package's aim is correct Conjure semantics
and richer errors at competitive or better throughput, not a blanket speedup.

## Struct Implementations

Generated structs must implement `MarshalJSONTo`, `UnmarshalJSONFrom`, and
`Equal` directly; `Struct[T]()` will not compile for a type missing any of the
three (route such types through `Any` instead). Field codecs are called directly.

`Equal` carries the struct's own Conjure-value semantics rather than Go's `==` or
`reflect.DeepEqual`: it compares scalar fields directly, slices and maps element
by element, and nested objects through their own `Equal`. The generator emits it
from the same field walk it uses for marshaling.

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (p Person) Equal(other Person) bool {
    return p.Name == other.Name && p.Age == other.Age
}

func (p Person) MarshalJSON() ([]byte, error) {
    return cj.Marshal(p, cj.Struct[Person]())
}

func (p *Person) UnmarshalJSON(data []byte) error {
    return cj.Unmarshal(data, p, cj.Struct[Person]())
}

func (p Person) MarshalJSONTo(enc *jsontext.Encoder) error {
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
    if err := enc.WriteToken(jsontext.EndObject); err != nil {
        return err
    }
    return nil
}

func (p *Person) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
    tok, err := dec.ReadToken()
    if err != nil {
        return cj.WrapSyntaxError(dec, err)
    }
    if kind := tok.Kind(); kind != jsontext.KindBeginObject {
        return cj.NewKindError[Person](dec, kind, "Person opening brace")
    }

    var seenName, seenAge bool
    var unknownMembers []string
    for {
        key, err := dec.ReadToken()
        if err != nil {
            return cj.WrapSyntaxError(dec, err)
        }
        switch kind := key.Kind(); kind {
        case jsontext.KindString:
            switch key.String() {
            case "name":
                if seenName {
                    return cj.NewDuplicateFieldKeyError[Person](dec)
                }
                if err := cj.String[string]().UnmarshalJSONFrom(dec, &p.Name); err != nil {
                    return err
                }
                seenName = true
            case "age":
                if seenAge {
                    return cj.NewDuplicateFieldKeyError[Person](dec)
                }
                if err := cj.Int32[int]().UnmarshalJSONFrom(dec, &p.Age); err != nil {
                    return err
                }
                seenAge = true
            default:
                unknownMembers = append(unknownMembers, key.String())
                if err := dec.SkipValue(); err != nil {
                    return cj.WrapSyntaxError(dec, err)
                }
            }
        case jsontext.KindEndObject:
            var missingFields []string
            if !seenName {
                missingFields = append(missingFields, "name")
            }
            if !seenAge {
                missingFields = append(missingFields, "age")
            }
            if len(missingFields) > 0 {
                return cj.NewMissingFieldsError[Person](dec, missingFields)
            }
            if len(unknownMembers) > 0 {
                if strict, _ := json.GetOption(dec.Options(), json.RejectUnknownMembers); strict {
                    return cj.NewUnknownFieldsError[Person](dec, unknownMembers)
                }
            }
            return nil
        default:
            return cj.NewKindError[Person](dec, kind, "field name")
        }
    }
}
```

Generated struct decoders are responsible for required-field tracking, duplicate
field detection, and unknown-field handling. `ServerDecoder` enables unknown
field rejection through `json.RejectUnknownMembers(true)`; client decoding leaves
unknown fields skipped.

## Error Handling

`cj` uses the JSON v2 error types instead of defining its own error hierarchy.

- Syntax and parser failures are reported as `*jsontext.SyntacticError`.
- Type mismatches, invalid Conjure values, duplicate decoded keys, missing
  fields, and unknown fields are reported as `*json.SemanticError`.
- Semantic errors include JSON pointer, byte offset, and JSON kind when that
  context is available from the direct codec path. The offending JSON *value* is
  never recorded: the pointer and kind locate the fault without copying
  potentially sensitive content (tokens, PII) into the error message or logs.
- Category sentinels support `errors.Is`: `ErrMissingFields`,
  `ErrUnknownFields`, `ErrUnknownEnum`, `ErrDuplicateField`,
  `ErrDuplicateMapKey`, and `ErrDuplicateSetItem`.

Example:

```go
var semantic *json.SemanticError
if errors.As(err, &semantic) {
    pointer := semantic.JSONPointer
    kind := semantic.JSONKind
    _ = pointer
    _ = kind
}

if errors.Is(err, cj.ErrDuplicateMapKey) {
    // The decoded Go map key appeared more than once.
}
```

When `cj.Unmarshal`, `UnmarshalRead`, `ClientDecoder`, or `ServerDecoder`
invoke a top-level codec, they allow duplicate JSON object names at the JSON
parser layer. This lets generated struct and map codecs enforce Conjure
semantics after map-key normalization. For example, an integer-keyed map can
reject both `"01"` and `"1"` as duplicate decoded key `1`.

## Strict and Lenient Decoding

Client decoders are lenient about unknown object members. Server decoders pass
`json.RejectUnknownMembers(true)` so generated struct decoders reject unknown
members after collecting them.

Duplicate object-name handling is separate from unknown-member handling. The
package enables duplicate names while decoding through its top-level helpers so
that generated code can produce Conjure-specific duplicate field or duplicate
map-key errors.

## Nil and Null Containers

By default:

- Nil slices encode as `[]`.
- Nil maps encode as `{}`.
- Optional nil pointers encode as `null`.

The JSON v2 options `json.FormatNilSliceAsNull(true)` and
`json.FormatNilMapAsNull(true)` can be passed to top-level helpers to encode nil
slices and maps as `null`.

On decode, `List`, `Set`, and map codecs consume JSON `null` and leave the Go
container initialized but empty. `Optional` consumes `null` and sets the pointer
to nil.

## Map Ordering

Map encoding is deterministic by default. `Map` sorts keys via the key codec's
`Sort` method, so each key type sorts as fast as it can: `cmp.Ordered` key types
(integer, safelong, double, string) use Go's built-in ordering, while key types
that do not support `<` (uuid, datetime, boolean) sort by a type-specific
comparison.

```go
cj.Map[map[string]Value](cj.String[string](), ValueCodec())
cj.Map[map[CustomKey]Value](CustomKeyCodec(), ValueCodec())
```

Pass `json.Deterministic(false)` to allow the encoder to use Go's map iteration
order.

## Conjure-Specific Values

`Float` encodes IEEE 754 special values as JSON strings:

- `NaN` as `"NaN"`
- positive infinity as `"Infinity"`
- negative infinity as `"-Infinity"`

`FloatMapKey` also uses these string forms, but rejects `NaN` during decoding
because `NaN` is not a usable Go map key for duplicate detection.

`Int32` decoders enforce the 32-bit signed integer range. `SafeLong` decoders
enforce the JavaScript safe integer range, from `-(2^53 - 1)` through
`2^53 - 1`.

`BearerToken` decoders reject empty tokens, non-ASCII input, and characters
outside the bearer-token grammar. Marshalers are intentionally more permissive
than unmarshalers and generally write the provided Go value.

---

## TODO

- use AppendText for types that support it (and add it to those that don't)

---

## Appendix: Ideas That Sound Good but Measure Worse

Recorded so they are not re-tried. Numbers are approximate; the reasoning is the
durable part.

- **`WriteValue`/`AppendQuote` for scalar string codecs**, to avoid the
  `[]byte`-to-`string` conversion. No win for `uuid`/`rid`/`datetime`: their bytes
  come from `String()`/`MarshalText()`, which allocate internally anyway, so the
  extra `WriteValue` scan only adds cost (uuid: 59 ns vs 46 ns, same 1 alloc).
  `rid` is far worse because its `MarshalText` also runs `validate()`. This wins
  only with an append-style formatter (`AppendText`/`base64.AppendEncode`) writing
  into the caller's buffer, as `binaryCodec` does.

- **`WithMarshalers`/`MarshalToFunc` at the entry points** instead of the
  `anonymousMarshaler` wrapper. The wrapper hits json/v2's native `MarshalerTo`
  fast path; the registry approach rebuilds a `*Marshalers` + closure + options
  every call and dispatches through reflection (marshal: 273 ns / 11 allocs vs
  112 ns / 2 allocs). `WithMarshalers` is still correct for the non-codec fallback
  paths (`json.Number`/`json.RawMessage`).

- **`reflect.DeepEqual` (or an auto-detected `==`) for struct equality.**
  `DeepEqual` boxes both args and compares Go representation, not Conjure value
  (datetime `*Location`, nil vs empty `[]byte` diverge); auto-detecting `==` is too
  permissive (a pointer field compares by identity, not pointee). The generator
  knows the fields, so it emits a field-wise `Equal` the codec delegates to
  (`DeepEqual` measured 2 allocs / ~43 ns, ~8× the devirtualized `Equal`).

- **Driving set dedup / map sort with a closure or comparator from the generic
  container**, instead of `Contains`/`Sort` methods on the codec. A type-parameter
  method call *inside a closure* handed to `slices.ContainsFunc`/`SortFunc`
  dispatches through the generic dictionary (indirect, no inlining); the same call
  as a direct method on the codec devirtualizes and inlines to `==`/`slices.Sort`.
  Set dedup of 100 strings: 6.6 µs (direct `Contains`) vs 10.4 µs (`Equal` closure),
  ~37%. This is why `Sort`/`Contains` live on `MapKeyCodec`/`SetItemCodec`, and why
  a `mapCodec[K comparable]` with no `Sort` could not avoid the slower comparator
  path for ordered keys.

- **Routing cj's string codecs through json/v2 to borrow its interned string
  cache.** v2's only structural allocation edge is that cache (recurring field
  names, enum values, bounded keys); for distinct or long (>256 B) strings both
  allocate identically (109 vs 110). Dispatching cj's strings through v2 to reach
  it backfired: same allocs and ~40% slower (3798 ns vs 2701 ns) from per-string
  reflection dispatch. cj stays at allocation parity by construction and wins on
  wall-clock even where the cache gives v2 fewer allocs.
