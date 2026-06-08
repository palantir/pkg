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

`MapKeyCodec[K]` is for comparable key types that need a custom deterministic
ordering:

```go
type MapKeyCodec[K comparable] interface {
    Codec[K]
    Compare(K, K) int
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
- `OrderedMap[map[K]V](keyCodec, valueCodec)`
- `ComparableMap[map[K]V](keyCodec, valueCodec)`
- `Struct[T json.MarshalerTo]()`
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
cj.OrderedMap[map[string]*CustomObject](
    cj.String[string](),
    cj.Optional[*CustomObject](cj.Struct[CustomObject]()),
)
```

The outer type argument usually remains explicit. The element, key, value, and
nested codec types are inferred from constructor arguments.

## Struct Implementations

Generated structs should implement `MarshalJSONTo` and `UnmarshalJSONFrom`
directly. Field codecs are called directly.

```go
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
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
    return enc.WriteToken(jsontext.EndObject)
}

func (p *Person) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
    tok, err := dec.ReadToken()
    if err != nil {
        return cj.WrapSyntaxError(dec, err)
    }
    if kind := tok.Kind(); kind != jsontext.KindBeginObject {
        return cj.NewKindMismatchError(dec, kind, "Person opening brace")
    }

    var seenName, seenAge bool
    var unknownMembers []string
    for {
        key, err := dec.ReadToken()
        if err != nil {
            return cj.WrapSyntaxError(dec, err)
        }
        switch kind := key.Kind(); kind {
        case jsontext.KindEndObject:
            var missingFields []string
            if !seenName {
                missingFields = append(missingFields, "name")
            }
            if !seenAge {
                missingFields = append(missingFields, "age")
            }
            if len(missingFields) > 0 {
                return cj.NewMissingFieldsError(dec, "Person", missingFields)
            }
            if len(unknownMembers) > 0 {
                if strict, _ := json.GetOption(dec.Options(), json.RejectUnknownMembers); strict {
                    return cj.NewUnknownFieldsError(dec, "Person", unknownMembers)
                }
            }
            return nil
        case jsontext.KindString:
            // Continue below.
        default:
            return cj.NewKindMismatchError(dec, kind, "field name")
        }

        switch key.String() {
        case "name":
            if seenName {
                return cj.NewDuplicateFieldKeyError(dec, `Person["name"]`)
            }
            if err := cj.String[string]().UnmarshalJSONFrom(dec, &p.Name); err != nil {
                return err
            }
            seenName = true
        case "age":
            if seenAge {
                return cj.NewDuplicateFieldKeyError(dec, `Person["age"]`)
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
- Semantic errors include JSON pointer, byte offset, JSON kind, and JSON value
  when that context is available from the direct codec path.
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

Map encoding is deterministic by default.

`OrderedMap` sorts keys with Go's built-in ordering for `cmp.Ordered` key types:

```go
cj.OrderedMap[map[string]Value](cj.String[string](), ValueCodec())
```

`ComparableMap` sorts keys with `KEY.Compare` for comparable types that do not
support `<`:

```go
cj.ComparableMap[map[CustomKey]Value](CustomKeyCodec(), ValueCodec())
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
