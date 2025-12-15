# ASN.1 Struct Tags for Go

This document describes the struct tags feature for the ASN.1 library, which allows you to encode and decode Go structs to/from ASN.1 using struct tags, similar to how `encoding/json` works.

## Overview

Instead of manually creating ASN.1 objects and sequences, you can now define your data structures as Go structs with ASN.1 tags and let the library handle the encoding/decoding automatically.

## Basic Usage

### Simple Struct

```go
type Document struct {
    ID      int64     `asn1:"integer"`
    Title   string    `asn1:"utf8string"`
    Created time.Time `asn1:"utctime"`
    Public  bool      `asn1:"boolean"`
    Content []byte    `asn1:"octetstring"`
}

// Encoding
doc := &Document{
    ID:      12345,
    Title:   "My Document",
    Created: time.Now(),
    Public:  true,
    Content: []byte("document content"),
}

encoded, err := asn1.Marshal(doc)
if err != nil {
    // handle error
}

// Decoding
var decoded Document
err = asn1.Unmarshal(encoded, &decoded)
if err != nil {
    // handle error
}
```

## Supported ASN.1 Types

| Go Type | ASN.1 Tag | ASN.1 Type | Example |
|---------|-----------|------------|---------|
| `bool` | `boolean` | BOOLEAN | `IsActive bool \`asn1:"boolean"\`` |
| `int64`, `int32`, `int` | `integer` | INTEGER | `ID int64 \`asn1:"integer"\`` |
| `uint64`, `uint32`, `uint` | `integer` | INTEGER | `Count uint64 \`asn1:"integer"\`` |
| `string` | `utf8string` | UTF8String | `Name string \`asn1:"utf8string"\`` |
| `string` | `printablestring` | PrintableString | `Code string \`asn1:"printablestring"\`` |
| `string` | `ia5string` | IA5String | `Email string \`asn1:"ia5string"\`` |
| `[]byte` | `octetstring` | OCTET STRING | `Data []byte \`asn1:"octetstring"\`` |
| `time.Time` | `utctime` | UTCTime | `Created time.Time \`asn1:"utctime"\`` |
| `time.Time` | `generalizedtime` | GeneralizedTime | `Expires time.Time \`asn1:"generalizedtime"\`` |
| `struct` | `sequence` | SEQUENCE | `Address Address \`asn1:"sequence"\`` |
| `[]T` | `sequence` | SEQUENCE OF | `Items []Item \`asn1:"sequence"\`` |
| `interface{}` | `choice` | CHOICE | `Content interface{} \`asn1:"choice"\`` |

## CHOICE Types

ASN.1 CHOICE types represent "one of several alternatives" and can be handled in three different ways:

### 1. Interface{} Approach (Recommended for Simple Cases)

```go
type Message struct {
    ID      int64       `asn1:"integer"`
    Content interface{} `asn1:"choice"` // Can hold any supported type
}

// Usage examples
msg1 := &Message{ID: 123, Content: true}           // Boolean choice
msg2 := &Message{ID: 123, Content: int64(42)}      // Integer choice  
msg3 := &Message{ID: 123, Content: "hello"}        // String choice
msg4 := &Message{ID: 123, Content: []byte("data")} // Bytes choice
msg5 := &Message{ID: 123, Content: time.Now()}     // Time choice

// Encoding/Decoding works seamlessly
encoded, err := asn1.Marshal(msg1)
var decoded Message
err = asn1.Unmarshal(encoded, &decoded)
```

### 2. Union Struct Approach (Type-Safe)

```go
type MessageChoice struct {
    BoolValue   *bool      `asn1:"boolean,optional,tag:0"`
    IntValue    *int64     `asn1:"integer,optional,tag:1"`  
    StringValue *string    `asn1:"utf8string,optional,tag:2"`
    BytesValue  *[]byte    `asn1:"octetstring,optional,tag:3"`
    TimeValue   *time.Time `asn1:"utctime,optional,tag:4"`
}

type Message struct {
    ID      int64         `asn1:"integer"`
    Content MessageChoice `asn1:"choice"`
}

// Usage - exactly one field should be non-nil
stringVal := "hello"
msg := &Message{
    ID: 123,
    Content: MessageChoice{StringValue: &stringVal},
}
```

