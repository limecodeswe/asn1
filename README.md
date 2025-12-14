# ASN.1 Library for Go

A complete Go library for ASN.1 BER encoding/decoding with struct tag support.

## Features

- **Complete ASN.1 Support**: All universal types (BOOLEAN, INTEGER, SEQUENCE, etc.)
- **Struct Tags**: Marshal/unmarshal like `encoding/json`
- **Context-Specific Tags**: Optional fields with custom tags
- **Type Safety**: Idiomatic Go interfaces
- **Production Ready**: Comprehensive testing and examples

## Installation

```bash
go get github.com/limecodeswe/asn1
```

## Quick Start

### Using Struct Tags (Recommended)

```go
type Document struct {
    ID      int64     `asn1:"integer"`
    Title   string    `asn1:"utf8string"`
    Created time.Time `asn1:"utctime"`
    Public  bool      `asn1:"boolean"`
    
    // Optional fields with context-specific tags
    Author *string `asn1:"utf8string,optional,tag:0"`
}

// Encode
doc := &Document{ID: 123, Title: "My Doc", Created: time.Now(), Public: true}
encoded, err := asn1.Marshal(doc)

// Decode
var decoded Document
err = asn1.Unmarshal(encoded, &decoded)
```

### Manual API

```go
// Create a sequence
person := asn1.NewSequence()
person.Add(asn1.NewUTF8String("John Doe"))
person.Add(asn1.NewInteger(30))
person.Add(asn1.NewBoolean(true))

// Encode
encoded, err := person.Encode()

// Decode
decoded, _, err := asn1.DecodeTLV(encoded)
```

## Struct Tag Options

| Tag | Description | Example |
|-----|-------------|---------|
| `boolean` | BOOLEAN | `bool \`asn1:"boolean"\`` |
| `integer` | INTEGER | `int64 \`asn1:"integer"\`` |
| `utf8string` | UTF8String | `string \`asn1:"utf8string"\`` |
| `octetstring` | OCTET STRING | `[]byte \`asn1:"octetstring"\`` |
| `sequence` | SEQUENCE | `struct \`asn1:"sequence"\`` |
| `choice` | CHOICE | `interface{} \`asn1:"choice"\`` |
| `optional,tag:N` | Context tag (IMPLICIT) | `*string \`asn1:"utf8string,optional,tag:0"\`` |
| `explicit` | Use EXPLICIT tagging | `*struct \`asn1:"sequence,tag:0,explicit"\`` |

## CHOICE Types

ASN.1 CHOICE types can be handled in several ways:

### 1. Interface{} Approach (Simple)
```go
type Message struct {
    ID      int64       `asn1:"integer"`
    Content interface{} `asn1:"choice"` // Can hold bool, int64, string, []byte, time.Time
}

msg := &Message{ID: 123, Content: "hello"}  // String choice
// or
msg := &Message{ID: 123, Content: int64(42)} // Integer choice
```

### 2. Union Struct Approach (Type-Safe)
```go
type MessageChoice struct {
    BoolValue   *bool   `asn1:"boolean,optional,tag:0"`
    IntValue    *int64  `asn1:"integer,optional,tag:1"`  
    StringValue *string `asn1:"utf8string,optional,tag:2"`
}
// Only one field should be non-nil

type Message struct {
    ID      int64         `asn1:"integer"`
    Content MessageChoice `asn1:"choice"`
}
```

### 3. Manual ASN1Choice
```go
type Message struct {
    ID      int64            `asn1:"integer"`
    Content *asn1.ASN1Choice // Direct use of ASN1Choice type
}
```

## Examples

Run the demo:
```bash
cd examples/demo
go run main.go
```

## Testing

```bash
go test -v
```

## License

Apache 2.0
