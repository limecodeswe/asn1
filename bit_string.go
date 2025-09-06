package asn1

import (
	"fmt"
	"strings"
)

// ASN1BitString represents an ASN.1 BIT STRING
type ASN1BitString struct {
	value     []byte
	unusedBits int // number of unused bits in the last byte (0-7)
}

// NewBitString creates a new ASN1BitString
func NewBitString(value []byte, unusedBits int) *ASN1BitString {
	if unusedBits < 0 || unusedBits > 7 {
		panic(fmt.Sprintf("unused bits must be between 0 and 7, got %d", unusedBits))
	}
	
	// Create a copy to prevent external modification
	copied := make([]byte, len(value))
	copy(copied, value)
	
	return &ASN1BitString{
		value:      copied,
		unusedBits: unusedBits,
	}
}

// NewBitStringFromBits creates a new ASN1BitString from a bit string
func NewBitStringFromBits(bits string) *ASN1BitString {
	if len(bits) == 0 {
		return &ASN1BitString{value: []byte{}, unusedBits: 0}
	}
	
	// Calculate how many bytes we need
	numBytes := (len(bits) + 7) / 8
	value := make([]byte, numBytes)
	unusedBits := (8 - (len(bits) % 8)) % 8
	
	// Fill in the bits
	for i, bit := range bits {
		if bit == '1' {
			byteIndex := i / 8
			bitIndex := 7 - (i % 8) // MSB first
			value[byteIndex] |= 1 << uint(bitIndex)
		} else if bit != '0' {
			panic(fmt.Sprintf("invalid bit character: %c", bit))
		}
	}
	
	return &ASN1BitString{
		value:      value,
		unusedBits: unusedBits,
	}
}

// Value returns the bit string value bytes
func (b *ASN1BitString) Value() []byte {
	// Return a copy to prevent external modification
	result := make([]byte, len(b.value))
	copy(result, b.value)
	return result
}

// UnusedBits returns the number of unused bits in the last byte
func (b *ASN1BitString) UnusedBits() int {
	return b.unusedBits
}

// BitLength returns the total number of bits
func (b *ASN1BitString) BitLength() int {
	if len(b.value) == 0 {
		return 0
	}
	return len(b.value)*8 - b.unusedBits
}

// ToBitString returns the bit string as a string of '0' and '1' characters
func (b *ASN1BitString) ToBitString() string {
	if len(b.value) == 0 {
		return ""
	}
	
	var bits strings.Builder
	totalBits := b.BitLength()
	
	for i := 0; i < totalBits; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8) // MSB first
		
		if (b.value[byteIndex] & (1 << uint(bitIndex))) != 0 {
			bits.WriteByte('1')
		} else {
			bits.WriteByte('0')
		}
	}
	
	return bits.String()
}

// Tag returns the ASN.1 tag for BIT STRING
func (b *ASN1BitString) Tag() Tag {
	return NewUniversalTag(TagBitString, false)
}

// Encode returns the BER encoding of the bit string
func (b *ASN1BitString) Encode() ([]byte, error) {
	// BIT STRING encoding: first byte is unused bits count, followed by the data
	content := make([]byte, 1+len(b.value))
	content[0] = byte(b.unusedBits)
	copy(content[1:], b.value)
	
	return EncodeTLV(b.Tag(), content)
}

// String returns a string representation of the bit string
func (b *ASN1BitString) String() string {
	bitStr := b.ToBitString()
	if len(bitStr) <= 32 {
		return fmt.Sprintf("BIT STRING '%s'", bitStr)
	}
	// For long bit strings, show first few bits and the length
	return fmt.Sprintf("BIT STRING '%s...' (%d bits)", bitStr[:32], b.BitLength())
}

// CompactString returns a compact hex representation
func (b *ASN1BitString) CompactString() string {
	return fmt.Sprintf("BIT STRING (0x%x, %d unused bits)", b.value, b.unusedBits)
}

// DecodeBitStringValue decodes a bit string value from raw bytes
func DecodeBitStringValue(data []byte) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("bit string value cannot be empty")
	}
	
	unusedBits := int(data[0])
	if unusedBits > 7 {
		return nil, 0, fmt.Errorf("unused bits count cannot exceed 7, got %d", unusedBits)
	}
	
	// If there's only the unused bits byte and no actual data
	if len(data) == 1 {
		if unusedBits != 0 {
			return nil, 0, fmt.Errorf("unused bits must be 0 when no data bytes present")
		}
		return []byte{}, 0, nil
	}
	
	value := make([]byte, len(data)-1)
	copy(value, data[1:])
	
	return value, unusedBits, nil
}

// DecodeBitString decodes an ASN1BitString from BER-encoded data
func DecodeBitString(data []byte) (*ASN1BitString, int, error) {
	asn1Value, consumed, err := DecodeTLV(data)
	if err != nil {
		return nil, 0, err
	}

	if asn1Value.tag.Class != 0 || asn1Value.tag.Number != TagBitString {
		return nil, 0, fmt.Errorf("expected BIT STRING tag, got class=%d number=%d", asn1Value.tag.Class, asn1Value.tag.Number)
	}

	value, unusedBits, err := DecodeBitStringValue(asn1Value.value)
	if err != nil {
		return nil, 0, err
	}

	return NewBitString(value, unusedBits), consumed, nil
}