### 3. Manual ASN1Choice Integration

```go
type Message struct {
    ID      int64            `asn1:"integer"`
    Content *asn1.ASN1Choice // Direct use of library's CHOICE type
}

// Usage
choice := asn1.NewChoiceWithID(asn1.NewBoolean(true), "approved")
msg := &Message{ID: 123, Content: choice}
```


## Optional Fields and Context-Specific Tags

Optional fields are represented as pointers and can have context-specific tags:

```go
type Person struct {
    // Required fields
    ID   int64  `asn1:"integer"`
    Name string `asn1:"utf8string"`
    
    // Optional fields with context-specific tags
    Department  *string    `asn1:"printablestring,optional,tag:0"`
    PhoneNumber *string    `asn1:"printablestring,optional,tag:1"`
    Birthday    *time.Time `asn1:"utctime,optional,tag:2"`
    Salary      *int64     `asn1:"integer,optional,tag:3"`
}
```

The generated ASN.1 structure will be:
```
Person ::= SEQUENCE {
    id           INTEGER,
    name         UTF8String,
    department   [0] PrintableString OPTIONAL,
    phoneNumber  [1] PrintableString OPTIONAL,
    birthday     [2] UTCTime OPTIONAL,
    salary       [3] INTEGER OPTIONAL
}
```

## Tag Options

### `optional`
Marks a field as optional. The field must be a pointer type.

```go
Name *string `asn1:"utf8string,optional"`
```

### `tag:N`
Specifies a context-specific tag number for the field. **Uses IMPLICIT tagging by default** (compatible with SS7 MAP/CAP protocols).

```go
Department *string `asn1:"printablestring,optional,tag:0"`  // IMPLICIT tagging (default)
```

### `explicit`
Forces EXPLICIT tagging for context-specific tags. Must be used with `tag:N`.

```go
Manager *Person `asn1:"sequence,optional,tag:5,explicit"`  // EXPLICIT tagging
```

**IMPLICIT vs EXPLICIT Tagging:**
- **IMPLICIT** (default): Replaces the original tag with the context-specific tag. More compact, used in MAP/CAP.
  - Example: `[0]` directly contains the INTEGER value
- **EXPLICIT**: Wraps the original tag with a context-specific tag wrapper. Preserves type information.
  - Example: `[0] { INTEGER value }` - both tags are present

```go
type Example struct {
    ImplicitInt int64  `asn1:"integer,tag:0"`          // [0] contains integer directly
    ExplicitInt int64  `asn1:"integer,tag:1,explicit"` // [1] wraps universal INTEGER
}
```

### `omitempty`
Skip encoding the field if it has a zero value (for non-pointer types).

```go
Description string `asn1:"utf8string,omitempty"`
```

### `-`
Ignore this field completely during marshaling/unmarshaling.

```go
TempData []byte `asn1:"-"`
```

## Complex Examples

### Nested Structures

```go
type Address struct {
    Street   string `asn1:"utf8string"`
    City     string `asn1:"utf8string"`
    PostCode string `asn1:"printablestring"`
    Country  string `asn1:"printablestring"`
}

type Employee struct {
    ID      int64   `asn1:"integer"`
    Name    string  `asn1:"utf8string"`
    Address Address `asn1:"sequence"`
}
```

### Arrays and Slices

```go
type Company struct {
    Name      string     `asn1:"utf8string"`
    Employees []Employee `asn1:"sequence"`  // SEQUENCE OF Employee
    Tags      []string   `asn1:"sequence"`  // SEQUENCE OF UTF8String
}
```

### Recursive Structures

```go
type PersonWithManager struct {
    ID      int64              `asn1:"integer"`
    Name    string             `asn1:"utf8string"`
    Manager *PersonWithManager `asn1:"sequence,optional,tag:0"`
}
```

## Advanced Usage

### Custom Marshal Options

```go
opts := &asn1.MarshalOptions{
    UseContextTags: false, // Disable context-specific tags
}

encoded, err := asn1.MarshalWithOptions(data, opts)
```

### Custom Marshaler/Unmarshaler Interfaces

