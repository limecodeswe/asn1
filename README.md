# ASN.1 Library for Go

A complete, production-ready Go library for working with ASN.1 data using BER encoding. This library implements all core ASN.1 types as dedicated Go structs with comprehensive encoding/decoding functionality, context-specific tag support, and realistic usage examples.

## Features

- **Complete ASN.1 Type Support**: All universal types implemented as dedicated structs
- **BER Encoding/Decoding**: Full Basic Encoding Rules (BER) implementation
- **Type-Safe API**: Idiomatic Go interfaces with strong typing
- **Context-Specific Tags**: Full support for optional and context-specific elements
- **Structured Types**: Unified handling of SEQUENCE and SET with nested structures
- **Realistic Examples**: Complete Person/Employee management example
- **Round-Trip Integrity**: Encode/decode operations preserve data integrity
- **Comprehensive Testing**: Full test coverage for all functionality
- **Easy Integration**: Simple, clean API designed for real-world applications

## Supported ASN.1 Types

### Primitive Types
- **BOOLEAN** - `ASN1Boolean`
- **INTEGER** - `ASN1Integer` (supports big.Int for arbitrary precision)
- **BIT STRING** - `ASN1BitString`
- **OCTET STRING** - `ASN1OctetString`
- **NULL** - `ASN1Null`
- **OBJECT IDENTIFIER** - `ASN1ObjectIdentifier`

### String Types
- **UTF8String** - `ASN1UTF8String`
- **PrintableString** - `ASN1PrintableString`
- **IA5String** - `ASN1IA5String`

### Structured Types
- **SEQUENCE** - `ASN1Structured`
- **SET** - `ASN1Structured`
- **Context-Specific Tags** - `ASN1Structured` with custom tags

## Installation

```bash
go get github.com/limecodeswe/asn1
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/limecodeswe/asn1"
)

func main() {
    // Create basic types
    name := asn1.NewUTF8String("John Doe")
    age := asn1.NewInteger(30)
    active := asn1.NewBoolean(true)
    
    // Create a sequence
    person := asn1.NewSequence()
    person.Add(name)
    person.Add(age)
    person.Add(active)
    
    // Encode to bytes
    encoded, err := person.Encode()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Encoded %d bytes: %02X\n", len(encoded), encoded)
    
    // Decode back
    decoded, _, err := asn1.DecodeTLV(encoded)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Decoded: %s\n", decoded.String())
}
```

## API Overview

### Core Interface

All ASN.1 objects implement the `ASN1Object` interface:

```go
type ASN1Object interface {
    Encode() ([]byte, error)  // BER encode the object
    String() string           // Human-readable representation
    Tag() Tag                 // Get the ASN.1 tag
}
```

### Creating Basic Types

```go
// Primitive types
boolean := asn1.NewBoolean(true)
integer := asn1.NewInteger(42)
octetString := asn1.NewOctetStringFromString("Hello")
null := asn1.NewNull()

// Bit strings
bitString := asn1.NewBitStringFromBits("10101010")

// Object identifiers
oid, _ := asn1.NewObjectIdentifierFromString("1.2.840.113549.1.1.1")

// String types
utf8Str := asn1.NewUTF8String("Hello, 世界!")
printableStr := asn1.NewPrintableString("Hello World")
ia5Str := asn1.NewIA5String("user@example.com")
```

### Creating Structured Types

```go
// SEQUENCE
seq := asn1.NewSequence()
seq.Add(asn1.NewInteger(1))
seq.Add(asn1.NewUTF8String("First"))
seq.Add(asn1.NewBoolean(true))

// SET
set := asn1.NewSet()
set.Add(asn1.NewInteger(100))
set.Add(asn1.NewPrintableString("Element"))
```

### Context-Specific Tags

```go
// Create optional field with context-specific tag [0]
contextTag := asn1.NewContextSpecificTag(0, false)
optional := asn1.NewStructured(contextTag)
optional.Add(asn1.NewUTF8String("Optional value"))

// Add to main sequence
mainSeq := asn1.NewSequence()
mainSeq.Add(asn1.NewInteger(1))        // Required field
mainSeq.Add(optional)                   // Optional field [0]
```

