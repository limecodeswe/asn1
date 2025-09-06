package asn1

import (
	"fmt"
	"math/big"
)

// ASN1Enumerated represents an ASN.1 ENUMERATED value
type ASN1Enumerated struct {
	value *big.Int
	name  string // Optional name for the enumerated value
}

// NewEnumerated creates a new ENUMERATED with the given integer value
func NewEnumerated(value int64) *ASN1Enumerated {
	return &ASN1Enumerated{
		value: big.NewInt(value),
	}
}

// NewEnumeratedWithName creates a new ENUMERATED with the given value and name
func NewEnumeratedWithName(value int64, name string) *ASN1Enumerated {
	return &ASN1Enumerated{
		value: big.NewInt(value),
		name:  name,
	}
}

// NewEnumeratedFromBigInt creates a new ENUMERATED from a big.Int
func NewEnumeratedFromBigInt(value *big.Int) *ASN1Enumerated {
	return &ASN1Enumerated{
		value: new(big.Int).Set(value),
	}
}

// Value returns the integer value as a big.Int
func (e *ASN1Enumerated) Value() *big.Int {
	return new(big.Int).Set(e.value)
}

// Int64 returns the value as an int64 (may panic if value doesn't fit)
func (e *ASN1Enumerated) Int64() int64 {
	if !e.value.IsInt64() {
		panic("enumerated value too large for int64")
	}
	return e.value.Int64()
}

// Name returns the optional name for this enumerated value
func (e *ASN1Enumerated) Name() string {
	return e.name
}

// SetName sets the name for this enumerated value
func (e *ASN1Enumerated) SetName(name string) {
	e.name = name
}

// Tag returns the ASN.1 tag for ENUMERATED
func (e *ASN1Enumerated) Tag() Tag {
	return NewUniversalTag(TagEnumerated, false)
}

// Encode returns the BER encoding of the ENUMERATED
func (e *ASN1Enumerated) Encode() ([]byte, error) {
	// ENUMERATED is encoded the same way as INTEGER
	valueBytes := encodeIntegerValue(e.value)
	return EncodeTLV(e.Tag(), valueBytes)
}

// String returns a string representation of the ENUMERATED
func (e *ASN1Enumerated) String() string {
	if e.name != "" {
		return fmt.Sprintf("ENUMERATED{%s (%s)}", e.name, e.value.String())
	}
	return fmt.Sprintf("ENUMERATED{%s}", e.value.String())
}

// TaggedString returns a string representation with tag information
func (e *ASN1Enumerated) TaggedString() string {
	if e.name != "" {
		return fmt.Sprintf("%s ENUMERATED: %s (%s)", e.Tag().TagString(), e.name, e.value.String())
	}
	return fmt.Sprintf("%s ENUMERATED: %s", e.Tag().TagString(), e.value.String())
}

// DecodeEnumeratedValue decodes the value bytes of an ENUMERATED
func DecodeEnumeratedValue(data []byte) (*big.Int, error) {
	// ENUMERATED uses the same encoding as INTEGER
	return DecodeIntegerValue(data)
}

// DecodeEnumerated decodes an ENUMERATED from BER data
func DecodeEnumerated(data []byte) (*ASN1Enumerated, int, error) {
	value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode TLV: %w", err)
	}

	expectedTag := NewUniversalTag(TagEnumerated, false)
	if value.Tag() != expectedTag {
		return nil, 0, fmt.Errorf("expected ENUMERATED tag %+v, got %+v", expectedTag, value.Tag())
	}

	intValue, err := DecodeEnumeratedValue(value.Value())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode enumerated value: %w", err)
	}

	return NewEnumeratedFromBigInt(intValue), consumed, nil
}

// Helper function to encode integer value (reused from integer.go logic)
func encodeIntegerValue(value *big.Int) []byte {
	if value.Sign() == 0 {
		return []byte{0x00}
	}

	// Get the byte representation
	bytes := value.Bytes()
	
	// For negative numbers, we need two's complement
	if value.Sign() < 0 {
		// Calculate two's complement
		// First, determine the minimum number of bytes needed
		bitLen := value.BitLen() + 1 // +1 for sign bit
		byteLen := (bitLen + 7) / 8
		
		// Create a mask for the required number of bytes
		maxVal := new(big.Int).Lsh(big.NewInt(1), uint(byteLen*8))
		
		// Calculate two's complement: maxVal + value (since value is negative)
		result := new(big.Int).Add(maxVal, value)
		bytes = result.Bytes()
		
		// Ensure we have the right number of bytes
		for len(bytes) < byteLen {
			bytes = append([]byte{0xFF}, bytes...)
		}
	} else {
		// For positive numbers, ensure the most significant bit is 0
		// If it's 1, we need to prepend a 0x00 byte
		if len(bytes) > 0 && bytes[0]&0x80 != 0 {
			bytes = append([]byte{0x00}, bytes...)
		}
	}
	
	return bytes
}