For types that require custom encoding logic (like TBCD for phone numbers, packed formats, or multi-byte structures), you can implement the `ASN1Marshaler` and `ASN1Unmarshaler` interfaces:

```go
// ASN1Marshaler allows types to provide custom ASN.1 encoding
type ASN1Marshaler interface {
    MarshalASN1() ([]byte, error)
}

// ASN1Unmarshaler allows types to provide custom ASN.1 decoding
type ASN1Unmarshaler interface {
    UnmarshalASN1([]byte) error
}
```

#### Example: TBCD-Encoded Phone Number

Telecom protocols often use TBCD (Telephony Binary Coded Decimal) encoding for phone numbers:

```go
type ISDNAddressString struct {
    Nature        NatureOfAddress
    NumberingPlan NumberingPlan
    Digits        string
}

// MarshalASN1 implements custom TBCD encoding
func (a *ISDNAddressString) MarshalASN1() ([]byte, error) {
    // Encode digits as TBCD (nibble-swapped BCD)
    tbcdDigits, err := encodeTBCD(a.Digits)
    if err != nil {
        return nil, err
    }
    
    // First byte contains nature and numbering plan
    firstByte := (byte(a.Nature) << 4) | byte(a.NumberingPlan)
    
    return append([]byte{firstByte}, tbcdDigits...), nil
}

// UnmarshalASN1 implements custom TBCD decoding
func (a *ISDNAddressString) UnmarshalASN1(data []byte) error {
    if len(data) < 1 {
        return fmt.Errorf("ISDN address too short")
    }
    
    a.Nature = NatureOfAddress((data[0] >> 4) & 0x07)
    a.NumberingPlan = NumberingPlan(data[0] & 0x0F)
    a.Digits, _ = decodeTBCD(data[1:])
    return nil
}

// Now use it seamlessly in structs with tags!
type CallRecord struct {
    ServiceKey        uint32            `asn1:"integer,tag:0"`
    CalledPartyNumber ISDNAddressString `asn1:"octetstring,tag:2"`
    CallingPartyNumber ISDNAddressString `asn1:"octetstring,tag:3"`
}

record := CallRecord{
    ServiceKey: 123,
    CalledPartyNumber: ISDNAddressString{
        Nature:        NatureInternational,
        NumberingPlan: NumberingE164,
        Digits:        "12345678",
    },
    CallingPartyNumber: ISDNAddressString{
        Nature:        NatureInternational,
        NumberingPlan: NumberingE164,
        Digits:        "87654321",
    },
}

// Marshal and unmarshal work automatically with custom encoding
encoded, _ := asn1.Marshal(&record)
var decoded CallRecord
asn1.Unmarshal(encoded, &decoded)
```

#### How Custom Marshalers Work

1. **During Marshaling**: The library checks if a field implements `ASN1Marshaler` before applying type-based encoding. If found, it calls `MarshalASN1()` to get the raw bytes, then wraps them with the appropriate ASN.1 tag specified in the struct tag.

2. **During Unmarshaling**: The library checks if a field implements `ASN1Unmarshaler`. If found, it extracts the raw value bytes (without the tag) and passes them to `UnmarshalASN1()`.

3. **Works with All Features**: Custom marshalers work with optional fields, context-specific tags, explicit tagging, and all other struct tag features.

#### When to Use Custom Marshalers

Use custom marshalers when:
- Your type uses specialized encoding (TBCD, packed formats, bit-level packing)
- Multiple fields need to be encoded together in a specific format
- Your encoding doesn't map cleanly to standard ASN.1 primitives
- You need fine-grained control over the byte representation

For simple structs that map to ASN.1 SEQUENCE, just use regular struct tags without custom marshalers.

## Error Handling

The library provides detailed error messages for common issues:

```go
// Invalid tag format
type Bad struct {
    ID int64 `asn1:"invalid_type"`  // Error: unsupported ASN.1 type
}

// Type mismatch
type Mismatch struct {
    ID string `asn1:"integer"`  // Error: expected integer type for integer
}

// Missing required field
type Required struct {
    ID   int64   `asn1:"integer"`
    Name *string `asn1:"utf8string"`  // Error: required field is nil
}
```