### Working with the Person Example

The library includes a realistic `Person` structure demonstrating real-world usage:

```go
// Create a person with required fields
person := asn1.NewPerson(123, "John Doe", "john@example.com", true)

// Add optional fields using method chaining
person.SetDepartment("Engineering").
       SetPhoneNumber("+1-555-0123").
       SetSalary(75000).
       SetEmployeeType(0). // 0=full-time, 1=part-time, 2=contractor
       SetPermissions([]string{"read", "write"})

// Set birthday
birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
person.SetBirthday(birthday)

// Create a manager relationship
manager := asn1.NewPerson(1, "Alice Johnson", "alice@example.com", true)
person.SetManager(manager)

// Encode the complete structure
encoded, err := person.Encode()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Person encoded to %d bytes\n", len(encoded))
fmt.Println(person.String()) // Pretty-print the structure
```

### Person Directory (SEQUENCE OF)

```go
// Create a directory of persons
directory := asn1.NewPersonDirectory()
directory.AddPerson(person1)
directory.AddPerson(person2)

// Encode entire directory
encoded, err := directory.Encode()
```

## BER Encoding/Decoding

### Low-Level Encoding

```go
// Direct BER encoding
tag := asn1.NewUniversalTag(asn1.TagInteger, false)
value := []byte{0x01, 0x23}
encoded, err := asn1.EncodeTLV(tag, value)
```

### Low-Level Decoding

```go
// Decode raw BER data
asn1Value, consumed, err := asn1.DecodeTLV(data)
if err != nil {
    return err
}

// Access tag and value
tag := asn1Value.Tag()
rawValue := asn1Value.Value()
```

### Decode All Objects

```go
// Decode multiple ASN.1 objects from data
objects, err := asn1.DecodeAll(data)
for i, obj := range objects {
    fmt.Printf("Object %d: %s\n", i, obj.String())
}
```

## Advanced Features

### Custom Tags

```go
// Application-specific tag [APPLICATION 1]
appTag := asn1.NewTag(1, false, 1) // class=1 (application)
appSpecific := asn1.NewStructured(appTag)

// Private tag [PRIVATE 5]
privateTag := asn1.NewTag(3, true, 5) // class=3 (private)
privateStruct := asn1.NewStructured(privateTag)
```

### Large Integers

```go
// The library supports arbitrary precision integers
import "math/big"

largeBig := new(big.Int)
largeBig.SetString("123456789012345678901234567890", 10)
largeInt := asn1.NewIntegerFromBigInt(largeBig)

encoded, _ := largeInt.Encode()
```

### Pretty Printing

```go
// Compact string representation
fmt.Println(structure.CompactString())

// Standard string representation  
fmt.Println(structure.String())
```

## Testing

Run the comprehensive test suite:

```bash
go test -v
```

The tests cover:
- All primitive types encoding/decoding
- Structured types (SEQUENCE/SET)
- Context-specific tags
- Object identifiers
- Person example structures
- Round-trip integrity
- Edge cases and error handling

## Examples

See the [examples/demo](examples/demo/) directory for a comprehensive demonstration of all library features:

```bash
cd examples/demo
go run main.go
```

This will demonstrate:
1. Basic ASN.1 types
2. Structured types (SEQUENCE/SET)
3. Realistic Person example
4. Context-specific tags
5. Object identifiers
6. Encoding/decoding round trips

## ASN.1 Specification Compliance

This library implements:
- ITU-T X.690 Basic Encoding Rules (BER)
- All universal ASN.1 types
- Context-specific and application-specific tags
- Proper tag, length, value (TLV) encoding
- Two's complement integer encoding
- Proper string type validation

## Performance

The library is designed for both correctness and performance:
- Zero-copy decoding where possible
- Efficient big integer handling
- Minimal memory allocations
- Fast BER encoding/decoding

## Contributing

Contributions are welcome! Please ensure that:
- All tests pass (`go test -v`)
- Code follows Go best practices
- New features include comprehensive tests
- Documentation is updated accordingly

## License

This project is open source. See LICENSE file for details.
